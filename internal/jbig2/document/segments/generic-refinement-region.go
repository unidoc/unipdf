/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// GenericRefinementRegion represtents jbig2 generic refinement region segment - 7.4.7.
type GenericRefinementRegion struct {
	t0 templater
	t1 templater

	r reader.StreamReader
	h *Header

	// Region segment information flags 7.4.1.
	RegionInfo *RegionSegment

	// Generic refinement region segment flags 7.4.7.2.
	IsTPGROn   bool
	TemplateID int8

	Template templater

	// Generic refinement region segment AT flags 7.4.7.3.
	GrAtX []int8
	GrAtY []int8

	// Decoded data as pixel values (use row stride/width to wrap line).
	RegionBitmap *bitmap.Bitmap

	// Variables for decoding.
	ReferenceBitmap *bitmap.Bitmap
	ReferenceDX     int32
	ReferenceDY     int32

	arithDecode *arithmetic.Decoder
	cx          *arithmetic.DecoderStats

	override     bool
	grAtOverride []bool
}

// Init implements Segmenter interface.
func (g *GenericRefinementRegion) Init(header *Header, r reader.StreamReader) error {
	g.h = header
	g.r = r
	g.RegionInfo = NewRegionSegment(r)
	return g.parseHeader()
}

// GetRegionBitmap implements Regioner interface.
func (g *GenericRefinementRegion) GetRegionBitmap() (*bitmap.Bitmap, error) {
	var err error
	common.Log.Trace("[GENERIC-REF-REGION] GetRegionBitmap begins...")
	defer func() {
		if err != nil {
			common.Log.Trace("[GENERIC-REF-REGION] GetRegionBitmap failed. %v", err)
		} else {
			common.Log.Trace("[GENERIC-REF-REGION] GetRegionBitmap finished.")
		}
	}()

	if g.RegionBitmap != nil {
		return g.RegionBitmap, nil
	}

	// 6.3.5.6 - 1)
	isLineTypicalPredicted := 0
	if g.ReferenceBitmap == nil {
		// Get the reference bitmap, which is the base of refinement process
		g.ReferenceBitmap, err = g.getGrReference()
		if err != nil {
			return nil, err
		}
	}

	if g.arithDecode == nil {
		g.arithDecode, err = arithmetic.New(g.r)
		if err != nil {
			return nil, err
		}
	}

	if g.cx == nil {
		g.cx = arithmetic.NewStats(8192, 1)
	}

	g.RegionBitmap = bitmap.New(int(g.RegionInfo.BitmapWidth), int(g.RegionInfo.BitmapHeight))

	if g.TemplateID == 0 {
		if err = g.updateOverride(); err != nil {
			return nil, err
		}
	}

	paddedWidth := (g.RegionBitmap.Width + 7) & -8
	var deltaRefStride int

	if g.IsTPGROn {
		deltaRefStride = int(-g.ReferenceDY) * g.ReferenceBitmap.RowStride
	}

	yOffset := deltaRefStride + 1

	// 6.3.5.6 - 3
	for y := 0; y < g.RegionBitmap.Height; y++ {
		// 6.3.5.6 - 3 b)
		if g.IsTPGROn {
			temp, err := g.decodeSLTP()
			if err != nil {
				return nil, err
			}
			isLineTypicalPredicted ^= temp
		}

		if isLineTypicalPredicted == 0 {
			// 6.3.5.6 - 3 c)
			err = g.decodeOptimized(y, g.RegionBitmap.Width, g.RegionBitmap.RowStride, g.ReferenceBitmap.RowStride, paddedWidth, deltaRefStride, yOffset)
			if err != nil {
				return nil, err
			}
		} else {
			// 6.3.5.6 - 3 d)
			err = g.decodeTypicalPredictedLine(y, g.RegionBitmap.Width, g.RegionBitmap.RowStride, g.ReferenceBitmap.RowStride, paddedWidth, deltaRefStride)
			if err != nil {
				return nil, err
			}
		}
	}

	// 6.3.5.6 - 4)
	return g.RegionBitmap, nil
}

// GetRegionInfo implements Regioner interface.
func (g *GenericRefinementRegion) GetRegionInfo() *RegionSegment {
	return g.RegionInfo
}

