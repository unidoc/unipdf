/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// PageSize represents the page size as a 2 element array representing the width and height in PDF document units (points).
type PageSize [2]float64

// PPI specifies the default PDF resolution in points/inch.
var PPI float64 = 72 // Points per inch. (Default resolution).

// PPMM specifies the default PDF resolution in points/mm.
var PPMM float64 = 72 * 1.0 / 25.4 // Points per mm. (Default resolution).

//
// Commonly used page sizes
//
var (
	PageSizeA3     = PageSize{297 * PPMM, 420 * PPMM}
	PageSizeA4     = PageSize{210 * PPMM, 297 * PPMM}
	PageSizeA5     = PageSize{148 * PPMM, 210 * PPMM}
	PageSizeLetter = PageSize{8.5 * PPI, 11 * PPI}
	PageSizeLegal  = PageSize{8.5 * PPI, 14 * PPI}
)

// TextAlignment options for paragraph.
type TextAlignment int

// The options supported for text alignment are:
// left - TextAlignmentLeft
// right - TextAlignmentRight
// center - TextAlignmentCenter
// justify - TextAlignmentJustify
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
