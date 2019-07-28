/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"errors"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
)

// document is a testing structure that implements Documenter interface.
type document struct {
	pages   []Pager
	globals []*Header
}

// GetPage implements Documenter interface.
func (d *document) GetPage(i int) (Pager, error) {
	i--
	if i <= len(d.pages)-1 {
		return d.pages[i], nil
	}
	return nil, errors.New("No Such page")
}

// GetGlobalSegment implements Documenter interface.
func (d *document) GetGlobalSegment(i int) *Header {
	i--
	if i <= len(d.globals)-1 {
		return d.globals[i]
	}
	return nil
}

// page is a testing structure that implements Pager interface.
type page struct {
	segments []*Header
	bm       *bitmap.Bitmap
}

// GetSegment implements Pager interface.
func (p *page) GetSegment(i int) *Header {
	i--
	if i <= len(p.segments)-1 {
		return p.segments[i]
	}
	return nil
}

// GetBitmap implements Pager interface.
func (p *page) GetBitmap() (*bitmap.Bitmap, error) {
	return p.bm, nil
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