func (g *GenericRefinementRegion) decodeSLTP() (int, error) {
	g.Template.setIndex(g.cx)
	return g.arithDecode.DecodeBit(g.cx)
}

func (g *GenericRefinementRegion) getGrReference() (*bitmap.Bitmap, error) {
	segments := g.h.RTSegments
	if len(segments) == 0 {
		return nil, errors.New("Referenced Segment not exists")
	}

	s, err := segments[0].GetSegmentData()
	if err != nil {
		return nil, err
	}

	r, ok := s.(Regioner)
	if !ok {
		return nil, fmt.Errorf("referred to Segment is not a Regioner: %T", s)
	}
	return r.GetRegionBitmap()
}

func (g *GenericRefinementRegion) decodeOptimized(
	lineNumber, width, rowStride, refRowStride, paddedWidth, deltaRefStride, lineOffset int,
) error {
	var (
		err       error
		rx        int
		tempIndex int
	)
	currentLine := lineNumber - int(g.ReferenceDY)
	if t := int(-g.ReferenceDX); t > 0 {
		rx = t
	}
	refByteIndex := g.ReferenceBitmap.GetByteIndex(rx, currentLine)

	if g.ReferenceDX > 0 {
		tempIndex = int(g.ReferenceDX)
	}

	byteIndex := g.RegionBitmap.GetByteIndex(tempIndex, lineNumber)

	switch g.TemplateID {
	case 0:
		err = g.decodeTemplate(lineNumber, width, rowStride, refRowStride, paddedWidth, deltaRefStride, lineOffset, byteIndex, currentLine, refByteIndex, g.t0)
	case 1:
		err = g.decodeTemplate(lineNumber, width, rowStride, refRowStride, paddedWidth, deltaRefStride, lineOffset, byteIndex, currentLine, refByteIndex, g.t1)
	}
	return err
}

func (g *GenericRefinementRegion) decodeTypicalPredictedLine(
	lineNumber, width, rowStride, refRowStride, paddedWidth, deltaRefStride int,
) error {
	// Offset of the reference bitmap with respect to the bitmap being decoded
	currentLine := lineNumber - int(g.ReferenceDY)
	refByteIndex := g.ReferenceBitmap.GetByteIndex(0, currentLine)
	byteIndex := g.RegionBitmap.GetByteIndex(0, lineNumber)
	var err error

	switch g.TemplateID {
	case 0:
		err = g.decodeTypicalPredictedLineTemplate0(lineNumber, width, rowStride, refRowStride, paddedWidth, deltaRefStride, byteIndex, currentLine, refByteIndex)
	case 1:
		err = g.decodeTypicalPredictedLineTemplate1(lineNumber, width, rowStride, refRowStride, paddedWidth, deltaRefStride, byteIndex, currentLine, refByteIndex)
	}
	return err
}

