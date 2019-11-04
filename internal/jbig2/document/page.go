/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package document

import (
	"fmt"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document/segments"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// Page represents JBIG2 Page structure.
// It contains all the included segments header definitions mapped to
// their number relation to the document and the resultant page bitmap.
type Page struct {
	// Segments relation of the page number to their structures.
	Segments []*segments.Header

	// PageNumber defines this page number.
	// NOTE: page numeration starts from 1.
	PageNumber int

	// Bitmap represents the page image.
	Bitmap *bitmap.Bitmap

	FinalHeight int
	FinalWidth  int
	ResolutionX int
	ResolutionY int

	// Document is a relation to page's document
	Document *Document
}

// Encode encodes all segments located on the given page and writes the results to the 'w' BinaryWriter.
func (p *Page) Encode(w writer.BinaryWriter) (n int, err error) {
	return n, nil
}

// GetBitmap implements segments.Pager interface.
func (p *Page) GetBitmap() (bm *bitmap.Bitmap, err error) {
	common.Log.Trace(fmt.Sprintf("[PAGE][#%d] GetBitmap begins...", p.PageNumber))
	defer func() {
		if err != nil {
			common.Log.Trace(fmt.Sprintf("[PAGE][#%d] GetBitmap failed. %v", p.PageNumber, err))
		} else {
			common.Log.Trace(fmt.Sprintf("[PAGE][#%d] GetBitmap finished", p.PageNumber))
		}
	}()

	if p.Bitmap != nil {
		return p.Bitmap, nil
	}

	err = p.composePageBitmap()
	if err != nil {
		return nil, err
	}

	return p.Bitmap, nil
}

// GetSegment implements segments.Pager interface.
func (p *Page) GetSegment(number int) (*segments.Header, error) {
	for _, h := range p.Segments {
		if h.SegmentNumber == uint32(number) {
			return h, nil
		}
	}
	return nil, errors.Error("Page.GetSegment", "segment not found")
}

// String implements Stringer interface.
func (p *Page) String() string {
	return fmt.Sprintf("Page #%d", p.PageNumber)
}

// newPage is the creator for the Page structure.
func newPage(d *Document, pageNumber int) *Page {
	return &Page{Document: d, PageNumber: pageNumber, Segments: []*segments.Header{}}
}

// composePageBitmap composes the segment's bitmaps
// as a single page Bitmap.
func (p *Page) composePageBitmap() error {
	const processName = "composePageBitmap"
	if p.PageNumber == 0 {
		return nil
	}
	h := p.getPageInformationSegment()
	if h == nil {
		return errors.Error(processName, "page information segment not found")
	}

	// get the Segment data
	seg, err := h.GetSegmentData()
	if err != nil {
		return err
	}

	pageInformation, ok := seg.(*segments.PageInformationSegment)
	if !ok {
		return errors.Error(processName, "page information segment is of invalid type")
	}

	if err = p.createPage(pageInformation); err != nil {
		return errors.Wrap(err, processName, "")
	}
	p.clearSegmentData()
	return nil
}

func (p *Page) createPage(i *segments.PageInformationSegment) error {
	var err error
	if !i.IsStripe || i.PageBMHeight != -1 {
		// Page 79, 4)
		err = p.createNormalPage(i)
	} else {
		err = p.createStripedPage(i)
	}
	return err
}

func (p *Page) createNormalPage(i *segments.PageInformationSegment) error {
	const processName = "createNormalPage"
	p.Bitmap = bitmap.New(i.PageBMWidth, i.PageBMHeight)

	// Page 79, 3)
	// if default pixel value is not 0, byte will be filled with 0xff
	if i.DefaultPixelValue() != 0 {
		p.Bitmap.SetDefaultPixel()
	}

	for _, h := range p.Segments {
		switch h.Type {
		case 6, 7, 22, 23, 38, 39, 42, 43:
			common.Log.Trace("Getting Segment: %d", h.SegmentNumber)
			s, err := h.GetSegmentData()
			if err != nil {
				return err
			}

			r, ok := s.(segments.Regioner)
			if !ok {
				common.Log.Debug("Segment: %T is not a Regioner", s)
				return errors.Errorf(processName, "invalid jbig2 segment type - not a Regioner: %T", s)
			}

			regionBitmap, err := r.GetRegionBitmap()
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			if p.fitsPage(i, regionBitmap) {
				p.Bitmap = regionBitmap
			} else {
				regionInfo := r.GetRegionInfo()
				op := p.getCombinationOperator(i, regionInfo.CombinaionOperator)
				err = bitmap.Blit(regionBitmap, p.Bitmap, int(regionInfo.XLocation), int(regionInfo.YLocation), op)
				if err != nil {
					return errors.Wrap(err, processName, "")
				}
			}
			break
		}
	}
	return nil
}

func (p *Page) createStripedPage(i *segments.PageInformationSegment) error {
	const processName = "createStripedPage"
	pageStripes, err := p.collectPageStripes()
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	var startLine int
	for _, sd := range pageStripes {
		if eos, ok := sd.(*segments.EndOfStripe); ok {
			startLine = eos.LineNumber() + 1
		} else {
			r := sd.(segments.Regioner)
			regionInfo := r.GetRegionInfo()
			op := p.getCombinationOperator(i, regionInfo.CombinaionOperator)
			regionBitmap, err := r.GetRegionBitmap()
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			err = bitmap.Blit(regionBitmap, p.Bitmap, int(regionInfo.XLocation), startLine, op)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}
		}
	}
	return nil
}

