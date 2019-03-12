package bitset

import (
	"errors"
	"fmt"
	"github.com/unidoc/unidoc/common"
	"math/bits"
)

const (
	uint64Bits int = 64
)

var (
	ErrIndexOutOfRange error = errors.New("Index out of range")
)

// BitSet is the fast set for the binary Data
// The Data is set into a slice of uint64 where each holds 64 bit Data
// The set and get operations use bitwise operaitons
type BitSet struct {
	Data []uint64

	// length is the number of bits used in the
	Length int
}

// NewBitSet creates new bitset of the provided length
func NewBitSet(length int) *BitSet {
	common.Log.Debug("Creating Bitset of length: %d", length)
	b := &BitSet{
		Length: length,
	}

	wcount := length / uint64Bits

	if (length % uint64Bits) != 0 {
		wcount += 1
	}
	// common.Log.Debug("Length: %d, WCount: %d", length, wcount)

	b.Data = make([]uint64, wcount)
	return b
}

// Checks if the values are equal
func (b *BitSet) Equals(s *BitSet) bool {
	if b.Length != s.Length {
		return false
	}

	for i, Data := range b.Data {
		if s.Data[i] != Data {
			return false
		}
	}
	return true
}

// Get gets the bit value at the 'index'
func (b *BitSet) Get(index uint) (bool, error) {
	var wIndex uint = index >> 6
	if wIndex > uint(len(b.Data)-1) {
		return false, ErrIndexOutOfRange
	}
	// get the value from [wIndex] uint64 at index 2^index

	value := ((b.Data[wIndex] & (uint64(1) << (index % 64))) != 0)

	return value, nil
}

// GetByte gets the byte at the 'bitIndex' position
// if the index is out of range returns error
func (b *BitSet) GetByte(bitIndex uint) (byte, error) {
	var wIndex uint = bitIndex >> 6
	if wIndex > uint(len(b.Data)-1) {
		return 0, ErrIndexOutOfRange
	}

	iIndex := bitIndex % 64

	value := byte(b.Data[wIndex] >> (iIndex))

	overflown := (iIndex + 8)
	if overflown > 64 {

		if wIndex+1 > uint(len(b.Data)-1) {
			return 0, ErrIndexOutOfRange
		}

		// get from the wIndex+1
		v2 := byte(b.Data[wIndex+1])

		v2 <<= (overflown % 64)

		value |= v2
	}

	return value, nil

}

// valueLength should be in range from 0 to 8
// i.e. the valueLength = 5 means that the byte 00011011 have 5 bits -> 11011
// bitOffset is the offset the the bitSet
func (b *BitSet) SetByteBitwiseOffset(bt byte, valueLength, bitOffset int, reverse bool) error {
	// common.Log.Debug("Setting byte at bitOffset: %d", bitOffset)
	if bitOffset > b.Length {
		return ErrIndexOutOfRange
	}

	if valueLength == 0 {
		return errors.New("No value provided")
	}

	if reverse {
		bt = byte(bits.Reverse8(uint8(bt)))

		common.Log.Debug("Reversed: %08b", bt)
	}

	// wIndex is the index of the Data within the b.Data
	wIndex := uint(bitOffset) >> 6

	// get the bitOffset
	// i index is the index within the uint64 Data
	iIndex := (bitOffset % 64)

	// apply the mask
	mask := byte(1<<byte(valueLength+1)) - 1
	btMasked := bt & mask

	// common.Log.Debug("Mask: %08b, btMasked: %08b", mask, btMasked)

	Data := b.Data[wIndex]

	// common.Log.Debug("Data before: %064b", Data)
	Data |= uint64(uint64(btMasked) << uint(iIndex))

	// common.Log.Debug("Data after: %064b", Data)

	b.Data[wIndex] = Data

	// if bitOffset + vlength overflows index within single int64
	overflown := iIndex + valueLength

	// i.e. bitOffset: 125 % 64 => 61 valueLength = 5 then overflown would be 66
	// last bits should be parsed to the next uint64

	if overflown > 64 {
		nData := b.Data[wIndex+1]

		// the shiftSize can't be greater than 8
		shiftSize := uint(overflown % 64)

		nData |= uint64(bt >> shiftSize)
		b.Data[wIndex+1] = nData
	}

	return nil
}

