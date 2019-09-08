/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package mmr

import (
	"io"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

const (
	maxRunDataBuffer int  = 1024 << 7
	minRunDataBuffer int  = 3
	codeOffset       uint = 24
)

type runData struct {
	r          *reader.SubstreamReader
	offset     int
	lastOffset int
	lastCode   int

	buffer     []byte
	bufferBase int
	bufferTop  int
}

func newRunData(r *reader.SubstreamReader) (*runData, error) {
	d := &runData{
		r:          r,
		offset:     0,
		lastOffset: 1,
	}

	length := minInt(maxInt(minRunDataBuffer, int(r.Length())), maxRunDataBuffer)
	d.buffer = make([]byte, length)

	if err := d.fillBuffer(0); err != nil {
		if err == io.EOF {
			d.buffer = make([]byte, 10)
			common.Log.Debug("FillBuffer failed: %v", err)
		} else {
			return nil, err
		}
	}
	return d, nil
}

// align skips to the next byte
func (r *runData) align() {
	r.offset = ((r.offset + 7) >> 3) << 3
}

func (r *runData) uncompressGetCode(table []*code) (*code, error) {
	return r.uncompressGetCodeLittleEndian(table)
}

func (r *runData) uncompressGetCodeLittleEndian(table []*code) (*code, error) {
	cd, err := r.uncompressGetNextCodeLittleEndian()
	if err != nil {
		common.Log.Debug("UncompressGetNextCodeLittleEndian failed: %v", err)
		return nil, err
	}

	cd &= 0xffffff
	index := cd >> (codeOffset - firstLevelTableSize)
	result := table[index]

	if result != nil && result.nonNilSubTable {
		index = (cd >> (codeOffset - firstLevelTableSize - secondLevelTableSize)) & secondLevelTableMask
		result = result.subTable[index]
	}
	return result, nil
}

func (r *runData) uncompressGetNextCodeLittleEndian() (int, error) {
	bitsToFill := r.offset - r.lastOffset

	// check whether we can refill, or need to fill in absolute mode
	if bitsToFill < 0 || bitsToFill > 24 {
		// refill at absolute offset
		byteOffset := (r.offset >> 3) - r.bufferBase

		if byteOffset >= r.bufferTop {
			byteOffset += r.bufferBase
			if err := r.fillBuffer(byteOffset); err != nil {
				return 0, err
			}

			byteOffset -= r.bufferBase
		}

		lastCode := (uint32(r.buffer[byteOffset]&0xFF) << 16) |
			(uint32(r.buffer[byteOffset+1]&0xFF) << 8) |
			(uint32(r.buffer[byteOffset+2] & 0xFF))
		bitOffset := uint32(r.offset & 7)
		lastCode <<= bitOffset
		r.lastCode = int(lastCode)
	} else {
		// the offset to the next byte boundary as seen from the last offset
		bitOffset := r.lastOffset & 7 // lastoffset % 8
		avail := 7 - bitOffset

		if bitsToFill <= avail {
			r.lastCode <<= uint(bitsToFill)
		} else {
			byteOffset := (r.lastOffset >> 3) + 3 - r.bufferBase

			if byteOffset >= r.bufferTop {
				byteOffset += r.bufferBase
				if err := r.fillBuffer(byteOffset); err != nil {
					return 0, err
				}
				byteOffset -= r.bufferBase
			}

			bitOffset = 8 - bitOffset

			for {
				r.lastCode <<= uint(bitOffset)
				r.lastCode |= int(uint(r.buffer[byteOffset]) & 0xFF)
				bitsToFill -= bitOffset
				byteOffset++
				bitOffset = 8

				if !(bitsToFill >= 8) {
					break
				}
			}

			r.lastCode <<= uint(bitsToFill)
		}
	}
	r.lastOffset = r.offset
	return r.lastCode, nil
}

func (r *runData) fillBuffer(byteOffset int) error {
	r.bufferBase = byteOffset

	_, err := r.r.Seek(int64(byteOffset), io.SeekStart)
	if err != nil {
		if err == io.EOF {
			common.Log.Debug("Seak EOF")
			r.bufferTop = -1
		} else {
			return err
		}
	}

	if err == nil {
		r.bufferTop, err = r.r.Read(r.buffer)
		if err != nil {
			if err == io.EOF {
				common.Log.Trace("Read EOF")
				r.bufferTop = -1
			} else {
				return err
			}
		}
	}

	// check filling degree
	if r.bufferTop > -1 && r.bufferTop < 3 {
		for r.bufferTop < 3 {
			b, err := r.r.ReadByte()
			if err != nil {
				if err == io.EOF {
					r.buffer[r.bufferTop] = 0
				} else {
					return err
				}
			} else {
				r.buffer[r.bufferTop] = b & 0xFF
			}
			r.bufferTop++
		}
	}

	// leave some room in order to save a few tests in the calling code
	r.bufferTop -= 3

	if r.bufferTop < 0 {
		// if we're at EOF just supply zero-bytes
		r.buffer = make([]byte, len(r.buffer))
		r.bufferTop = len(r.buffer) - 3
	}

	return nil
}
