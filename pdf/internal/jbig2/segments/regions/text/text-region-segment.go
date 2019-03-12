package text

import (
	"encoding/binary"
	"errors"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/huffman"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/pageinformation"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/regions"
	"io"
)

// TextRegionSegment is a region segment that is defined to contain text data
type TextRegionSegment struct {
	*regions.RegionSegment

	// Basic Text Region flags
	TRFlags *Flags

	// Huffman Text Region flags
	TRHuffmanFlags *HuffmanFlags

	inlineImage bool

	// The locations of the adaptive template pixel RA1,2 at X coordinate
	SymbolRegionATX []int8

	// The location of the adaptive template pixel RA1,2 at Y coordinate
	SymbolRegionATY []int8

	bm *bitmap.Bitmap
}

func New(d *container.Decoder, h *header.Header, inline bool) *TextRegionSegment {
	t := &TextRegionSegment{
		RegionSegment:  regions.NewRegionSegment(d, h),
		TRFlags:        newFlags(),
		TRHuffmanFlags: newHuffmanFlags(),
		inlineImage:    inline,
	}
	return t
}

// Decode decodes the segment from the provided reader stream
// Returns error if the segment is incorrectly encoded or the stream is empty(io.EOF).
func (t *TextRegionSegment) Decode(r *reader.Reader) (err error) {
	common.Log.Debug("[DECODE] Text Region Segment begin")
	defer func() { common.Log.Debug("[DECODE] Text Region Segment finished") }()

	// Decode the basics of the region segment
	if err = t.RegionSegment.Decode(r); err != nil {
		if err != io.EOF {
			common.Log.Error("TextRegionSegment Decode failed. %v", err)
		}
		return
	}

	// decode text region flags
	if err = t.decodeTRFlags(r); err != nil {
		return
	}

	// make 4 bytes long buffer for the first block
	var buf []byte = make([]byte, 4)

	if _, err = r.Read(buf); err != nil {
		common.Log.Error("TextRegionSegment Read first block failed. %v", err)
		return err
	}

	var (
		referencedSegments []segment.Segmenter
		symbolsNumber      int
	)

	symbolInstancesNumber := binary.BigEndian.Uint32(buf)

	for i := 0; i < t.Header.ReferredToSegmentCount; i++ {
		common.Log.Debug("Getting ReferredToSegment: %d", t.Header.ReferredToSegments[i])
		seg := t.Decoders.FindSegment(t.Header.ReferredToSegments[i])
		if seg == nil {
			common.Log.Debug("Refered Segment not found, continues...")
			continue
		}
		tp := seg.Kind()
		if tp == kind.SymbolDictionary {
			referencedSegments = append(referencedSegments, seg)

			// add the symbols number of the referenced segment
			symbolsNumber += int(seg.(segment.SymbolDictionarySegmenter).AmmountOfExportedSymbols())
		} else if tp == kind.Tables {

		}
	}

	var (
		symbolCodeLength int
		count            int = 1
	)

	for count < symbolsNumber {
		symbolCodeLength += 1
		count <<= 1
	}

	var symbols []*bitmap.Bitmap

	for _, refSegment := range referencedSegments {
		if refSegment.Kind() == kind.SymbolDictionary {
			//append the bitmap from the reference
			symbols = append(symbols, refSegment.(bitmap.BitmapsLister).ListBitmaps()...)
		}
	}

	if len(symbols) != symbolsNumber {
		common.Log.Debug("SymbolsNumber: '%d' doesn't match symbols slice length: '%d'", symbolsNumber, len(symbols))
	}

	common.Log.Debug("Text region contains: %d symbols.", symbolsNumber)

	var (
		huffmanFSTable, huffmanDSTable, huffmanDTTable,
		huffmanRDWTable, huffmanRDHTable,
		huffmanRDXTable, huffmanRDYTable, huffmanRSizeTable [][]int
	)

	var sbHuffman bool = t.TRFlags.GetValue(SbHuff) != 0

	var i int

	// set the huffman tables
	if sbHuffman {
		sbHuffFs := t.TRHuffmanFlags.GetValue(SBHUFFFS)
		switch sbHuffFs {
		case 0:
			huffmanFSTable = huffman.TableF
		case 1:
			huffmanFSTable = huffman.TableG
		default:
		}

		sbHuffDS := t.TRHuffmanFlags.GetValue(SBHUFFDS)
		switch sbHuffDS {
		case 0:
			huffmanDSTable = huffman.TableH
		case 1:
			huffmanDSTable = huffman.TableI
		case 2:
			huffmanDSTable = huffman.TableJ
		default:
		}

		sbHuffDT := t.TRHuffmanFlags.GetValue(SBHUFFDT)
		switch sbHuffDT {
		case 0:
			huffmanDTTable = huffman.TableK
		case 1:
			huffmanDTTable = huffman.TableL
		case 2:
			huffmanDTTable = huffman.TableM
		default:
		}

		sbHuffRDW := t.TRHuffmanFlags.GetValue(SBHUFFRDW)
		switch sbHuffRDW {
		case 0:
			huffmanRDWTable = huffman.TableN
		case 1:
			huffmanRDWTable = huffman.TableO
		default:
		}

		sbHuffRDH := t.TRHuffmanFlags.GetValue(SBHUFFRDH)
		switch sbHuffRDH {
		case 0:
			huffmanRDHTable = huffman.TableN
		case 1:
			huffmanRDHTable = huffman.TableO
		default:
		}

		sbHuffRDX := t.TRHuffmanFlags.GetValue(SBHUFFRDX)
		switch sbHuffRDX {
		case 0:
			huffmanRDXTable = huffman.TableN
		case 1:
			huffmanRDXTable = huffman.TableO
		default:
		}

		sbHuffRDY := t.TRHuffmanFlags.GetValue(SBHUFFRDY)
		switch sbHuffRDY {
		case 0:
			huffmanRDYTable = huffman.TableN
		case 1:
			huffmanRDYTable = huffman.TableO
		default:
		}

		sbHuffRSize := t.TRHuffmanFlags.GetValue(SBHUFFRSIZE)
		if sbHuffRSize == 0 {
			huffmanRSizeTable = huffman.TableA
		} else {
			//
		}
	}

	var runLengthTable [][]int = make([][]int, 36)
	for i := range runLengthTable {
		runLengthTable[i] = make([]int, 4)
	}

	var symbolCodeTable [][]int = make([][]int, symbolsNumber+1)
	for i := range symbolCodeTable {
		symbolCodeTable[i] = make([]int, 4)
	}

	if sbHuffman {
		r.ConsumeRemainingBits()

		for i := 0; i < 32; i++ {
			bits, err := r.ReadBits(4)
			if err != nil {
				return err
			}

			runLengthTable[i] = []int{i, int(bits), 0, 0}
		}

		bits, err := r.ReadBits(4)
		if err != nil {
			return err
		}

		runLengthTable[32] = []int{0x103, int(bits), 2, 0}

		bits, err = r.ReadBits(4)
		if err != nil {
			return err
		}

		runLengthTable[33] = []int{0x203, int(bits), 3, 0}

		bits, err = r.ReadBits(4)
		if err != nil {
			return err
		}

		runLengthTable[34] = []int{0x20b, int(bits), 7, 0}

		runLengthTable[35] = []int{0, 0, huffman.EOT, 0}

		runLengthTable, err = t.Decoders.Huffman.BuildTable(runLengthTable, 35)
		if err != nil {
			return err
		}

		for i := 0; i < symbolsNumber; i++ {
			symbolCodeTable[i] = []int{i, 0, 0, 0}
		}

		// pg: 253
		for i < symbolsNumber {

			j, _, err := t.Decoders.Huffman.DecodeInt(r, runLengthTable)
			if err != nil {
				return err
			}

			if j > 0x200 {
				for j -= 0x200; j != 0 && i < symbolsNumber; j-- {
					symbolCodeTable[i][1] = 0
					i++
				}
			} else if j > 0x100 {
				for j -= 0x100; j != 0 && j < symbolsNumber; j-- {
					symbolCodeTable[i][1] = symbolCodeTable[i-1][1]
					i++
				}
			} else {
				symbolCodeTable[i][1] = j
				i++
			}
		}
		symbolCodeTable[symbolsNumber][1] = 0
		symbolCodeTable[symbolsNumber][2] = huffman.EOT
		symbolCodeTable, err = t.Decoders.Huffman.BuildTable(symbolCodeTable, symbolsNumber)
		if err != nil {
			return err
		}
		r.ConsumeRemainingBits()
	} else {
		symbolCodeTable = nil
		arithm := t.Decoders.Arithmetic

		arithm.ResetIntStats(symbolCodeLength)
		if err := arithm.Start(r); err != nil {
			return err
		}
	}

	var (
		symbolRefine        bool = t.TRFlags.GetValue(SBRefine) != 0
		logStrips           int  = t.TRFlags.GetValue(LogSbStripes)
		defaultPixel        int  = t.TRFlags.GetValue(SbDefPixel)
		combinationOperator int  = t.TRFlags.GetValue(SbCombOp)
		transposed          bool = t.TRFlags.GetValue(Transposed) != 0
		referenceCorner     int  = t.TRFlags.GetValue(RefCorner)
		sOffset             int  = t.TRFlags.GetValue(SbDsOffset)
		template            int  = t.TRFlags.GetValue(SbRTemplate)
	)

	// Check if the symbols is refine
	if symbolRefine {
		t.Decoders.Arithmetic.ResetRefinementStats(template, nil)
	}

	bm := bitmap.New(t.BMWidth, t.BMHeight, t.Decoders)

	common.Log.Debug("bm.ReadTextRegion Reader Position: %04X", r.CurrentBytePosition())
	err = bm.ReadTextRegion(r, sbHuffman, symbolRefine, uint(symbolInstancesNumber), uint(logStrips), uint(symbolsNumber), symbolCodeTable, symbolCodeLength, symbols, defaultPixel, int64(combinationOperator), transposed, referenceCorner, sOffset,
		huffmanFSTable, huffmanDSTable, huffmanDTTable, huffmanRDWTable, huffmanRDHTable, huffmanRDXTable, huffmanRDYTable, huffmanRSizeTable,
		template, t.SymbolRegionATX, t.SymbolRegionATY,
	)

	if err != nil {
		common.Log.Error("ReadTextRegion failed: %v", err)
		return err
	}

	if t.inlineImage {
		common.Log.Debug("Looking for PageSegment")
		s := t.Decoders.FindPageSegment(t.Header.PageAssociation)
		if s == nil {
			common.Log.Debug("No page segment found for: %d page.", t.Header.PageAssociation)
			return errors.New("Can't find page segment")
		}

		// The segmenter must be a PageInformationSegment
		ps := s.(*pageinformation.PageInformationSegment)

		externalCombinationOperator := t.RegionFlags.GetValue(regions.ExternalCombinationOperator)
		err = ps.PageBitmap.Combine(bm, t.BMXLocation, t.BMYLocation, int64(externalCombinationOperator))
		if err != nil {
			common.Log.Error("PageBitmap.Combine failed: %v", err)
			return err
		}

	} else {
		bm.BitmapNumber = t.Header.SegmentNumber
		t.bm = bm
	}

	r.ConsumeRemainingBits()
	return nil
}

