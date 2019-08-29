/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/huffman"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// SymbolDictionary is the model for the JBIG2 Symbol Dictionary Segment - see 7.4.2.
type SymbolDictionary struct {
	r reader.StreamReader

	// Symbol Dictionary flags, 7.4.2.1.1
	sdrTemplate                 int8
	sdTemplate                  int8
	isCodingContextRetained     bool
	isCodingContextUsed         bool
	sdHuffAggInstanceSelection  int8
	sdHuffBMSizeSelection       int8
	sdHuffDecodeWidthSelection  int8
	sdHuffDecodeHeightSelection int8
	useRefinementAggregation    bool
	isHuffmanEncoded            bool

	// Symbol Dictionary AT flags 7.4.2.1.2
	sdATX []int8
	sdATY []int8

	// Symbol Dictionary refinement AT flags 7.4.2.1.3
	sdrATX []int8
	sdrATY []int8

	// Number of exported symbols, 7.4.2.1.4
	numberOfExportedSymbols uint32

	// Number of new symbols 7.4.2.1.5
	numberOfNewSymbols uint32

	// Further parameters
	Header                  *Header
	numberOfImportedSymbols uint32
	importSymbols           []*bitmap.Bitmap
	numberOfDecodedSymbols  uint32
	newSymbols              []*bitmap.Bitmap

	// User-supplied tables
	dhTable      huffman.Tabler
	dwTable      huffman.Tabler
	bmSizeTable  huffman.Tabler
	aggInstTable huffman.Tabler

	// Return value of that segment
	exportSymbols []*bitmap.Bitmap
	sbSymbols     []*bitmap.Bitmap

	arithmeticDecoder *arithmetic.Decoder

	textRegion              *TextRegion
	genericRegion           *GenericRegion
	genericRefinementRegion *GenericRefinementRegion
	cx                      *arithmetic.DecoderStats

	cxIADH  *arithmetic.DecoderStats
	cxIADW  *arithmetic.DecoderStats
	cxIAAI  *arithmetic.DecoderStats
	cxIAEX  *arithmetic.DecoderStats
	cxIARDX *arithmetic.DecoderStats
	cxIARDY *arithmetic.DecoderStats
	cxIADT  *arithmetic.DecoderStats

	cxIAID       *arithmetic.DecoderStats
	sbSymCodeLen int8
}

// NumberOfExportedSymbols defines how many symbols are being exported by this SymbolDictionary.
func (s *SymbolDictionary) NumberOfExportedSymbols() uint32 {
	return s.numberOfExportedSymbols
}

// NumberOfNewSymbols returns the amount of new symbols defined by the Symbol Dictionary.
func (s *SymbolDictionary) NumberOfNewSymbols() uint32 {
	return s.numberOfNewSymbols
}

