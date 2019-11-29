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
	// NumberOfPagesUnknown defines if the ammount of the pages is known.
	NumberOfPagesUnknown bool
	// NumberOfPages - D.4.3 - Number of pages field (4 bytes). Only presented if NumberOfPagesUnknown is true.
	NumberOfPages uint32
	// GBUseExtTemplate defines wether extended Template is used.
	GBUseExtTemplate bool
	// SubInputStream is the source data stream wrapped into a SubInputStream.
	InputStream reader.StreamReader
	// GlobalSegments contains all segments that aren't associated with a page.
	GlobalSegments Globals
	// OrganisationType is the document segment organization.
	OrganizationType segments.OrganizationType

	// Encoder variables
	// Classer is the document encoding classifier.
	Classer classer.Classer
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
}

// DecodeDocument decodes provided document based on the provided 'input' data stream
// and with optional Global defined segments 'globals'.
func DecodeDocument(input reader.StreamReader, globals ...Globals) (*Document, error) {
	var globalsMap Globals
	if len(globals) == 1 {
		globalsMap = globals[0]
	}
	return decodeWithGlobals(input, globalsMap)
}

// InitEncodeDocument initializes the jbig2 document for the encoding process.
func InitEncodeDocument(fullheaders bool) *Document {
	return &Document{FullHeaders: fullheaders, w: writer.BufferedMSB(), Pages: map[int]*Page{}}
}

// AddGenericPage creates the jbig2 page based on the provided bitmap. The data provided
func (d *Document) AddGenericPage(bm *bitmap.Bitmap, duplicateLineRemoval bool) (err error) {
	const processName = "Document.AddGenericPage"
	// check if this is PDFMode and there is already a page
	if !d.FullHeaders && d.NumberOfPages != 0 {
		return errors.Error(processName, "document already contains page. FileMode disallows addoing more than one page")
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
	d.NumberOfPages++
	page.PageNumber = int(d.NumberOfPages)
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
		last uint32
		ok   bool
		h    *segments.Header
		page *Page
	)
	currentPageNumber := 1
	// iterate over segment numbers
	for i := 0; i < int(d.CurrentSegmentNumber); i++ {
		// check if the page is defined
		if page == nil {
			// find page in the document pages
			if page, ok = d.Pages[currentPageNumber]; !ok {
				return nil, errors.Errorf(processName, "page: '#%d' not found", currentPageNumber)
			}
		}
		// check if the last segment number in the page is known
		if last == 0 {
			if last, err = page.lastSegmentNumber(); err != nil {
				return nil, errors.Wrapf(err, processName, "")
			}
		}

		// check if current page number is within given page range
		if last < uint32(i) {
			common.Log.Trace("currentPageNumber: %d, last: %d", currentPageNumber, last)
			currentPageNumber++
			last = 0
			// find page in the document pages
			if page, ok = d.Pages[currentPageNumber]; !ok {
				return nil, errors.Errorf(processName, "page: '#%d' not found", currentPageNumber)
			}
		}

		// check if the segment is in the given page
		h, err = page.GetSegment(i)
		if err != nil {
			// otherwise check if globals contains this segment
			h, ok = d.GlobalSegments[i]
			if !ok {
				return nil, errors.Wrapf(err, processName, "segment '#%d' not found", i)
			}
		}

		if temp, err = h.Encode(d.w); err != nil {
			return nil, errors.Wrapf(err, processName, "segment #%d", i)
		}
		n += temp
		common.Log.Trace("Encoded segment: %s", h.String())
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
			d.mapData()
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
	//	0th bit - file organisation type - '1' for sequential, '0' for random-access
	// 	1st bit - unknown number of pages - '1' if true
	// this encoder stores only sequential organisation type with known number of pages
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
				d.GlobalSegments.AddSegment(int(segment.SegmentNumber), segment)
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
	// Bit 3-7 are reserverd and must be 0
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

	// Bit 0 - Indicates file organisation type.
	b, err = d.InputStream.ReadBit()
	if err != nil {
		return errors.Wrap(err, processName, "organisation type")
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

// func (d *Document) uniteTemplatesWithIndexes(firstTempIndex, secondTempIndex int) error {
// 	if len(d.Classer.Pixat) < firstTempIndex || len(d.Classer.Pixat) < secondTempIndex {
// 		return errors.New("index doesn't point to template array")
// 	}
// 	for i, class := range d.Classer.ClassIDs {
// 		if class == secondTempIndex {
// 			d.Classer.ClassIDs[i] = firstTempIndex
// 		}
// 	}
//
// 	var (
// 		endPix *bitmap.Bitmap
// 		//		copiedPix *bitmap.Bitmap
// 		//		boxa      []*image.Rectangle
// 	)
//
// 	index := len(d.Classer.Pixat) - 1
// 	if index != secondTempIndex {
// 		endPix = d.Classer.Pixat[index]
// 		d.Classer.Pixat[secondTempIndex] = endPix
//
// 		for i, class := range d.Classer.ClassIDs {
// 			if class == index {
// 				d.Classer.ClassIDs[i] = secondTempIndex
// 			}
// 		}
// 	}
//
// 	// remove the bitmap
// 	d.Classer.Pixat = append(d.Classer.Pixat[:index], d.Classer.Pixat[index+1:]...)
// 	d.Classer.NClass--
// 	return nil
// }

func decodeWithGlobals(input reader.StreamReader, globals Globals) (*Document, error) {
	d := &Document{
		Pages:                make(map[int]*Page),
		InputStream:          input,
		OrganizationType:     segments.OSequential,
		NumberOfPagesUnknown: true,
		GlobalSegments:       globals,
		fileHeaderLength:     9,
	}

	if d.GlobalSegments == nil {
		d.GlobalSegments = Globals(make(map[int]*segments.Header))
	}

	// mapData map the data stream
	if err := d.mapData(); err != nil {
		return nil, err
	}
	return d, nil
}
