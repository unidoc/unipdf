package jbig2

import (
	"github.com/pkg/errors"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/mmr"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// GenericRegionSegment represents a generic region segment.<br>
// Parsing is done as described in 7.4.5.<br>
// Decoding procedure is done as described in 6.2.5.7 and 7.4.6.4.
type GenericRegionSegment struct {
	r *reader.Reader

	DataHeaderOffset int64
	DataHeaderLength int64
	DataOffset       int64
	DataLength       int64

	/** Region segment information field, 7.4.1 */
	RegionSegment *RegionSegment

	/** Generic region segment flags, 7.4.6.2 */
	UseExtTemplates bool
	IsTPGDon        bool
	GBTemplate      byte
	IsMMREncoded    bool

	inlineImage, unknownLength bool
	UseMMR                     bool

	/** Generic region segment AT flags, 7.4.6.3 */
	GBAtX        []int8
	GBAtY        []int8
	GBAtOverride []bool

	// override if true AT pixels are not on their normal location and have to be overwriten
	override bool

	// Decoded data as pixel values
	Bitmap *bitmap.Bitmap

	arithDecoder *arithmetic.Decoder
	cx           *arithmetic.DecoderStats
	mmrDecoder   *mmr.MmrDecoder
}

// NewGenericRegionSegment creates new GenericRegionSegment
func NewGenericRegionSegment(
	h *SegmentHeader,
	r *reader.Reader,
) (*GenericRegionSegment, error) {
	g := &GenericRegionSegment{
		RegionSegment: NewRegionSegment(r),
		r:             r,
	}
	if err := g.ParseHeader(); err != nil {
		return nil, err
	}

	return g, nil
}

// func (g *GenericRegionSegment) Init(h *SegmentHeader, r *reader.Reader) error {
// 	g.RegionSegment = NewRegionSegment(r)
// 	g.r = r
// 	return g.ParseHeader()
// }

func (g *GenericRegionSegment) GetRegionBitmap() (bm *bitmap.Bitmap, err error) {
	if g.Bitmap == nil {
		if g.IsMMREncoded {
			// MMR Decoder Call
			g.mmrDecoder, err = mmr.New(
				g.r,
				g.RegionSegment.BitmapWidth, g.RegionSegment.BitmapHeight,
				g.DataOffset, g.DataLength,
			)
			if err != nil {
				return
			}

			// Uncompress the bitmap
			g.Bitmap, err = g.mmrDecoder.UncompressMMR()
			if err != nil {
				return
			}
		} else {
			/*
			 * ARITHMETIC DECODER PROCEDURE for generic region segments
			 */

			if err = g.updateOverrideFlags(); err != nil {
				return
			}
			/* 6.2.5.7 - 1) */
			var ltp int
			if g.arithDecoder == nil {
				g.arithDecoder, err = arithmetic.New(g.r)
				if err != nil {
					return
				}
			}

			if g.cx == nil {
				g.cx = arithmetic.NewStats(65536, 1)
			}

			/* 6.2.5.7 - 2) */
			g.Bitmap = bitmap.New(g.RegionSegment.BitmapWidth, g.RegionSegment.BitmapHeight)

			paddedWidth := (g.Bitmap.Width + 7) & -8

			/* 6.2.5.7 - 3) */
			for line := 0; line < g.Bitmap.Height; line++ {
				common.Log.Debug("Decoding line: %d", line)
				common.Log.Debug("----------================================----------------")
				/* 6.2.5.7 - 3 c) */
				if g.IsTPGDon {
					var temp int
					temp, err = g.decodeSLTP()
					if err != nil {
						return
					}
					ltp ^= temp
				}

				/* 6.2.5.7 - 3 c) */
				if ltp == 1 {
					if line > 0 {
						if err = g.copyLineAbove(line); err != nil {
							return
						}
					}
				} else {
					common.Log.Debug("Debug Line: %d", line)
					if err = g.decodeLine(line, g.Bitmap.Width, paddedWidth); err != nil {
						return
					}
				}
			}
		}
	}

	return g.Bitmap, nil
}

func (g *GenericRegionSegment) ParseHeader() (err error) {
	if err = g.RegionSegment.ParseHeader(); err != nil {
		return err
	}

	/* Bit 5-7 */
	if _, err = g.r.ReadBits(3); err != nil {
		return err
	}

	/* Bit 4 */
	var b int
	b, err = g.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		g.UseExtTemplates = true
	}

	/* Bit 3 */
	b, err = g.r.ReadBit()
	if err != nil {
		return err
	}

	if b == 1 {
		g.IsTPGDon = true
	}

	/* Bit 1-2 */
	var bits uint64
	bits, err = g.r.ReadBits(2)
	if err != nil {
		return err
	}
	g.GBTemplate = byte(bits & 0xf)

	/* Bit 0 */
	b, err = g.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		g.IsMMREncoded = true
	}

	if !g.IsMMREncoded {
		var amountOfGbAt int
		if g.GBTemplate == 0 {
			if g.UseExtTemplates {
				amountOfGbAt = 12
			} else {
				amountOfGbAt = 4
			}
		} else {
			amountOfGbAt = 1
		}
		if err = g.readGBAtPixels(amountOfGbAt); err != nil {
			return err
		}
	}

	/* Segment data structure */
	if err = g.computeSegmentDataStructure(); err != nil {
		return err
	}

	return g.checkInput()
}

