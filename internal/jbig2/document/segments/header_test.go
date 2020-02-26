/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// TestDecodeHeader test the segment header model decode process.
func TestDecodeHeader(t *testing.T) {
	data := []byte{
		// header
		0x00, 0x00, 0x00, 0x0B, 0x27, 0x00, 0x02, 0x00, 0x00, 0x00, 0x23,

		// data
		0x00, 0x00, 0x00, 0x36, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00, 0x00, 0x04,
		0x00, 0x00, 0x00, 0x0B, 0x00, 0x08, 0x03, 0xFF, 0xFD, 0xFF, 0x02, 0xFE,
		0xFE, 0xFE, 0x04, 0xEE, 0xED, 0x87, 0xFB, 0xCB, 0x2B, 0xFF, 0xAC,
	}

	r := reader.New(data)
	d := &document{}
	h, err := NewHeader(d, r, 0, OSequential)
	require.NoError(t, err)

	assert.Equal(t, int64(11), h.HeaderLength)
	assert.Equal(t, uint64(11), h.SegmentDataStartOffset)

	s, err := h.subInputReader()
	require.NoError(t, err)

	b, err := s.ReadByte()
	require.NoError(t, err)
	assert.Equal(t, byte(0x00), b)

	three := make([]byte, 3)
	read, err := s.Read(three)
	require.NoError(t, err)

	assert.Equal(t, 3, read)
	assert.Equal(t, byte(0x36), three[2])
}

