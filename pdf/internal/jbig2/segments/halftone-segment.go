/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"github.com/pkg/errors"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"math"
)

// HalftoneRegion is the model for the JBIG2 Halftone Region
type HalftoneRegion struct {
	r reader.StreamReader

	h *Header

	DataHeaderOffset int64
	DataHeaderLength int64
	DataOffset       int64
	DataLength       int64

	/** Region segment information field, 7.4.1 */
	RegionSegment *RegionSegment

	/** Halftone segment information field, 7.4.5.1.1 */
	HDefaultPixel       int
	CombinationOperator bitmap.CombinationOperator
	HSkipEnabled        bool
	HTemplate           byte
	IsMMREncoded        bool

	/** Halftone grid position and size, 7.4.5.1.2 */

	/** Width of the gray-scale image, 7.4.5.1.2.1 */
	HGridWidth int
	/** Height of the gray-scale image, 7.4.5.1.2.2 */
	HGridHeight int
	/** Horizontal offset of the grid, 7.4.5.1.2.3 */
	HGridX int
	/** Vertical offset of the grid, 7.4.5.1.2.4 */
	HGridY int

	/** Halftone grid vector, 7.4.5.1.3 */
	/** Horizontal coordinate of the halftone grid vector, 7.4.5.1.3.1 */
	HRegionX int
	/** Vertical coordinate of the halftone grod vector, 7.4.5.1.3.2 */
	HRegionY int

	/** Decoded data */
	HalftoneRegionBitmap *bitmap.Bitmap

	/**
	 * Previously decoded data from other regions or dictionaries, stored to use as patterns in this region.
	 */
	Patterns []*bitmap.Bitmap
}

func newHalftoneRegion(r *reader.Reader) *HalftoneRegion {
	hr := &HalftoneRegion{
		r:             r,
		RegionSegment: NewRegionSegment(r),
	}

	return hr
}

// Init initializes the HalfotoneRegion
func (h *HalftoneRegion) Init(hd *Header, r reader.StreamReader) error {
	h.r = r
	h.h = hd
	h.RegionSegment = NewRegionSegment(r)
	return h.parseHeader()
}

func (h *HalftoneRegion) parseHeader() error {
	if err := h.RegionSegment.parseHeader(); err != nil {
		return err
	}
	/* Bit 7 */

	b, err := h.r.ReadBit()
	if err != nil {
		return err
	}
	h.HDefaultPixel = b

	/* Bit 4-6 */

	temp, err := h.r.ReadBits(3)
	if err != nil {
		return err
	}

	h.CombinationOperator = bitmap.CombinationOperator(temp & 0xf)

	/* Bit 3 */
	b, err = h.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		h.HSkipEnabled = true
	}

	/* Bit 1 - 2 */
	temp, err = h.r.ReadBits(2)
	if err != nil {
		return err
	}

	h.HTemplate = byte(temp & 0xf)

	/* Bit 0 */
	b, err = h.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		h.IsMMREncoded = true
	}

	temp, err = h.r.ReadBits(32)
	if err != nil {
		return err
	}

	h.HGridWidth = int(temp) & 0xFFFFFFFF

	temp, err = h.r.ReadBits(32)
	if err != nil {
		return err
	}

	h.HGridHeight = int(temp) & 0xFFFFFFFF

	temp, err = h.r.ReadBits(32)
	if err != nil {
		return err
	}
	h.HGridX = int(temp)

	temp, err = h.r.ReadBits(32)
	if err != nil {
		return err
	}
	h.HGridY = int(temp)

	temp, err = h.r.ReadBits(16)
	if err != nil {
		return err
	}

	h.HRegionX = int(temp) & 0xFFFF

	temp, err = h.r.ReadBits(16)
	if err != nil {
		return err
	}

	h.HRegionY = int(temp) & 0xFFFF

	if err = h.computeSegmentDataStructure(); err != nil {
		return err
	}

	return h.checkInput()

}

func (h *HalftoneRegion) computeSegmentDataStructure() error {
	h.DataOffset = h.r.StreamPosition()
	h.DataHeaderLength = h.DataOffset - h.DataHeaderOffset
	h.DataLength = int64(h.r.Length()) - h.DataHeaderLength
	return nil
}

func (h *HalftoneRegion) checkInput() error {
	if h.IsMMREncoded {
		if h.HTemplate != 0 {
			common.Log.Info("HTemplate = %d should contain the value 0", h.HTemplate)
		}

		if h.HSkipEnabled {
			common.Log.Info("HSkipEnabled 0 %v (should contain the value false)", h.HSkipEnabled)
		}
	}
	return nil
}

