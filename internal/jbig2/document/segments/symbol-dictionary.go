/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/decoder/huffman"
	encoder "github.com/unidoc/unipdf/v3/internal/jbig2/encoder/arithmetic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// SymbolDictionary is the model for the JBIG2 Symbol Dictionary Segment - see 7.4.2.
type SymbolDictionary struct {
	r reader.StreamReader

	// Symbol Dictionary flags, 7.4.2.1.1
	SdrTemplate                 int8
	SdTemplate                  int8
	isCodingContextRetained     bool
	isCodingContextUsed         bool
	SdHuffAggInstanceSelection  bool
	SdHuffBMSizeSelection       int8
	SdHuffDecodeWidthSelection  int8
	SdHuffDecodeHeightSelection int8
	UseRefinementAggregation    bool
	IsHuffmanEncoded            bool

	// Symbol Dictionary AT flags 7.4.2.1.2
	SdATX []int8
	SdATY []int8

	// Symbol Dictionary refinement AT flags 7.4.2.1.3
	SdrATX []int8
	SdrATY []int8

	// Number of exported symbols, 7.4.2.1.4
	NumberOfExportedSymbols uint32

	// Number of new symbols 7.4.2.1.5
	NumberOfNewSymbols uint32

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

	// encoder parameters
	symbols         *bitmap.Bitmaps
	symbolList      []int
	symbolMap       map[int]int
	unborderSymbols bool
}

// InitEncode initializes the symbol dictionary for the encode method.
func (s *SymbolDictionary) InitEncode(symbols *bitmap.Bitmaps, symbolList []int, symbolMap map[int]int, unborderSymbols bool) error {
	const processName = "SymbolDictionary.InitEncode"
	s.SdATX = []int8{3, -3, 2, -2}
	s.SdATY = []int8{-1, -1, -2, -2}
	s.symbols = symbols
	s.symbolList = make([]int, len(symbolList))
	copy(s.symbolList, symbolList)
	// the symbols and symbollist should have the same length
	if len(s.symbolList) != s.symbols.Size() {
		return errors.Error(processName, "symbols and symbolList of different size")
	}
	s.NumberOfNewSymbols = uint32(symbols.Size())
	s.NumberOfExportedSymbols = uint32(symbols.Size())
	s.symbolMap = symbolMap
	s.unborderSymbols = unborderSymbols
	return nil
}

// BorderSize is the constant border size for symbols.
const BorderSize = 6

