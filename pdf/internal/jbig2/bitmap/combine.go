package bitmap

import (
	"github.com/unidoc/unidoc/common"
	"image"
)

// CombineBytes combines the provided bytes with respect to the CombinationOperator
func CombineBytes(oldByte, newByte byte, op CombinationOperator) byte {
	return combineBytes(oldByte, newByte, op)
}

func combineBytes(oldByte, newByte byte, op CombinationOperator) byte {
	switch op {
	case CmbOpOr:
		return newByte | oldByte
	case CmbOpAnd:
		return newByte & oldByte
	case CmbOpXor:
		return newByte ^ oldByte
	case CmbOpXNor:
		return ^(newByte ^ oldByte)
	default:
		return newByte
	}
}

func Extract(roi image.Rectangle, src *Bitmap) (*Bitmap, error) {
	dst := New(roi.Dx(), roi.Dy())

	upShift := roi.Min.X & 0x07
	downShift := 8 - upShift

	var dstLineStartIdx int

	padding := uint(8 - dst.Width&0x07)

	srcLineStartIdx := src.GetByteIndex(roi.Min.X, roi.Min.Y)
	srcLineEndIdx := src.GetByteIndex(roi.Max.X-1, roi.Min.Y)

	common.Log.Debug("SrcLineStartIdx: %d", srcLineStartIdx)
	common.Log.Debug("SrcLineEndIdx: %d", srcLineEndIdx)
	usePadding := dst.RowStride == srcLineEndIdx+1-srcLineStartIdx

	for y := roi.Min.Y; y < roi.Max.Y; y++ {
		srcIdx := srcLineStartIdx
		dstIdx := dstLineStartIdx

		if srcLineStartIdx == srcLineEndIdx {

			pixels, err := src.GetByte(srcIdx)
			if err != nil {
				return nil, err
			}

			pixels <<= uint(upShift)
			common.Log.Debug("Pixels Byte: %08b", pixels)
			err = dst.SetByte(dstIdx, unpad(padding, pixels))
			if err != nil {
				return nil, err
			}
		} else if upShift == 0 {

			for x := srcLineStartIdx; x <= srcLineEndIdx; x++ {

				value, err := src.GetByte(srcIdx)
				if err != nil {
					return nil, err
				}
				srcIdx += 1

				common.Log.Debug("Value Byte: %08b", value)

				if x == srcLineEndIdx && usePadding {
					value = unpad(padding, value)
				}
				err = dst.SetByte(dstIdx, value)
				if err != nil {
					return nil, err
				}
				dstIdx += 1
			}

		} else {
			err := copyLine(src, dst, uint(upShift), uint(downShift), padding, srcLineStartIdx, srcLineEndIdx, usePadding, srcIdx, dstIdx)
			if err != nil {
				return nil, err
			}
		}
		srcLineStartIdx += src.RowStride
		srcLineEndIdx += src.RowStride
		dstLineStartIdx += dst.RowStride
	}

	return dst, nil
}

func copyLine(
	src, dst *Bitmap,
	sourceUpShift, sourceDownShift, padding uint,
	firstSourceByteOfLine, lastSourceByteOfLine int,
	usePadding bool, sourceOffset, targetOffset int,
) error {
	for x := firstSourceByteOfLine; x < lastSourceByteOfLine; x++ {

		if sourceOffset+1 < len(src.Data) {
			isLastByte := x+1 == lastSourceByteOfLine
			v1, err := src.GetByte(sourceOffset)
			if err != nil {
				return err
			}
			sourceOffset++
			v1 <<= sourceUpShift

			v2, err := src.GetByte(sourceOffset)
			if err != nil {
				return err
			}

			v2 >>= sourceDownShift

			value := v1 | v2

			if isLastByte && !usePadding {
				value = unpad(padding, value)
			}

			// common.Log.Debug("Value Byte in CopyLine: %08b", value)
			err = dst.SetByte(targetOffset, value)
			if err != nil {
				return err
			}
			targetOffset += 1

			if isLastByte && usePadding {
				temp, err := src.GetByte(sourceOffset)
				if err != nil {
					return err
				}
				temp <<= sourceUpShift
				value = unpad(padding, temp)

				if err = dst.SetByte(targetOffset, value); err != nil {
					return err
				}
			}

		} else {
			value, err := src.GetByte(sourceOffset)
			if err != nil {
				return err
			}
			value <<= sourceUpShift
			sourceOffset += 1
			err = dst.SetByte(targetOffset, value)
			if err != nil {
				return err
			}
			targetOffset += 1
		}
	}
	return nil
}
