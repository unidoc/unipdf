/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"os"

	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
)

// Loads the template from path as a list of pages.
func loadPagesFromFile(f *os.File) ([]*model.PdfPage, error) {
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}

	// Load the pages.
	var pages []*model.PdfPage
	for i := 0; i < numPages; i++ {
		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			return nil, err
		}

		pages = append(pages, page)
	}

	return pages, nil
}

// Rotates rectangle (0,0,w,h) by the specified angle and returns the its
// bounding box. The origin of rotation is the top left corner of the rectangle.
func rotateRect(w, h, angle float64) (x, y, rotatedWidth, rotatedHeight float64) {
	if angle == 0 {
		return 0, 0, w, h
	}

	// Get rotated size
	bbox := draw.Path{Points: []draw.Point{
		draw.NewPoint(0, 0).Rotate(angle),
		draw.NewPoint(w, 0).Rotate(angle),
		draw.NewPoint(0, h).Rotate(angle),
		draw.NewPoint(w, h).Rotate(angle),
	}}.GetBoundingBox()

	return bbox.X, bbox.Y, bbox.Width, bbox.Height
}
