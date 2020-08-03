/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// Header is the segment header used to define the segment parameters - see 7.2.
type Header struct {
	SegmentNumber            uint32
	Type                     Type
	RetainFlag               bool
	PageAssociation          int
	PageAssociationFieldSize bool
	RTSegments               []*Header
	HeaderLength             int64
	SegmentDataLength        uint64
	SegmentDataStartOffset   uint64
	Reader                   reader.StreamReader
	SegmentData              Segmenter
	// RTSNumbers is the list of numbers where the segment is referred to.
	RTSNumbers []int
	// RetainBits are the  flags for the given segment.
	RetainBits []uint8
}

// NewHeader creates new segment header for the provided document from the stream reader.
func NewHeader(d Documenter, r reader.StreamReader, offset int64, organizationType OrganizationType) (*Header, error) {
	h := &Header{Reader: r}
	if err := h.parse(d, r, offset, organizationType); err != nil {
		return nil, errors.Wrap(err, "NewHeader", "")
	}
	return h, nil
}

// CleanSegmentData cleans the segment's data setting its segment data to nil.
func (h *Header) CleanSegmentData() {
	if h.SegmentData != nil {
		h.SegmentData = nil
	}
}

// GetSegmentData gets the segment's data returning the Segmenter instance.
func (h *Header) GetSegmentData() (Segmenter, error) {
	var segmentDataPart Segmenter

	if h.SegmentData != nil {
		segmentDataPart = h.SegmentData
	}

	if segmentDataPart == nil {
		creator, ok := kindMap[h.Type]
		if !ok {
			return nil, fmt.Errorf("type: %s/ %d creator not found. ", h.Type, h.Type)
		}
		segmentDataPart = creator()

		common.Log.Trace("[SEGMENT-HEADER][#%d] GetSegmentData at Offset: %04X", h.SegmentNumber, h.SegmentDataStartOffset)
		subReader, err := h.subInputReader()
		if err != nil {
			return nil, err
		}

		if err := segmentDataPart.Init(h, subReader); err != nil {
			common.Log.Debug("Init failed: %v for type: %T", err, segmentDataPart)
			return nil, err
		}
		h.SegmentData = segmentDataPart
	}
	return segmentDataPart, nil
}

