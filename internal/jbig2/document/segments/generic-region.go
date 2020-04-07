/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"fmt"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/mmr"
	enc "github.com/unidoc/unipdf/v3/internal/jbig2/encoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// GenericRegion represents a generic region segment.
// Parsing is done as described in 7.4.5.
// Decoding procedure is done as described in 6.2.5.7 and 7.4.6.4.
type GenericRegion struct {
	r reader.StreamReader

	DataHeaderOffset int64
	DataHeaderLength int64
	DataOffset       int64
	DataLength       int64

	// Region segment information field, 7.4.1
	RegionSegment *RegionSegment

	// Generic region segment flags, 7.4.6.2
	UseExtTemplates bool
	IsTPGDon        bool
	GBTemplate      byte
	IsMMREncoded    bool
	UseMMR          bool

	// Generic region segment AT flags, 7.4.6.3
	GBAtX        []int8
	GBAtY        []int8
	GBAtOverride []bool

	// override defines if AT pixels are not on their normal location and have to be overwriten.
	override bool

	// Bitmap is the decoded generic region image.
	Bitmap *bitmap.Bitmap

	arithDecoder *arithmetic.Decoder
	cx           *arithmetic.DecoderStats
	mmrDecoder   *mmr.Decoder
}

// NewGenericRegion creates new GenericRegion segment.
func NewGenericRegion(r reader.StreamReader) *GenericRegion {
	return &GenericRegion{RegionSegment: NewRegionSegment(r), r: r}
}

// compile time check for the SegmentEncoder interface.
var _ SegmentEncoder = &GenericRegion{}

