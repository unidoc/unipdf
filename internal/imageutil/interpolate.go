package imageutil

import (
	"math"
)

// LinearInterpolate is the simple linear interpolation from the PDF manual.
func LinearInterpolate(x, xmin, xmax, ymin, ymax float64) float64 {
	if math.Abs(xmax-xmin) < 0.000001 {
		return ymin
	}

	y := ymin + (x-xmin)*(ymax-ymin)/(xmax-xmin)
	return y
}
