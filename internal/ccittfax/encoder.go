/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ccittfax

import (
	"math"
)

var (
	white byte = 1
	black byte = 0
)

// Encoder implements a CCITT facsimile (fax) Group3 and Group4 encoding/decoding algorithms.
type Encoder struct {
	K                      int
	EndOfLine              bool
	EncodedByteAlign       bool
	Columns                int
	Rows                   int
	EndOfBlock             bool
	BlackIs1               bool
	DamagedRowsBeforeError int
}

// Encode encodes the original image pixels.
// Note: `pixels` here are the pixels of the image where each byte value is 1 for white
// pixels and 0 for the black ones.
func (e *Encoder) Encode(pixels [][]byte) []byte {
	if e.BlackIs1 {
		white = 0
		black = 1
	} else {
		white = 1
		black = 0
	}

	if e.K == 0 {
		// do Group3 1-dimensional encoding
		return e.encodeG31D(pixels)
	}

	if e.K > 0 {
		// do Group3 mixed (1D/2D) encoding
		return e.encodeG32D(pixels)
	}

	if e.K < 0 {
		// do Group4 encoding
		return e.encodeG4(pixels)
	}

	// never happens
	return nil
}

// encodeG31D encodes the image data using the CCITT facsimile (fax) Group3 1-dimensional encoding.
func (e *Encoder) encodeG31D(pixels [][]byte) []byte {
	var encoded []byte

	prevBitPos := 0
	for i := range pixels {
		if e.Rows > 0 && !e.EndOfBlock && i == e.Rows {
			break
		}

		encodedRow, bitPos := encodeRow1D(pixels[i], prevBitPos, eol)
		encoded = e.appendEncodedRow(encoded, encodedRow, prevBitPos)

		if e.EncodedByteAlign {
			// align to the byte border
			bitPos = 0
		}

		prevBitPos = bitPos
	}

	if e.EndOfBlock {
		// put RTC
		encodedRTC, _ := encodeRTC(prevBitPos)
		encoded = e.appendEncodedRow(encoded, encodedRTC, prevBitPos)
	}

	return encoded
}

// encodeG32D encodes the image data using the CCITT facsimile (fax) Group3 mixed (1D/2D) dimensional encoding.
func (e *Encoder) encodeG32D(pixels [][]byte) []byte {
	var encoded []byte
	var prevBitPos int
	for i := 0; i < len(pixels); i += e.K {
		if e.Rows > 0 && !e.EndOfBlock && i == e.Rows {
			break
		}

		// encode 1st of K rows 1-dimensionally
		encodedRow, bitPos := encodeRow1D(pixels[i], prevBitPos, eol1)

		encoded = e.appendEncodedRow(encoded, encodedRow, prevBitPos)

		if e.EncodedByteAlign {
			// align to byte border
			bitPos = 0
		}

		prevBitPos = bitPos

		// encode the rest K-1 rows 2-dimensionally
		for j := i + 1; j < (i+e.K) && j < len(pixels); j++ {
			if e.Rows > 0 && !e.EndOfBlock && j == e.Rows {
				break
			}

			encodedRow, bitPos := addCode(nil, prevBitPos, eol0)

			var a1, b1, b2 int

			a0 := -1

			for a0 < len(pixels[j]) {
				a1 = seekChangingElem(pixels[j], a0)
				b1 = seekB1(pixels[j], pixels[j-1], a0)
				b2 = seekChangingElem(pixels[j-1], b1)

				if b2 < a1 {
					// do pass mode
					encodedRow, bitPos = encodePassMode(encodedRow, bitPos)

					a0 = b2
				} else {
					if math.Abs(float64(b1-a1)) > 3 {
						// do horizontal mode
						encodedRow, bitPos, a0 = encodeHorizontalMode(pixels[j], encodedRow, bitPos, a0, a1)
					} else {
						// do vertical mode
						encodedRow, bitPos = encodeVerticalMode(encodedRow, bitPos, a1, b1)

						a0 = a1
					}
				}
			}

			encoded = e.appendEncodedRow(encoded, encodedRow, prevBitPos)

			if e.EncodedByteAlign {
				//align to byte border
				bitPos = 0
			}

			prevBitPos = bitPos % 8
		}
	}

	if e.EndOfBlock {
		// put RTC
		encodedRTC, _ := encodeRTC2D(prevBitPos)

		encoded = e.appendEncodedRow(encoded, encodedRTC, prevBitPos)
	}

	return encoded
}

