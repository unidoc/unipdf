/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package endian

import (
	"encoding/binary"
	"unsafe"
)

var (
	// ByteOrder is the current system byte order.
	ByteOrder   binary.ByteOrder
	isBigEndian bool
)

func init() {
	const intSize = int(unsafe.Sizeof(0))
	i := 1
	byteSlice := (*[intSize]byte)(unsafe.Pointer(&i))
	if byteSlice[0] == 0 {
		isBigEndian = true
		ByteOrder = binary.BigEndian
	} else {
		ByteOrder = binary.LittleEndian
	}
}

// IsBig checks if the machine uses the Big Endian byte order.
func IsBig() bool {
	return isBigEndian
}

// IsLittle checks if the machine uses Little Endian byte ordering.
func IsLittle() bool {
	return !isBigEndian
}