// Encode implements SegmentEncoder interface.
func (g *GenericRegion) Encode(w writer.BinaryWriter) (n int, err error) {
	const processName = "GenericRegion.Encode"
	if g.Bitmap == nil {
		return 0, errors.Error(processName, "provided nil bitmap")
	}

	// at first encode the region segment.
	i, err := g.RegionSegment.Encode(w)
	if err != nil {
		return 0, errors.Wrap(err, processName, "RegionSegment")
	}
	n += i

	// skip 4 reserved bits
	if err = w.SkipBits(4); err != nil {
		return n, errors.Wrap(err, processName, "skip reserved bits")
	}
	var bit int
	if g.IsTPGDon {
		bit = 1
	}

	// Write TPGDON bit
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "tpgdon")
	}
	bit = 0

	// write two bits of the gb template
	if err = w.WriteBit(int(g.GBTemplate>>1) & 0x01); err != nil {
		return n, errors.Wrap(err, processName, "first gbtemplate bit")
	}
	if err = w.WriteBit(int(g.GBTemplate) & 0x01); err != nil {
		return n, errors.Wrap(err, processName, "second gbtemplate bit")
	}

	// encode mmr flag
	if g.UseMMR {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "use MMR bit")
	}
	n++

	// write GB At pixels
	if i, err = g.writeGBAtPixels(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += i

	// encode the region using arithmetic encoder
	ctx := enc.New()
	if err = ctx.EncodeBitmap(g.Bitmap, g.IsTPGDon); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	ctx.Final()

	// write the encoded bitmap value to the 'w' writer
	var n64 int64
	if n64, err = ctx.WriteTo(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += int(n64)
	return n, nil
}

// GetRegionBitmap gets the bitmap for the generic region segment.
func (g *GenericRegion) GetRegionBitmap() (bm *bitmap.Bitmap, err error) {
	if g.Bitmap != nil {
		return g.Bitmap, nil
	}

	// Check how the data is encoded.
	if g.IsMMREncoded {
		// MMR Decoder Call
		if g.mmrDecoder == nil {
			g.mmrDecoder, err = mmr.New(g.r, int(g.RegionSegment.BitmapWidth), int(g.RegionSegment.BitmapHeight), g.DataOffset, g.DataLength)
			if err != nil {
				return nil, err
			}
		}

		// Uncompress the bitmap.
		g.Bitmap, err = g.mmrDecoder.UncompressMMR()
		return g.Bitmap, err
	}

	// Arithmetic decoder process for generic region segments.
	if err = g.updateOverrideFlags(); err != nil {
		return nil, err
	}

	// 6.2.5.7 - 1)
	var ltp int
	if g.arithDecoder == nil {
		g.arithDecoder, err = arithmetic.New(g.r)
		if err != nil {
			return nil, err
		}
	}
	if g.cx == nil {
		g.cx = arithmetic.NewStats(65536, 1)
	}

	// 6.2.5.7 - 2)
	g.Bitmap = bitmap.New(int(g.RegionSegment.BitmapWidth), int(g.RegionSegment.BitmapHeight))
	paddedWidth := int(uint32(g.Bitmap.Width+7) & (^uint32(7)))

	// 6.2.5.7 - 3)
	for line := 0; line < g.Bitmap.Height; line++ {
		// 6.2.5.7 - 3 c)
		if g.IsTPGDon {
			var temp int
			temp, err = g.decodeSLTP()
			if err != nil {
				return nil, err
			}
			ltp ^= temp
		}

		// 6.2.5.7 - 3 d)
		if ltp == 1 {
			if line > 0 {
				if err = g.copyLineAbove(line); err != nil {
					return nil, err
				}
			}
		} else {
			if err = g.decodeLine(line, g.Bitmap.Width, paddedWidth); err != nil {
				return nil, err
			}
		}
	}
	return g.Bitmap, nil
}

// GetRegionInfo implements Regioner interface.
func (g *GenericRegion) GetRegionInfo() *RegionSegment {
	return g.RegionSegment
}

// Init implements Segmenter interface.
func (g *GenericRegion) Init(h *Header, r reader.StreamReader) error {
	g.RegionSegment = NewRegionSegment(r)
	g.r = r
	return g.parseHeader()
}

// InitEncode initializes the generic region for the provided bitmap 'bm', it's 'xLoc', 'yLoc' locations and if it has to remove duplicated lines.
func (g *GenericRegion) InitEncode(bm *bitmap.Bitmap, xLoc, yLoc, template int, duplicateLineRemoval bool) error {
	const processName = "GenericRegion.InitEncode"
	if bm == nil {
		return errors.Error(processName, "provided nil bitmap")
	}
	if xLoc < 0 || yLoc < 0 {
		return errors.Error(processName, "x and y location must be greater than 0")
	}
	g.Bitmap = bm
	g.GBTemplate = byte(template)
	// use the default GB template 0 with 4 at pixel bytes
	switch g.GBTemplate {
	case 0:
		g.GBAtX = []int8{3, -3, 2, -2}
		g.GBAtY = []int8{-1, -1, -2, -2}
	case 1:
		g.GBAtX = []int8{3}
		g.GBAtY = []int8{-1}
	case 2, 3:
		g.GBAtX = []int8{2}
		g.GBAtY = []int8{-1}
	default:
		return errors.Errorf(processName, "provided template: '%d' not in valid range {0,1,2,3}", template)
	}
	g.RegionSegment = &RegionSegment{
		BitmapHeight: uint32(bm.Height),
		BitmapWidth:  uint32(bm.Width),
		XLocation:    uint32(xLoc),
		YLocation:    uint32(yLoc),
	}
	g.IsTPGDon = duplicateLineRemoval
	return nil
}

// Size returns the byte size of the generic region.
func (g *GenericRegion) Size() int {
	// region size + flags + 2 * gb pixel size
	return g.RegionSegment.Size() + 1 + 2*len(g.GBAtX)
}

// String implements Stringer interface.
func (g *GenericRegion) String() string {
	sb := &strings.Builder{}

	sb.WriteString("\n[GENERIC REGION]\n")
	sb.WriteString(g.RegionSegment.String() + "\n")
	sb.WriteString(fmt.Sprintf("\t- UseExtTemplates: %v\n", g.UseExtTemplates))
	sb.WriteString(fmt.Sprintf("\t- IsTPGDon: %v\n", g.IsTPGDon))
	sb.WriteString(fmt.Sprintf("\t- GBTemplate: %d\n", g.GBTemplate))
	sb.WriteString(fmt.Sprintf("\t- IsMMREncoded: %v\n", g.IsMMREncoded))
	sb.WriteString(fmt.Sprintf("\t- GBAtX: %v\n", g.GBAtX))
	sb.WriteString(fmt.Sprintf("\t- GBAtY: %v\n", g.GBAtY))
	sb.WriteString(fmt.Sprintf("\t- GBAtOverride: %v\n", g.GBAtOverride))
	return sb.String()
}
func (g *GenericRegion) parseHeader() (err error) {
	common.Log.Trace("[GENERIC-REGION] ParsingHeader...")
	defer func() {
		if err != nil {
			common.Log.Trace("[GENERIC-REGION] ParsingHeader Finished with error. %v", err)
		} else {
			common.Log.Trace("[GENERIC-REGION] ParsingHeader Finished Successfully...")
		}
	}()
	var (
		b    int
		bits uint64
	)

	if err = g.RegionSegment.parseHeader(); err != nil {
		return err
	}

	// Bit 5-7
	if _, err = g.r.ReadBits(3); err != nil {
		return err
	}

	// Bit 4
	b, err = g.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		g.UseExtTemplates = true
	}

	// Bit 3
	b, err = g.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		g.IsTPGDon = true
	}

	// Bit 1-2
	bits, err = g.r.ReadBits(2)
	if err != nil {
		return err
	}
	g.GBTemplate = byte(bits & 0xf)

	// Bit 0
	b, err = g.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		g.IsMMREncoded = true
	}

	if !g.IsMMREncoded {
		numberOfGbAt := 1
		if g.GBTemplate == 0 {
			numberOfGbAt = 4
			if g.UseExtTemplates {
				numberOfGbAt = 12
			}
		}
		if err = g.readGBAtPixels(numberOfGbAt); err != nil {
			return err
		}
	}

	// Segment data structure
	if err = g.computeSegmentDataStructure(); err != nil {
		return err
	}
	common.Log.Trace("%s", g)
	return nil
}