func (g *GenericRegionSegment) computeSegmentDataStructure() error {
	g.DataOffset = g.r.CurrentBytePosition()
	g.DataHeaderOffset = g.DataOffset - g.DataHeaderOffset
	g.DataLength = g.r.Size() - g.DataHeaderOffset
	return nil
}

func (g *GenericRegionSegment) copyLineAbove(line int) error {
	targetByteIndex := line * g.Bitmap.RowStride
	sourceByteIndex := targetByteIndex - g.Bitmap.RowStride
	for i := 0; i < g.Bitmap.RowStride; i++ {
		b, err := g.Bitmap.GetByte(sourceByteIndex)
		if err != nil {
			return err
		}
		if err = g.Bitmap.SetByte(targetByteIndex, b); err != nil {
			return err
		}

		targetByteIndex += 1
		sourceByteIndex += 1
	}
	return nil
}

func (g *GenericRegionSegment) checkInput() error {
	return nil
}

func (g *GenericRegionSegment) decodeSLTP() (int, error) {
	switch g.GBTemplate {
	case 0:
		g.cx.SetIndex(0x9B25)
	case 1:
		g.cx.SetIndex(0x795)
	case 2:
		g.cx.SetIndex(0xE5)
	case 3:
		g.cx.SetIndex(0x195)
	}

	return g.arithDecoder.DecodeBit(g.cx)
}

func (g *GenericRegionSegment) decodeLine(line, width, paddedWidth int) error {
	byteIndex := g.Bitmap.GetByteIndex(0, line)

	// idx originalyl byteIndex - rowStride
	idx := byteIndex - g.Bitmap.RowStride

	switch g.GBTemplate {
	case 0:
		if !g.UseExtTemplates {
			common.Log.Debug("DecodeTemplate0a")
			return g.decodeTemplate0a(line, width, paddedWidth, byteIndex, idx)
		} else {
			common.Log.Debug("DecodeTemplate0b")
			return g.decodeTemplate0b(line, width, paddedWidth, byteIndex, idx)
		}
	case 1:
		common.Log.Debug("DecodeTemplate1")
		return g.decodeTemplate1(line, width, paddedWidth, byteIndex, idx)
	case 2:
		common.Log.Debug("DecodeTemplate2")
		return g.decodeTemplate2(line, width, paddedWidth, byteIndex, idx)
	case 3:
		common.Log.Debug("DecodeTemplate3")
		return g.decodeTemplate3(line, width, paddedWidth, byteIndex, idx)
	}
	return errors.Errorf("Invalid GBTemplate provided: %d", g.GBTemplate)

	return nil
}

func (g *GenericRegionSegment) decodeTemplate0a(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	var (
		context, overriddenContext int
		line1, line2               int

		temp byte
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return err
		}

		line1 = int(temp)
	}

	if line >= 2 {
		temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride)
		if err != nil {
			return err
		}
		line2 = int(temp) << 6
	}
	common.Log.Debug("Line1: %d", line1)
	common.Log.Debug("Line2: %d", line2)

	context = (line1 & 0xf0) | (line2 & 0x3800)
	common.Log.Debug("Context: %d", context)

	var nextByte int

	for x := 0; x < paddedWidth; x = nextByte {
		/* 6.2.5.7 3d */

		var result byte
		nextByte = x + 8

		var minorWidth int
		if d := width - x; d > 8 {
			minorWidth = 8
		} else {
			minorWidth = d
		}

		if line > 0 {
			line1 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx + 1)
				if err != nil {
					return err
				}
				line1 |= int(temp)
			}
			common.Log.Debug("Line1 Set to: %b", line1)
		}

		if line > 1 {
			line2 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride + 1)
				if err != nil {
					return err
				}
				line2 |= (int(temp) << 6)
			}
			common.Log.Debug("Line2 Set to: %b", line2)
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			toShift := uint(7 - minorX)

			if g.override {
				overriddenContext = g.overrideAtTemplate0a(context, x+minorX, line, int(result), minorX, int(toShift))
				g.cx.SetIndex(overriddenContext)
			} else {
				common.Log.Debug("Setting context to %b", context)
				g.cx.SetIndex(context)
			}

			var bit int
			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return err
			}

			common.Log.Debug("Minor at: %d, bit: %d", minorX, bit)

			result |= byte(bit << uint(toShift))
			common.Log.Debug("Result: %b", result)
			common.Log.Debug("Shift to: %d, Line1: %d, Line2: %d", toShift, line1, line2)

			context = ((context & 0x7bf7) << 1) | bit | ((line1 >> toShift) & 0x10) | ((line2 >> toShift) & 0x800)
			common.Log.Debug("New context: %d", context)
		}
		common.Log.Debug("Result: %08b, %d", result, byteIndex)
		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return err
		}
		// byteIndex + 1
		byteIndex += 1
		idx += 1
	}

	return nil
}

