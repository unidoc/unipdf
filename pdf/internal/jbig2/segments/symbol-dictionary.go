package segments

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/huffman"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"image"
	"math"
	"strings"
	"time"
)

// SymbolDictionary is the struct that represents Symbol Dictionary Segment for JBIG2
// encoding
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
	amountOfExportedSymbols int

	// Number of new symbols 7.4.2.1.5
	amountOfNewSymbols int

	// Further parameters
	Header                  *Header
	amountOfImportedSymbols int
	importSymbols           []*bitmap.Bitmap
	amountOfDecodedSymbols  int
	newSymbols              []*bitmap.Bitmap

	// User-supplied tables
	dhTable      huffman.HuffmanTabler
	dwTable      huffman.HuffmanTabler
	bmSizeTable  huffman.HuffmanTabler
	aggInstTable huffman.HuffmanTabler

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
	sbSymCodeLen int
}

// NewSymbolDictionary creates new SymbolDictionary
func NewSymbolDictionary(h *Header, r reader.StreamReader,
) *SymbolDictionary {
	s := &SymbolDictionary{
		r:      r,
		Header: h,
	}

	return s
}

// AmmountOfExportedSymbols defines how many symbols are being exported by this SymbolDictionary
func (s *SymbolDictionary) AmountOfExportedSymbols() int {
	return s.amountOfExportedSymbols
}

// AmmountOfNewSymbols returns the ammount of new symbols defined by the Symbol Dictionary
func (s *SymbolDictionary) AmmountOfNewSymbols() int {
	return s.amountOfNewSymbols
}

// Init initialize the SymbolDicitonary segment
func (s *SymbolDictionary) Init(h *Header, r reader.StreamReader) error {
	s.Header = h
	s.r = r

	return s.parseHeader()
}

// IsHuffmanEncoded defines if the segment is HuffmanEncoded
func (s *SymbolDictionary) IsHuffmanEncoded() bool {
	return s.isHuffmanEncoded
}

// UseRefinementAggregation defines if the SymbolDictionary uses refinement aggergation
func (s *SymbolDictionary) UseRefinementAggregation() bool {
	return s.useRefinementAggregation
}

func (s *SymbolDictionary) parseHeader() (err error) {
	common.Log.Debug("[SYMBOL DICTIONARY][PARSE-HEADER] begins...")
	defer func() {
		if err != nil {
			common.Log.Debug("[SYMBOL DICTIONARY][PARSE-HEADER] failed. %v", err)
		} else {
			common.Log.Debug("[SYMBOL DICTIONARY][PARSE-HEADER] finished.")
		}
	}()
	if err = s.readRegionFlags(); err != nil {
		return
	}
	if err = s.setAtPixels(); err != nil {
		return
	}
	if err = s.setRefinementAtPixels(); err != nil {
		return
	}
	if err = s.readAmountOfExportedSymbols(); err != nil {
		return
	}
	if err = s.readAmmountOfNewSymbols(); err != nil {
		return
	}
	if err = s.setInSyms(); err != nil {
		return
	}

	if s.isCodingContextUsed {

		rtSegments := s.Header.RTSegments
		for i := len(rtSegments) - 1; i >= 0; i-- {

			if rtSegments[i].Type == 0 {
				symbolDictionary, ok := rtSegments[i].SegmentData.(*SymbolDictionary)
				if !ok {
					err = errors.Errorf("Related Segment: %v is not a Symbol Dictionary Segment", rtSegments[i])
					return
				}

				if symbolDictionary.isCodingContextUsed {
					if err = s.setRetainedCodingContexts(symbolDictionary); err != nil {
						return
					}
				}
				break
			}
		}
	}
	common.Log.Debug("%s", s.String())
	return s.checkInput()
}

