package jbig2

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"io"
	"strings"
)

var log common.Logger = common.Log

// SegmentHeader is the segment header used to define the segment parameters
type SegmentHeader struct {
	SegmentNumber            uint32
	SegmentType              SegmentType
	RetainFlag               bool
	PageAssociation          int
	PageAssociationFieldSize bool
	RTSegments               []*SegmentHeader
	SegmentHeaderLength      int64
	SegmentDataLength        uint64
	SegmentDataStartOffset   uint64
	Reader                   reader.StreamReader
	SegmentData              Segmenter
}

// NewHeader creates new segment header
func NewHeader(
	d *Document, r reader.StreamReader,
	offset int64, organisationType uint8,
) (*SegmentHeader, error) {
	h := &SegmentHeader{Reader: r}
	if err := h.parse(d, r, offset, organisationType); err != nil {
		return nil, err
	}

	return h, nil
}

// Parse parses the current segment header for the provided document 'd'.
func (h *SegmentHeader) parse(
	d *Document, r reader.StreamReader,
	offset int64, organisationType uint8,
) error {
	common.Log.Debug("[SEGMENT-HEADER][PARSE] Begins")
	defer func() { common.Log.Debug("[SEGMENT-HEADER][PARSE] Finished") }()

	_, err := r.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	// 7.2.2 Segment Number
	if err = h.readSegmentNumber(r); err != nil {
		return err
	}

	// 7.2.3 Segment header flags
	if err = h.readSegmentHeaderFlags(r); err != nil {
		return err
	}

	// 7.2.4 Amount of referred-to segment
	var countOfRTS uint64

	countOfRTS, err = h.readAmmountOfReferredToSegments(r)
	if err != nil {
		return err
	}

	// 7.2.5 Refered-tp segment numbers
	var rtsNumbers []int
	rtsNumbers, err = h.readReferedToSegmentNumbers(r, int(countOfRTS))
	if err != nil {
		return err
	}

	// 7.2.6 Segment page association
	err = h.readSegmentPageAssociation(d, r, countOfRTS, rtsNumbers...)
	if err != nil {
		return err
	}

	// 7.2.7 Segment data length (Contains the length of the data)
	if err = h.readSegmentDataLength(r); err != nil {
		return err
	}

	h.readDataStartOffset(r, organisationType)
	h.readSegmentHeaderLength(r, offset)

	return nil
}

// readSegmentNumber - 7.2.2
func (h *SegmentHeader) readSegmentNumber(r reader.StreamReader) error {
	// get 4 bytes
	b := make([]byte, 4)
	_, err := r.Read(b)
	if err != nil {
		return err
	}

	// BigEndian number
	h.SegmentNumber = binary.BigEndian.Uint32(b)

	return nil
}

// readSegmentHeaderFlags - 7.2.3
func (h *SegmentHeader) readSegmentHeaderFlags(r reader.StreamReader) error {
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

	h.SegmentType = SegmentType(int(tp))
	// common.Log.Debug("Segment Type: %v", h.SegmentType)
	// common.Log.Debug("PageAssociationSizeSet: %v", h.PageAssociationFieldSize)
	// common.Log.Debug("DeferredNonRetainSet: %v", h.RetainFlag)

	return nil
}

// readAmmountOfReferredToSegments - 7.2.4 get the amount of referred-to segments
func (h *SegmentHeader) readAmmountOfReferredToSegments(r reader.StreamReader) (uint64, error) {
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
		countOfRTS &= 0xffffffff
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

// readReferedToSegmentNumbers - 7.2.5 Gathers all segment numbers of referred-to segments. The
// segment itself is in rtSegments the array.
func (h *SegmentHeader) readReferedToSegmentNumbers(
	r reader.StreamReader, countOfRTS int,
) ([]int, error) {

	var rtsNumbers []int = make([]int, countOfRTS)

	if countOfRTS > 0 {

		var rtsSize byte = 1

		if h.SegmentNumber > 256 {

			rtsSize = 2

			if h.SegmentNumber > 65536 {
				rtsSize = 4
			}
		}

		h.RTSegments = make([]*SegmentHeader, countOfRTS)
		var (
			bits uint64
			err  error
		)
		for i := 0; i < countOfRTS; i++ {
			bits, err = r.ReadBits(rtsSize << 3)
			if err != nil {
				return nil, err
			}
			rtsNumbers[i] = int(bits & 0xffffffff)
		}
	}

	return rtsNumbers, nil
}

// readSegmentPageAssociation - 7.2.6
func (h *SegmentHeader) readSegmentPageAssociation(
	d *Document, r reader.StreamReader,
	countOfRTS uint64, rtsNumbers ...int,
) error {

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

		h.PageAssociation = int(bits & 0xFFFFFFFF)
	}

	if countOfRTS > 0 {
		page, err := d.GetPage(h.PageAssociation)
		if err != nil {
			return err
		}

		var i uint64

		for i = 0; i < countOfRTS; i++ {
			if page != nil {
				common.Log.Debug("RTSNumbers: %d", rtsNumbers[i])
				h.RTSegments[i] = page.GetSegment(rtsNumbers[i])
			} else {

				h.RTSegments[i] = d.GlobalSegments[rtsNumbers[i]]
				// TODO: d.getGlobalSegment(rtsNumber[i])
			}
		}
	}

	return nil
}