func (g *GenericRegionSegment) decodeTemplate0b(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	var (
		context, overriddenContext int
		line1, line2               int

		temp byte
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return err
		}

		line1 = int(temp)
	}

	if line >= 2 {
		temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride)
		if err != nil {
			return err
		}
		line2 = int(temp) << 6
	}
	common.Log.Debug("Line1: %d", line1)
	common.Log.Debug("Line2: %d", line1)

	context = (line1 & 0xf0) | (line2 & 0x3800)
	common.Log.Debug("Context: %d", context)

	var nextByte byte

	for x := 0; x < paddedWidth; x = int(nextByte) {
		/* 6.2.5.7 3d */

		var result byte
		nextByte = byte(x) + 8

		var minorWidth int
		if d := width - x; d > 8 {
			minorWidth = 8
		} else {
			minorWidth = d
		}

		if line > 0 {
			line1 <<= 8

			if int(nextByte) < width {
				temp, err = g.Bitmap.GetByte(idx + 1)
				if err != nil {
					return err
				}
				line1 |= int(temp)
			}
		}

		if line > 1 {
			line2 <<= 8

			if int(nextByte) < width {
				temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride + 1)
				if err != nil {
					return err
				}
				line2 |= (int(temp) << 6)
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			toShift := uint(7 - minorX)

			if g.override {
				overriddenContext = g.overrideAtTemplate0a(context, x+minorX, line, int(result), minorX, int(toShift))
				g.cx.SetIndex(overriddenContext)
			} else {
				g.cx.SetIndex(context)
			}

			var bit int
			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return err
			}

			common.Log.Debug("Minor at: %d, bit: %d", minorX, bit)

			result |= byte(bit << uint(toShift))
			common.Log.Debug("Result: %08b", result)

			context = ((context & 0x7bf7) << 1) | bit | ((line1 >> toShift) & 0x10) | ((line2 >> toShift) & 0x800)
		}
		common.Log.Debug("Result: %08b, %d", result, byteIndex)
		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return err
		}
		// byteIndex + 1
		byteIndex += 1
		idx += 1
	}

	return nil
}

func (g *GenericRegionSegment) decodeTemplate1(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	var (
		context, overriddenContext int
		line1, line2               int

		temp byte
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return err
		}

		line1 = int(temp)
	}

	if line >= 2 {
		temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride)
		if err != nil {
			return err
		}
		line2 = int(temp) << 5
	}
	common.Log.Debug("Line1: %d", line1)
	common.Log.Debug("Line2: %d", line1)

	context = (line1 & 0x1f8) | (line2 & 0x1e00)
	common.Log.Debug("Context: %d", context)

	var nextByte byte

	for x := 0; x < paddedWidth; x = int(nextByte) {
		/* 6.2.5.7 3d */

		var result byte
		nextByte = byte(x) + 8

		var minorWidth int
		if d := width - x; d > 8 {
			minorWidth = 8
		} else {
			minorWidth = d
		}

		if line > 0 {
			line1 <<= 8

			if int(nextByte) < width {
				temp, err = g.Bitmap.GetByte(idx + 1)
				if err != nil {
					return err
				}
				line1 |= int(temp)
			}
		}

		if line > 1 {
			line2 <<= 8

			if int(nextByte) < width {
				temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride + 1)
				if err != nil {
					return err
				}
				line2 |= (int(temp) << 5)
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			toShift := uint(7 - minorX)

			if g.override {
				overriddenContext = g.overrideAtTemplate0a(context, x+minorX, line, int(result), minorX, int(toShift))
				g.cx.SetIndex(overriddenContext)
			} else {
				g.cx.SetIndex(context)
			}

			var bit int
			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return err
			}

			common.Log.Debug("Minor at: %d, bit: %d", minorX, bit)

			result |= byte(bit << uint(toShift))
			common.Log.Debug("Result: %08b", result)

			context = ((context & 0xefb) << 1) | bit | ((line1 >> toShift) & 0x8) | ((line2 >> toShift) & 0x200)
		}
		common.Log.Debug("Result: %08b, %d", result, byteIndex)
		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return err
		}
		// byteIndex + 1
		byteIndex += 1
		idx += 1
	}

	return nil
}