func (g *GenericRefinementRegion) decodeTypicalPredictedLineTemplate0(
	lineNumber, width, rowStride, refRowStride, paddedWidth,
	deltaRefStride, byteIndex, currentLine, refByteIndex int,
) error {
	var (
		context, overridenContext, previousLine, previousReferenceLine,
		currentReferenceLine, nextReferenceLine int
		temp byte
		err  error
	)

	if lineNumber > 0 {
		temp, err = g.RegionBitmap.GetByte(byteIndex - rowStride)
		if err != nil {
			return err
		}
		previousLine = int(temp)
	}

	if currentLine > 0 && currentLine <= g.ReferenceBitmap.Height {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride + deltaRefStride)
		if err != nil {
			return err
		}
		previousReferenceLine = int(temp) << 4
	}

	if currentLine >= 0 && currentLine < g.ReferenceBitmap.Height {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex + deltaRefStride)
		if err != nil {
			return err
		}
		currentReferenceLine = int(temp) << 1
	}

	if currentLine > -2 && currentLine < g.ReferenceBitmap.Height-1 {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride + deltaRefStride)
		if err != nil {
			return err
		}
		nextReferenceLine = int(temp)
	}

	context = ((previousLine >> 5) & 0x6) | ((nextReferenceLine >> 2) & 0x30) | (currentReferenceLine & 0x180) | (previousReferenceLine & 0xc00)
	var nextByte int

	for x := 0; x < paddedWidth; x = nextByte {
		var result int
		nextByte = x + 8
		var minorWidth int

		if minorWidth = width - x; minorWidth > 8 {
			minorWidth = 8
		}

		readNextByte := nextByte < width
		refReadNextByte := nextByte < g.ReferenceBitmap.Width
		yOffset := deltaRefStride + 1

		if lineNumber > 0 {
			temp = 0
			if readNextByte {
				temp, err = g.RegionBitmap.GetByte(byteIndex - rowStride + 1)
				if err != nil {
					return err
				}
			}
			previousLine = (previousLine << 8) | int(temp)
		}

		if currentLine > 0 && currentLine <= g.ReferenceBitmap.Height {
			var tempVal int
			if refReadNextByte {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride + yOffset)
				if err != nil {
					return err
				}
				tempVal = int(temp) << 4
			}
			previousReferenceLine = (previousReferenceLine << 8) | tempVal
		}

		if currentLine >= 0 && currentLine < g.ReferenceBitmap.Height {
			var tempVal int
			if refReadNextByte {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex + yOffset)
				if err != nil {
					return err
				}
				tempVal = int(temp) << 1
			}
			currentReferenceLine = (currentReferenceLine << 8) | tempVal
		}

		if currentLine > -2 && currentLine < (g.ReferenceBitmap.Height-1) {
			temp = 0
			if refReadNextByte {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride + yOffset)
				if err != nil {
					return err
				}
			}
			nextReferenceLine = (nextReferenceLine << 8) | int(temp)
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			var bit int
			isPixelTypicalPredicted := false
			bitmapValue := (context >> 4) & 0x1ff

			if bitmapValue == 0x1ff {
				isPixelTypicalPredicted = true
				bit = 1
			} else if bitmapValue == 0x00 {
				isPixelTypicalPredicted = true
			}

			if !isPixelTypicalPredicted {
				if g.override {
					overridenContext = g.overrideAtTemplate0(context, x+minorX, lineNumber, result, minorX)
					g.cx.SetIndex(int32(overridenContext))
				} else {
					g.cx.SetIndex(int32(context))
				}

				bit, err = g.arithDecode.DecodeBit(g.cx)
				if err != nil {
					return err
				}
			}

			toShift := uint(7 - minorX)
			result |= int(bit << toShift)
			context = ((context & 0xdb6) << 1) | bit | (previousLine>>toShift+5)&0x002 |
				((nextReferenceLine>>toShift + 2) & 0x010) |
				((currentReferenceLine >> toShift) & 0x080) |
				((previousReferenceLine >> toShift) & 0x400)
		}

		err = g.RegionBitmap.SetByte(byteIndex, byte(result))
		if err != nil {
			return err
		}

		byteIndex++
		refByteIndex++
	}
	return nil
}

