/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package document

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/document/segments"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// Globals store segments that aren't associated to a page.
// If the data is embedded in another format, for example PDF, this
// segments might be stored separately in the file.
// These segments will be decoded on demand, all results are stored in the document.
type Globals struct {
	Segments []*segments.Header
}

// AddSegment adds the segment to the globals store.
func (g *Globals) AddSegment(segment *segments.Header) {
	g.Segments = append(g.Segments, segment)
}

// GetSegment gets the global segment header.
func (g *Globals) GetSegment(segmentNumber int) (*segments.Header, error) {
	const processName = "Globals.GetSegment"
	if g == nil {
		return nil, errors.Error(processName, "globals not defined")
	}

	if len(g.Segments) == 0 {
		return nil, errors.Error(processName, "globals are empty")
	}

	var segment *segments.Header
	for _, segment = range g.Segments {
		if segment.SegmentNumber == uint32(segmentNumber) {
			break
		}
	}
	if segment == nil {
		return nil, errors.Error(processName, "segment not found")
	}
	return segment, nil
}

// GetSegmentByIndex gets segments header by 'index' in the Globals.
func (g *Globals) GetSegmentByIndex(index int) (*segments.Header, error) {
	const processName = "Globals.GetSegmentByIndex"
	if g == nil {
		return nil, errors.Error(processName, "globals not defined")
	}

	if len(g.Segments) == 0 {
		return nil, errors.Error(processName, "globals are empty")
	}
	if index > len(g.Segments)-1 {
		return nil, errors.Error(processName, "index out of range")
	}
	return g.Segments[index], nil
}

// GetSymbolDictionary gets global symbol dictionary.
func (g *Globals) GetSymbolDictionary() (*segments.Header, error) {
	const processName = "Globals.GetSymbolDictionary"
	if g == nil {
		return nil, errors.Error(processName, "globals not defined")
	}

	if len(g.Segments) == 0 {
		return nil, errors.Error(processName, "globals are empty")
	}
	for _, seg := range g.Segments {
		if seg.Type == segments.TSymbolDictionary {
			return seg, nil
		}
	}
	return nil, errors.Error(processName, "global symbol dictionary not found")
}