func (g *GenericRegion) computeSegmentDataStructure() error {
	g.DataOffset = g.r.StreamPosition()
	g.DataHeaderLength = g.DataOffset - g.DataHeaderOffset
	g.DataLength = int64(g.r.Length()) - g.DataHeaderLength
	return nil
}

func (g *GenericRegion) copyLineAbove(line int) error {
	targetByteIndex := line * g.Bitmap.RowStride
	sourceByteIndex := targetByteIndex - g.Bitmap.RowStride
	for i := 0; i < g.Bitmap.RowStride; i++ {
		b, err := g.Bitmap.GetByte(sourceByteIndex)
		if err != nil {
			return err
		}
		sourceByteIndex++
		if err = g.Bitmap.SetByte(targetByteIndex, b); err != nil {
			return err
		}
		targetByteIndex++
	}
	return nil
}

func (g *GenericRegion) decodeSLTP() (int, error) {
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

func (g *GenericRegion) decodeLine(line, width, paddedWidth int) error {
	const processName = "decodeLine"
	byteIndex := g.Bitmap.GetByteIndex(0, line)
	idx := byteIndex - g.Bitmap.RowStride

	switch g.GBTemplate {
	case 0:
		if !g.UseExtTemplates {
			return g.decodeTemplate0a(line, width, paddedWidth, byteIndex, idx)
		}
		return g.decodeTemplate0b(line, width, paddedWidth, byteIndex, idx)
	case 1:
		return g.decodeTemplate1(line, width, paddedWidth, byteIndex, idx)
	case 2:
		return g.decodeTemplate2(line, width, paddedWidth, byteIndex, idx)
	case 3:
		return g.decodeTemplate3(line, width, paddedWidth, byteIndex, idx)
	}
	return errors.Errorf(processName, "invalid GBTemplate provided: %d", g.GBTemplate)
}

func (g *GenericRegion) decodeTemplate0a(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	const processName = "decodeTemplate0a"
	var (
		context, overriddenContext int
		line1, line2               int
		temp                       byte
		nextByte                   int
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return errors.Wrap(err, processName, "line >= 1")
		}
		line1 = int(temp)
	}

	if line >= 2 {
		temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride)
		if err != nil {
			return errors.Wrap(err, processName, "line >= 2")
		}
		line2 = int(temp) << 6
	}
	context = (line1 & 0xf0) | (line2 & 0x3800)

	for x := 0; x < paddedWidth; x = nextByte {
		// 6.2.5.7 3d
		var (
			result     byte
			minorWidth int
		)
		nextByte = x + 8

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
					return errors.Wrap(err, processName, "line > 0")
				}
				line1 |= int(temp)
			}
		}

		if line > 1 {
			index := idx - g.Bitmap.RowStride + 1
			line2 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(index)
				if err != nil {
					return errors.Wrap(err, processName, "line > 1")
				}
				line2 |= int(temp) << 6
			} else {
				line2 |= 0
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			toShift := uint(7 - minorX)

			if g.override {
				overriddenContext = g.overrideAtTemplate0a(context, x+minorX, line, int(result), minorX, int(toShift))
				g.cx.SetIndex(int32(overriddenContext))
			} else {
				g.cx.SetIndex(int32(context))
			}

			var bit int
			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			result |= byte(bit) << uint(toShift)
			context = ((context & 0x7bf7) << 1) | bit | ((line1 >> toShift) & 0x10) | ((line2 >> toShift) & 0x800)
		}

		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return errors.Wrap(err, processName, "")
		}

		byteIndex++
		idx++
	}
	return nil
}