func (p *Page) collectPageStripes() (stripes []segments.Segmenter, err error) {
	const processName = "collectPageStripes"
	var s segments.Segmenter

	for _, h := range p.Segments {
		switch h.Type {
		case 6, 7, 22, 23, 38, 39, 42, 43:
			s, err = h.GetSegmentData()
			if err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
			stripes = append(stripes, s)
		case 50:
			s, err = h.GetSegmentData()
			if err != nil {
				return nil, err
			}

			eos, ok := s.(*segments.EndOfStripe)
			if !ok {
				return nil, errors.Errorf(processName, "EndOfStripe is not of valid type: '%T'", s)
			}

			stripes = append(stripes, eos)
			p.FinalHeight = eos.LineNumber()
		}
	}
	return stripes, nil
}

func (p *Page) clearSegmentData() {
	for i := range p.Segments {
		p.Segments[i].CleanSegmentData()
	}
}

func (p *Page) clearPageData() {
	p.Bitmap = nil
}

// countRegions counts the region segments in the Page.
func (p *Page) countRegions() int {
	var regionCount int

	for _, h := range p.Segments {
		switch h.Type {
		case 6, 7, 22, 23, 38, 39, 42, 43:
			regionCount++
		}
	}
	return regionCount
}

// encodeSegment encodes the segment data and segment header for given 'segmentNumber'.
// Then the function writes it's encoded data into 'w' writer.
func (p *Page) encodeSegment(w writer.BinaryWriter, segmentNumber int) (n int, err error) {
	const processName = "encodeSegment"
	// get the segment for given 'segmentNumber'
	seg, err := p.GetSegment(segmentNumber)
	if err != nil {
		return n, errors.Wrap(err, processName, "")
	}

	var encoded []byte
	if seg.SegmentData != nil {
		if se, ok := seg.SegmentData.(segments.SegmentEncoder); ok {
			encoded, err = se.Encode()
			if err != nil {
				return 0, errors.Wrap(err, processName, "")
			}
		}
	}
	seg.SegmentDataLength = uint64(len(encoded))
	if n, err = seg.Encode(w); err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	if encoded != nil {
		var nd int
		nd, err = w.Write(encoded)
		if err != nil {
			return n, errors.Wrap(err, processName, "")
		}
		n += nd
	}
	return n, nil
}

func (p *Page) fitsPage(i *segments.PageInformationSegment, regionBitmap *bitmap.Bitmap) bool {
	return p.countRegions() == 1 &&
		i.DefaultPixelValue() == 0 &&
		i.PageBMWidth == regionBitmap.Width &&
		i.PageBMHeight == regionBitmap.Height
}

func (p *Page) getCombinationOperator(i *segments.PageInformationSegment, newOperator bitmap.CombinationOperator) bitmap.CombinationOperator {
	if i.CombinationOperatorOverrideAllowed() {
		return newOperator
	}
	return i.CombinationOperator()
}

// getPageInformationSegment returns the associated page information segment.
func (p *Page) getPageInformationSegment() *segments.Header {
	for _, s := range p.Segments {
		if s.Type == segments.TPageInformation {
			return s
		}
	}
	common.Log.Debug("Page information segment not found for page: %s.", p)
	return nil
}

func (p *Page) getHeight() (int, error) {
	const processName = "getHeight"
	if p.FinalHeight != 0 {
		return p.FinalHeight, nil
	}

	h := p.getPageInformationSegment()
	if h == nil {
		return 0, errors.Error(processName, "nil page information")
	}

	s, err := h.GetSegmentData()
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}

	pi, ok := s.(*segments.PageInformationSegment)
	if !ok {
		return 0, errors.Errorf(processName, "page information segment is of invalid type: '%T'", s)
	}

	if pi.PageBMHeight == math.MaxInt32 {
		_, err = p.GetBitmap()
		if err != nil {
			return 0, errors.Wrap(err, processName, "")
		}
	} else {
		p.FinalHeight = pi.PageBMHeight
	}
	return p.FinalHeight, nil
}

func (p *Page) getWidth() (int, error) {
	const processName = "getWidth"
	if p.FinalWidth != 0 {
		return p.FinalWidth, nil
	}

	h := p.getPageInformationSegment()
	if h == nil {
		return 0, errors.Error(processName, "nil page information")
	}

	s, err := h.GetSegmentData()
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}

	pi, ok := s.(*segments.PageInformationSegment)
	if !ok {
		return 0, errors.Errorf(processName, "page information segment is of invalid type: '%T'", s)
	}
	p.FinalWidth = pi.PageBMWidth
	return p.FinalWidth, nil
}

func (p *Page) getResolutionX() (int, error) {
	const processName = "getResolutionX"
	if p.ResolutionX != 0 {
		return p.ResolutionX, nil
	}
	h := p.getPageInformationSegment()
	if h == nil {
		return 0, errors.Error(processName, "nil page information")
	}

	s, err := h.GetSegmentData()
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}

	pi, ok := s.(*segments.PageInformationSegment)
	if !ok {
		return 0, errors.Errorf(processName, "page information segment is of invalid type: '%T'", s)
	}
	p.ResolutionX = pi.ResolutionX
	return p.ResolutionX, nil
}

func (p *Page) getResolutionY() (int, error) {
	const processName = "getResolutionY"
	if p.ResolutionY != 0 {
		return p.ResolutionY, nil
	}
	h := p.getPageInformationSegment()
	if h == nil {
		return 0, errors.Error(processName, "nil page information")
	}

	s, err := h.GetSegmentData()
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}

	pi, ok := s.(*segments.PageInformationSegment)
	if !ok {
		return 0, errors.Errorf(processName, "page information segment is of invalid type:'%T'", s)
	}
	p.ResolutionY = pi.ResolutionY
	return p.ResolutionY, nil
}
