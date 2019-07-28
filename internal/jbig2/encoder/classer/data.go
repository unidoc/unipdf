/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"image"
	"io"

	"github.com/unidoc/unipdf/internal/jbig2/bitmap"
)

// Data holds all the data required for the compressed jbig-type representation
// of a set of images.
type Data struct {
	// Pix is the template composite for all classes.
	Pix *bitmap.Pix
	// PageNumber is the number of pages.
	PageNumber int
	// MaxWidth is the max width of original page.
	MaxWidth int
	// MaxHeight is the max height of original page.
	MaxHeight int
	// ClassNumber is the number of classes.
	ClassNumber int
	// LatticeWidth is the lattice width for template.
	LatticeWidth int
	// LatticeHeight is the lattice height for template.
	LatticeHeight int
	// ClassIDs is the slice of class ids for each component.
	ClassIDs []int
	// PageNumbers is the slice of page numbers for each component.
	PageNumbers []int
	// PtaUL is the slice of UL corners at which the template
	// is to be placed for each component
	PtaUL []image.Point
}

// Write implements io.Writer interface.
func (d *Data) Write(p []byte) (int, error) {
	// TODO: jbclass.c:1951
	return 0, nil
}

// Read implements io.Reader interface.
func (d *Data) Read(p []byte) (int, error) {
	// TODO: jbclass.c:2011
	return 0, nil
}

// Render renders provided data into *bitmap.Pixa instance.
func (d *Data) Render() (*bitmap.Pixa, error) {
	return nil, nil
}
