/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"image"

	"github.com/unidoc/unipdf/v3/common"
)

// CombineBytes combines the provided bytes with respect to the CombinationOperator.
func CombineBytes(oldByte, newByte byte, op CombinationOperator) byte {
	return combineBytes(oldByte, newByte, op)
}

// Extract extracts the rectangle of given size from the source 'src' Bitmap.
func Extract(roi image.Rectangle, src *Bitmap) (*Bitmap, error) {
	dst := New(roi.Dx(), roi.Dy())
	upShift := roi.Min.X & 0x07
	downShift := 8 - upShift
	padding := uint(8 - dst.Width&0x07)
	srcLineStartIdx := src.GetByteIndex(roi.Min.X, roi.Min.Y)
	srcLineEndIdx := src.GetByteIndex(roi.Max.X-1, roi.Min.Y)
	usePadding := dst.RowStride == srcLineEndIdx+1-srcLineStartIdx

	var dstLineStartIdx int

	for y := roi.Min.Y; y < roi.Max.Y; y++ {
		srcIdx := srcLineStartIdx
		dstIdx := dstLineStartIdx

		switch {
		case srcLineStartIdx == srcLineEndIdx:
			pixels, err := src.GetByte(srcIdx)
			if err != nil {
				return nil, err
			}

			pixels <<= uint(upShift)

			err = dst.SetByte(dstIdx, unpad(padding, pixels))
			if err != nil {
				return nil, err
			}
		case upShift == 0:
			for x := srcLineStartIdx; x <= srcLineEndIdx; x++ {
				value, err := src.GetByte(srcIdx)
				if err != nil {
					return nil, err
				}
				srcIdx++

				if x == srcLineEndIdx && usePadding {
					value = unpad(padding, value)
				}

				err = dst.SetByte(dstIdx, value)
				if err != nil {
					return nil, err
				}

				dstIdx++
			}
		default:
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

// combineBitmap combines two bitmaps with respect to the 'op' combination operator and returns result as new Bitmap.
func combineBitmap(first, second *Bitmap, op CombinationOperator) *Bitmap {
	result := New(first.Width, first.Height)

	for i := 0; i < len(result.Data); i++ {
		result.Data[i] = combineBytes(first.Data[i], second.Data[i], op)
	}
	return result
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
	case CmbOpNot:
		return ^(newByte)
	default:
		return newByte
	}
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

			err = dst.SetByte(targetOffset, value)
			if err != nil {
				return err
			}
			targetOffset++

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
			continue
		}
		value, err := src.GetByte(sourceOffset)
		if err != nil {
			common.Log.Debug("Getting the value at: %d failed: %s", sourceOffset, err)
			return err
		}
		value <<= sourceUpShift
		sourceOffset++
		err = dst.SetByte(targetOffset, value)
		if err != nil {
			return err
		}
		targetOffset++
	}
	return nil
}
