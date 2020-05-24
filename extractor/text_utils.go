/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"

	"github.com/unidoc/unipdf/v3/model"
)

// TOL is the tolerance for coordinates to be consideted equal. It is big enough to cover all
// rounding errors and small enough that TOL point differences on a page aren't visible.
const TOL = 1.0e-6

// isZero returns true if x is with TOL of 0.0
func isZero(x float64) bool {
	return math.Abs(x) < TOL
}

// rectUnion returns the smallest axis-aligned rectangle that contains `b1` and `b2`.
func rectUnion(b1, b2 model.PdfRectangle) model.PdfRectangle {
	return model.PdfRectangle{
		Llx: math.Min(b1.Llx, b2.Llx),
		Lly: math.Min(b1.Lly, b2.Lly),
		Urx: math.Max(b1.Urx, b2.Urx),
		Ury: math.Max(b1.Ury, b2.Ury),
	}
}

// rectIntersection returns the largest axis-aligned rectangle that is contained by `b1` and `b2`.
func rectIntersection(b1, b2 model.PdfRectangle) (model.PdfRectangle, bool) {
	if !intersects(b1, b2) {
		return model.PdfRectangle{}, false
	}
	return model.PdfRectangle{
		Llx: math.Max(b1.Llx, b2.Llx),
		Urx: math.Min(b1.Urx, b2.Urx),
		Lly: math.Max(b1.Lly, b2.Lly),
		Ury: math.Min(b1.Ury, b2.Ury),
	}, true
}

// intersects returns true if `r0` and `r1` overlap in the x and y axes.
func intersects(b1, b2 model.PdfRectangle) bool {
	return intersectsX(b1, b2) && intersectsY(b1, b2)
}

// intersectsX returns true if `r0` and `r1` overlap in the x axis.
func intersectsX(b1, b2 model.PdfRectangle) bool {
	return b1.Llx <= b2.Urx && b2.Llx <= b1.Urx
}

// intersectsY returns true if `r0` and `r1` overlap in the y axis.
func intersectsY(b1, b2 model.PdfRectangle) bool {
	return b1.Lly <= b2.Ury && b2.Lly <= b1.Ury
}

func fileLine(skip int, doSecond bool) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}
	depth := fmt.Sprintf("%s:%-4d", file, line)
	if !doSecond {
		return depth
	}
	_, _, line2, _ := runtime.Caller(skip + 2)
	return fmt.Sprintf("%s:%-4d", depth, line2)
}