// readSegmentDataLength 7.2.7 - contains the length of the data part in bytes
func (h *SegmentHeader) readSegmentDataLength(r reader.StreamReader) (err error) {
	common.Log.Debug("SegmentData BitPosition: %d", r.BitPosition())
	h.SegmentDataLength, err = r.ReadBits(32)
	if err != nil {
		return err
	}

	// Set the 4bytes only mask
	h.SegmentDataLength &= 0xffffffff
	return nil
}

// readDataStartOffset sets the offset of the current reader if the organisation type
// is OSequential.
func (h *SegmentHeader) readDataStartOffset(r reader.StreamReader, organisationType uint8) {
	if organisationType == OSequential {
		common.Log.Debug("Sequential organisation")
		h.SegmentDataStartOffset = uint64(r.StreamPosition())
	}

}

func (h *SegmentHeader) readSegmentHeaderLength(r reader.StreamReader, offset int64) {
	h.SegmentHeaderLength = r.StreamPosition() - offset
}

func (h *SegmentHeader) cleanSegmentData() {
	if h.SegmentData != nil {
		h.SegmentData = nil
	}
}

func (h *SegmentHeader) subInputReader() (reader.StreamReader, error) {
	return reader.NewSubstreamReader(h.Reader, h.SegmentDataStartOffset, h.SegmentDataLength)
}

