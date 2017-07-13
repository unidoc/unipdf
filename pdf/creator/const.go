/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

type PageSize [2]float64

//
// Common page sizes in mm:
// A3: {297, 420}
// A4: {210, 297}
// A5: {148, 210}
// Letter: {216, 280}
// Legal: {216, 356}
// Converted to points at standard resolution 72ppi and rounded to 2 digits:
//

var PPI float64 = 72               // Points per inch. (Default resolution).
var PPMM float64 = 72 * 1.0 / 25.4 // Points per mm. (Default resolution).

var (
	PageSizeA3     = PageSize{297 * PPMM, 420 * PPMM}
	PageSizeA4     = PageSize{210 * PPMM, 297 * PPMM}
	PageSizeA5     = PageSize{148 * PPMM, 210 * PPMM}
	PageSizeLetter = PageSize{8.5 * PPI, 11 * PPI}
	PageSizeLegal  = PageSize{8.5 * PPI, 14 * PPI}
)

type TextAlignment int

const (
	TextAlignmentLeft TextAlignment = iota
	TextAlignmentRight
	TextAlignmentCenter
	TextAlignmentJustify
)

// Relative and absolute positioning types.
type positioning int

const (
	positionRelative positioning = iota
	positionAbsolute
)

func (p positioning) isRelative() bool {
	return p == positionRelative
}
func (p positioning) isAbsolute() bool {
	return p == positionAbsolute
}