func (s *SymbolDictionary) readRegionFlags() error {
	var (
		bits uint64
		bit  int
	)

	/* Bit 13 - 15 */
	_, err := s.r.ReadBits(3) // Dirty read
	if err != nil {
		return err
	}

	/* Bit 12 - SDRTemplate*/
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	s.sdrTemplate = int8(bit)

	/* Bit 10 - 11 - SDTemplate */
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.sdTemplate = int8(bits & 0xf)

	/* Bit 9 - isCodingContextRetained */
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.isCodingContextRetained = true
	}

	/* Bit 8 - isCodingContextUsed */
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.isCodingContextUsed = true
	}

	/* Bit 7 - sdHuffAggInstanceSelection */
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	s.sdHuffBMSizeSelection = int8(bit)

	/* Bit 6 - sdHuffBMSizeSelection */
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	s.sdHuffBMSizeSelection = int8(bit)

	/* Bit 4 - 5 - sdHuffDecodeWidthSelection */
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.sdHuffDecodeWidthSelection = int8(bits & 0xf)

	/* Bit 2 - 3 - sdHuffDecodeWidthSelection */
	bits, err = s.r.ReadBits(2)
	if err != nil {
		return err
	}
	s.sdHuffDecodeHeightSelection = int8(bits & 0xf)

	/* Bit 1 */
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.useRefinementAggregation = true
	}

	/* Bit 0 */
	bit, err = s.r.ReadBit()
	if err != nil {
		return err
	}
	if bit == 1 {
		s.isHuffmanEncoded = true
	}

	return nil
}

