package jbig2

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// Segmenter is the interface for all data pars of segments
type Segmenter interface {
	Init(header *SegmentHeader, r reader.StreamReader) error
}

// Regioner is the interface for all JBIG2 region segments
type Regioner interface {
	// GetRegionBitmap decodes and returns a regions content
	GetRegionBitmap() (*bitmap.Bitmap, error)

	// Returns RegionInfo
	GetRegionInfo() *RegionSegment
}
