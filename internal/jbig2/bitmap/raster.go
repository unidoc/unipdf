/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// RasterOperator is the raster operation flag operator.
// There are following raster operations:
//      PixClr                      	0000   0x0
//      PixSet                      	1111   0xf
//      PixSrc              s        	1100   0xc
//      PixDst              d        	1010   0xa
//      PixNotSrc		    ~s        	0011   0x3
//      PixNotDst		    ~d        	0101   0x5
//      PixSrcOrDst		    s | d       1110   0xe
//      PixSrcAndDst	    s & d       1000   0x8
//      PixSrcXorDst	    s ^ d       0110   0x6
//      PixNotSrcOrDst		~s | d 	    1011   0xb
//      PixNotSrcAndDst		~s & d		0010   0x2
//      PixSrcOrNotDst		s | ~d		1101   0xd
//      PixSrcAndNotDst		s & ~d		0100   0x4
//      PixNotPixSrcOrDst	~(s | d)	0001   0x1
//      PixNotPixSrcAndDst	~(s & d)	0111   0x7
//      PixNotPixSrcXorDst	~(s ^ d)	1001   0X9
type RasterOperator int

// Raster operator constant definitions.
const (
	PixSrc    RasterOperator = 0xc
	PixDst    RasterOperator = 0xa
	PixNotSrc RasterOperator = 0x3
	PixNotDst RasterOperator = 0x5

	PixClr RasterOperator = 0x0
	PixSet RasterOperator = 0xf

	PixSrcOrDst        RasterOperator = 0xe
	PixSrcAndDst       RasterOperator = 0x8
	PixSrcXorDst       RasterOperator = 0x6
	PixNotSrcOrDst     RasterOperator = 0xb
	PixNotSrcAndDst    RasterOperator = 0x2
	PixSrcOrNotDst     RasterOperator = 0xd
	PixSrcAndNotDst    RasterOperator = 0x4
	PixNotPixSrcOrDst  RasterOperator = 0x1
	PixNotPixSrcAndDst RasterOperator = 0x7
	PixNotPixSrcXorDst RasterOperator = 0x9

	PixPaint    = PixSrcOrDst
	PixSubtract = PixNotSrcAndDst
	PixMask     = PixSrcAndDst
)

// RasterOperation does the rastering operation on the provided 'dest' bitmap.
// There are 18 operations, described by the 'op' RasterOperator.
// The PixDst is a no-op.
//
// PixClr, PixSet, PixNotPixDst operate only on the 'dest'.
//
// The other 14 involve both 'src' and 'dest' bitmaps and depends on the bit values of either just the src or the
// both 'src' and 'dest'.
// Out of these 14 operators there are only 12 unique logical combinations. ~(s) ^ d) == ~(s ^ d) == s ^ ~(d).
// Parameters:
// 	'dest'	 	'dest' bitmap
//	'dx'		x val of UL corner of 'dest' bitmap
//	'dy'		y val of UL corner of 'dest' bitmap
//	'dw'		is the width of the operational rectangle on the bitmap 'dest'
//	'dh'		is the height of the operational rectangle on the bitmap 'dest'
//	'op'		raster operator code
//	'src'	 	'src' bitmap
//	'sx'		x val of UL corner of 'src' bitmap
//	'sy'		y val of UL corner of 'src' bitmap
func RasterOperation(dest *Bitmap, dx, dy, dw, dh int, op RasterOperator, src *Bitmap, sx, sy int) error {
	return rasterOperation(dest, dx, dy, dw, dh, op, src, sx, sy)
}

// RasterOperation has the same function as the RasterOperation package function, where the
// 'b' bitmap is the 'dest'.
func (b *Bitmap) RasterOperation(dx, dy, dw, dh int, op RasterOperator, src *Bitmap, sx, sy int) error {
	return rasterOperation(b, dx, dy, dw, dh, op, src, sx, sy)
}

