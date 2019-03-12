package jbig2

import (
	"container/list"
	"errors"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"io"
)

var (
	/** ID string in file header, see ISO/IEC 14492:2001, D.4.1 */
	fileHeaderID []byte = []byte{0x97, 0x4A, 0x42, 0x32, 0x0D, 0x0A, 0x1A, 0x0A}
)

// According to D.4.2. - File header bit 0
const (
	ORandom     uint8 = 0
	OSequential uint8 = 1
)

// Document is the structure of jbig2 document with it's pages  and global segments
type Document struct {
	// Pages contains all pages of this document
	Pages map[int]*Page

	// AmountOfPagesUnknown defines if the ammount of the pages is knownw
	AmountOfPagesUnknown bool

	// AmountOfPages - D.4.3 - Number of pages field (4 bytes). Only presented if
	// AmountOfPagesUnknown is true
	AmountOfPages uint32

	// GBUseExtTemplate defines wether extended Template is used
	GBUseExtTemplate bool

	// SubInputStream is the source data stream wrapped into a SubInputStream
	InputStream *reader.Reader

	// GlobalSegments contains all segments that aren't associated with a page
	GlobalSegments Globals

	OrgainsationType uint8

	fileHeaderLength uint8
}

// NewDocument creates new jbig2.Document for the provided reader
func NewDocument(data []byte) (*Document, error) {
	return NewDocumentWithGlobals(data, nil)
}

// NewDocumentWithGlobals creates new jbig2.Document
func NewDocumentWithGlobals(data []byte, globals Globals) (*Document, error) {
	d := &Document{
		Pages:                make(map[int]*Page),
		InputStream:          reader.New(data),
		OrgainsationType:     OSequential,
		AmountOfPagesUnknown: true,
		GlobalSegments:       globals,
		fileHeaderLength:     9,
	}

	if d.GlobalSegments == nil {
		d.GlobalSegments = Globals(make(map[int]*SegmentHeader))
	}

	// mapData map the data stream
	if err := d.mapData(); err != nil {
		return nil, err
	}

	return d, nil
}

// MapData maps the data and stores all segments
func (d *Document) mapData() error {
	// Get the header list
	segments := list.New()

	var (
		offset      uint64
		segmentType int
	)

	isFileHeaderPresent, err := d.isFileHeaderPresent()
	if err != nil {
		return err
	}

	// Parse the file header if there is one
	if isFileHeaderPresent {
		if err = d.parseFileHeader(); err != nil {
			return err
		}

		offset += uint64(d.fileHeaderLength)
	}

	// var (
	// 	page      *Page
	// 	segmentNo int
	// )

	for segmentType != 51 {
		// TODO: Finish mapping the data
		if segments.Front() == nil {

		}
	}

	return nil

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

func (d *Document) determineRandomDataOffsets(segments *list.List, offset *uint64) {

	if d.OrgainsationType == ORandom {

		for s := segments.Front(); s != nil; s.Next() {
			// h := s.Value.(*Header)
			// TODO: h.setSegmentDataStartOffset(offset)

			// TODO: offset += h.getSegmentDataLength
		}

	}
}

// parseFileHeader - this method reads the stream and sets variables for information about
// organization type and length etc.
func (d *Document) parseFileHeader() error {
	// D.4.1 ID string read will be skipped
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

	// Bit 1 - Indicates if amount of pages are unknown
	b, err = d.InputStream.ReadBit()
	if err != nil {
		return err
	}
	if b != 1 {
		d.AmountOfPagesUnknown = false
	}

	// Bit 0 - Indicates file organisation type
	b, err = d.InputStream.ReadBit()
	if err != nil {
		return err
	}
	d.OrgainsationType = uint8(b)

	// D.4.3 Number of pages
	if !d.AmountOfPagesUnknown {
		d.AmountOfPages, err = d.InputStream.ReadUnsignedInt()
		if err != nil {
			return err
		}
		d.fileHeaderLength = 13
	}

	return nil
}

// GetPage gets the gage for the provided 'pageNumber'
func (d *Document) GetPage(pageNumber int) (*Page, error) {
	p, ok := d.Pages[pageNumber]
	if !ok {
		return nil, errors.New("No page found")
	}

	return p, nil
}

func (d *Document) GetAmountOfPages() (uint32, error) {
	if d.AmountOfPagesUnknown || d.AmountOfPages == 0 {
		if len(d.Pages) == 0 {
			d.mapData()
		}

		return uint32(len(d.Pages)), nil
	}
	return d.AmountOfPages, nil
}

// func (d *Document) reachedEOF(offset uint64) (bool, error) {
// 	_, err := d.InputStream.Seek(offset, io.SeekStart)

// }