// Encode encodes the jbi2 header structure to the provided 'w' BinaryWriter.
func (h *Header) Encode(w writer.BinaryWriter) (n int, err error) {
	const processName = "Header.Write"
	var tw writer.BinaryWriter
	common.Log.Trace("[SEGMENT-HEADER][ENCODE] Begins")
	defer func() {
		if err != nil {
			common.Log.Trace("[SEGMENT-HEADER][ENCODE] Failed. %v", err)
		} else {
			common.Log.Trace("[SEGMENT-HEADER] %v", h)
			common.Log.Trace("[SEGMENT-HEADER][ENCODE] Finished")
		}
	}()
	// For safety Finish the current byte in the 'w'.
	w.FinishByte()
	// check if the header contains any data
	if h.SegmentData != nil {
		se, ok := h.SegmentData.(SegmentEncoder)
		if !ok {
			return 0, errors.Errorf(processName, "Segment: %T doesn't implement SegmentEncoder interface", h.SegmentData)
		}

		// create temporary writer for the segment data
		tw = writer.BufferedMSB()

		// encode the segment data
		n, err = se.Encode(tw)
		if err != nil {
			return 0, errors.Wrap(err, processName, "")
		}
		h.SegmentDataLength = uint64(n)
	}

	// the header contains following bitwise parts:
	// [SegmentNumber] - 4 byte uint32
	// [segment header flags] - 1 byte
	// [referred-to count and retention flags] 1 or more bytes
	// [referred-to segment numbers] 1 or more bytes
	// [segment page association] - 1 or more bytes
	// [segment data length] - 4 bytes - uint32
	if h.pageSize() == 4 {
		h.PageAssociationFieldSize = true
	}

	var temp int
	// 7.2.2 Segment number
	temp, err = h.writeSegmentNumber(w)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	n += temp

	// 7.2.3 Segment header flags
	if err = h.writeFlags(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n++

	// 7.2.4 Referred-to segment count and retention flags
	temp, err = h.writeReferredToCount(w)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	n += temp

	// 7.2.5 Referred-to segment numbers
	temp, err = h.writeReferredToSegments(w)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	n += temp

	// 7.2.6 Segment page association
	temp, err = h.writeSegmentPageAssociation(w)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	n += temp

	// 7.2.7 Segment data length
	temp, err = h.writeSegmentDataLength(w)
	if err != nil {
		return 0, errors.Wrap(err, processName, "")
	}
	n += temp

	// compute header length
	h.HeaderLength = int64(n) - int64(h.SegmentDataLength)

	// if the segment contains any data writer it into the 'w' writer
	if tw != nil {
		if _, err = w.Write(tw.Data()); err != nil {
			return n, errors.Wrap(err, processName, "write segment data")
		}
	}
	return n, nil
}

// String implements Stringer interface.
func (h *Header) String() string {
	sb := &strings.Builder{}
	sb.WriteString("\n[SEGMENT-HEADER]\n")
	sb.WriteString(fmt.Sprintf("\t- SegmentNumber: %v\n", h.SegmentNumber))
	sb.WriteString(fmt.Sprintf("\t- Type: %v\n", h.Type))
	sb.WriteString(fmt.Sprintf("\t- RetainFlag: %v\n", h.RetainFlag))
	sb.WriteString(fmt.Sprintf("\t- PageAssociation: %v\n", h.PageAssociation))
	sb.WriteString(fmt.Sprintf("\t- PageAssociationFieldSize: %v\n", h.PageAssociationFieldSize))
	sb.WriteString("\t- RTSEGMENTS:\n")
	for _, rt := range h.RTSNumbers {
		sb.WriteString(fmt.Sprintf("\t\t- %d\n", rt))
	}
	sb.WriteString(fmt.Sprintf("\t- HeaderLength: %v\n", h.HeaderLength))
	sb.WriteString(fmt.Sprintf("\t- SegmentDataLength: %v\n", h.SegmentDataLength))
	sb.WriteString(fmt.Sprintf("\t- SegmentDataStartOffset: %v\n", h.SegmentDataStartOffset))

	return sb.String()
}

/**

Header private methods

*/

// pageSize returns the size of the segment page association field.:w
func (h *Header) pageSize() uint {
	if h.PageAssociation <= 255 {
		return 1
	}
	return 4
}

// parses the current segment header for the provided document 'd'.
func (h *Header) parse(
	d Documenter, r reader.StreamReader,
	offset int64, organizationType OrganizationType,
) (err error) {
	const processName = "parse"
	common.Log.Trace("[SEGMENT-HEADER][PARSE] Begins")
	defer func() {
		if err != nil {
			common.Log.Trace("[SEGMENT-HEADER][PARSE] Failed. %v", err)
		} else {
			common.Log.Trace("[SEGMENT-HEADER][PARSE] Finished")
		}
	}()

	_, err = r.Seek(offset, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, processName, "seek start")
	}

	// 7.2.2 Segment Number.
	if err = h.readSegmentNumber(r); err != nil {
		return errors.Wrap(err, processName, "")
	}

	// 7.2.3 Segment header flags.
	if err = h.readHeaderFlags(); err != nil {
		return errors.Wrap(err, processName, "")
	}

	// 7.2.4 Amount of referred-to segment.
	var countOfRTS uint64
	countOfRTS, err = h.readNumberOfReferredToSegments(r)
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	// 7.2.5 referred-tp segment numbers.
	h.RTSNumbers, err = h.readReferredToSegmentNumbers(r, int(countOfRTS))
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	// 7.2.6 Segment page association.
	err = h.readSegmentPageAssociation(d, r, countOfRTS, h.RTSNumbers...)
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	if h.Type != TEndOfFile {
		// 7.2.7 Segment data length (Contains the length of the data).
		if err = h.readSegmentDataLength(r); err != nil {
			return errors.Wrap(err, processName, "")
		}
	}
	h.readDataStartOffset(r, organizationType)
	h.readHeaderLength(r, offset)

	common.Log.Trace("%s", h)
	return nil
}