func rasterOperation(dest *Bitmap, dx, dy, dw, dh int, op RasterOperator, src *Bitmap, sx, sy int) error {
	const processName = "rasterOperation"
	if dest == nil {
		return errors.Error(processName, "nil 'dest' Bitmap")
	}
	if op == PixDst {
		return nil
	}

	switch op {
	case PixClr, PixSet, PixNotDst:
		rasterOpUniLow(dest, dx, dy, dw, dh, op)
		return nil
	}

	if src == nil {
		common.Log.Debug("RasterOperation source bitmap is not defined")
		return errors.Error(processName, "nil 'src' bitmap")
	}
	if err := rasterOpLow(dest, dx, dy, dw, dh, op, src, sx, sy); err != nil {
		return errors.Wrap(err, processName, "")
	}
	return nil
}

// rasterOpLow scales width, performs clipping, checks alignment and dispatches for the 'op' RasterOperator.
func rasterOpLow(dest *Bitmap, dx, dy int, dw, dh int, op RasterOperator, src *Bitmap, sx, sy int) error {
	var dHangW, sHangW, dHangH, sHangH int
	// first, clip horizontally (sx, dx, dw)
	if dx < 0 {
		sx -= dx
		dw += dx
		dx = 0
	}
	if sx < 0 {
		dx -= sx
		dw += sx
		sx = 0
	}
	// rect ovhang dest to right
	dHangW = dx + dw - dest.Width
	if dHangW > 0 {
		dw -= dHangW
	}
	// rect ovhang src to right
	sHangW = sx + dw - src.Width
	if sHangW > 0 {
		dw -= sHangW
	}

	// clip vertically (sy, dy, dh)
	if dy < 0 {
		sy -= dy
		dh += dy
		dy = 0
	}
	if sy < 0 {
		dy -= sy
		dh += sy
		sy = 0
	}

	// rect ovhang dest below
	dHangH = dy + dh - dest.Height
	if dHangH > 0 {
		dh -= dHangH
	}
	// rect ovhang src below
	sHangH = sy + dh - src.Height
	if sHangH > 0 {
		dh -= sHangH
	}

	// quit if clipped entirely
	if dw <= 0 || dh <= 0 {
		return nil
	}

	// dispatch to aligned or non-aligned blitters.
	var err error
	switch {
	case dx&7 == 0 && sx&7 == 0:
		err = rasterOpByteAlignedLow(dest, dx, dy, dw, dh, op, src, sx, sy)
	case dx&7 == sx&7:
		err = rasterOpVAlignedLow(dest, dx, dy, dw, dh, op, src, sx, sy)
	default:
		err = rasterOpGeneralLow(dest, dx, dy, dw, dh, op, src, sx, sy)
	}
	if err != nil {
		return errors.Wrap(err, "rasterOpLow", "")
	}
	return nil
}

