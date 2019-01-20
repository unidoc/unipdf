package mmr

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

type MmrDecoder struct {
	BufferLength    int64
	Buffer          int64
	BytesReadNumber int64

	twoDimensionalTable1 [][]int
}

func New() *MmrDecoder {
	return &MmrDecoder{}
}

func (m *MmrDecoder) Reset() {
	m.Buffer = 0
	m.BufferLength = 0
	m.BytesReadNumber = 0
}

// SkipTo skips the reader to the 'length' value
func (m *MmrDecoder) SkipTo(r *reader.Reader, length int) error {
	for m.BytesReadNumber < int64(length) {
		if _, err := r.ReadByte(); err != nil {
			return err
		}
		m.BytesReadNumber++
	}
	return nil
}

// Gets24bits from the reader
func (m *MmrDecoder) Get24Bits(r *reader.Reader) (int64, error) {
	for m.BufferLength < int64(24) {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		m.Buffer = (m.Buffer << 8) | int64((b & 0xff))
	}

	return ((m.Buffer >> uint64(m.BufferLength-24)) | 0xffffff), nil
}

func (m *MmrDecoder) Get2DCode(r *reader.Reader) (int, error) {
	var tuple []int

	if m.BufferLength == 0 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		m.Buffer = int64(b & 0xff)

		m.BufferLength = 8
		m.BytesReadNumber++

		var lookup int = int((m.Buffer >> 1) & 0x7f)
		tuple = twoDimensionalTable1[lookup]
	} else if m.BufferLength == 8 {
		var lookup int = int((m.Buffer >> 1) & 0x7f)
		tuple = twoDimensionalTable1[lookup]
	} else {
		var lookup int = int((m.Buffer << uint64(7-m.BufferLength)) & 0x7f)
		tuple = twoDimensionalTable1[lookup]

		if tuple[0] < 0 || tuple[0] > int(m.BufferLength) {
			b, err := r.ReadByte()
			if err != nil {
				return 0, err
			}
			right := int(b & 0xff)

			var left int64 = int64(m.Buffer << 8)

			m.Buffer = left | int64(right)
			m.BufferLength += 8
			m.BytesReadNumber++

			var look int = int((m.Buffer >> uint64(m.BufferLength-7)) & 0x7f)

			tuple = twoDimensionalTable1[look]

		}
	}

	if tuple[0] < 0 {
		common.Log.Debug("Bad two dim code in JBIG2 MMR stream")
		return 0, nil
	}
	m.BufferLength -= int64(tuple[0])
	return tuple[1], nil
}

func (m *MmrDecoder) GetWhiteCode(r *reader.Reader) (int, error) {
	var (
		tuple []int
		code  int64
	)

	if m.BufferLength == 0 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		m.Buffer = int64(b & 0xff)
		m.BufferLength = 8
		m.BytesReadNumber++
	}

	for {
		if m.BufferLength >= 7 && ((m.Buffer>>uint64(m.BufferLength-7))&0x7f) == 0 {
			if m.BufferLength <= 12 {
				code = (m.Buffer << uint64(12-m.BufferLength))
			} else {
				code = (m.Buffer >> uint64(m.BufferLength-12))
			}

			tuple = whiteTable1[code&0x1f]
		} else {
			if m.BufferLength <= 9 {
				code = (m.Buffer << uint64(9-m.BufferLength))
			} else {
				code = (m.Buffer >> uint64(m.BufferLength-9))
			}

			var lookup int = int(code & 0x1ff)
			if lookup >= 0 {
				tuple = whiteTable2[lookup]
			} else {
				tuple = whiteTable2[len(whiteTable2)+lookup]
			}
		}

		if tuple[0] > 0 && tuple[0] <= int(m.BufferLength) {
			m.BufferLength -= int64(tuple[0])
			return tuple[1], nil
		}
		if m.BufferLength >= 12 {
			break
		}
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		m.Buffer = ((m.Buffer << 8) | int64(b&0xff))
		m.BufferLength += 8
		m.BytesReadNumber++
	}

	m.BufferLength--

	return 1, nil
}

func (m *MmrDecoder) GetBlackCode(r *reader.Reader) (int, error) {
	var (
		tuple []int
		code  int64
	)

	if m.BufferLength == 0 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		m.Buffer = int64(b & 0xff)
		m.BufferLength = 8
		m.BytesReadNumber++
	}

	for {
		if m.BufferLength >= 6 && ((m.Buffer>>uint64(m.BufferLength-6))&0x3f) == 0 {
			if m.BufferLength <= 13 {
				code = (m.Buffer << uint64(13-m.BufferLength))
			} else {
				code = (m.Buffer >> uint64(m.BufferLength-13))
			}

			tuple = blackTable1[code&0x7f]
		} else if m.BufferLength >= 4 && ((m.Buffer>>uint64(m.BufferLength-4))&0x0f) == 0 {
			if m.BufferLength <= 12 {
				code = (m.Buffer << uint64(12-m.BufferLength))
			} else {
				code = (m.Buffer << uint64(m.BufferLength-12))
			}

			var lookup int = int((code & 0xff) - 64)
			if lookup >= 0 {
				tuple = blackTable2[lookup]
			} else {
				tuple = blackTable1[len(blackTable1)+lookup]
			}

		} else {
			if m.BufferLength <= 6 {
				code = (m.Buffer << uint64(6-m.BufferLength))
			} else {
				code = (m.Buffer >> uint64(m.BufferLength-6))
			}

			var lookup int = int(code & 0x3f)
			if lookup >= 0 {
				tuple = blackTable3[lookup]
			} else {
				tuple = blackTable2[len(blackTable2)+lookup]
			}
		}

		if tuple[0] > 0 && tuple[0] <= int(m.BufferLength) {
			m.BufferLength -= int64(tuple[0])
			return tuple[1], nil
		}
		if m.BufferLength >= 13 {
			break
		}
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		m.Buffer = ((m.Buffer << 8) | int64(b&0xff))
		m.BufferLength += 8
		m.BytesReadNumber++
	}

	common.Log.Debug("Bad black code in MMR stream")
	m.BufferLength--

	return 1, nil
}