func (g *GenericRegionSegment) decodeTemplate2(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	var (
		context, overriddenContext int
		line1, line2               int

		temp byte
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return err
		}

		line1 = int(temp)
	}

	if line >= 2 {
		temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride)
		if err != nil {
			return err
		}
		line2 = int(temp) << 4
	}
	common.Log.Debug("Line1: %d", line1)
	common.Log.Debug("Line2: %d", line1)

	context = (line1 & 0x7c) | (line2 & 0x380)
	common.Log.Debug("Context: %d", context)

	var nextByte int

	for x := 0; x < paddedWidth; x = nextByte {
		/* 6.2.5.7 3d */

		var result byte
		nextByte = x + 8

		var minorWidth int
		if d := width - x; d > 8 {
			minorWidth = 8
		} else {
			minorWidth = d
		}

		if line > 0 {
			line1 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx + 1)
				if err != nil {
					return err
				}
				line1 |= int(temp)
			}
		}

		if line > 1 {
			line2 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride + 1)
				if err != nil {
					return err
				}
				line2 |= (int(temp) << 4)
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			toShift := uint(7 - minorX)

			if g.override {
				overriddenContext = g.overrideAtTemplate2(context, x+minorX, line, int(result), minorX)
				g.cx.SetIndex(overriddenContext)
			} else {
				g.cx.SetIndex(context)
			}

			var bit int
			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return err
			}

			common.Log.Debug("Minor at: %d, bit: %d", minorX, bit)

			result |= byte(bit << uint(toShift))
			common.Log.Debug("Result: %08b", result)

			context = ((context & 0x7bf7) << 1) | bit | ((line1 >> toShift) & 0x10) | ((line2 >> toShift) & 0x800)
		}
		common.Log.Debug("Result: %08b, %d", result, byteIndex)
		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return err
		}
		// byteIndex + 1
		byteIndex += 1
		idx += 1
	}

	return nil
}

func (g *GenericRegionSegment) decodeTemplate3(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	var (
		context, overriddenContext int
		line1                      int

		temp byte
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return err
		}

		line1 = int(temp)
	}

	common.Log.Debug("Line1: %d", line1)

	context = (line1 >> 1) & 0x70
	common.Log.Debug("Context: %d", context)

	var nextByte int

	for x := 0; x <= paddedWidth; x = nextByte {
		/* 6.2.5.7 3d */

		var result byte
		nextByte = x + 8

		var minorWidth int
		if d := width - x; d > 8 {
			minorWidth = 8
		} else {
			minorWidth = d
		}

		if line >= 1 {
			line1 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx + 1)
				if err != nil {
					return err
				}
				line1 |= int(temp)
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			toShift := uint(7 - minorX)

			if g.override {
				overriddenContext = g.overrideAtTemplate3(context, x+minorX, line, int(result), minorX)
				g.cx.SetIndex(overriddenContext)
			} else {
				g.cx.SetIndex(context)
			}

			var bit int
			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return err
			}

			common.Log.Debug("Minor at: %d, bit: %d", minorX, bit)

			result |= byte(bit << uint(toShift))
			common.Log.Debug("Result: %08b", result)

			context = ((context & 0x1f7) << 1) | bit | ((line1 >> uint(8-minorX)) & 0x010)
		}
		common.Log.Debug("Result: %08b, %d", result, byteIndex)
		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return err
		}
		// byteIndex + 1
		byteIndex += 1
		idx += 1
	}

	return nil
}

func (g *GenericRegionSegment) getPixel(x, y int) int8 {
	if x < 0 || x >= g.Bitmap.Width {
		return 0
	}
	if y < 0 || y >= g.Bitmap.Height {
		return 0
	}

	if g.Bitmap.GetPixel(x, y) {
		return 1
	}
	return 0
}

