package imageutil

import (
	"image"
	"image/color"
	"math"
)

// AutoThresholdTriangle is a function that returns threshold value on the base of provided histogram.
// It is calculated using triangle method.
func AutoThresholdTriangle(histogram [256]int) uint8 {
	var minV, dMax, maxV, min2 int
	// find the minV and the maxV of the histogram
	for i := 0; i < len(histogram); i++ {
		if histogram[i] > 0 {
			minV = i
			break
		}
	}
	if minV > 0 {
		minV--
	}
	for i := 255; i > 0; i-- {
		if histogram[i] > 0 {
			min2 = i
			break
		}
	}
	if min2 < 255 {
		min2++
	}

	for i := 0; i < 256; i++ {
		if histogram[i] > dMax {
			maxV = i
			dMax = histogram[i]
		}
	}

	// find which is the furthest side
	var inverted bool
	if (maxV - minV) < (min2 - maxV) {
		// reverse the histogram
		inverted = true
		var left int
		right := 255
		for left < right {
			temp := histogram[left]
			histogram[left] = histogram[right]
			histogram[right] = temp
			left++
			right--
		}
		minV = 255 - min2
		maxV = 255 - maxV
	}

	if minV == maxV {
		return uint8(minV)
	}
	// nx is the maxV frequency
	nx := float64(histogram[maxV])
	ny := float64(minV - maxV)
	d := math.Sqrt(nx*nx + ny*ny)
	nx /= d
	ny /= d
	d = nx*float64(minV) + ny*float64(histogram[minV])

	// find the split point
	split := minV
	var splitDistance float64
	for i := minV + 1; i <= maxV; i++ {
		newDistance := nx*float64(i) + ny*float64(histogram[i]) - d
		if newDistance > splitDistance {
			split = i
			splitDistance = newDistance
		}
	}
	split--
	if inverted {
		var left int
		right := 255
		for left < right {
			temp := histogram[left]
			histogram[left] = histogram[right]
			histogram[right] = temp
			left++
			right--
		}
		return uint8(255 - split)
	}
	return uint8(split)
}

// GrayImageHistogram gets histogram for the provided Gray 'img'.
func GrayImageHistogram(img *image.Gray) (histogram [256]int) {
	for _, pix := range img.Pix {
		histogram[pix]++
	}
	return histogram
}

// ImgToBinary gets the binary (black/white) image from the given image 'i' and provided threshold 'threshold'
func ImgToBinary(i image.Image, threshold uint8) *image.Gray {
	switch img := i.(type) {
	case *image.Gray:
		if isGrayBlackWhite(img) {
			return img
		}
		return grayImageToBlackWhite(img, threshold)
	case *image.Gray16:
		return gray16ImageToBlackWhite(img, threshold)
	default:
		return rgbImageToBlackWhite(img, threshold)
	}
}

// ImgToGray gets the gray-scaled image from the given 'i' image.
func ImgToGray(i image.Image) *image.Gray {
	if g, ok := i.(*image.Gray); ok {
		return g
	}
	bounds := i.Bounds()
	g := image.NewGray(bounds)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := i.At(x, y)
			g.Set(x, y, c)
		}
	}
	return g
}

// IsGrayImgBlackAndWhite checks if provided gray image is BlackAndWhite - Binary image.
func IsGrayImgBlackAndWhite(i *image.Gray) bool {
	return isGrayBlackWhite(i)
}

func blackOrWhite(c, threshold uint8) uint8 {
	if c < threshold {
		return 255
	}
	return 0
}

func gray16ImageToBlackWhite(img *image.Gray16, th uint8) *image.Gray {
	bounds := img.Bounds()
	d := image.NewGray(bounds)
	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			pix := img.Gray16At(x, y)
			d.SetGray(x, y, color.Gray{Y: blackOrWhite(uint8(pix.Y/256), th)})
		}
	}
	return d
}

// grayImageToBlackWhite gets black and white image on the base of provided
// Gray 'img' and a threshold 'th'.
func grayImageToBlackWhite(img *image.Gray, th uint8) *image.Gray {
	bounds := img.Bounds()
	d := image.NewGray(bounds)
	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			c := img.GrayAt(x, y)
			d.SetGray(x, y, color.Gray{Y: blackOrWhite(c.Y, th)})
		}
	}
	return d
}

func isGrayBlackWhite(img *image.Gray) bool {
	for i := 0; i < len(img.Pix); i++ {
		if !isPix8BlackWhite(img.Pix[i]) {
			return false
		}
	}
	return true
}

func isPix8BlackWhite(pix uint8) bool {
	if pix == 0 || pix == 255 {
		return true
	}
	return false
}

func rgbImageToBlackWhite(i image.Image, th uint8) *image.Gray {
	bounds := i.Bounds()
	gray := image.NewGray(bounds)
	var (
		c  color.Color
		cg color.Gray
	)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			// get the color at x,y
			c = i.At(x, y)
			// set it to the grayscale
			gray.Set(x, y, c)
			// get the grayscale color value
			cg = gray.GrayAt(x, y)
			// set the black/white pixel at 'x', 'y'
			gray.SetGray(x, y, color.Gray{Y: blackOrWhite(cg.Y, th)})
		}
	}
	return gray
}
