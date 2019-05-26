/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package jbig2

import (
	"errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/segments"
)

// common errors definition
var (
	ErrNoGlobalsYet  error = errors.New("No global segment added yet.")
	ErrNoGlobalFound error = errors.New("No global segment found.")
)

// Globals store segments that aren't associated to a page
// If the data is embedded in another format, for example PDF, this segments might be stored
// separately in the file.
// This segments will be decoded on demand and all results are stored in the document object and
// can be retrieved from there.
type Globals map[int]*segments.Header

// GetSegment gets the global segment
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

// AddSegment adds the segment to the globals
func (g Globals) AddSegment(segmentNumber int, segment *segments.Header) {
	g[segmentNumber] = segment
}