func (g *GenericRegionSegment) updateOverrideFlags() error {
	if g.GBAtX == nil || g.GBAtY == nil {
		return nil
	}

	if len(g.GBAtX) != len(g.GBAtY) {
		return errors.Errorf("Incosistent AT pixel. Amount of 'x' pixels: %d, Amount of 'y' pixels: %d", len(g.GBAtX), len(g.GBAtY))
	}

	g.GBAtOverride = make([]bool, len(g.GBAtX))

	switch g.GBTemplate {
	case 0:
		if !g.UseExtTemplates {
			if g.GBAtX[0] != 3 || g.GBAtY[0] != -1 {
				g.setOverrideFlag(0)
			}
			if g.GBAtX[1] != -3 || g.GBAtY[1] != -1 {
				g.setOverrideFlag(1)
			}
			if g.GBAtX[2] != 2 || g.GBAtY[2] != -2 {
				g.setOverrideFlag(2)
			}
			if g.GBAtX[3] != -2 || g.GBAtY[3] != -2 {
				g.setOverrideFlag(3)
			}
		} else {
			if g.GBAtX[0] != -2 || g.GBAtY[0] != 0 {
				g.setOverrideFlag(0)
			}
			if g.GBAtX[1] != 0 || g.GBAtY[1] != -2 {
				g.setOverrideFlag(1)
			}
			if g.GBAtX[2] != -2 || g.GBAtY[2] != -1 {
				g.setOverrideFlag(2)
			}
			if g.GBAtX[3] != -1 || g.GBAtY[3] != -2 {
				g.setOverrideFlag(3)
			}
			if g.GBAtX[4] != 1 || g.GBAtY[4] != -2 {
				g.setOverrideFlag(4)
			}
			if g.GBAtX[5] != 2 || g.GBAtY[5] != -1 {
				g.setOverrideFlag(5)
			}
			if g.GBAtX[6] != -3 || g.GBAtY[6] != 0 {
				g.setOverrideFlag(6)
			}
			if g.GBAtX[7] != -4 || g.GBAtY[7] != 0 {
				g.setOverrideFlag(7)
			}
			if g.GBAtX[8] != 2 || g.GBAtY[8] != -2 {
				g.setOverrideFlag(8)
			}
			if g.GBAtX[9] != 3 || g.GBAtY[9] != -1 {
				g.setOverrideFlag(9)
			}
			if g.GBAtX[10] != -2 || g.GBAtY[10] != -2 {
				g.setOverrideFlag(10)
			}
			if g.GBAtX[11] != -3 || g.GBAtY[11] != -1 {
				g.setOverrideFlag(11)
			}
		}
	case 1:
		if g.GBAtX[0] != 3 || g.GBAtY[0] != -1 {
			g.setOverrideFlag(0)
		}
	case 2:
		if g.GBAtX[0] != 2 || g.GBAtY[0] != -1 {
			g.setOverrideFlag(0)
		}
	case 3:
		if g.GBAtX[0] != 2 || g.GBAtY[0] != -1 {
			g.setOverrideFlag(0)
		}
	}
	return nil
}

func (g *GenericRegionSegment) overrideAtTemplate0a(ctx, x, y, result, minorX, toShift int) int {
	if g.GBAtOverride[0] {
		ctx &= 0xFFEF
		if g.GBAtY[0] == 0 && g.GBAtX[0] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[0]&0x1)) << 4
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[0]), y+int(g.GBAtY[0]))) << 4
		}
	}

	if g.GBAtOverride[1] {
		ctx &= 0xFBFF
		if g.GBAtY[1] == 0 && g.GBAtX[1] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[1]&0x1)) << 10
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[1]), y+int(g.GBAtY[1]))) << 10
		}
	}

	if g.GBAtOverride[2] {
		ctx &= 0xF7FF
		if g.GBAtY[2] == 0 && g.GBAtX[2] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[2]&0x1)) << 11
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[2]), y+int(g.GBAtY[2]))) << 11
		}
	}

	if g.GBAtOverride[3] {
		ctx &= 0x7FFF
		if g.GBAtY[3] == 0 && g.GBAtX[3] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[3]&0x1)) << 15
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[3]), y+int(g.GBAtY[3]))) << 15
		}
	}
	return ctx
}

