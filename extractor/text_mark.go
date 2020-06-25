/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"image/color"
	"math"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/internal/transform"
	"github.com/unidoc/unipdf/v3/model"
)

// textMark represents text drawn on a page and its position in device coordinates.
// All dimensions are in device coordinates.
type textMark struct {
	model.PdfRectangle                    // Bounding box oriented so character base is at bottom
	orient             int                // Orientation
	text               string             // The text (decoded via ToUnicode).
	original           string             // Original text (decoded).
	font               *model.PdfFont     // The font the mark was drawn with.
	fontsize           float64            // The font size the mark was drawn with.
	charspacing        float64            // TODO (peterwilliams97: Should this be exposed in TextMark?
	trm                transform.Matrix   // The current text rendering matrix (TRM above).
	end                transform.Point    // The end of character device coordinates.
	originaBBox        model.PdfRectangle // Bounding box without orientation correction.
	fillColor          color.Color        // Text fill color.
	strokeColor        color.Color        // Text stroke color.
}

// newTextMark returns a textMark for text `text` rendered with text rendering matrix (TRM) `trm`
// and end of character device coordinates `end`. `spaceWidth` is our best guess at the width of a
// space in the font the text is rendered in device coordinates.
func (to *textObject) newTextMark(text string, trm transform.Matrix, end transform.Point,
	spaceWidth float64, font *model.PdfFont, charspacing float64,
	fillColor, strokeColor color.Color) (textMark, bool) {
	theta := trm.Angle()
	orient := nearestMultiple(theta, orientationGranularity)
	var height float64
	if orient%180 != 90 {
		height = trm.ScalingFactorY()
	} else {
		height = trm.ScalingFactorX()
	}

	start := translation(trm)
	bbox := model.PdfRectangle{Llx: start.X, Lly: start.Y, Urx: end.X, Ury: end.Y}
	switch orient % 360 {
	case 90:
		bbox.Urx -= height
	case 180:
		bbox.Ury -= height
	case 270:
		bbox.Urx += height
	case 0:
		bbox.Ury += height
	default:
		// This is a hack to capture diagonal text.
		// TODO(peterwilliams97): Extract diagonal text.
		orient = 0
		bbox.Ury += height
	}
	if bbox.Llx > bbox.Urx {
		bbox.Llx, bbox.Urx = bbox.Urx, bbox.Llx
	}
	if bbox.Lly > bbox.Ury {
		bbox.Lly, bbox.Ury = bbox.Ury, bbox.Lly
	}

	clipped, onPage := rectIntersection(bbox, to.e.mediaBox)
	if !onPage {
		common.Log.Debug("Text mark outside page. bbox=%g mediaBox=%g text=%q",
			bbox, to.e.mediaBox, text)
	}
	bbox = clipped

	// The orientedBBox is bbox rotated and translated so the base of the character is at Lly.
	orientedBBox := bbox
	orientedMBox := to.e.mediaBox

	switch orient % 360 {
	case 90:
		orientedMBox.Urx, orientedMBox.Ury = orientedMBox.Ury, orientedMBox.Urx
		orientedBBox = model.PdfRectangle{
			Llx: orientedMBox.Urx - bbox.Ury,
			Urx: orientedMBox.Urx - bbox.Lly,
			Lly: bbox.Llx,
			Ury: bbox.Urx}
	case 180:
		orientedBBox = model.PdfRectangle{
			Llx: orientedMBox.Urx - bbox.Llx,
			Urx: orientedMBox.Urx - bbox.Urx,
			Lly: orientedMBox.Ury - bbox.Lly,
			Ury: orientedMBox.Ury - bbox.Ury}
	case 270:
		orientedMBox.Urx, orientedMBox.Ury = orientedMBox.Ury, orientedMBox.Urx
		orientedBBox = model.PdfRectangle{
			Llx: bbox.Ury,
			Urx: bbox.Lly,
			Lly: orientedMBox.Ury - bbox.Llx,
			Ury: orientedMBox.Ury - bbox.Urx}
	}
	if orientedBBox.Llx > orientedBBox.Urx {
		orientedBBox.Llx, orientedBBox.Urx = orientedBBox.Urx, orientedBBox.Llx
	}
	if orientedBBox.Lly > orientedBBox.Ury {
		orientedBBox.Lly, orientedBBox.Ury = orientedBBox.Ury, orientedBBox.Lly
	}

	tm := textMark{
		text:         text,
		PdfRectangle: orientedBBox,
		originaBBox:  bbox,
		font:         font,
		fontsize:     height,
		charspacing:  charspacing,
		trm:          trm,
		end:          end,
		orient:       orient,
		fillColor:    fillColor,
		strokeColor:  strokeColor,
	}
	if verboseGeom {
		common.Log.Info("newTextMark: start=%.2f end=%.2f %s", start, end, tm.String())
	}
	return tm, onPage
}

// String returns a description of `tm`.
func (tm *textMark) String() string {
	return fmt.Sprintf("%.2f fontsize=%.2f \"%s\"", tm.PdfRectangle, tm.fontsize, tm.text)
}

// bbox makes textMark implement the `bounded` interface.
func (tm *textMark) bbox() model.PdfRectangle {
	return tm.PdfRectangle
}

// ToTextMark returns the public view of `tm`.
func (tm *textMark) ToTextMark() TextMark {
	return TextMark{
		Text:        tm.text,
		Original:    tm.original,
		BBox:        tm.originaBBox,
		Font:        tm.font,
		FontSize:    tm.fontsize,
		FillColor:   tm.fillColor,
		StrokeColor: tm.strokeColor,
	}
}

// inDiacriticArea returns true if `diacritic` is in the area where it could be a diacritic of `tm`.
func (tm *textMark) inDiacriticArea(diacritic *textMark) bool {
	dLlx := tm.Llx - diacritic.Llx
	dUrx := tm.Urx - diacritic.Urx
	dLly := tm.Lly - diacritic.Lly
	return math.Abs(dLlx+dUrx) < tm.Width()*diacriticRadiusR &&
		math.Abs(dLly) < tm.Height()*diacriticRadiusR
}

// appendTextMark appends `mark` to `marks` and updates `offset`, the offset of `mark` in the extracted
// text.
func appendTextMark(marks []TextMark, offset *int, mark TextMark) []TextMark {
	mark.Offset = *offset
	marks = append(marks, mark)
	*offset += len(mark.Text)
	return marks
}

// appendSpaceMark appends a spaceMark with space character `space` to `marks` and updates `offset`,
// the offset of `mark` in the extracted text.
func appendSpaceMark(marks []TextMark, offset *int, spaceChar string) []TextMark {
	mark := spaceMark
	mark.Text = spaceChar
	return appendTextMark(marks, offset, mark)
}

// nearestMultiple return the integer multiple of `m` that is closest to `x`.
func nearestMultiple(x float64, m int) int {
	if m == 0 {
		m = 1
	}
	fac := float64(m)
	return int(math.Round(x/fac) * fac)
}
