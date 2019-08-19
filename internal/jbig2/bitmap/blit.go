/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

// Blit blits the source Bitmap 'src' into Destination bitmap: 'dst' on the provided 'x' and 'y' coordinates
// with respect to the combination operator 'op'.
func Blit(src *Bitmap, dst *Bitmap, x, y int32, op CombinationOperator) error {
	var startLine, srcStartIdx int32
	srcEndIdx := src.RowStride - 1

	// ignore those parts of source bitmap placed outside target bitmap.
	if x < 0 {
		srcStartIdx = -x
		x = 0
	} else if x+src.Width > dst.Width {
		srcEndIdx -= src.Width + x - dst.Width
	}

	if y < 0 {
		startLine = -y
		y = 0
		srcStartIdx += src.RowStride
		srcEndIdx += src.RowStride
	} else if y+src.Height > dst.Height {
		startLine = src.Height + y - dst.Height
	}

	var (
		lastLine int32
		err      error
	)
	shiftVal1 := uint8(x & 0x07)
	shiftVal2 := 8 - shiftVal1
	padding := uint8(src.Width & 0x07)
	toShift := shiftVal2 - padding
	useShift := shiftVal2&0x07 != 0
	specialCase := src.Width <= ((srcEndIdx-srcStartIdx)<<3)+int32(shiftVal2)
	dstStartIdx := dst.GetByteIndex(x, y)

	// get math.min()
	temp := startLine + dst.Height
	if src.Height > temp {
		lastLine = temp
	} else {
		lastLine = src.Height
	}

	switch {
	case !useShift:
		err = blitUnshifted(src, dst, startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx, op)
	case specialCase:
		err = blitSpecialShifted(src, dst, startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx, toShift, shiftVal1, shiftVal2, op)
	default:
		err = blitShifted(src, dst, startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx, toShift, shiftVal1, shiftVal2, op, padding)
	}
	return err
}

func blitUnshifted(
	src, dst *Bitmap,
	startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx int32,
	op CombinationOperator,
) error {
	var dstLine int32
	increaser := func() {
		dstLine++
		dstStartIdx += dst.RowStride
		srcStartIdx += src.RowStride
		srcEndIdx += src.RowStride
	}

	for dstLine = startLine; dstLine < lastLine; increaser() {
		dstIdx := dstStartIdx
		for srcIdx := srcStartIdx; srcIdx <= srcEndIdx; srcIdx++ {
			oldByte, err := dst.GetByte(dstIdx)
			if err != nil {
				return err
			}

			newByte, err := src.GetByte(srcIdx)
			if err != nil {
				return err
			}

			if err = dst.SetByte(dstIdx, combineBytes(oldByte, newByte, op)); err != nil {
				return err
			}
			dstIdx++
		}
	}
	return nil
}

func blitShifted(
	src, dst *Bitmap,
	startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx int32,
	toShift, shiftVal1, shiftVal2 uint8,
	op CombinationOperator, padding uint8,
) error {
	var dstLine int32
	increaser := func() {
		dstLine++
		dstStartIdx += dst.RowStride
		srcStartIdx += src.RowStride
		srcEndIdx += src.RowStride
	}

	for dstLine = startLine; dstLine < lastLine; increaser() {
		var register uint16
		dstIdx := dstStartIdx

		for srcIdx := srcStartIdx; srcIdx <= srcEndIdx; srcIdx++ {
			oldByte, err := dst.GetByte(dstIdx)
			if err != nil {
				return err
			}

			newByte, err := src.GetByte(srcIdx)
			if err != nil {
				return err
			}
			register = (register | (uint16(newByte) & 0xff)) << uint(shiftVal2)

			newByte = byte(register >> 8)
			dst.SetByte(dstIdx, combineBytes(oldByte, newByte, op))
			dstIdx++

			register <<= uint(shiftVal1)

			if srcIdx == srcEndIdx {
				newByte = byte(register >> (8 - uint8(shiftVal2)))

				if padding != 0 {
					newByte = unpad(8+toShift, newByte)
				}

				oldByte, err = dst.GetByte(dstIdx)
				if err != nil {
					return err
				}

				if err = dst.SetByte(dstIdx, combineBytes(oldByte, newByte, op)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func blitSpecialShifted(
	src, dst *Bitmap,
	startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx int32,
	toShift, shiftVal1, shiftVal2 uint8,
	op CombinationOperator,
) error {
	var dstLine int32
	increaser := func() {
		dstLine++
		dstStartIdx += dst.RowStride
		srcStartIdx += src.RowStride
		srcEndIdx += src.RowStride
	}

	for dstLine = startLine; dstLine < lastLine; increaser() {
		var register uint16
		dstIdx := dstStartIdx

		for srcIdx := srcStartIdx; srcIdx <= srcEndIdx; srcIdx++ {
			oldByte, err := dst.GetByte(dstIdx)
			if err != nil {
				return err
			}

			newByte, err := src.GetByte(srcIdx)
			if err != nil {
				return err
			}
			register = (register | uint16(newByte)) << uint(shiftVal2)
			newByte = byte(register >> 8)

			if srcIdx == srcEndIdx {
				newByte = unpad(toShift, newByte)
			}

			if err = dst.SetByte(dstIdx, combineBytes(oldByte, newByte, op)); err != nil {
				return err
			}
			dstIdx++

			register <<= uint(shiftVal1)
		}
	}
	return nil
}

func unpad(padding uint8, b byte) byte {
	return b >> padding << padding
}