func (g *GenericRefinementRegion) decodeTypicalPredictedLineTemplate1(
	lineNumber, width, rowStride, refRowStride, paddedWidth,
	deltaRefStride, byteIndex, currentLine, refByteIndex int,
) (err error) {
	var (
		context, grReferenceValue               int
		previousLine, previousReferenceLine     int
		currentReferenceLine, nextReferenceLine int
		temp                                    byte
	)

	if lineNumber > 0 {
		temp, err = g.RegionBitmap.GetByte(byteIndex - rowStride)
		if err != nil {
			return
		}
		previousLine = int(temp)
	}

	if currentLine > 0 && currentLine <= g.ReferenceBitmap.Height {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride + deltaRefStride)
		if err != nil {
			return
		}
		previousReferenceLine = int(temp) << 2
	}

	if currentLine >= 0 && currentLine < g.ReferenceBitmap.Height {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex + deltaRefStride)
		if err != nil {
			return
		}
		currentReferenceLine = int(temp)
	}

	if currentLine > -2 && currentLine < g.ReferenceBitmap.Height-1 {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride + deltaRefStride)
		if err != nil {
			return
		}
		nextReferenceLine = int(temp)
	}

	context = ((previousLine >> 5) & 0x6) | ((nextReferenceLine >> 2) & 0x30) |
		(currentReferenceLine & 0xc0) | (previousReferenceLine & 0x200)
	grReferenceValue = ((nextReferenceLine >> 2) & 0x70) | (currentReferenceLine & 0xc0) |
		(previousReferenceLine & 0x700)
	var nextByte int

	for x := 0; x < paddedWidth; x = nextByte {
		var (
			minorWidth int
			result     int
		)
		nextByte = x + 8

		if minorWidth = width - x; minorWidth > 8 {
			minorWidth = 8
		}

		readNextByte := nextByte < width
		refReadNextByte := nextByte < g.ReferenceBitmap.Width
		yOffset := deltaRefStride + 1

		if lineNumber > 0 {
			temp = 0
			if readNextByte {
				temp, err = g.RegionBitmap.GetByte(byteIndex - rowStride + 1)
				if err != nil {
					return
				}
			}
			previousLine = (previousLine << 8) | int(temp)
		}

		if currentLine > 0 && currentLine <= g.ReferenceBitmap.Height {
			var tempVal int
			if refReadNextByte {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride + yOffset)
				if err != nil {
					return
				}
				tempVal = int(temp) << 2
			}
			previousReferenceLine = (previousReferenceLine << 8) | tempVal
		}

		if currentLine >= 0 && currentLine < g.ReferenceBitmap.Height {
			temp = 0
			if refReadNextByte {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex + yOffset)
				if err != nil {
					return
				}
			}
			currentReferenceLine = (currentReferenceLine << 8) | int(temp)
		}

		if currentLine > -2 && currentLine < (g.ReferenceBitmap.Height-1) {
			temp = 0
			if refReadNextByte {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride + yOffset)
				if err != nil {
					return
				}
			}
			nextReferenceLine = (nextReferenceLine << 8) | int(temp)
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			var bit int
			bitmapValue := (grReferenceValue >> 4) & 0x1ff

			switch bitmapValue {
			case 0x1ff:
				bit = 1
			case 0x00:
				bit = 0
			default:
				g.cx.SetIndex(int32(context))
				bit, err = g.arithDecode.DecodeBit(g.cx)
				if err != nil {
					return
				}
			}

			toShift := uint(7 - minorX)
			result |= int(bit << toShift)
			context = ((context & 0x0d6) << 1) | bit | (previousLine>>toShift+5)&0x002 |
				((nextReferenceLine>>toShift + 2) & 0x010) |
				((currentReferenceLine >> toShift) & 0x040) |
				((previousReferenceLine >> toShift) & 0x200)
			grReferenceValue = ((grReferenceValue & 0xdb) << 1) |
				((nextReferenceLine>>toShift + 2) & 0x010) |
				((currentReferenceLine >> toShift) & 0x080) |
				((previousReferenceLine >> toShift) & 0x400)
		}

		err = g.RegionBitmap.SetByte(byteIndex, byte(result))
		if err != nil {
			return
		}

		byteIndex++
		refByteIndex++
	}
	return nil
}

