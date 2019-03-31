package jbig2

import (
	"errors"
	"fmt"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
)

// Page represents JBIG2 Page structure
type Page struct {
	// All segments for this page
	Segments map[int]*SegmentHeader

	// NOTE: page number != segmentList.index
	PageNumber int

	// The page bitmap represents the page buffer
	Bitmap *bitmap.Bitmap

	FinalHeight int
	FinalWidth  int
	ResolutionX int
	ResolutionY int

	Document *Document
}

// NewPage is the creator for the Page model
func NewPage(d *Document, pageNumber int) *Page {
	common.Log.Debug("Creating Page #%d...", pageNumber)
	return &Page{Document: d, PageNumber: pageNumber, Segments: map[int]*SegmentHeader{}}
}

// GetSegment searches for a segment specified by its number
func (p *Page) GetSegment(number int) *SegmentHeader {
	s, ok := p.Segments[number]
	if ok {
		return s
	}

	if !ok || s == nil {
		s, _ = p.Document.GlobalSegments.GetSegment(number)
		return s
	}

	common.Log.Info("Segment not found, returning nil.")
	return nil
}

// getPageInformationSegment Returns the associated page information segment
func (p *Page) getPageInformationSegment() *SegmentHeader {
	for _, s := range p.Segments {
		if s.SegmentType == TPageInformation {
			return s
		}
	}

	common.Log.Info("Page information segment not found.")

	return nil
}

// GetBitmap returns the decoded bitmap if present
func (p *Page) GetBitmap() (*bitmap.Bitmap, error) {
	if p.Bitmap == nil {
		err := p.composePageBitmap()
		if err != nil {
			return nil, err
		}
	}
	return p.Bitmap, nil
}

// composePageBitmap composes the segment's bitmaps to a page and stores the page as a Bitmap
func (p *Page) composePageBitmap() error {
	if p.PageNumber > 0 {
		h := p.getPageInformationSegment()
		if h == nil {
			return errors.New("Page Information segment not found")
		}

		// get the Segment data
		seg, err := h.getSegmentData()
		if err != nil {
			return err
		}

		pageInformation, ok := seg.(*PageInformationSegment)
		if !ok {
			return errors.New("PageInformation Segment is of invalid type")
		}

		if err = p.createPage(pageInformation); err != nil {
			return err
		}
		p.clearSegmentData()
	}

	return nil
}

func (p *Page) createPage(i *PageInformationSegment) (err error) {
	if !i.IsStripe || i.PageBMHeight != -1 {
		// Page 79, 4)
		err = p.createNormalPage(i)
	} else {
		err = p.createStripedPage(i)
	}
	return
}

func (p *Page) createNormalPage(i *PageInformationSegment) error {
	p.Bitmap = bitmap.New(i.PageBMWidth, i.PageBMHeight)

	// Page 79, 3)
	// if default pixel value is not 0, byte will be filled with 0xff
	if i.defaultPixelValue != 0 {
		p.Bitmap.SetDefaultPixel()

	}

	for _, h := range p.Segments {
		switch h.SegmentType {
		case 6, 7, 22, 23, 38, 39, 42, 43:
			s, err := h.getSegmentData()
			if err != nil {
				return err
			}

			r, ok := s.(Regioner)
			if !ok {
				common.Log.Debug("Segment: %T is not a Regioner", s)
				return errors.New("Current segment type is not a regioner")
			}

			regionBitmap, err := r.GetRegionBitmap()
			if err != nil {
				return err
			}

			if p.fitsPage(i, regionBitmap) {
				p.Bitmap = regionBitmap
			} else {
				regionInfo := r.GetRegionInfo()
				op := p.getCombinationOperator(i, regionInfo.CombinaionOperator)
				err = bitmap.Blit(regionBitmap, p.Bitmap, regionInfo.XLocation, regionInfo.YLocation, op)
				if err != nil {
					return err
				}
			}
			break
		}
	}

	return nil
}

