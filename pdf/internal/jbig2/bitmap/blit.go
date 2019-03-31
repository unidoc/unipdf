package bitmap

import (
	"github.com/unidoc/unidoc/common"
)

func Blit(src *Bitmap, dst *Bitmap, x, y int, op CombinationOperator) (err error) {

	common.Log.Debug("Blitting bitmaps with operator: %s, at X: %d, Y: %d", op, x, y)
	common.Log.Debug("Source bitmap: %s", src)
	common.Log.Debug("Dst bitmap: %s", dst)

	var startLine, srcStartIdx, srcEndIdx int
	srcEndIdx = src.RowStride - 1

	// ignore those parts of source bitmap which would be placed outside target bitmap
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
		shiftVal1   int  = x & 0x07
		shiftVal2   int  = 8 - shiftVal1
		padding     int  = src.Width & 0x07
		toShift     int  = shiftVal2 - padding
		useShift    bool = shiftVal2&0x07 != 0
		specialCase bool = src.Width <= ((srcEndIdx-srcStartIdx)<<3)+shiftVal2
		dstStartIdx int  = dst.GetByteIndex(x, y)
		lastLine    int
	)

	// get math.min()
	temp := startLine + dst.Height
	if src.Height > temp {
		lastLine = temp
	} else {
		lastLine = src.Height
	}

	if !useShift {
		err = blitUnshifted(src, dst, startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx, op)
	} else if specialCase {
		err = blitSpecialShifted(src, dst, startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx, toShift, shiftVal1, shiftVal2, op)
	} else {
		err = blitShifted(src, dst, startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx, toShift, shiftVal1, shiftVal2, op, padding)
	}

	return
}

func blitUnshifted(
	src, dst *Bitmap,
	startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx int,
	op CombinationOperator,
) error {
	common.Log.Debug("Blit unshifted")
	var dstLine int

	increaser := func() {
		dstLine += 1
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
			dstIdx += 1
		}
	}

	return nil
}

func blitShifted(
	src, dst *Bitmap,
	startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx, toShift, shiftVal1, shiftVal2 int,
	op CombinationOperator, padding int,
) error {
	common.Log.Debug("Blit shifted")
	var dstLine int

	// increaser increases the for loop values
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
			dst.SetByte(dstIdx, combineBytes(oldByte, newByte, op))
			dstIdx++

			register <<= uint(shiftVal1)

			if srcIdx == srcEndIdx {

				newByte = byte(register >> (8 - uint8(shiftVal2)))

				if padding != 0 {
					newByte = unpad(uint(8+toShift), newByte)
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
	startLine, lastLine, dstStartIdx, srcStartIdx, srcEndIdx, toShift, shiftVal1, shiftVal2 int,
	op CombinationOperator,
) error {
	common.Log.Debug("Blit special shifted")
	var dstLine int
	increaser := func() {
		dstLine += 1
		dstStartIdx += dst.RowStride
		srcStartIdx += src.RowStride
		srcEndIdx += src.RowStride
	}

	for dstLine = startLine; dstLine < lastLine; increaser() {
		var register uint16 = 0
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
				newByte = unpad(uint(toShift), newByte)
			}

			if err = dst.SetByte(dstIdx, combineBytes(oldByte, newByte, op)); err != nil {
				return err
			}
			dstIdx += 1

			register <<= uint(shiftVal1)
		}
	}

	return nil
}

func unpad(padding uint, b byte) byte {
	return b >> padding << padding

}