// GetDictionary gets the decoded dictionary symbols as a bitmap slice.
func (s *SymbolDictionary) GetDictionary() ([]*bitmap.Bitmap, error) {
	common.Log.Trace("[SYMBOL-DICTIONARY] GetDictionary begins...")
	defer func() {
		common.Log.Trace("[SYMBOL-DICTIONARY] GetDictionary finished")
	}()

	if s.exportSymbols == nil {
		var err error
		if s.useRefinementAggregation {
			s.sbSymCodeLen = s.getSbSymCodeLen()
		}

		if !s.isHuffmanEncoded {
			if err = s.setCodingStatistics(); err != nil {
				return nil, err
			}
		}

		// 6.5.5. 1)
		s.newSymbols = make([]*bitmap.Bitmap, s.numberOfNewSymbols)

		// 6.5.5. 2)
		var newSymbolsWidth []int
		if s.isHuffmanEncoded && !s.useRefinementAggregation {
			newSymbolsWidth = make([]int, s.numberOfNewSymbols)
		}

		if err = s.setSymbolsArray(); err != nil {
			return nil, err
		}

		// 6.5.5 3)
		var heightClassHeight, temp int64
		s.numberOfDecodedSymbols = 0

		// 6.5.5 4 a)
		for s.numberOfDecodedSymbols < s.numberOfNewSymbols {
			// 6.5.5 4 b)
			temp, err = s.decodeHeightClassDeltaHeight()
			if err != nil {
				return nil, err
			}
			heightClassHeight += temp

			var symbolWidth, totalWidth uint32
			heightClassFirstSymbolIndex := int64(s.numberOfDecodedSymbols)

			// 6.5.5 4 c)
			// Repeat until OOB - OOB sends a break
			for {
				// 4 c) i)
				var differenceWidth int64
				differenceWidth, err = s.decodeDifferenceWidth()
				if err != nil {
					return nil, err
				}

				if differenceWidth == int64(math.MaxInt64) || s.numberOfDecodedSymbols >= s.numberOfNewSymbols {
					break
				}

				symbolWidth += uint32(differenceWidth)
				totalWidth += symbolWidth

				//* 4 c) ii)
				if !s.isHuffmanEncoded || s.useRefinementAggregation {
					if !s.useRefinementAggregation {
						// 6.5.8.1 - Directly coded
						err = s.decodeDirectlyThroughGenericRegion(symbolWidth, uint32(heightClassHeight))
						if err != nil {
							return nil, err
						}
					} else {
						// 6.5.8.2 - Refinement / Aggregate -coded
						err = s.decodeAggregate(symbolWidth, uint32(heightClassHeight))
						if err != nil {
							return nil, err
						}
					}
				} else if s.isHuffmanEncoded && !s.useRefinementAggregation {
					// 4 c) iii)
					newSymbolsWidth[s.numberOfDecodedSymbols] = int(symbolWidth)
				}
				s.numberOfDecodedSymbols++
			}

			// 6.5.5 4 d)
			if s.isHuffmanEncoded && !s.useRefinementAggregation {
				var bmSize int64
				if s.sdHuffBMSizeSelection == 0 {
					var st huffman.Tabler
					st, err = huffman.GetStandardTable(1)
					if err != nil {
						return nil, err
					}
					bmSize, err = st.Decode(s.r)
					if err != nil {
						return nil, err
					}
				} else {
					bmSize, err = s.huffDecodeBmSize()
					if err != nil {
						return nil, err
					}
				}

				s.r.Align()

				var heightClassCollectiveBitmap *bitmap.Bitmap
				heightClassCollectiveBitmap, err = s.decodeHeightClassCollectiveBitmap(
					bmSize, uint32(heightClassHeight), totalWidth)
				if err != nil {
					return nil, err
				}

				err = s.decodeHeightClassBitmap(
					heightClassCollectiveBitmap, heightClassFirstSymbolIndex,
					int(heightClassHeight), newSymbolsWidth,
				)
				if err != nil {
					return nil, err
				}
			}
		}
		// 5)
		// 6.5.10 1) - 5
		exFlags, err := s.getToExportFlags()
		if err != nil {
			return nil, err
		}
		s.setExportedSymbols(exFlags)
	}
	return s.exportSymbols, nil
}

// Init implements Segmenter interface.
func (s *SymbolDictionary) Init(h *Header, r reader.StreamReader) error {
	s.Header = h
	s.r = r
	return s.parseHeader()
}

// IsHuffmanEncoded defines if the segment is encoded using huffman tables.
func (s *SymbolDictionary) IsHuffmanEncoded() bool {
	return s.isHuffmanEncoded
}

// String implements the Stringer interface.
func (s *SymbolDictionary) String() string {
	sb := &strings.Builder{}
	sb.WriteString("\n[SYMBOL-DICTIONARY]\n")
	sb.WriteString(fmt.Sprintf("\t- sdrTemplate %v\n", s.sdrTemplate))
	sb.WriteString(fmt.Sprintf("\t- sdTemplate %v\n", s.sdTemplate))
	sb.WriteString(fmt.Sprintf("\t- isCodingContextRetained %v\n", s.isCodingContextRetained))
	sb.WriteString(fmt.Sprintf("\t- isCodingContextUsed %v\n", s.isCodingContextUsed))
	sb.WriteString(fmt.Sprintf("\t- sdHuffAggInstanceSelection %v\n", s.sdHuffAggInstanceSelection))
	sb.WriteString(fmt.Sprintf("\t- sdHuffBMSizeSelection %v\n", s.sdHuffBMSizeSelection))
	sb.WriteString(fmt.Sprintf("\t- sdHuffDecodeWidthSelection %v\n", s.sdHuffDecodeWidthSelection))
	sb.WriteString(fmt.Sprintf("\t- sdHuffDecodeHeightSelection %v\n", s.sdHuffDecodeHeightSelection))
	sb.WriteString(fmt.Sprintf("\t- useRefinementAggregation %v\n", s.useRefinementAggregation))
	sb.WriteString(fmt.Sprintf("\t- isHuffmanEncoded %v\n", s.isHuffmanEncoded))
	sb.WriteString(fmt.Sprintf("\t- sdATX %v\n", s.sdATX))
	sb.WriteString(fmt.Sprintf("\t- sdATY %v\n", s.sdATY))
	sb.WriteString(fmt.Sprintf("\t- sdrATX %v\n", s.sdrATX))
	sb.WriteString(fmt.Sprintf("\t- sdrATY %v\n", s.sdrATY))
	sb.WriteString(fmt.Sprintf("\t- numberOfExportedSymbols %v\n", s.numberOfExportedSymbols))
	sb.WriteString(fmt.Sprintf("\t- numberOfNewSymbols %v\n", s.numberOfNewSymbols))
	sb.WriteString(fmt.Sprintf("\t- numberOfImportedSymbols %v\n", s.numberOfImportedSymbols))
	sb.WriteString(fmt.Sprintf("\t- numberOfDecodedSymbols %v\n", s.numberOfDecodedSymbols))
	return sb.String()
}

