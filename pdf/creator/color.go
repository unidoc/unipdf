/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"fmt"

	"math"

	"github.com/unidoc/unipdf/v3/common"
)

// Color interface represents colors in the PDF creator.
type Color interface {
	ToRGB() (float64, float64, float64)
}

// Represents RGB color values.
type rgbColor struct {
	// Arithmetic representation of r,g,b (range 0-1).
	r, g, b float64
}

func (col rgbColor) ToRGB() (float64, float64, float64) {
	return col.r, col.g, col.b
}

// Commonly used colors.
var (
	ColorBlack  = ColorRGBFromArithmetic(0, 0, 0)
	ColorWhite  = ColorRGBFromArithmetic(1, 1, 1)
	ColorRed    = ColorRGBFromArithmetic(1, 0, 0)
	ColorGreen  = ColorRGBFromArithmetic(0, 1, 0)
	ColorBlue   = ColorRGBFromArithmetic(0, 0, 1)
	ColorYellow = ColorRGBFromArithmetic(1, 1, 0)
)

// ColorRGBFromHex converts color hex code to rgb color for using with creator.
// NOTE: If there is a problem interpreting the string, then will use black color and log a debug message.
// Example hex code: #ffffff -> (1,1,1) white.
func ColorRGBFromHex(hexStr string) Color {
	color := rgbColor{}
	if (len(hexStr) != 4 && len(hexStr) != 7) || hexStr[0] != '#' {
		common.Log.Debug("Invalid hex code: %s", hexStr)
		return color
	}

	var r, g, b int
	if len(hexStr) == 4 {
		// Special case: 4 digits: #abc ; where r = a*16+a, e.g. #ffffff -> #fff
		var tmp1, tmp2, tmp3 int
		n, err := fmt.Sscanf(hexStr, "#%1x%1x%1x", &tmp1, &tmp2, &tmp3)

		if err != nil {
			common.Log.Debug("Invalid hex code: %s, error: %v", hexStr, err)
			return color
		}
		if n != 3 {
			common.Log.Debug("Invalid hex code: %s", hexStr)
			return color
		}

		r = tmp1*16 + tmp1
		g = tmp2*16 + tmp2
		b = tmp3*16 + tmp3
	} else {
		// Default case: 7 digits: #rrggbb
		n, err := fmt.Sscanf(hexStr, "#%2x%2x%2x", &r, &g, &b)
		if err != nil {
			common.Log.Debug("Invalid hex code: %s", hexStr)
			return color
		}
		if n != 3 {
			common.Log.Debug("Invalid hex code: %s, n != 3 (%d)", hexStr, n)
			return color
		}
	}

	rNorm := float64(r) / 255.0
	gNorm := float64(g) / 255.0
	bNorm := float64(b) / 255.0

	color.r = rNorm
	color.g = gNorm
	color.b = bNorm

	return color
}

// ColorRGBFrom8bit creates a Color from 8bit (0-255) r,g,b values.
// Example:
//   red := ColorRGBFrom8Bit(255, 0, 0)
func ColorRGBFrom8bit(r, g, b byte) Color {
	color := rgbColor{}
	color.r = float64(r) / 255.0
	color.g = float64(g) / 255.0
	color.b = float64(b) / 255.0
	return color
}

// ColorRGBFromArithmetic creates a Color from arithmetic (0-1.0) color values.
// Example:
//   green := ColorRGBFromArithmetic(0, 1.0, 0)
func ColorRGBFromArithmetic(r, g, b float64) Color {
	// Ensure is in the range 0-1:
	r = math.Max(math.Min(r, 1.0), 0.0)
	g = math.Max(math.Min(g, 1.0), 0.0)
	b = math.Max(math.Min(b, 1.0), 0.0)

	color := rgbColor{}
	color.r = r
	color.g = g
	color.b = b
	return color
}
