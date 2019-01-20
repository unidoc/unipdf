package bitmap

import (
	"errors"
)

const (
	uint64Bits int = 64
)

var (
	ErrIndexOutOfRange error = errors.New("Index out of range")
)

// BitSet is the fast set for the binary data
// The data is set into a slice of uint64 where each holds
// 64 bit data
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

func (b *BitSet) Or(startIndex uint, set *BitSet, setStartIndex uint, length int) {
	var shift uint = startIndex - setStartIndex

	var k uint64 = set.data[setStartIndex>>6]

	k = (k << shift) | (k >> uint(64-shift))

	if (setStartIndex&63)+uint(length) <= 64 {
		setStartIndex += shift

		for i := 0; i < length; i++ {
			b.data[(startIndex)>>6] |= k & (uint64(1) << setStartIndex)
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
