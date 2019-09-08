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

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
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
	RTSNumbers               []int
}

// NewHeader creates new segment header for the provided document from the stream reader.
func NewHeader(d Documenter, r reader.StreamReader, offset int64, organizationType OrganizationType) (*Header, error) {
	h := &Header{Reader: r}
	if err := h.parse(d, r, offset, organizationType); err != nil {
		return nil, err
	}
	return h, nil
}

// CleanSegmentData cleans the segment's data setting it's segment data to nil.
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

// parses the current segment header for the provided document 'd'.
func (h *Header) parse(
	d Documenter, r reader.StreamReader,
	offset int64, organizationType OrganizationType,
) (err error) {
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
		return err
	}

	// 7.2.2 Segment Number.
	if err = h.readSegmentNumber(r); err != nil {
		return err
	}

	// 7.2.3 Segment header flags.
	if err = h.readHeaderFlags(r); err != nil {
		return err
	}

	// 7.2.4 Amount of referred-to segment.
	var countOfRTS uint64
	countOfRTS, err = h.readNumberOfReferredToSegments(r)
	if err != nil {
		return err
	}

	// 7.2.5 Refered-tp segment numbers.
	h.RTSNumbers, err = h.readReferedToSegmentNumbers(r, int(countOfRTS))
	if err != nil {
		return err
	}

	// 7.2.6 Segment page association.
	err = h.readSegmentPageAssociation(d, r, countOfRTS, h.RTSNumbers...)
	if err != nil {
		return err
	}

	if h.Type != TEndOfFile {
		// 7.2.7 Segment data length (Contains the length of the data).
		if err = h.readSegmentDataLength(r); err != nil {
			return err
		}
	}

	h.readDataStartOffset(r, organizationType)
	h.readHeaderLength(r, offset)

	common.Log.Trace("%s", h)
	return nil
}

// readSegmentNumber reads the segment number.
func (h *Header) readSegmentNumber(r reader.StreamReader) error {
	// 7.2.2
	b := make([]byte, 4)
	_, err := r.Read(b)
	if err != nil {
		return err
	}

	// BigEndian number
	h.SegmentNumber = binary.BigEndian.Uint32(b)
	return nil
}

// readHeaderFlags reads the header flag values.
func (h *Header) readHeaderFlags(r reader.StreamReader) error {
	// 7.2.3
	bit, err := h.Reader.ReadBit()
	if err != nil {
		return err
	}

	// Bit 7: Retain Flag
	if bit != 0 {
		h.RetainFlag = true
	}

	bit, err = h.Reader.ReadBit()
	if err != nil {
		return err
	}

	// Bit 6: Size of the page association field
	if bit != 0 {
		h.PageAssociationFieldSize = true
	}

	// Bit 5-0 Contains the values (between 0 - 62 with gaps) for segment types
	tp, err := h.Reader.ReadBits(6)
	if err != nil {
		return err
	}
	h.Type = Type(int(tp))

	return nil
}

// readNumberOfReferredToSegments gets the amount of referred-to segments.
func (h *Header) readNumberOfReferredToSegments(r reader.StreamReader) (uint64, error) {
	// 7.2.4
	countOfRTS, err := r.ReadBits(3)
	if err != nil {
		return 0, err
	}
	countOfRTS &= 0xf
	var retainBit []byte

	if countOfRTS <= 4 {
		// short format
		retainBit = make([]byte, 5)
		for i := 0; i <= 4; i++ {
			b, err := r.ReadBit()
			if err != nil {
				return 0, err
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
				return 0, err
			}
			retainBit[i] = byte(b)
		}
	}
	return countOfRTS, nil
}

// readReferedToSegmentNumbers gathers all segment numbers of referred-to segments. The
// segment itself is in rtSegments the array.
func (h *Header) readReferedToSegmentNumbers(r reader.StreamReader, countOfRTS int) ([]int, error) {
	// 7.2.5
	rtsNumbers := make([]int, countOfRTS)

	if countOfRTS > 0 {
		rtsSize := byte(1)

		if h.SegmentNumber > 256 {
			rtsSize = 2

			if h.SegmentNumber > 65536 {
				rtsSize = 4
			}
		}

		h.RTSegments = make([]*Header, countOfRTS)
		var (
			bits uint64
			err  error
		)
		for i := 0; i < countOfRTS; i++ {
			bits, err = r.ReadBits(rtsSize << 3)
			if err != nil {
				return nil, err
			}
			rtsNumbers[i] = int(bits & math.MaxInt32)
		}
	}
	return rtsNumbers, nil
}

// readSegmentPageAssociation gets the segment's associated page number.
func (h *Header) readSegmentPageAssociation(d Documenter, r reader.StreamReader, countOfRTS uint64, rtsNumbers ...int) error {
	// 7.2.6
	if !h.PageAssociationFieldSize {
		// Short format
		bits, err := r.ReadBits(8)
		if err != nil {
			return err
		}

		h.PageAssociation = int(bits & 0xFF)
	} else {
		// Long format
		bits, err := r.ReadBits(32)
		if err != nil {
			return err
		}

		h.PageAssociation = int(bits & math.MaxInt32)
	}

	if countOfRTS > 0 {
		page, _ := d.GetPage(h.PageAssociation)
		var i uint64

		for i = 0; i < countOfRTS; i++ {
			if page != nil {
				h.RTSegments[i] = page.GetSegment(rtsNumbers[i])
			} else {
				h.RTSegments[i] = d.GetGlobalSegment(rtsNumbers[i])
			}
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
// the organisation type is OSequential.
func (h *Header) readDataStartOffset(r reader.StreamReader, organizationType OrganizationType) {
	if organizationType == OSequential {
		h.SegmentDataStartOffset = uint64(r.StreamPosition())
	}
}

func (h *Header) readHeaderLength(r reader.StreamReader, offset int64) {
	h.HeaderLength = r.StreamPosition() - offset
}

func (h *Header) subInputReader() (reader.StreamReader, error) {
	return reader.NewSubstreamReader(h.Reader, h.SegmentDataStartOffset, h.SegmentDataLength)
}