// encodeG4 encodes the image data using the CCITT facsimile (fax) Group4 encoding.
func (e *Encoder) encodeG4(pixelsToEncode [][]byte) []byte {
	// copy the outer slice of pixels in order to avoid the modification of the original data
	pixels := make([][]byte, len(pixelsToEncode))
	copy(pixels, pixelsToEncode)
	// append white reference line
	pixels = appendWhiteReferenceLine(pixels)

	var encoded []byte
	var prevBitPos int
	for i := 1; i < len(pixels); i++ {
		if e.Rows > 0 && !e.EndOfBlock && i == (e.Rows+1) {
			break
		}

		var encodedRow []byte
		var a1, b1, b2 int

		bitPos := prevBitPos
		a0 := -1

		for a0 < len(pixels[i]) {
			a1 = seekChangingElem(pixels[i], a0)
			b1 = seekB1(pixels[i], pixels[i-1], a0)
			b2 = seekChangingElem(pixels[i-1], b1)

			if b2 < a1 {
				// do pass mode
				encodedRow, bitPos = addCode(encodedRow, bitPos, p)

				a0 = b2
			} else {
				if math.Abs(float64(b1-a1)) > 3 {
					encodedRow, bitPos, a0 = encodeHorizontalMode(pixels[i], encodedRow, bitPos, a0, a1)
				} else {
					// do vertical mode
					encodedRow, bitPos = encodeVerticalMode(encodedRow, bitPos, a1, b1)

					a0 = a1
				}
			}
		}

		encoded = e.appendEncodedRow(encoded, encodedRow, prevBitPos)
		if e.EncodedByteAlign {
			// align to byte border
			bitPos = 0
		}

		prevBitPos = bitPos % 8
	}

	if e.EndOfBlock {
		// put EOFB
		encodedEOFB, _ := encodeEOFB(prevBitPos)
		encoded = e.appendEncodedRow(encoded, encodedEOFB, prevBitPos)
	}

	return encoded
}

// encodeRTC encodes the RTC code for the G3 1D (6 EOL in a row). `bitPos` is equal to the one in
// `encodeRow`.
func encodeRTC(bitPos int) ([]byte, int) {
	var encoded []byte

	for i := 0; i < 6; i++ {
		encoded, bitPos = addCode(encoded, bitPos, eol)
	}

	return encoded, bitPos % 8
}

// encodeRTC2D encodes the RTC code for the G3 2D (6 EOL + 1 bit in a row). `bitPos` is equal to the one in
// `encodeRow`.
func encodeRTC2D(bitPos int) ([]byte, int) {
	var encoded []byte

	for i := 0; i < 6; i++ {
		encoded, bitPos = addCode(encoded, bitPos, eol1)
	}

	return encoded, bitPos % 8
}

// encodeEOFB encodes the EOFB code for the G4 (2 EOL in a row). `bitPos` is equal to the one in
// `encodeRow`.
func encodeEOFB(bitPos int) ([]byte, int) {
	var encoded []byte

	for i := 0; i < 2; i++ {
		encoded, bitPos = addCode(encoded, bitPos, eol)
	}

	return encoded, bitPos % 8
}

// encodeRow1D encodes single row of the image pixels in 1D mode. `bitPos` is the bit position
// global for the `row` array. `bitPos` is used to indicate where to start the
// encoded sequences. It is used for the EncodedByteAlign option implementation.
// Returns the encoded data and the number of the bits taken from the last byte.
func encodeRow1D(row []byte, bitPos int, prefix code) ([]byte, int) {
	// always start with whites
	isWhite := true
	var encoded []byte

	// always add EOL before the scan line
	encoded, bitPos = addCode(nil, bitPos, prefix)

	bytePos := 0
	var runLen int
	for bytePos < len(row) {
		runLen, bytePos = calcRun(row, isWhite, bytePos)

		encoded, bitPos = encodeRunLen(encoded, bitPos, runLen, isWhite)

		// switch color
		isWhite = !isWhite
	}

	return encoded, bitPos % 8
}

