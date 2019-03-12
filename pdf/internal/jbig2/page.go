package jbig2

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
)

// Page represents JBIG2 Page structure
type Page struct {
	// All segments for this page
	Segments map[int]*SegmentHeader

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
	return &Page{Document: d, PageNumber: pageNumber}
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

// GetPageInformationSegment Returns the associated page information segment
func (p *Page) GetPageInformationSegment() *SegmentHeader {
	for _, s := range p.Segments {
		if s.SegmentType == TPageInformation {
			return s
		}
	}

	common.Log.Info("Page information segment not found.")

	return nil
}

func (p *Page) GetBitmap() (*bitmap.Bitmap, error) {
	if p.PageNumber > 0 {
		// TODO
	}
	return nil, nil
}

// composePageBitmap composes the segment's bitmaps to a page and stores the page as a Bitmap
func (p *Page) composePageBitmap() error {
	p.GetPageInformationSegment()
	return nil
}
