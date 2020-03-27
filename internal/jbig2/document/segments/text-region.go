/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/huffman"
	encoder "github.com/unidoc/unipdf/v3/internal/jbig2/encoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// TextRegion is the model for the jbig2 text region segment - see 7.4.1.
type TextRegion struct {
	r reader.StreamReader

	// Region segment information field 7.4.1.
	RegionInfo *RegionSegment

	// Text Region segment flags 7.4.3.1.1.
	SbrTemplate         int8
	SbdsOffset          int8
	DefaultPixel        int8
	CombinationOperator bitmap.CombinationOperator
	IsTransposed        int8
	ReferenceCorner     int16
	LogSBStrips         int16
	UseRefinement       bool
	IsHuffmanEncoded    bool

	// Text region segment huffman flags 7.4.3.1.2.
	SbHuffRSize    int8
	SbHuffRDY      int8
	SbHuffRDX      int8
	SbHuffRDHeight int8
	SbHuffRDWidth  int8
	SbHuffDT       int8
	SbHuffDS       int8
	SbHuffFS       int8

	// Text region refinement AT flags 7.4.3.1.3.
	SbrATX []int8
	SbrATY []int8

	// Number of symbol instances 7.4.3.1.3.
	NumberOfSymbolInstances uint32

	currentS        int64
	SbStrips        int8
	NumberOfSymbols uint32

	RegionBitmap *bitmap.Bitmap
	Symbols      []*bitmap.Bitmap

	arithmDecoder           *arithmetic.Decoder
	genericRefinementRegion *GenericRefinementRegion

	cxIADT  *arithmetic.DecoderStats
	cxIAFS  *arithmetic.DecoderStats
	cxIADS  *arithmetic.DecoderStats
	cxIAIT  *arithmetic.DecoderStats
	cxIARI  *arithmetic.DecoderStats
	cxIARDW *arithmetic.DecoderStats
	cxIARDH *arithmetic.DecoderStats
	cxIAID  *arithmetic.DecoderStats
	cxIARDX *arithmetic.DecoderStats
	cxIARDY *arithmetic.DecoderStats
	cx      *arithmetic.DecoderStats

	// symbolCodeTable includes a code to each symbol used in that region.
	symbolCodeLength int8
	symbolCodeTable  *huffman.FixedSizeTable
	Header           *Header

	fsTable    huffman.Tabler
	dsTable    huffman.Tabler
	table      huffman.Tabler
	rdwTable   huffman.Tabler
	rdhTable   huffman.Tabler
	rdxTable   huffman.Tabler
	rdyTable   huffman.Tabler
	rSizeTable huffman.Tabler

	// encoding context variables
	globalSymbolsMap, localSymbolsMap map[int]int
	// componentNumbers is the slice of components used within this TextRegion
	componentNumbers []int
	// is a slice of lower-left component points
	inLL *bitmap.Points
	// is a slice of symbols
	symbols *bitmap.Bitmaps
	// assignments contain the 'symbols' indexes based on the symbol number from components.
	assignments         *basic.IntSlice
	stripWidth, symBits int

	boxes *bitmap.Boxes
}

// Compile time checks for the TextRegion interfaces implementation.
var (
	_ Regioner  = &TextRegion{}
	_ Segmenter = &TextRegion{}
)