// UseRefinementAggregation defines if the SymbolDictionary uses refinement aggregation.
func (s *SymbolDictionary) UseRefinementAggregation() bool {
	return s.useRefinementAggregation
}

func (s *SymbolDictionary) checkInput() error {
	if s.sdHuffDecodeHeightSelection == 2 {
		common.Log.Debug("Symbol Dictionary Decode Heigh Selection: %d value not permitted", s.sdHuffDecodeHeightSelection)
	}

	if s.sdHuffDecodeWidthSelection == 2 {
		common.Log.Debug("Symbol Dictionary Decode Width Selection: %d value not permitted", s.sdHuffDecodeWidthSelection)
	}

	if s.isHuffmanEncoded {
		if s.sdTemplate != 0 {
			common.Log.Debug("SDTemplate = %d (should be 0)", s.sdTemplate)
		}

		if !s.useRefinementAggregation {
			if !s.useRefinementAggregation {
				if s.isCodingContextRetained {
					common.Log.Debug("IsCodingContextRetained = true (should be false)")
					s.isCodingContextRetained = false
				}

				if s.isCodingContextUsed {
					common.Log.Debug("isCodingContextUsed = true (should be false)")
					s.isCodingContextUsed = false
				}
			}
		}
	} else {
		if s.sdHuffBMSizeSelection != 0 {
			common.Log.Debug("sdHuffBMSizeSelection should be 0")
			s.sdHuffBMSizeSelection = 0
		}

		if s.sdHuffDecodeWidthSelection != 0 {
			common.Log.Debug("sdHuffDecodeWidthSelection should be 0")
			s.sdHuffDecodeWidthSelection = 0
		}

		if s.sdHuffDecodeHeightSelection != 0 {
			common.Log.Debug("sdHuffDecodeHeightSelection should be 0")
			s.sdHuffDecodeHeightSelection = 0
		}
	}

	if !s.useRefinementAggregation {
		if s.sdrTemplate != 0 {
			common.Log.Debug("SDRTemplate = %d (should be 0)", s.sdrTemplate)
			s.sdrTemplate = 0
		}
	}

	if !s.isHuffmanEncoded || !s.useRefinementAggregation {
		if s.sdHuffAggInstanceSelection != 0 {
			common.Log.Debug("sdHuffAggInstanceSelection = %d (should be 0)", s.sdHuffAggInstanceSelection)
		}
	}
	return nil
}

func (s *SymbolDictionary) addSymbol(region Regioner) error {
	symbol, err := region.GetRegionBitmap()
	if err != nil {
		return err
	}
	s.newSymbols[s.numberOfDecodedSymbols] = symbol
	s.sbSymbols = append(s.sbSymbols, symbol)
	return nil
}

