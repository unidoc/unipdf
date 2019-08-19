/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"fmt"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// HalftoneRegion is the model for the jbig2 halftone region segment implementation - 7.4.5.1.
type HalftoneRegion struct {
	r                reader.StreamReader
	h                *Header
	DataHeaderOffset int64
	DataHeaderLength int64
	DataOffset       int64
	DataLength       int64

	// Region segment information field, 7.4.1.
	RegionSegment *RegionSegment

	// Halftone segment information field, 7.4.5.1.1.
	HDefaultPixel       int
	CombinationOperator bitmap.CombinationOperator
	HSkipEnabled        bool
	HTemplate           byte
	IsMMREncoded        bool

	// Halftone grid position and size, 7.4.5.1.2
	// Width of the gray-scale image, 7.4.5.1.2.1
	HGridWidth uint32
	// Height of the gray-scale image, 7.4.5.1.2.2
	HGridHeight uint32
	// Horizontal offset of the grid, 7.4.5.1.2.3
	HGridX int32
	// Vertical offset of the grid, 7.4.5.1.2.4
	HGridY int32

	// Halftone grid vector, 7.4.5.1.3
	// Horizontal coordinate of the halftone grid vector, 7.4.5.1.3.1
	HRegionX uint16
	// Vertical coordinate of the halftone grod vector, 7.4.5.1.3.2
	HRegionY uint16

	// Decoded data
	HalftoneRegionBitmap *bitmap.Bitmap

	// Previously decoded data from other regions or dictionaries, stored to use as patterns in this region.
	Patterns []*bitmap.Bitmap
}

// Init implements Segmenter interface.
func (h *HalftoneRegion) Init(hd *Header, r reader.StreamReader) error {
	h.r = r
	h.h = hd
	h.RegionSegment = NewRegionSegment(r)
	return h.parseHeader()
}

// GetRegionBitmap implements Regioner interface.
func (h *HalftoneRegion) GetRegionBitmap() (*bitmap.Bitmap, error) {
	if h.HalftoneRegionBitmap != nil {
		return h.HalftoneRegionBitmap, nil
	}
	var err error

	// 6.6.5 1)
	h.HalftoneRegionBitmap = bitmap.New(h.RegionSegment.BitmapWidth, h.RegionSegment.BitmapHeight)

	if h.Patterns == nil || len(h.Patterns) == 0 {
		h.Patterns, err = h.GetPatterns()
		if err != nil {
			return nil, err
		}
	}

	if h.HDefaultPixel == 1 {
		h.HalftoneRegionBitmap.SetDefaultPixel()
	}

	// 3)
	bitsPerValueF := math.Ceil(math.Log(float64(len(h.Patterns))) / math.Log(2))
	bitsPerValue := int(bitsPerValueF)

	// 4)
	var grayScaleValues [][]int32
	grayScaleValues, err = h.grayScaleDecoding(bitsPerValue)
	if err != nil {
		return nil, err
	}

	if err = h.renderPattern(grayScaleValues); err != nil {
		return nil, err
	}

	return h.HalftoneRegionBitmap, nil
}

// GetRegionInfo implements Regioner interface.
func (h *HalftoneRegion) GetRegionInfo() *RegionSegment {
	return h.RegionSegment
}

// GetPatterns gets the HalftoneRegion patterns.
func (h *HalftoneRegion) GetPatterns() ([]*bitmap.Bitmap, error) {
	var (
		patterns []*bitmap.Bitmap
		err      error
	)

	for _, s := range h.h.RTSegments {
		var data Segmenter
		data, err = s.GetSegmentData()
		if err != nil {
			common.Log.Debug("GetSegmentData failed: %v", err)
			return nil, err
		}

		pattern, ok := data.(*PatternDictionary)
		if !ok {
			err = fmt.Errorf("related segment not a pattern dictionary: %T", data)
			return nil, err
		}

		var tempPatterns []*bitmap.Bitmap
		tempPatterns, err = pattern.GetDictionary()
		if err != nil {
			common.Log.Debug("pattern GetDictionary failed: %v", err)
			return nil, err
		}

		patterns = append(patterns, tempPatterns...)
	}
	return patterns, nil
}

func (h *HalftoneRegion) checkInput() error {
	if h.IsMMREncoded {
		if h.HTemplate != 0 {
			common.Log.Debug("HTemplate = %d should contain the value 0", h.HTemplate)
		}

		if h.HSkipEnabled {
			common.Log.Debug("HSkipEnabled 0 %v (should contain the value false)", h.HSkipEnabled)
		}
	}
	return nil
}