func (s *SymbolDictionary) setAtPixels() error {
	if !s.isHuffmanEncoded {
		if s.sdTemplate == 0 {
			if err := s.readAtPixels(4); err != nil {
				return err
			}
		} else {
			if err := s.readAtPixels(1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SymbolDictionary) setRefinementAtPixels() error {
	if s.useRefinementAggregation && s.sdrTemplate == 0 {
		if err := s.readRefinementAtPixels(2); err != nil {
			return err
		}
	}
	return nil
}

func (s *SymbolDictionary) setInSyms() error {
	if s.Header.RTSegments != nil {
		return s.retrieveImportSymbols()
	}
	s.importSymbols = make([]*bitmap.Bitmap, 0)
	return nil
}

func (s *SymbolDictionary) setRetainedCodingContexts(sd *SymbolDictionary) error {

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

	return nil
}

func (s *SymbolDictionary) readAtPixels(amountOfPixels int) error {
	s.sdATX = make([]int8, amountOfPixels)
	s.sdATY = make([]int8, amountOfPixels)

	var (
		b   byte
		err error
	)

	for i := 0; i < amountOfPixels; i++ {
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

func (s *SymbolDictionary) readRefinementAtPixels(amountOfAtPixels int) error {
	s.sdrATX = make([]int8, amountOfAtPixels)
	s.sdrATY = make([]int8, amountOfAtPixels)

	var (
		b   byte
		err error
	)

	for i := 0; i < amountOfAtPixels; i++ {
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

func (s *SymbolDictionary) readAmountOfExportedSymbols() error {
	bits, err := s.r.ReadBits(32)
	if err != nil {
		return err
	}
	s.amountOfExportedSymbols = int(bits & 0xffffffff)
	return nil
}

func (s *SymbolDictionary) readAmmountOfNewSymbols() error {
	bits, err := s.r.ReadBits(32)
	if err != nil {
		return err
	}
	s.amountOfNewSymbols = int(bits & 0xffffffff)
	return nil
}

func (s *SymbolDictionary) checkInput() error {
	if s.sdHuffDecodeHeightSelection == 2 {
		common.Log.Info("Symbol Dictionary Decode Heigh Selection: %d value not permitted", s.sdHuffDecodeHeightSelection)
	}

	if s.sdHuffDecodeWidthSelection == 2 {
		common.Log.Info("Symbol Dictionary Decode Width Selection: %d value not permitted", s.sdHuffDecodeWidthSelection)
	}

	if s.isHuffmanEncoded {
		if s.sdTemplate != 0 {
			common.Log.Info("SDTemplate = %d (should be 0)", s.sdTemplate)
		}

		if !s.useRefinementAggregation {
			if !s.useRefinementAggregation {
				if s.isCodingContextRetained {
					common.Log.Info("IsCodingContextRetained = true (should be false)")
					s.isCodingContextRetained = false
				}

				if s.isCodingContextUsed {
					common.Log.Info("isCodingContextUsed = true (should be false)")
					s.isCodingContextUsed = false
				}

			}
		}
	} else {
		if s.sdHuffBMSizeSelection != 0 {
			common.Log.Info("sdHuffBMSizeSelection should be 0")
			s.sdHuffBMSizeSelection = 0
		}

		if s.sdHuffDecodeWidthSelection != 0 {
			common.Log.Info("sdHuffDecodeWidthSelection should be 0")
			s.sdHuffDecodeWidthSelection = 0
		}

		if s.sdHuffDecodeHeightSelection != 0 {
			common.Log.Info("sdHuffDecodeHeightSelection should be 0")
			s.sdHuffDecodeHeightSelection = 0
		}

	}

	if !s.useRefinementAggregation {
		if s.sdrTemplate != 0 {
			common.Log.Info("SDRTemplate = %d (should be 0)", s.sdrTemplate)
			s.sdrTemplate = 0
		}
	}

	if !s.isHuffmanEncoded || !s.useRefinementAggregation {
		if s.sdHuffAggInstanceSelection != 0 {
			common.Log.Info("sdHuffAggInstanceSelection = %d (should be 0)", s.sdHuffAggInstanceSelection)
		}
	}
	return nil
}

// GetDictionary gets the decoded dictionary symbols
func (s *SymbolDictionary) GetDictionary() ([]*bitmap.Bitmap, error) {
	common.Log.Debug("[SYMBOL-DICTIONARY] GetDictionary begins...")
	defer func() {
		common.Log.Debug("[SYMBOL-DICTIONARY] GetDictionary finished")
	}()
	ts := time.Now()

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

		/* 6.5.5. 1) */
		s.newSymbols = make([]*bitmap.Bitmap, s.amountOfNewSymbols)

		/* 6.5.5. 2) */
		var newSymbolsWidth []int
		if s.isHuffmanEncoded && !s.useRefinementAggregation {
			newSymbolsWidth = make([]int, s.amountOfNewSymbols)
		}

		if err = s.setSymbolsArray(); err != nil {
			return nil, err
		}

		/* 6.5.5 3) */
		var heightClassHeight, temp int64
		s.amountOfDecodedSymbols = 0

		/* 6.5.5 4 a) */
		for s.amountOfDecodedSymbols < s.amountOfNewSymbols {
			common.Log.Debug("Decoding Symbol: %d", s.amountOfDecodedSymbols+1)
			/* 6.5.5 4 b) */
			temp, err = s.decodeHeightClassDeltaHeight()
			if err != nil {
				return nil, err
			}
			heightClassHeight += temp
			var symbolWidth, totalWidth int
			heightClassFirstSymbolIndex := s.amountOfDecodedSymbols

			/* 6.5.5 4 c) */

			// Repeat until OOB - OOB sends a break
			for {
				common.Log.Debug("HeightClassHeight: %d", heightClassHeight)
				/* 4 c) i) */
				var differenceWidth int64
				differenceWidth, err = s.decodeDifferenceWidth()
				if err != nil {
					return nil, err
				}

				common.Log.Debug("Difference Width: %d", differenceWidth)
				common.Log.Debug("MaxInt64: %d", math.MaxInt64)

				/*
				 * If result is OOB, then all the symbols in this height class has been decoded;
				 * proceed to step 4 d). Also exit, if the expected number of symbols have been
				 * decoded.
				 * The latter exit condition guards against pathological cases where a symbol's DW
				 * never contains OOB and thus never terminates
				 */
				if differenceWidth == math.MaxInt64 || s.amountOfDecodedSymbols >= s.amountOfNewSymbols {
					common.Log.Debug("Break")
					break
				}

				symbolWidth += int(differenceWidth)
				totalWidth += symbolWidth

				/* 4 c) ii) */
				if !s.isHuffmanEncoded || s.useRefinementAggregation {

					if !s.useRefinementAggregation {
						/* 6.5.8.1 - Directly coded */
						err = s.decodeDirectlyThroughGenericRegion(symbolWidth, int(heightClassHeight))
						if err != nil {
							return nil, err
						}
					} else {
						/* 6.5.8.2 - Refinement / Aggregate -coded*/
						err = s.decodeAggregate(symbolWidth, int(heightClassHeight))
						if err != nil {
							return nil, err
						}
					}
				} else if s.isHuffmanEncoded && !s.useRefinementAggregation {
					/* 4 c) iii) */
					newSymbolsWidth[s.amountOfDecodedSymbols] = symbolWidth
				}
				s.amountOfDecodedSymbols++
			}

			/* 6.5.5 4 d) */
			if s.isHuffmanEncoded && !s.useRefinementAggregation {
				var bmSize int64
				if s.sdHuffBMSizeSelection == 0 {
					var st huffman.HuffmanTabler
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
					bmSize, int(heightClassHeight), totalWidth)
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
		/* 5) */
		/* 6.5.10 1) - 5 */
		exFlags, err := s.getToExportFlags()
		if err != nil {
			return nil, err
		}
		s.setExportedSymbols(exFlags)
	}

	common.Log.Debug("PERFORMANCE TEST: Symbol Decoding %d ns", time.Since(ts).Nanoseconds())

	return s.exportSymbols, nil
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

func (s *SymbolDictionary) decodeHeightClassBitmap(
	heightClassCollectiveBitmap *bitmap.Bitmap,
	heightClassFirstSymbol, heightClassHeight int,
	newSymbolsWidths []int,
) error {
	for i := heightClassFirstSymbol; i < s.amountOfDecodedSymbols; i++ {

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

// decodeAggregate - 6.5.8.2.1
func (s *SymbolDictionary) decodeAggregate(symbolWidth, heightClassHeight int) error {
	var (
		amountOfRefinementAggregationInstanves int64
		err                                    error
	)

	if s.isHuffmanEncoded {

		amountOfRefinementAggregationInstanves, err = s.huffDecodeRefAggNInst()
		if err != nil {
			return err
		}
	} else {
		t, err := s.arithmeticDecoder.DecodeInt(s.cxIAAI)
		if err != nil {
			return err
		}
		amountOfRefinementAggregationInstanves = int64(t)
	}

	if amountOfRefinementAggregationInstanves > 1 {
		// 6.5.8.2

		return s.decodeThroughTextRegion(symbolWidth, heightClassHeight, amountOfRefinementAggregationInstanves)
	} else if amountOfRefinementAggregationInstanves == 1 {
		return s.decodeRefinedSymbol(symbolWidth, heightClassHeight)
	}

	return nil
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
			var aggregationInstanceNumber int

			if s.sdHuffDecodeHeightSelection == 3 {
				aggregationInstanceNumber += 1
			}

			if s.sdHuffDecodeWidthSelection == 3 {
				aggregationInstanceNumber += 1
			}

			if s.sdHuffBMSizeSelection == 3 {
				aggregationInstanceNumber += 1
			}

			var err error
			s.aggInstTable, err = s.getUserTable(aggregationInstanceNumber)
			if err != nil {
				return 0, err
			}
		}
		return s.aggInstTable.Decode(s.r)
	}
	return 0, nil
}

func (s *SymbolDictionary) decodeThroughTextRegion(
	symbolWidth, heightClassHeight int, amountOfRefinementAggregationInstances int64,
) error {
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

	s.textRegion.SetParameters(s.arithmeticDecoder, s.isHuffmanEncoded, true, symbolWidth,
		heightClassHeight, amountOfRefinementAggregationInstances, 1, (s.amountOfImportedSymbols + s.amountOfDecodedSymbols), 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, s.sdrTemplate, s.sdrATX, s.sdrATY, s.sbSymbols, s.sbSymCodeLen)

	return s.addSymbol(s.textRegion)
}

func (s *SymbolDictionary) decodeRefinedSymbol(
	symbolWidth, heightClassHeight int,
) error {
	var (
		id, rdx, rdy int
	)

	if s.isHuffmanEncoded {
		/* 2) - 4)*/
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

		rdx = int(iv)

		iv, err = st.Decode(s.r)
		if err != nil {
			return err
		}
		rdy = int(iv)

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
		s.r.Align()
		// make sure that the processed bytes are equal to the value read in step 5
	}
	return nil
}

func (s *SymbolDictionary) decodeNewSymbols(
	symWidth, hcHeight int, ibo *bitmap.Bitmap, rdx, rdy int,
) (err error) {
	if s.genericRefinementRegion == nil {

		s.genericRefinementRegion = newGenericRefinementRegion(s.r, nil)

		if s.arithmeticDecoder == nil {
			s.arithmeticDecoder, err = arithmetic.New(s.r)
			if err != nil {
				return
			}
		}

		if s.cx == nil {
			s.cx = arithmetic.NewStats(65536, 1)
		}
	}

	s.genericRefinementRegion.SetParameters(s.cx, s.arithmeticDecoder, s.sdrTemplate, symWidth, hcHeight, ibo, rdx, rdy, false, s.sdrATX, s.sdrATY)

	return s.addSymbol(s.genericRefinementRegion)
}

func (s *SymbolDictionary) decodeDirectlyThroughGenericRegion(
	symWidth, hcHeight int,
) error {
	if s.genericRegion == nil {
		s.genericRegion = NewGenericRegion(s.r)
	}

	s.genericRegion.setParametersWithAt(false, byte(s.sdTemplate), false, false, s.sdATX, s.sdATY, symWidth, hcHeight, s.cx, s.arithmeticDecoder)

	return s.addSymbol(s.genericRegion)
}

func (s *SymbolDictionary) addSymbol(region Regioner) error {
	symbol, err := region.GetRegionBitmap()
	if err != nil {
		return err
	}
	s.newSymbols[s.amountOfDecodedSymbols] = symbol
	s.sbSymbols = append(s.sbSymbols, symbol)
	return nil
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
					dwNr += 1
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
		return int64(i), nil
	}
	return 0, nil
}

func (s *SymbolDictionary) decodeHeightClassDeltaHeight() (int64, error) {
	if s.isHuffmanEncoded {
		return s.decodeHeightClassDeltaHeightWithHuffman()
	} else {

		common.Log.Debug("Reader: %d", s.r.StreamPosition())
		i, err := s.arithmeticDecoder.DecodeInt(s.cxIADH)
		if err != nil {
			return 0, err
		}

		return int64(i), nil
	}
}

// decodeHeightClassDeltaHeightWithHuffman - 6.5.6 if isHuffmanEncoded
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
	bmSize int64, heightClassHeight, totalWidth int,
) (*bitmap.Bitmap, error) {
	if bmSize == 0 {
		heightClassColleciveBitmap := bitmap.New(totalWidth, heightClassHeight)
		var (
			b   byte
			err error
		)
		for i := 0; i < len(heightClassColleciveBitmap.Data); i++ {
			b, err = s.r.ReadByte()
			if err != nil {
				return nil, err
			}
			// common.Log.Debug("Setting HeightClassCollectiveBitmap Byte: %08b at index: %d", b, i)
			if err = heightClassColleciveBitmap.SetByte(i, b); err != nil {
				return nil, err
			}
		}
		return heightClassColleciveBitmap, nil
	} else {
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
}

func (s *SymbolDictionary) setExportedSymbols(toExportFlags []int) {
	for i := 0; i < s.amountOfImportedSymbols+s.amountOfNewSymbols; i++ {
		if toExportFlags[i] == 1 {
			if i < s.amountOfImportedSymbols {
				s.exportSymbols = append(s.exportSymbols, s.importSymbols[i])
			} else {
				s.exportSymbols = append(s.exportSymbols, s.newSymbols[i-s.amountOfImportedSymbols])
			}
		}
	}
}

func (s *SymbolDictionary) getToExportFlags() ([]int, error) {
	var (
		currentExportFlag int
		exRunLength       int
		err               error
		totalNewSymbols   = s.amountOfImportedSymbols + s.amountOfNewSymbols
		exportFlags       = make([]int, totalNewSymbols)
	)

	for exportIndex := 0; exportIndex < totalNewSymbols; exportIndex += exRunLength {

		if s.isHuffmanEncoded {
			t, err := huffman.GetStandardTable(1)
			if err != nil {
				return nil, err
			}

			i, err := t.Decode(s.r)
			if err != nil {
				return nil, err
			}
			exRunLength = int(i)
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

func (s *SymbolDictionary) huffDecodeBmSize() (int64, error) {
	if s.bmSizeTable == nil {
		var bmNr int

		if s.sdHuffDecodeHeightSelection == 3 {
			bmNr += 1
		}

		if s.sdHuffDecodeWidthSelection == 3 {
			bmNr += 1
		}

		var err error

		s.bmSizeTable, err = s.getUserTable(bmNr)
		if err != nil {
			return 0, err
		}
	}

	return s.bmSizeTable.Decode(s.r)
}

// getSbSymCodeLen 6.5.8.2.3 - Setting SBSYMCODES
func (s *SymbolDictionary) getSbSymCodeLen() int {
	first := int(
		math.Ceil(
			math.Log(float64(s.amountOfImportedSymbols+s.amountOfNewSymbols)) / math.Log(2)))

	if s.isHuffmanEncoded && first < 1 {
		return 1
	}

	return first
}

// setSymbolsArray 6.5.8.2.4 - Setting SBSYMS
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

func (s *SymbolDictionary) getUserTable(tablePosition int) (huffman.HuffmanTabler, error) {
	var tableCounter int

	for _, h := range s.Header.RTSegments {
		if h.Type == 53 {
			if tableCounter == tablePosition {
				d, err := h.GetSegmentData()
				if err != nil {
					return nil, err
				}
				t := d.(huffman.Tabler)
				return huffman.NewEncodedTable(t)
			} else {
				tableCounter += 1
			}
		}
	}
	return nil, nil
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
				return errors.Errorf("Provided Segment Data is not a SymbolDictionary Segment: %T", sd)
			}

			relatedDict, err := dict.GetDictionary()
			if err != nil {
				return errors.Wrapf(err, "Related segment with index: %d getDictionary failed.", h.SegmentNumber)
			}
			s.importSymbols = append(s.importSymbols, relatedDict...)
			s.amountOfImportedSymbols += dict.amountOfExportedSymbols
		}
	}
	return nil
}

// String implements the Stringer interface
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
	sb.WriteString(fmt.Sprintf("\t- amountOfExportedSymbols %v\n", s.amountOfExportedSymbols))
	sb.WriteString(fmt.Sprintf("\t- amountOfNewSymbols %v\n", s.amountOfNewSymbols))
	sb.WriteString(fmt.Sprintf("\t- amountOfImportedSymbols %v\n", s.amountOfImportedSymbols))
	sb.WriteString(fmt.Sprintf("\t- amountOfDecodedSymbols %v\n", s.amountOfDecodedSymbols))

	return sb.String()
}
