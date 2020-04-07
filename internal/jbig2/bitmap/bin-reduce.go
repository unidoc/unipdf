/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"encoding/binary"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// reduceBinaryCascade performs up to four cascades 2x rank reductions.
func reduceRankBinaryCascade(s *Bitmap, levels ...int) (d *Bitmap, err error) {
	const processName = "reduceRankBinaryCascade"
	if s == nil {
		return nil, errors.Error(processName, "source bitmap not defined")
	}

	if len(levels) == 0 || len(levels) > 4 {
		return nil, errors.Error(processName, "there must be at least one and at most 4 levels")
	}

	if levels[0] <= 0 {
		common.Log.Debug("level1 <= 0 - no reduction")
		d, err = copyBitmap(nil, s)
		if err != nil {
			return nil, errors.Wrap(err, processName, "level1 <= 0")
		}
		return d, nil
	}

	tab := makeSubsampleTab2x()
	d = s // no need to copy the 's' as the 'reduceRankBinary2 creates new 'd' Bitmap.
	for i, level := range levels {
		if level <= 0 {
			break
		}
		d, err = reduceRankBinary2(d, level, tab)
		if err != nil {
			return nil, errors.Wrapf(err, processName, "level%d reduction", i)
		}
	}
	return d, nil
}

// reduceRankBinary2 reduces the size of the 'source' bitmap by 2.
// The 'level' is the rank threshold which defines the number of pixels in each 2x2 region
// that are required to set corresponding pixel 'ON' int the 'd' Bitmap.
func reduceRankBinary2(s *Bitmap, level int, tab []byte) (d *Bitmap, err error) {
	const processName = "reduceRankBinary2"
	if s == nil {
		return nil, errors.Error(processName, "source bitmap not defined")
	}
	if level < 1 || level > 4 {
		return nil, errors.Error(processName, "level must be in set {1,2,3,4}")
	}
	if s.Height <= 1 {
		return nil, errors.Errorf(processName, "source height must be at least '2' - is: '%d'", s.Height)
	}

	// create a bitmap half a side of the 's' bitmap
	d = New(s.Width/2, s.Height/2)

	if tab == nil {
		tab = makeSubsampleTab2x()
	}

	// get the data in the []uint32  word form
	// let's assume big endian
	// wpl := (s.Width + 31) >> 5
	wpl := min(s.RowStride, 2*d.RowStride)

	switch level {
	case 1:
		err = reduceRankBinary2Level1(s, d, level, tab, wpl)
	case 2:
		err = reduceRankBinary2Level2(s, d, level, tab, wpl)
	case 3:
		err = reduceRankBinary2Level3(s, d, level, tab, wpl)
	case 4:
		err = reduceRankBinary2Level4(s, d, level, tab, wpl)
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func reduceRankBinary2Level1(s, d *Bitmap, level int, tab []byte, wpl int) (err error) {
	const processName = "reduceRankBinary2Level1"
	var (
		lineS, lineD, i, id, j, jd, k, index int
		word1, word2                         uint32
		byte0, byte1                         byte
		twoBytes                             uint16
	)
	bytes1 := make([]byte, 4)
	bytes2 := make([]byte, 4)

	for i = 0; i < s.Height-1; i, id = i+2, id+1 {
		lineS = i * s.RowStride
		lineD = id * d.RowStride
		for j, jd = 0, 0; j < wpl; j, jd = j+4, jd+1 {
			// get four bytes from the bitmap 's' and 'd' at the j'th index
			for k = 0; k < 4; k++ {
				index = lineS + j + k
				// check if the bitmap 's' has byte at 'index'
				if index <= len(s.Data)-1 && index < lineS+s.RowStride {
					bytes1[k] = s.Data[index]
				} else {
					bytes1[k] = 0x00
				}
				index = lineS + s.RowStride + j + k
				if index <= len(s.Data)-1 && index < lineS+(2*s.RowStride) {
					bytes2[k] = s.Data[index]
				} else {
					bytes2[k] = 0x00
				}
			}
			word1 = binary.BigEndian.Uint32(bytes1)
			word2 = binary.BigEndian.Uint32(bytes2)

			word2 |= word1
			word2 |= word2 << 1

			word2 &= 0xaaaaaaaa
			word1 = word2 | (word2 << 7)

			byte0 = byte(word1 >> 24)
			byte1 = byte((word1 >> 8) & 0xff)

			index = lineD + jd
			if index+1 == len(d.Data)-1 || index+1 >= lineD+d.RowStride {
				// set only one byte0 if there is no place for 'two' bytes.
				d.Data[index] = tab[byte0]
			} else {
				twoBytes = (uint16(tab[byte0]) << 8) | uint16(tab[byte1])
				if err = d.setTwoBytes(index, twoBytes); err != nil {
					return errors.Wrapf(err, processName, "setting two bytes failed, index: %d", index)
				}
				// while setting two bytes the index should increase.
				jd++
			}
		}
	}
	return nil
}

func reduceRankBinary2Level2(s, d *Bitmap, level int, tab []byte, wpl int) (err error) {
	const processName = "reduceRankBinary2Level2"
	var (
		lineS, lineD, i, id, j, jd, k, index int
		word1, word2, word3, word4           uint32
		byte0, byte1                         byte
		twoBytes                             uint16
	)
	bytes1 := make([]byte, 4)
	bytes2 := make([]byte, 4)

	for i = 0; i < s.Height-1; i, id = i+2, id+1 {
		lineS = i * s.RowStride
		lineD = id * d.RowStride
		for j, jd = 0, 0; j < wpl; j, jd = j+4, jd+1 {
			// get four bytes from the bitmap 's' and 'd' at the j'th index
			for k = 0; k < 4; k++ {
				index = lineS + j + k
				// check if the bitmap 's' has byte at 'index'
				if index <= len(s.Data)-1 && index < lineS+s.RowStride {
					bytes1[k] = s.Data[index]
				} else {
					bytes1[k] = 0x00
				}
				index = lineS + s.RowStride + j + k
				if index <= len(s.Data)-1 && index < lineS+(2*s.RowStride) {
					bytes2[k] = s.Data[index]
				} else {
					bytes2[k] = 0x00
				}
			}
			word1 = binary.BigEndian.Uint32(bytes1)
			word2 = binary.BigEndian.Uint32(bytes2)

			word3 = word1 & word2
			word3 |= word3 << 1
			word4 = word1 | word2
			word4 &= word4 << 1
			word2 = word3 | word4

			word2 &= 0xaaaaaaaa
			word1 = word2 | (word2 << 7)

			byte0 = byte(word1 >> 24)
			byte1 = byte((word1 >> 8) & 0xff)

			index = lineD + jd
			if index+1 == len(d.Data)-1 || index+1 >= lineD+d.RowStride {
				// set only one byte0 if there is no place for 'two' bytes.
				if err = d.SetByte(index, tab[byte0]); err != nil {
					return errors.Wrapf(err, processName, "index: %d", index)
				}
			} else {
				twoBytes = (uint16(tab[byte0]) << 8) | uint16(tab[byte1])
				if err = d.setTwoBytes(index, twoBytes); err != nil {
					return errors.Wrapf(err, processName, "setting two bytes failed, index: %d", index)
				}
				// while setting two bytes the index should increase.
				jd++
			}
		}
	}
	return nil
}

func reduceRankBinary2Level3(s, d *Bitmap, level int, tab []byte, wpl int) (err error) {
	const processName = "reduceRankBinary2Level3"
	var (
		lineS, lineD, i, id, j, jd, k, index int
		word1, word2, word3, word4           uint32
		byte0, byte1                         byte
		twoBytes                             uint16
	)
	bytes1 := make([]byte, 4)
	bytes2 := make([]byte, 4)

	for i = 0; i < s.Height-1; i, id = i+2, id+1 {
		lineS = i * s.RowStride
		lineD = id * d.RowStride
		for j, jd = 0, 0; j < wpl; j, jd = j+4, jd+1 {
			// get four bytes from the bitmap 's' and 'd' at the j'th index
			for k = 0; k < 4; k++ {
				index = lineS + j + k
				// check if the bitmap 's' has byte at 'index'
				if index <= len(s.Data)-1 && index < lineS+s.RowStride {
					bytes1[k] = s.Data[index]
				} else {
					bytes1[k] = 0x00
				}
				index = lineS + s.RowStride + j + k
				if index <= len(s.Data)-1 && index < lineS+(2*s.RowStride) {
					bytes2[k] = s.Data[index]
				} else {
					bytes2[k] = 0x00
				}
			}
			word1 = binary.BigEndian.Uint32(bytes1)
			word2 = binary.BigEndian.Uint32(bytes2)

			word3 = word1 & word2
			word3 |= word3 << 1
			word4 = word1 | word2
			word4 &= word4 << 1
			word2 = word3 & word4

			word2 &= 0xaaaaaaaa
			word1 = word2 | (word2 << 7)

			byte0 = byte(word1 >> 24)
			byte1 = byte((word1 >> 8) & 0xff)

			index = lineD + jd
			if index+1 == len(d.Data)-1 || index+1 >= lineD+d.RowStride {
				// set only one byte0 if there is no place for 'two' bytes.
				if err = d.SetByte(index, tab[byte0]); err != nil {
					return errors.Wrapf(err, processName, "index: %d", index)
				}
			} else {
				twoBytes = (uint16(tab[byte0]) << 8) | uint16(tab[byte1])
				if err = d.setTwoBytes(index, twoBytes); err != nil {
					return errors.Wrapf(err, processName, "setting two bytes failed, index: %d", index)
				}
				// while setting two bytes the index should increase.
				jd++
			}
		}
	}
	return nil
}

func reduceRankBinary2Level4(s, d *Bitmap, level int, tab []byte, wpl int) (err error) {
	const processName = "reduceRankBinary2Level4"
	var (
		lineS, lineD, i, id, j, jd, k, index int
		word1, word2                         uint32
		byte0, byte1                         byte
		twoBytes                             uint16
	)
	bytes1 := make([]byte, 4)
	bytes2 := make([]byte, 4)

	for i = 0; i < s.Height-1; i, id = i+2, id+1 {
		lineS = i * s.RowStride
		lineD = id * d.RowStride
		for j, jd = 0, 0; j < wpl; j, jd = j+4, jd+1 {
			// get four bytes from the bitmap 's' and 'd' at the j'th index
			for k = 0; k < 4; k++ {
				index = lineS + j + k
				// check if the bitmap 's' has byte at 'index'
				if index <= len(s.Data)-1 && index < lineS+s.RowStride {
					bytes1[k] = s.Data[index]
				} else {
					bytes1[k] = 0x00
				}
				index = lineS + s.RowStride + j + k
				if index <= len(s.Data)-1 && index < lineS+(2*s.RowStride) {
					bytes2[k] = s.Data[index]
				} else {
					bytes2[k] = 0x00
				}
			}
			word1 = binary.BigEndian.Uint32(bytes1)
			word2 = binary.BigEndian.Uint32(bytes2)

			word2 &= word1
			word2 &= word2 << 1

			word2 &= 0xaaaaaaaa
			word1 = word2 | (word2 << 7)

			byte0 = byte(word1 >> 24)
			byte1 = byte((word1 >> 8) & 0xff)

			index = lineD + jd
			if index+1 == len(d.Data)-1 || index+1 >= lineD+d.RowStride {
				// set only one byte0 if there is no place for 'two' bytes.
				d.Data[index] = tab[byte0]
				if err = d.SetByte(index, tab[byte0]); err != nil {
					return errors.Wrapf(err, processName, "index: %d", index)
				}
			} else {
				twoBytes = (uint16(tab[byte0]) << 8) | uint16(tab[byte1])
				if err = d.setTwoBytes(index, twoBytes); err != nil {
					return errors.Wrapf(err, processName, "setting two bytes failed, index: %d", index)
				}
				// while setting two bytes the index should increase.
				jd++
			}
		}
	}
	return nil
}

func makeSubsampleTab2x() (tab []byte) {
	tab = make([]byte, 256)
	for i := 0; i < 256; i++ {
		i := byte(i)
		tab[i] = (i & 0x01) | /* 7 */
			((i & 0x04) >> 1) | /* 6 */
			((i & 0x10) >> 2) | /* 5 */
			((i & 0x40) >> 3) | /* 4 */
			((i & 0x02) << 3) | /* 3 */
			((i & 0x08) << 2) | /* 2 */
			((i & 0x20) << 1) | /* 1 */
			(i & 0x80) /* 0 */
	}
	return tab
}