// Encode encodes the symbol dictionary structure into 'w' writer. Returns
// number of bytes written and the error if occurs.
func (s *SymbolDictionary) Encode(w writer.BinaryWriter) (n int, err error) {
	const processName = "SymbolDictionary.Encode"
	if s == nil {
		return 0, errors.Error(processName, "symbol dictionary not defined")
	}
	// Flags
	if n, err = s.encodeFlags(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}

	// AT Flags
	tmp, err := s.encodeATFlags(w)
	if err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += tmp

	// Refinement AT Flags
	if tmp, err = s.encodeRefinementATFlags(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += tmp

	// SDNUMEXSYM SDNUMNEWSYM
	if tmp, err = s.encodeNumSyms(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += tmp

	// encode all the symbols set for given symbol dictionary
	if tmp, err = s.encodeSymbols(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += tmp

	return n, nil
}

func (s *SymbolDictionary) encodeATFlags(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeATFlags"
	if s.IsHuffmanEncoded || s.SdTemplate != 0 {
		return 0, nil
	}
	// if sdTemplate is 0 then there are eight bytes of at flags
	max := 4
	if s.SdTemplate != 0 {
		// otherwise there is only sdAtX0 and sdAtY0
		max = 1
	}
	for i := 0; i < max; i++ {
		if err = w.WriteByte(byte(s.SdATX[i])); err != nil {
			return n, errors.Wrapf(err, processName, "SdATX[%d]", i)
		}
		n++

		if err = w.WriteByte(byte(s.SdATY[i])); err != nil {
			return n, errors.Wrapf(err, processName, "SdATY[%d]", i)
		}
		n++
	}
	return n, nil
}

func (s *SymbolDictionary) encodeRefinementATFlags(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeRefinementATFlags"
	if !s.UseRefinementAggregation || s.SdrTemplate != 0 {
		// no refinement aggregation AT flags
		return 0, nil
	}
	for i := 0; i < 2; i++ {
		if err = w.WriteByte(byte(s.SdrATX[i])); err != nil {
			return n, errors.Wrapf(err, processName, "SdrATX[%d]", i)
		}
		n++

		if err = w.WriteByte(byte(s.SdrATY[i])); err != nil {
			return n, errors.Wrapf(err, processName, "SdrATY[%d]", i)
		}
		n++
	}
	return n, nil
}

func (s *SymbolDictionary) encodeNumSyms(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeNumSyms"

	// SDNUMEXSYMS
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, s.NumberOfExportedSymbols)
	if n, err = w.Write(temp); err != nil {
		return n, errors.Wrap(err, processName, "exported symbols")
	}

	// SDNUMNEWSYMS
	binary.BigEndian.PutUint32(temp, s.NumberOfNewSymbols)
	tmp, err := w.Write(temp)
	if err != nil {
		return n, errors.Wrap(err, processName, "new symbols")
	}
	return n + tmp, nil
}

func (s *SymbolDictionary) encodeSymbols(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeSymbol"
	ectx := encoder.New()
	ectx.Init()

	symbols, err := s.symbols.SelectByIndexes(s.symbolList)
	if err != nil {
		return 0, errors.Wrap(err, processName, "initial")
	}
	mapping := map[*bitmap.Bitmap]int{}
	for i, bm := range symbols.Values {
		mapping[bm] = i
	}

	// sort symbols by height.
	symbols.SortByHeight()

	// heightClassIndexes and symbol numbers
	var hcHeight, number int

	// group the symbols by height.
	symHeights, err := symbols.GroupByHeight()
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}

	// iterate over the height classes.
	for _, heightSymbols := range symHeights.Values {
		height := heightSymbols.Values[0].Height
		deltaHeight := height - hcHeight
		if err = ectx.EncodeInteger(encoder.IADH, deltaHeight); err != nil {
			return 0, errors.Wrapf(err, processName, "IADH for dh: '%d'", deltaHeight)
		}
		hcHeight = height
		// group the symbols by width now
		symWidths, err := heightSymbols.GroupByWidth()
		if err != nil {
			return 0, errors.Wrapf(err, processName, "height: '%d'", height)
		}

		// define the width class
		var wc int
		for _, widthSymbols := range symWidths.Values {
			// iterate over all symbols with given width
			for _, bm := range widthSymbols.Values {
				// get the width of the first symbol.
				width := bm.Width
				// encode the width difference
				deltaWidth := width - wc
				if err = ectx.EncodeInteger(encoder.IADW, deltaWidth); err != nil {
					return 0, errors.Wrapf(err, processName, "IADW for dw: '%d'", deltaWidth)
				}
				// increase current width class
				wc += deltaWidth
				if err = ectx.EncodeBitmap(bm, false); err != nil {
					return 0, errors.Wrapf(err, processName, "Height: %d Width: %d", height, width)
				}
				idx := mapping[bm]
				s.symbolMap[idx] = number
				number++
			}
		}
		if err = ectx.EncodeOOB(encoder.IADW); err != nil {
			return 0, errors.Wrap(err, processName, "IADW")
		}
	}

	// encode runlength of exported symbols
	if err = ectx.EncodeInteger(encoder.IAEX, 0); err != nil {
		return 0, errors.Wrap(err, processName, "exported symbols")
	}
	if err = ectx.EncodeInteger(encoder.IAEX, len(s.symbolList)); err != nil {
		return 0, errors.Wrap(err, processName, "number of symbols")
	}
	ectx.Final()
	temp, err := ectx.WriteTo(w)
	if err != nil {
		return 0, errors.Wrap(err, processName, "writing encoder context to 'w' writer")
	}
	return int(temp), nil
}

func (s *SymbolDictionary) getSymbol(i int) (*bitmap.Bitmap, error) {
	const processName = "getSymbol"
	bm, err := s.symbols.GetBitmap(s.symbolList[i])
	if err != nil {
		return nil, errors.Wrap(err, processName, "can't get symbol")
	}
	return bm, nil
}

func (s *SymbolDictionary) encodeFlags(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeFlags"
	// skip three bits
	if err = w.SkipBits(3); err != nil {
		return 0, errors.Wrap(err, processName, "empty bits")
	}
	var bit int
	// write sdrtemplate flag
	if s.SdrTemplate > 0 {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdrtemplate")
	}
	// write first sdtemplate bit
	bit = 0
	if s.SdTemplate > 1 {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdtemplate")
	}
	// write second sdtemplate bit.
	bit = 0
	if s.SdTemplate == 1 || s.SdTemplate == 3 {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdtemplate")
	}
	// bitmap coding context retained
	bit = 0
	if s.isCodingContextRetained {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "coding context retained")
	}
	// bitmap coding context used
	bit = 0
	if s.isCodingContextUsed {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "coding context used")
	}
	// sdhuffagginst
	bit = 0
	if s.SdHuffAggInstanceSelection {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdhuffagginst")
	}
	//	sdhuffbmsize
	bit = int(s.SdHuffBMSizeSelection)
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdhuffbmsize")
	}
	// sdhuffwidth
	bit = 0
	if s.SdHuffDecodeWidthSelection > 1 {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdhuffwidth")
	}
	bit = 0
	switch s.SdHuffDecodeWidthSelection {
	case 1, 3:
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdhuffwidth")
	}
	// sdhuff height
	bit = 0
	if s.SdHuffDecodeHeightSelection > 1 {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdhuffheight")
	}
	bit = 0
	switch s.SdHuffDecodeHeightSelection {
	case 1, 3:
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdhuffheight")
	}
	// SdRefAgg
	bit = 0
	if s.UseRefinementAggregation {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdrefagg")
	}
	// SdHuff
	bit = 0
	if s.IsHuffmanEncoded {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return n, errors.Wrap(err, processName, "sdhuff")
	}
	return 2, nil
}