// Encode writes the TextRegion segment data into 'w' binary writer.
func (t *TextRegion) Encode(w writer.BinaryWriter) (n int, err error) {
	const processName = "TextRegion.Encode"
	// region info
	if n, err = t.RegionInfo.Encode(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}

	// write flags
	var temp int
	if temp, err = t.encodeFlags(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += temp

	// write huffman flags
	// not implemnted

	// write refinement flags
	// not implemented
	if temp, err = t.encodeSymbols(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += temp
	return n, nil
}

// InitEncode initializes text region for the Encode method.
func (t *TextRegion) InitEncode(globalSymbolsMap, localSymbolsMap map[int]int, comps []int, inLL *bitmap.Points, symbols *bitmap.Bitmaps, classIDs *basic.IntSlice, boxes *bitmap.Boxes, width, height, symbits int) {
	t.RegionInfo = &RegionSegment{
		BitmapWidth:  uint32(width),
		BitmapHeight: uint32(height),
	}
	t.globalSymbolsMap = globalSymbolsMap
	t.localSymbolsMap = localSymbolsMap
	t.componentNumbers = comps
	t.inLL = inLL
	t.symbols = symbols
	t.assignments = classIDs
	t.boxes = boxes
	t.symBits = symbits
}

// GetRegionBitmap implements Regioner interface.
func (t *TextRegion) GetRegionBitmap() (*bitmap.Bitmap, error) {
	if t.RegionBitmap != nil {
		return t.RegionBitmap, nil
	}

	if !t.IsHuffmanEncoded {
		if err := t.setCodingStatistics(); err != nil {
			return nil, err
		}
	}

	if err := t.createRegionBitmap(); err != nil {
		return nil, err
	}

	if err := t.decodeSymbolInstances(); err != nil {
		return nil, err
	}
	return t.RegionBitmap, nil
}

// GetRegionInfo implements Regioner interface.
func (t *TextRegion) GetRegionInfo() *RegionSegment {
	return t.RegionInfo
}

// Init implements Segmenter interface.
func (t *TextRegion) Init(header *Header, r reader.StreamReader) error {
	t.Header = header
	t.r = r
	t.RegionInfo = NewRegionSegment(t.r)
	return t.parseHeader()
}

// String implements the Stringer interface.
func (t *TextRegion) String() string {
	sb := &strings.Builder{}
	sb.WriteString("\n[TEXT REGION]\n")
	sb.WriteString(t.RegionInfo.String() + "\n")
	sb.WriteString(fmt.Sprintf("\t- SbrTemplate: %v\n", t.SbrTemplate))
	sb.WriteString(fmt.Sprintf("\t- SbdsOffset: %v\n", t.SbdsOffset))
	sb.WriteString(fmt.Sprintf("\t- DefaultPixel: %v\n", t.DefaultPixel))
	sb.WriteString(fmt.Sprintf("\t- CombinationOperator: %v\n", t.CombinationOperator))
	sb.WriteString(fmt.Sprintf("\t- IsTransposed: %v\n", t.IsTransposed))
	sb.WriteString(fmt.Sprintf("\t- ReferenceCorner: %v\n", t.ReferenceCorner))
	sb.WriteString(fmt.Sprintf("\t- UseRefinement: %v\n", t.UseRefinement))
	sb.WriteString(fmt.Sprintf("\t- IsHuffmanEncoded: %v\n", t.IsHuffmanEncoded))

	if t.IsHuffmanEncoded {
		sb.WriteString(fmt.Sprintf("\t- SbHuffRSize: %v\n", t.SbHuffRSize))
		sb.WriteString(fmt.Sprintf("\t- SbHuffRDY: %v\n", t.SbHuffRDY))
		sb.WriteString(fmt.Sprintf("\t- SbHuffRDX: %v\n", t.SbHuffRDX))
		sb.WriteString(fmt.Sprintf("\t- SbHuffRDHeight: %v\n", t.SbHuffRDHeight))
		sb.WriteString(fmt.Sprintf("\t- SbHuffRDWidth: %v\n", t.SbHuffRDWidth))
		sb.WriteString(fmt.Sprintf("\t- SbHuffDT: %v\n", t.SbHuffDT))
		sb.WriteString(fmt.Sprintf("\t- SbHuffDS: %v\n", t.SbHuffDS))
		sb.WriteString(fmt.Sprintf("\t- SbHuffFS: %v\n", t.SbHuffFS))
	}

	sb.WriteString(fmt.Sprintf("\t- SbrATX: %v\n", t.SbrATX))
	sb.WriteString(fmt.Sprintf("\t- SbrATY: %v\n", t.SbrATY))
	sb.WriteString(fmt.Sprintf("\t- NumberOfSymbolInstances: %v\n", t.NumberOfSymbolInstances))
	sb.WriteString(fmt.Sprintf("\t- SbrATX: %v\n", t.SbrATX))
	return sb.String()
}

func (t *TextRegion) blit(ib *bitmap.Bitmap, tc int64) error {
	if t.IsTransposed == 0 && (t.ReferenceCorner == 2 || t.ReferenceCorner == 3) {
		t.currentS += int64(ib.Width - 1)
	} else if t.IsTransposed == 1 && (t.ReferenceCorner == 0 || t.ReferenceCorner == 2) {
		t.currentS += int64(ib.Height - 1)
	}

	// VII)
	s := t.currentS

	// VIII)
	if t.IsTransposed == 1 {
		s, tc = tc, s
	}

	switch t.ReferenceCorner {
	case 0:
		// BL
		tc -= int64(ib.Height - 1)
	case 2:
		// BR
		tc -= int64(ib.Height - 1)
		s -= int64(ib.Width - 1)
	case 3:
		// TR
		s -= int64(ib.Width - 1)
	}

	err := bitmap.Blit(ib, t.RegionBitmap, int(s), int(tc), t.CombinationOperator)
	if err != nil {
		return err
	}

	// X)
	if t.IsTransposed == 0 && (t.ReferenceCorner == 0 || t.ReferenceCorner == 1) {
		t.currentS += int64(ib.Width - 1)
	}

	if t.IsTransposed == 1 && (t.ReferenceCorner == 1 || t.ReferenceCorner == 3) {
		t.currentS += int64(ib.Height - 1)
	}
	return nil
}

func (t *TextRegion) computeSymbolCodeLength() error {
	if t.IsHuffmanEncoded {
		return t.symbolIDCodeLengths()
	}

	t.symbolCodeLength = int8(math.Ceil(math.Log(float64(t.NumberOfSymbols)) / math.Log(2)))
	return nil
}

func (t *TextRegion) checkInput() error {
	const processName = "checkInput"
	if !t.UseRefinement {
		if t.SbrTemplate != 0 {
			common.Log.Debug("SbrTemplate should be 0")
			t.SbrTemplate = 0
		}
	}

	if t.SbHuffFS == 2 || t.SbHuffRDWidth == 2 || t.SbHuffRDHeight == 2 || t.SbHuffRDX == 2 || t.SbHuffRDY == 2 {
		return errors.Error(processName, "huffman flag value is not permitted")
	}

	if !t.UseRefinement {
		if t.SbHuffRSize != 0 {
			common.Log.Debug("SbHuffRSize should be 0")
			t.SbHuffRSize = 0
		}

		if t.SbHuffRDY != 0 {
			common.Log.Debug("SbHuffRDY should be 0")
			t.SbHuffRDY = 0
		}

		if t.SbHuffRDX != 0 {
			common.Log.Debug("SbHuffRDX should be 0")
			t.SbHuffRDX = 0
		}

		if t.SbHuffRDWidth != 0 {
			common.Log.Debug("SbHuffRDWidth should be 0")
			t.SbHuffRDWidth = 0
		}

		if t.SbHuffRDHeight != 0 {
			common.Log.Debug("SbHuffRDHeight should be 0")
			t.SbHuffRDHeight = 0
		}
	}
	return nil
}

func (t *TextRegion) createRegionBitmap() error {
	// 6.4.5
	t.RegionBitmap = bitmap.New(int(t.RegionInfo.BitmapWidth), int(t.RegionInfo.BitmapHeight))
	if t.DefaultPixel != 0 {
		t.RegionBitmap.SetDefaultPixel()
	}
	return nil
}

func (t *TextRegion) decodeStripT() (stripT int64, err error) {
	// 2)
	if t.IsHuffmanEncoded {
		// 6.4.6
		if t.SbHuffDT == 3 {
			if t.table == nil {
				var dtNr int
				if t.SbHuffFS == 3 {
					dtNr++
				}

				if t.SbHuffDS == 3 {
					dtNr++
				}

				t.table, err = t.getUserTable(dtNr)
				if err != nil {
					return 0, err
				}
			}

			stripT, err = t.table.Decode(t.r)
			if err != nil {
				return 0, err
			}
		} else {
			var table huffman.Tabler
			table, err = huffman.GetStandardTable(11 + int(t.SbHuffDT))
			if err != nil {
				return 0, err
			}

			stripT, err = table.Decode(t.r)
			if err != nil {
				return 0, err
			}
		}
	} else {
		var temp int32
		temp, err = t.arithmDecoder.DecodeInt(t.cxIADT)
		if err != nil {
			return 0, err
		}
		stripT = int64(temp)
	}
	stripT *= int64(-t.SbStrips)
	return stripT, nil
}

func (t *TextRegion) decodeSymbolInstances() error {
	stripT, err := t.decodeStripT()
	if err != nil {
		return err
	}

	// Last two sentences of 6.4.5 2)
	var (
		firstS          int64
		instanceCounter uint32
	)

	// 6.4.5 3)
	for instanceCounter < t.NumberOfSymbolInstances {
		dt, err := t.decodeDT()
		if err != nil {
			return err
		}

		stripT += dt

		// 3 c) symbol instances in the strip
		var dfs int64
		first := true
		t.currentS = 0

		// loop until OOB
		for {
			if first {
				// 6.4.7
				dfs, err = t.decodeDfs()
				if err != nil {
					return err
				}

				firstS += dfs
				t.currentS = firstS
				first = false
				// 3 c) II) - the remaining symbol instances in the strip
			} else {
				// 6.4.8
				idS, err := t.decodeIds()
				if err != nil {
					return err
				}

				if idS == math.MaxInt32 || instanceCounter >= t.NumberOfSymbolInstances {
					break
				}

				t.currentS += idS + int64(t.SbdsOffset)
			}
			// 3 c) III)
			currentT, err := t.decodeCurrentT()
			if err != nil {
				return err
			}

			tt := stripT + currentT

			// 3 c) IV)
			id, err := t.decodeID()
			if err != nil {
				return err
			}

			// 3 c) V)
			r, err := t.decodeRI()
			if err != nil {
				return err
			}

			// 6.4.11
			ib, err := t.decodeIb(r, id)
			if err != nil {
				return err
			}
			if err = t.blit(ib, tt); err != nil {
				return err
			}

			instanceCounter++
		}
	}
	return nil
}

func (t *TextRegion) decodeDT() (dT int64, err error) {
	// 3) b)
	// 6.4.6
	if t.IsHuffmanEncoded {
		if t.SbHuffDT == 3 {
			dT, err = t.table.Decode(t.r)
			if err != nil {
				return 0, err
			}
		} else {
			var st huffman.Tabler
			st, err = huffman.GetStandardTable(11 + int(t.SbHuffDT))
			if err != nil {
				return 0, err
			}

			dT, err = st.Decode(t.r)
			if err != nil {
				return 0, err
			}
		}
	} else {
		var temp int32
		temp, err = t.arithmDecoder.DecodeInt(t.cxIADT)
		if err != nil {
			return
		}
		dT = int64(temp)
	}
	dT *= int64(t.SbStrips)
	return dT, nil
}

func (t *TextRegion) decodeDfs() (int64, error) {
	if t.IsHuffmanEncoded {
		if t.SbHuffFS == 3 {
			if t.fsTable == nil {
				var err error
				t.fsTable, err = t.getUserTable(0)
				if err != nil {
					return 0, err
				}
			}
			return t.fsTable.Decode(t.r)
		}
		st, err := huffman.GetStandardTable(6 + int(t.SbHuffFS))
		if err != nil {
			return 0, err
		}
		return st.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIAFS)
	if err != nil {
		return 0, err
	}
	return int64(temp), nil
}

func (t *TextRegion) decodeCurrentT() (int64, error) {
	if t.SbStrips != 1 {
		if t.IsHuffmanEncoded {
			bits, err := t.r.ReadBits(byte(t.LogSBStrips))
			return int64(bits), err
		}
		temp, err := t.arithmDecoder.DecodeInt(t.cxIAIT)
		if err != nil {
			return 0, err
		}
		return int64(temp), nil
	}
	return 0, nil
}

func (t *TextRegion) decodeID() (int64, error) {
	if t.IsHuffmanEncoded {
		if t.symbolCodeTable == nil {
			bits, err := t.r.ReadBits(byte(t.symbolCodeLength))
			return int64(bits), err
		}
		return t.symbolCodeTable.Decode(t.r)
	}
	return t.arithmDecoder.DecodeIAID(uint64(t.symbolCodeLength), t.cxIAID)
}

func (t *TextRegion) decodeRI() (int64, error) {
	if !t.UseRefinement {
		return 0, nil
	}

	if t.IsHuffmanEncoded {
		temp, err := t.r.ReadBit()
		return int64(temp), err
	}

	temp, err := t.arithmDecoder.DecodeInt(t.cxIARI)
	return int64(temp), err
}

func (t *TextRegion) decodeIb(r, id int64) (*bitmap.Bitmap, error) {
	const processName = "decodeIb"
	var (
		err error
		ib  *bitmap.Bitmap
	)

	if r == 0 {
		if int(id) > len(t.Symbols)-1 {
			return nil, errors.Error(processName, "decoding IB bitmap. index out of range")
		}
		return t.Symbols[int(id)], nil
	}
	// 1) - 4)
	var rdw, rdh, rdx, rdy int64

	rdw, err = t.decodeRdw()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	rdh, err = t.decodeRdh()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	rdx, err = t.decodeRdx()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	rdy, err = t.decodeRdy()
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	// 5)
	if t.IsHuffmanEncoded {
		if _, err = t.decodeSymInRefSize(); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		t.r.Align()
	}

	// 6)
	ibo := t.Symbols[id]
	wo := uint32(ibo.Width)
	ho := uint32(ibo.Height)
	genericRegionReferenceDX := int32(uint32(rdw)>>1) + int32(rdx)
	genericRegionReferenceDY := int32(uint32(rdh)>>1) + int32(rdy)

	if t.genericRefinementRegion == nil {
		t.genericRefinementRegion = newGenericRefinementRegion(t.r, nil)
	}

	t.genericRefinementRegion.setParameters(t.cx, t.arithmDecoder, t.SbrTemplate,
		wo+uint32(rdw), ho+uint32(rdh), ibo, genericRegionReferenceDX, genericRegionReferenceDY, false, t.SbrATX, t.SbrATY)

	ib, err = t.genericRefinementRegion.GetRegionBitmap()
	if err != nil {
		return nil, errors.Wrap(err, processName, "grf")
	}

	// 7)
	if t.IsHuffmanEncoded {
		t.r.Align()
	}
	return ib, nil
}

func (t *TextRegion) decodeIds() (int64, error) {
	const processName = "decodeIds"
	if t.IsHuffmanEncoded {
		if t.SbHuffDS == 3 {
			if t.dsTable == nil {
				dsNr := 0
				if t.SbHuffFS == 3 {
					dsNr++
				}

				var err error
				t.dsTable, err = t.getUserTable(dsNr)
				if err != nil {
					return 0, errors.Wrap(err, processName, "")
				}
			}
			return t.dsTable.Decode(t.r)
		}
		st, err := huffman.GetStandardTable(8 + int(t.SbHuffDS))
		if err != nil {
			return 0, errors.Wrap(err, processName, "")
		}
		return st.Decode(t.r)
	}

	i, err := t.arithmDecoder.DecodeInt(t.cxIADS)
	if err != nil {
		return 0, errors.Wrap(err, processName, "cxIADS")
	}
	return int64(i), nil
}

func (t *TextRegion) decodeRdw() (int64, error) {
	const processName = "decodeRdw"
	if t.IsHuffmanEncoded {
		if t.SbHuffRDWidth == 3 {
			if t.rdwTable == nil {
				var (
					rdwNr int
					err   error
				)

				if t.SbHuffFS == 3 {
					rdwNr++
				}

				if t.SbHuffDS == 3 {
					rdwNr++
				}

				if t.SbHuffDT == 3 {
					rdwNr++
				}

				t.rdwTable, err = t.getUserTable(rdwNr)
				if err != nil {
					return 0, errors.Wrap(err, processName, "")
				}
			}
			return t.rdwTable.Decode(t.r)
		}
		ts, err := huffman.GetStandardTable(14 + int(t.SbHuffRDWidth))
		if err != nil {
			return 0, errors.Wrap(err, processName, "")
		}
		return ts.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIARDW)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	return int64(temp), nil
}

func (t *TextRegion) decodeRdh() (int64, error) {
	const processName = "decodeRdh"
	if t.IsHuffmanEncoded {
		if t.SbHuffRDHeight == 3 {
			if t.rdhTable == nil {
				var (
					rdhNr int
					err   error
				)

				if t.SbHuffFS == 3 {
					rdhNr++
				}

				if t.SbHuffDS == 3 {
					rdhNr++
				}

				if t.SbHuffDT == 3 {
					rdhNr++
				}

				if t.SbHuffRDWidth == 3 {
					rdhNr++
				}

				t.rdhTable, err = t.getUserTable(rdhNr)
				if err != nil {
					return 0, errors.Wrap(err, processName, "")
				}
			}
			return t.rdhTable.Decode(t.r)
		}
		ts, err := huffman.GetStandardTable(14 + int(t.SbHuffRDHeight))
		if err != nil {
			return 0, errors.Wrap(err, processName, "")
		}

		return ts.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIARDH)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	return int64(temp), nil
}

func (t *TextRegion) decodeRdx() (int64, error) {
	const processName = "decodeRdx"
	if t.IsHuffmanEncoded {
		if t.SbHuffRDX == 3 {
			if t.rdxTable == nil {
				var (
					rdxNr int
					err   error
				)

				if t.SbHuffFS == 3 {
					rdxNr++
				}

				if t.SbHuffDS == 3 {
					rdxNr++
				}

				if t.SbHuffDT == 3 {
					rdxNr++
				}

				if t.SbHuffRDWidth == 3 {
					rdxNr++
				}

				if t.SbHuffRDHeight == 3 {
					rdxNr++
				}

				t.rdxTable, err = t.getUserTable(rdxNr)
				if err != nil {
					return 0, errors.Wrap(err, processName, "")
				}
			}
			return t.rdxTable.Decode(t.r)
		}
		ts, err := huffman.GetStandardTable(14 + int(t.SbHuffRDX))
		if err != nil {
			return 0, errors.Wrap(err, processName, "")
		}
		return ts.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIARDX)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	return int64(temp), nil
}

func (t *TextRegion) decodeRdy() (int64, error) {
	const processName = "decodeRdy"
	if t.IsHuffmanEncoded {
		if t.SbHuffRDY == 3 {
			if t.rdyTable == nil {
				var (
					rdyNr int
					err   error
				)

				if t.SbHuffFS == 3 {
					rdyNr++
				}

				if t.SbHuffDS == 3 {
					rdyNr++
				}

				if t.SbHuffDT == 3 {
					rdyNr++
				}

				if t.SbHuffRDWidth == 3 {
					rdyNr++
				}

				if t.SbHuffRDHeight == 3 {
					rdyNr++
				}

				if t.SbHuffRDX == 3 {
					rdyNr++
				}

				t.rdyTable, err = t.getUserTable(rdyNr)
				if err != nil {
					return 0, errors.Wrap(err, processName, "")
				}
			}
			return t.rdyTable.Decode(t.r)
		}
		ts, err := huffman.GetStandardTable(14 + int(t.SbHuffRDY))
		if err != nil {
			return 0, err
		}
		return ts.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIARDY)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	return int64(temp), nil
}

func (t *TextRegion) decodeSymInRefSize() (int64, error) {
	const processName = "decodeSymInRefSize"
	if t.SbHuffRSize == 0 {
		ts, err := huffman.GetStandardTable(1)
		if err != nil {
			return 0, errors.Wrap(err, processName, "")
		}
		return ts.Decode(t.r)
	}
	if t.rSizeTable == nil {
		var (
			rSizeNr int
			err     error
		)

		if t.SbHuffFS == 3 {
			rSizeNr++
		}

		if t.SbHuffDS == 3 {
			rSizeNr++
		}

		if t.SbHuffDT == 3 {
			rSizeNr++
		}

		if t.SbHuffRDWidth == 3 {
			rSizeNr++
		}

		if t.SbHuffRDHeight == 3 {
			rSizeNr++
		}

		if t.SbHuffRDX == 3 {
			rSizeNr++
		}

		if t.SbHuffRDY == 3 {
			rSizeNr++
		}

		t.rSizeTable, err = t.getUserTable(rSizeNr)
		if err != nil {
			return 0, errors.Wrap(err, processName, "")
		}
	}
	temp, err := t.rSizeTable.Decode(t.r)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	return temp, nil
}

func (t *TextRegion) encodeFlags(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeFlags"
	if err = w.WriteBit(int(t.SbrTemplate)); err != nil {
		return n, errors.Wrap(err, processName, "sbrTemplate")
	}
	if _, err = w.WriteBits(uint64(t.SbdsOffset), 5); err != nil {
		return n, errors.Wrap(err, processName, "sbdsOffset")
	}
	if err = w.WriteBit(int(t.DefaultPixel)); err != nil {
		return n, errors.Wrap(err, processName, "DefaultPixel")
	}
	if _, err = w.WriteBits(uint64(t.CombinationOperator), 2); err != nil {
		return n, errors.Wrap(err, processName, "CombinationOperator")
	}
	if err = w.WriteBit(int(t.IsTransposed)); err != nil {
		return n, errors.Wrap(err, processName, "is transposed")
	}
	if _, err = w.WriteBits(uint64(t.ReferenceCorner), 2); err != nil {
		return n, errors.Wrap(err, processName, "reference corner")
	}
	if _, err = w.WriteBits(uint64(t.LogSBStrips), 2); err != nil {
		return n, errors.Wrap(err, processName, "LogSBStrips")
	}
	var bit int
	if t.UseRefinement {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "use refinement")
	}
	bit = 0
	if t.IsHuffmanEncoded {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "use huffman")
	}
	n = 2
	return n, nil
}

