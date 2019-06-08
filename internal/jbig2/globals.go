/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package jbig2

import (
	"errors"

	"github.com/unidoc/unipdf/v3/internal/jbig2/segments"
)

// Common errors definitions.
var (
	ErrNoGlobalsYet  = errors.New("no global segment added yet")
	ErrNoGlobalFound = errors.New("no global segment found")
)

// Globals store segments that aren't associated to a page.
// If the data is embedded in another format, for example PDF, this
// segments might be stored separately in the file.
// This segments will be decoded on demand, all results are stored in the document.
type Globals map[int]*segments.Header

// GetSegment gets the global segment header.
func (g Globals) GetSegment(segmentNumber int) (*segments.Header, error) {
	if len(g) == 0 {
		return nil, ErrNoGlobalsYet
	}

	v, ok := g[segmentNumber]
	if !ok {
		return nil, ErrNoGlobalFound
	}

	return v, nil
}

// AddSegment adds the segment to the globals store.
func (g Globals) AddSegment(segmentNumber int, segment *segments.Header) {
	g[segmentNumber] = segment
}
