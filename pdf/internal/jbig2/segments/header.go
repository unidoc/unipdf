package segment

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
	"io"
	"math"
)

var log common.Logger = common.Log

// Header is the segment header used to define the segment parameters
type Header struct {
	SegmentNumber int
	SegmentType   int

	PageAssociationSizeSet bool
	DeferredNonRetainSet   bool

	ReferredToSegmentCount int
	RententionFlags        []int

	ReferredToSegments []int
	PageAssociation    int
	DataLength         int

	SegmentData *
}

// Decode reads the header from the provided reader
// Implements io.ReaderFrom
func (h *Header) Decode(r *reader.Reader) (int, error) {
	common.Log.Debug("[SEGMENT-HEADER][DECODE] Begins")
	defer func() { common.Log.Debug("[SEGMENT-HEADER][DECODE] Finished") }()

	common.Log.Debug("Reader ReadIndex: %04X", r.CurrentBytePosition())
	var bytesRead int
	n, err := h.handleSegmentNumber(r)
	if err != nil {
		if err != io.EOF {
			common.Log.Debug("handleSegmentNumber failed: %v.", err)
		}
		return bytesRead, err
	}

	bytesRead += n

	err = h.handleSegmentHeaderFlags(r)
	if err != nil {
		return bytesRead, err
	}

	bytesRead += 1

	n, err = h.handleSegmentReferredToCountAndRententionFlags(r)
	if err != nil {
		return bytesRead, err
	}
	bytesRead += n

	n, err = h.handleReferedSegmentNumbers(r)
	if err != nil {
		return bytesRead, err
	}

	bytesRead += n

	n, err = h.handlePageAssociation(r)
	if err != nil {
		return bytesRead, err
	}

	bytesRead += n

	if h.SegmentType != int(kind.EndOfFile) {
		n, err = h.handleSegmentDataLength(r)
		if err != nil {
			return bytesRead, err
		}

		bytesRead += n
	}

	return bytesRead, nil

}

func (h *Header) handleSegmentNumber(r *reader.Reader) (int, error) {
	b := make([]byte, 4)

	n, err := r.Read(b)
	if err != nil {
		return n, err
	}

	//

	sgNumber := binary.BigEndian.Uint32(b)
	h.SegmentNumber = int(sgNumber)
	return n, nil
}

func (h *Header) handleSegmentHeaderFlags(r *reader.Reader) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}

	h.setFlags(b)

	return nil
}

func (h *Header) handleSegmentReferredToCountAndRententionFlags(r *reader.Reader) (int, error) {
	var bytesRead int
	b, err := r.ReadByte()
	if err != nil {
		return bytesRead, err
	}

	bytesRead += 1

	referedToSgmCountAndFlags := uint8(b)

	var referedToSegmentCount int

	referedToSegmentCount = int((referedToSgmCountAndFlags & 0xE0) >> 5)

	var retentionFlags []byte

	firstByte := byte(referedToSgmCountAndFlags & 31)

	if referedToSegmentCount <= 4 {
		// short form
		retentionFlags = make([]byte, 1)
		retentionFlags[0] = firstByte

	} else if referedToSegmentCount == 7 {
		// long form
		var longFormCountAndFlags []byte = make([]byte, 4)

		// add the first byte
		longFormCountAndFlags[0] = firstByte

		for i := 1; i < 4; i++ {

			sb, err := r.ReadByte()
			if err != nil {
				return bytesRead, err
			}

			bytesRead += 1

			longFormCountAndFlags[i] = sb
		}

		// get the count of the referred to segments
		referedToSegmentCount = int(binary.BigEndian.Uint32(longFormCountAndFlags))

		// calculate the number of bytes in this field

		var bytesCount float64 = float64(referedToSegmentCount) + 1.0

		var noOfBytesInField int = 4 + int(math.Ceil(bytesCount/8.0))

		var noOfRententionFlagBytes = noOfBytesInField - 4

		retentionFlags = make([]byte, noOfRententionFlagBytes)

		n, err := r.Read(retentionFlags)
		if err != nil {
			return n, err
		}

		bytesRead += n

	} else {
		return bytesRead, errors.Errorf("3 bit Segment count field = %v must not contain value of 5 and 6", referedToSegmentCount)
	}

	h.ReferredToSegmentCount = referedToSegmentCount
	common.Log.Debug("ReferredToSegmentCount: %v", referedToSegmentCount)

	h.RententionFlags = make([]int, len(retentionFlags))
	for i := 0; i < len(retentionFlags); i++ {
		h.RententionFlags[i] = int(retentionFlags[i])
	}

	return bytesRead, nil
}

func (h *Header) handleReferedSegmentNumbers(r *reader.Reader) (int, error) {
	h.ReferredToSegments = make([]int, h.ReferredToSegmentCount)
	var (
		bytesRead int
	)

	if h.SegmentNumber <= 256 {

		for i := 0; i < h.ReferredToSegmentCount; i++ {
			b, err := r.ReadByte()
			if err != nil {
				return bytesRead, err
			}
			h.ReferredToSegments[i] = int(b)
			bytesRead += 1
		}
	} else if h.SegmentNumber <= 65536 {

		b := make([]byte, 2)
		for i := 0; i < h.ReferredToSegmentCount; i++ {
			n, err := r.Read(b)
			if err != nil {
				return bytesRead, err
			}

			bytesRead += n
			s := binary.BigEndian.Uint16(b)

			h.ReferredToSegments[i] = int(s)

		}
	} else {
		b := make([]byte, 4)
		for i := 0; i < h.ReferredToSegmentCount; i++ {
			n, err := r.Read(b)
			if err != nil {
				return bytesRead, err
			}
			bytesRead += n
			s := binary.BigEndian.Uint32(b)
			h.ReferredToSegments[i] = int(s)
		}
	}

	return bytesRead, nil
}

func (h *Header) handlePageAssociation(r *reader.Reader) (int, error) {
	var bytesRead int
	if h.PageAssociationSizeSet {
		buf := make([]byte, 4)
		n, err := r.Read(buf)
		if err != nil {
			return n, err
		}

		bytesRead = 4
		h.PageAssociation = int(binary.BigEndian.Uint32(buf))
	} else {
		b, err := r.ReadByte()
		if err != nil {
			return 1, err
		}
		bytesRead = 1
		h.PageAssociation = int(b)
	}

	common.Log.Debug("Page Association: %v", h.PageAssociation)
	return bytesRead, nil
}

func (h *Header) handleSegmentDataLength(r *reader.Reader) (int, error) {
	length := make([]byte, 4)
	n, err := r.Read(length)
	if err != nil {
		return n, err
	}

	dataLength := binary.BigEndian.Uint32(length)
	h.DataLength = int(dataLength)

	common.Log.Debug("Data length: %v", dataLength)
	return n, nil
}

func (h *Header) setFlags(flags byte) {

	h.SegmentType = int(flags & 63)
	common.Log.Debug("Segment Type: %v", h.SegmentType)

	h.PageAssociationSizeSet = (flags & (1 << 6)) != 0
	common.Log.Debug("PageAssociationSizeSet: %v", h.PageAssociationSizeSet)

	h.DeferredNonRetainSet = (flags & (1 << 7)) != 0
	common.Log.Debug("DeferredNonRetainSet: %v", h.DeferredNonRetainSet)
}
