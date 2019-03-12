package jbig2

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// Segmenter is the interface for all data pars of segments
type Segmenter interface {
	Init(header *SegmentHeader, r *reader.Reader)
}
