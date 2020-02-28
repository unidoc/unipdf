/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

func expandBinaryFactor2(d, s *Bitmap) (err error) {
	const processName = "expandBinaryFactor2"

	bplS := s.RowStride
	bplD := d.RowStride
	var (
		source                    byte
		expanded                  uint16
		lineS, lineD, i, j, index int
	)
	for i = 0; i < s.Height; i++ {
		lineS = i * bplS
		lineD = 2 * i * bplD
		// set two bytes per line in 'd' bitmap.
		for j = 0; j < bplS; j++ {
			source = s.Data[lineS+j]
			expanded = tabExpand2x[source]
			index = lineD + j*2

			// the 'd' Bitmap might not have rowstride = 2*s.Rowstride
			// thus setting two bytes might set byte on the next row.
			if d.RowStride != s.RowStride*2 && (j+1)*2 > d.RowStride {
				err = d.SetByte(index, byte(expanded>>8))
			} else {
				err = d.setTwoBytes(index, expanded)
			}
			if err != nil {
				return errors.Wrap(err, processName, "")
			}
		}
		// copy this 'doubled' line for the next line
		for j = 0; j < bplD; j++ {
			index = lineD + bplD + j
			source = d.Data[lineD+j]
			if err = d.SetByte(index, source); err != nil {
				return errors.Wrapf(err, processName, "copy doubled line: '%d', Byte: '%d'", lineD+j, lineD+bplD+j)
			}
		}
	}
	return nil
}

func expandBinaryFactor4(d, s *Bitmap) (err error) {
	const processName = "expandBinaryFactor4"
	bplS := s.RowStride
	bplD := d.RowStride
	diff := s.RowStride*4 - d.RowStride
	var (
		source, temp                         byte
		expanded                             uint32
		lineS, lineD, i, j, k, index, iindex int
	)
	for i = 0; i < s.Height; i++ {
		lineS = i * bplS
		lineD = 4 * i * bplD
		// set four bytes per line in 'd' bitmap.
		for j = 0; j < bplS; j++ {
			source = s.Data[lineS+j]
			expanded = tabExpand4x[source]
			index = lineD + j*4

			// the 'd' Bitmap might not have rowstride = 4*s.Rowstride
			// i.e.:
			// s.width: 18 -> rowstride = 3; d.Width = 72 -> 9   | 3 * 4 = 12 | 12 - 9  = 3
			// s.width: 20 -> rowstride = 3; d.Width = 80 -> 10  | 3 * 4 = 12 | 12 - 10 = 2
			// s.width: 46 -> rowstride = 6; d.Width = 184 -> 23 | 4 * 6 = 24 | 24 - 23 = 1

			// thus setting two bytes might set byte on the next row.
			if diff != 0 && (j+1)*4 > d.RowStride {
				for k = diff; k > 0; k-- {
					temp = byte((expanded >> uint(k*8)) & 0xff)
					iindex = index + (diff - k)
					if err = d.SetByte(iindex, temp); err != nil {
						return errors.Wrapf(err, processName, "Different rowstrides. K: %d", k)
					}
				}
			} else if err = d.setFourBytes(index, expanded); err != nil {
				return errors.Wrap(err, processName, "")
			}

			if err = d.setFourBytes(lineD+j*4, tabExpand4x[s.Data[lineS+j]]); err != nil {
				return errors.Wrap(err, processName, "")
			}
		}

		// copy this 'quadrable' line for the next 3 lines too
		for k = 1; k < 4; k++ {
			for j = 0; j < bplD; j++ {
				if err = d.SetByte(lineD+k*bplD+j, d.Data[lineD+j]); err != nil {
					return errors.Wrapf(err, processName, "copy 'quadrable' line: '%d', byte: '%d'", k, j)
				}
			}
		}
	}
	return nil
}

func expandBinaryFactor8(d, s *Bitmap) (err error) {
	const processName = "expandBinaryFactor8"
	bplS := s.RowStride
	bplD := d.RowStride

	var lineS, lineD, i, j, k int
	for i = 0; i < s.Height; i++ {
		lineS = i * bplS
		lineD = 8 * i * bplD
		// set four bytes per line in 'd' bitmap.
		for j = 0; j < bplS; j++ {
			if err = d.setEightBytes(lineD+j*8, tabExpand8x[s.Data[lineS+j]]); err != nil {
				return errors.Wrap(err, processName, "")
			}
		}

		// copy this factor * 8 line for the next 7 lines too
		for k = 1; k < 8; k++ {
			for j = 0; j < bplD; j++ {
				if err = d.SetByte(lineD+k*bplD+j, d.Data[lineD+j]); err != nil {
					return errors.Wrap(err, processName, "")
				}
			}
		}
	}
	return nil
}

func expandBinaryPower2(s *Bitmap, factor int) (*Bitmap, error) {
	const processName = "expandBinaryPower2"
	if s == nil {
		return nil, errors.Error(processName, "source not defined")
	}
	if factor == 1 {
		return copyBitmap(nil, s)
	}
	if factor != 2 && factor != 4 && factor != 8 {
		return nil, errors.Error(processName, "factor must be in {2,4,8} range")
	}
	wd := factor * s.Width
	hd := factor * s.Height
	d := New(wd, hd)

	var err error
	switch factor {
	case 2:
		err = expandBinaryFactor2(d, s)
	case 4:
		err = expandBinaryFactor4(d, s)
	case 8:
		err = expandBinaryFactor8(d, s)
	}
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return d, nil
}

