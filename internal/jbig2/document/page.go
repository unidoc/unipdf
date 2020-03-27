/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package document

import (
	"fmt"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document/segments"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// EncodingMethod defines the method of encoding for given page,
type EncodingMethod int

// enums that defines encoding method.
const (
	GenericEM EncodingMethod = iota
	CorrelationEM
	RankHausEM
)

// Page represents JBIG2 Page structure.
// It contains all the included segments header definitions mapped to
// their number relation to the document and the resultant page bitmap.
// NOTE: page numeration starts from 1 and the association to 0'th page means the segments
// are associated to global segments.
type Page struct {
	// Segments relation of the page number to their structures.
	Segments []*segments.Header
	// PageNumber defines this page number.
	PageNumber int
	// Bitmap represents the page image.
	Bitmap *bitmap.Bitmap

	// Page parameters
	FinalHeight int
	FinalWidth  int
	ResolutionX int
	ResolutionY int

	IsLossless bool

	// Document is a relation to page's document
	Document *Document
	// FirstSegmentNumber defines first segment number for given page
	FirstSegmentNumber int
	// EncodingMethod defines
	EncodingMethod EncodingMethod
}

// AddEndOfPageSegment adds the end of page segment.
func (p *Page) AddEndOfPageSegment() {
	seg := &segments.Header{
		Type:            segments.TEndOfPage,
		PageAssociation: p.PageNumber,
	}
	p.Segments = append(p.Segments, seg)
}

// AddGenericRegion adds the generic region to the page context.
// 'bm' 					- bitmap containing data to encode
// 'xloc' 					- x location of the generic region
// 'yloc'					- y location of the generic region
// 'template'				- generic region template
// 'tp'						- is the generic region type
// 'duplicateLineRemoval'	- is the flag that defines if the generic region segment should remove duplicated lines
func (p *Page) AddGenericRegion(bm *bitmap.Bitmap, xloc, yloc, template int, tp segments.Type, duplicateLineRemoval bool) error {
	const processName = "Page.AddGenericRegion"
	// create generic region segment
	genReg := &segments.GenericRegion{}
	if err := genReg.InitEncode(bm, xloc, yloc, template, duplicateLineRemoval); err != nil {
		return errors.Wrap(err, processName, "")
	}
	// create segment header for the generic region
	genRegSegmentHeader := &segments.Header{
		Type:            segments.TImmediateGenericRegion,
		PageAssociation: p.PageNumber,
		SegmentData:     genReg,
	}
	p.Segments = append(p.Segments, genRegSegmentHeader)
	return nil
}

// AddPageInformationSegment adds the page information segment to the page segments.
func (p *Page) AddPageInformationSegment() {
	// prepare page info segment data
	pageInfo := &segments.PageInformationSegment{
		PageBMWidth:  p.FinalWidth,
		PageBMHeight: p.FinalHeight,
		ResolutionX:  p.ResolutionX,
		ResolutionY:  p.ResolutionY,
		IsLossless:   p.IsLossless,
	}

	// and the page info header
	pageInfoHeader := &segments.Header{
		PageAssociation:   p.PageNumber,
		SegmentDataLength: uint64(pageInfo.Size()),
		SegmentData:       pageInfo,
		Type:              segments.TPageInformation,
	}
	p.Segments = append(p.Segments, pageInfoHeader)
}

// addTextRegionSegment adds text region segment to the given page.
// arguments:
// - referredTo is the referred to segments header
// - globalSymbolsMap 	- is the mapping between global symbols id and their classes.
// - localSymbolsMap 	- is the mapping between this page exclusive symbols id and their' classes.
// - comps 				- are the components numbers for this page.
// - inLL 				- is the slice of the lower-left corners of the boxes for each symbol
// - symbols 			- the slice of symbols
// - assignments 		-
func (p *Page) addTextRegionSegment(referredTo []*segments.Header, globalSymbolsMap, localSymbolsMap map[int]int, comps []int, inLL *bitmap.Points, symbols *bitmap.Bitmaps, classIDs *basic.IntSlice, boxes *bitmap.Boxes, symbits, sbNumInstances int) {
	textRegion := &segments.TextRegion{NumberOfSymbols: uint32(sbNumInstances)}
	textRegion.InitEncode(globalSymbolsMap, localSymbolsMap, comps, inLL, symbols, classIDs, boxes, p.FinalWidth, p.FinalHeight, symbits)

	textRegionHeader := &segments.Header{
		RTSegments:      referredTo,
		SegmentData:     textRegion,
		PageAssociation: p.PageNumber,
		Type:            segments.TImmediateTextRegion,
	}

	// if the text region referes only to global symbol dictionary
	// it shold be stored just after page information segment
	tp := segments.TPageInformation
	if localSymbolsMap != nil {
		// otherwise store it after local symbol dictionary
		tp = segments.TSymbolDictionary
	}

	var index int
	for ; index < len(p.Segments); index++ {
		if p.Segments[index].Type == tp {
			index++
			break
		}
	}
	p.Segments = append(p.Segments, nil)
	copy(p.Segments[index+1:], p.Segments[index:])
	p.Segments[index] = textRegionHeader
}

// Encode encodes segments into provided 'w' writer.
func (p *Page) Encode(w writer.BinaryWriter) (n int, err error) {
	const processName = "Page.Encode"
	var temp int
	for _, seg := range p.Segments {
		if temp, err = seg.Encode(w); err != nil {
			return n, errors.Wrap(err, processName, "")
		}
		n += temp
	}
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

// GetHeight gets the page height.
func (p *Page) GetHeight() (int, error) {
	return p.getHeight()
}

// GetResolutionX gets the 'x' resolution of the page.
func (p *Page) GetResolutionX() (int, error) {
	return p.getResolutionX()
}

// GetResolutionY gets the 'y' resolution of the page.
func (p *Page) GetResolutionY() (int, error) {
	return p.getResolutionY()
}

// GetWidth gets the page width.
func (p *Page) GetWidth() (int, error) {
	return p.getWidth()
}

// GetSegment implements segments.Pager interface.
func (p *Page) GetSegment(number int) (*segments.Header, error) {
	const processName = "Page.GetSegment"

	for _, h := range p.Segments {
		if h.SegmentNumber == uint32(number) {
			return h, nil
		}
	}
	containedIDS := make([]uint32, len(p.Segments))
	for i, h := range p.Segments {
		containedIDS[i] = h.SegmentNumber
	}
	return nil, errors.Errorf(processName, "segment with number: '%d' not found in the page: '%d'. Known segment numbers: %v", number, p.PageNumber, containedIDS)
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

// lastSegmentNumber gets the number of the last segment in the page.
func (p *Page) lastSegmentNumber() (last uint32, err error) {
	const processName = "lastSegmentNumber"
	if len(p.Segments) == 0 {
		return last, errors.Errorf(processName, "no segments found in the page '%d'", p.PageNumber)
	}
	return p.Segments[len(p.Segments)-1].SegmentNumber, nil
}

func (p *Page) nextSegmentNumber() uint32 {
	return p.Document.nextSegmentNumber()
}
