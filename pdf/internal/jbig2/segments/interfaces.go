/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// Documenter is the interface used for the document model
type Documenter interface {
	GetPage(int) (Pager, error)
	GetGlobalSegment(int) *Header
}

// Pager is the interface used as a Page model
type Pager interface {
	GetSegment(int) *Header
	GetBitmap() (*bitmap.Bitmap, error)
}

// Segmenter is the interface for all data pars of segments
type Segmenter interface {
	Init(header *Header, r reader.StreamReader) error
}

// Regioner is the interface for all JBIG2 region segments
type Regioner interface {
	// GetRegionBitmap decodes and returns a regions content
	GetRegionBitmap() (*bitmap.Bitmap, error)

	// Returns RegionInfo
	GetRegionInfo() *RegionSegment
}