func (t *TextRegion) encodeSymbols(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeSymbols"
	// store the number of new symbols value.
	tm := make([]byte, 4)
	binary.BigEndian.PutUint32(tm, t.NumberOfSymbols)
	if n, err = w.Write(tm); err != nil {
		return n, errors.Wrap(err, processName, "NumberOfSymbolInstances")
	}

	// write segment symbol id huffman decoded table
	// not implemented
	symPoints, err := bitmap.NewClassedPoints(t.inLL, basic.IntSlice(t.componentNumbers))
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}

	var stripeT, firsts int
	// initialize encode context
	encodeCtx := encoder.New()
	encodeCtx.Init()

	// begin with encoding IADT.
	if err = encodeCtx.EncodeInteger(encoder.IADT, 0); err != nil {
		return n, errors.Wrap(err, processName, "initial DT")
	}

	// group the points by their 'y' coordinate.
	// and treat as the stripes.
	stripes, err := symPoints.GroupByY()
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	// iterate over the stripes with different 'y' position of their lower left point.
	for _, stripe := range stripes {
		stripeY := int(stripe.YAtIndex(0))
		deltaT := stripeY - stripeT

		// encode the height difference
		if err = encodeCtx.EncodeInteger(encoder.IADT, deltaT); err != nil {
			return n, errors.Wrap(err, processName, "")
		}

		// deltaS is the difference in the 'x' value between the symbols.
		var currentS int
		// iterate over the symbols in the stripe
		for i, symbol := range stripe.IntSlice {
			switch i {
			case 0:
				// the first symbol is encoded with the 'IAFS'
				deltaFS := int(stripe.XAtIndex(i)) - firsts
				if err = encodeCtx.EncodeInteger(encoder.IAFS, deltaFS); err != nil {
					return n, errors.Wrap(err, processName, "")
				}
				firsts += deltaFS
				currentS = firsts
			default:
				// all other symbols in the stripe are encoded using 'IADS'
				// encode only the difference between the last symbol 'x' and this one.
				deltaS := int(stripe.XAtIndex(i)) - currentS
				if err = encodeCtx.EncodeInteger(encoder.IADS, deltaS); err != nil {
					return n, errors.Wrap(err, processName, "")
				}
				currentS += deltaS
			}
			// get the assigned symbol index
			assigned, err := t.assignments.Get(symbol)
			if err != nil {
				return n, errors.Wrap(err, processName, "")
			}

			// try to find the symbol in the global map
			symbolID, ok := t.globalSymbolsMap[assigned]
			if !ok {
				// and in the local map
				symbolID, ok = t.localSymbolsMap[assigned]
				if !ok {
					return n, errors.Errorf(processName, "Symobl: '%d' is not found in global and local symbol map", assigned)
				}
			}
			// encode the symbol id.
			if err = encodeCtx.EncodeIAID(t.symBits, symbolID); err != nil {
				return n, errors.Wrap(err, processName, "")
			}
		}

		// terminate the strip with the OOB
		if err = encodeCtx.EncodeOOB(encoder.IADS); err != nil {
			return n, errors.Wrap(err, processName, "")
		}
	}
	encodeCtx.Final()
	total, err := encodeCtx.WriteTo(w)
	if err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += int(total)
	return n, nil
}