func (p *Page) createStripedPage(i *PageInformationSegment) error {
	pageStripes, err := p.collectPageStripes()
	if err != nil {
		return err
	}

	var startLine int
	for _, sd := range pageStripes {
		if eos, ok := sd.(*EndOfStripe); ok {
			startLine = eos.lineNumber + 1
		} else {
			r := sd.(Regioner)
			regionInfo := r.GetRegionInfo()
			op := p.getCombinationOperator(i, regionInfo.CombinaionOperator)
			regionBitmap, err := r.GetRegionBitmap()
			if err != nil {
				return err
			}

			err = bitmap.Blit(regionBitmap, p.Bitmap, regionInfo.XLocation, startLine, op)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Page) fitsPage(i *PageInformationSegment, regionBitmap *bitmap.Bitmap) bool {
	return p.countRegions() != 1 &&
		i.defaultPixelValue == 0 &&
		i.PageBMHeight == regionBitmap.Width &&
		i.PageBMHeight == regionBitmap.Height
}

// countRegions count the regions segments in the Page
func (p *Page) countRegions() int {
	var regionCount int

	for _, h := range p.Segments {
		switch h.SegmentType {
		case 6, 7, 22, 23, 38, 39, 42, 43:
			regionCount++
		}
	}
	return regionCount
}

func (p *Page) getCombinationOperator(i *PageInformationSegment, newOperator bitmap.CombinationOperator) bitmap.CombinationOperator {
	if i.combinaitonOperatorOverrideAllowed {
		return newOperator
	}
	return i.combinationOperator
}

func (p *Page) collectPageStripes() (stripes []Segmenter, err error) {
	for _, h := range p.Segments {
		switch h.SegmentType {
		case 6, 7, 22, 23, 38, 39, 42, 43:
			var s Segmenter
			s, err = h.getSegmentData()
			if err != nil {
				return
			}
			stripes = append(stripes, s)
		case 50:
			var s Segmenter
			s, err = h.getSegmentData()
			if err != nil {
				return
			}
			eos, ok := s.(*EndOfStripe)
			if !ok {
				err = errors.New("EndOfStrip Segmenter is of invalid type")
				return
			}

			stripes = append(stripes, eos)
			p.FinalHeight = eos.lineNumber
		}
	}
	return
}

func (p *Page) clearSegmentData() {
	for i := range p.Segments {
		p.Segments[i].cleanSegmentData()
	}
}

func (p *Page) clearPageData() {
	p.Bitmap = nil
}

func (p *Page) getHeight() (int, error) {
	if p.FinalHeight == 0 {
		h := p.getPageInformationSegment()
		// if err != nil {
		// 	return 0, err
		// }
		if h == nil {
			return 0, errors.New("Nil page information")
		}

		s, err := h.getSegmentData()
		if err != nil {
			return 0, err
		}

		pi, ok := s.(*PageInformationSegment)
		if !ok {
			return 0, errors.New("PageInformationSegment is of invalid type")
		}
		if pi.PageBMHeight == 0xffffffff {
			_, err = p.GetBitmap()
			if err != nil {
				return 0, err
			}
		} else {
			p.FinalHeight = pi.PageBMHeight
		}
	}

	return p.FinalHeight, nil
}

func (p *Page) getWidth() (int, error) {
	if p.FinalWidth == 0 {
		h := p.getPageInformationSegment()
		// if err != nil {
		// 	return 0, err
		// }
		if h == nil {
			return 0, errors.New("Nil page information")
		}

		s, err := h.getSegmentData()
		if err != nil {
			return 0, err
		}

		pi, ok := s.(*PageInformationSegment)
		if !ok {
			return 0, errors.New("PageInformationSegment is of invalid type")
		}

		p.FinalWidth = pi.PageBMWidth
	}

	return p.FinalWidth, nil
}

func (p *Page) getResolutionX() (int, error) {
	if p.ResolutionX == 0 {
		h := p.getPageInformationSegment()
		// if err != nil {
		// 	return 0, err
		// }
		if h == nil {
			return 0, errors.New("Nil page information")
		}

		s, err := h.getSegmentData()
		if err != nil {
			return 0, err
		}

		pi, ok := s.(*PageInformationSegment)
		if !ok {
			return 0, errors.New("PageInformationSegment is of invalid type")
		}

		p.ResolutionX = pi.ResolutionX
	}

	return p.ResolutionX, nil
}

func (p *Page) getResolutionY() (int, error) {
	if p.ResolutionY == 0 {
		h := p.getPageInformationSegment()
		// if err != nil {
		// 	return 0, err
		// }
		if h == nil {
			return 0, errors.New("Nil page information")
		}

		s, err := h.getSegmentData()
		if err != nil {
			return 0, err
		}

		pi, ok := s.(*PageInformationSegment)
		if !ok {
			return 0, errors.New("PageInformationSegment is of invalid type")
		}

		p.ResolutionY = pi.ResolutionY
	}

	return p.ResolutionY, nil
}

// String implements Stringer interface
func (p *Page) String() string {
	return fmt.Sprintf("Page number: %d", p.PageNumber)
}