func (g *GenericRegionSegment) overrideAtTemplate0b(ctx, x, y, result, minorX, toShift int) int {
	if g.GBAtOverride[0] {
		ctx &= 0xFFFD
		if g.GBAtY[0] == 0 && g.GBAtX[0] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[0]&0x1)) << 1
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[0]), y+int(g.GBAtY[0]))) << 1
		}
	}

	if g.GBAtOverride[1] {
		ctx &= 0xDFFF
		if g.GBAtY[1] == 0 && g.GBAtX[1] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[1]&0x1)) << 13
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[1]), y+int(g.GBAtY[1]))) << 13
		}
	}

	if g.GBAtOverride[2] {
		ctx &= 0xFDFF
		if g.GBAtY[2] == 0 && g.GBAtX[2] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[2]&0x1)) << 9
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[2]), y+int(g.GBAtY[2]))) << 9
		}
	}

	if g.GBAtOverride[3] {
		ctx &= 0xBFFF
		if g.GBAtY[3] == 0 && g.GBAtX[3] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[3]&0x1)) << 14
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[3]), y+int(g.GBAtY[3]))) << 14
		}
	}
	if g.GBAtOverride[4] {
		ctx &= 0xEFFF
		if g.GBAtY[4] == 0 && g.GBAtX[4] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[4]&0x1)) << 12
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[4]), y+int(g.GBAtY[4]))) << 12
		}
	}

	if g.GBAtOverride[5] {
		ctx &= 0xFFDF
		if g.GBAtY[5] == 0 && g.GBAtX[5] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[5]&0x1)) << 5
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[5]), y+int(g.GBAtY[5]))) << 5
		}
	}

	if g.GBAtOverride[6] {
		ctx &= 0xFFFB
		if g.GBAtY[6] == 0 && g.GBAtX[6] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[6]&0x1)) << 2
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[6]), y+int(g.GBAtY[6]))) << 2
		}
	}

	if g.GBAtOverride[7] {
		ctx &= 0xFFF7
		if g.GBAtY[7] == 0 && g.GBAtX[7] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[7]&0x1)) << 3
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[7]), y+int(g.GBAtY[7]))) << 3
		}
	}
	if g.GBAtOverride[8] {
		ctx &= 0xF7FF
		if g.GBAtY[8] == 0 && g.GBAtX[8] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[8]&0x1)) << 11
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[8]), y+int(g.GBAtY[8]))) << 11
		}
	}

	if g.GBAtOverride[9] {
		ctx &= 0xFFEF
		if g.GBAtY[9] == 0 && g.GBAtX[9] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[9]&0x1)) << 4
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[9]), y+int(g.GBAtY[9]))) << 4
		}
	}

	if g.GBAtOverride[10] {
		ctx &= 0x7FFF
		if g.GBAtY[10] == 0 && g.GBAtX[10] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[10]&0x1)) << 15
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[10]), y+int(g.GBAtY[10]))) << 15
		}
	}

	if g.GBAtOverride[11] {
		ctx &= 0xFDFF
		if g.GBAtY[11] == 0 && g.GBAtX[11] >= -int8(minorX) {
			ctx |= (result >> uint(int8(toShift)-g.GBAtX[11]&0x1)) << 10
		} else {
			ctx |= int(g.getPixel(x+int(g.GBAtX[11]), y+int(g.GBAtY[11]))) << 10
		}
	}

	return ctx
}

func (g *GenericRegionSegment) overrideAtTemplate1(ctx, x, y, result, minorX int) int {

	ctx &= 0x1FF7
	if g.GBAtY[0] == 0 && g.GBAtX[0] >= -int8(minorX) {
		ctx |= (result >> uint(7-(int8(minorX)+g.GBAtX[0]&0x1))) << 3
	} else {
		ctx |= int(g.getPixel(x+int(g.GBAtX[0]), y+int(g.GBAtY[0]))) << 3
	}

	return ctx
}

func (g *GenericRegionSegment) overrideAtTemplate2(ctx, x, y, result, minorX int) int {

	ctx &= 0x3FB
	if g.GBAtY[0] == 0 && g.GBAtX[0] >= -int8(minorX) {
		ctx |= (result >> uint(7-(int8(minorX)+g.GBAtX[0]&0x1))) << 2
	} else {
		ctx |= int(g.getPixel(x+int(g.GBAtX[0]), y+int(g.GBAtY[0]))) << 2
	}

	return ctx
}

func (g *GenericRegionSegment) overrideAtTemplate3(ctx, x, y, result, minorX int) int {

	ctx &= 0x3EF
	if g.GBAtY[0] == 0 && g.GBAtX[0] >= -int8(minorX) {
		ctx |= (result >> uint(7-(int8(minorX)+g.GBAtX[0]&0x1))) << 4
	} else {
		ctx |= int(g.getPixel(x+int(g.GBAtX[0]), y+int(g.GBAtY[0]))) << 4
	}

	return ctx
}

func (g *GenericRegionSegment) readGBAtPixels(amountOfGbAt int) error {
	g.GBAtX = make([]int8, amountOfGbAt)
	g.GBAtY = make([]int8, amountOfGbAt)

	for i := 0; i < amountOfGbAt; i++ {
		b, err := g.r.ReadByte()
		if err != nil {
			return err
		}

		g.GBAtX[i] = int8(b)

		b, err = g.r.ReadByte()
		if err != nil {
			return err
		}
		g.GBAtY[i] = int8(b)
	}

	return nil
}

func (g *GenericRegionSegment) setOverrideFlag(index int) {
	g.GBAtOverride[index] = true
	g.override = true
}