// GetRegionBitmap - gets the Halftone Bitmap
// Implements Regioner method
// 6.6.5
func (h *HalftoneRegion) GetRegionBitmap() (bm *bitmap.Bitmap, err error) {

	if h.HalftoneRegionBitmap != nil {
		bm = h.HalftoneRegionBitmap
		return
	}

	// 6.6.5 page 40
	// 1)
	h.HalftoneRegionBitmap = bitmap.New(h.RegionSegment.BitmapWidth, h.RegionSegment.BitmapHeight)

	if h.Patterns == nil || len(h.Patterns) == 0 {
		h.Patterns, err = h.GetPatterns()
		if err != nil {
			return
		}
		common.Log.Trace("Taken patterns: %v", h.Patterns)
	}

	if h.HDefaultPixel == 1 {
		h.HalftoneRegionBitmap.SetDefaultPixel()
	}
	// 2)
	//6.6.5.1 Computing hSkip - At the moment SKIP is not used... we are not able to test it.

	// 3)
	bitsPerValueF := math.Ceil(math.Log(float64(len(h.Patterns))) / math.Log(2))

	common.Log.Trace("Patterns: %v", len(h.Patterns))
	common.Log.Trace("PatternsLog: %v", math.Log(float64(len(h.Patterns))))
	common.Log.Trace("Log:%v", math.Log(2))
	common.Log.Trace("BitsPerValue: %v", bitsPerValueF)
	bitsPerValue := int(bitsPerValueF)

	common.Log.Trace("BitsPerValue: %d", bitsPerValue)
	// 4)
	var grayScaleValues [][]int
	grayScaleValues, err = h.grayScaleDecoding(bitsPerValue)
	if err != nil {
		return
	}
	common.Log.Trace("Grayscale values: %v", grayScaleValues)
	if err = h.renderPattern(grayScaleValues); err != nil {
		return
	}

	return h.HalftoneRegionBitmap, nil
}

// GetRegionInfo implements Regioner interface method
func (h *HalftoneRegion) GetRegionInfo() *RegionSegment {
	return h.RegionSegment
}

// GetPatterns gets the HalftoneRegion patterns
func (h *HalftoneRegion) GetPatterns() (patterns []*bitmap.Bitmap, err error) {
	common.Log.Trace("RT Segments: %v", h.h.RTSegments)
	for _, s := range h.h.RTSegments {
		var data Segmenter
		data, err = s.GetSegmentData()
		if err != nil {
			common.Log.Trace("GetSegmentData failed: %v", err)
			return
		}

		common.Log.Trace("Data :%v", data)
		pattern, ok := data.(*PatternDictionary)
		if !ok {
			err = errors.Errorf("Related segment not a pattern dictionary: %T", data)
			return
		}
		var tempPatterns []*bitmap.Bitmap
		tempPatterns, err = pattern.GetDictionary()
		if err != nil {
			common.Log.Trace("GetDictionary failed: %v", err)
			return
		}

		patterns = append(patterns, tempPatterns...)
	}
	return
}

func (h *HalftoneRegion) grayScaleDecoding(bitsPerValue int) ([][]int, error) {
	var (
		gbAtX []int8
		gbAtY []int8
	)

	if !h.IsMMREncoded {
		gbAtX = make([]int8, 4)
		gbAtY = make([]int8, 4)

		if h.HTemplate <= 1 {
			gbAtX[0] = 3
		} else if h.HTemplate >= 2 {
			gbAtX[0] = 2
		}
		gbAtY[0] = -1

		gbAtX[1] = -3
		gbAtY[1] = -1

		gbAtX[2] = 2
		gbAtY[2] = -2

		gbAtX[3] = -2
		gbAtY[3] = -2
	}

	var grayScalePlanes []*bitmap.Bitmap = make([]*bitmap.Bitmap, bitsPerValue)

	// 1)
	common.Log.Trace("BeforeNewGenericSegment StreamPos: %d", h.r.StreamPosition())
	genericRegion := NewGenericRegion(h.r)
	genericRegion.setParametersMMR(h.IsMMREncoded, h.DataOffset, h.DataLength, h.HGridHeight, h.HGridWidth, h.HTemplate, false, h.HSkipEnabled, gbAtX, gbAtY)

	// 2)
	j := bitsPerValue - 1

	common.Log.Trace("J - bits per value: %d, %d", j, bitsPerValue)

	var err error
	grayScalePlanes[j], err = genericRegion.GetRegionBitmap()
	if err != nil {
		return nil, err
	}

	common.Log.Trace("First bm: %s", grayScalePlanes[j].String())

	for j > 0 {
		j--
		genericRegion.Bitmap = nil
		grayScalePlanes[j], err = genericRegion.GetRegionBitmap()
		if err != nil {
			return nil, err
		}

		common.Log.Trace("Before at j: %d, %s", j, grayScalePlanes[j].String())
		if err = h.combineGrayscalePlanes(grayScalePlanes, j); err != nil {
			return nil, err
		}

		common.Log.Trace("After at j: %d, %s", j, grayScalePlanes[j].String())
	}

	return h.computeGrayScalePlanes(grayScalePlanes, bitsPerValue)
}