func (g *GenericRefinementRegion) decodeTemplate(
	lineNumber, width, rowStride, refRowStride, paddedWidth,
	deltaRefStride, lineOffset, byteIndex, currentLine, refByteIndex int,
	templateFormation templater,
) (err error) {
	var (
		c1, c2, c3, c4, c5 int16
		w1, w2, w3, w4     int
		temp               byte
	)

	if currentLine >= 1 && (currentLine-1) < g.ReferenceBitmap.Height {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride)
		if err != nil {
			return
		}
		w1 = int(temp)
	}

	if currentLine >= 0 && (currentLine) < g.ReferenceBitmap.Height {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex)
		if err != nil {
			return
		}
		w2 = int(temp)
	}

	if currentLine >= -1 && (currentLine+1) < g.ReferenceBitmap.Height {
		temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride)
		if err != nil {
			return
		}
		w3 = int(temp)
	}
	refByteIndex++

	if lineNumber >= 1 {
		temp, err = g.RegionBitmap.GetByte(byteIndex - rowStride)
		if err != nil {
			return
		}
		w4 = int(temp)
	}
	byteIndex++

	modReferenceDX := g.ReferenceDX % 8
	shiftOffset := 6 + modReferenceDX
	modRefByteIdx := refByteIndex % refRowStride

	if shiftOffset >= 0 {
		if shiftOffset < 8 {
			c1 = int16(w1>>uint(shiftOffset)) & 0x07
		}

		if shiftOffset < 8 {
			c2 = int16(w2>>uint(shiftOffset)) & 0x07
		}

		if shiftOffset < 8 {
			c3 = int16(w3>>uint(shiftOffset)) & 0x07
		}

		if shiftOffset == 6 && modRefByteIdx > 1 {
			if currentLine >= 1 && (currentLine-1) < g.ReferenceBitmap.Height {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride - 2)
				if err != nil {
					return err
				}
				c1 |= int16(temp<<2) & 0x04
			}

			if currentLine >= 0 && currentLine < g.ReferenceBitmap.Height {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex - 2)
				if err != nil {
					return err
				}
				c2 |= int16(temp<<2) & 0x04
			}

			if currentLine >= -1 && currentLine+1 < g.ReferenceBitmap.Height {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride - 2)
				if err != nil {
					return err
				}
				c3 |= int16(temp<<2) & 0x04
			}
		}

		if shiftOffset == 0 {
			w1 = 0
			w2 = 0
			w3 = 0

			if modRefByteIdx < refRowStride-1 {
				if currentLine >= 1 && currentLine-1 < g.ReferenceBitmap.Height {
					temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride)
					if err != nil {
						return err
					}
					w1 = int(temp)
				}

				if currentLine >= 0 && currentLine < g.ReferenceBitmap.Height {
					temp, err = g.ReferenceBitmap.GetByte(refByteIndex)
					if err != nil {
						return err
					}
					w2 = int(temp)
				}

				if currentLine >= -1 && currentLine+1 < g.ReferenceBitmap.Height {
					temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride)
					if err != nil {
						return err
					}
					w3 = int(temp)
				}
			}
			refByteIndex++
		}
	} else {
		c1 = int16(w1<<1) & 0x07
		c2 = int16(w2<<1) & 0x07
		c3 = int16(w3<<1) & 0x07
		w1 = 0
		w2 = 0
		w3 = 0

		if modRefByteIdx < refRowStride-1 {
			if currentLine >= 1 && currentLine-1 < g.ReferenceBitmap.Height {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride)
				if err != nil {
					return err
				}
				w1 = int(temp)
			}

			if currentLine >= 0 && currentLine < g.ReferenceBitmap.Height {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex)
				if err != nil {
					return err
				}
				w2 = int(temp)
			}

			if currentLine >= -1 && currentLine+1 < g.ReferenceBitmap.Height {
				temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride)
				if err != nil {
					return err
				}
				w3 = int(temp)
			}
			refByteIndex++
		}

		c1 |= int16((w1 >> 7) & 0x07)
		c2 |= int16((w2 >> 7) & 0x07)
		c3 |= int16((w3 >> 7) & 0x07)
	}
	c4 = int16(w4 >> 6)
	c5 = 0

	modBitsToTrim := (2 - modReferenceDX) % 8
	w1 <<= uint(modBitsToTrim)
	w2 <<= uint(modBitsToTrim)
	w3 <<= uint(modBitsToTrim)
	w4 <<= 2
	var bit int

	for x := 0; x < width; x++ {
		minorX := x & 0x07
		tval := templateFormation.form(c1, c2, c3, c4, c5)

		if g.override {
			temp, err = g.RegionBitmap.GetByte(g.RegionBitmap.GetByteIndex(x, lineNumber))
			if err != nil {
				return err
			}
			g.cx.SetIndex(int32(g.overrideAtTemplate0(int(tval), x, lineNumber, int(temp), minorX)))
		} else {
			g.cx.SetIndex(int32(tval))
		}

		bit, err = g.arithDecode.DecodeBit(g.cx)
		if err != nil {
			return err
		}

		if err = g.RegionBitmap.SetPixel(x, lineNumber, byte(bit)); err != nil {
			return err
		}

		c1 = ((c1 << 1) | 0x01&int16(w1>>7)) & 0x07
		c2 = ((c2 << 1) | 0x01&int16(w2>>7)) & 0x07
		c3 = ((c3 << 1) | 0x01&int16(w3>>7)) & 0x07
		c4 = ((c4 << 1) | 0x01&int16(w4>>7)) & 0x07
		c5 = int16(bit)

		if (x-int(g.ReferenceDX))%8 == 5 {
			w1 = 0
			w2 = 0
			w3 = 0

			if ((x-int(g.ReferenceDX))/8)+1 < g.ReferenceBitmap.RowStride {
				if currentLine >= 1 && (currentLine-1) < g.ReferenceBitmap.Height {
					temp, err = g.ReferenceBitmap.GetByte(refByteIndex - refRowStride)
					if err != nil {
						return err
					}
					w1 = int(temp)
				}

				if currentLine >= 0 && currentLine < g.ReferenceBitmap.Height {
					temp, err = g.ReferenceBitmap.GetByte(refByteIndex)
					if err != nil {
						return err
					}
					w2 = int(temp)
				}

				if currentLine >= -1 && (currentLine+1) < g.ReferenceBitmap.Height {
					temp, err = g.ReferenceBitmap.GetByte(refByteIndex + refRowStride)
					if err != nil {
						return err
					}
					w3 = int(temp)
				}
			}
			refByteIndex++
		} else {
			w1 <<= 1
			w2 <<= 1
			w3 <<= 1
		}

		if minorX == 5 && lineNumber >= 1 {
			if ((x >> 3) + 1) >= g.RegionBitmap.RowStride {
				w4 = 0
			} else {
				temp, err = g.RegionBitmap.GetByte(byteIndex - rowStride)
				if err != nil {
					return err
				}
				w4 = int(temp)
			}
			byteIndex++
		} else {
			w4 <<= 1
		}
	}
	return nil
}