func (g *GenericRegion) decodeTemplate0b(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	const processName = "decodeTemplate0b"
	var (
		context, overriddenContext int
		line1, line2               int
		temp                       byte
		nextByte                   int
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return errors.Wrap(err, processName, "line >= 1")
		}
		line1 = int(temp)
	}

	if line >= 2 {
		temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride)
		if err != nil {
			return errors.Wrap(err, processName, "line >= 2")
		}
		line2 = int(temp) << 6
	}
	context = (line1 & 0xf0) | (line2 & 0x3800)

	for x := 0; x < paddedWidth; x = nextByte {
		// 6.2.5.7 3d
		var (
			result     byte
			minorWidth int
		)
		nextByte = x + 8

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
					return errors.Wrap(err, processName, "line > 0")
				}
				line1 |= int(temp)
			}
		}

		if line > 1 {
			line2 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride + 1)
				if err != nil {
					return errors.Wrap(err, processName, "line > 1")
				}
				line2 |= int(temp) << 6
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			toShift := uint(7 - minorX)

			if g.override {
				overriddenContext = g.overrideAtTemplate0b(context, x+minorX, line, int(result), minorX, int(toShift))
				g.cx.SetIndex(int32(overriddenContext))
			} else {
				g.cx.SetIndex(int32(context))
			}

			var bit int
			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			result |= byte(bit << uint(toShift))
			context = ((context & 0x7bf7) << 1) | bit | ((line1 >> toShift) & 0x10) | ((line2 >> toShift) & 0x800)
		}

		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return errors.Wrap(err, processName, "")
		}

		byteIndex++
		idx++
	}
	return nil
}

func (g *GenericRegion) decodeTemplate1(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	const processName = "decodeTemplate1"
	var (
		context, overriddenContext int
		line1, line2               int
		temp                       byte
		nextByte, bit              int
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return errors.Wrap(err, processName, "line >= 1")
		}

		line1 = int(temp)
	}

	if line >= 2 {
		temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride)
		if err != nil {
			return errors.Wrap(err, processName, "line >= 2")
		}
		line2 = int(temp) << 5
	}

	context = ((line1 >> 1) & 0x1f8) | ((line2 >> 1) & 0x1e00)

	for x := 0; x < paddedWidth; x = nextByte {
		// 6.2.5.7 3d
		var (
			result     byte
			minorWidth int
		)
		nextByte = x + 8

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
					return errors.Wrap(err, processName, "line > 0")
				}
				line1 |= int(temp)
			}
		}

		if line > 1 {
			line2 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride + 1)
				if err != nil {
					return errors.Wrap(err, processName, "line > 1")
				}
				line2 |= int(temp) << 5
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			if g.override {
				overriddenContext = g.overrideAtTemplate1(context, x+minorX, line, int(result), minorX)
				g.cx.SetIndex(int32(overriddenContext))
			} else {
				g.cx.SetIndex(int32(context))
			}

			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			result |= byte(bit) << uint(7-minorX)
			toShift := uint(8 - minorX)
			context = ((context & 0xefb) << 1) | bit | ((line1 >> toShift) & 0x8) | ((line2 >> toShift) & 0x200)
		}

		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return errors.Wrap(err, processName, "")
		}

		byteIndex++
		idx++
	}
	return nil
}

