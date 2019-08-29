/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/huffman"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// TextRegion is the model for the jbig2 text region segment - see 7.4.1.
type TextRegion struct {
	r reader.StreamReader

	// Region segment information field 7.4.1.
	regionInfo *RegionSegment

	// Text Region segment flags 7.4.3.1.1.
	sbrTemplate         int8
	sbdsOffset          int8
	defaultPixel        int8
	combinationOperator bitmap.CombinationOperator
	isTransposed        int8
	referenceCorner     int16
	logSBStrips         int16
	useRefinement       bool
	isHuffmanEncoded    bool

	// Text region segment huffman flags 7.4.3.1.2.
	sbHuffRSize    int8
	sbHuffRDY      int8
	sbHuffRDX      int8
	sbHuffRDHeight int8
	sbHuffRDWidth  int8
	sbHuffDT       int8
	sbHuffDS       int8
	sbHuffFS       int8

	// Text region refinement AT flags 7.4.3.1.3.
	sbrATX []int8
	sbrATY []int8

	// Number of symbol instances 7.4.3.1.3.
	numberOfSymbolInstances uint32

	currentS        int64
	sbStrips        int8
	numberOfSymbols uint32

	regionBitmap *bitmap.Bitmap
	symbols      []*bitmap.Bitmap

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
}

// Compile time checks for the TextRegion interfaces implementation.
var (
	_ Regioner  = &TextRegion{}
	_ Segmenter = &TextRegion{}
)

