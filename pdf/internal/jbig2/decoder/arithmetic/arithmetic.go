package arithmetic

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"io"
)

var (
	qe [][4]uint32 = [][4]uint32{{0x5601, 1, 1, 1}, {0x3401, 2, 6, 0},
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
		{0x5601, 46, 46, 0}}

	qeTable []int = []int{0x56010000, 0x34010000, 0x18010000, 0x0AC10000, 0x05210000, 0x02210000, 0x56010000, 0x54010000, 0x48010000, 0x38010000, 0x30010000, 0x24010000, 0x1C010000, 0x16010000, 0x56010000, 0x54010000, 0x51010000, 0x48010000, 0x38010000, 0x34010000, 0x30010000, 0x28010000, 0x24010000, 0x22010000, 0x1C010000, 0x18010000, 0x16010000, 0x14010000, 0x12010000, 0x11010000, 0x0AC10000, 0x09C10000, 0x08A10000, 0x05210000, 0x04410000, 0x02A10000, 0x02210000, 0x01410000, 0x01110000, 0x00850000, 0x00490000, 0x00250000, 0x00150000, 0x00090000, 0x00050000, 0x00010000,
		0x56010000}

	nmpsTable []int = []int{
		1, 2, 3, 4, 5, 38, 7, 8, 9, 10, 11, 12, 13, 29, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
		30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43,
		44, 45, 45, 46,
	}

	nlpsTable []int = []int{
		1, 6, 9, 12, 29, 33, 6, 14, 14, 14, 17, 18, 20, 21, 14, 14, 15, 16, 17, 18, 19,
		19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
		39, 40, 41, 42, 43, 46,
	}

	switchTable []int = []int{
		1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
)

// Decoder is the arithmetic Decoder structure that is used to decode the
// segments in an arithmetic method.
type Decoder struct {
	GenericRegionStats, RefinementRegionStats *DecoderStats

	// IaaiStats is used to decode the number of symbol instances in an aggregation
	IaaiStats *DecoderStats

	// IadhStats is used to decode the difference in height between two height classes
	IadhStats *DecoderStats

	// IadwStats is used to decode the difference in width between two symbols in a height class
	IadwStats *DecoderStats

	// IaexStats is used to decode export flags
	IaexStats *DecoderStats

	// IadtStats is used to decode the T coordinate of the second and subsequent symbol instances
	// in a strip
	IadtStats *DecoderStats

	// IaitStats is used to decode the T coordinate of the symbol instances in a strip
	IaitStats *DecoderStats

	// IafsStats is used to decode the S coordinate of the first symbol instance in a strip
	IafsStats *DecoderStats

	// IadsStats is used to decode the S coordinate of the second and subsequent symbol instances in a strip
	IadsStats *DecoderStats

	// IardxStats is used to decode the delta X position of symbol instance refinements
	IardxStats *DecoderStats

	// IardyStats is used to decode the delta Y position of symbol instance refinements
	IardyStats *DecoderStats

	// IardwStats is used to decode the delta width of symbol instance refinements
	IardwStats *DecoderStats

	// IardhStats is used to decode the delta height of symbol instance refinements
	IardhStats *DecoderStats

	// IariStats is used to decode the R_i bit of symbol instances
	IariStats *DecoderStats

	// IaidStats is used to decode the symbol IDs of symbol instances
	IaidStats *DecoderStats

	ContextSize          []int
	ReferedToContextSize []int

	b int

	c        uint64
	a        uint32
	previous int64

	ct int

	prvCtr int

	streamPosition int64
}

// New creates new arithmetic Decoder
func New() *Decoder {
	d := &Decoder{
		GenericRegionStats:    NewStats(1<<1, 1),
		RefinementRegionStats: NewStats(1<<1, 1),

		IadhStats:            NewStats(512, 1),
		IadwStats:            NewStats(512, 1),
		IaexStats:            NewStats(512, 1),
		IaaiStats:            NewStats(512, 1),
		IadtStats:            NewStats(512, 1),
		IaitStats:            NewStats(512, 1),
		IafsStats:            NewStats(512, 1),
		IadsStats:            NewStats(512, 1),
		IardxStats:           NewStats(512, 1),
		IardyStats:           NewStats(512, 1),
		IardwStats:           NewStats(512, 1),
		IardhStats:           NewStats(512, 1),
		IariStats:            NewStats(512, 1),
		IaidStats:            NewStats(512, 1),
		ContextSize:          []int{16, 13, 10, 10},
		ReferedToContextSize: []int{13, 10},
	}

	return d
}

// ResetIntStats resets the context stats for the decoder
func (d *Decoder) ResetIntStats(symbolCodeLength int) {
	d.IadhStats.Reset()
	d.IadwStats.Reset()
	d.IaexStats.Reset()
	d.IaaiStats.Reset()
	d.IadtStats.Reset()
	d.IaitStats.Reset()
	d.IafsStats.Reset()
	d.IadsStats.Reset()
	d.IardxStats.Reset()
	d.IardyStats.Reset()
	d.IardwStats.Reset()
	d.IardhStats.Reset()
	d.IariStats.Reset()

	if d.IaidStats.contextSize == (1 << uint(symbolCodeLength+1)) {
		d.IaidStats.Reset()
	} else {
		d.IaidStats = NewStats(1<<uint(symbolCodeLength+1), 1)
	}
}

// ResetGenericStats resets the decoder's generic stats.
// If the previousStats are not nil then the decoder would copy
// the genericRegionStats from the previous stats
func (d *Decoder) ResetGenericStats(template int, previousStats *DecoderStats) {
	size := d.ContextSize[template]

	if previousStats != nil && previousStats.contextSize == size {
		if d.GenericRegionStats.contextSize == size {
			d.GenericRegionStats.Overwrite(previousStats)
		} else {
			d.GenericRegionStats = previousStats.Copy()
		}
	} else {
		if d.GenericRegionStats.contextSize == size {
			d.GenericRegionStats.Reset()
		} else {
			d.GenericRegionStats = NewStats(1<<uint(size), 1)
		}
	}
}

// ResetRefinementStats resets the RefinementsRegionStats for the given template
// If the previouseStats are not 'nil' then the values are set from the previousStats
func (d *Decoder) ResetRefinementStats(template int, previousStats *DecoderStats) {
	size := d.ContextSize[template]

	if previousStats != nil && previousStats.contextSize == size {
		if d.RefinementRegionStats.contextSize == size {
			d.RefinementRegionStats.Overwrite(previousStats)
		} else {
			d.RefinementRegionStats = previousStats.Copy()
		}
	} else {
		if d.RefinementRegionStats.contextSize == size {
			d.RefinementRegionStats.Reset()
		} else {
			d.RefinementRegionStats = NewStats(1<<uint(size), 1)
		}
	}
}

func (d *Decoder) Start(r *reader.Reader) error {
	d.streamPosition = r.CurrentBytePosition()
	b, err := r.ReadByte()
	if err != nil {
		common.Log.Debug("Buffer0 readByte failed. %v", err)
		return err
	}
	d.b = int(b)

	d.c = (uint64(b) << 16)
	if err = d.readByte(r); err != nil {
		return err
	}

	// Shift c value again
	d.c <<= 7

	// set
	d.ct -= 7

	d.a = 0x8000

	d.prvCtr += 1
	// common.Log.Debug("Decoder 'a' value: %b", d.a)
	common.Log.Debug("%d, C: %08x A: %04x, CTR: %d, B0: %02x", d.prvCtr, d.c, d.a, d.ct, d.b)

	return nil
}

func (d *Decoder) readByte(r *reader.Reader) error {

	if r.CurrentBytePosition() > d.streamPosition {
		if _, err := r.Seek(-1, io.SeekCurrent); err != nil {
			return err
		}
	}

	b, err := r.ReadByte()
	if err != nil {
		return err
	}

	d.b = int(b)

	if d.b == 0xFF {
		b1, err := r.ReadByte()
		if err != nil {
			return err
		}
		if b1 > 0x8F {
			d.c += 0xFF00
			d.ct = 8
			if _, err := r.Seek(-2, io.SeekCurrent); err != nil {
				return err
			}
		} else {
			d.c += uint64(b1) << 9
			d.ct = 7
		}
	} else {
		b, err = r.ReadByte()
		if err != nil {
			return err
		}
		d.b = int(b)

		d.c += uint64(d.b) << 8
		d.ct = 8
	}

	d.c &= 0xFFFFFFFFFF

	return nil
}

func (d *Decoder) DecodeInt(r *reader.Reader, stats *DecoderStats) (int, bool, error) {
	var (
		value, bit, s, bitsToRead, offset int
		err                               error
	)

	d.previous = 1

	// first bit defines the sign of the integer
	s, err = d.decodeIntBit(r, stats)
	if err != nil {
		return 0, false, err
	}

	// common.Log.Debug("Sign int bit: '%01b'", s)

	bit, err = d.decodeIntBit(r, stats)
	if err != nil {
		return 0, false, err
	}

	// common.Log.Debug("First bit value: %b", bit)
	// First read
	if bit == 1 {
		bit, err = d.decodeIntBit(r, stats)
		if err != nil {
			return 0, false, err
		}

		// Second Read
		if bit == 1 {

			bit, err = d.decodeIntBit(r, stats)
			if err != nil {
				return 0, false, err
			}

			// Third Read
			if bit == 1 {

				bit, err = d.decodeIntBit(r, stats)
				if err != nil {
					return 0, false, err
				}

				// Fourth Read
				if bit == 1 {

					bit, err = d.decodeIntBit(r, stats)
					if err != nil {
						return 0, false, err
					}
					// Fifth Read
					if bit != 1 {

						bitsToRead = 32
						offset = 4436

					} else {
						// Fifth Read

						bitsToRead = 12
						offset = 340
					}
				} else {
					// Fourth Read
					bitsToRead = 8
					offset = 84

				}
			} else {
				// Third Read
				bitsToRead = 6
				offset = 20

			}
		} else {
			// SecondRead
			bitsToRead = 4
			offset = 4

		}
	} else {
		// First read
		bitsToRead = 2
		offset = 0

		// common.Log.Debug("Read First value: %v ", value)
	}

	for i := 0; i < bitsToRead; i++ {
		bit, err = d.decodeIntBit(r, stats)
		if err != nil {
			return 0, false, err
		}
		value = (value << 1) | bit
	}
	value += offset

	common.Log.Debug("Value decoded: %v with sign: %b", value, s)

	if s == 0 {

		return int(value), true, nil
	} else if s == 1 && value > 0 {

		return int(-value), true, nil
	}

	return int(0), false, nil
}

func (d *Decoder) DecodeBit(r *reader.Reader, stats *DecoderStats) (int, error) {
	var (
		bit     int
		qeValue uint32 = qe[stats.cx()][0]
		icx            = int(stats.cx())
	)

	defer func() {
		d.prvCtr += 1
		// common.Log.Debug("Decoder 'a' value: %b", d.a)
		// common.Log.Debug("%d, D: %01b C: %08X A: %04X, CTR: %d, B: %02X  QE: %04X", d.prvCtr, bit, d.c, d.a, d.ct, d.b, qeValue)
	}()

	d.a -= qeValue

	if (d.c >> 16) < uint64(qeValue) {
		bit = d.lpsExchange(stats, icx, qeValue)

		if err := d.renormalize(r); err != nil {
			return 0, err
		}
	} else {
		d.c -= (uint64(qeValue) << 16)

		if (d.a & 0x8000) == 0 {
			bit = d.mpsExchange(stats, icx)
			if err := d.renormalize(r); err != nil {
				return 0, err
			}
		} else {
			return int(stats.cx()), nil
		}
	}

	return bit, nil
}

func (d *Decoder) decodeIntBit(r *reader.Reader, stats *DecoderStats) (int, error) {

	stats.SetIndex(int(d.previous))
	bit, err := d.DecodeBit(r, stats)
	if err != nil {
		common.Log.Debug("ArithmeticDecoder 'decodeIntBit'-> DecodeBit failed. %v", err)
		return bit, err
	}

	common.Log.Debug("bit: %1b", bit)
	// common.Log.Debug("'previous' before: %b", d.previous)

	// if prev < 256
	if d.previous < 256 {

		d.previous = ((d.previous << uint(1)) | int64(bit)) & 0x1FF
	} else {

		d.previous = (((d.previous<<uint(1) | int64(bit)) & 511) | 256) & 0x1FF
	}

	return bit, nil
}

func (d *Decoder) DecodeIAID(r *reader.Reader, codeLen uint64, stats *DecoderStats) (int64, error) {
	d.previous = 1
	stats.SetIndex(int(d.previous))
	var i uint64
	for i = 0; i < codeLen; i++ {
		bit, err := d.DecodeBit(r, stats)
		if err != nil {
			return 0, err
		}

		d.previous = ((d.previous << 1) | int64(bit))
	}
	return d.previous - (1 << codeLen), nil
}

func (d *Decoder) renormalize(r *reader.Reader) error {
fl:
	for {
		if d.ct == 0 {
			if err := d.readByte(r); err != nil {
				return err
			}
		}

		d.a <<= 1
		d.c <<= 1
		d.ct -= 1

		if (d.a & 0x8000) != 0 {
			break fl
		}
	}

	d.c &= 0xFFFFFFFF
	return nil
}

func (d *Decoder) mpsExchange(stats *DecoderStats, icx int) int {
	mps := stats.mps[stats.index]

	if d.a < qe[icx][0] {
		if qe[icx][3] == 1 {
			stats.toggleMps()
		}

		stats.SetEntry(int(qe[icx][2]))
		return int(1 - mps)
	} else {
		stats.SetEntry(int(qe[icx][1]))
		return int(mps)
	}
}

func (d *Decoder) lpsExchange(stats *DecoderStats, icx int, qeValue uint32) int {
	mps := stats.mps[stats.index]

	if d.a < qeValue {
		stats.SetEntry(int(qe[icx][1]))
		d.a = qeValue
		return int(mps)
	} else {

		if qe[icx][3] == 1 {
			stats.toggleMps()
		}

		stats.SetEntry(int(qe[icx][2]))
		d.a = qeValue
		return int(1 - mps)
	}
}