func (g *GenericRefinementRegion) getPixel(b *bitmap.Bitmap, x, y int) int {
	if x < 0 || x >= b.Width {
		return 0
	}

	if y < 0 || y >= b.Height {
		return 0
	}

	if b.GetPixel(x, y) {
		return 1
	}
	return 0
}

func (g *GenericRefinementRegion) overrideAtTemplate0(context, x, y, result, minorX int) int {
	if g.grAtOverride[0] {
		context &= 0xfff7
		if g.GrAtY[0] == 0 && int(g.GrAtX[0]) >= -minorX {
			context |= (result >> uint(7-(minorX+int(g.GrAtX[0]))) & 0x1) << 3
		} else {
			context |= g.getPixel(g.RegionBitmap, x+int(g.GrAtX[0]), y+int(g.GrAtY[0])) << 3
		}
	}

	if g.grAtOverride[1] {
		context &= 0xefff
		if g.GrAtY[1] == 0 && int(g.GrAtX[1]) >= -minorX {
			context |= (result >> uint(7-(minorX+int(g.GrAtX[1]))) & 0x1) << 12
		} else {
			context |= g.getPixel(g.ReferenceBitmap, x+int(g.GrAtX[1]), y+int(g.GrAtY[1]))
		}
	}
	return context
}

func (g *GenericRefinementRegion) parseHeader() (err error) {
	common.Log.Trace("[GENERIC-REF-REGION] parsing Header...")
	ts := time.Now()
	defer func() {
		if err == nil {
			common.Log.Trace("[GENERIC-REF-REGION] parsing header finishid in: %d ns", time.Since(ts).Nanoseconds())
		} else {
			common.Log.Trace("[GENERIC-REF-REGION] parsing header failed: %s", err)
		}
	}()

	if err = g.RegionInfo.parseHeader(); err != nil {
		return err
	}

	// Bit 2-7
	_, err = g.r.ReadBits(6) // Dirty Read
	if err != nil {
		return err
	}

	// Bit 1
	g.IsTPGROn, err = g.r.ReadBool()
	if err != nil {
		return err
	}

	// Bit 0
	var templateID int
	templateID, err = g.r.ReadBit()
	if err != nil {
		return err
	}
	g.TemplateID = int8(templateID)

	switch g.TemplateID {
	case 0:
		g.Template = g.t0
		if err = g.readAtPixels(); err != nil {
			return
		}
	case 1:
		g.Template = g.t1
	}
	return nil
}

func (g *GenericRefinementRegion) readAtPixels() error {
	g.GrAtX = make([]int8, 2)
	g.GrAtY = make([]int8, 2)

	// Byte 0
	temp, err := g.r.ReadByte()
	if err != nil {
		return err
	}
	g.GrAtX[0] = int8(temp)

	// Byte 1
	temp, err = g.r.ReadByte()
	if err != nil {
		return err
	}
	g.GrAtY[0] = int8(temp)

	// Byte 2
	temp, err = g.r.ReadByte()
	if err != nil {
		return err
	}
	g.GrAtX[1] = int8(temp)

	// Byte 3
	temp, err = g.r.ReadByte()
	if err != nil {
		return err
	}
	g.GrAtY[1] = int8(temp)
	return nil
}