func (g *GenericRegion) decodeTemplate2(lineNumber, width, paddedWidth int, byteIndex, idx int) (err error) {
	const processName = "decodeTemplate2"
	var (
		context, overriddenContext int
		line1, line2               int
		temp                       byte
		nextByte, bit              int
	)

	if lineNumber >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return errors.Wrap(err, processName, "lineNumber >= 1")
		}
		line1 = int(temp)
	}

	if lineNumber >= 2 {
		temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride)
		if err != nil {
			return errors.Wrap(err, processName, "lineNumber >= 2")
		}
		line2 = int(temp) << 4
	}

	context = (line1 >> 3 & 0x7c) | (line2 >> 3 & 0x380)

	for x := 0; x < paddedWidth; x = nextByte {
		// 6.2.5.7 3d
		var (
			result     byte
			minorWidth int
		)
		nextByte = x + 8

		if d := width - x; d > 8 {
			minorWidth = 8
		} else {
			minorWidth = d
		}

		if lineNumber > 0 {
			line1 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx + 1)
				if err != nil {
					return errors.Wrap(err, processName, "lineNumber > 0")
				}
				line1 |= int(temp)
			}
		}

		if lineNumber > 1 {
			line2 <<= 8

			if nextByte < width {
				temp, err = g.Bitmap.GetByte(idx - g.Bitmap.RowStride + 1)
				if err != nil {
					return errors.Wrap(err, processName, "lineNumber > 1")
				}
				line2 |= int(temp) << 4
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			toShift := uint(10 - minorX)

			if g.override {
				overriddenContext = g.overrideAtTemplate2(context, x+minorX, lineNumber, int(result), minorX)
				g.cx.SetIndex(int32(overriddenContext))
			} else {
				g.cx.SetIndex(int32(context))
			}

			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			result |= byte(bit << uint(7-minorX))
			context = ((context & 0x1bd) << 1) | bit | ((line1 >> toShift) & 0x4) | ((line2 >> toShift) & 0x80)
		}

		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return errors.Wrap(err, processName, "")
		}

		byteIndex++
		idx++
	}
	return nil
}

func (g *GenericRegion) decodeTemplate3(line, width, paddedWidth int, byteIndex, idx int) (err error) {
	const processName = "decodeTemplate3"
	var (
		context, overriddenContext int
		line1                      int
		temp                       byte
		nextByte, bit              int
	)

	if line >= 1 {
		temp, err = g.Bitmap.GetByte(idx)
		if err != nil {
			return errors.Wrap(err, processName, "line >= 1")
		}
		line1 = int(temp)
	}

	context = (line1 >> 1) & 0x70

	for x := 0; x < paddedWidth; x = nextByte {
		// 6.2.5.7 3d
		var (
			result     byte
			minorWidth int
		)
		nextByte = x + 8

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
					return errors.Wrap(err, processName, "inner - line >= 1")
				}
				line1 |= int(temp)
			}
		}

		for minorX := 0; minorX < minorWidth; minorX++ {
			if g.override {
				overriddenContext = g.overrideAtTemplate3(context, x+minorX, line, int(result), minorX)
				g.cx.SetIndex(int32(overriddenContext))
			} else {
				g.cx.SetIndex(int32(context))
			}

			bit, err = g.arithDecoder.DecodeBit(g.cx)
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			result |= byte(bit) << byte(7-minorX)
			context = ((context & 0x1f7) << 1) | bit | ((line1 >> uint(8-minorX)) & 0x010)
		}

		if err := g.Bitmap.SetByte(byteIndex, result); err != nil {
			return errors.Wrap(err, processName, "")
		}

		byteIndex++
		idx++
	}
	return nil
}