// GetRegionBitmap implements Regioner interface.
func (t *TextRegion) GetRegionBitmap() (*bitmap.Bitmap, error) {
	if t.regionBitmap != nil {
		return t.regionBitmap, nil
	}

	if !t.isHuffmanEncoded {
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
	return t.regionBitmap, nil
}

// GetRegionInfo implements Regioner interface.
func (t *TextRegion) GetRegionInfo() *RegionSegment {
	return t.regionInfo
}

// Init implements Segmenter interface.
func (t *TextRegion) Init(header *Header, r reader.StreamReader) error {
	t.Header = header
	t.r = r
	t.regionInfo = NewRegionSegment(t.r)
	return t.parseHeader()
}

// String implements the Stringer interface.
func (t *TextRegion) String() string {
	sb := &strings.Builder{}
	sb.WriteString("\n[TEXT REGION]\n")
	sb.WriteString(t.regionInfo.String() + "\n")
	sb.WriteString(fmt.Sprintf("\t- sbrTemplate: %v\n", t.sbrTemplate))
	sb.WriteString(fmt.Sprintf("\t- sbdsOffset: %v\n", t.sbdsOffset))
	sb.WriteString(fmt.Sprintf("\t- defaultPixel: %v\n", t.defaultPixel))
	sb.WriteString(fmt.Sprintf("\t- combinationOperator: %v\n", t.combinationOperator))
	sb.WriteString(fmt.Sprintf("\t- isTransposed: %v\n", t.isTransposed))
	sb.WriteString(fmt.Sprintf("\t- referenceCorner: %v\n", t.referenceCorner))
	sb.WriteString(fmt.Sprintf("\t- useRefinement: %v\n", t.useRefinement))
	sb.WriteString(fmt.Sprintf("\t- isHuffmanEncoded: %v\n", t.isHuffmanEncoded))

	if t.isHuffmanEncoded {
		sb.WriteString(fmt.Sprintf("\t- sbHuffRSize: %v\n", t.sbHuffRSize))
		sb.WriteString(fmt.Sprintf("\t- sbHuffRDY: %v\n", t.sbHuffRDY))
		sb.WriteString(fmt.Sprintf("\t- sbHuffRDX: %v\n", t.sbHuffRDX))
		sb.WriteString(fmt.Sprintf("\t- sbHuffRDHeight: %v\n", t.sbHuffRDHeight))
		sb.WriteString(fmt.Sprintf("\t- sbHuffRDWidth: %v\n", t.sbHuffRDWidth))
		sb.WriteString(fmt.Sprintf("\t- sbHuffDT: %v\n", t.sbHuffDT))
		sb.WriteString(fmt.Sprintf("\t- sbHuffDS: %v\n", t.sbHuffDS))
		sb.WriteString(fmt.Sprintf("\t- sbHuffFS: %v\n", t.sbHuffFS))
	}

	sb.WriteString(fmt.Sprintf("\t- sbrATX: %v\n", t.sbrATX))
	sb.WriteString(fmt.Sprintf("\t- sbrATY: %v\n", t.sbrATY))
	sb.WriteString(fmt.Sprintf("\t- numberOfSymbolInstances: %v\n", t.numberOfSymbolInstances))
	sb.WriteString(fmt.Sprintf("\t- sbrATX: %v\n", t.sbrATX))
	return sb.String()
}

func (t *TextRegion) blit(ib *bitmap.Bitmap, tc int64) error {
	if t.isTransposed == 0 && (t.referenceCorner == 2 || t.referenceCorner == 3) {
		t.currentS += int64(ib.Width - 1)
	} else if t.isTransposed == 1 && (t.referenceCorner == 0 || t.referenceCorner == 2) {
		t.currentS += int64(ib.Height - 1)
	}

	// VII)
	s := t.currentS

	// VIII)
	if t.isTransposed == 1 {
		s, tc = tc, s
	}

	if t.referenceCorner != 1 {
		if t.referenceCorner == 0 {
			// BL
			tc -= int64(ib.Height - 1)
		} else if t.referenceCorner == 2 {
			// BR
			tc -= int64(ib.Height - 1)
			s -= int64(ib.Width - 1)
		} else if t.referenceCorner == 3 {
			// TR
			s -= int64(ib.Width - 1)
		}
	}

	err := bitmap.Blit(ib, t.regionBitmap, int(s), int(tc), t.combinationOperator)
	if err != nil {
		return err
	}

	// X)
	if t.isTransposed == 0 && (t.referenceCorner == 0 || t.referenceCorner == 1) {
		t.currentS += int64(ib.Width - 1)
	}

	if t.isTransposed == 1 && (t.referenceCorner == 1 || t.referenceCorner == 3) {
		t.currentS += int64(ib.Height - 1)
	}
	return nil
}

func (t *TextRegion) computeSymbolCodeLength() error {
	if t.isHuffmanEncoded {
		return t.symbolIDCodeLengths()
	}

	t.symbolCodeLength = int8(math.Ceil(math.Log(float64(t.numberOfSymbols)) / math.Log(2)))
	return nil
}

func (t *TextRegion) checkInput() error {
	if !t.useRefinement {
		if t.sbrTemplate != 0 {
			common.Log.Debug("sbrTemplate should be 0")
			t.sbrTemplate = 0
		}
	}

	if t.sbHuffFS == 2 || t.sbHuffRDWidth == 2 || t.sbHuffRDHeight == 2 || t.sbHuffRDX == 2 || t.sbHuffRDY == 2 {
		return errors.New("huffman flag value in text region segment is not permitted")
	}

	if !t.useRefinement {
		if t.sbHuffRSize != 0 {
			common.Log.Debug("sbHuffRSize should be 0")
			t.sbHuffRSize = 0
		}

		if t.sbHuffRDY != 0 {
			common.Log.Debug("sbHuffRDY should be 0")
			t.sbHuffRDY = 0
		}

		if t.sbHuffRDX != 0 {
			common.Log.Debug("sbHuffRDX should be 0")
			t.sbHuffRDX = 0
		}

		if t.sbHuffRDWidth != 0 {
			common.Log.Debug("sbHuffRDWidth should be 0")
			t.sbHuffRDWidth = 0
		}

		if t.sbHuffRDHeight != 0 {
			common.Log.Debug("sbHuffRDHeight should be 0")
			t.sbHuffRDHeight = 0
		}
	}
	return nil
}

func (t *TextRegion) createRegionBitmap() error {
	// 6.4.5
	t.regionBitmap = bitmap.New(int(t.regionInfo.BitmapWidth), int(t.regionInfo.BitmapHeight))
	if t.defaultPixel != 0 {
		t.regionBitmap.SetDefaultPixel()
	}
	return nil
}

func (t *TextRegion) decodeStripT() (stripT int64, err error) {
	// 2)
	if t.isHuffmanEncoded {
		// 6.4.6
		if t.sbHuffDT == 3 {
			if t.table == nil {
				var dtNr int
				if t.sbHuffFS == 3 {
					dtNr++
				}

				if t.sbHuffDS == 3 {
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
			table, err = huffman.GetStandardTable(11 + int(t.sbHuffDT))
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
	stripT *= int64(-t.sbStrips)
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
	for instanceCounter < t.numberOfSymbolInstances {
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

				if idS == math.MaxInt32 || instanceCounter >= t.numberOfSymbolInstances {
					break
				}

				t.currentS += (idS + int64(t.sbdsOffset))
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
	if t.isHuffmanEncoded {
		if t.sbHuffDT == 3 {
			dT, err = t.table.Decode(t.r)
			if err != nil {
				return 0, err
			}
		} else {
			var st huffman.Tabler
			st, err = huffman.GetStandardTable(11 + int(t.sbHuffDT))
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
	dT *= int64(t.sbStrips)
	return dT, nil
}

func (t *TextRegion) decodeDfs() (int64, error) {
	if t.isHuffmanEncoded {
		if t.sbHuffFS == 3 {
			if t.fsTable == nil {
				var err error
				t.fsTable, err = t.getUserTable(0)
				if err != nil {
					return 0, err
				}
			}
			return t.fsTable.Decode(t.r)
		}
		st, err := huffman.GetStandardTable(6 + int(t.sbHuffFS))
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
	if t.sbStrips != 1 {
		if t.isHuffmanEncoded {
			bits, err := t.r.ReadBits(byte(t.logSBStrips))
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
	if t.isHuffmanEncoded {
		if t.symbolCodeTable == nil {
			bits, err := t.r.ReadBits(byte(t.symbolCodeLength))
			return int64(bits), err
		}
		return t.symbolCodeTable.Decode(t.r)
	}
	return t.arithmDecoder.DecodeIAID(uint64(t.symbolCodeLength), t.cxIAID)
}

func (t *TextRegion) decodeRI() (int64, error) {
	if !t.useRefinement {
		return 0, nil
	}

	if t.isHuffmanEncoded {
		temp, err := t.r.ReadBit()
		return int64(temp), err
	}

	temp, err := t.arithmDecoder.DecodeInt(t.cxIARI)
	return int64(temp), err
}

func (t *TextRegion) decodeIb(r, id int64) (*bitmap.Bitmap, error) {
	var (
		err error
		ib  *bitmap.Bitmap
	)

	if r == 0 {
		if int(id) > len(t.symbols)-1 {
			return nil, errors.New("decoding IB bitmap. index out of range")
		}
		return t.symbols[int(id)], nil
	}
	// 1) - 4)
	var rdw, rdh, rdx, rdy int64

	rdw, err = t.decodeRdw()
	if err != nil {
		return nil, err
	}

	rdh, err = t.decodeRdh()
	if err != nil {
		return nil, err
	}

	rdx, err = t.decodeRdx()
	if err != nil {
		return nil, err
	}

	rdy, err = t.decodeRdy()
	if err != nil {
		return nil, err
	}

	// 5)
	if t.isHuffmanEncoded {
		if _, err = t.decodeSymInRefSize(); err != nil {
			return nil, err
		}
		t.r.Align()
	}

	// 6)
	ibo := t.symbols[id]
	wo := uint32(ibo.Width)
	ho := uint32(ibo.Height)
	genericRegionReferenceDX := int32(uint32(rdw)>>1) + int32(rdx)
	genericRegionReferenceDY := int32(uint32(rdh)>>1) + int32(rdy)

	if t.genericRefinementRegion == nil {
		t.genericRefinementRegion = newGenericRefinementRegion(t.r, nil)
	}

	t.genericRefinementRegion.setParameters(t.cx, t.arithmDecoder, t.sbrTemplate,
		wo+uint32(rdw), ho+uint32(rdh), ibo, genericRegionReferenceDX, genericRegionReferenceDY, false, t.sbrATX, t.sbrATY)

	ib, err = t.genericRefinementRegion.GetRegionBitmap()
	if err != nil {
		return nil, err
	}

	// 7)
	if t.isHuffmanEncoded {
		t.r.Align()
	}
	return ib, nil
}

func (t *TextRegion) decodeIds() (int64, error) {
	if t.isHuffmanEncoded {
		if t.sbHuffDS == 3 {
			if t.dsTable == nil {
				dsNr := 0
				if t.sbHuffFS == 3 {
					dsNr++
				}

				var err error
				t.dsTable, err = t.getUserTable(dsNr)
				if err != nil {
					return 0, err
				}
			}
			return t.dsTable.Decode(t.r)
		}
		st, err := huffman.GetStandardTable(8 + int(t.sbHuffDS))
		if err != nil {
			return 0, err
		}
		return st.Decode(t.r)
	}

	i, err := t.arithmDecoder.DecodeInt(t.cxIADS)
	if err != nil {
		return 0, err
	}
	return int64(i), nil
}

func (t *TextRegion) decodeRdw() (int64, error) {
	if t.isHuffmanEncoded {
		if t.sbHuffRDWidth == 3 {
			if t.rdwTable == nil {
				var (
					rdwNr int
					err   error
				)

				if t.sbHuffFS == 3 {
					rdwNr++
				}

				if t.sbHuffDS == 3 {
					rdwNr++
				}

				if t.sbHuffDT == 3 {
					rdwNr++
				}

				t.rdwTable, err = t.getUserTable(rdwNr)
				if err != nil {
					return 0, err
				}
			}
			return t.rdwTable.Decode(t.r)
		}
		ts, err := huffman.GetStandardTable(14 + int(t.sbHuffRDWidth))
		if err != nil {
			return 0, err
		}
		return ts.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIARDW)
	return int64(temp), err
}

func (t *TextRegion) decodeRdh() (int64, error) {
	if t.isHuffmanEncoded {
		if t.sbHuffRDHeight == 3 {
			if t.rdhTable == nil {
				var (
					rdhNr int
					err   error
				)

				if t.sbHuffFS == 3 {
					rdhNr++
				}

				if t.sbHuffDS == 3 {
					rdhNr++
				}

				if t.sbHuffDT == 3 {
					rdhNr++
				}

				if t.sbHuffRDWidth == 3 {
					rdhNr++
				}

				t.rdhTable, err = t.getUserTable(rdhNr)
				if err != nil {
					return 0, err
				}
			}
			return t.rdhTable.Decode(t.r)
		}
		ts, err := huffman.GetStandardTable(14 + int(t.sbHuffRDHeight))
		if err != nil {
			return 0, err
		}

		return ts.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIARDH)
	return int64(temp), err
}

func (t *TextRegion) decodeRdx() (int64, error) {
	if t.isHuffmanEncoded {
		if t.sbHuffRDX == 3 {
			if t.rdxTable == nil {
				var (
					rdxNr int
					err   error
				)

				if t.sbHuffFS == 3 {
					rdxNr++
				}

				if t.sbHuffDS == 3 {
					rdxNr++
				}

				if t.sbHuffDT == 3 {
					rdxNr++
				}

				if t.sbHuffRDWidth == 3 {
					rdxNr++
				}

				if t.sbHuffRDHeight == 3 {
					rdxNr++
				}

				t.rdxTable, err = t.getUserTable(rdxNr)
				if err != nil {
					return 0, err
				}
			}
			return t.rdxTable.Decode(t.r)
		}
		ts, err := huffman.GetStandardTable(14 + int(t.sbHuffRDX))
		if err != nil {
			return 0, err
		}
		return ts.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIARDX)
	return int64(temp), err
}

func (t *TextRegion) decodeRdy() (int64, error) {
	if t.isHuffmanEncoded {
		if t.sbHuffRDY == 3 {
			if t.rdyTable == nil {
				var (
					rdyNr int
					err   error
				)

				if t.sbHuffFS == 3 {
					rdyNr++
				}

				if t.sbHuffDS == 3 {
					rdyNr++
				}

				if t.sbHuffDT == 3 {
					rdyNr++
				}

				if t.sbHuffRDWidth == 3 {
					rdyNr++
				}

				if t.sbHuffRDHeight == 3 {
					rdyNr++
				}

				if t.sbHuffRDX == 3 {
					rdyNr++
				}

				t.rdyTable, err = t.getUserTable(rdyNr)
				if err != nil {
					return 0, err
				}
			}
			return t.rdyTable.Decode(t.r)
		}
		ts, err := huffman.GetStandardTable(14 + int(t.sbHuffRDY))
		if err != nil {
			return 0, err
		}
		return ts.Decode(t.r)
	}
	temp, err := t.arithmDecoder.DecodeInt(t.cxIARDY)
	return int64(temp), err
}

func (t *TextRegion) decodeSymInRefSize() (int64, error) {
	if t.sbHuffRSize == 0 {
		ts, err := huffman.GetStandardTable(1)
		if err != nil {
			return 0, err
		}
		return ts.Decode(t.r)
	}
	if t.rSizeTable == nil {
		var (
			rSizeNr int
			err     error
		)

		if t.sbHuffFS == 3 {
			rSizeNr++
		}

		if t.sbHuffDS == 3 {
			rSizeNr++
		}

		if t.sbHuffDT == 3 {
			rSizeNr++
		}

		if t.sbHuffRDWidth == 3 {
			rSizeNr++
		}

		if t.sbHuffRDHeight == 3 {
			rSizeNr++
		}

		if t.sbHuffRDX == 3 {
			rSizeNr++
		}

		if t.sbHuffRDY == 3 {
			rSizeNr++
		}

		t.rSizeTable, err = t.getUserTable(rSizeNr)
		if err != nil {
			return 0, err
		}
	}
	return t.rSizeTable.Decode(t.r)
}

func (t *TextRegion) getSymbols() error {
	if t.Header.RTSegments != nil {
		return t.initSymbols()
	}
	return nil
}

func (t *TextRegion) getUserTable(tablePosition int) (huffman.Tabler, error) {
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
					return nil, errors.New("jbig2 document - internal segment error")
				}
				return huffman.NewEncodedTable(ts)
			}
			tableCounter++
		}
	}
	return nil, nil
}

func (t *TextRegion) initSymbols() error {
	for _, segment := range t.Header.RTSegments {
		if segment == nil {
			return errors.New("jbig2 - internal error - nil segment provided for the text region symbols")
		}

		if segment.Type == 0 {
			s, err := segment.GetSegmentData()
			if err != nil {
				return err
			}

			sd, ok := s.(*SymbolDictionary)
			if !ok {
				return errors.New("text region - Refered To Segment is not a SymbolDictionary")
			}

			sd.cxIAID = t.cxIAID
			dict, err := sd.GetDictionary()
			if err != nil {
				return err
			}

			t.symbols = append(t.symbols, dict...)
		}
	}
	t.numberOfSymbols = uint32(len(t.symbols))
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

	if err = t.regionInfo.parseHeader(); err != nil {
		return err
	}

	if err = t.readRegionFlags(); err != nil {
		return err
	}

	if t.isHuffmanEncoded {
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
	t.sbrTemplate = int8(bit)

	// Bit 10 - 14
	bits, err = t.r.ReadBits(5)
	if err != nil {
		return err
	}

	t.sbdsOffset = int8(bits)
	if t.sbdsOffset > 0x0f {
		t.sbdsOffset -= 0x20
	}

	// Bit 9
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}
	t.defaultPixel = int8(bit)

	// Bit 7 - 8
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.combinationOperator = bitmap.CombinationOperator(int(bits) & 0x3)

	// Bit 6
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}
	t.isTransposed = int8(bit)

	// Bit 4 - 5
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.referenceCorner = int16(bits) & 0x3

	// Bit 2 - 3
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.logSBStrips = int16(bits) & 0x3
	t.sbStrips = 1 << uint(t.logSBStrips)

	// Bit 1
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		t.useRefinement = true
	}

	// Bit 0
	bit, err = t.r.ReadBit()
	if err != nil {
		return err
	}

	if bit == 1 {
		t.isHuffmanEncoded = true
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
	t.sbHuffRSize = int8(bit)

	// Bit 12 - 13
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.sbHuffRDY = int8(bits) & 0xf

	// Bit 10 - 11
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.sbHuffRDX = int8(bits) & 0xf

	// Bit 8 - 9
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.sbHuffRDHeight = int8(bits) & 0xf

	// Bit 6 - 7
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.sbHuffRDWidth = int8(bits) & 0xf

	// Bit 4 - 5
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.sbHuffDT = int8(bits) & 0xf

	// Bit 2 - 3
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.sbHuffDS = int8(bits) & 0xf

	// Bit 0 - 1
	bits, err = t.r.ReadBits(2)
	if err != nil {
		return err
	}
	t.sbHuffFS = int8(bits) & 0xf
	return nil
}

func (t *TextRegion) readUseRefinement() error {
	if !t.useRefinement || t.sbrTemplate != 0 {
		return nil
	}

	var (
		temp byte
		err  error
	)
	t.sbrATX = make([]int8, 2)
	t.sbrATY = make([]int8, 2)

	// Byte 0
	temp, err = t.r.ReadByte()
	if err != nil {
		return err
	}
	t.sbrATX[0] = int8(temp)

	// Byte 1
	temp, err = t.r.ReadByte()
	if err != nil {
		return err
	}
	t.sbrATY[0] = int8(temp)

	// Byte 2
	temp, err = t.r.ReadByte()
	if err != nil {
		return err
	}
	t.sbrATX[1] = int8(temp)

	// Byte 3
	temp, err = t.r.ReadByte()
	if err != nil {
		return err
	}
	t.sbrATY[1] = int8(temp)

	return nil
}

func (t *TextRegion) readAmountOfSymbolInstances() error {
	bits, err := t.r.ReadBits(32)
	if err != nil {
		return err
	}
	t.numberOfSymbolInstances = uint32(bits & math.MaxUint32)
	pixels := t.regionInfo.BitmapWidth * t.regionInfo.BitmapHeight

	if pixels < t.numberOfSymbolInstances {
		common.Log.Debug("Limiting the number of decoded symbol instances to one per pixel ( %d instead of %d)", pixels, t.numberOfSymbolInstances)
		t.numberOfSymbolInstances = pixels
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
	sbrATX, sbrATY []int8, sbSyms []*bitmap.Bitmap, sbSymCodeLen int8,
) {
	t.arithmDecoder = arithmeticDecoder

	t.isHuffmanEncoded = isHuffmanEncoded
	t.useRefinement = sbRefine

	t.regionInfo.BitmapWidth = sbw
	t.regionInfo.BitmapHeight = sbh

	t.numberOfSymbolInstances = sbNumInstances
	t.sbStrips = sbStrips
	t.numberOfSymbols = sbNumSyms
	t.defaultPixel = sbDefaultPixel
	t.combinationOperator = sbCombinationOperator
	t.isTransposed = transposed
	t.referenceCorner = refCorner
	t.sbdsOffset = sbdsOffset

	t.sbHuffFS = sbHuffFS
	t.sbHuffDS = sbHuffDS
	t.sbHuffDT = sbHuffDT
	t.sbHuffRDWidth = sbHuffRDWidth
	t.sbHuffRDHeight = sbHuffRDHeight
	t.sbHuffRDX = sbHuffRDX
	t.sbHuffRDY = sbHuffRDY

	t.sbrTemplate = sbrTemplate
	t.sbrATX = sbrATX
	t.sbrATY = sbrATY

	t.symbols = sbSyms
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

	for counter < t.numberOfSymbols {
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
		regionInfo: NewRegionSegment(r),
	}
	return t
}