// GetDictionary gets the decoded dictionary symbols as a bitmap slice.
func (s *SymbolDictionary) GetDictionary() ([]*bitmap.Bitmap, error) {
	common.Log.Trace("[SYMBOL-DICTIONARY] GetDictionary begins...")
	defer func() {
		common.Log.Trace("[SYMBOL-DICTIONARY] GetDictionary finished")
		common.Log.Trace("[SYMBOL-DICTIONARY] Dictionary. \nEx: '%s', \nnew:'%s'", s.exportSymbols, s.newSymbols)
	}()

	if s.exportSymbols == nil {
		var err error
		if s.UseRefinementAggregation {
			s.sbSymCodeLen = s.getSbSymCodeLen()
		}

		if !s.IsHuffmanEncoded {
			if err = s.setCodingStatistics(); err != nil {
				return nil, err
			}
		}

		// 6.5.5. 1)
		s.newSymbols = make([]*bitmap.Bitmap, s.NumberOfNewSymbols)

		// 6.5.5. 2)
		var newSymbolsWidth []int
		if s.IsHuffmanEncoded && !s.UseRefinementAggregation {
			newSymbolsWidth = make([]int, s.NumberOfNewSymbols)
		}

		if err = s.setSymbolsArray(); err != nil {
			return nil, err
		}

		// 6.5.5 3)
		var heightClassHeight, temp int64
		s.numberOfDecodedSymbols = 0

		// 6.5.5 4 a)
		for s.numberOfDecodedSymbols < s.NumberOfNewSymbols {
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

				if differenceWidth == int64(math.MaxInt64) || s.numberOfDecodedSymbols >= s.NumberOfNewSymbols {
					break
				}

				symbolWidth += uint32(differenceWidth)
				totalWidth += symbolWidth

				//* 4 c) ii)
				if !s.IsHuffmanEncoded || s.UseRefinementAggregation {
					if !s.UseRefinementAggregation {
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
				} else if s.IsHuffmanEncoded && !s.UseRefinementAggregation {
					// 4 c) iii)
					newSymbolsWidth[s.numberOfDecodedSymbols] = int(symbolWidth)
				}
				s.numberOfDecodedSymbols++
			}

			// 6.5.5 4 d)
			if s.IsHuffmanEncoded && !s.UseRefinementAggregation {
				var bmSize int64
				if s.SdHuffBMSizeSelection == 0 {
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

// String implements the Stringer interface.
func (s *SymbolDictionary) String() string {
	sb := &strings.Builder{}
	sb.WriteString("\n[SYMBOL-DICTIONARY]\n")
	sb.WriteString(fmt.Sprintf("\t- SdrTemplate %v\n", s.SdrTemplate))
	sb.WriteString(fmt.Sprintf("\t- SdTemplate %v\n", s.SdTemplate))
	sb.WriteString(fmt.Sprintf("\t- isCodingContextRetained %v\n", s.isCodingContextRetained))
	sb.WriteString(fmt.Sprintf("\t- isCodingContextUsed %v\n", s.isCodingContextUsed))
	sb.WriteString(fmt.Sprintf("\t- SdHuffAggInstanceSelection %v\n", s.SdHuffAggInstanceSelection))
	sb.WriteString(fmt.Sprintf("\t- SdHuffBMSizeSelection %v\n", s.SdHuffBMSizeSelection))
	sb.WriteString(fmt.Sprintf("\t- SdHuffDecodeWidthSelection %v\n", s.SdHuffDecodeWidthSelection))
	sb.WriteString(fmt.Sprintf("\t- SdHuffDecodeHeightSelection %v\n", s.SdHuffDecodeHeightSelection))
	sb.WriteString(fmt.Sprintf("\t- UseRefinementAggregation %v\n", s.UseRefinementAggregation))
	sb.WriteString(fmt.Sprintf("\t- isHuffmanEncoded %v\n", s.IsHuffmanEncoded))
	sb.WriteString(fmt.Sprintf("\t- SdATX %v\n", s.SdATX))
	sb.WriteString(fmt.Sprintf("\t- SdATY %v\n", s.SdATY))
	sb.WriteString(fmt.Sprintf("\t- SdrATX %v\n", s.SdrATX))
	sb.WriteString(fmt.Sprintf("\t- SdrATY %v\n", s.SdrATY))
	sb.WriteString(fmt.Sprintf("\t- NumberOfExportedSymbols %v\n", s.NumberOfExportedSymbols))
	sb.WriteString(fmt.Sprintf("\t- NumberOfNewSymbols %v\n", s.NumberOfNewSymbols))
	sb.WriteString(fmt.Sprintf("\t- numberOfImportedSymbols %v\n", s.numberOfImportedSymbols))
	sb.WriteString(fmt.Sprintf("\t- numberOfDecodedSymbols %v\n", s.numberOfDecodedSymbols))
	return sb.String()
}

func (s *SymbolDictionary) checkInput() error {
	if s.SdHuffDecodeHeightSelection == 2 {
		common.Log.Debug("Symbol Dictionary Decode Heigh Selection: %d value not permitted", s.SdHuffDecodeHeightSelection)
	}

	if s.SdHuffDecodeWidthSelection == 2 {
		common.Log.Debug("Symbol Dictionary Decode Width Selection: %d value not permitted", s.SdHuffDecodeWidthSelection)
	}

	if s.IsHuffmanEncoded {
		if s.SdTemplate != 0 {
			common.Log.Debug("SDTemplate = %d (should be 0)", s.SdTemplate)
		}

		if !s.UseRefinementAggregation {
			if !s.UseRefinementAggregation {
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
		if s.SdHuffBMSizeSelection != 0 {
			common.Log.Debug("SdHuffBMSizeSelection should be 0")
			s.SdHuffBMSizeSelection = 0
		}

		if s.SdHuffDecodeWidthSelection != 0 {
			common.Log.Debug("SdHuffDecodeWidthSelection should be 0")
			s.SdHuffDecodeWidthSelection = 0
		}

		if s.SdHuffDecodeHeightSelection != 0 {
			common.Log.Debug("SdHuffDecodeHeightSelection should be 0")
			s.SdHuffDecodeHeightSelection = 0
		}
	}

	if !s.UseRefinementAggregation {
		if s.SdrTemplate != 0 {
			common.Log.Debug("SDRTemplate = %d (should be 0)", s.SdrTemplate)
			s.SdrTemplate = 0
		}
	}

	if !s.IsHuffmanEncoded || !s.UseRefinementAggregation {
		if s.SdHuffAggInstanceSelection {
			common.Log.Debug("SdHuffAggInstanceSelection = %d (should be 0)", s.SdHuffAggInstanceSelection)
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
	common.Log.Trace("[SYMBOL DICTIONARY] Added symbol: %s", symbol)
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
	if s.IsHuffmanEncoded {
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

	s.textRegion.setParameters(s.arithmeticDecoder, s.IsHuffmanEncoded, true, symbolWidth,
		heightClassHeight, numberOfRefinementAggregationInstances, 1, s.numberOfImportedSymbols+s.numberOfDecodedSymbols,
		0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, s.SdrTemplate, s.SdrATX, s.SdrATY, s.sbSymbols, s.sbSymCodeLen)
	return s.addSymbol(s.textRegion)
}

func (s *SymbolDictionary) decodeRefinedSymbol(symbolWidth, heightClassHeight uint32) error {
	var (
		id       int
		rdx, rdy int32
	)

	if s.IsHuffmanEncoded {
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

	if s.IsHuffmanEncoded {
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
	s.genericRefinementRegion.setParameters(s.cx, s.arithmeticDecoder, s.SdrTemplate, symWidth, hcHeight, ibo, rdx, rdy, false, s.SdrATX, s.SdrATY)
	return s.addSymbol(s.genericRefinementRegion)
}

func (s *SymbolDictionary) decodeDirectlyThroughGenericRegion(symWidth, hcHeight uint32) error {
	if s.genericRegion == nil {
		s.genericRegion = NewGenericRegion(s.r)
	}
	s.genericRegion.setParametersWithAt(false, byte(s.SdTemplate), false, false, s.SdATX, s.SdATY, symWidth, hcHeight, s.cx, s.arithmeticDecoder)
	return s.addSymbol(s.genericRegion)
}

func (s *SymbolDictionary) decodeDifferenceWidth() (int64, error) {
	if s.IsHuffmanEncoded {
		switch s.SdHuffDecodeWidthSelection {
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
				if s.SdHuffDecodeHeightSelection == 3 {
					dwNr++
				}
				t, err := s.getUserTable(dwNr)
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
	if s.IsHuffmanEncoded {
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
	switch s.SdHuffDecodeHeightSelection {
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
		math.Log(float64(s.numberOfImportedSymbols+s.NumberOfNewSymbols)) / math.Log(2)))

	if s.IsHuffmanEncoded && first < 1 {
		return 1
	}

	return first
}

func (s *SymbolDictionary) getToExportFlags() ([]int, error) {
	var (
		currentExportFlag int
		exRunLength       int32
		err               error
		totalNewSymbols   = int32(s.numberOfImportedSymbols + s.NumberOfNewSymbols)
		exportFlags       = make([]int, totalNewSymbols)
	)

	for exportIndex := int32(0); exportIndex < totalNewSymbols; exportIndex += exRunLength {
		if s.IsHuffmanEncoded {
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

		if s.SdHuffDecodeHeightSelection == 3 {
			bmNr++
		}

		if s.SdHuffDecodeWidthSelection == 3 {
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
	if !s.SdHuffAggInstanceSelection {
		t, err := huffman.GetStandardTable(1)
		if err != nil {
			return 0, err
		}
		return t.Decode(s.r)
	}
	if s.aggInstTable == nil {
		var (
			aggregationInstanceNumber int
			err                       error
		)
		if s.SdHuffDecodeHeightSelection == 3 {
			aggregationInstanceNumber++
		}

		if s.SdHuffDecodeWidthSelection == 3 {
			aggregationInstanceNumber++
		}

		if s.SdHuffBMSizeSelection == 3 {
			aggregationInstanceNumber++
		}

		s.aggInstTable, err = s.getUserTable(aggregationInstanceNumber)
		if err != nil {
			return 0, err
		}
	}
	return s.aggInstTable.Decode(s.r)
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
	s.SdrTemplate = int8(bit)

	//* Bit 10 - 11 - SDTemplate
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.SdTemplate = int8(bits & 0xf)

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

	// Bit 7 - SdHuffAggInstanceSelection
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.SdHuffAggInstanceSelection = true
	}

	// Bit 6 - SdHuffBMSizeSelection
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	s.SdHuffBMSizeSelection = int8(bit)

	// Bit 4 - 5 - SdHuffDecodeWidthSelection
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.SdHuffDecodeWidthSelection = int8(bits & 0xf)

	// Bit 2 - 3 - SdHuffDecodeWidthSelection
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.SdHuffDecodeHeightSelection = int8(bits & 0xf)

	// Bit 1
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.UseRefinementAggregation = true
	}

	// Bit 0
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.IsHuffmanEncoded = true
	}
	return nil
}

func (s *SymbolDictionary) readAtPixels(pixelsNumber int) error {
	s.SdATX = make([]int8, pixelsNumber)
	s.SdATY = make([]int8, pixelsNumber)
	var (
		b   byte
		err error
	)

	for i := 0; i < pixelsNumber; i++ {
		b, err = s.r.ReadByte()
		if err != nil {
			return err
		}
		s.SdATX[i] = int8(b)
		b, err = s.r.ReadByte()
		if err != nil {
			return err
		}
		s.SdATY[i] = int8(b)
	}
	return nil
}

func (s *SymbolDictionary) readRefinementAtPixels(numberOfAtPixels int) error {
	s.SdrATX = make([]int8, numberOfAtPixels)
	s.SdrATY = make([]int8, numberOfAtPixels)
	var (
		b   byte
		err error
	)

	for i := 0; i < numberOfAtPixels; i++ {
		b, err = s.r.ReadByte()
		if err != nil {
			return err
		}
		s.SdrATX[i] = int8(b)
		b, err = s.r.ReadByte()
		if err != nil {
			return err
		}
		s.SdrATY[i] = int8(b)
	}
	return nil
}

func (s *SymbolDictionary) readNumberOfExportedSymbols() error {
	bits, err := s.r.ReadBits(32)
	if err != nil {
		return err
	}
	s.NumberOfExportedSymbols = uint32(bits & math.MaxUint32)
	return nil
}

func (s *SymbolDictionary) readNumberOfNewSymbols() error {
	bits, err := s.r.ReadBits(32)
	if err != nil {
		return err
	}
	s.NumberOfNewSymbols = uint32(bits & math.MaxUint32)
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
			s.numberOfImportedSymbols += dict.NumberOfExportedSymbols
		}
	}
	return nil
}

func (s *SymbolDictionary) setAtPixels() error {
	if s.IsHuffmanEncoded {
		return nil
	}
	index := 1
	if s.SdTemplate == 0 {
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

	if s.UseRefinementAggregation && s.cxIAID == nil {
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
	for i := uint32(0); i < s.numberOfImportedSymbols+s.NumberOfNewSymbols; i++ {
		if toExportFlags[i] == 1 {
			var symbol *bitmap.Bitmap
			if i < s.numberOfImportedSymbols {
				symbol = s.importSymbols[i]
			} else {
				symbol = s.newSymbols[i-s.numberOfImportedSymbols]
			}
			common.Log.Trace("[SYMBOL-DICTIONARY] Add ExportedSymbol: '%s'", symbol)
			s.exportSymbols = append(s.exportSymbols, symbol)
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
	if !s.UseRefinementAggregation || s.SdrTemplate != 0 {
		return nil
	}

	if err := s.readRefinementAtPixels(2); err != nil {
		return err
	}
	return nil
}

func (s *SymbolDictionary) setRetainedCodingContexts(sd *SymbolDictionary) {
	s.arithmeticDecoder = sd.arithmeticDecoder
	s.IsHuffmanEncoded = sd.IsHuffmanEncoded
	s.UseRefinementAggregation = sd.UseRefinementAggregation
	s.SdTemplate = sd.SdTemplate
	s.SdrTemplate = sd.SdrTemplate
	s.SdATX = sd.SdATX
	s.SdATY = sd.SdATY
	s.SdrATX = sd.SdrATX
	s.SdrATY = sd.SdrATY
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