func (t *TextRegion) getSymbols() error {
	if t.Header.RTSegments != nil {
		return t.initSymbols()
	}
	return nil
}

func (t *TextRegion) getUserTable(tablePosition int) (huffman.Tabler, error) {
	const processName = "getUserTable"
	var tableCounter int
	for _, rts := range t.Header.RTSegments {
		if rts.Type == 53 {
			if tableCounter == tablePosition {
				sd, err := rts.GetSegmentData()
				if err != nil {
					return nil, err
				}

				ts, ok := sd.(*TableSegment)
				if !ok {
					common.Log.Debug(fmt.Sprintf("segment with Type 53 - and index: %d not a TableSegment", rts.SegmentNumber))
					return nil, errors.Error(processName, "segment with Type 53 is not a *TableSegment")
				}
				return huffman.NewEncodedTable(ts)
			}
			tableCounter++
		}
	}
	return nil, nil
}

func (t *TextRegion) initSymbols() error {
	const processName = "initSymbols"
	for _, segment := range t.Header.RTSegments {
		if segment == nil {
			return errors.Error(processName, "nil segment provided for the text region Symbols")
		}

		if segment.Type == 0 {
			s, err := segment.GetSegmentData()
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			sd, ok := s.(*SymbolDictionary)
			if !ok {
				return errors.Error(processName, "referred To Segment is not a SymbolDictionary")
			}

			sd.cxIAID = t.cxIAID
			dict, err := sd.GetDictionary()
			if err != nil {
				return errors.Wrap(err, processName, "")
			}

			t.Symbols = append(t.Symbols, dict...)
		}
	}
	t.NumberOfSymbols = uint32(len(t.Symbols))
	return nil
}