func expandBinaryPower2Low(d *Bitmap, s *Bitmap, factor int) (err error) {
	const processName = "expandBinaryPower2Low"
	switch factor {
	case 2:
		err = expandBinaryFactor2(d, s)
	case 4:
		err = expandBinaryFactor4(d, s)
	case 8:
		err = expandBinaryFactor8(d, s)
	default:
		return errors.Error(processName, "expansion factor not in {2,4,8} range")
	}
	if err != nil {
		err = errors.Wrap(err, processName, "")
	}
	return err
}

func expandBinaryReplicate(s *Bitmap, xFact, yFact int) (*Bitmap, error) {
	const processName = "expandBinaryReplicate"
	if s == nil {
		return nil, errors.Error(processName, "source not defined")
	}
	if xFact <= 0 || yFact <= 0 {
		return nil, errors.Error(processName, "invalid scale factor: <= 0")
	}

	if xFact == yFact {
		if xFact == 1 {
			bm, err := copyBitmap(nil, s)
			if err != nil {
				return nil, errors.Wrap(err, processName, "xFact == yFact")
			}
			return bm, nil
		}
		if xFact == 2 || xFact == 4 || xFact == 8 {
			bm, err := expandBinaryPower2(s, xFact)
			if err != nil {
				return nil, errors.Wrap(err, processName, "xFact in {2,4,8}")
			}
			return bm, nil
		}
	}

	wd := xFact * s.Width
	hd := yFact * s.Height
	d := New(wd, hd)

	bplD := d.RowStride

	var (
		lineD, i, j, k, start int
		bt                    byte
		err                   error
	)

	for i = 0; i < s.Height; i++ {
		// lineS = i * bplS
		lineD = yFact * i * bplD
		// replicate pixels on single line
		for j = 0; j < s.Width; j++ {
			// get bit at
			if pix := s.GetPixel(j, i); pix {
				start = xFact * j

				for k = 0; k < xFact; k++ {
					d.setBit(lineD*8 + start + k)
				}
			}
		}
		// replicate the line
		for k = 1; k < yFact; k++ {
			indexD := lineD + k*bplD
			// iterate over all bytes of line
			for bi := 0; bi < bplD; bi++ {
				if bt, err = d.GetByte(lineD + bi); err != nil {
					return nil, errors.Wrapf(err, processName, "replicating line: '%d'", k)
				}
				if err = d.SetByte(indexD+bi, bt); err != nil {
					return nil, errors.Wrap(err, processName, "Setting byte failed")
				}
			}
		}
	}
	return d, nil
}

func makeExpandTab2x() (tab [256]uint16) {
	for i := 0; i < 256; i++ {
		if i&0x01 != 0 {
			tab[i] |= 0x3
		}
		if i&0x02 != 0 {
			tab[i] |= 0xc
		}
		if i&0x04 != 0 {
			tab[i] |= 0x30
		}
		if i&0x08 != 0 {
			tab[i] |= 0xc0
		}
		if i&0x10 != 0 {
			tab[i] |= 0x300
		}
		if i&0x20 != 0 {
			tab[i] |= 0xc00
		}
		if i&0x40 != 0 {
			tab[i] |= 0x3000
		}
		if i&0x80 != 0 {
			tab[i] |= 0xc000
		}
	}
	return tab
}

func makeExpandTab4x() (tab [256]uint32) {
	for i := 0; i < 256; i++ {
		if i&0x01 != 0 {
			tab[i] |= 0xf
		}
		if i&0x02 != 0 {
			tab[i] |= 0xf0
		}
		if i&0x04 != 0 {
			tab[i] |= 0xf00
		}
		if i&0x08 != 0 {
			tab[i] |= 0xf000
		}
		if i&0x10 != 0 {
			tab[i] |= 0xf0000
		}
		if i&0x20 != 0 {
			tab[i] |= 0xf00000
		}
		if i&0x40 != 0 {
			tab[i] |= 0xf000000
		}
		if i&0x80 != 0 {
			tab[i] |= 0xf0000000
		}
	}
	return tab
}

func makeExpandTab8x() (tab [256]uint64) {
	for i := 0; i < 256; i++ {
		if i&0x01 != 0 {
			tab[i] |= 0xff
		}
		if i&0x02 != 0 {
			tab[i] |= 0xff00
		}
		if i&0x04 != 0 {
			tab[i] |= 0xff0000
		}
		if i&0x08 != 0 {
			tab[i] |= 0xff000000
		}
		if i&0x10 != 0 {
			tab[i] |= 0xff00000000
		}
		if i&0x20 != 0 {
			tab[i] |= 0xff0000000000
		}
		if i&0x40 != 0 {
			tab[i] |= 0xff000000000000
		}
		if i&0x80 != 0 {
			tab[i] |= 0xff00000000000000
		}
	}
	return tab
}