// rasterOpByteAlignedLow called when both the src and dest bitmaps are left aligned on 8-bit boundaries - dx & 7 == 0 AND sx & 7 == 0.
func rasterOpByteAlignedLow(dest *Bitmap, dx, dy, dw, dh int, op RasterOperator, src *Bitmap, sx, sy int) error {
	var (
		// mask for lat partial byte
		lwMask byte
		// index of the first src byte
		psfWord int
		// index of the first dest byte
		pdfWord int
		// line index for source and dest bitmap
		lines, lined int
		i, j         int
	)
	// get the number of full bytes
	fullBytesNumber := dw >> 3
	// get the number of ovrhang bits in last partial word
	lwBits := dw & 7
	if lwBits > 0 {
		lwMask = lmaskByte[lwBits]
	}
	psfWord = src.RowStride*sy + (sx >> 3)
	pdfWord = dest.RowStride*dy + (dx >> 3)

	switch op {
	case PixSrc:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = src.Data[lines]
				lined++
				lines++
			}
			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], src.Data[lines], lwMask)
			}
		}
	case PixNotSrc:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = ^(src.Data[lines])
				lined++
				lines++
			}
			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], ^src.Data[lines], lwMask)
			}
		}
	case PixSrcOrDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] |= src.Data[lines]
				lined++
				lines++
			}
			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], src.Data[lines]|dest.Data[lined], lwMask)
			}
		}
	case PixSrcAndDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] &= src.Data[lines]
				lined++
				lines++
			}
			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], src.Data[lines]&dest.Data[lined], lwMask)
			}
		}
	case PixSrcXorDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] ^= src.Data[lines]
				lined++
				lines++
			}
			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], src.Data[lines]^dest.Data[lined], lwMask)
			}
		}
	case PixNotSrcOrDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] |= ^(src.Data[lines])
				lined++
				lines++
			}

			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], ^(src.Data[lines])|dest.Data[lined], lwMask)
			}
		}
	case PixNotSrcAndDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] &= ^(src.Data[lines])
				lined++
				lines++
			}

			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], ^(src.Data[lines])&dest.Data[lined], lwMask)
			}
		}
	case PixSrcOrNotDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = src.Data[lines] | ^(dest.Data[lined])
				lined++
				lines++
			}

			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], src.Data[lines]|^(dest.Data[lined]), lwMask)
			}
		}
	case PixSrcAndNotDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = src.Data[lines] & ^(dest.Data[lined])
				lined++
				lines++
			}

			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], src.Data[lines]&^(dest.Data[lined]), lwMask)
			}
		}
	case PixNotPixSrcOrDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = ^(src.Data[lines] | dest.Data[lined])
				lined++
				lines++
			}

			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], ^(src.Data[lines] | dest.Data[lined]), lwMask)
			}
		}
	case PixNotPixSrcAndDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = ^(src.Data[lines] & dest.Data[lined])
				lined++
				lines++
			}

			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], ^(src.Data[lines] & dest.Data[lined]), lwMask)
			}
		}
	case PixNotPixSrcXorDst:
		for i = 0; i < dh; i++ {
			lines = psfWord + i*src.RowStride
			lined = pdfWord + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = ^(src.Data[lines] ^ dest.Data[lined])
				lined++
				lines++
			}

			if lwBits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], ^(src.Data[lines] ^ dest.Data[lined]), lwMask)
			}
		}
	default:
		common.Log.Debug("Provided invalid raster operator: %v", op)
		return errors.Error("rasterOpByteAlignedLow", "invalid raster operator")
	}
	return nil
}