func (t *TextRegion) parseHeader() error {
	var err error
	common.Log.Trace("[TEXT REGION][PARSE-HEADER] begins...")
	defer func() {
		if err != nil {
			common.Log.Trace("[TEXT REGION][PARSE-HEADER] failed. %v", err)
		} else {
			common.Log.Trace("[TEXT REGION][PARSE-HEADER] finished.")
		}
	}()

	if err = t.RegionInfo.parseHeader(); err != nil {
		return err
	}

	if err = t.readRegionFlags(); err != nil {
		return err
	}

	if t.IsHuffmanEncoded {
		if err = t.readHuffmanFlags(); err != nil {
			return err
		}
	}

	if err = t.readUseRefinement(); err != nil {
		return err
	}

	if err = t.readAmountOfSymbolInstances(); err != nil {
		return err
	}

	// 7.4.3.1.7
	if err = t.getSymbols(); err != nil {
		return err
	}

	if err = t.computeSymbolCodeLength(); err != nil {
		return err
	}

	if err = t.checkInput(); err != nil {
		return err
	}
	common.Log.Trace("%s", t.String())
	return nil
}

func (t *TextRegion) readRegionFlags() error {
	var (
		bit  int
		bits uint64
		err  error
	)

	// Bit 15
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}
	t.SbrTemplate = int8(bit)

	// Bit 10 - 14
	bits, err = t.r.ReadBits(5)
	if err != nil {
		return err
	}

	t.SbdsOffset = int8(bits)
	if t.SbdsOffset > 0x0f {
		t.SbdsOffset -= 0x20
	}

	// Bit 9
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}
	t.DefaultPixel = int8(bit)

	// Bit 7 - 8
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.CombinationOperator = bitmap.CombinationOperator(int(bits) & 0x3)

	// Bit 6
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}
	t.IsTransposed = int8(bit)

	// Bit 4 - 5
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.ReferenceCorner = int16(bits) & 0x3

	// Bit 2 - 3
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.LogSBStrips = int16(bits) & 0x3
	t.SbStrips = 1 << uint(t.LogSBStrips)

	// Bit 1
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		t.UseRefinement = true
	}

	// Bit 0
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}

	if bit == 1 {
		t.IsHuffmanEncoded = true
	}
	return nil
}