func (h *HalftoneRegion) combineGrayscalePlanes(grayScalePlanes []*bitmap.Bitmap, j int) error {
	byteIndex := int32(0)

	for y := int32(0); y < grayScalePlanes[j].Height; y++ {
		for x := int32(0); x < grayScalePlanes[j].Width; x += 8 {
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

func (h *HalftoneRegion) computeGrayScalePlanes(grayScalePlanes []*bitmap.Bitmap, bitsPerValue int) ([][]int32, error) {
	grayScaleValues := make([][]int32, h.HGridHeight)

	for i := 0; i < len(grayScaleValues); i++ {
		grayScaleValues[i] = make([]int32, h.HGridWidth)
	}

	for y := uint32(0); y < h.HGridHeight; y++ {
		for x := uint32(0); x < h.HGridWidth; x += 8 {
			var minorWidth uint32

			if d := h.HGridWidth - x; d > 8 {
				minorWidth = 8
			} else {
				minorWidth = d
			}

			byteIndex := grayScalePlanes[0].GetByteIndex(int32(x), int32(y))

			for minorX := uint32(0); minorX < minorWidth; minorX++ {
				i := minorX + x
				grayScaleValues[y][i] = 0

				for j := 0; j < bitsPerValue; j++ {
					bv, err := grayScalePlanes[j].GetByte(byteIndex)
					if err != nil {
						return nil, err
					}
					shifted := (bv >> uint(7-i&7))
					and1 := shifted & 1
					multiplier := int32(1) << uint(j)
					v := int32(and1) * multiplier
					grayScaleValues[y][i] += v
				}
			}
		}
	}
	return grayScaleValues, nil
}

func (h *HalftoneRegion) computeSegmentDataStructure() error {
	h.DataOffset = h.r.StreamPosition()
	h.DataHeaderLength = h.DataOffset - h.DataHeaderOffset
	h.DataLength = int64(h.r.Length()) - h.DataHeaderLength
	return nil
}

func (h *HalftoneRegion) computeX(m, n uint32) uint32 {
	return h.shiftAndFill(uint32(h.HGridX) + m*uint32(h.HRegionY) + n*uint32(h.HRegionX))
}

func (h *HalftoneRegion) computeY(m, n uint32) uint32 {
	return h.shiftAndFill(uint32(h.HGridY) + m*uint32(h.HRegionX) - n*uint32(h.HRegionY))
}

func (h *HalftoneRegion) grayScaleDecoding(bitsPerValue int) ([][]int32, error) {
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

	grayScalePlanes := make([]*bitmap.Bitmap, bitsPerValue)

	// 1)
	genericRegion := NewGenericRegion(h.r)
	genericRegion.setParametersMMR(h.IsMMREncoded, h.DataOffset, h.DataLength, int32(h.HGridHeight), int32(h.HGridWidth), h.HTemplate, false, h.HSkipEnabled, gbAtX, gbAtY)

	// 2)
	j := bitsPerValue - 1

	var err error
	grayScalePlanes[j], err = genericRegion.GetRegionBitmap()
	if err != nil {
		return nil, err
	}

	for j > 0 {
		j--
		genericRegion.Bitmap = nil
		grayScalePlanes[j], err = genericRegion.GetRegionBitmap()
		if err != nil {
			return nil, err
		}

		if err = h.combineGrayscalePlanes(grayScalePlanes, j); err != nil {
			return nil, err
		}
	}
	return h.computeGrayScalePlanes(grayScalePlanes, bitsPerValue)
}

func (h *HalftoneRegion) parseHeader() error {
	if err := h.RegionSegment.parseHeader(); err != nil {
		return err
	}

	// Bit 7
	b, err := h.r.ReadBit()
	if err != nil {
		return err
	}
	h.HDefaultPixel = b

	// Bit 4-6
	temp, err := h.r.ReadBits(3)
	if err != nil {
		return err
	}

	h.CombinationOperator = bitmap.CombinationOperator(temp & 0xf)

	// Bit 3
	b, err = h.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		h.HSkipEnabled = true
	}

	// Bit 1 - 2
	temp, err = h.r.ReadBits(2)
	if err != nil {
		return err
	}
	h.HTemplate = byte(temp & 0xf)

	// Bit 0
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
	h.HGridWidth = uint32(temp) & 0xFFFFFFFF

	temp, err = h.r.ReadBits(32)
	if err != nil {
		return err
	}
	h.HGridHeight = uint32(temp) & 0xFFFFFFFF

	temp, err = h.r.ReadBits(32)
	if err != nil {
		return err
	}
	h.HGridX = int32(temp)

	temp, err = h.r.ReadBits(32)
	if err != nil {
		return err
	}
	h.HGridY = int32(temp)

	temp, err = h.r.ReadBits(16)
	if err != nil {
		return err
	}
	h.HRegionX = uint16(temp) & 0xFFFF

	temp, err = h.r.ReadBits(16)
	if err != nil {
		return err
	}
	h.HRegionY = uint16(temp) & 0xFFFF

	if err = h.computeSegmentDataStructure(); err != nil {
		return err
	}
	return h.checkInput()
}

// renderPattern draws the pattern into the region bitmap, as described in 6.6.5.2.
func (h *HalftoneRegion) renderPattern(grayScaleValues [][]int32) (err error) {
	var x, y uint32

	for m := uint32(0); m < h.HGridHeight; m++ {
		for n := uint32(0); n < h.HGridWidth; n++ {
			x = h.computeX(m, n)
			y = h.computeY(m, n)
			patternBitmap := h.Patterns[grayScaleValues[m][n]]

			if err = bitmap.Blit(
				patternBitmap, h.HalftoneRegionBitmap,
				int32(x)+h.HGridX, int32(y)+h.HGridY, h.CombinationOperator,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func newHalftoneRegion(r *reader.Reader) *HalftoneRegion {
	return &HalftoneRegion{r: r, RegionSegment: NewRegionSegment(r)}
}

func findMSB(n uint32) uint32 {
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

func (h *HalftoneRegion) shiftAndFill(value uint32) uint32 {
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
