package segments

import (
	"errors"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
)

// this file contains test structs used only for testing purpose

type document struct {
	pages   []Pager
	globals []*Header
}

func (d *document) GetPage(i int) (Pager, error) {
	i--
	if i <= len(d.pages)-1 {
		return d.pages[i], nil
	}
	return nil, errors.New("No Such page")
}

func (d *document) GetGlobalSegment(i int) *Header {
	i--
	if i <= len(d.globals)-1 {
		return d.globals[i]
	}
	return nil
}

type page struct {
	segments []*Header
	bm       *bitmap.Bitmap
}

func (p *page) setSegment(h *Header) {
	for len(p.segments)-1 <= int(h.SegmentNumber) {
		var length int
		if len(p.segments) == 0 {
			length = 4
		} else {
			length = len(p.segments) * 2
		}
		temp := make([]*Header, length)
		copy(temp, p.segments)
		p.segments = temp
	}

	p.segments[int(h.SegmentNumber)] = h
}

func (p *page) GetSegment(i int) *Header {
	i--
	if i <= len(p.segments)-1 {
		return p.segments[i]
	}
	return nil
}

func (p *page) GetBitmap() (*bitmap.Bitmap, error) {
	return p.bm, nil
}
