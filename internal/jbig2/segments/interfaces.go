/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// Documenter is the interface used for the document model.
type Documenter interface {
	// GetPage gets the page at given page number.
	GetPage(int) (Pager, error)

	// GetGlobalSegment gets the global segment header at given segment number.
	GetGlobalSegment(int) *Header
}

// Pager is the interface used as a Page model.
type Pager interface {
	// GetSegment gets the segment Header with the given segment number.
	GetSegment(int) *Header

	// GetBitmap gets the decoded bitmap.Bitmap.
	GetBitmap() (*bitmap.Bitmap, error)
}

// Segmenter is the interface for all data pars of segments.
type Segmenter interface {
	// Init initializes the segment from the provided data stream 'r'.
	Init(header *Header, r reader.StreamReader) error
}

// Regioner is the interface for all JBIG2 region segments.
type Regioner interface {
	// GetRegionBitmap decodes and returns a regions content.
	GetRegionBitmap() (*bitmap.Bitmap, error)

	// GetRegionInfo returns RegionSegment information.
	GetRegionInfo() *RegionSegment
}
