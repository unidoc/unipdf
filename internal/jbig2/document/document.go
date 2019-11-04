/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package document

import (
	"errors"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document/segments"
	"github.com/unidoc/unipdf/v3/internal/jbig2/encoder/classer"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
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
	CurrentSegmentNumber int
	// CurrentPageNumber stores the number of given page.
	CurrentPageNumber int

	// SymbolTableSegment is the segment number of the symbol table.
	SymbolTableSegment int

	// AverageTemplates are the grayed templates.
	AverageTemplates *bitmap.Bitmaps
	BaseIndexes      []int

	Refinement  bool
	RefineLevel int

	fileHeaderLength uint8
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
func InitEncodeDocument(class classer.Classer, xres, yres int, fullheaders bool, refineLevel int) *Document {
	return &Document{
		Classer:            class,
		XRes:               xres,
		YRes:               yres,
		SymbolTableSegment: -1,
		Refinement:         refineLevel >= 0,
		RefineLevel:        refineLevel,
		FullHeaders:        fullheaders,
	}
}

// AddGenericPage creates the jbig2 page based on the provided bitmap. The data provided
func (d *Document) AddGenericPage(bm *bitmap.Bitmap, duplicateLineRemoval bool) (err error) {
	page := &Page{Segments: []*segments.Header{}, Bitmap: bm}
	page.PageNumber = int(d.NumberOfPages)
	d.NumberOfPages++

	pageInfo := &segments.PageInformationSegment{
		PageBMWidth:  bm.Width,
		PageBMHeight: bm.Height,
		ResolutionX:  bm.XResolution,
		ResolutionY:  bm.YResolution,
		IsLossless:   true,
	}
	// add page info segment
	h := &segments.Header{
		Type:            segments.TPageInformation,
		SegmentNumber:   uint32(d.CurrentSegmentNumber),
		PageAssociation: page.PageNumber,
		SegmentData:     pageInfo,
		// add segment data length of the page info size.
	}

	//  add page info
	page.Segments = append(page.Segments, h)
	// create a generic segment header
	generic := &segments.GenericRegion{
		RegionSegment: &segments.RegionSegment{
			BitmapWidth:  uint32(bm.Width),
			BitmapHeight: uint32(bm.Height),
		},
		Bitmap: bm,
	}
	if duplicateLineRemoval {
		generic.IsTPGDon = true
	}
	generic.GBAtX = []int8{3, -3, 2, -2}
	generic.GBAtY = []int8{-1, -1, -2, -2}

	// add generic segment header
	h = &segments.Header{
		Type:            segments.TImmediateGenericRegion,
		SegmentData:     generic,
		PageAssociation: page.PageNumber,
		SegmentNumber:   uint32(d.CurrentSegmentNumber),
	}
	d.CurrentSegmentNumber++
	// set the header to the page.
	page.Segments = append(page.Segments, h)

	// page end segment
	h = &segments.Header{
		Type:            segments.TEndOfPage,
		SegmentNumber:   uint32(d.CurrentSegmentNumber),
		PageAssociation: page.PageNumber,
	}
	page.Segments = append(page.Segments, h)
	d.CurrentSegmentNumber++
	return nil
}

// AutoThreshold gathers classes of symbols and uses a single representative to stand for them all.
func (d *Document) AutoThreshold() error {
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
	return nil
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
	if pageNumber < 0 {
		common.Log.Debug("JBIG2 Page - GetPage: %d. Page cannot be lower than 0. %s", pageNumber, debug.Stack())
		return nil, fmt.Errorf("invalid jbig2 document - provided invalid page number: %d", pageNumber)
	}

	if pageNumber > len(d.Pages) {
		common.Log.Debug("Page not found: %d. %s", pageNumber, debug.Stack())
		return nil, errors.New("invalid jbig2 document - page not found")
	}

	p, ok := d.Pages[pageNumber]
	if !ok {
		common.Log.Debug("Page not found: %d. %s", pageNumber, debug.Stack())
		return nil, errors.New("invalid jbig2 document - page not found")
	}

	return p, nil
}

// GetGlobalSegment implements segments.Documenter interface.
func (d *Document) GetGlobalSegment(i int) (*segments.Header, error) {
	return d.GlobalSegments.GetSegment(i)
}

func (d *Document) determineRandomDataOffsets(segmentHeaders []*segments.Header, offset uint64) {
	if d.OrganizationType != segments.ORandom {
		return
	}

	for _, s := range segmentHeaders {
		s.SegmentDataStartOffset = offset
		offset += s.SegmentDataLength
	}
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
	// Get the header list
	var (
		segmentHeaders []*segments.Header
		offset         int64
		kind           segments.Type
	)

	isFileHeaderPresent, err := d.isFileHeaderPresent()
	if err != nil {
		return err
	}

	// Parse the file header if exists.
	if isFileHeaderPresent {
		if err = d.parseFileHeader(); err != nil {
			return err
		}
		offset += int64(d.fileHeaderLength)
	}

	var (
		page       *Page
		segmentNo  int
		reachedEOF bool
	)

	// type 51 is the EndOfFile segment kind
	for kind != 51 && !reachedEOF {
		segmentNo++

		// get new segment
		segment, err := segments.NewHeader(d, d.InputStream, offset, d.OrganizationType)
		if err != nil {
			return err
		}

		common.Log.Trace("Decoding segment number: %d, Type: %s", segmentNo, segment.Type)

		kind = segment.Type
		if kind != segments.TEndOfFile {
			if segment.PageAssociation != 0 {
				page = d.Pages[segment.PageAssociation]

				if page == nil {
					page = newPage(d, segment.PageAssociation)
					d.Pages[segment.PageAssociation] = page
				}

				page.Segments[int(segment.SegmentNumber)] = segment
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
			return err
		}
	}
	d.determineRandomDataOffsets(segmentHeaders, uint64(offset))
	return nil
}

func (d *Document) parseFileHeader() error {
	// D.4.1 ID string read will be skipped.
	_, err := d.InputStream.Seek(8, io.SeekStart)
	if err != nil {
		return err
	}

	// D.4.2 Header flag (1 byte)
	// Bit 3-7 are reserverd and must be 0
	_, err = d.InputStream.ReadBits(5)
	if err != nil {
		return err
	}

	// Bit 2 - extended templates are used
	b, err := d.InputStream.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		d.GBUseExtTemplate = true
	}

	// Bit 1 - Indicates if amount of pages are unknown.
	b, err = d.InputStream.ReadBit()
	if err != nil {
		return err
	}
	if b != 1 {
		d.NumberOfPagesUnknown = false
	}

	// Bit 0 - Indicates file organisation type.
	b, err = d.InputStream.ReadBit()
	if err != nil {
		return err
	}
	d.OrganizationType = segments.OrganizationType(b)

	// D.4.3 Number of pages
	if !d.NumberOfPagesUnknown {
		d.NumberOfPages, err = d.InputStream.ReadUint32()
		if err != nil {
			return err
		}
		d.fileHeaderLength = 13
	}
	return nil
}

func (d *Document) reachedEOF(offset int64) (bool, error) {
	_, err := d.InputStream.Seek(offset, io.SeekStart)
	if err != nil {
		common.Log.Debug("reachedEOF - d.InputStream.Seek failed: %v", err)
		return false, err
	}

	_, err = d.InputStream.ReadBits(32)
	if err == io.EOF {
		return true, nil
	} else if err != nil {
		return false, err
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