// readSegmentNumber reads the segment number.
func (h *Header) readSegmentNumber(r reader.StreamReader) error {
	const processName = "readSegmentNumber"
	// 7.2.2
	b := make([]byte, 4)
	_, err := r.Read(b)
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	// BigEndian number
	h.SegmentNumber = binary.BigEndian.Uint32(b)
	return nil
}

// readHeaderFlags reads the header flag values.
func (h *Header) readHeaderFlags() error {
	const processName = "readHeaderFlags"
	// 7.2.3
	bit, err := h.Reader.ReadBit()
	if err != nil {
		return errors.Wrap(err, processName, "retain flag")
	}

	// Bit 7: Retain Flag
	if bit != 0 {
		h.RetainFlag = true
	}

	bit, err = h.Reader.ReadBit()
	if err != nil {
		return errors.Wrap(err, processName, "page association")
	}

	// Bit 6: Size of the page association field
	if bit != 0 {
		h.PageAssociationFieldSize = true
	}

	// Bit 5-0 Contains the values (between 0 - 62 with gaps) for segment types
	tp, err := h.Reader.ReadBits(6)
	if err != nil {
		return errors.Wrap(err, processName, "segment type")
	}
	h.Type = Type(int(tp))

	return nil
}

// readNumberOfReferredToSegments gets the amount of referred-to segments.
func (h *Header) readNumberOfReferredToSegments(r reader.StreamReader) (uint64, error) {
	const processName = "readNumberOfReferredToSegments"
	// 7.2.4
	countOfRTS, err := r.ReadBits(3)
	if err != nil {
		return 0, errors.Wrap(err, processName, "count of rts")
	}
	countOfRTS &= 0xf
	var retainBit []byte

	if countOfRTS <= 4 {
		// short format
		retainBit = make([]byte, 5)
		for i := 0; i <= 4; i++ {
			b, err := r.ReadBit()
			if err != nil {
				return 0, errors.Wrap(err, processName, "short format")
			}
			retainBit[i] = byte(b)
		}
	} else {
		// long format
		countOfRTS, err = r.ReadBits(29)
		if err != nil {
			return 0, err
		}
		countOfRTS &= math.MaxInt32
		arrayLength := (countOfRTS + 8) >> 3
		arrayLength <<= 3
		retainBit = make([]byte, arrayLength)

		var i uint64
		for i = 0; i < arrayLength; i++ {
			b, err := r.ReadBit()
			if err != nil {
				return 0, errors.Wrap(err, processName, "long format")
			}
			retainBit[i] = byte(b)
		}
	}
	return countOfRTS, nil
}

// readReferredToSegmentNumbers gathers all segment numbers of referred-to segments. The
// segment itself is in rtSegments the array.
func (h *Header) readReferredToSegmentNumbers(r reader.StreamReader, countOfRTS int) ([]int, error) {
	const processName = "readReferredToSegmentNumbers"
	// 7.2.5
	rtsNumbers := make([]int, countOfRTS)

	if countOfRTS > 0 {
		h.RTSegments = make([]*Header, countOfRTS)
		var (
			bits uint64
			err  error
		)
		for i := 0; i < countOfRTS; i++ {
			bits, err = r.ReadBits(byte(h.referenceSize()) << 3)
			if err != nil {
				return nil, errors.Wrapf(err, processName, "'%d' referred segment number", i)
			}
			rtsNumbers[i] = int(bits & math.MaxInt32)
		}
	}
	return rtsNumbers, nil
}

