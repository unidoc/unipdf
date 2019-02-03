package bitmap

import (
	"errors"
	"fmt"
	"github.com/unidoc/unidoc/common"
)

const (
	uint64Bits int = 64
)

var (
	ErrIndexOutOfRange error = errors.New("Index out of range")
)

// BitSet is the fast set for the binary data
// The data is set into a slice of uint64 where each holds 64 bit data
// The set and get operations use bitwise operaitons
type BitSet struct {
	data []uint64

	// length is the number of bits used in the
	length int
}

func NewBitSet(length int) *BitSet {
	b := &BitSet{
		length: length,
	}

	wcount := length / uint64Bits

	if (length % uint64Bits) != 0 {
		wcount += 1
	}
	// common.Log.Debug("Length: %d, WCount: %d", length, wcount)

	b.data = make([]uint64, wcount)
	return b
}

func (b *BitSet) Equals(s *BitSet) bool {
	if b.length != s.length {
		return false
	}

	for i, data := range b.data {
		if s.data[i] != data {
			return false
		}
	}
	return true
}

// Get gets the bit value at the 'index'
func (b *BitSet) Get(index uint) (bool, error) {
	var wIndex uint = index >> 6
	if wIndex > uint(len(b.data)-1) {
		return false, ErrIndexOutOfRange
	}
	// get the value from [wIndex] uint64 at index 2^index

	value := ((b.data[wIndex] & (uint64(1) << (index % 64))) != 0)

	return value, nil
}

// Or makes a union of two sets saving the result into the 'b' Bitset
// Arguments
// - startIndex - index where union begins in the 'b' set
// - setStartIndex - index where union begins in the 'set' set
// - set - the bitset that is being unioned with the 'b' set
// - length - the length of the bitset
func (b *BitSet) Or(startIndex uint, set *BitSet, setStartIndex uint, length int) {

	var shift uint = startIndex - setStartIndex

	var k uint64 = set.data[setStartIndex>>6]

	k = (k << shift) | (k >> uint(64-shift))

	if (setStartIndex&63)+uint(length) <= 64 {
		setStartIndex += shift

		for i := 0; i < length; i++ {
			index := startIndex >> 6

			b.data[index] |= (k & (uint64(1) << setStartIndex))
			setStartIndex++
			startIndex++
		}
	} else {
		for i := 0; i < length; i++ {
			if (setStartIndex & 63) == 0 {

				k = set.data[(setStartIndex)>>6]
				k = (k << shift) | (k >> (64 - shift))
			}
			b.data[(startIndex)>>6] |= k & (uint64(1) << (setStartIndex + shift))
			setStartIndex++
			startIndex++
		}
	}

}

// SetAll sets all the values in the BitSet
// to '1' if value is true and '0' if  value is false
func (b *BitSet) SetAll(value bool) {
	if value {
		var max uint64
		max -= 1
		for i := 0; i < len(b.data); i++ {
			b.data[i] = max
		}
	} else {
		for i := 0; i < len(b.data); i++ {
			b.data[i] = 0
		}
	}
}

// SetAtRange sets the value in the bit set in the range
// starting from 'start' up to 'end'
func (b *BitSet) SetAtRange(start, end uint, value bool) error {
	for i := start; i < end; i++ {
		if err := b.Set(i, value); err != nil {
			return err
		}
	}
	return nil
}

// Set sets the values
func (b *BitSet) Set(index uint, value bool) error {
	var wIndex uint = uint(index) / 64
	if wIndex > uint(len(b.data)-1) {
		return ErrIndexOutOfRange
	}

	rIndex := index % 64

	// common.Log.Debug("Index: %v, WIndex: %v, RIndex: %v", index, wIndex, rIndex)

	if value {
		b.data[wIndex] |= (uint64(1) << rIndex)
	} else {
		b.data[wIndex] &= ^(uint64(1) << rIndex)
	}

	// common.Log.Debug("WIndex: %v, %064b, %v", wIndex, b.data[wIndex], value)

	return nil
}

// Size returns the BitSet length
func (b *BitSet) Size() int {
	return b.length
}

func (b *BitSet) String() string {
	var val string
	for i := 0; i < b.length; i++ {
		bv, err := b.Get(uint(i))
		if err != nil {
			panic(err)
		}
		if bv {
			if val != "" {
				val = fmt.Sprintf("1%s", val)
			} else {
				val = "1"
			}
		} else {
			if val != "" {
				val = fmt.Sprintf("0%s", val)
			} else {
				val = "0"
			}
		}
	}
	return val
}

// Bytes returns the data representation as the byte stream
func (b *BitSet) Bytes() []byte {

	byteLen := (b.length / 8) + 1

	common.Log.Debug("Bytes len: %v", byteLen)
	var data []byte = make([]byte, byteLen)

	for i := 0; i < len(b.data); i++ {
		// data length is the
		for j := 0; j < 8; j++ {

			if (i*8 + j) >= byteLen {
				break
			}
			v := uint8(b.data[i] >> uint(j*8))
			common.Log.Debug("Setting data: %08b at index: %v", v, i*8+j)
			data[i*8+j] = v
		}
	}

	return data
}

// 	fmt.Println("Hello, playground")

// 	var b uint64 = 2<<63 - 1

// 	b &= ^(uint64(1) << 26)
// 	b &= ^(uint64(1) << 35)
// 	b &= ^(uint64(1) << 42)

// 	fmt.Printf("%64b\n", b)

// 	for i := 1; i < 64/8; i++ {
// 		fmt.Printf("%08b\n", uint8(b>>uint(i*8)))
// 	}
// }
