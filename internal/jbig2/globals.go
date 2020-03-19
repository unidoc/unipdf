package jbig2

import (
	"sort"

	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document/segments"
)

// Globals is the v3 mapping of the jbig2 segments to header mapping.
type Globals map[int]*segments.Header

// ToDocumentGlobals converts 'jbig2.Globals' into '*document.Globals'
func (g Globals) ToDocumentGlobals() *document.Globals {
	if g == nil {
		return nil
	}
	headers := []*segments.Header{}
	// add all segments to the slice.
	for _, segment := range g {
		headers = append(headers, segment)
	}
	// sort by the segment number value
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].SegmentNumber < headers[j].SegmentNumber
	})
	return &document.Globals{Segments: headers}
}
