/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package jbig2

import (
	"errors"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/segments"
)

// fileHeaderID first byte slices of the jbig2 encoded file, see D.4.1.
var fileHeaderID = []byte{0x97, 0x4A, 0x42, 0x32, 0x0D, 0x0A, 0x1A, 0x0A}

// Document is the jbig2 document model containing pages and global segments.
// By creating new document with method NewDocument or NewDocumentWithGlobals
// all the jbig2 encoded data segment headers are decoded.
// In order to decode whole document, all of it's pages should be decoded using GetBitmap method.
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
	InputStream *reader.Reader

	// GlobalSegments contains all segments that aren't associated with a page.
	GlobalSegments Globals

	// OrganisationType is the document segment organization.
	OrganizationType segments.OrganizationType

	fileHeaderLength uint8
}

// NewDocument creates new Document for the 'data' byte slice.
func NewDocument(data []byte) (*Document, error) {
	return NewDocumentWithGlobals(data, nil)
}

// NewDocumentWithGlobals creates new Document for the provided encoded 'data'
// byte slice and the 'globals' Globals.
func NewDocumentWithGlobals(data []byte, globals Globals) (*Document, error) {
	d := &Document{
		Pages:                make(map[int]*Page),
		InputStream:          reader.New(data),
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
func (d *Document) GetGlobalSegment(i int) *segments.Header {
	if d.GlobalSegments == nil {
		common.Log.Debug("Trying to get Global segment from nil Globals")
		return nil
	}
	return d.GlobalSegments[i]
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
		d.NumberOfPages, err = d.InputStream.ReadUnsignedInt()
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