// rasterOpVAlignedLow called when the left side of the src and dest bitmaps have the same alignment
// relative to 8-bit boundaries. That is: dx & 7 == sx & 7.
func rasterOpVAlignedLow(dest *Bitmap, dx, dy, dw, dh int, op RasterOperator, src *Bitmap, sx, sy int) error {
	var (
		// boolean if first dest byte is doubly partial
		dfwPart2B bool
		// boolean if there exist a full dest byte
		dfwFullB bool
		// number of full bytes in dest
		dnFullBytes int
		// index of first full dest byte
		pdfwFull int
		// index of first full src byte
		psfwFull int
		// boolean if last dest byte is partial
		dlwPartB bool
		// mask for last partial dest byte
		dlwMask byte
		// last byte dest bits in ovrhang
		dlwBits int
		// index of last partial dest byte
		pdlwPart int
		// index of last partial src byte
		pslwPart int
		i, j     int
	)

	// first byte dest bits in ovrhang
	dfwBits := 8 - (dx & 7)
	// mask for first partial dest byte
	dfwMask := rmaskByte[dfwBits]
	// index of first partial dest byte
	pdfwPart := dest.RowStride*dy + (dx >> 3)
	// index of first partial src byte
	psfwPart := src.RowStride*sy + (sx >> 3)

	// is the first word doubly partial?
	if dw < dfwBits {
		dfwPart2B = true
		dfwMask &= lmaskByte[8-dfwBits+dw]
	}

	// is there a full dest byte?
	if !dfwPart2B {
		dnFullBytes = (dw - dfwBits) >> 3
		if dnFullBytes > 0 {
			// there is a full dest byte
			dfwFullB = true
			pdfwFull = pdfwPart + 1
			psfwFull = psfwPart + 1
		}
	}

	dlwBits = (dx + dw) & 7
	// is the last word partial?
	if !(dfwPart2B || dlwBits == 0) {
		dlwPartB = true
		dlwMask = lmaskByte[dlwBits]
		pdlwPart = pdfwPart + 1 + dnFullBytes
		pslwPart = psfwPart + 1 + dnFullBytes
	}

	switch op {
	case PixSrc:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], src.Data[psfwPart], dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] = src.Data[psfwFull+j]
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], src.Data[pslwPart], dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotSrc:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^src.Data[psfwPart], dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] = ^src.Data[psfwFull+j]
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^src.Data[pslwPart], dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcOrDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], src.Data[psfwPart]|dest.Data[pdfwPart], dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] |= src.Data[psfwFull+j]
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], src.Data[pslwPart]|dest.Data[pdlwPart], dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcAndDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], src.Data[psfwPart]&dest.Data[pdfwPart], dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] &= src.Data[psfwFull+j]
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], src.Data[pslwPart]&dest.Data[pdlwPart], dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcXorDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], src.Data[psfwPart]^dest.Data[pdfwPart], dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] ^= src.Data[psfwFull+j]
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], src.Data[pslwPart]^dest.Data[pdlwPart], dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotSrcOrDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^(src.Data[psfwPart])|dest.Data[pdfwPart], dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] |= ^(src.Data[psfwFull+j])
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^(src.Data[pslwPart])|dest.Data[pdlwPart], dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotSrcAndDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^(src.Data[psfwPart])&dest.Data[pdfwPart], dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] &= ^src.Data[psfwFull+j]
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^(src.Data[pslwPart])&dest.Data[pdlwPart], dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcOrNotDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], src.Data[psfwPart]|^(dest.Data[pdfwPart]), dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] = src.Data[psfwFull+j] | ^(dest.Data[pdfwFull+j])
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], src.Data[pslwPart]|^(dest.Data[pdlwPart]), dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcAndNotDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], src.Data[psfwPart]&^(dest.Data[pdfwPart]), dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] = src.Data[psfwFull+j] & ^(dest.Data[pdfwFull+j])
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], src.Data[pslwPart]&^(dest.Data[pdlwPart]), dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotPixSrcOrDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^(src.Data[psfwPart] | dest.Data[pdfwPart]), dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] = ^(src.Data[psfwFull+j] | dest.Data[pdfwFull+j])
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^(src.Data[pslwPart] | dest.Data[pdlwPart]), dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotPixSrcAndDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^(src.Data[psfwPart] & dest.Data[pdfwPart]), dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] = ^(src.Data[psfwFull+j] & dest.Data[pdfwFull+j])
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^(src.Data[pslwPart] & dest.Data[pdlwPart]), dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotPixSrcXorDst:
		// do the first partial byte
		for i = 0; i < dh; i++ {
			dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^(src.Data[psfwPart] ^ dest.Data[pdfwPart]), dfwMask)
			pdfwPart += dest.RowStride
			psfwPart += src.RowStride
		}

		// do the full bytes
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					dest.Data[pdfwFull+j] = ^(src.Data[psfwFull+j] ^ dest.Data[pdfwFull+j])
				}
				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial byte
		if dlwPartB {
			for i = 0; i < dh; i++ {
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^(src.Data[pslwPart] ^ dest.Data[pdlwPart]), dlwMask)
				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	default:
		common.Log.Debug("Invalid raster operator: %d", op)
		return errors.Error("rasterOpVAlignedLow", "invalid raster operator")
	}
	return nil
}