func (g *GenericRegionSegment) setParameters(
	isMMREncoded bool,
	dataOffset, dataLength int64,
	gbh, gbw int,
) {
	g.IsMMREncoded = isMMREncoded
	g.DataOffset = dataOffset
	g.DataLength = dataLength
	g.RegionSegment.BitmapHeight = gbh
	g.RegionSegment.BitmapWidth = gbw

	g.mmrDecoder = nil
	g.Bitmap = nil
}

func (g *GenericRegionSegment) setParametersWithAt(
	isMMREncoded bool,
	SDTemplate byte,
	isTPGDon, useSkip bool,
	sDAtX, sDAtY []int8,
	symWidth, hcHeight int,
	cx *arithmetic.DecoderStats, a *arithmetic.Decoder,
) {
	g.IsMMREncoded = isMMREncoded
	g.GBTemplate = SDTemplate
	g.IsTPGDon = isTPGDon
	g.GBAtX = sDAtX
	g.GBAtY = sDAtY
	g.RegionSegment.BitmapHeight = hcHeight
	g.RegionSegment.BitmapWidth = symWidth
	if cx != nil {
		g.cx = cx
	}
	if a != nil {
		g.arithDecoder = a
	}

	g.mmrDecoder = nil
	g.Bitmap = nil
}

// // Decode reads the segment for the provided reader.
// // Implements Segmenter interface
// func (g *GenericRegionSegment) Decode(r *reader.Reader) error {
// 	common.Log.Debug("[READ] Generic Region Segment started.")
// 	defer common.Log.Debug("[READ] Generic Region Segment finished.")

// 	common.Log.Debug("Decode RegionSegment")
// 	// Read the segment basics
// 	if err := g.RegionSegment.Decode(r); err != nil {
// 		common.Log.Debug("RegionSegment.Decode failed.")
// 		return err
// 	}

// 	common.Log.Debug("ReadGenericSegmentFlags")
// 	// Read generic region segment flags
// 	if err := g.readGenericRegionFlags(r); err != nil {
// 		common.Log.Debug("readGenericRegionFlags failed. %v", err)
// 		return err
// 	}

// 	g.UseMMR = g.GenericRegionFlags.GetValue(MMR) != 0
// 	var (
// 		template = g.GenericRegionFlags.GetValue(GBTemplate)

// 		AdaptiveTemplateX, AdaptiveTemplateY []int8 = make([]int8, 4), make([]int8, 4)
// 	)
// 	common.Log.Debug("UseMMR: %v, Template: %v", g.UseMMR, template)

// 	if !g.UseMMR {

// 		// set Adaptive Templates
// 		var err error
// 		var buf []byte = make([]byte, 2)
// 		if template == 0 {

// 			for i := 0; i < 4; i++ {
// 				if _, err := r.Read(buf); err != nil {
// 					common.Log.Debug("Reading AdaptiveTemplate at %d failed", i)
// 					return err
// 				}

// 				AdaptiveTemplateX[i] = int8(buf[0])
// 				AdaptiveTemplateY[i] = int8(buf[1])

// 			}

// 		} else {
// 			if _, err := r.Read(buf); err != nil {
// 				common.Log.Debug("Reading first AdaptiveTemplate failed")
// 				return err
// 			}
// 			AdaptiveTemplateX[0] = int8(buf[0])
// 			AdaptiveTemplateY[0] = int8(buf[1])
// 		}
// 		common.Log.Debug("AdaptiveTemplateX %v", AdaptiveTemplateX)
// 		common.Log.Debug("AdaptiveTemplateY %v", AdaptiveTemplateY)

// 		g.Decoders.Arithmetic.ResetGenericStats(template, nil)

// 		if err = g.Decoders.Arithmetic.Start(r); err != nil {
// 			common.Log.Debug("ArithmeticDecoder Start failed. %v", err)
// 			return err
// 		}

// 		common.Log.Debug("Read Adaptive Template X, and Adaptive Template Y finished.")
// 		common.Log.Debug("AdaptiveTemplateX: %v", AdaptiveTemplateX)
// 		common.Log.Debug("AdaptiveTemplateY: %v", AdaptiveTemplateY)
// 	}

// 	var (
// 		typicalPrediction bool = g.GenericRegionFlags.GetValue(TPGDOn) != 0
// 		length            int  = g.Header.DataLength
// 	)

// 	common.Log.Debug("Typical Prediction: %v", typicalPrediction)

// 	if length == -1 {
// 		// If lenght of data is unknown it needs to be determined through examination of the data
// 		// JBIG2 7.2.7

// 		common.Log.Debug("Length set to: %v", length)

// 		g.unknownLength = true

// 		var match1, match2 byte