// readSegmentPageAssociation gets the segment's associated page number.
func (h *Header) readSegmentPageAssociation(d Documenter, r reader.StreamReader, countOfRTS uint64, rtsNumbers ...int) (err error) {
	const processName = "readSegmentPageAssociation"
	// 7.2.6
	if !h.PageAssociationFieldSize {
		// Short format
		bits, err := r.ReadBits(8)
		if err != nil {
			return errors.Wrap(err, processName, "short format")
		}
		h.PageAssociation = int(bits & 0xFF)
	} else {
		// Long format
		bits, err := r.ReadBits(32)
		if err != nil {
			return errors.Wrap(err, processName, "long format")
		}
		h.PageAssociation = int(bits & math.MaxInt32)
	}

	// check if there are any related to segments
	if countOfRTS == 0 {
		return nil
	}

	// map the related segment headers from the pages.
	// check if the page number is different than 0 - 'Globals' are stored under '0'
	if h.PageAssociation != 0 {
		page, err := d.GetPage(h.PageAssociation)
		if err != nil {
			return errors.Wrap(err, processName, "associated page not found")
		}

		var relatedSegmentNumber int
		for i := uint64(0); i < countOfRTS; i++ {
			relatedSegmentNumber = rtsNumbers[i]
			h.RTSegments[i], err = page.GetSegment(relatedSegmentNumber)
			if err != nil {
				var er error
				h.RTSegments[i], er = d.GetGlobalSegment(relatedSegmentNumber)
				if er != nil {
					return errors.Wrapf(err, processName, "reference segment not found at page: '%d' nor in globals", h.PageAssociation)
				}
			}
		}
		return nil
	}

	for i := uint64(0); i < countOfRTS; i++ {
		h.RTSegments[i], err = d.GetGlobalSegment(rtsNumbers[i])
		if err != nil {
			return errors.Wrapf(err, processName, "global segment: '%d' not found", rtsNumbers[i])
		}
	}
	return nil
}

// readSegmentDataLength contains the length of the data part in bytes.
func (h *Header) readSegmentDataLength(r reader.StreamReader) (err error) {
	// 7.2.7
	h.SegmentDataLength, err = r.ReadBits(32)
	if err != nil {
		return err
	}

	// Set the 4bytes mask
	h.SegmentDataLength &= math.MaxInt32
	return nil
}

// readDataStartOffset sets the offset of the current reader if
// the organization type is OSequential.
func (h *Header) readDataStartOffset(r reader.StreamReader, organizationType OrganizationType) {
	if organizationType == OSequential {
		h.SegmentDataStartOffset = uint64(r.StreamPosition())
	}
}

func (h *Header) readHeaderLength(r reader.StreamReader, offset int64) {
	h.HeaderLength = r.StreamPosition() - offset
}

// referenceSize returns the size of the segment reference for this segment. Segments can only refer to previous
// segments, so the bits needed is determined by the number of this segment.
func (h *Header) referenceSize() uint {
	switch {
	case h.SegmentNumber <= 255:
		return 1
	case h.SegmentNumber <= 65535:
		return 2
	default:
		return 4
	}
}

func (h *Header) subInputReader() (reader.StreamReader, error) {
	return reader.NewSubstreamReader(h.Reader, h.SegmentDataStartOffset, h.SegmentDataLength)
}
func (h *Header) writeFlags(w writer.BinaryWriter) (err error) {
	const processName = "Header.writeFlags"
	// the header flags is composed of 8 bits.
	// MSB 00000000 LSB
	//     ^				the first bit is the 'retain' flag.
	//	    ^				the second defines if the page has association field size flag.
	// The last 6 bits are the segment header type number.
	//
	// at first write the type number.
	// segment type can take value from 0 - 63 - 6-bits
	bt := byte(h.Type)
	if err = w.WriteByte(bt); err != nil {
		return errors.Wrap(err, processName, "writing segment type number failed")
	}

	// check if the header should contain bits that defines other flags
	if !h.RetainFlag && !h.PageAssociationFieldSize {
		return nil
	}

	// skip back the bits to the previous pointer.
	if err = w.SkipBits(-8); err != nil {
		return errors.Wrap(err, processName, "skipping back the bits failed")
	}

	var bit int
	// first bit is the deferred non retain flag.
	if h.RetainFlag {
		bit = 1
	}

	if err = w.WriteBit(bit); err != nil {
		return errors.Wrap(err, processName, "retain retain flags")
	}
	// reset the bit
	bit = 0
	// second bit is the page association field size flag.
	if h.PageAssociationFieldSize {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return errors.Wrap(err, processName, "page association flag")
	}

	// finish the byte so that it contains all the flags and the type number and the byte pointer
	// is moved to the next byte and the bitPointer is reset.
	w.FinishByte()
	return nil
}