// SetByteOffset sets the byte at the provided offset
func (b *BitSet) SetByteOffset(bt byte, byteOffset int) error {
	// Check if it is within the length
	if (byteOffset*8 + 8) > b.Length {
		return ErrIndexOutOfRange
	}

	// DataIndex is the index of the int64 in the Data
	DataIndex := byteOffset / 8 // 15 / 8 = 1

	// byteDataIndex is the index within the provided int64 Data
	byteDataIndex := byteOffset % 8 // 15 % 8 = 7

	// get the Data
	Data := b.Data[DataIndex]

	common.Log.Debug("Data before shifting: %08b", bt)

	// shift the
	Data |= (uint64(bt) << uint(byteDataIndex*8))

	common.Log.Debug("Setting Data: %064b at index: %d", Data, DataIndex)

	b.Data[DataIndex] = Data

	return nil
}

// Or makes a union of two sets saving the result into the 'b' Bitset
// Arguments
// - startIndex - index where union begins in the 'b' set
// - setStartIndex - index where union begins in the 'set' set
// - set - the bitset that is being unioned with the 'b' set
// - length - the length of the bitset
func (b *BitSet) Or(startIndex uint, set *BitSet, setStartIndex uint, length int) {

	var shift uint = startIndex - setStartIndex

	var k uint64 = set.Data[setStartIndex>>6]

	k = (k << shift) | (k >> uint(64-shift))

	if (setStartIndex&63)+uint(length) <= 64 {
		setStartIndex += shift

		for i := 0; i < length; i++ {
			index := startIndex >> 6

			b.Data[index] |= (k & (uint64(1) << setStartIndex))
			setStartIndex++
			startIndex++
		}
	} else {
		for i := 0; i < length; i++ {
			if (setStartIndex & 63) == 0 {

				k = set.Data[(setStartIndex)>>6]
				k = (k << shift) | (k >> (64 - shift))
			}
			b.Data[(startIndex)>>6] |= k & (uint64(1) << (setStartIndex + shift))
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
		for i := 0; i < len(b.Data); i++ {
			b.Data[i] = max
		}
	} else {
		for i := 0; i < len(b.Data); i++ {
			b.Data[i] = 0
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
	if wIndex > uint(len(b.Data)-1) {
		return ErrIndexOutOfRange
	}

	rIndex := index % 64

	// common.Log.Debug("Index: %v, WIndex: %v, RIndex: %v", index, wIndex, rIndex)

	if value {
		b.Data[wIndex] |= (uint64(1) << rIndex)
	} else {
		b.Data[wIndex] &= ^(uint64(1) << rIndex)
	}

	// common.Log.Debug("WIndex: %v, %064b, %v", wIndex, b.Data[wIndex], value)

	return nil
}

// Size returns the BitSet length
func (b *BitSet) Size() int {
	return b.Length
}

func (b *BitSet) String() string {
	var val string
	for i := 0; i < b.Length; i++ {
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

// Bytes returns the Data representation as the byte stream
func (b *BitSet) Bytes() []byte {

	byteLen := (b.Length / 8) + 1

	common.Log.Debug("Bytes len: %v", byteLen)
	var Data []byte = make([]byte, byteLen)

	for i := 0; i < len(b.Data); i++ {
		// Data length is the
		for j := 0; j < 8; j++ {

			if (i*8 + j) >= byteLen {
				break
			}
			v := uint8(b.Data[i] >> uint(j*8))
			common.Log.Debug("Setting Data: %08b at index: %v", v, i*8+j)
			Data[i*8+j] = v
		}
	}

	return Data
}