// 		// Match1 And Match2 are the bytes to look for
// 		if g.UseMMR {
// 			match1 = 0x00
// 			match2 = 0x00
// 		} else {
// 			match1 = 0xFF
// 			match2 = 0xAC
// 		}

// 		var bytesRead int64

// 		for {
// 			//
// 			b1, err := r.ReadByte()
// 			if err != nil {
// 				common.Log.Debug("GenericSegment->Decode->ReadByte failed. %v", err)
// 				return err
// 			}

// 			bytesRead += 1

// 			// check if b1 matches match1
// 			// continue to read the
// 			if b1 == match1 {
// 				b2, err := r.ReadByte()
// 				if err != nil {
// 					common.Log.Debug("GenericSegment->Decode->ReadByte Byte2 failed. %v", err)
// 					return err
// 				}
// 				bytesRead += 1

// 				// if b2 matches match2
// 				if b2 == match2 {
// 					length = int(bytesRead) - 2
// 					break
// 				}
// 			}
// 		}

// 		common.Log.Debug("Bytes read: %d. Rewind", bytesRead)
// 		_, err := r.Seek(-bytesRead, io.SeekCurrent)
// 		if err != nil {
// 			log.Debug("GenericSegment->Decode->Seek() failed: %v", err)
// 			return err
// 		}
// 	}

// 	common.Log.Debug("Position: %02X", r.CurrentBytePosition())

// 	common.Log.Debug("Create new bitmap - w: %d, h: %d", g.BMHeight, g.BMWidth)
// 	bm := bitmap.New(g.BMWidth, g.BMHeight, g.Decoders)
// 	bm.Clear(false)

// 	var mmrDataLength int
// 	if !g.UseMMR {
// 		mmrDataLength = length - 18
// 	}

// 	common.Log.Debug("Reading Generic Bitmap with: MMR-%v,template-%v,typicalPredition:%v,ATX:%v, ATY:%v,mmrDL:%v", g.UseMMR, template, typicalPrediction, AdaptiveTemplateX, AdaptiveTemplateY, mmrDataLength)
// 	err := bm.Read(
// 		r,
// 		g.UseMMR,
// 		template,
// 		typicalPrediction,
// 		false,
// 		nil,
// 		AdaptiveTemplateX,
// 		AdaptiveTemplateY,
// 		mmrDataLength,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	common.Log.Debug("Generic Bitmap: \n%s", bm)

// 	if g.inlineImage {
// 		// If the segment is an inline image get the page info segment
// 		// and combine it's bitmap with this 'bitmap'
// 		pageSegmenter := g.Decoders.FindPageSegment(g.PageAssociation())
// 		if pageSegmenter != nil {
// 			// the page segmenter must be pageinformation.PISegment

// 			pageSegment := pageSegmenter.(*pageinformation.PageInformationSegment)

// 			// get page bitmap
// 			pageBM := pageSegment.PageBitmap

// 			extComboOp := g.RegionFlags.GetValue(regions.ExternalCombinationOperator)

// 			if pageSegment.PageBMHeight == -1 && g.BMYLocation+g.BMHeight > pageBM.Height {
// 				pageBM.Expand(g.BMYLocation+g.BMHeight, pageSegment.PageInfoFlags.GetValue(pageinformation.DefaultPixelValue))

// 			}

// 			common.Log.Debug("Combine Operation: %d", extComboOp)
// 			err := pageBM.Combine(bm, g.BMXLocation, g.BMYLocation, int64(extComboOp))
// 			if err != nil {
// 				common.Log.Debug("PageBitmap Combine failed: %v", err)
// 				return err
// 			}

// 		} else {
// 			common.Log.Debug("Page segment is nil.")
// 		}
// 	} else {
// 		// if not an inline image set the bitmap to the segment variable
// 		bm.BitmapNumber = g.Header.SegmentNumber
// 		g.Bitmap = bm
// 	}

// 	if g.unknownLength {
// 		if _, err := r.Seek(4, io.SeekCurrent); err != nil {
// 			common.Log.Debug("Seek 4 from current failed: %v", err)
// 			return err
// 		}

// 	}

// 	return nil
// }

// // GetBitmap implements bitmap.BitmapGetter interface
// func (g *GenericRegionSegment) GetBitmap() *bitmap.Bitmap {
// 	return g.Bitmap
// }

// // func (g *GenericRegionSegment) readGenericRegionFlags(r io.ByteReader) error {
// // 	GenericRegionFlags, err := r.ReadByte()
// // 	if err != nil {
// // 		return err
// // 	}

// // 	g.GenericRegionFlags.SetValue(int(GenericRegionFlags))

// // 	common.Log.Debug("Generic Region Segment Flags: %d", GenericRegionFlags)
// // 	return nil
// // }