// TestWriteHeader tests the write function of the segment Header.
func TestWriteHeader(t *testing.T) {
	t.Run("ReferenceSize#1", func(t *testing.T) {
		// the header with reference size '1' have it's segment number within the range (0, 255>
		// it is refered to under 4 segments - this means that their values are represented
		// in the uint8 - byte format.
		// the associated page number may fit in a single byte - the page association flag should be false.
		initial := &Header{
			PageAssociation:   1,
			RTSegments:        []*Header{{SegmentNumber: 2}, {SegmentNumber: 3}, {SegmentNumber: 4}},
			SegmentDataLength: 64,
			SegmentNumber:     2,
			Type:              TImmediateTextRegion,
		}
		w := writer.BufferedMSB()

		n, err := initial.Encode(w)
		require.NoError(t, err)
		assert.Equal(t, 14, n)

		// the data should look like:
		//
		// 00000000 00000000 00000000 00000010  - 0x00, 0x00, 0x00, 0x02 - segment number
		// 00000110 - 0x06 - segment flags
		// 01100000 - 0x60 - referred to count and retain flags
		// 00000010 00000011 00000100 - 0x02, 0x03, 0x04 - segment referred to numbers
		// 00000001 - 0x01 - segment page association
		// 00000000 00000000 00000000 01000000 - 0x40 - segment data length
		expected := []byte{
			0x00, 0x00, 0x00, 0x02,
			0x06,
			0x60,
			0x02, 0x03, 0x04,
			0x01,
			0x00, 0x00, 0x00, 0x40,
		}
		assert.Equal(t, expected, w.Data())

		// check if the encoded data is now decodable with the same results.
		// as the 'SegmentDataLength' = 64 the source data needs 64 bytes of empty data to
		// read the header properly.
		r := reader.New(append(w.Data(), make([]byte, 64)...))
		d := &document{pages: []Pager{&page{segments: []*Header{{}, {}, {}, {}, {}}}, &page{}}}

		h, err := NewHeader(d, r, 0, OSequential)
		require.NoError(t, err)

		assert.Equal(t, initial.PageAssociation, h.PageAssociation)
		assert.Equal(t, initial.SegmentDataLength, h.SegmentDataLength)
		assert.Equal(t, initial.SegmentNumber, h.SegmentNumber)
		assert.Equal(t, initial.Type, h.Type)
		assert.Equal(t, initial.RTSNumbers, h.RTSNumbers)
		assert.False(t, h.PageAssociationFieldSize)
	})

	t.Run("ReferenceSize#2", func(t *testing.T) {
		// the reference size is equal to '2' when the segment has a number at least 256
		// and at most 65535.
		// this example has also referred to symbols number greater than 4 - which
		// enforces the refered to count segment to be more than a byte long and the
		// refered to segments to be of uint16 size.
		//
		initial := &Header{
			PageAssociation:   2,
			RTSNumbers:        []int{2, 3, 4, 5, 6}, // the size > 4
			SegmentDataLength: 64,
			SegmentNumber:     257,
			Type:              TIntermediateGenericRegion, // 36
		}
		for _, nm := range initial.RTSNumbers {
			initial.RTSegments = append(initial.RTSegments, &Header{SegmentNumber: uint32(nm)})
		}
		w := writer.BufferedMSB()

		n, err := initial.Encode(w)
		require.NoError(t, err)

		// the data should look like:
		//
		// 00000000 00000000 00000001 00000001  - 0x00, 0x00, 0x01, 0x01 - segment number
		// 00100100 - 0x24 - segment flags
		// there should be 4+ceil((len(rts) + 1) / 8) bytes = 4 + ceil((5+1)/8) = 5 bytes.
		// 11100000 00000000 00000000 00000101 00000000 - 0xE0, 0x00, 0x00, 0x05, 0x00 - number of related segments and retain flags
		// the refered to numbers should be of 16bits size each.
		// 00000000 00000010 - 0x00, 0x02
		// 00000000 00000011 - 0x00, 0x03
		// 00000000 00000100 - 0x00, 0x04
		// 00000000 00000101 - 0x00, 0x05
		// 00000000 00000110 - 0x00, 0x06 - refered to segment numbers
		// 00000010 - 0x02 - segment page association
		// 00000000 00000000 00000000 01000000 - 0x00, 0x00, 0x00, 0x40 - segment data length
		//
		// at total 25 bytes for the header.
		f := assert.Equal(t, 25, n)
		expected := []byte{
			0x00, 0x00, 0x01, 0x01,
			0x24,
			0xE0, 0x00, 0x00, 0x05, 0x00,
			0x00, 0x02, 0x00, 0x03, 0x00, 0x04, 0x00, 0x05, 0x00, 0x06,
			0x02,
			0x00, 0x00, 0x00, 0x40,
		}
		s := assert.Equal(t, expected, w.Data())
		require.True(t, f && s)

		// check if the decoder would decode that header.
		r := reader.New(append(w.Data(), make([]byte, 64)...))
		p := &page{segments: make([]*Header, 7)}
		d := &document{pages: []Pager{&page{}, p}}

		h, err := NewHeader(d, r, 0, OSequential)
		require.NoError(t, err)

		assert.Equal(t, initial.PageAssociation, h.PageAssociation)
		assert.Equal(t, initial.SegmentDataLength, h.SegmentDataLength)
		assert.Equal(t, initial.SegmentNumber, h.SegmentNumber)
		assert.Equal(t, initial.Type, h.Type)
		assert.Equal(t, initial.RTSNumbers, h.RTSNumbers)
	})

	t.Run("ReferenceSize#4", func(t *testing.T) {
		// the reference size is equal to '4' when the segment has a number at least 65536.
		// this example has also referred to symbols number greater than 255 - which
		// enforces the refered to count segment to be more than a byte long and the refered to segment numbers
		// to be of uint32 size.
		//
		initial := &Header{
			PageAssociation:   256,
			RTSegments:        make([]*Header, 256), // the size > 4
			SegmentDataLength: 64,
			SegmentNumber:     65536,
			Type:              TIntermediateGenericRegion, // 36
		}

		// seed the refered to segment numbers.
		for i := uint32(0); i < 256; i++ {
			initial.RTSegments[i] = &Header{SegmentNumber: i + initial.SegmentNumber}
		}
		w := writer.BufferedMSB()

		n, err := initial.Encode(w)
		require.NoError(t, err)

		// the data should look like:
		//
		// 00000000 00000001 00000000 00000000  - 0x00, 0x01, 0x00, 0x00 - segment number
		// 01100100 - 0x64 - segment flags, the page association flag is 'on'.
		// there should be 4+ceil((len(rts) + 1) / 8) bytes = 4 + ceil((256+1)/8) = 4 + 33 = 37 bytes.
		// 11100000 00000000 00000001 00000000 - 0xE0, 0x00, 0x01, 0x00 - number of related segments
		// 37 * 00000000 (0x00) - retain flags
		// the refered to numbers should be of 32bits size each.
		// 00000000 00000001 00000000 00000001 - the first refered to segment number
		// 00000000 00000001 00000000 00000010
		// 00000000 00000001 00000000 00000011
		// ...................................
		// 00000000 00000001 00000001 00000000 - the last refered to segment number
		// 00000000 00000000 00000001 00000000 - 0x00, 0x00, 0x01, 0x00 - segment page association
		// 00000000 00000000 00000000 01000000 - 0x00, 0x00, 0x00, 0x40 - segment data length
		//
		// at total 4 + 1 + 37 + 4 * 256 + 4 + 4 = 1074 bytes
		f := assert.Equal(t, 1074, n)
		expected := []byte{
			0x00, 0x01, 0x00, 0x00, // segment number
			0x64,                   // header flags
			0xE0, 0x00, 0x01, 0x00, // number of referred to segments and '7' bitwise value on the first three bits.
		}

		// seed the retain flags
		for i := 0; i < 33; i++ {
			expected = append(expected, 0x00)
		}

		// initialize referred to segment numbers
		referedToSegments := make([]byte, 4*256)
		for i := 0; i < 256; i++ {
			binary.BigEndian.PutUint32(referedToSegments[i*4:], uint32(i)+initial.SegmentNumber)
		}
		expected = append(expected, referedToSegments...)
		// append the page info segment and the segment data length
		expected = append(expected, []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x40}...)

		// the writer data should be equal to the 'expected' data.
		s := assert.Equal(t, expected, w.Data())
		require.True(t, f && s)

		// check if the decoder would decode that header.
		r := reader.New(append(w.Data(), make([]byte, 64)...))
		// initialize testing document with the page no 256
		d := &document{pages: make([]Pager, 257)}
		// initialize page segments
		p := &page{segments: make([]*Header, 65537+256)}
		// fill the page with segments
		for i := 0; i < 256; i++ {
			p.segments[65536+i] = &Header{}
		}
		// set the document's page to 'p'
		d.pages[255] = p

		// try to decode the header.
		h, err := NewHeader(d, r, 0, OSequential)
		require.NoError(t, err)

		assert.Equal(t, initial.PageAssociation, h.PageAssociation)
		assert.Equal(t, initial.SegmentDataLength, h.SegmentDataLength)
		assert.Equal(t, initial.SegmentNumber, h.SegmentNumber)
		assert.Equal(t, initial.Type, h.Type)
		assert.Equal(t, initial.RTSNumbers, h.RTSNumbers)
		assert.True(t, h.PageAssociationFieldSize)
	})
}