func (h *Header) writeReferredToCount(w writer.BinaryWriter) (n int, err error) {
	const processName = "writeReferredToCount"
	// the referred to count segment is one byte long if there are up to maximum of 4
	// referred to other headers.

	// copy the segment numbers from the referred to slice
	h.RTSNumbers = make([]int, len(h.RTSegments))
	for i, seg := range h.RTSegments {
		h.RTSNumbers[i] = int(seg.SegmentNumber)
	}

	// if the referred to count is <= 4 the header uses short format
	if len(h.RTSNumbers) <= 4 {
		// use the short format
		var bt byte
		if len(h.RetainBits) >= 1 {
			bt = byte(h.RetainBits[0])
		}
		// first three bits are the number of referred to segments
		// and the other five bits are the retain bits flags.
		bt |= byte(len(h.RTSNumbers)) << 5
		if err = w.WriteByte(bt); err != nil {
			return 0, errors.Wrap(err, processName, "short format")
		}
		return 1, nil
	}
	// use the long format
	count := uint32(len(h.RTSNumbers))
	bts := make([]byte, 4+basic.Ceil(len(h.RTSNumbers)+1, 8))
	count |= 0x7 << 29
	binary.BigEndian.PutUint32(bts, count)
	// copy retain bits to the bts
	copy(bts[1:], h.RetainBits)

	n, err = w.Write(bts)
	if err != nil {
		return 0, errors.Wrap(err, processName, "long format")
	}
	return n, nil
}

func (h *Header) writeReferredToSegments(w writer.BinaryWriter) (n int, err error) {
	const processName = "writeReferredToSegments"
	var (
		short uint16
		long  uint32
	)
	referenceSize := h.referenceSize()
	// temp bytes written size initialized to '1' for uint8u
	temp := 1
	bts := make([]byte, referenceSize)
	for _, referred := range h.RTSNumbers {
		switch referenceSize {
		case 4:
			// referred to integers are uint32
			long = uint32(referred)
			binary.BigEndian.PutUint32(bts, long)
			temp, err = w.Write(bts)
			if err != nil {
				return 0, errors.Wrap(err, processName, "uint32 size")
			}
		case 2:
			// referred to integers are uint16
			short = uint16(referred)
			binary.BigEndian.PutUint16(bts, short)
			temp, err = w.Write(bts)
			if err != nil {
				return 0, errors.Wrap(err, processName, "uint16")
			}
		default:
			// referred to integers are uint8
			if err = w.WriteByte(byte(referred)); err != nil {
				return 0, errors.Wrap(err, processName, "uint8")
			}
		}
		n += temp
	}
	return n, nil
}

func (h *Header) writeSegmentDataLength(w writer.BinaryWriter) (n int, err error) {
	// uint32 segment number
	temp := make([]byte, 4)
	// the multibytes integers are stored in the big endian manner.
	binary.BigEndian.PutUint32(temp, uint32(h.SegmentDataLength))
	if n, err = w.Write(temp); err != nil {
		return 0, errors.Wrap(err, "Header.writeSegmentDataLength", "")
	}
	return n, nil
}

func (h *Header) writeSegmentNumber(w writer.BinaryWriter) (n int, err error) {
	// uint32 segment number
	temp := make([]byte, 4)
	// the multibytes integers are stored in the big endian manner.
	binary.BigEndian.PutUint32(temp, h.SegmentNumber)
	if n, err = w.Write(temp); err != nil {
		return 0, errors.Wrap(err, "Header.writeSegmentNumber", "")
	}
	return n, nil
}

func (h *Header) writeSegmentPageAssociation(w writer.BinaryWriter) (n int, err error) {
	const processName = "writeSegmentPageAssociation"
	if h.pageSize() != 4 {
		if err = w.WriteByte(byte(h.PageAssociation)); err != nil {
			return 0, errors.Wrap(err, processName, "pageSize != 4")
		}
		return 1, nil
	}
	// the page size is written as 4 bytes - header flag page association field size is set to '1'
	bts := make([]byte, 4)
	binary.BigEndian.PutUint32(bts, uint32(h.PageAssociation))
	if n, err = w.Write(bts); err != nil {
		return 0, errors.Wrap(err, processName, "4 byte page number")
	}
	return n, nil
}