func (h *HalftoneRegion) combineGrayscalePlanes(
	grayScalePlanes []*bitmap.Bitmap, j int,
) error {
	byteIndex := 0
	for y := 0; y < grayScalePlanes[j].Height; y++ {
		for x := 0; x < grayScalePlanes[j].Width; x += 8 {
			newValue, err := grayScalePlanes[j+1].GetByte(byteIndex)
			if err != nil {
				return err
			}
			oldValue, err := grayScalePlanes[j].GetByte(byteIndex)
			if err != nil {
				return err
			}

			err = grayScalePlanes[j].SetByte(byteIndex, bitmap.CombineBytes(oldValue, newValue, bitmap.CmbOpXor))
			if err != nil {
				return err
			}
			byteIndex++
		}
	}
	return nil
}

func (h *HalftoneRegion) computeGrayScalePlanes(
	grayScalePlanes []*bitmap.Bitmap, bitsPerValue int,
) ([][]int, error) {
	grayScaleValues := make([][]int, h.HGridHeight)
	for i := 0; i < len(grayScaleValues); i++ {
		grayScaleValues[i] = make([]int, h.HGridWidth)
	}

	common.Log.Trace("ComputGrayScale patterns init value: %v", grayScaleValues)
	common.Log.Trace("Base\n: %s", grayScalePlanes[0].String())
	for y := 0; y < h.HGridHeight; y++ {
		for x := 0; x < h.HGridWidth; x += 8 {
			var minorWidth int
			if d := h.HGridWidth - x; d > 8 {
				minorWidth = 8
			} else {
				minorWidth = d
			}

			byteIndex := grayScalePlanes[0].GetByteIndex(x, y)

			for minorX := 0; minorX < minorWidth; minorX++ {
				i := minorX + x

				grayScaleValues[y][i] = 0
				for j := 0; j < bitsPerValue; j++ {
					bv, err := grayScalePlanes[j].GetByte(byteIndex)
					if err != nil {
						return nil, err
					}
					shifted := (bv >> uint(7-i&7))
					and1 := shifted & 1
					multiplier := 1 << uint(j)
					v := int(and1) * multiplier
					common.Log.Trace("GrayScaleValue[%d][%d] Bv: %d, Shifter: %d, and1: %v, multiplier: %d, v: %d", y, i, bv, shifted, and1, multiplier, v)
					grayScaleValues[y][i] += v
				}
			}
		}
	}
	return grayScaleValues, nil
}

// renderPattern This method draws the pattern into the region bitmap ({@code htReg}), as
// described in 6.6.5.2, page 42
func (h *HalftoneRegion) renderPattern(grayScaleValues [][]int) (err error) {
	var x, y int
	common.Log.Trace("Rendering pattern begins")
	for m := 0; m < h.HGridHeight; m++ {
		for n := 0; n < h.HGridWidth; n++ {

			// common.Log.Trace("X: %d, Y: %d", x, y)
			x = h.computeX(m, n)
			y = h.computeY(m, n)

			common.Log.Trace("Getting pattern at: %d, %d", m, n)
			patternBitmap := h.Patterns[grayScaleValues[m][n]]

			if err = bitmap.Blit(
				patternBitmap, h.HalftoneRegionBitmap,
				x+h.HGridX, y+h.HGridY, h.CombinationOperator,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *HalftoneRegion) computeX(m, n int) int {
	return h.shiftAndFill(h.HGridX + m*h.HRegionY + n*h.HRegionX)
}

func (h *HalftoneRegion) computeY(m, n int) int {
	return h.shiftAndFill(h.HGridY + m*h.HRegionX - n*h.HRegionY)
}

func findMSB(n int) int {
	if n == 0 {
		return 0
	}
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return (n + 1) >> 1
}

func (h *HalftoneRegion) shiftAndFill(value int) int {
	value >>= 8
	if value < 0 {
		bitPosition := int(math.Log(float64(findMSB(value))) / math.Log(2))
		l := 31 - bitPosition
		for i := 1; i < l; i++ {
			value |= (1 << uint(31-i))
		}
	}
	return value
}
