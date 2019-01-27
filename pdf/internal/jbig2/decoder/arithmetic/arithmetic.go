package arithmetic

import (
	"github.com/unidoc/unidoc/common"
	"io"
)

var (
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

	Buffer0 int64
	Buffer1 int64

	c, a     int64
	previous int64

	counter int
}

// New creates new arithmetic Decoder
func New() *Decoder {
	d := &Decoder{
		GenericRegionStats:    NewStats(1 << 1),
		RefinementRegionStats: NewStats(1 << 1),

		IadhStats:            NewStats(int(uint(1) << uint(9))),
		IadwStats:            NewStats(int(uint(1) << uint(9))),
		IaexStats:            NewStats(int(uint(1) << uint(9))),
		IaaiStats:            NewStats(int(uint(1) << uint(9))),
		IadtStats:            NewStats(int(uint(1) << uint(9))),
		IaitStats:            NewStats(int(uint(1) << uint(9))),
		IafsStats:            NewStats(int(uint(1) << uint(9))),
		IadsStats:            NewStats(int(uint(1) << uint(9))),
		IardxStats:           NewStats(int(uint(1) << uint(9))),
		IardyStats:           NewStats(int(uint(1) << uint(9))),
		IardwStats:           NewStats(int(uint(1) << uint(9))),
		IardhStats:           NewStats(int(uint(1) << uint(9))),
		IariStats:            NewStats(int(uint(1) << uint(9))),
		IaidStats:            NewStats(int(uint(1) << uint(1))),
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
		d.IaidStats = NewStats(1 << uint(symbolCodeLength+1))
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
			d.GenericRegionStats = NewStats(1 << uint(size))
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
			d.RefinementRegionStats = NewStats(1 << uint(size))
		}
	}
}

func (d *Decoder) Start(r io.ByteReader) error {
	b, err := r.ReadByte()
	if err != nil {
		common.Log.Debug("Buffer0 readByte failed. %v", err)
		return err
	}
	d.Buffer0 = int64(b)
	// common.Log.Debug("Decoder Buffer0 init value: %b, %02x", d.Buffer0, d.Buffer0)
	b, err = r.ReadByte()
	if err != nil {
		common.Log.Debug("Buffer0 readByte failed. %v", err)
		return err
	}

	d.Buffer1 = int64(b)
	// common.Log.Debug("Decoder Buffer1 init value: %b, %02x", d.Buffer1, d.Buffer1)
	// the values of Buffer0 and Buffer1 are set

	// Shift buffer0
	d.c = ((d.Buffer0 ^ 0xFF) << 16)

	// Decoder.readByte()
	if err = d.readByte(r); err != nil {
		return err
	}

	// Shift c value again
	d.c <<= 7
	// common.Log.Debug("Decoder 'c' value: %b", d.c)

	// set
	d.counter -= 7

	// common.Log.Debug("Decoder Counter: %v", d.counter)

	d.a = 0x800000001

	// common.Log.Debug("Decoder 'a' value: %b", d.a)

	return nil
}

func (d *Decoder) readByte(r io.ByteReader) error {
	if d.Buffer0 == 0xFF {
		if d.Buffer1 > 0x8F {
			d.counter = 8
		} else {
			d.Buffer0 = d.Buffer1
			b, err := r.ReadByte()
			if err != nil {
				return err
			}
			d.Buffer1 = int64(b)
			d.c = d.c + 0xfe00 - int64(d.Buffer0<<9)
			d.counter = 7
		}
	} else {
		d.Buffer0 = d.Buffer1
		b, err := r.ReadByte()
		if err != nil {
			return err
		}

		d.Buffer1 = int64(b)

		d.c = d.c + 0xff00 - int64(d.Buffer0<<8)
		d.counter = 8
	}
	return nil
}

func (d *Decoder) DecodeInt(r io.ByteReader, stats *DecoderStats) (int, bool, error) {
	var value int32

	d.previous = 1

	// first bit defines the sign of the integer
	s, err := d.decodeIntBit(r, stats)
	if err != nil {
		return 0, false, err
	}
	// common.Log.Debug("Sign int bit: '%01b'", s)

	bit, err := d.decodeIntBit(r, stats)
	if err != nil {
		return 0, false, err
	}

	// common.Log.Debug("First bit value: %b", bit)
	// First read
	if bit != 0 {

		bit, err := d.decodeIntBit(r, stats)
		if err != nil {
			return 0, false, err
		}
		// Second Read
		if bit != 0 {

			bit, err := d.decodeIntBit(r, stats)
			if err != nil {
				return 0, false, err
			}
			// Third Read
			if bit != 0 {

				bit, err := d.decodeIntBit(r, stats)
				if err != nil {
					return 0, false, err
				}
				// Fourth Read
				if bit != 0 {

					bit, err := d.decodeIntBit(r, stats)
					if err != nil {
						return 0, false, err
					}
					// Fifth Read
					if bit != 0 {
						value = 0

						for i := 0; i < 32; i++ {
							bit, err := d.decodeIntBit(r, stats)
							if err != nil {
								return 0, false, err
							}

							value = ((value << 1) | int32(bit))
						}

						value += 4436
						// common.Log.Debug("Read sixth value: %v ", value)

					} else {
						// Fifth Read

						value = 0
						for i := 0; i < 12; i++ {
							bit, err := d.decodeIntBit(r, stats)
							if err != nil {
								return 0, false, err
							}

							value = ((value << 1) | int32(bit))
						}

						value += 340

						// common.Log.Debug("Read fifth value: %v ", value)
					}
				} else {
					// Fourth Read
					value = 0
					for i := 0; i < 8; i++ {
						bit, err := d.decodeIntBit(r, stats)
						if err != nil {
							return 0, false, err
						}

						value = ((value << 1) | int32(bit))
					}
					value += 84

					// common.Log.Debug("Read fourth value: %v ", value)
				}
			} else {
				// Third Read
				value = 0
				for i := 0; i < 6; i++ {
					bit, err := d.decodeIntBit(r, stats)
					if err != nil {
						return 0, false, err
					}

					value = ((value << 1) | int32(bit))
				}

				value += 20

				// common.Log.Debug("Read third value: %v ", value)
			}
		} else {
			// SecondRead
			bit, err := d.decodeIntBit(r, stats)
			if err != nil {
				return 0, false, err
			}

			value = int32(bit)

			for i := 0; i < 3; i++ {
				bit, err := d.decodeIntBit(r, stats)
				if err != nil {
					return 0, false, err
				}
				value = ((value << 1) | int32(bit))
			}

			value += 4

			// common.Log.Debug("Read Second value: %v ", value)
		}
	} else {
		// First read
		bit, err := d.decodeIntBit(r, stats)
		if err != nil {
			return 0, false, err
		}
		value = int32(bit)

		bit, err = d.decodeIntBit(r, stats)
		if err != nil {
			return 0, false, err
		}
		value = ((value << 1) | int32(bit))

		// common.Log.Debug("Read First value: %v ", value)
	}

	var decodedInt int
	if s != 0 {
		if value == 0 {
			return int(value), false, nil
		}
		decodedInt = -(int(value))
	} else {
		decodedInt = int(value)
	}

	return decodedInt, true, nil
}