func (s *SymbolDictionary) decodeHeightClassBitmap(
	heightClassCollectiveBitmap *bitmap.Bitmap,
	heightClassFirstSymbol int64, heightClassHeight int,
	newSymbolsWidths []int,
) error {
	for i := heightClassFirstSymbol; i < int64(s.numberOfDecodedSymbols); i++ {
		var startColumn int

		for j := heightClassFirstSymbol; j <= i-1; j++ {
			startColumn += newSymbolsWidths[j]
		}

		roi := image.Rect(startColumn, 0, startColumn+newSymbolsWidths[i], heightClassHeight)
		symbolBitmap, err := bitmap.Extract(roi, heightClassCollectiveBitmap)
		if err != nil {
			return err
		}
		s.newSymbols[i] = symbolBitmap
		s.sbSymbols = append(s.sbSymbols, symbolBitmap)
	}
	return nil
}

func (s *SymbolDictionary) decodeAggregate(symbolWidth, heightClassHeight uint32) error {
	var (
		numberOfRefinementAggregationInstances int64
		err                                    error
	)

	//  6.5.8.2.1.
	if s.isHuffmanEncoded {
		numberOfRefinementAggregationInstances, err = s.huffDecodeRefAggNInst()
		if err != nil {
			return err
		}
	} else {
		t, err := s.arithmeticDecoder.DecodeInt(s.cxIAAI)
		if err != nil {
			return err
		}
		numberOfRefinementAggregationInstances = int64(t)
	}

	if numberOfRefinementAggregationInstances > 1 {
		// 6.5.8.2
		return s.decodeThroughTextRegion(symbolWidth, heightClassHeight, uint32(numberOfRefinementAggregationInstances))
	} else if numberOfRefinementAggregationInstances == 1 {
		return s.decodeRefinedSymbol(symbolWidth, heightClassHeight)
	}
	return nil
}

func (s *SymbolDictionary) decodeThroughTextRegion(symbolWidth, heightClassHeight, numberOfRefinementAggregationInstances uint32) error {
	if s.textRegion == nil {
		s.textRegion = newTextRegion(s.r, nil)

		s.textRegion.setContexts(
			s.cx,
			arithmetic.NewStats(512, 1), // IADT
			arithmetic.NewStats(512, 1), // IAFS
			arithmetic.NewStats(512, 1), // IADS
			arithmetic.NewStats(512, 1), // IAIT
			s.cxIAID,
			arithmetic.NewStats(512, 1), // IARDW
			arithmetic.NewStats(512, 1), // IARDH
			arithmetic.NewStats(512, 1), // IARDX
			arithmetic.NewStats(512, 1), // IARDY
		)
	}
	// 6.5.8.2.4 Concatenating the array used as parameter later
	if err := s.setSymbolsArray(); err != nil {
		return err
	}

	s.textRegion.setParameters(s.arithmeticDecoder, s.isHuffmanEncoded, true, symbolWidth,
		heightClassHeight, numberOfRefinementAggregationInstances, 1, (s.numberOfImportedSymbols + s.numberOfDecodedSymbols),
		0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, s.sdrTemplate, s.sdrATX, s.sdrATY, s.sbSymbols, s.sbSymCodeLen)
	return s.addSymbol(s.textRegion)
}

func (s *SymbolDictionary) decodeRefinedSymbol(symbolWidth, heightClassHeight uint32) error {
	var (
		id       int
		rdx, rdy int32
	)

	if s.isHuffmanEncoded {
		// 2) - 4)
		v, err := s.r.ReadBits(byte(s.sbSymCodeLen))
		if err != nil {
			return err
		}
		id = int(v)

		st, err := huffman.GetStandardTable(15)
		if err != nil {
			return err
		}

		iv, err := st.Decode(s.r)
		if err != nil {
			return err
		}
		rdx = int32(iv)

		iv, err = st.Decode(s.r)
		if err != nil {
			return err
		}
		rdy = int32(iv)

		st, err = huffman.GetStandardTable(1)
		if err != nil {
			return err
		}
		if _, err = st.Decode(s.r); err != nil {
			return err
		}

		// skip over remaining bits
		s.r.Align()
	} else {
		tid, err := s.arithmeticDecoder.DecodeIAID(uint64(s.sbSymCodeLen), s.cxIAID)
		if err != nil {
			return err
		}
		id = int(tid)

		rdx, err = s.arithmeticDecoder.DecodeInt(s.cxIARDX)
		if err != nil {
			return err
		}

		rdy, err = s.arithmeticDecoder.DecodeInt(s.cxIARDY)
		if err != nil {
			return err
		}
	}

	if err := s.setSymbolsArray(); err != nil {
		return err
	}

	ibo := s.sbSymbols[id]

	if err := s.decodeNewSymbols(symbolWidth, heightClassHeight, ibo, rdx, rdy); err != nil {
		return err
	}

	if s.isHuffmanEncoded {
		// Make sure that the processed bytes are equal to the value read in step 5.
		s.r.Align()
	}
	return nil
}