// GetBitmap returns TextRegionSegment bitmap
// If the segment was inline the bitmap returned would be nil
func (t *TextRegionSegment) GetBitmap() *bitmap.Bitmap {
	if t.inlineImage {
		return nil
	}
	return t.bm
}

// decodeTRFlags extracts the text region segment flag
func (t *TextRegionSegment) decodeTRFlags(r *reader.Reader) error {

	// Extract text region segment flag
	var textRegionFlagsField []byte = make([]byte, 2)

	_, err := r.Read(textRegionFlagsField)
	if err != nil {
		common.Log.Error("decoding TRFlags failed to read flags. %v", err)
		return err
	}

	flags := binary.BigEndian.Uint16(textRegionFlagsField)
	t.TRFlags.SetValue(int(flags))

	common.Log.Debug("Text Region Segment Flags: %d", flags)

	if sbHuff := t.TRFlags.GetValue(SbHuff) != 0; sbHuff {
		// Extract huffman flags
		var trHuffmanFlags []byte = make([]byte, 2)

		if _, err = r.Read(trHuffmanFlags); err != nil {
			common.Log.Error("decoding TRHuffmanFlags - failed to read huffman flags. %v", err)
			return err
		}

		flags = binary.BigEndian.Uint16(trHuffmanFlags)
		t.TRHuffmanFlags.SetValue(int(flags))

		common.Log.Debug("Text Region Segment Huffman Flags: %d", flags)
	}

	sbRefine := t.TRFlags.GetValue(SBRefine) != 0
	sbrTemplate := t.TRFlags.GetValue(SbRTemplate)

	if sbRefine && sbrTemplate == 0 {
		var b byte
		b, err = r.ReadByte()
		if err != nil {
			common.Log.Error("Text Region Segment read Adaptive template X[0] failed. %v", err)
			return err
		}
		t.SymbolRegionATX[0] = int8(b)

		b, err = r.ReadByte()
		if err != nil {
			common.Log.Error("Text Region Segment read Adaptive template Y[0] failed. %v", err)
			return err
		}
		t.SymbolRegionATY[0] = int8(b)

		b, err = r.ReadByte()
		if err != nil {
			common.Log.Error("Text Region Segment read Adaptive template X[1] failed. %v", err)
			return err
		}
		t.SymbolRegionATX[1] = int8(b)

		b, err = r.ReadByte()
		if err != nil {
			common.Log.Error("Text Region Segment read Adaptive template Y[1] failed. %v", err)
			return err
		}
		t.SymbolRegionATY[1] = int8(b)
	}

	return nil
}