func (h *SegmentHeader) getSegmentData() (Segmenter, error) {
	var segmentDataPart Segmenter
	if h.SegmentData != nil {
		segmentDataPart = h.SegmentData
	}

	if segmentDataPart == nil {
		creator, ok := kindMap[h.SegmentType]
		if !ok {
			return nil, errors.Errorf("SegmentType: %s/ %d creator not found. ", h.SegmentType, h.SegmentType)
		}
		segmentDataPart = creator()

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

func (h *SegmentHeader) String() string {
	b := strings.Builder{}

	if h.RTSegments != nil {
		for _, s := range h.RTSegments {
			b.WriteString(fmt.Sprintf("%d ", s.SegmentNumber))
		}
	} else {
		b.WriteString("none")
	}

	return fmt.Sprintf("\n#SegmentNr: %d\n SegmentType: %s\n PageAssociation: %d\n Referred-to segments: %s\n", h.SegmentNumber, h.SegmentType, h.PageAssociation, b.String())
}

// // Decode reads the header from the provided reader
// // Implements io.ReaderFrom
// func (h *Header) Decode(r reader.StreamReader) (int, error) {
// 	common.Log.Debug("[SEGMENT-HEADER][DECODE] Begins")
// 	defer func() { common.Log.Debug("[SEGMENT-HEAD-ER][DECODE] Finished") }()

// 	common.Log.Debug("Reader ReadIndex: %04X", r.CurrentBytePosition())
// 	var bytesRead int
// 	n, err := h.handleSegmentNumber(r)
// 	if err != nil {
// 		if err != io.EOF {
// 			common.Log.Debug("handleSegmentNumber failed: %v.", err)
// 		}
// 		return bytesRead, err
// 	}

// 	bytesRead += n

// 	err = h.handleSegmentHeaderFlags(r)
// 	if err != nil {
// 		return bytesRead, err
// 	}

// 	bytesRead += 1

// 	n, err = h.handleSegmentReferredToCountAndRententionFlags(r)
// 	if err != nil {
// 		return bytesRead, err
// 	}
// 	bytesRead += n

// 	n, err = h.handleReferedSegmentNumbers(r)
// 	if err != nil {
// 		return bytesRead, err
// 	}

// 	bytesRead += n

// 	n, err = h.handlePageAssociation(r)
// 	if err != nil {
// 		return bytesRead, err
// 	}

// 	bytesRead += n

// 	if h.SegmentType != int(kind.EndOfFile) {
// 		n, err = h.handleSegmentDataLength(r)
// 		if err != nil {
// 			return bytesRead, err
// 		}

// 		bytesRead += n
// 	}

// 	return bytesRead, nil

// func (h *Header) handleSegmentReferredToCountAndRententionFlags(r reader.StreamReader) (int, error) {
// 	var bytesRead int
// 	b, err := r.ReadByte()
// 	if err != nil {
// 		return bytesRead, err
// 	}

// 	bytesRead += 1

// 	referedToSgmCountAndFlags := uint8(b)

// 	var referedToSegmentCount int

// 	referedToSegmentCount = int((referedToSgmCountAndFlags & 0xE0) >> 5)

// 	var retentionFlags []byte

// 	firstByte := byte(referedToSgmCountAndFlags & 31)

// 	if referedToSegmentCount <= 4 {
// 		// short form
// 		retentionFlags = make([]byte, 1)
// 		retentionFlags[0] = firstByte

// 	} else if referedToSegmentCount == 7 {
// 		// long form
// 		var longFormCountAndFlags []byte = make([]byte, 4)

// 		// add the first byte
// 		longFormCountAndFlags[0] = firstByte

// 		for i := 1; i < 4; i++ {

// 			sb, err := r.ReadByte()
// 			if err != nil {
// 				return bytesRead, err
// 			}

// 			bytesRead += 1

// 			longFormCountAndFlags[i] = sb
// 		}

// 		// get the count of the referred to segments
// 		referedToSegmentCount = int(binary.BigEndian.Uint32(longFormCountAndFlags))

// 		// calculate the number of bytes in this field

// 		var bytesCount float64 = float64(referedToSegmentCount) + 1.0

// 		var noOfBytesInField int = 4 + int(math.Ceil(bytesCount/8.0))

// 		var noOfRententionFlagBytes = noOfBytesInField - 4

// 		retentionFlags = make([]byte, noOfRententionFlagBytes)

// 		n, err := r.Read(retentionFlags)
// 		if err != nil {
// 			return n, err
// 		}

// 		bytesRead += n

// 	} else {
// 		return bytesRead, errors.Errorf("3 bit Segment count field = %v must not contain value of 5 and 6", referedToSegmentCount)
// 	}

// 	h.ReferredToSegmentCount = referedToSegmentCount
// 	common.Log.Debug("ReferredToSegmentCount: %v", referedToSegmentCount)

// 	h.RententionFlags = make([]int, len(retentionFlags))
// 	for i := 0; i < len(retentionFlags); i++ {
// 		h.RententionFlags[i] = int(retentionFlags[i])
// 	}

// 	return bytesRead, nil
// }

// func (h *Header) handlePageAssociation(r reader.StreamReader) (int, error) {
// 	var bytesRead int
// 	if h.PageAssociationSizeSet {
// 		buf := make([]byte, 4)
// 		n, err := r.Read(buf)
// 		if err != nil {
// 			return n, err
// 		}

// 		bytesRead = 4
// 		h.PageAssociation = int(binary.BigEndian.Uint32(buf))
// 	} else {
// 		b, err := r.ReadByte()
// 		if err != nil {
// 			return 1, err
// 		}
// 		bytesRead = 1
// 		h.PageAssociation = int(b)
// 	}

// 	common.Log.Debug("Page Association: %v", h.PageAssociation)
// 	return bytesRead, nil
// }

// func (h *Header) handleSegmentDataLength(r reader.StreamReader) (int, error) {
// 	length := make([]byte, 4)
// 	n, err := r.Read(length)
// 	if err != nil {
// 		return n, err
// 	}

// 	dataLength := binary.BigEndian.Uint32(length)
// 	h.DataLength = int(dataLength)

// 	common.Log.Debug("Data length: %v", dataLength)
// 	return n, nil
// }

// func (h *Header) setFlags(flags byte) {

// 	h.SegmentType = int(flags & 63)
// 	common.Log.Debug("Segment Type: %v", h.SegmentType)

// 	h.PageAssociationSizeSet = (flags & (1 << 6)) != 0
// 	common.Log.Debug("PageAssociationSizeSet: %v", h.PageAssociationSizeSet)

// 	h.DeferredNonRetainSet = (flags & (1 << 7)) != 0
// 	common.Log.Debug("DeferredNonRetainSet: %v", h.DeferredNonRetainSet)
// }