func (s *SymbolDictionary) decodeNewSymbols(symWidth, hcHeight uint32, ibo *bitmap.Bitmap, rdx, rdy int32) error {
	if s.genericRefinementRegion == nil {
		s.genericRefinementRegion = newGenericRefinementRegion(s.r, nil)

		if s.arithmeticDecoder == nil {
			var err error
			s.arithmeticDecoder, err = arithmetic.New(s.r)
			if err != nil {
				return err
			}
		}

		if s.cx == nil {
			s.cx = arithmetic.NewStats(65536, 1)
		}
	}
	s.genericRefinementRegion.setParameters(s.cx, s.arithmeticDecoder, s.sdrTemplate, symWidth, hcHeight, ibo, rdx, rdy, false, s.sdrATX, s.sdrATY)
	return s.addSymbol(s.genericRefinementRegion)
}

func (s *SymbolDictionary) decodeDirectlyThroughGenericRegion(symWidth, hcHeight uint32) error {
	if s.genericRegion == nil {
		s.genericRegion = NewGenericRegion(s.r)
	}
	s.genericRegion.setParametersWithAt(false, byte(s.sdTemplate), false, false, s.sdATX, s.sdATY, symWidth, hcHeight, s.cx, s.arithmeticDecoder)
	return s.addSymbol(s.genericRegion)
}

func (s *SymbolDictionary) decodeDifferenceWidth() (int64, error) {
	if s.isHuffmanEncoded {
		switch s.sdHuffDecodeWidthSelection {
		case 0:
			t, err := huffman.GetStandardTable(2)
			if err != nil {
				return 0, err
			}
			return t.Decode(s.r)
		case 1:
			t, err := huffman.GetStandardTable(3)
			if err != nil {
				return 0, err
			}
			return t.Decode(s.r)
		case 3:
			if s.dwTable == nil {
				var dwNr int
				if s.sdHuffDecodeHeightSelection == 3 {
					dwNr++
				}
				t, err := huffman.GetStandardTable(2)
				if err != nil {
					return 0, err
				}
				s.dwTable = t
			}
			return s.dwTable.Decode(s.r)
		}
	} else {
		i, err := s.arithmeticDecoder.DecodeInt(s.cxIADW)
		if err != nil {
			return 0, err
		}
		if i == math.MaxInt32 {
			return int64(math.MaxInt64), nil
		}
		return int64(i), nil
	}
	return 0, nil
}

func (s *SymbolDictionary) decodeHeightClassDeltaHeight() (int64, error) {
	if s.isHuffmanEncoded {
		return s.decodeHeightClassDeltaHeightWithHuffman()
	}

	i, err := s.arithmeticDecoder.DecodeInt(s.cxIADH)
	if err != nil {
		return 0, err
	}

	return int64(i), nil
}

// decodeHeightClassDeltaHeightWithHuffman - 6.5.6 decodes the symbol dictionary
// when the height class is encoded using huffman tables.
func (s *SymbolDictionary) decodeHeightClassDeltaHeightWithHuffman() (int64, error) {
	switch s.sdHuffDecodeHeightSelection {
	case 0:
		t, err := huffman.GetStandardTable(4)
		if err != nil {
			return 0, err
		}
		return t.Decode(s.r)
	case 1:
		t, err := huffman.GetStandardTable(5)
		if err != nil {
			return 0, err
		}
		return t.Decode(s.r)
	case 3:
		if s.dhTable == nil {
			t, err := huffman.GetStandardTable(0)
			if err != nil {
				return 0, err
			}
			s.dhTable = t
		}
		return s.dhTable.Decode(s.r)
	}
	return 0, nil
}