// encodeRunLen writes the calculated run length to the `encoded` starting with `bitPos` bit.
func encodeRunLen(encoded []byte, bitPos int, runLen int, isWhite bool) ([]byte, int) {
	var (
		code       code
		isTerminal bool
	)

	for !isTerminal {
		code, runLen, isTerminal = getRunCode(runLen, isWhite)
		encoded, bitPos = addCode(encoded, bitPos, code)
	}

	return encoded, bitPos
}

// getRunCode gets the code for the specified run. If the code is not
// terminal, returns the remainder to be determined later. Otherwise
// returns 0 remainder. Also returns the bool flag to indicate if
// the code is terminal.
func getRunCode(runLen int, isWhite bool) (code, int, bool) {
	if runLen < 64 {
		if isWhite {
			return wTerms[runLen], 0, true
		} else {
			return bTerms[runLen], 0, true
		}
	} else {
		multiplier := runLen / 64

		// stands for lens more than 2560 which are not
		// covered by the Huffman codes
		if multiplier > 40 {
			return commonMakeups[2560], runLen - 2560, false
		}

		// stands for lens more than 1728. These should be common
		// for both colors
		if multiplier > 27 {
			return commonMakeups[multiplier*64], runLen - multiplier*64, false
		}

		// for lens < 27 we use the specific makeups for each color
		if isWhite {
			return wMakeups[multiplier*64], runLen - multiplier*64, false
		} else {
			return bMakeups[multiplier*64], runLen - multiplier*64, false
		}
	}
}

// calcRun calculates the nex pixel run. Returns the number of the
// pixels and the new position in the original array.
func calcRun(row []byte, isWhite bool, pos int) (int, int) {
	count := 0
	for pos < len(row) {
		if isWhite {
			if row[pos] != white {
				break
			}
		} else {
			if row[pos] != black {
				break
			}
		}

		count++
		pos++
	}

	return count, pos
}

// addCode writes the specified `code` to the `encoded` starting with `pos` bit. `pos`
// bit is a bit position global to the whole encoded array.
func addCode(encoded []byte, pos int, code code) ([]byte, int) {
	i := 0
	for i < code.BitsWritten {
		bytePos := pos / 8
		bitPos := pos % 8

		if bytePos >= len(encoded) {
			encoded = append(encoded, 0)
		}

		toWrite := 8 - bitPos
		leftToWrite := code.BitsWritten - i
		if toWrite > leftToWrite {
			toWrite = leftToWrite
		}

		if i < 8 {
			encoded[bytePos] = encoded[bytePos] | byte(code.Code>>uint(8+bitPos-i))&masks[8-toWrite-bitPos]
		} else {
			encoded[bytePos] = encoded[bytePos] | (byte(code.Code<<uint(i-8))&masks[8-toWrite])>>uint(bitPos)
		}

		pos += toWrite

		i += toWrite
	}

	return encoded, pos
}

// appendEncodedRow appends the newly encoded row to the array of the encoded data.
// `bitPos` is a bit position in the last byte of the encoded data. `bitPos` points where to
// write the next piece of data.
func (e *Encoder) appendEncodedRow(encoded, newRow []byte, bitPos int) []byte {
	if len(encoded) > 0 && bitPos != 0 && !e.EncodedByteAlign {
		encoded[len(encoded)-1] = encoded[len(encoded)-1] | newRow[0]

		encoded = append(encoded, newRow[1:]...)
	} else {
		encoded = append(encoded, newRow...)
	}

	return encoded
}

// seekChangingElem gets the position of the changing elem in the `row` based on
// the position of the `currElem`.
func seekChangingElem(row []byte, currElem int) int {
	if currElem >= len(row) {
		return currElem
	}
	if currElem < -1 {
		// A current element of -1 means white color starting prior to first element (0).
		// Should not be able to go any further left than that.
		currElem = -1
	}

	var color byte
	if currElem > -1 {
		color = row[currElem]
	} else {
		color = white
	}

	i := currElem + 1
	for i < len(row) {
		if row[i] != color {
			break
		}

		i++
	}

	return i
}

