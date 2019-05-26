/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package mmr

import (
	"errors"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// Decoder is the mmr decoder model
type Decoder struct {
	width, height int
	data          *runData

	whiteTable []*code
	blackTable []*code
	modeTable  []*code
}

// New creates new mmr decoder
func New(r reader.StreamReader, width, height int, dataOffset, dataLength int64) (*Decoder, error) {
	m := &Decoder{
		width:  width,
		height: height,
	}

	s, err := reader.NewSubstreamReader(r, uint64(dataOffset), uint64(dataLength))
	if err != nil {
		return nil, err
	}

	rd, err := newRunData(s)
	if err != nil {
		return nil, err
	}

	m.data = rd

	if err := m.initTables(); err != nil {
		return nil, err
	}

	return m, nil
}

// Data returns the decoder RunData
func (m *Decoder) Data() *runData {
	return m.data
}

func (m *Decoder) initTables() (err error) {
	if m.whiteTable == nil {
		m.whiteTable, err = m.createLittleEndianTable(WhiteCodes)
		if err != nil {
			return
		}
		m.blackTable, err = m.createLittleEndianTable(BlackCodes)
		if err != nil {
			return
		}
		m.modeTable, err = m.createLittleEndianTable(ModeCodes)
		if err != nil {
			return
		}
	}
	return nil
}

// Uncompress2d uncopmresses the 2d rows
func (m *Decoder) Uncompress2d(
	rd *runData,
	referenceOffsets []int,
	refRunLength int,
	runOffsets []int,
	width int,
) (int, error) {

	var (
		referenceBuffOffset    int
		currentBuffOffset      int
		currentLineBitPosition int

		whiteRun = true
		err      error

		c *code
	)

	// common.Log.Debug("refRunLength: %v", refRunLength)
	referenceOffsets[refRunLength] = width
	referenceOffsets[refRunLength+1] = width
	referenceOffsets[refRunLength+2] = width + 1
	referenceOffsets[refRunLength+3] = width + 1

decodeLoop:
	for currentLineBitPosition < width {
		// common.Log.Debug("CurrentLineBitPosition: %v, Width: %v", currentLineBitPosition, m.width)

		// common.Log.Debug("Mode Table")

		// Get the mode code
		c, err = rd.uncompressGetCode(m.modeTable)
		if err != nil {
			// common.Log.Debug("UncompressedGetCode failed: %v", err)
			return eol, nil
		}

		// common.Log.Debug("Code: %v", c)

		if c == nil {
			rd.offset++
			break decodeLoop
		}

		rd.offset += c.bitLength
		// common.Log.Debug("MMRCode: %v", c.runLength)

		switch mmrCode(c.runLength) {
		case codeV0:

			currentLineBitPosition = referenceOffsets[referenceBuffOffset]
		case codeVR1:
			currentLineBitPosition = referenceOffsets[referenceBuffOffset] + 1
		case codeVL1:

			currentLineBitPosition = referenceOffsets[referenceBuffOffset] - 1
		case codeH:
			var ever = 1
			// common.Log.Debug("CodeH")
			for ever > 0 {
				var table []*code
				// common.Log.Debug("WhiteRun: %v", whiteRun)
				if whiteRun {
					table = m.whiteTable
				} else {
					table = m.blackTable
				}

				// common.Log.Debug("UncompressGetCode")
				c, err = rd.uncompressGetCode(table)
				if err != nil {
					return 0, err
				}

				if c == nil {
					// common.Log.Debug("uncompressGetCode no code found.")
					break decodeLoop
				}

				rd.offset += c.bitLength

				// common.Log.Debug("RunLength: %v", c.runLength)
				if c.runLength < 64 {
					if c.runLength < 0 {
						// common.Log.Debug("runLength smaller than 0")
						runOffsets[currentBuffOffset] = currentLineBitPosition
						currentBuffOffset++
						c = nil
						break decodeLoop
					}
					currentLineBitPosition += c.runLength
					runOffsets[currentBuffOffset] = currentLineBitPosition
					currentBuffOffset++
					break
				}
				currentLineBitPosition += c.runLength
			}

			firstHalfBitPost := currentLineBitPosition

			var ever1 = 1

		ever1Loop:
			for ever1 > 0 {
				var table []*code
				// common.Log.Debug("Ever1: WhiteRun: %v", whiteRun)
				if !whiteRun {
					table = m.whiteTable
				} else {
					table = m.blackTable
				}

				// common.Log.Debug("UncompressGetCode")
				c, err = rd.uncompressGetCode(table)
				if err != nil {
					return 0, err
				}

				if c == nil {
					// common.Log.Debug("Nil code")
					break decodeLoop
				}

				// common.Log.Debug("Ever1 runLength: %v", c.runLength)
				rd.offset += c.bitLength
				if c.runLength < 64 {
					if c.runLength < 0 {
						runOffsets[currentBuffOffset] = currentLineBitPosition
						currentBuffOffset++
						break decodeLoop
					}

					currentLineBitPosition += c.runLength
					// don't generate 0-length run at EOL for cases where the line ends in an H-run
					if currentLineBitPosition < width ||
						currentLineBitPosition != firstHalfBitPost {

						runOffsets[currentBuffOffset] = currentLineBitPosition
						currentBuffOffset++
					}
					break ever1Loop
				}
				currentLineBitPosition += c.runLength
			}

			for currentLineBitPosition < width &&
				referenceOffsets[referenceBuffOffset] <= currentLineBitPosition {
				referenceBuffOffset += 2
			}
			continue decodeLoop

		case codeP:
			referenceBuffOffset++
			currentLineBitPosition = referenceOffsets[referenceBuffOffset]
			referenceBuffOffset++
			continue decodeLoop
		case codeVR2:
			currentLineBitPosition = referenceOffsets[referenceBuffOffset] + 2
		case codeVL2:
			currentLineBitPosition = referenceOffsets[referenceBuffOffset] - 2
		case codeVR3:
			currentLineBitPosition = referenceOffsets[referenceBuffOffset] + 3
		case codeVL3:
			currentLineBitPosition = referenceOffsets[referenceBuffOffset] - 3
		default:
			// common.Log.Debug("Unknown MMR Code type: %v", c.runLength)

			// Possibly MMR Decoded
			if rd.offset == 12 && c.runLength == eol {
				rd.offset = 0
				m.Uncompress1d(rd, referenceOffsets, width)
				rd.offset++
				m.Uncompress1d(rd, runOffsets, width)
				retCode, err := m.Uncompress1d(rd, referenceOffsets, width)
				if err != nil {
					return eof, err
				}
				rd.offset++
				return retCode, nil
			}
			currentLineBitPosition = width
			continue decodeLoop
		}

		if currentLineBitPosition <= width {
			whiteRun = !whiteRun

			runOffsets[currentBuffOffset] = currentLineBitPosition
			currentBuffOffset++

			if referenceBuffOffset > 0 {
				referenceBuffOffset--
			} else {
				referenceBuffOffset++
			}

			for currentLineBitPosition < width &&
				referenceOffsets[referenceBuffOffset] <= currentLineBitPosition {

				referenceBuffOffset += 2
			}
		}
	}

	if runOffsets[currentBuffOffset] != width {
		runOffsets[currentBuffOffset] = width
	}

	// common.Log.Debug("Code is nil.")

	if c == nil {
		// common.Log.Debug("EOL")
		return eol, nil
	}
	return currentBuffOffset, nil
}

// Uncompress1d decopmresses 1 row of data
func (m *Decoder) Uncompress1d(data *runData, runOffsets []int, width int) (int, error) {

	var (
		whiteRun  = true
		iBitPos   int
		cd        *code
		refOffset int
		err       error
	)

outer:
	for iBitPos < width {

	inner:
		for {
			if whiteRun {
				// common.Log.Debug("White table")
				cd, err = data.uncompressGetCode(m.whiteTable)
				if err != nil {
					return 0, err
				}
			} else {
				// common.Log.Debug("White table")
				cd, err = data.uncompressGetCode(m.blackTable)
				if err != nil {
					return 0, err
				}
			}
			data.offset += cd.bitLength

			if cd.runLength < 0 {
				break outer
			}

			iBitPos += cd.runLength

			if cd.runLength < 64 {
				whiteRun = !whiteRun
				runOffsets[refOffset] = iBitPos
				refOffset++
				break inner
			}
		}
	}

	if runOffsets[refOffset] != width {
		runOffsets[refOffset] = width
	}
	var result int
	if cd != nil && cd.runLength != eol {
		result = refOffset
	} else {
		result = eol
	}
	return result, nil
}

func (m *Decoder) createLittleEndianTable(codes [][3]int) ([]*code, error) {

	var firstLevelTable = make([]*code, firstLevelTablemask+1)

	// common.Log.Debug("CodesLength: %v", len(codes))
	for i := 0; i < len(codes); i++ {
		// common.Log.Debug("I: %v", i)
		var cd = newCode(codes[i])

		if cd.bitLength <= firstLevelTableSize {

			variantLength := firstLevelTableSize - cd.bitLength
			// common.Log.Debug("CodeWord: %v", cd.codeWord)
			baseWord := cd.codeWord << uint(variantLength)
			// common.Log.Debug("BaseWord: %v VariantLength %v", baseWord, variantLength)
			for variant := (1 << uint(variantLength)) - 1; variant >= 0; variant-- {
				index := baseWord | variant
				// common.Log.Debug("Variant: %v, index: %v", variant, index)
				firstLevelTable[index] = cd
			}

		} else {

			firstLevelIndex := cd.codeWord >> uint(cd.bitLength-firstLevelTableSize)
			// common.Log.Debug("SL: FirstLevelIndex: %v", firstLevelIndex)

			if firstLevelTable[firstLevelIndex] == nil {

				var firstLevelCode = newCode([3]int{})
				firstLevelCode.subTable = make([]*code, secondLevelTableMask+1)

				firstLevelTable[firstLevelIndex] = firstLevelCode
			}

			if cd.bitLength <= firstLevelTableSize+secondLevelTableSize {

				// var secondLevelTable []*code = firstLevelTable[firstLevelIndex].subTable

				variantLength := firstLevelTableSize + secondLevelTableSize - cd.bitLength
				baseWord := (cd.codeWord << uint(variantLength)) & secondLevelTableMask
				// common.Log.Debug("SL BaseWord: %v VariantLength %v", baseWord, variantLength)
				firstLevelTable[firstLevelIndex].nonNilSubTable = true
				for variant := (1 << uint(variantLength)) - 1; variant >= 0; variant-- {
					// common.Log.Debug("Variant: %v", variant)
					firstLevelTable[firstLevelIndex].subTable[baseWord|variant] = cd
				}
			} else {
				return nil, errors.New("Code table overflow in MMRDecoder")
			}
		}
	}
	// common.Log.Debug("%s", firstLevelTable)
	return firstLevelTable, nil
}

// DetectAndSkipEOL detects and skip the eond of line
func (m *Decoder) DetectAndSkipEOL() error {
	for {
		cd, err := m.data.uncompressGetCode(m.modeTable)
		if err != nil {
			return err
		}

		if cd != nil && cd.runLength == eol {
			m.data.offset += cd.bitLength
		} else {
			return nil
		}
	}
}

// UncompressMMR decompress the
func (m *Decoder) UncompressMMR() (b *bitmap.Bitmap, err error) {

	b = bitmap.New(m.width, m.height)

	// define offsets
	var (
		currentOffsets   = make([]int, b.Width+5)
		referenceOffsets = make([]int, b.Width+5)
	)

	referenceOffsets[0] = b.Width
	var refRunLength = 1

	count := 0

	// common.Log.Debug("Height: %v", b.Height)
	for line := 0; line < b.Height; line++ {
		common.Log.Debug("Line: %d", line)
		count, err = m.Uncompress2d(m.data, referenceOffsets, refRunLength, currentOffsets, b.Width)
		if err != nil {
			return
		}

		if count == EOF {
			break
		}

		if count > 0 {
			err = m.FillBitmap(b, line, currentOffsets, count)
			if err != nil {
				return
			}
		}
		// swap lines
		referenceOffsets, currentOffsets = currentOffsets, referenceOffsets
		refRunLength = count
	}
	if err = m.DetectAndSkipEOL(); err != nil {
		return
	}

	m.Data().Align()
	return
}

// FillBitmap fills the bitmap with the current line and offsets from the Decoder
func (m *Decoder) FillBitmap(b *bitmap.Bitmap, line int, currentOffsets []int, count int) error {
	x := 0
	// common.Log.Debug("Filling MMR. Line: %d", line)
	// common.Log.Debug("CurrentOffsets: %v. Count: %d", currentOffsets, count)

	targetByte := b.GetByteIndex(x, line)

	var targetByteValue byte

	for index := 0; index < count; index++ {
		// common.Log.Debug("Index: %d", index)
		offset := currentOffsets[index]

		// common.Log.Debug("Offset: %d", offset)
		var value byte

		if (index & 1) == 0 {
			// common.Log.Debug("Is zero")
			value = 0
		} else {
			// common.Log.Debug("Non zero")
			value = 1
		}

		for x < offset {
			targetByteValue = (targetByteValue << 1) | value
			x++

			if (x & 7) == 0 {

				if err := b.SetByte(targetByte, targetByteValue); err != nil {
					// common.Log.Debug("SetByte error: %v", err)
					return err
				}
				// // err := b.Data.SetByteBitwiseOffset(targetByteValue, count, targetBit, true)
				// if err != nil {

				// 	return err
				// }
				// b.Data.SetByteOffset(targetByteValue, targetByte)
				targetByte++
				targetByteValue = 0
			}
		}

		// common.Log.Debug("TargetByteValue: %08b", targetByteValue)
	}

	if (x & 7) != 0 {
		targetByteValue <<= uint(8 - (x & 7))

		if err := b.SetByte(targetByte, targetByteValue); err != nil {
			// common.Log.Debug("SetByte error: %v", err)
			return err
		}
		// the value here is in little endian manner
		// we should
		// common.Log.Debug("TargetByteValue at the end: %08b", targetByteValue)
		// b.Data.SetByteOffset(targetByteValue, targetByte)
		// err := b.Data.SetByteBitwiseOffset(targetByteValue, count, targetBit, true)
		// if err != nil {
		// common.Log.Debug("SetByteBitwiseOffset failed: %v", err)
		// 	return err
		// }
	}

	return nil

}