// rasterOpGeneralLow called when the 'src' and 'dst' rects don't have the same byte alignment.
func rasterOpGeneralLow(dest *Bitmap, dx, dy, dw, dh int, op RasterOperator, src *Bitmap, sx, sy int) error {
	var (
		// dfwPartB boolean if first dest byte is partial.
		dfwPartB bool
		// dfwPart2B boolean if first dest byte is doubly partial.
		dfwPart2B bool
		// dfwMask mask for first partial 'dest'byte.
		dfwMask byte
		// dfwBits are first byte 'dest' bits in overhang.
		dfwBits int
		// dHang is the 'dest' overhang in first partial byte or if the
		// 'dest' is byte aligned.
		dHang int
		// pdfwPart is an index of the first partial 'dest' byte.
		pdfwPart int
		// psfwPart is an index of the first partial 'src' byte.
		psfwPart int

		// dfwFullB a boolean if there exists a full 'dest' byte.
		dfwFullB bool
		// dnFullBytes number of full bytes in 'dest'.
		dnFullBytes int
		// pdfwFull is the index of the first full 'dest' byte.
		pdfwFull int
		// psfwFull is the index of the first full 'src' byte.
		psfwFull int

		// dlwPartB boolean if last 'dest' word is partial.
		dlwPartB bool
		// dlwMask is a mask for the last partial 'dest' byte.
		dlwMask byte
		// dlwBits are the last word 'dest' bits in ovrhang.
		dlwBits int
		// pdlwPart is the index of the last partial 'dest' byte.
		pdlwPart int
		// pslwPart is the index of the last partial 'src' byte.
		pslwPart int

		// sBytes compose 'src' byte aligned with the 'dest' bytes.
		sBytes byte
		// sfwBits is a first word 'src' bits in overhang, or 8 if src is byte aligned.
		sfwBits int
		// sHang is a 'src' overhang in the first partial byte or 0 if 'src' is word aligned.
		sHang int
		// sleftShift are the bits to shift left for source bytes to align
		// with the 'dest'. Also the number of bits that get shifted to the right to align with the 'dest'.
		sleftShift uint
		// srightShift are the bits to shift right for source byte to align with the 'dest'. Also the number of bits
		// that get shifted to the left to align with the 'dest'.
		srightShift uint
		// srightMask is the mask for selecting sleftshift bits that have been shifted right by srightshift bits.
		srightMask byte
		// shift direction
		sfwShiftDir shift
		// sfwAddB is additional sfw right shift needed
		sfwAddB bool
		// slwAddB is additional slw right shift needed
		slwAddB bool
		i, j    int
	)

	if sx&7 != 0 {
		sHang = 8 - (sx & 7)
	}
	if dx&7 != 0 {
		dHang = 8 - (dx & 7)
	}

	if sHang == 0 && dHang == 0 {
		srightMask = rmaskByte[0]
	} else {
		if dHang > sHang {
			sleftShift = uint(dHang - sHang)
		} else {
			sleftShift = uint(8 - (sHang - dHang))
		}
		srightShift = 8 - sleftShift
		srightMask = rmaskByte[sleftShift]
	}

	// is the first dest word partial?
	if (dx & 7) != 0 {
		dfwPartB = true
		dfwBits = 8 - (dx & 7)
		dfwMask = rmaskByte[dfwBits]

		pdfwPart = dest.RowStride*dy + (dx >> 3)
		psfwPart = src.RowStride*sy + (sx >> 3)

		sfwBits = 8 - (sx & 7)
		if dfwBits > sfwBits {
			sfwShiftDir = shiftLeft
			if dw >= sHang {
				sfwAddB = true
			}
		} else {
			sfwShiftDir = shiftRight
		}
	}

	if dw < dfwBits {
		dfwPart2B = true
		dfwMask &= lmaskByte[8-dfwBits+dw]
	}

	if !dfwPart2B {
		dnFullBytes = (dw - dfwBits) >> 3
		if dnFullBytes != 0 {
			dfwFullB = true
			pdfwFull = dest.RowStride*dy + ((dx + dHang) >> 3)
			psfwFull = src.RowStride*sy + ((sx + dHang) >> 3)
		}
	}

	dlwBits = (dx + dw) & 7
	if !(dfwPart2B || dlwBits == 0) {
		dlwPartB = true
		dlwMask = lmaskByte[dlwBits]
		pdlwPart = dest.RowStride*dy + ((dx + dHang) >> 3) + dnFullBytes
		pslwPart = src.RowStride*sy + ((sx + dHang) >> 3) + dnFullBytes

		if dlwBits > int(srightShift) {
			slwAddB = true
		}
	}

	switch op {
	case PixSrc:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], sBytes, dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] = sBytes
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], sBytes, dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotSrc:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^sBytes, dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] = ^sBytes
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^sBytes, dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcOrDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], sBytes|dest.Data[pdfwPart], dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] |= sBytes
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], sBytes|dest.Data[pdlwPart], dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcAndDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], sBytes&dest.Data[pdfwPart], dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] &= sBytes
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], sBytes&dest.Data[pdlwPart], dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcXorDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], sBytes^dest.Data[pdfwPart], dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] ^= sBytes
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], sBytes^dest.Data[pdlwPart], dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotSrcOrDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^sBytes|dest.Data[pdfwPart], dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] |= ^sBytes
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^sBytes|dest.Data[pdlwPart], dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotSrcAndDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^sBytes&dest.Data[pdfwPart], dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] &= ^sBytes
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^sBytes&dest.Data[pdlwPart], dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcOrNotDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], sBytes|^dest.Data[pdfwPart], dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] = sBytes | ^dest.Data[pdfwFull+j]
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], sBytes|^dest.Data[pdlwPart], dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixSrcAndNotDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], sBytes & ^dest.Data[pdfwPart], dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] = sBytes & ^dest.Data[pdfwFull+j]
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], sBytes & ^dest.Data[pdlwPart], dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotPixSrcOrDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^(sBytes | dest.Data[pdfwPart]), dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] = ^(sBytes | dest.Data[pdfwFull+j])
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^(sBytes | dest.Data[pdlwPart]), dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotPixSrcAndDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^(sBytes & dest.Data[pdfwPart]), dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] = ^(sBytes & dest.Data[pdfwFull+j])
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^(sBytes & dest.Data[pdlwPart]), dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	case PixNotPixSrcXorDst:
		// do the first partial word
		if dfwPartB {
			for i = 0; i < dh; i++ {
				if sfwShiftDir == shiftLeft {
					sBytes = src.Data[psfwPart] << sleftShift
					if sfwAddB {
						sBytes = combinePartial(sBytes, src.Data[psfwPart+1]>>srightShift, srightMask)
					}
				} else {
					sBytes = src.Data[psfwPart] >> srightShift
				}

				dest.Data[pdfwPart] = combinePartial(dest.Data[pdfwPart], ^(sBytes ^ dest.Data[pdfwPart]), dfwMask)
				pdfwPart += dest.RowStride
				psfwPart += src.RowStride
			}
		}

		// do the full words
		if dfwFullB {
			for i = 0; i < dh; i++ {
				for j = 0; j < dnFullBytes; j++ {
					sBytes = combinePartial(src.Data[psfwFull+j]<<sleftShift, src.Data[psfwFull+j+1]>>srightShift, srightMask)
					dest.Data[pdfwFull+j] = ^(sBytes ^ dest.Data[pdfwFull+j])
				}

				pdfwFull += dest.RowStride
				psfwFull += src.RowStride
			}
		}

		// do the last partial word
		if dlwPartB {
			for i = 0; i < dh; i++ {
				sBytes = src.Data[pslwPart] << sleftShift
				if slwAddB {
					sBytes = combinePartial(sBytes, src.Data[pslwPart+1]>>srightShift, srightMask)
				}
				dest.Data[pdlwPart] = combinePartial(dest.Data[pdlwPart], ^(sBytes ^ dest.Data[pdlwPart]), dlwMask)

				pdlwPart += dest.RowStride
				pslwPart += src.RowStride
			}
		}
	default:
		common.Log.Debug("Operation: '%d' not permitted", op)
		return errors.Error("rasterOpGeneralLow", "raster operation not permitted")
	}

	return nil
}