func (t *TextRegion) readHuffmanFlags() error {
	var (
		bit  int
		bits uint64
		err  error
	)
	// Bit 15 - dirty read
	_, err = t.r.ReadBit()
	if err != nil {
		return err
	}

	// Bit 14
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}
	t.SbHuffRSize = int8(bit)

	// Bit 12 - 13
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.SbHuffRDY = int8(bits) & 0xf

	// Bit 10 - 11
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.SbHuffRDX = int8(bits) & 0xf

	// Bit 8 - 9
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.SbHuffRDHeight = int8(bits) & 0xf

	// Bit 6 - 7
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.SbHuffRDWidth = int8(bits) & 0xf

	// Bit 4 - 5
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.SbHuffDT = int8(bits) & 0xf

	// Bit 2 - 3
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.SbHuffDS = int8(bits) & 0xf

	// Bit 0 - 1
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.SbHuffFS = int8(bits) & 0xf
	return nil
}

func (t *TextRegion) readUseRefinement() error {
	if !t.UseRefinement || t.SbrTemplate != 0 {
		return nil
	}

	var (
		temp byte
		err  error
	)
	t.SbrATX = make([]int8, 2)
	t.SbrATY = make([]int8, 2)

	// Byte 0
	temp, err = t.r.ReadByte()
	if err != nil {
		return err
	}
	t.SbrATX[0] = int8(temp)

	// Byte 1
	temp, err = t.r.ReadByte()
	if err != nil {
		return err
	}
	t.SbrATY[0] = int8(temp)

	// Byte 2
	temp, err = t.r.ReadByte()
	if err != nil {
		return err
	}
	t.SbrATX[1] = int8(temp)

	// Byte 3
	temp, err = t.r.ReadByte()
	if err != nil {
		return err
	}
	t.SbrATY[1] = int8(temp)

	return nil
}

