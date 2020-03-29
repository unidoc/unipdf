/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package document

import (
	"encoding/binary"
	"io"
	"runtime/debug"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document/segments"
	"github.com/unidoc/unipdf/v3/internal/jbig2/encoder/classer"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// fileHeaderID first byte slices of the jbig2 encoded file, see D.4.1.
var fileHeaderID = []byte{0x97, 0x4A, 0x42, 0x32, 0x0D, 0x0A, 0x1A, 0x0A}

// Document is the jbig2 document model containing pages and global segments.
// By creating new document with method New or NewWithGlobals all the jbig2
// encoded data segment headers are decoded. In order to decode whole
// document, all of it's pages should be decoded using GetBitmap method.
// PDF encoded documents should contains only one Page with the number 1.
type Document struct {
	// Pages contains all pages of this document.
	Pages map[int]*Page
	// NumberOfPagesUnknown defines if the amout of the pages is known.
	NumberOfPagesUnknown bool
	// NumberOfPages - D.4.3 - Number of pages field (4 bytes). Only presented if NumberOfPagesUnknown is true.
	NumberOfPages uint32
	// GBUseExtTemplate defines wether extended Template is used.
	GBUseExtTemplate bool
	// SubInputStream is the source data stream wrapped into a SubInputStream.
	InputStream reader.StreamReader
	// GlobalSegments contains all segments that aren't associated with a page.
	GlobalSegments *Globals
	// OrganizationType is the document segment organization.
	OrganizationType segments.OrganizationType

	// Encoder variables
	// Classer is the document encoding classifier.
	Classer *classer.Classer
	// XRes and YRes are the PPI for the x and y direction.
	XRes, YRes int
	// FullHeaders is a flag that defines if the encoder should produce full JBIG2 files.
	FullHeaders bool

	// CurrentSegmentNumber current symbol number.
	CurrentSegmentNumber uint32

	// AverageTemplates are the grayed templates.
	AverageTemplates *bitmap.Bitmaps
	BaseIndexes      []int

	Refinement  bool
	RefineLevel int

	fileHeaderLength uint8

	w *writer.Buffer

	EncodeGlobals bool
	// globalSymbolsNumber is the number of globally defined symbols.
	globalSymbolsNumber int
	// singleUseSymbols is the mapping between the page and exclusively used components for that page.
	// This means that if some symbols are mapped to the page 1 only then that information would be stored here.
	singleUseSymbols map[int][]int
	// pageComponents is the mapping from page number to the list of connected components for that page.
	// i.e.  1 map[1]=[]int{1,2,3} means that page 1 contains components [1,2,3]
	pageComponents map[int][]int
	// symbolsUsed is the array that contains the number of a symbol being used.
	// Array index is the class-id of the symbol, where the value is it's occurrence.
	symbolsUsed []int
	// symbolIndexMap is the mapping between symbol class and the number it represents.
	symbolIndexMap map[int]int
}

// DecodeDocument decodes provided document based on the provided 'input' data stream
// and with optional Global defined segments 'globals'.
func DecodeDocument(input reader.StreamReader, globals *Globals) (*Document, error) {
	return decodeWithGlobals(input, globals)
}

// InitEncodeDocument initializes the jbig2 document for the encoding process.
func InitEncodeDocument(fullHeaders bool) *Document {
	return &Document{FullHeaders: fullHeaders, w: writer.BufferedMSB(), Pages: map[int]*Page{}, singleUseSymbols: map[int][]int{}, symbolIndexMap: map[int]int{}, pageComponents: map[int][]int{}}
}

// AddGenericPage creates the jbig2 page based on the provided bitmap. The data provided
func (d *Document) AddGenericPage(bm *bitmap.Bitmap, duplicateLineRemoval bool) (err error) {
	const processName = "Document.AddGenericPage"
	// check if this is PDFMode and there is already a page
	if !d.FullHeaders && d.NumberOfPages != 0 {
		return errors.Error(processName, "document already contains page. FileMode disallows adding more than one page")
	}
	// initialize page
	page := &Page{
		Segments:    []*segments.Header{},
		Bitmap:      bm,
		Document:    d,
		FinalHeight: bm.Height,
		FinalWidth:  bm.Width,
		IsLossless:  true,
	}
	page.PageNumber = int(d.nextPageNumber())
	d.Pages[page.PageNumber] = page

	// add page information segment.
	page.AddPageInformationSegment()
	// add the generic region for given bitmap
	if err = page.AddGenericRegion(bm, 0, 0, 0, segments.TImmediateGenericRegion, duplicateLineRemoval); err != nil {
		return errors.Wrap(err, processName, "")
	}
	if d.FullHeaders {
		// finish the page
		page.AddEndOfPageSegment()
	}
	return nil
}

// AddClassifiedPage adds the bitmap page with a classification 'method'.
func (d *Document) AddClassifiedPage(bm *bitmap.Bitmap, method classer.Method) (err error) {
	const processName = "Document.AddClassifiedPage"
	// check if this is PDFMode and there is already a page
	if !d.FullHeaders && d.NumberOfPages != 0 {
		return errors.Error(processName, "document already contains page. FileMode disallows adding more than one page")
	}
	// initialize the classer if not set yet
	if d.Classer == nil {
		if d.Classer, err = classer.Init(classer.DefaultSettings()); err != nil {
			return errors.Wrap(err, processName, "")
		}
	}

	pageNumber := int(d.nextPageNumber())
	p := &Page{
		Segments:    []*segments.Header{},
		Bitmap:      bm,
		Document:    d,
		FinalHeight: bm.Height,
		FinalWidth:  bm.Width,
		PageNumber:  pageNumber,
	}
	d.Pages[pageNumber] = p
	switch method {
	case classer.RankHaus:
		p.EncodingMethod = RankHausEM
	case classer.Correlation:
		p.EncodingMethod = CorrelationEM
	}

	p.AddPageInformationSegment()
	if err = d.Classer.AddPage(bm, pageNumber, method); err != nil {
		return errors.Wrap(err, processName, "")
	}

	if d.FullHeaders {
		// finish the page
		p.AddEndOfPageSegment()
	}
	return nil
}

// completeClassifiedPages completes the pages with the classification encoding type.
// creates global symbol dictionary segment.
func (d *Document) completeClassifiedPages() (err error) {
	const processName = "completeClassifiedPages"

	// check if the Classer is initialized
	if d.Classer == nil {
		return nil
	}
	// map symbol number to the times it was used
	d.symbolsUsed = make([]int, d.Classer.UndilatedTemplates.Size())
	for i := 0; i < d.Classer.ClassIDs.Size(); i++ {
		classID, err := d.Classer.ClassIDs.Get(i)
		if err != nil {
			return errors.Wrapf(err, processName, "class with id: '%d'", i)
		}
		d.symbolsUsed[classID]++
	}

	// find if there are any symbols used on multiple pages at once - Globals.
	var globalSymbols []int
	for i := 0; i < d.Classer.UndilatedTemplates.Size(); i++ {
		if d.NumberOfPages == 1 || d.symbolsUsed[i] > 1 {
			globalSymbols = append(globalSymbols, i)
		}
	}
	// build page components map
	var (
		page *Page
		ok   bool
	)
	// iterate over all classified pages.
	// create a page components map which maps the page number to the list of connected components
	// for that page.
	for i, pageNumber := range *d.Classer.ComponentPageNumbers {
		if page, ok = d.Pages[pageNumber]; !ok {
			return errors.Errorf(processName, "page: '%d' not found", i)
		}
		// make sure that the page is not Generic Encoded.
		if page.EncodingMethod == GenericEM {
			common.Log.Error("Generic page with number: '%d' mapped as classified page", i)
			continue
		}
		// add component to the given page number components.
		d.pageComponents[pageNumber] = append(d.pageComponents[pageNumber], i)

		// check if the symbol was used on this page exclusively.
		symbol, err := d.Classer.ClassIDs.Get(i)
		if err != nil {
			return errors.Wrapf(err, processName, "no such classID: %d", i)
		}
		if d.symbolsUsed[symbol] == 1 && d.NumberOfPages != 1 {
			sus := append(d.singleUseSymbols[pageNumber], symbol)
			d.singleUseSymbols[pageNumber] = sus
		}
	}
	if err = d.Classer.ComputeLLCorners(); err != nil {
		return errors.Wrap(err, processName, "")
	}

	// TODO: set the page number of the symold dictionary depending if there are more than one page.
	// This is related with the mapping[*bitmap.Bitmap]int
	// create global symbols
	if _, err = d.addSymbolDictionary(0, d.Classer.UndilatedTemplates, globalSymbols, d.symbolIndexMap, false); err != nil {
		return errors.Wrap(err, processName, "")
	}
	return nil
}

func (d *Document) produceClassifiedPages() (err error) {
	const processName = "produceClassifiedPages"
	if d.Classer == nil {
		return nil
	}
	var (
		page     *Page
		ok       bool
		globalSD *segments.Header
	)
	// iterate over all pages and find if it is of type different than GenericEM.
	for i := 1; i <= int(d.NumberOfPages); i++ {
		if page, ok = d.Pages[i]; !ok {
			return errors.Errorf(processName, "page: '%d' not found", i)
		}
		if page.EncodingMethod == GenericEM {
			continue
		}
		if globalSD == nil {
			if globalSD, err = d.GlobalSegments.GetSymbolDictionary(); err != nil {
				return errors.Wrap(err, processName, "")
			}
		}
		// produce new classified page.
		if err = d.produceClassifiedPage(page, globalSD); err != nil {
			return errors.Wrapf(err, processName, "page: '%d'", i)
		}
	}
	return nil
}

func (d *Document) produceClassifiedPage(page *Page, globalSD *segments.Header) (err error) {
	const processName = "produceClassifiedPage"
	// if given page contains it's own unique symbols
	// create additional symbols dictionary
	var secondSymbolMap map[int]int
	numSyms := d.globalSymbolsNumber
	refererTo := []*segments.Header{globalSD}

	// check if there are symbols used only for given page. This would be present only if there is
	// more then 1 page in the document.
	if len(d.singleUseSymbols[page.PageNumber]) > 0 {
		// create new symbols dictionary
		secondSymbolMap = map[int]int{}
		extraSDHeader, err := d.addSymbolDictionary(page.PageNumber, d.Classer.UndilatedTemplates, d.singleUseSymbols[page.PageNumber], secondSymbolMap, false)
		if err != nil {
			return errors.Wrap(err, processName, "")
		}
		refererTo = append(refererTo, extraSDHeader)
		numSyms += len(d.singleUseSymbols[page.PageNumber])
	}
	// get components array for given page.
	comps := d.pageComponents[page.PageNumber]
	common.Log.Debug("Page: '%d' comps: %v", page.PageNumber, comps)
	// add the text region to the page
	page.addTextRegionSegment(refererTo, d.symbolIndexMap, secondSymbolMap, d.pageComponents[page.PageNumber],
		d.Classer.PtaLL, d.Classer.UndilatedTemplates, d.Classer.ClassIDs, nil,
		log2up(numSyms), len(d.pageComponents[page.PageNumber]))

	return nil
}

func log2up(v int) int {
	r := 0
	isPow2 := (v & (v - 1)) == 0

	v >>= 1
	for ; v != 0; v >>= 1 {
		r++
	}
	if isPow2 {
		return r
	}
	return r + 1
}

func (d *Document) addSymbolDictionary(
	pageNumber int, symbols *bitmap.Bitmaps, symbolList []int,
	symbolMap map[int]int, unborderSymbols bool,
) (*segments.Header, error) {
	const processName = "addSymbolDictionary"
	// add symbolTable
	sd := &segments.SymbolDictionary{}
	if err := sd.InitEncode(symbols, symbolList, symbolMap, unborderSymbols); err != nil {
		return nil, err
	}

	sh := &segments.Header{
		Type:            segments.TSymbolDictionary,
		PageAssociation: pageNumber,
		SegmentData:     sd,
	}
	if pageNumber == 0 {
		if d.GlobalSegments == nil {
			d.GlobalSegments = &Globals{}
		}
		d.GlobalSegments.AddSegment(sh)
		return sh, nil
	}

	// find the page at Number: 'pageNumber'
	// add the symbol dictionary after page info header
	page, ok := d.Pages[pageNumber]
	if !ok {
		return nil, errors.Errorf(processName, "page: '%d' not found", pageNumber)
	}
	var (
		i   int
		seg *segments.Header
	)
	for i, seg = range page.Segments {
		if seg.Type == segments.TPageInformation {
			break
		}
	}
	// the segment should be after page info
	i++
	page.Segments = append(page.Segments, nil)
	copy(page.Segments[i+1:], page.Segments[i:])
	page.Segments[i] = sh

	return sh, nil
}

func (d *Document) completeSymbols() (err error) {
	const processName = "completeSymbols"
	// check if classer was used.
	if d.Classer == nil {
		return nil
	}

	if d.Classer.UndilatedTemplates == nil {
		return errors.Error(processName, "no templates defined for the classer")
	}

	// select the symbols used only once per page
	// map these symbols
	singlePage := len(d.Pages) == 1
	symbolsUsed := make([]int, d.Classer.UndilatedTemplates.Size())
	var n int
	for i := 0; i < d.Classer.ClassIDs.Size(); i++ {
		n, err = d.Classer.ClassIDs.Get(i)
		if err != nil {
			return errors.Wrap(err, processName, "class ID's")
		}
		symbolsUsed[n]++
	}

	var multiuseSymbols []int
	for i := 0; i < d.Classer.UndilatedTemplates.Size(); i++ {
		if symbolsUsed[i] == 0 {
			return errors.Error(processName, "no symbols instances found for given class? ")
		}
		if symbolsUsed[i] > 1 || singlePage {
			multiuseSymbols = append(multiuseSymbols, i)
		}
	}
	d.globalSymbolsNumber = len(multiuseSymbols)

	// build page components map for page number to the list of connected components for that page.
	var pageNum, symNum int
	for i := 0; i < d.Classer.ComponentPageNumbers.Size(); i++ {
		pageNum, err = d.Classer.ComponentPageNumbers.Get(i)
		if err != nil {
			return errors.Wrapf(err, processName, "page: '%d' not found in the classer pagenumbers", i)
		}
		symNum, err = d.Classer.ClassIDs.Get(i)
		if err != nil {
			return errors.Wrapf(err, processName, "can't get symbol for page '%d' from classer", pageNum)
		}
		if symbolsUsed[symNum] == 1 && !singlePage {
			d.singleUseSymbols[pageNum] = append(d.singleUseSymbols[pageNum], symNum)
		}
	}

	if err = d.Classer.ComputeLLCorners(); err != nil {
		return errors.Wrap(err, processName, "")
	}

	return nil
}

// AutoThreshold gathers classes of symbols and uses a single representative to stand for them all.
// func (d *Document) AutoThreshold() error {
// 	bms := d.Classer.Pixat
// 	for i := 0; i < len(bm); i++ {
// 		bm := bms[i]
//
// 		for j := i + 1; j < len(bms); j++ {
// 			if bm.Equivalent(bms[j]) {
// 				if err := d.uniteTemplatesWithIndexes(i, j); err != nil {
// 					return err
// 				}
// 				j--
// 			}
// 		}
// 	}
// 	return nil
// }

// Encode encodes the given document and stores into 'w' writer.
func (d *Document) Encode() (data []byte, err error) {
	const processName = "Document.Encode"
	var n, temp int
	// if the full headers flag is on, encode file header
	if d.FullHeaders {
		if n, err = d.encodeFileHeader(d.w); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
	}
	var (
		ok   bool
		seg  *segments.Header
		page *Page
	)
	// complete classified pages
	if err = d.completeClassifiedPages(); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	// produce classified pages
	if err = d.produceClassifiedPages(); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	if d.GlobalSegments != nil {
		for _, seg = range d.GlobalSegments.Segments {
			if err = d.encodeSegment(seg, &n); err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		}
	}

	for i := 1; i <= int(d.NumberOfPages); i++ {
		if page, ok = d.Pages[i]; !ok {
			return nil, errors.Errorf(processName, "page: '%d' not found", i)
		}
		for _, seg = range page.Segments {
			if err = d.encodeSegment(seg, &n); err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		}
	}

	// with full headers the End Of File header must be encoded
	if d.FullHeaders {
		if temp, err = d.encodeEOFHeader(d.w); err != nil {
			return nil, errors.Wrap(err, processName, "")
		}
		n += temp
	}

	// get the encoded data from the 'w' writer
	data = d.w.Data()
	if len(data) != n {
		common.Log.Debug("Bytes written (n): '%d' is not equal to the length of the data encoded: '%d'", n, len(data))
	}
	return data, nil
}

func (d *Document) encodeSegment(seg *segments.Header, n *int) error {
	const processName = "encodeSegment"
	seg.SegmentNumber = d.nextSegmentNumber()

	temp, err := seg.Encode(d.w)
	if err != nil {
		return errors.Wrapf(err, processName, "segment: '%d'", seg.SegmentNumber)
	}
	*n += temp
	return nil
}

// GetGlobalSegment implements segments.Documenter interface.
func (d *Document) GetGlobalSegment(i int) (*segments.Header, error) {
	h, err := d.GlobalSegments.GetSegment(i)
	if err != nil {
		return nil, errors.Wrap(err, "GetGlobalSegment", "")
	}
	return h, nil
}

// GetNumberOfPages gets the amount of Pages in the given document.
func (d *Document) GetNumberOfPages() (uint32, error) {
	if d.NumberOfPagesUnknown || d.NumberOfPages == 0 {
		if len(d.Pages) == 0 {
			if err := d.mapData(); err != nil {
				return 0, errors.Wrap(err, "Document.GetNumberOfPages", "")
			}
		}
		return uint32(len(d.Pages)), nil
	}
	return d.NumberOfPages, nil
}

// GetPage implements segments.Documenter interface.
// NOTE: in order to decode all document images, get page by page (page numeration starts from '1') and
// decode them by calling 'GetBitmap' method.
func (d *Document) GetPage(pageNumber int) (segments.Pager, error) {
	const processName = "Document.GetPage"
	if pageNumber < 0 {
		common.Log.Debug("JBIG2 Page - GetPage: %d. Page cannot be lower than 0. %s", pageNumber, debug.Stack())
		return nil, errors.Errorf(processName, "invalid jbig2 document - provided invalid page number: %d", pageNumber)
	}

	if pageNumber > len(d.Pages) {
		common.Log.Debug("Page not found: %d. %s", pageNumber, debug.Stack())
		return nil, errors.Error(processName, "invalid jbig2 document - page not found")
	}

	p, ok := d.Pages[pageNumber]
	if !ok {
		common.Log.Debug("Page not found: %d. %s", pageNumber, debug.Stack())
		return nil, errors.Errorf(processName, "invalid jbig2 document - page not found")
	}

	return p, nil
}

/**

Private document methods

*/

func (d *Document) determineRandomDataOffsets(segmentHeaders []*segments.Header, offset uint64) {
	if d.OrganizationType != segments.ORandom {
		return
	}

	for _, s := range segmentHeaders {
		s.SegmentDataStartOffset = offset
		offset += s.SegmentDataLength
	}
}

// encodeEOFHeader writes the end of file header segment into 'w' writer.
func (d *Document) encodeEOFHeader(w writer.BinaryWriter) (n int, err error) {
	h := &segments.Header{SegmentNumber: d.nextSegmentNumber(), Type: segments.TEndOfFile}
	if n, err = h.Encode(w); err != nil {
		return 0, errors.Wrap(err, "encodeEOFHeader", "")
	}
	return n, nil
}

// encodeFileHeader writes the file header segment into the 'w' writer.
func (d *Document) encodeFileHeader(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeFileHeader"
	// file header contains following fields:
	// ID string - constant 8-byte sequence
	n, err = w.Write(fileHeaderID)
	if err != nil {
		return n, errors.Wrap(err, processName, "id")
	}

	// file header flags one byte field
	// where:
	//	0th bit - file organization type - '1' for sequential, '0' for random-access
	// 	1st bit - unknown number of pages - '1' if true
	// this encoder stores only sequential organization type with known number of pages
	// thus file header flags are equal to 0x01 byte
	if err = w.WriteByte(0x01); err != nil {
		return n, errors.Wrap(err, processName, "flags")
	}
	n++

	temp := make([]byte, 4)
	// last 4 bytes is the number of pages - used if the number of pages is known
	binary.BigEndian.PutUint32(temp, d.NumberOfPages)
	nt, err := w.Write(temp)
	if err != nil {
		return nt, errors.Wrap(err, processName, "page number")
	}
	n += nt
	return n, nil
}

func (d *Document) isFileHeaderPresent() (bool, error) {
	d.InputStream.Mark()

	for _, magicByte := range fileHeaderID {
		b, err := d.InputStream.ReadByte()
		if err != nil {
			return false, err
		}

		if magicByte != b {
			d.InputStream.Reset()
			return false, nil
		}
	}

	d.InputStream.Reset()
	return true, nil
}

func (d *Document) mapData() error {
	const processName = "mapData"
	// Get the header list
	var (
		segmentHeaders []*segments.Header
		offset         int64
		kind           segments.Type
	)

	isFileHeaderPresent, err := d.isFileHeaderPresent()
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	// Parse the file header if exists.
	if isFileHeaderPresent {
		if err = d.parseFileHeader(); err != nil {
			return errors.Wrap(err, processName, "")
		}
		offset += int64(d.fileHeaderLength)
		d.FullHeaders = true
	}

	var (
		page       *Page
		reachedEOF bool
	)

	// type 51 is the EndOfFile segment kind
	for kind != 51 && !reachedEOF {
		// get new segment
		segment, err := segments.NewHeader(d, d.InputStream, offset, d.OrganizationType)
		if err != nil {
			return errors.Wrap(err, processName, "")
		}

		common.Log.Trace("Decoding segment number: %d, Type: %s", segment.SegmentNumber, segment.Type)

		kind = segment.Type
		if kind != segments.TEndOfFile {
			if segment.PageAssociation != 0 {
				page = d.Pages[segment.PageAssociation]
				if page == nil {
					page = newPage(d, segment.PageAssociation)
					d.Pages[segment.PageAssociation] = page
					if d.NumberOfPagesUnknown {
						d.NumberOfPages++
					}
				}
				page.Segments = append(page.Segments, segment)
			} else {
				d.GlobalSegments.AddSegment(segment)
			}
		}

		segmentHeaders = append(segmentHeaders, segment)
		offset = d.InputStream.StreamPosition()

		if d.OrganizationType == segments.OSequential {
			offset += int64(segment.SegmentDataLength)
		}

		reachedEOF, err = d.reachedEOF(offset)
		if err != nil {
			common.Log.Debug("jbig2 document reached EOF with error: %v", err)
			return errors.Wrap(err, processName, "")
		}
	}
	d.determineRandomDataOffsets(segmentHeaders, uint64(offset))
	return nil
}

func (d *Document) nextPageNumber() uint32 {
	d.NumberOfPages++
	return d.NumberOfPages
}

func (d *Document) nextSegmentNumber() uint32 {
	s := d.CurrentSegmentNumber
	d.CurrentSegmentNumber++
	return s
}

func (d *Document) parseFileHeader() error {
	const processName = "parseFileHeader"
	// D.4.1 ID string read will be skipped.
	_, err := d.InputStream.Seek(8, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, processName, "id")
	}

	// D.4.2 Header flag (1 byte)
	// Bit 3-7 are reserved and must be 0
	_, err = d.InputStream.ReadBits(5)
	if err != nil {
		return errors.Wrap(err, processName, "reserved bits")
	}

	// Bit 2 - extended templates are used
	b, err := d.InputStream.ReadBit()
	if err != nil {
		return errors.Wrap(err, processName, "extended templates")
	}
	if b == 1 {
		d.GBUseExtTemplate = true
	}

	// Bit 1 - Indicates if amount of pages are unknown.
	b, err = d.InputStream.ReadBit()
	if err != nil {
		return errors.Wrap(err, processName, "unknown page number")
	}
	if b != 1 {
		d.NumberOfPagesUnknown = false
	}

	// Bit 0 - Indicates file organization type.
	b, err = d.InputStream.ReadBit()
	if err != nil {
		return errors.Wrap(err, processName, "organization type")
	}
	d.OrganizationType = segments.OrganizationType(b)

	// D.4.3 Number of pages
	if !d.NumberOfPagesUnknown {
		d.NumberOfPages, err = d.InputStream.ReadUint32()
		if err != nil {
			return errors.Wrap(err, processName, "number of pages")
		}
		d.fileHeaderLength = 13
	}
	return nil
}

func (d *Document) reachedEOF(offset int64) (bool, error) {
	const processName = "reachedEOF"
	_, err := d.InputStream.Seek(offset, io.SeekStart)
	if err != nil {
		common.Log.Debug("reachedEOF - d.InputStream.Seek failed: %v", err)
		return false, errors.Wrap(err, processName, "input stream seek failed")
	}

	_, err = d.InputStream.ReadBits(32)
	if err == io.EOF {
		return true, nil
	} else if err != nil {
		return false, errors.Wrap(err, processName, "")
	}
	return false, nil
}

func decodeWithGlobals(input reader.StreamReader, globals *Globals) (*Document, error) {
	d := &Document{
		Pages:                make(map[int]*Page),
		InputStream:          input,
		OrganizationType:     segments.OSequential,
		NumberOfPagesUnknown: true,
		GlobalSegments:       globals,
		fileHeaderLength:     9,
	}

	if d.GlobalSegments == nil {
		d.GlobalSegments = &Globals{}
	}

	// mapData map the data stream
	if err := d.mapData(); err != nil {
		return nil, err
	}
	return d, nil
}
