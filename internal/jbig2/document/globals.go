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
type Globals map[int]*segments.Header

// AddSegment adds the segment to the globals store.
func (g Globals) AddSegment(segmentNumber int, segment *segments.Header) {
	g[segmentNumber] = segment
}

// GetSegment gets the global segment header.
func (g Globals) GetSegment(segmentNumber int) (*segments.Header, error) {
	const processName = "Globals.GetSegment"
	if g == nil {
		return nil, errors.Error(processName, "globals not defined")
	}

	if len(g) == 0 {
		return nil, errors.Error(processName, "globals are empty")
	}

	v, ok := g[segmentNumber]
	if !ok {
		return nil, errors.Error(processName, "segment not found")
	}
	return v, nil
}
