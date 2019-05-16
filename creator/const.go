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
// Points per mm. (Default resolution).
var PPMM = float64(72 * 1.0 / 25.4)

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

// TextRenderingMode determines whether showing text shall cause glyph
// outlines to be stroked, filled, used as a clipping boundary, or some
// combination of the three.
// See section 9.3 "Text State Parameters and Operators" and
// Table 106 (pp. 254-255 PDF32000_2008).
type TextRenderingMode int

const (
	// TextRenderingModeFill (default) - Fill text.
	TextRenderingModeFill TextRenderingMode = iota

	// TextRenderingModeStroke - Stroke text.
	TextRenderingModeStroke

	// TextRenderingModeFillStroke - Fill, then stroke text.
	TextRenderingModeFillStroke

	// TextRenderingModeInvisible - Neither fill nor stroke text (invisible).
	TextRenderingModeInvisible

	// TextRenderingModeFillClip - Fill text and add to path for clipping.
	TextRenderingModeFillClip

	// TextRenderingModeStrokeClip - Stroke text and add to path for clipping.
	TextRenderingModeStrokeClip

	// TextRenderingModeFillStrokeClip - Fill, then stroke text and add to path for clipping.
	TextRenderingModeFillStrokeClip

	// TextRenderingModeClip - Add text to path for clipping.
	TextRenderingModeClip
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

// HorizontalAlignment represents the horizontal alignment of components
// within a page.
type HorizontalAlignment int

// Horizontal alignment options.
const (
	HorizontalAlignmentLeft HorizontalAlignment = iota
	HorizontalAlignmentCenter
	HorizontalAlignmentRight
)