func (g *GenericRegion) getPixel(x, y int) int8 {
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

func (g *GenericRegion) updateOverrideFlags() error {
	const processName = "updateOverrideFlags"
	if g.GBAtX == nil || g.GBAtY == nil {
		return nil
	}

	if len(g.GBAtX) != len(g.GBAtY) {
		return errors.Errorf(processName, "incosistent AT pixel. Amount of 'x' pixels: %d, Amount of 'y' pixels: %d", len(g.GBAtX), len(g.GBAtY))
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

func (g *GenericRegion) overrideAtTemplate0a(ctx, x, y, result, minorX, toShift int) int {
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

func (g *GenericRegion) overrideAtTemplate0b(ctx, x, y, result, minorX, toShift int) int {
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

func (g *GenericRegion) overrideAtTemplate1(ctx, x, y, result, minorX int) int {
	ctx &= 0x1FF7
	if g.GBAtY[0] == 0 && g.GBAtX[0] >= -int8(minorX) {
		ctx |= (result >> uint(7-(int8(minorX)+g.GBAtX[0])) & 0x1) << 3
	} else {
		ctx |= int(g.getPixel(x+int(g.GBAtX[0]), y+int(g.GBAtY[0]))) << 3
	}
	return ctx
}

func (g *GenericRegion) overrideAtTemplate2(ctx, x, y, result, minorX int) int {
	ctx &= 0x3FB
	if g.GBAtY[0] == 0 && g.GBAtX[0] >= -int8(minorX) {
		ctx |= (result >> uint(7-(int8(minorX)+g.GBAtX[0])) & 0x1) << 2
	} else {
		ctx |= int(g.getPixel(x+int(g.GBAtX[0]), y+int(g.GBAtY[0]))) << 2
	}
	return ctx
}

func (g *GenericRegion) overrideAtTemplate3(ctx, x, y, result, minorX int) int {
	ctx &= 0x3EF
	if g.GBAtY[0] == 0 && g.GBAtX[0] >= -int8(minorX) {
		ctx |= (result >> uint(7-(int8(minorX)+g.GBAtX[0])) & 0x1) << 4
	} else {
		ctx |= int(g.getPixel(x+int(g.GBAtX[0]), y+int(g.GBAtY[0]))) << 4
	}
	return ctx
}

func (g *GenericRegion) readGBAtPixels(numberOfGbAt int) error {
	const processName = "readGBAtPixels"
	g.GBAtX = make([]int8, numberOfGbAt)
	g.GBAtY = make([]int8, numberOfGbAt)

	for i := 0; i < numberOfGbAt; i++ {
		b, err := g.r.ReadByte()
		if err != nil {
			return errors.Wrapf(err, processName, "X at i: '%d'", i)
		}

		g.GBAtX[i] = int8(b)

		b, err = g.r.ReadByte()
		if err != nil {
			return errors.Wrapf(err, processName, "Y at i: '%d'", i)
		}
		g.GBAtY[i] = int8(b)
	}
	return nil
}

func (g *GenericRegion) setOverrideFlag(index int) {
	g.GBAtOverride[index] = true
	g.override = true
}

func (g *GenericRegion) setParameters(isMMREncoded bool, dataOffset, dataLength int64, gbh, gbw uint32) {
	g.IsMMREncoded = isMMREncoded
	g.DataOffset = dataOffset
	g.DataLength = dataLength
	g.RegionSegment.BitmapHeight = gbh
	g.RegionSegment.BitmapWidth = gbw
	g.mmrDecoder = nil
	g.Bitmap = nil
}

func (g *GenericRegion) setParametersWithAt(
	isMMREncoded bool,
	sdTemplate byte,
	isTPGDon, useSkip bool,
	sDAtX, sDAtY []int8,
	symWidth, hcHeight uint32,
	cx *arithmetic.DecoderStats, a *arithmetic.Decoder,
) {
	g.IsMMREncoded = isMMREncoded
	g.GBTemplate = sdTemplate
	g.IsTPGDon = isTPGDon
	g.GBAtX = sDAtX
	g.GBAtY = sDAtY
	g.RegionSegment.BitmapHeight = hcHeight
	g.RegionSegment.BitmapWidth = symWidth
	g.mmrDecoder = nil
	g.Bitmap = nil

	if cx != nil {
		g.cx = cx
	}

	if a != nil {
		g.arithDecoder = a
	}
	common.Log.Trace("[GENERIC-REGION] setParameters SDAt: %s", g)
}

func (g *GenericRegion) setParametersMMR(
	isMMREncoded bool,
	dataOffset, dataLength int64,
	gbh, gbw uint32,
	gbTemplate byte,
	isTPGDon, useSkip bool,
	gbAtX, gbAtY []int8,
) {
	g.DataOffset = dataOffset
	g.DataLength = dataLength
	g.RegionSegment = &RegionSegment{}
	g.RegionSegment.BitmapHeight = gbh
	g.RegionSegment.BitmapWidth = gbw
	g.GBTemplate = gbTemplate
	g.IsMMREncoded = isMMREncoded
	g.IsTPGDon = isTPGDon
	g.GBAtX = gbAtX
	g.GBAtY = gbAtY

}

func (g *GenericRegion) writeGBAtPixels(w writer.BinaryWriter) (n int, err error) {
	const processName = "writeGBAtPixels"
	if g.UseMMR {
		return 0, nil
	}
	pairNumber := 1
	if g.GBTemplate == 0 {
		pairNumber = 4
	} else if g.UseExtTemplates {
		pairNumber = 12
	}
	if len(g.GBAtX) != pairNumber {
		return 0, errors.Errorf(processName, "gb at pair number doesn't match to GBAtX slice len")
	}
	if len(g.GBAtY) != pairNumber {
		return 0, errors.Errorf(processName, "gb at pair number doesn't match to GBAtY slice len")
	}

	// write all the at pixels to te 'w' writer
	for i := 0; i < pairNumber; i++ {
		if err = w.WriteByte(byte(g.GBAtX[i])); err != nil {
			return n, errors.Wrap(err, processName, "write GBAtX")
		}
		n++
		if err = w.WriteByte(byte(g.GBAtY[i])); err != nil {
			return n, errors.Wrap(err, processName, "write GBAtY")
		}
		n++
	}
	return n, nil
}
