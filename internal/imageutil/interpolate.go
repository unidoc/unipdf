/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package imageutil

import (
	"math"
)

// LinearInterpolate is the simple linear interpolation from the PDF manual - PDF 32000-1:2008 - 7.10.2.
func LinearInterpolate(x, xmin, xmax, ymin, ymax float64) float64 {
	if math.Abs(xmax-xmin) < 0.000001 {
		return ymin
	}

	y := ymin + (x-xmin)*(ymax-ymin)/(xmax-xmin)
	return y
}