// setParameters sets the parameters for the Generic Refinemenet Region.
func (g *GenericRefinementRegion) setParameters(
	cx *arithmetic.DecoderStats, arithmDecoder *arithmetic.Decoder,
	grTemplate int8, regionWidth, regionHeight uint32,
	grReference *bitmap.Bitmap, grReferenceDX, grReferenceDY int32,
	isTPGRon bool, grAtX []int8, grAtY []int8,
) {
	common.Log.Trace("[GENERIC-REF-REGION] setParameters")
	if cx != nil {
		g.cx = cx
	}

	if arithmDecoder != nil {
		g.arithDecode = arithmDecoder
	}

	g.TemplateID = grTemplate
	g.RegionInfo.BitmapWidth = regionWidth
	g.RegionInfo.BitmapHeight = regionHeight
	g.ReferenceBitmap = grReference
	g.ReferenceDX = grReferenceDX
	g.ReferenceDY = grReferenceDY
	g.IsTPGROn = isTPGRon
	g.GrAtX = grAtX
	g.GrAtY = grAtY
	g.RegionBitmap = nil

	common.Log.Trace("[GENERIC-REF-REGION] setParameters finished. %s", g)
}

func (g *GenericRefinementRegion) updateOverride() error {
	if g.GrAtX == nil || g.GrAtY == nil {
		return errors.New("AT pixels not set")
	}

	if len(g.GrAtX) != len(g.GrAtY) {
		return errors.New("AT pixel inconsistent")
	}

	g.grAtOverride = make([]bool, len(g.GrAtX))

	switch g.TemplateID {
	case 0:
		if g.GrAtX[0] != -1 && g.GrAtY[0] != -1 {
			g.grAtOverride[0] = true
			g.override = true
		}

		if g.GrAtX[1] != -1 && g.GrAtY[1] != -1 {
			g.grAtOverride[1] = true
			g.override = true
		}
	case 1:
		g.override = false
	}
	return nil
}

type templater interface {
	form(c1, c2, c3, c4, c5 int16) int16
	setIndex(cx *arithmetic.DecoderStats)
}

type template0 struct{}

func (t *template0) form(c1, c2, c3, c4, c5 int16) int16 {
	return (c1 << 10) | (c2 << 7) | (c3 << 4) | (c4 << 1) | c5
}

func (t *template0) setIndex(cx *arithmetic.DecoderStats) {
	// Figure 14, page 22
	cx.SetIndex(0x100)
}

var _ templater = &template0{}

type template1 struct{}

func (t *template1) form(c1, c2, c3, c4, c5 int16) int16 {
	return ((c1 & 0x02) << 8) | (c2 << 6) | ((c3 & 0x03) << 4) | (c4 << 1) | c5
}

func (t *template1) setIndex(cx *arithmetic.DecoderStats) {
	// Figure 15, page 22
	cx.SetIndex(0x080)
}

var _ templater = &template1{}

// String implements the Stringer interface.
func (g *GenericRefinementRegion) String() string {
	sb := &strings.Builder{}
	sb.WriteString("\n[GENERIC REGION]\n")
	sb.WriteString(g.RegionInfo.String() + "\n")
	sb.WriteString(fmt.Sprintf("\t- IsTPGRon: %v\n", g.IsTPGROn))
	sb.WriteString(fmt.Sprintf("\t- TemplateID: %v\n", g.TemplateID))
	sb.WriteString(fmt.Sprintf("\t- GrAtX: %v\n", g.GrAtX))
	sb.WriteString(fmt.Sprintf("\t- GrAtY: %v\n", g.GrAtY))
	sb.WriteString(fmt.Sprintf("\t- ReferenceDX %v\n", g.ReferenceDX))
	sb.WriteString(fmt.Sprintf("\t- ReferencDeY: %v\n", g.ReferenceDY))
	return sb.String()
}

// newGenericRefinementRegion is a creator for the Generic Refinement Region.
func newGenericRefinementRegion(r reader.StreamReader, h *Header) *GenericRefinementRegion {
	return &GenericRefinementRegion{
		r:          r,
		RegionInfo: NewRegionSegment(r),
		h:          h,
		t0:         &template0{},
		t1:         &template1{},
	}
}
