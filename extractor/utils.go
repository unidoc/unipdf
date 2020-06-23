/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// RenderMode specifies the text rendering mode (Tmode), which determines whether showing text shall cause
// glyph outlines to be  stroked, filled, used as a clipping boundary, or some combination of the three.
// Stroking, filling, and clipping shall have the same effects for a text object as they do for a path object
// (see 8.5.3, "Path-Painting Operators" and 8.5.4, "Clipping Path Operators").
type RenderMode int

// Render mode type.
const (
	RenderModeStroke RenderMode = 1 << iota // Stroke
	RenderModeFill                          // Fill
	RenderModeClip                          // Clip
)

// toFloatXY returns `objs` as 2 floats, if that's what `objs` is, or an error if it isn't.
func toFloatXY(objs []core.PdfObject) (x, y float64, err error) {
	if len(objs) != 2 {
		return 0, 0, fmt.Errorf("invalid number of params: %d", len(objs))
	}
	floats, err := core.GetNumbersAsFloat(objs)
	if err != nil {
		return 0, 0, err
	}
	return floats[0], floats[1], nil
}

// minFloat returns the lesser of `a` and `b`.
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxFloat returns the greater of `a` and `b`.
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func procBuf(pt *PageText) {
	if isTesting {
		return
	}

	lk := license.GetLicenseKey()
	if lk != nil && lk.IsLicensed() {
		return
	}
	fmt.Printf("Unlicensed copy of unidoc\n")
	fmt.Printf("To get rid of the watermark and keep entire text - Please get a license on https://unidoc.io\n")

	var buf bytes.Buffer
	buf.WriteString(pt.viewText)

	s := "- [Unlicensed UniDoc - Get a license on https://unidoc.io]"
	if buf.Len() > 100 {
		s = "... [Truncated - Unlicensed UniDoc - Get a license on https://unidoc.io]"
		buf.Truncate(buf.Len() - 100)
	}
	buf.WriteString(s)
	pt.viewText = buf.String()

	if len(pt.marks) > 200 {
		pt.marks = pt.marks[:200]
	}
	if len(pt.viewMarks) > 200 {
		pt.viewMarks = pt.viewMarks[:200]
	}
}

// truncate returns the first `n` characters in string `s`.
func truncate(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// pdfColorToGoColor converts the specified color to a Go color, using the
// provided colorspace. If unsuccessful, color.Black is returned.
func pdfColorToGoColor(space model.PdfColorspace, c model.PdfColor) color.Color {
	if space == nil || c == nil {
		return color.Black
	}

	conv, err := space.ColorToRGB(c)
	if err != nil {
		common.Log.Debug("WARN: could not convert color %v (%v) to RGB: %s", space, c, err)
		return color.Black
	}
	rgb, ok := conv.(*model.PdfColorDeviceRGB)
	if !ok {
		common.Log.Debug("WARN: converted color is not in the RGB colorspace: %v", conv)
		return color.Black
	}

	return color.NRGBA{
		R: uint8(rgb.R() * 255),
		G: uint8(rgb.G() * 255),
		B: uint8(rgb.B() * 255),
		A: uint8(255),
	}
}