func (d *Decoder) DecodeBit(r io.ByteReader, context int64, stats *DecoderStats) (int, error) {
	var iCX int = (stats.codingContextTable[context] >> 1) & 0xFF
	var mpsCX int = stats.codingContextTable[context] & 1
	var qe int = qeTable[iCX]

	common.Log.Debug("mpsCX: %v, iCX: %v, qe: %v, context: %v", mpsCX, iCX, qe, context)

	d.a -= int64(qe)

	var bit int

	if d.c < d.a {

		// 0x80000000 is OOB vvalue
		if (d.a & 0x80000000) != 0 {
			bit = mpsCX
			// common.Log.Debug("Bit set to mpsCX: %b", bit)
		} else {

			if d.a < int64(qe) {
				bit = 1 - mpsCX
				// common.Log.Debug("Bit set to '1 - mpsCX': %b", bit)
				if switchTable[iCX] != 0 {
					stats.codingContextTable[context] = ((nlpsTable[iCX] << 1) | (1 - mpsCX))
				} else {
					stats.codingContextTable[context] = ((nlpsTable[iCX] << 1) | mpsCX)
				}
			} else {
				bit = mpsCX

				stats.codingContextTable[context] = ((nmpsTable[iCX] << 1) | mpsCX)
			}

			// common.Log.Debug("Bit set to mpsCX: %b start looping ", bit)

			for {
				if d.counter == 0 {
					if err := d.readByte(r); err != nil {
						return 0, err
					}
				}
				d.a <<= 1
				d.c <<= 1

				d.counter -= 1
				if !((d.a & 0x80000000) == 0) {
					break
				}
			}
		}
	} else {
		d.c -= d.a

		if d.a < int64(qe) {
			bit = mpsCX
			// common.Log.Debug("Bit set to mpsCX: %b", bit)
			stats.codingContextTable[context] = ((nmpsTable[iCX] << 1) | mpsCX)
		} else {

			bit = 1 - mpsCX
			// common.Log.Debug("Bit set to 1 - mpsCX: %b", bit)
			if switchTable[iCX] != 0 {
				stats.codingContextTable[context] = ((nlpsTable[iCX] << 1) | (1 - mpsCX))
			} else {
				stats.codingContextTable[context] = ((nlpsTable[iCX] << 1) | mpsCX)
			}
		}
		d.a = int64(qe)

		for {
			if d.counter == 0 {
				if err := d.readByte(r); err != nil {
					return 0, err
				}
			}

			d.a <<= 1
			d.c <<= 1

			d.counter -= 1

			if !((d.a & 0x80000000) == 0) {
				break
			}
		}
	}
	return bit, nil
}

func (d *Decoder) decodeIntBit(r io.ByteReader, stats *DecoderStats) (int, error) {
	// get bit
	// common.Log.Debug("[decodeIntBit] starts")
	// defer func() { common.Log.Debug("[decodeIntBit] finished") }()
	bit, err := d.DecodeBit(r, d.previous, stats)
	if err != nil {
		common.Log.Debug("ArithmeticDecoder 'decodeIntBit'-> DecodeBit failed. %v", err)
		return bit, err
	}
	common.Log.Debug("bit: %1b", bit)
	// common.Log.Debug("'previous' before: %b", d.previous)

	// if prev < 256
	if d.previous < 0x100 {
		d.previous = (d.previous << uint(1)) | int64(bit)
	} else {
		d.previous = ((d.previous<<uint(1) | int64(bit)) & 0x1ff) | 0x100
	}

	// common.Log.Debug("'previous' after: %b", d.previous)

	return bit, nil
}

func (d *Decoder) DecodeIAID(r io.ByteReader, codeLen uint64, stats *DecoderStats) (int64, error) {
	d.previous = 1

	var i uint64
	for i = 0; i < codeLen; i++ {
		bit, err := d.DecodeBit(r, d.previous, stats)
		if err != nil {
			return 0, err
		}

		d.previous = ((d.previous << 1) | int64(bit))
	}
	return d.previous - (1 << codeLen), nil
}