// rasterOpUniLow performs clipping, checks aligment and dispatches for the rasterop.
func rasterOpUniLow(dest *Bitmap, dx, dy, dw, dh int, op RasterOperator) {
	// clip horizontally (dx, dw)
	if dx < 0 {
		dw += dx
		dx = 0
	}

	dHangw := dx + dw - dest.Width
	if dHangw > 0 {
		// reduce dw
		dw -= dHangw
	}

	// clip vertically(dy, dh)
	if dy < 0 {
		dh += dy
		dy = 0
	}

	dHangh := dy + dh - dest.Height
	if dHangh > 0 {
		// reduce dh
		dh -= dHangh
	}

	// if clipped entirely quit
	if dw <= 0 || dh <= 0 {
		return
	}

	if (dx & 7) == 0 {
		rasterOpUniWordAlignedLow(dest, dx, dy, dw, dh, op)
	} else {
		rasterOpUniGeneralLow(dest, dx, dy, dw, dh, op)
	}
}

// rasterOpUniWordAligneLow is called when the 'dest' bitmap is left aligned. This means dx & 7 == 0.
func rasterOpUniWordAlignedLow(dest *Bitmap, dx, dy int, dw, dh int, op RasterOperator) {
	var (
		// firstWordIndex is the byte index of the first byte
		firstWordIndex int
		// lwMask is the mask for the last partial byte
		lwmask byte
		i, j   int
		lined  int
	)
	// fullBytesNumber is the number of Full bytes.
	fullBytesNumber := dw >> 3
	// lwBits is the number of ovrhang bits in last partial byte.
	lwbits := dw & 7
	if lwbits > 0 {
		lwmask = lmaskByte[lwbits]
	}

	firstWordIndex = dest.RowStride*dy + (dx >> 3)

	switch op {
	case PixClr:
		for i = 0; i < dh; i++ {
			lined = firstWordIndex + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = 0x0
				lined++
			}
			if lwbits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], 0x0, lwmask)
			}
		}
	case PixSet:
		for i = 0; i < dh; i++ {
			lined = firstWordIndex + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = 0xff
				lined++
			}
			if lwbits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], 0xff, lwmask)
			}
		}
	case PixNotDst:
		for i = 0; i < dh; i++ {
			lined = firstWordIndex + i*dest.RowStride
			for j = 0; j < fullBytesNumber; j++ {
				dest.Data[lined] = ^dest.Data[lined]
				lined++
			}
			if lwbits > 0 {
				dest.Data[lined] = combinePartial(dest.Data[lined], ^dest.Data[lined], lwmask)
			}
		}
	}
}

