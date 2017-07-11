/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"os"

	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model"
)

// Convert color hex code to rgb colors. Returns an error flag if there is a problem with interpreting the hex code.
// Example hex code: #ffffff -> (1,1,1) white.
func ColorRGBFromHex(hexStr string) (float64, float64, float64, error) {
	if (len(hexStr) != 4 && len(hexStr) != 7) || hexStr[0] != '#' {
		common.Log.Debug("Invalid hex code: %s", hexStr)
		return 0, 0, 0, errors.New("Invalid hex color code")
	}

	var r, g, b int
	if len(hexStr) == 4 {
		// Special case: 4 digits: #abc ; where r = a*16+a, e.g. #ffffff -> #fff
		var tmp1, tmp2, tmp3 int
		n, err := fmt.Sscanf(hexStr, "#%1x%1x%1x", &tmp1, &tmp2, &tmp3)

		if err != nil {
			return 0, 0, 0, err
		}
		if n != 3 {
			return 0, 0, 0, errors.New("Invalid hex code")
		}

		r = tmp1*16 + tmp1
		g = tmp2*16 + tmp2
		b = tmp3*16 + tmp3
	} else {
		// Default case: 7 digits: #rrggbb
		n, err := fmt.Sscanf(hexStr, "#%2x%2x%2x", &r, &g, &b)
		if err != nil {
			return 0, 0, 0, err
		}
		if n != 3 {
			return 0, 0, 0, errors.New("Invalid hex code")
		}
	}

	rNorm := float64(r) / 255.0
	gNorm := float64(g) / 255.0
	bNorm := float64(b) / 255.0

	return rNorm, gNorm, bNorm, nil
}

func dimensionsMMtoPoints(dimensionsMM [2]float64, ppi float64) [2]float64 {
	width := dimensionsMM[0] / 25.4 * ppi
	height := dimensionsMM[1] / 25.4 * ppi
	return [2]float64{width, height}
}

// Loads the template from path as a list of pages.
func loadPagesFromFile(path string) ([]*model.PdfPage, error) {
	// Read the input pdf file.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}

	// Load the pages.
	pages := []*model.PdfPage{}
	for i := 0; i < numPages; i++ {
		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			return nil, err
		}

		pages = append(pages, page)
	}

	return pages, nil
}