// seekB1 gets the position of b1 based on the position of `a0`.
func seekB1(codingLine, refLine []byte, a0 int) int {
	changingElem := seekChangingElem(refLine, a0)

	if changingElem < len(refLine) && (a0 == -1 && refLine[changingElem] == white ||
		a0 >= 0 && a0 < len(codingLine) && codingLine[a0] == refLine[changingElem] ||
		a0 >= len(codingLine) && codingLine[a0-1] != refLine[changingElem]) {
		changingElem = seekChangingElem(refLine, changingElem)
	}

	return changingElem
}

// seekB12D gets the position of b1 based on the position of `a0`.
// Note: Used for the Group4 encoding.
func seekB12D(codingLine, refLine []byte, a0 int, a0isWhite bool) int {
	changingElem := seekChangingElem(refLine, a0)

	if changingElem < len(refLine) && (a0 == -1 && refLine[changingElem] == white ||
		a0 >= 0 && a0 < len(codingLine) && codingLine[a0] == refLine[changingElem] ||
		a0 >= len(codingLine) && a0isWhite && refLine[changingElem] == white ||
		a0 >= len(codingLine) && !a0isWhite && refLine[changingElem] == black) {
		changingElem = seekChangingElem(refLine, changingElem)
	}

	return changingElem
}

// encodePassMode encodes the pass mode to the `encodedRow` starting with `bitPos` bit.
func encodePassMode(encodedRow []byte, bitPos int) ([]byte, int) {
	return addCode(encodedRow, bitPos, p)
}

// encodeHorizontalMode encodes the horizontal mode to the `encodedRow` starting with `bitPos` bit.
func encodeHorizontalMode(row, encodedRow []byte, bitPos, a0, a1 int) ([]byte, int, int) {
	a2 := seekChangingElem(row, a1)

	isWhite := a0 >= 0 && row[a0] == white || a0 == -1

	encodedRow, bitPos = addCode(encodedRow, bitPos, h)
	var a0a1RunLen int
	if a0 > -1 {
		a0a1RunLen = a1 - a0
	} else {
		a0a1RunLen = a1 - a0 - 1
	}

	encodedRow, bitPos = encodeRunLen(encodedRow, bitPos, a0a1RunLen, isWhite)

	isWhite = !isWhite

	a1a2RunLen := a2 - a1

	encodedRow, bitPos = encodeRunLen(encodedRow, bitPos, a1a2RunLen, isWhite)

	a0 = a2

	return encodedRow, bitPos, a0
}

// encodeVerticalMode encodes vertical mode to the encodedRow starting with `bitPos` bit.
func encodeVerticalMode(encodedRow []byte, bitPos, a1, b1 int) ([]byte, int) {
	vCode := getVCode(a1, b1)

	encodedRow, bitPos = addCode(encodedRow, bitPos, vCode)

	return encodedRow, bitPos
}

// getVCode get the code for the vertical mode based on
// the locations of `a1` and `b1`.
func getVCode(a1, b1 int) code {
	var vCode code

	switch b1 - a1 {
	case -1:
		vCode = v1r
	case -2:
		vCode = v2r
	case -3:
		vCode = v3r
	case 0:
		vCode = v0
	case 1:
		vCode = v1l
	case 2:
		vCode = v2l
	case 3:
		vCode = v3l
	}

	return vCode
}

// appendWhiteReferenceLine appends the line full of white pixels just before
// the first image line.
func appendWhiteReferenceLine(pixels [][]byte) [][]byte {
	whiteRefLine := make([]byte, len(pixels[0]))
	for i := range whiteRefLine {
		whiteRefLine[i] = white
	}

	pixels = append(pixels, []byte{})
	for i := len(pixels) - 1; i > 0; i-- {
		pixels[i] = pixels[i-1]
	}

	pixels[0] = whiteRefLine

	return pixels
}