// rasterOpUniGeneralLow static low-level uni rasterop without byte aligment.
func rasterOpUniGeneralLow(dest *Bitmap, dx, dy int, dw, dh int, op RasterOperator) {
	var (
		// fbDoublyPartial boolean if first 'dest' byte is doubly partial.
		fbDoublyPartial bool
		// fbFull boolean if there exists a full 'dest' byte word.
		fbFull bool
		// fullBytesNumber is the number of full 'dest' bytes.
		fullBytesNumber int
		// firstFullByteIndex is an index in 'dest'.Data' of the first full byte.
		firstFullByteIndex int

		// lbBits last byte 'dest' bits in ovrhang.
		lbBits int
		// lastBytePartialIndex is an index in 'dest'.Data of the last partial byte.
		lastBytePartialIndex int
		// lbPartial boolean if last 'dest' byte is partial
		lbPartial bool
		// lbMask mask for the last partial 'dest' byte.
		lbMask byte
	)
	// fbBits first byte 'dest' bits int ovrhang.
	fbBits := 8 - (dx & 7)
	// fbMask is a mask for first partial 'dest' byte.
	fbMask := rmaskByte[fbBits]
	// firstBytePartialIndex is an index in 'dest'.Data of the first partial byte.
	firstBytePartialIndex := dest.RowStride*dy + (dx >> 3)

	// is the first byte is doubly partial?
	if dw < fbBits {
		fbDoublyPartial = true
		fbMask &= lmaskByte[8-fbBits+dw]
	}

	// is there a full 'dest' byte?
	if !fbDoublyPartial {
		fullBytesNumber = (dw - fbBits) >> 3
		if fullBytesNumber != 0 {
			fbFull = true
			firstFullByteIndex = firstBytePartialIndex + 1
		}
	}

	// is the last byte partial?
	lbBits = (dx + dw) & 7
	if !(fbDoublyPartial || lbBits == 0) {
		lbPartial = true
		lbMask = lmaskByte[lbBits]
		lastBytePartialIndex = firstBytePartialIndex + 1 + fullBytesNumber
	}

	var i, j int
	switch op {
	case PixClr:
		// do the first partial word
		for i = 0; i < dh; i++ {
			dest.Data[firstBytePartialIndex] = combinePartial(dest.Data[firstBytePartialIndex], 0x0, fbMask)
			firstBytePartialIndex += dest.RowStride
		}

		// do the full words
		if fbFull {
			for i = 0; i < dh; i++ {
				for j = 0; j < fullBytesNumber; j++ {
					dest.Data[firstFullByteIndex+j] = 0x0
				}
				firstFullByteIndex += dest.RowStride
			}
		}

		// last partial word
		if lbPartial {
			for i = 0; i < dh; i++ {
				dest.Data[lastBytePartialIndex] = combinePartial(dest.Data[lastBytePartialIndex], 0x0, lbMask)
				lastBytePartialIndex += dest.RowStride
			}
		}
	case PixSet:
		// do the first partial word
		for i = 0; i < dh; i++ {
			dest.Data[firstBytePartialIndex] = combinePartial(dest.Data[firstBytePartialIndex], 0xff, fbMask)
			firstBytePartialIndex += dest.RowStride
		}

		// do the full words
		if fbFull {
			for i = 0; i < dh; i++ {
				for j = 0; j < fullBytesNumber; j++ {
					dest.Data[firstFullByteIndex+j] = 0xff
				}
				firstFullByteIndex += dest.RowStride
			}
		}

		// last partial word
		if lbPartial {
			for i = 0; i < dh; i++ {
				dest.Data[lastBytePartialIndex] = combinePartial(dest.Data[lastBytePartialIndex], 0xff, lbMask)
				lastBytePartialIndex += dest.RowStride
			}
		}
	case PixNotDst:
		// do the first partial word
		for i = 0; i < dh; i++ {
			dest.Data[firstBytePartialIndex] = combinePartial(dest.Data[firstBytePartialIndex], ^dest.Data[firstBytePartialIndex], fbMask)
			firstBytePartialIndex += dest.RowStride
		}

		// do the full words
		if fbFull {
			for i = 0; i < dh; i++ {
				for j = 0; j < fullBytesNumber; j++ {
					dest.Data[firstFullByteIndex+j] = ^(dest.Data[firstFullByteIndex+j])
				}
				firstFullByteIndex += dest.RowStride
			}
		}

		// last partial word
		if lbPartial {
			for i = 0; i < dh; i++ {
				dest.Data[lastBytePartialIndex] = combinePartial(dest.Data[lastBytePartialIndex], ^dest.Data[lastBytePartialIndex], lbMask)
				lastBytePartialIndex += dest.RowStride
			}
		}
	}
}

var (
	lmaskByte = []byte{
		// 00000000
		0x00,
		// 10000000
		0x80,
		// 11000000
		0xC0,
		// 11100000
		0xE0,
		// 11110000
		0xF0,
		// 11111000
		0xF8,
		// 11111100
		0xFC,
		// 11111110
		0xFE,
		// 11111111
		0xFF,
	}
	rmaskByte = []byte{
		// 00000000
		0x00,
		// 00000001
		0x01,
		// 00000011
		0x03,
		// 00000111
		0x07,
		// 00001111
		0x0F,
		// 00011111
		0x1F,
		// 00111111
		0x3F,
		// 01111111
		0x7F,
		// 11111111
		0xFF,
	}
)

func combinePartial(d, s, m byte) byte {
	return (d & ^(m)) | (s & m)
}

type shift int

const (
	shiftLeft shift = iota
	shiftRight
)