func (t *TextRegion) readAmountOfSymbolInstances() error {
	bits, err := t.r.ReadBits(32)
	if err != nil {
		return err
	}
	t.NumberOfSymbolInstances = uint32(bits & math.MaxUint32)
	pixels := t.RegionInfo.BitmapWidth * t.RegionInfo.BitmapHeight

	if pixels < t.NumberOfSymbolInstances {
		common.Log.Debug("Limiting the number of decoded symbol instances to one per pixel ( %d instead of %d)", pixels, t.NumberOfSymbolInstances)
		t.NumberOfSymbolInstances = pixels
	}
	return nil
}

func (t *TextRegion) setCodingStatistics() error {
	if t.cxIADT == nil {
		t.cxIADT = arithmetic.NewStats(512, 1)
	}

	if t.cxIAFS == nil {
		t.cxIAFS = arithmetic.NewStats(512, 1)
	}

	if t.cxIADS == nil {
		t.cxIADS = arithmetic.NewStats(512, 1)
	}

	if t.cxIAIT == nil {
		t.cxIAIT = arithmetic.NewStats(512, 1)
	}

	if t.cxIARI == nil {
		t.cxIARI = arithmetic.NewStats(512, 1)
	}

	if t.cxIARDW == nil {
		t.cxIARDW = arithmetic.NewStats(512, 1)
	}

	if t.cxIARDH == nil {
		t.cxIARDH = arithmetic.NewStats(512, 1)
	}

	if t.cxIAID == nil {
		t.cxIAID = arithmetic.NewStats(1<<uint(t.symbolCodeLength), 1)
	}

	if t.cxIARDX == nil {
		t.cxIARDX = arithmetic.NewStats(512, 1)
	}

	if t.cxIARDY == nil {
		t.cxIARDY = arithmetic.NewStats(512, 1)
	}

	if t.arithmDecoder == nil {
		var err error
		t.arithmDecoder, err = arithmetic.New(t.r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TextRegion) setContexts(
	cx *arithmetic.DecoderStats, cxIADT *arithmetic.DecoderStats,
	cxIAFS *arithmetic.DecoderStats, cxIADS *arithmetic.DecoderStats,
	cxIAIT *arithmetic.DecoderStats, cxIAID *arithmetic.DecoderStats,
	cxIARDW *arithmetic.DecoderStats, cxIARDH *arithmetic.DecoderStats,
	cxIARDX *arithmetic.DecoderStats, cxIARDY *arithmetic.DecoderStats,
) {
	t.cxIADT = cxIADT
	t.cxIAFS = cxIAFS
	t.cxIADS = cxIADS
	t.cxIAIT = cxIAIT
	t.cxIARDW = cxIARDW
	t.cxIARDH = cxIARDH
	t.cxIAID = cxIAID
	t.cxIARDX = cxIARDX
	t.cxIARDY = cxIARDY
	t.cx = cx
}

// setParameters sets the text region segment parameters.
func (t *TextRegion) setParameters(
	arithmeticDecoder *arithmetic.Decoder,
	isHuffmanEncoded, sbRefine bool, sbw, sbh uint32,
	sbNumInstances uint32, sbStrips int8, sbNumSyms uint32,
	sbDefaultPixel int8, sbCombinationOperator bitmap.CombinationOperator,
	transposed int8, refCorner int16, sbdsOffset, sbHuffFS, sbHuffDS, sbHuffDT, sbHuffRDWidth,
	sbHuffRDHeight, sbHuffRDX, sbHuffRDY, sbHuffRSize, sbrTemplate int8,
	sbrAtx, sbrAty []int8, sbSyms []*bitmap.Bitmap, sbSymCodeLen int8,
) {
	t.arithmDecoder = arithmeticDecoder

	t.IsHuffmanEncoded = isHuffmanEncoded
	t.UseRefinement = sbRefine

	t.RegionInfo.BitmapWidth = sbw
	t.RegionInfo.BitmapHeight = sbh

	t.NumberOfSymbolInstances = sbNumInstances
	t.SbStrips = sbStrips
	t.NumberOfSymbols = sbNumSyms
	t.DefaultPixel = sbDefaultPixel
	t.CombinationOperator = sbCombinationOperator
	t.IsTransposed = transposed
	t.ReferenceCorner = refCorner
	t.SbdsOffset = sbdsOffset

	t.SbHuffFS = sbHuffFS
	t.SbHuffDS = sbHuffDS
	t.SbHuffDT = sbHuffDT
	t.SbHuffRDWidth = sbHuffRDWidth
	t.SbHuffRDHeight = sbHuffRDHeight
	t.SbHuffRDX = sbHuffRDX
	t.SbHuffRDY = sbHuffRDY

	t.SbrTemplate = sbrTemplate
	t.SbrATX = sbrAtx
	t.SbrATY = sbrAty

	t.Symbols = sbSyms
	t.symbolCodeLength = sbSymCodeLen
}

func (t *TextRegion) symbolIDCodeLengths() error {
	var (
		runCodeTable []*huffman.Code
		bits         uint64
		ht           huffman.Tabler
		err          error
	)

	for i := 0; i < 35; i++ {
		bits, err = t.r.ReadBits(4)
		if err != nil {
			return err
		}

		prefLen := int(bits & 0xf)
		if prefLen > 0 {
			runCodeTable = append(runCodeTable, huffman.NewCode(int32(prefLen), 0, int32(i), false))
		}
	}

	ht, err = huffman.NewFixedSizeTable(runCodeTable)
	if err != nil {
		return err
	}

	// 3) - 5)
	var (
		previousCodeLength int64
		counter            uint32
		sbSymCodes         []*huffman.Code
		code               int64
	)

	for counter < t.NumberOfSymbols {
		code, err = ht.Decode(t.r)
		if err != nil {
			return err
		}

		if code < 32 {
			if code > 0 {
				sbSymCodes = append(sbSymCodes, huffman.NewCode(int32(code), 0, int32(counter), false))
			}
			previousCodeLength = code
			counter++
		} else {
			var runLength, currCodeLength int64

			switch code {
			case 32:
				bits, err = t.r.ReadBits(2)
				if err != nil {
					return err
				}
				runLength = 3 + int64(bits)
				if counter > 0 {
					currCodeLength = previousCodeLength
				}
			case 33:
				bits, err = t.r.ReadBits(3)
				if err != nil {
					return err
				}
				runLength = 3 + int64(bits)
			case 34:
				bits, err = t.r.ReadBits(7)
				if err != nil {
					return err
				}
				runLength = 11 + int64(bits)
			}

			for j := 0; j < int(runLength); j++ {
				if currCodeLength > 0 {
					sbSymCodes = append(sbSymCodes, huffman.NewCode(int32(currCodeLength), 0, int32(counter), false))
				}
				counter++
			}
		}
	}
	// 6) Skip over remaining bits in the last Byte read
	t.r.Align()

	// 7)
	t.symbolCodeTable, err = huffman.NewFixedSizeTable(sbSymCodes)
	return err
}

func newTextRegion(r reader.StreamReader, h *Header) *TextRegion {
	t := &TextRegion{
		r:          r,
		Header:     h,
		RegionInfo: NewRegionSegment(r),
	}
	return t
}
