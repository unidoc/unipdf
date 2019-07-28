/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"
)

// Pix is the model used for encoding images.
// Based on the apache leptonica Pix struct.
type Pix struct {
	// Width in pixels
	Width uint
	// Height in pixels
	Heiht uint
	// Depth in bits (bpp)
	Depth uint
	// Number of samples per pixel
	SamplesPerLine uint
	// 32-bit words/line
	WordsPerLine uint
	// Reference count (1 if no clones)
	RefCount uint

	// Image resolution (ppi) in 'x' direction.
	XRes int
	// Image resolution (ppi) in 'y' direction.
	YRes int

	// Special instructions for I/O.
	Special int
	// Text string associated with the pix.
	Text string
	// Image data.
	Data []uint32
}

// Pixa is the slice of pixels and boxes container.
type Pixa struct {
	Pixels []*Pix
	Boxes  *Boxa
}

// Pixaa is the slice of slices of Pix.
type Pixaa struct {
	Pixels []*Pixa
	Boxes  *Boxa
}

// Box is image.Rectangle wrapper that contains reference count variable.
type Box struct {
	image.Rectangle
	// RefCount is the reference count (1 if no clones).
	RefCount uint
}

// Boxa is the slice of boxes with the reference count.
type Boxa struct {
	Boxes []*Box
	// RefCount is the reference count (1 if no clones).
	RefCount uint
}

// Boxaa is the slice of 'Boxa' pointers.
type Boxaa []*Boxa