func (s *SymbolDictionary) decodeHeightClassCollectiveBitmap(
	bmSize int64, heightClassHeight, totalWidth uint32,
) (*bitmap.Bitmap, error) {
	if bmSize == 0 {
		heightClassColleciveBitmap := bitmap.New(int(totalWidth), int(heightClassHeight))
		var (
			b   byte
			err error
		)
		for i := 0; i < len(heightClassColleciveBitmap.Data); i++ {
			b, err = s.r.ReadByte()
			if err != nil {
				return nil, err
			}

			if err = heightClassColleciveBitmap.SetByte(i, b); err != nil {
				return nil, err
			}
		}
		return heightClassColleciveBitmap, nil
	}
	if s.genericRegion == nil {
		s.genericRegion = NewGenericRegion(s.r)
	}

	s.genericRegion.setParameters(true, s.r.StreamPosition(), bmSize, heightClassHeight, totalWidth)

	bm, err := s.genericRegion.GetRegionBitmap()
	if err != nil {
		return nil, err
	}
	return bm, nil
}

// getSbSymCodeLen 6.5.8.2.3 sets the SBSYMCODES variable.
func (s *SymbolDictionary) getSbSymCodeLen() int8 {
	first := int8(math.Ceil(
		math.Log(float64(s.numberOfImportedSymbols+s.numberOfNewSymbols)) / math.Log(2)))

	if s.isHuffmanEncoded && first < 1 {
		return 1
	}

	return first
}

func (s *SymbolDictionary) getToExportFlags() ([]int, error) {
	var (
		currentExportFlag int
		exRunLength       int32
		err               error
		totalNewSymbols   = int32(s.numberOfImportedSymbols + s.numberOfNewSymbols)
		exportFlags       = make([]int, totalNewSymbols)
	)

	for exportIndex := int32(0); exportIndex < totalNewSymbols; exportIndex += exRunLength {
		if s.isHuffmanEncoded {
			t, err := huffman.GetStandardTable(1)
			if err != nil {
				return nil, err
			}

			i, err := t.Decode(s.r)
			if err != nil {
				return nil, err
			}
			exRunLength = int32(i)
		} else {
			exRunLength, err = s.arithmeticDecoder.DecodeInt(s.cxIAEX)
			if err != nil {
				return nil, err
			}
		}

		if exRunLength != 0 {
			for index := exportIndex; index < exportIndex+exRunLength; index++ {
				exportFlags[index] = currentExportFlag
			}
		}

		if currentExportFlag == 0 {
			currentExportFlag = 1
		} else {
			currentExportFlag = 0
		}
	}
	return exportFlags, nil
}

func (s *SymbolDictionary) getUserTable(tablePosition int) (huffman.Tabler, error) {
	var tableCounter int

	for _, h := range s.Header.RTSegments {
		if h.Type == 53 {
			if tableCounter == tablePosition {
				d, err := h.GetSegmentData()
				if err != nil {
					return nil, err
				}
				t := d.(huffman.BasicTabler)
				return huffman.NewEncodedTable(t)
			}
			tableCounter++
		}
	}
	return nil, nil
}

func (s *SymbolDictionary) huffDecodeBmSize() (int64, error) {
	if s.bmSizeTable == nil {
		var (
			bmNr int
			err  error
		)

		if s.sdHuffDecodeHeightSelection == 3 {
			bmNr++
		}

		if s.sdHuffDecodeWidthSelection == 3 {
			bmNr++
		}

		s.bmSizeTable, err = s.getUserTable(bmNr)
		if err != nil {
			return 0, err
		}
	}
	return s.bmSizeTable.Decode(s.r)
}

func (s *SymbolDictionary) huffDecodeRefAggNInst() (int64, error) {
	if s.sdHuffAggInstanceSelection == 0 {
		t, err := huffman.GetStandardTable(1)
		if err != nil {
			return 0, err
		}
		return t.Decode(s.r)
	} else if s.sdHuffAggInstanceSelection == 1 {
		if s.aggInstTable == nil {
			var (
				aggregationInstanceNumber int
				err                       error
			)
			if s.sdHuffDecodeHeightSelection == 3 {
				aggregationInstanceNumber++
			}

			if s.sdHuffDecodeWidthSelection == 3 {
				aggregationInstanceNumber++
			}

			if s.sdHuffBMSizeSelection == 3 {
				aggregationInstanceNumber++
			}

			s.aggInstTable, err = s.getUserTable(aggregationInstanceNumber)
			if err != nil {
				return 0, err
			}
		}
		return s.aggInstTable.Decode(s.r)
	}
	return 0, nil
}

