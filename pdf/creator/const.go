/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

type PageSize int

const (
	PageSizeA3     PageSize = iota
	PageSizeA4              = iota
	PageSizeA5              = iota
	PageSizeLetter          = iota
	PageSizeLegal           = iota
)

// Page sizes in mm.
var pageSizesMM map[PageSize][2]float64 = map[PageSize][2]float64{
	PageSizeA3:     {297, 420},
	PageSizeA4:     {210, 297},
	PageSizeA5:     {148, 210},
	PageSizeLetter: {216, 280},
	PageSizeLegal:  {216, 356},
}

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
