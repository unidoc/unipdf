/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package mmr

import (
	"errors"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// Decoder is the jbig2 mmr data decoder.
type Decoder struct {
	width, height int
	data          *runData

	whiteTable []*code
	blackTable []*code
	modeTable  []*code
}

// New creates new jbig2 mmr decoder for the provided data stream.
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

// UncompressMMR decompress the stream value using MMR method. As a result returns the bitmap.Bitmap.
func (m *Decoder) UncompressMMR() (b *bitmap.Bitmap, err error) {
	b = bitmap.New(m.width, m.height)

	// define offsets
	currentOffsets := make([]int, b.Width+5)
	referenceOffsets := make([]int, b.Width+5)
	referenceOffsets[0] = b.Width

	refRunLength := 1
	var count int

	for line := 0; line < b.Height; line++ {
		count, err = m.uncompress2d(m.data, referenceOffsets, refRunLength, currentOffsets, b.Width)
		if err != nil {
			return nil, err
		}

		if count == EOF {
			break
		}

		if count > 0 {
			err = m.fillBitmap(b, line, currentOffsets, count)
			if err != nil {
				return nil, err
			}
		}

		// swap lines
		referenceOffsets, currentOffsets = currentOffsets, referenceOffsets
		refRunLength = count
	}
	if err = m.detectAndSkipEOL(); err != nil {
		return nil, err
	}

	m.data.align()
	return b, nil
}

func (m *Decoder) createLittleEndianTable(codes [][3]int) ([]*code, error) {
	firstLevelTable := make([]*code, firstLevelTablemask+1)

	for i := 0; i < len(codes); i++ {
		cd := newCode(codes[i])

		if cd.bitLength <= firstLevelTableSize {
			variantLength := firstLevelTableSize - cd.bitLength
			baseWord := cd.codeWord << uint(variantLength)
			for variant := (1 << uint(variantLength)) - 1; variant >= 0; variant-- {
				index := baseWord | variant
				firstLevelTable[index] = cd
			}
		} else {
			firstLevelIndex := cd.codeWord >> uint(cd.bitLength-firstLevelTableSize)
			if firstLevelTable[firstLevelIndex] == nil {
				var firstLevelCode = newCode([3]int{})
				firstLevelCode.subTable = make([]*code, secondLevelTableMask+1)
				firstLevelTable[firstLevelIndex] = firstLevelCode
			}

			if cd.bitLength <= firstLevelTableSize+secondLevelTableSize {
				variantLength := firstLevelTableSize + secondLevelTableSize - cd.bitLength
				baseWord := (cd.codeWord << uint(variantLength)) & secondLevelTableMask
				firstLevelTable[firstLevelIndex].nonNilSubTable = true

				for variant := (1 << uint(variantLength)) - 1; variant >= 0; variant-- {
					firstLevelTable[firstLevelIndex].subTable[baseWord|variant] = cd
				}
			} else {
				return nil, errors.New("Code table overflow in MMRDecoder")
			}
		}
	}
	return firstLevelTable, nil
}

// detectAndSkipEOL detects and skip the EOL.
func (m *Decoder) detectAndSkipEOL() error {
	for {
		cd, err := m.data.uncompressGetCode(m.modeTable)
		if err != nil {
			return err
		}

		if cd != nil && cd.runLength == EOL {
			m.data.offset += cd.bitLength
		} else {
			return nil
		}
	}
}

// fillBitmap fills the bitmap with the current line and offsets from the Decoder.
func (m *Decoder) fillBitmap(b *bitmap.Bitmap, line int, currentOffsets []int, count int) error {
	var targetByteValue byte
	x := 0
	targetByte := b.GetByteIndex(x, line)

	for index := 0; index < count; index++ {
		value := byte(1)
		offset := currentOffsets[index]

		if (index & 1) == 0 {
			value = 0
		}

		for x < offset {
			targetByteValue = (targetByteValue << 1) | value
			x++

			if (x & 7) == 0 {
				if err := b.SetByte(targetByte, targetByteValue); err != nil {
					return err
				}
				targetByte++
				targetByteValue = 0
			}
		}
	}

	if (x & 7) != 0 {
		targetByteValue <<= uint(8 - (x & 7))
		if err := b.SetByte(targetByte, targetByteValue); err != nil {
			return err
		}
	}
	return nil
}

func (m *Decoder) initTables() (err error) {
	if m.whiteTable == nil {
		m.whiteTable, err = m.createLittleEndianTable(whiteCodes)
		if err != nil {
			return
		}
		m.blackTable, err = m.createLittleEndianTable(blackCodes)
		if err != nil {
			return
		}
		m.modeTable, err = m.createLittleEndianTable(modeCodes)
		if err != nil {
			return
		}
	}
	return nil
}

func (m *Decoder) uncompress1d(data *runData, runOffsets []int, width int) (int, error) {
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
				cd, err = data.uncompressGetCode(m.whiteTable)
				if err != nil {
					return 0, err
				}
			} else {
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
	result := EOL
	if cd != nil && cd.runLength != EOL {
		result = refOffset
	}
	return result, nil
}

func (m *Decoder) uncompress2d(
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
		whiteRun               = true
		err                    error
		c                      *code
	)

	referenceOffsets[refRunLength] = width
	referenceOffsets[refRunLength+1] = width
	referenceOffsets[refRunLength+2] = width + 1
	referenceOffsets[refRunLength+3] = width + 1

decodeLoop:
	for currentLineBitPosition < width {
		// Get the mode code
		c, err = rd.uncompressGetCode(m.modeTable)
		if err != nil {
			return EOL, nil
		}

		if c == nil {
			rd.offset++
			break decodeLoop
		}
		rd.offset += c.bitLength

		switch mmrCode(c.runLength) {
		case codeV0:
			currentLineBitPosition = referenceOffsets[referenceBuffOffset]
		case codeVR1:
			currentLineBitPosition = referenceOffsets[referenceBuffOffset] + 1
		case codeVL1:
			currentLineBitPosition = referenceOffsets[referenceBuffOffset] - 1
		case codeH:
			ever := 1
			for ever > 0 {
				var table []*code
				if whiteRun {
					table = m.whiteTable
				} else {
					table = m.blackTable
				}

				c, err = rd.uncompressGetCode(table)
				if err != nil {
					return 0, err
				}

				if c == nil {
					break decodeLoop
				}

				rd.offset += c.bitLength

				if c.runLength < 64 {
					if c.runLength < 0 {
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
			ever1 := 1

		ever1Loop:
			for ever1 > 0 {
				var table []*code

				if !whiteRun {
					table = m.whiteTable
				} else {
					table = m.blackTable
				}

				c, err = rd.uncompressGetCode(table)
				if err != nil {
					return 0, err
				}
				if c == nil {
					break decodeLoop
				}

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

			for currentLineBitPosition < width && referenceOffsets[referenceBuffOffset] <= currentLineBitPosition {
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
			// Possibly MMR Decoded
			if rd.offset == 12 && c.runLength == EOL {
				rd.offset = 0
				if _, err = m.uncompress1d(rd, referenceOffsets, width); err != nil {
					return 0, err
				}
				rd.offset++
				if _, err = m.uncompress1d(rd, runOffsets, width); err != nil {
					return 0, err
				}
				retCode, err := m.uncompress1d(rd, referenceOffsets, width)
				if err != nil {
					return EOF, err
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

			for currentLineBitPosition < width && referenceOffsets[referenceBuffOffset] <= currentLineBitPosition {
				referenceBuffOffset += 2
			}
		}
	}

	if runOffsets[currentBuffOffset] != width {
		runOffsets[currentBuffOffset] = width
	}

	if c == nil {
		return EOL, nil
	}
	return currentBuffOffset, nil
}
