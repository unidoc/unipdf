/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

import (
	"io"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// Define the constant arithmetic decoder tables.
var (
	qe = [][4]uint32{
		{0x5601, 1, 1, 1}, {0x3401, 2, 6, 0},
		{0x1801, 3, 9, 0}, {0x0AC1, 4, 12, 0}, {0x0521, 5, 29, 0}, {0x0221, 38, 33, 0},
		{0x5601, 7, 6, 1}, {0x5401, 8, 14, 0}, {0x4801, 9, 14, 0}, {0x3801, 10, 14, 0},
		{0x3001, 11, 17, 0}, {0x2401, 12, 18, 0}, {0x1C01, 13, 20, 0},
		{0x1601, 29, 21, 0}, {0x5601, 15, 14, 1}, {0x5401, 16, 14, 0},
		{0x5101, 17, 15, 0}, {0x4801, 18, 16, 0}, {0x3801, 19, 17, 0},
		{0x3401, 20, 18, 0}, {0x3001, 21, 19, 0}, {0x2801, 22, 19, 0},
		{0x2401, 23, 20, 0}, {0x2201, 24, 21, 0}, {0x1C01, 25, 22, 0},
		{0x1801, 26, 23, 0}, {0x1601, 27, 24, 0}, {0x1401, 28, 25, 0},
		{0x1201, 29, 26, 0}, {0x1101, 30, 27, 0}, {0x0AC1, 31, 28, 0},
		{0x09C1, 32, 29, 0}, {0x08A1, 33, 30, 0}, {0x0521, 34, 31, 0},
		{0x0441, 35, 32, 0}, {0x02A1, 36, 33, 0}, {0x0221, 37, 34, 0},
		{0x0141, 38, 35, 0}, {0x0111, 39, 36, 0}, {0x0085, 40, 37, 0},
		{0x0049, 41, 38, 0}, {0x0025, 42, 39, 0}, {0x0015, 43, 40, 0},
		{0x0009, 44, 41, 0}, {0x0005, 45, 42, 0}, {0x0001, 45, 43, 0},
		{0x5601, 46, 46, 0},
	}
)

// Decoder is the arithmetic Decoder structure, used to decode the jbig2 Segments.
type Decoder struct {
	// ContextSize is the current decoder context size
	ContextSize          []uint32
	ReferedToContextSize []uint32

	r              reader.StreamReader
	b              uint8
	c              uint64
	a              uint32
	previous       int64
	ct             int32
	prvCtr         int32
	streamPosition int64
}

// New creates new arithmetic Decoder.
func New(r reader.StreamReader) (*Decoder, error) {
	d := &Decoder{
		r:                    r,
		ContextSize:          []uint32{16, 13, 10, 10},
		ReferedToContextSize: []uint32{13, 10},
	}

	// initialize the decoder from the reader
	if err := d.init(); err != nil {
		return nil, err
	}

	return d, nil
}

// DecodeBit decodes a single bit using provided decoder stats.
func (d *Decoder) DecodeBit(stats *DecoderStats) (int, error) {
	var (
		bit     int
		qeValue = qe[stats.cx()][0]
		icx     = int32(stats.cx())
	)

	defer func() {
		d.prvCtr++
	}()

	d.a -= qeValue

	if (d.c >> 16) < uint64(qeValue) {
		bit = d.lpsExchange(stats, icx, qeValue)

		if err := d.renormalize(); err != nil {
			return 0, err
		}
	} else {
		d.c -= uint64(qeValue) << 16

		if (d.a & 0x8000) == 0 {
			bit = d.mpsExchange(stats, icx)
			if err := d.renormalize(); err != nil {
				return 0, err
			}
		} else {
			bit = int(stats.getMps())
		}
	}
	return bit, nil
}

// DecodeInt decodes the Integer from the arithmetic Decoder for the provided DecoderStats.
func (d *Decoder) DecodeInt(stats *DecoderStats) (int32, error) {
	var (
		value, offset      int32
		bit, s, bitsToRead int
		err                error
	)
	if stats == nil {
		stats = NewStats(512, 1)
	}
	d.previous = 1

	// First bit defines the sign of the integer.
	s, err = d.decodeIntBit(stats)
	if err != nil {
		return 0, err
	}

	bit, err = d.decodeIntBit(stats)
	if err != nil {
		return 0, err
	}

	// Read first bit.
	if bit == 1 {
		bit, err = d.decodeIntBit(stats)
		if err != nil {
			return 0, err
		}

		// Read second bit.
		if bit == 1 {
			bit, err = d.decodeIntBit(stats)
			if err != nil {
				return 0, err
			}

			// Read third bit.
			if bit == 1 {
				bit, err = d.decodeIntBit(stats)
				if err != nil {
					return 0, err
				}

				// Read fourth bit.
				if bit == 1 {
					bit, err = d.decodeIntBit(stats)
					if err != nil {
						return 0, err
					}

					// Read fifth bit.
					if bit == 1 {
						bitsToRead = 32
						offset = 4436
					} else {
						// Set fifth bit variables.
						bitsToRead = 12
						offset = 340
					}
				} else {
					// Set fourth bit variables.
					bitsToRead = 8
					offset = 84
				}
			} else {
				// Set third bit variables.
				bitsToRead = 6
				offset = 20
			}
		} else {
			// Set second bit variables.
			bitsToRead = 4
			offset = 4
		}
	} else {
		// Set first bit variables.
		bitsToRead = 2
		offset = 0
	}

	for i := 0; i < bitsToRead; i++ {
		bit, err = d.decodeIntBit(stats)
		if err != nil {
			return 0, err
		}
		value = (value << 1) | int32(bit)
	}
	value += offset

	if s == 0 {
		return value, nil
	} else if s == 1 && value > 0 {
		return -value, nil
	}
	return math.MaxInt32, nil
}

// DecodeIAID decodes the IAID procedure, Annex A.3.
func (d *Decoder) DecodeIAID(codeLen uint64, stats *DecoderStats) (int64, error) {
	// A.3 1)
	d.previous = 1
	var i uint64

	// A.3 2)
	for i = 0; i < codeLen; i++ {
		stats.SetIndex(int32(d.previous))
		bit, err := d.DecodeBit(stats)
		if err != nil {
			return 0, err
		}

		d.previous = (d.previous << 1) | int64(bit)
	}

	// A.3 3) & 5)
	result := d.previous - (1 << codeLen)
	return result, nil
}

func (d *Decoder) init() error {
	d.streamPosition = d.r.StreamPosition()
	b, err := d.r.ReadByte()
	if err != nil {
		common.Log.Debug("Buffer0 readByte failed. %v", err)
		return err
	}

	d.b = b
	d.c = uint64(b) << 16

	if err = d.readByte(); err != nil {
		return err
	}

	d.c <<= 7
	d.ct -= 7
	d.a = 0x8000
	d.prvCtr++

	return nil
}

func (d *Decoder) readByte() error {
	if d.r.StreamPosition() > d.streamPosition {
		if _, err := d.r.Seek(-1, io.SeekCurrent); err != nil {
			return err
		}
	}

	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	d.b = b

	if d.b == 0xFF {
		b1, err := d.r.ReadByte()
		if err != nil {
			return err
		}

		if b1 > 0x8F {
			d.c += 0xFF00
			d.ct = 8
			if _, err := d.r.Seek(-2, io.SeekCurrent); err != nil {
				return err
			}
		} else {
			d.c += uint64(b1) << 9
			d.ct = 7
		}
	} else {
		b, err = d.r.ReadByte()
		if err != nil {
			return err
		}
		d.b = b

		d.c += uint64(d.b) << 8
		d.ct = 8
	}
	d.c &= 0xFFFFFFFFFF
	return nil
}

func (d *Decoder) renormalize() error {
	for {
		if d.ct == 0 {
			if err := d.readByte(); err != nil {
				return err
			}
		}

		d.a <<= 1
		d.c <<= 1
		d.ct--

		if (d.a & 0x8000) != 0 {
			break
		}
	}

	d.c &= 0xffffffff
	return nil
}

func (d *Decoder) decodeIntBit(stats *DecoderStats) (int, error) {
	stats.SetIndex(int32(d.previous))
	bit, err := d.DecodeBit(stats)
	if err != nil {
		common.Log.Debug("ArithmeticDecoder 'decodeIntBit'-> DecodeBit failed. %v", err)
		return bit, err
	}

	if d.previous < 256 {
		d.previous = ((d.previous << uint64(1)) | int64(bit)) & 0x1ff
	} else {
		d.previous = (((d.previous<<uint64(1) | int64(bit)) & 511) | 256) & 0x1ff
	}
	return bit, nil
}

func (d *Decoder) mpsExchange(stats *DecoderStats, icx int32) int {
	mps := stats.mps[stats.index]

	if d.a < qe[icx][0] {
		if qe[icx][3] == 1 {
			stats.toggleMps()
		}

		stats.setEntry(int(qe[icx][2]))
		return int(1 - mps)
	}
	stats.setEntry(int(qe[icx][1]))
	return int(mps)

}

func (d *Decoder) lpsExchange(stats *DecoderStats, icx int32, qeValue uint32) int {
	mps := stats.getMps()
	if d.a < qeValue {
		stats.setEntry(int(qe[icx][1]))
		d.a = qeValue
		return int(mps)
	}

	if qe[icx][3] == 1 {
		stats.toggleMps()
	}

	stats.setEntry(int(qe[icx][2]))
	d.a = qeValue
	return int(1 - mps)
}