func (s *SymbolDictionary) parseHeader() (err error) {
	common.Log.Trace("[SYMBOL DICTIONARY][PARSE-HEADER] begins...")
	defer func() {
		if err != nil {
			common.Log.Trace("[SYMBOL DICTIONARY][PARSE-HEADER] failed. %v", err)
		} else {
			common.Log.Trace("[SYMBOL DICTIONARY][PARSE-HEADER] finished.")
		}
	}()

	if err = s.readRegionFlags(); err != nil {
		return err
	}
	if err = s.setAtPixels(); err != nil {
		return err
	}
	if err = s.setRefinementAtPixels(); err != nil {
		return err
	}
	if err = s.readNumberOfExportedSymbols(); err != nil {
		return err
	}
	if err = s.readNumberOfNewSymbols(); err != nil {
		return err
	}
	if err = s.setInSyms(); err != nil {
		return err
	}

	if s.isCodingContextUsed {
		rtSegments := s.Header.RTSegments

		for i := len(rtSegments) - 1; i >= 0; i-- {
			if rtSegments[i].Type == 0 {
				symbolDictionary, ok := rtSegments[i].SegmentData.(*SymbolDictionary)
				if !ok {
					err = fmt.Errorf("related Segment: %v is not a Symbol Dictionary Segment", rtSegments[i])
					return err
				}

				if symbolDictionary.isCodingContextUsed {
					s.setRetainedCodingContexts(symbolDictionary)
				}
				break
			}
		}
	}

	if err = s.checkInput(); err != nil {
		return err
	}
	return nil
}

func (s *SymbolDictionary) readRegionFlags() error {
	var (
		bits uint64
		bit  int
	)
	// Bit 13 - 15
	_, err := s.r.ReadBits(3) // Dirty read
	if err != nil {
		return err
	}

	// Bit 12 - SDRTemplate
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	s.sdrTemplate = int8(bit)

	//* Bit 10 - 11 - SDTemplate
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.sdTemplate = int8(bits & 0xf)

	// Bit 9 - isCodingContextRetained
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.isCodingContextRetained = true
	}

	// Bit 8 - isCodingContextUsed
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.isCodingContextUsed = true
	}

	// Bit 7 - sdHuffAggInstanceSelection
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	s.sdHuffBMSizeSelection = int8(bit)

	// Bit 6 - sdHuffBMSizeSelection
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	s.sdHuffBMSizeSelection = int8(bit)

	// Bit 4 - 5 - sdHuffDecodeWidthSelection
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.sdHuffDecodeWidthSelection = int8(bits & 0xf)

	// Bit 2 - 3 - sdHuffDecodeWidthSelection
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.sdHuffDecodeHeightSelection = int8(bits & 0xf)

	// Bit 1
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.useRefinementAggregation = true
	}

	// Bit 0
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.isHuffmanEncoded = true
	}
	return nil
}

func (s *SymbolDictionary) readAtPixels(pixelsNumber int) error {
	s.sdATX = make([]int8, pixelsNumber)
	s.sdATY = make([]int8, pixelsNumber)
	var (
		b   byte
		err error
	)

	for i := 0; i < pixelsNumber; i++ {
		b, err = s.r.ReadByte()
		if err != nil {
			return err
		}
		s.sdATX[i] = int8(b)
		b, err = s.r.ReadByte()
		if err != nil {
			return err
		}
		s.sdATY[i] = int8(b)
	}
	return nil
}

func (s *SymbolDictionary) readRefinementAtPixels(numberOfAtPixels int) error {
	s.sdrATX = make([]int8, numberOfAtPixels)
	s.sdrATY = make([]int8, numberOfAtPixels)
	var (
		b   byte
		err error
	)

	for i := 0; i < numberOfAtPixels; i++ {
		b, err = s.r.ReadByte()
		if err != nil {
			return err
		}
		s.sdrATX[i] = int8(b)
		b, err = s.r.ReadByte()
		if err != nil {
			return err
		}
		s.sdrATY[i] = int8(b)
	}
	return nil
}

func (s *SymbolDictionary) readNumberOfExportedSymbols() error {
	bits, err := s.r.ReadBits(32)
	if err != nil {
		return err
	}
	s.numberOfExportedSymbols = uint32(bits & math.MaxUint32)
	return nil
}

func (s *SymbolDictionary) readNumberOfNewSymbols() error {
	bits, err := s.r.ReadBits(32)
	if err != nil {
		return err
	}
	s.numberOfNewSymbols = uint32(bits & math.MaxUint32)
	return nil
}

// retrieveImportSymbols Concatenates symbols from all referred-to segments.
func (s *SymbolDictionary) retrieveImportSymbols() error {
	for _, h := range s.Header.RTSegments {
		if h.Type == 0 {
			sd, err := h.GetSegmentData()
			if err != nil {
				return err
			}

			dict, ok := sd.(*SymbolDictionary)
			if !ok {
				return fmt.Errorf("provided Segment Data is not a SymbolDictionary Segment: %T", sd)
			}

			relatedDict, err := dict.GetDictionary()
			if err != nil {
				return fmt.Errorf("related segment with index: %d getDictionary failed. %s", h.SegmentNumber, err.Error())
			}
			s.importSymbols = append(s.importSymbols, relatedDict...)
			s.numberOfImportedSymbols += dict.numberOfExportedSymbols
		}
	}
	return nil
}

func (s *SymbolDictionary) setAtPixels() error {
	if s.isHuffmanEncoded {
		return nil
	}
	index := 1
	if s.sdTemplate == 0 {
		index = 4
	}

	if err := s.readAtPixels(index); err != nil {
		return err
	}
	return nil
}

func (s *SymbolDictionary) setCodingStatistics() error {
	if s.cxIADT == nil {
		s.cxIADT = arithmetic.NewStats(512, 1)
	}

	if s.cxIADH == nil {
		s.cxIADH = arithmetic.NewStats(512, 1)
	}

	if s.cxIADW == nil {
		s.cxIADW = arithmetic.NewStats(512, 1)
	}

	if s.cxIAAI == nil {
		s.cxIAAI = arithmetic.NewStats(512, 1)
	}
	if s.cxIAEX == nil {
		s.cxIAEX = arithmetic.NewStats(512, 1)
	}

	if s.useRefinementAggregation && s.cxIAID == nil {
		s.cxIAID = arithmetic.NewStats(1<<uint(s.sbSymCodeLen), 1)
		s.cxIARDX = arithmetic.NewStats(512, 1)
		s.cxIARDY = arithmetic.NewStats(512, 1)
	}

	if s.cx == nil {
		s.cx = arithmetic.NewStats(65536, 1)
	}

	if s.arithmeticDecoder == nil {
		var err error
		s.arithmeticDecoder, err = arithmetic.New(s.r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SymbolDictionary) setExportedSymbols(toExportFlags []int) {
	for i := uint32(0); i < s.numberOfImportedSymbols+s.numberOfNewSymbols; i++ {
		if toExportFlags[i] == 1 {
			if i < s.numberOfImportedSymbols {
				s.exportSymbols = append(s.exportSymbols, s.importSymbols[i])
			} else {
				s.exportSymbols = append(s.exportSymbols, s.newSymbols[i-s.numberOfImportedSymbols])
			}
		}
	}
}

func (s *SymbolDictionary) setInSyms() error {
	if s.Header.RTSegments != nil {
		return s.retrieveImportSymbols()
	}
	s.importSymbols = make([]*bitmap.Bitmap, 0)
	return nil
}

func (s *SymbolDictionary) setRefinementAtPixels() error {
	if !s.useRefinementAggregation || s.sdrTemplate != 0 {
		return nil
	}

	if err := s.readRefinementAtPixels(2); err != nil {
		return err
	}
	return nil
}

func (s *SymbolDictionary) setRetainedCodingContexts(sd *SymbolDictionary) {
	s.arithmeticDecoder = sd.arithmeticDecoder
	s.isHuffmanEncoded = sd.isHuffmanEncoded
	s.useRefinementAggregation = sd.useRefinementAggregation
	s.sdTemplate = sd.sdTemplate
	s.sdrTemplate = sd.sdrTemplate
	s.sdATX = sd.sdATX
	s.sdATY = sd.sdATY
	s.sdrATX = sd.sdrATX
	s.sdrATY = sd.sdrATY
	s.cx = sd.cx
}

// setSymbolsArray 6.5.8.2.4 sets the SBSYMS variable.
func (s *SymbolDictionary) setSymbolsArray() error {
	if s.importSymbols == nil {
		if err := s.retrieveImportSymbols(); err != nil {
			return err
		}
	}

	if s.sbSymbols == nil {
		s.sbSymbols = append(s.sbSymbols, s.importSymbols...)
	}
	return nil
}
