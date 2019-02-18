/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ccittfax

import (
	"errors"
)

var (
	// errEOFBCorrupt is returned when the corrupt EOFB (end-of-block) code is found.
	errEOFBCorrupt = errors.New("EOFB code is corrupted")
	// errRTCCorrupt is returned when the corrupt RTC (return-the-carriage) code is found.
	errRTCCorrupt = errors.New("RTC code is corrupted")
	// errWrongCodeInHorizontalMode is returned when entered the horizontal mode and unknown bit
	// sequence met.
	errWrongCodeInHorizontalMode = errors.New("wrong code in horizontal mode")
	// errNoEOLFound is returned when the EndOfLine parameter is true in filter but no EOL (end-of-line) met.
	errNoEOLFound = errors.New("no EOL found while the EndOfLine parameter is true")
	// errInvalidEOL is returned when the EOL code is corrupt.
	errInvalidEOL = errors.New("invalid EOL")
	// errInvalid2DCode is returned when the invalid 2 dimensional code is met. 2 dimensional code
	// according to the CCITT reccommendations is one of the following: H, P, V0, V1L, V2L, V3L, V1R, V2R, V3R.
	errInvalid2DCode = errors.New("invalid 2D code")
)

// trees represent the finite state machine for parsing bit sequences and fetching pixel run lengths
var (
	whiteTree = &decodingTreeNode{
		Val: 255,
	}
	blackTree = &decodingTreeNode{
		Val: 255,
	}
	twoDimTree = &decodingTreeNode{
		Val: 255,
	}
)

func init() {
	for runLen, code := range wTerms {
		addNode(whiteTree, code, 0, runLen)
	}

	for runLen, code := range wMakeups {
		addNode(whiteTree, code, 0, runLen)
	}

	for runLen, code := range bTerms {
		addNode(blackTree, code, 0, runLen)
	}

	for runLen, code := range bMakeups {
		addNode(blackTree, code, 0, runLen)
	}

	for runLen, code := range commonMakeups {
		addNode(whiteTree, code, 0, runLen)
		addNode(blackTree, code, 0, runLen)
	}

	addNode(twoDimTree, p, 0, 0)
	addNode(twoDimTree, h, 0, 0)
	addNode(twoDimTree, v0, 0, 0)
	addNode(twoDimTree, v1r, 0, 0)
	addNode(twoDimTree, v2r, 0, 0)
	addNode(twoDimTree, v3r, 0, 0)
	addNode(twoDimTree, v1l, 0, 0)
	addNode(twoDimTree, v2l, 0, 0)
	addNode(twoDimTree, v3l, 0, 0)
}

// Decode performs decoding operation on the encoded image using the Group3 or Group4
// CCITT facsimile (fax) algorithm.
func (e *Encoder) Decode(encoded []byte) ([][]byte, error) {
	if e.BlackIs1 {
		white = 0
		black = 1
	} else {
		white = 1
		black = 0
	}

	if e.K == 0 {
		// decode Group3 1-dimensional
		return e.decodeG31D(encoded)
	}

	if e.K > 0 {
		// decode Group3 mixed dimensional (1D/2D)
		return e.decodeG32D(encoded)
	}

	if e.K < 4 {
		// decode Group4
		return e.decodeG4(encoded)
	}

	return nil, nil
}

// decodeG31D decodes the encoded image data using the Group3 1-dimensional
// CCITT facsimile (fax) decoding algorithm.
func (e *Encoder) decodeG31D(encoded []byte) ([][]byte, error) {
	var pixels [][]byte

	// do g31d decoding
	var bitPos int
	for (bitPos / 8) < len(encoded) {
		var gotEOL bool

		gotEOL, bitPos = tryFetchEOL(encoded, bitPos)
		if !gotEOL {
			if e.EndOfLine {
				return nil, errNoEOLFound
			}
		} else {
			// 5 EOLs left to fill RTC
			for i := 0; i < 5; i++ {
				gotEOL, bitPos = tryFetchEOL(encoded, bitPos)

				if !gotEOL {
					if i == 0 {
						break
					}

					return nil, errInvalidEOL
				}
			}

			// got RTC
			if gotEOL {
				break
			}
		}

		var row []byte
		row, bitPos = e.decodeRow1D(encoded, bitPos)

		if e.EncodedByteAlign && bitPos%8 != 0 {
			// align to byte border
			bitPos += 8 - bitPos%8
		}

		pixels = append(pixels, row)

		if e.Rows > 0 && !e.EndOfBlock && len(pixels) >= e.Rows {
			break
		}
	}

	return pixels, nil
}

// decodeG32D decodes the encoded image data using the Group3 mixed (1D/2D) dimensional
// CCITT facsimile (fax) decoding algorithm.
func (e *Encoder) decodeG32D(encoded []byte) ([][]byte, error) {
	var (
		pixels [][]byte
		bitPos int
		err    error
	)
byteLoop:
	for (bitPos / 8) < len(encoded) {
		var gotEOL bool
		gotEOL, bitPos, err = tryFetchRTC2D(encoded, bitPos)
		if err != nil {
			return nil, err
		}

		if gotEOL {
			break
		}

		gotEOL, bitPos = tryFetchEOL1(encoded, bitPos)

		if !gotEOL {
			if e.EndOfLine {
				return nil, errNoEOLFound
			}
		}

		// decode 1st of K rows as 1D
		var row []byte
		row, bitPos = e.decodeRow1D(encoded, bitPos)

		if e.EncodedByteAlign && bitPos%8 != 0 {
			// align to byte border
			bitPos += 8 - bitPos%8
		}

		if row != nil {
			pixels = append(pixels, row)
		}

		if e.Rows > 0 && !e.EndOfBlock && len(pixels) >= e.Rows {
			break
		}

		// decode K-1 rows as 2D
		for i := 1; i < e.K && (bitPos/8) < len(encoded); i++ {
			gotEOL, bitPos = tryFetchEOL0(encoded, bitPos)
			if !gotEOL {
				// only EOL0 or RTC should be met here. If neither of these met -
				// the data is considered corrupt
				gotEOL, bitPos, err = tryFetchRTC2D(encoded, bitPos)
				if err != nil {
					return nil, err
				}

				if gotEOL {
					break byteLoop
				} else {
					if e.EndOfLine {
						return nil, errNoEOLFound
					}
				}
			}

			var (
				twoDimCode code
				ok         bool
			)

			isWhite := true
			var pixelsRow []byte
			a0 := -1
			for twoDimCode, bitPos, ok = fetchNext2DCode(encoded, bitPos); ok; twoDimCode, bitPos, ok = fetchNext2DCode(encoded, bitPos) {
				switch twoDimCode {
				case p:
					// do pass mode decoding
					pixelsRow, a0 = decodePassMode(pixels, pixelsRow, isWhite, a0)
				case h:
					// do horizontal mode decoding
					pixelsRow, bitPos, a0, err = decodeHorizontalMode(encoded, pixelsRow, bitPos, isWhite, a0)
					if err != nil {
						return nil, err
					}
				case v0:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 0)
					isWhite = !isWhite
				case v1r:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 1)
					isWhite = !isWhite
				case v2r:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 2)
					isWhite = !isWhite
				case v3r:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 3)
					isWhite = !isWhite
				case v1l:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -1)
					isWhite = !isWhite
				case v2l:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -2)
					isWhite = !isWhite
				case v3l:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -3)
					isWhite = !isWhite
				}

				if len(pixelsRow) >= e.Columns {
					break
				}
			}

			if e.EncodedByteAlign && bitPos%8 != 0 {
				// align to byte border
				bitPos += 8 - bitPos%8
			}

			if pixelsRow != nil {
				pixels = append(pixels, pixelsRow)
			}

			if e.Rows > 0 && !e.EndOfBlock && len(pixels) >= e.Rows {
				break byteLoop
			}
		}
	}

	return pixels, nil
}

// decodeG4 decodes the encoded image data using the Group4 CCITT facsimile (fax)
// decoding algorithm.
func (e *Encoder) decodeG4(encoded []byte) ([][]byte, error) {
	// append white reference line
	whiteReferenceLine := make([]byte, e.Columns)
	for i := range whiteReferenceLine {
		whiteReferenceLine[i] = white
	}

	pixels := make([][]byte, 1)
	pixels[0] = whiteReferenceLine

	var (
		gotEOL bool
		err    error
		bitPos int
	)
	for (bitPos / 8) < len(encoded) {
		// try get EOFB
		gotEOL, bitPos, err = tryFetchEOFB(encoded, bitPos)
		if err != nil {
			return nil, err
		}
		if gotEOL {
			break
		}

		var (
			twoDimCode code
			ok         bool
		)

		isWhite := true
		var pixelsRow []byte
		a0 := -1
		for a0 < e.Columns {
			twoDimCode, bitPos, ok = fetchNext2DCode(encoded, bitPos)
			if !ok {
				return nil, errInvalid2DCode
			}

			switch twoDimCode {
			case p:
				// do pass mode decoding
				pixelsRow, a0 = decodePassMode(pixels, pixelsRow, isWhite, a0)
			case h:
				// do horizontal mode decoding
				pixelsRow, bitPos, a0, err = decodeHorizontalMode(encoded, pixelsRow, bitPos, isWhite, a0)
				if err != nil {
					return nil, err
				}
			case v0:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 0)
				isWhite = !isWhite
			case v1r:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 1)
				isWhite = !isWhite
			case v2r:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 2)
				isWhite = !isWhite
			case v3r:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 3)
				isWhite = !isWhite
			case v1l:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -1)
				isWhite = !isWhite
			case v2l:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -2)
				isWhite = !isWhite
			case v3l:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -3)
				isWhite = !isWhite
			}

			if len(pixelsRow) >= e.Columns {
				break
			}
		}

		if e.EncodedByteAlign && bitPos%8 != 0 {
			// align to byte border
			bitPos += 8 - bitPos%8
		}

		pixels = append(pixels, pixelsRow)

		if e.Rows > 0 && !e.EndOfBlock && len(pixels) >= (e.Rows+1) {
			break
		}
	}

	// remove the white reference line
	pixels = pixels[1:]

	return pixels, nil
}

// decodeVerticalMode decodes the part of data using the vertical mode. Returns the moved `a0` and the
// pixels row filled with the decoded pixels.
func decodeVerticalMode(pixels [][]byte, pixelsRow []byte, isWhite bool, a0, shift int) ([]byte, int) {
	b1 := seekB12D(pixelsRow, pixels[len(pixels)-1], a0, isWhite)
	// true for V0
	a1 := b1 + shift

	if a0 == -1 {
		pixelsRow = drawPixels(pixelsRow, isWhite, a1-a0-1)
	} else {
		pixelsRow = drawPixels(pixelsRow, isWhite, a1-a0)
	}

	a0 = a1

	return pixelsRow, a0
}

// decodePassMode decodes the part of data using the pass mode. Returns the moved `a0` and the
// pixels row filled with the decoded pixels.
func decodePassMode(pixels [][]byte, pixelsRow []byte, isWhite bool, a0 int) ([]byte, int) {
	b1 := seekB12D(pixelsRow, pixels[len(pixels)-1], a0, isWhite)
	b2 := seekChangingElem(pixels[len(pixels)-1], b1)

	if a0 == -1 {
		pixelsRow = drawPixels(pixelsRow, isWhite, b2-a0-1)
	} else {
		pixelsRow = drawPixels(pixelsRow, isWhite, b2-a0)
	}

	a0 = b2

	return pixelsRow, a0
}

// decodeHorizontalMode decodes the part of data using the horizontal mode. Returns the moved `a0`, the moved
// global bit position and the pixels row filled with the decoded pixels. The returned bit position is not
// moved is the error occurs.
func decodeHorizontalMode(encoded, pixelsRow []byte, bitPos int, isWhite bool, a0 int) ([]byte, int, int, error) {
	startingBitPos := bitPos

	var err error
	pixelsRow, bitPos, err = decodeNextRunLen(encoded, pixelsRow, bitPos, isWhite)
	if err != nil {
		return pixelsRow, startingBitPos, a0, err
	}

	isWhite = !isWhite

	pixelsRow, bitPos, err = decodeNextRunLen(encoded, pixelsRow, bitPos, isWhite)
	if err != nil {
		return pixelsRow, startingBitPos, a0, err
	}

	// the last code was the code of a1a2 run. a0 was moved to a2
	// while encoding. that's why we put a0 on the a2 and continue
	a0 = len(pixelsRow)

	return pixelsRow, bitPos, a0, nil
}

// decodeNextRunLen decodes tries to decode the next part of data using the Group3 1-dimensional code.
// Returns moved bit position and the pixels row filled with the decoded pixels. The returned bit position
// is not moved if the error occurs. Returns `errWrongCodeInHorizontalMode` if none of the 1-dimensional codes found.
func decodeNextRunLen(encoded, pixelsRow []byte, bitPos int, isWhite bool) ([]byte, int, error) {
	startingBitPos := bitPos

	var runLen int

	for runLen, bitPos = fetchNextRunLen(encoded, bitPos, isWhite); runLen != -1; runLen, bitPos = fetchNextRunLen(encoded, bitPos, isWhite) {
		pixelsRow = drawPixels(pixelsRow, isWhite, runLen)

		// got terminal code, switch color
		if runLen < 64 {
			break
		}
	}

	if runLen == -1 {
		return pixelsRow, startingBitPos, errWrongCodeInHorizontalMode
	}

	return pixelsRow, bitPos, nil
}

// tryFetchRTC2D tries to fetch the RTC code (0000000000011 X 6) for Group3 mixed (1D/2D) dimensional encoding from
// the encoded data. Returns the moved bit position if the code was found. The other way returns the
// the original bit position. The `errRTCCorrupt` is returned if the RTC code is corrupt. The RTC code is considered
// corrupt if there are more than one EOL1 code (0000000000011) is met.
func tryFetchRTC2D(encoded []byte, bitPos int) (bool, int, error) {
	startingBitPos := bitPos
	var gotEOL = false

	// 6 EOL1s to fill RTC
	for i := 0; i < 6; i++ {
		gotEOL, bitPos = tryFetchEOL1(encoded, bitPos)

		if !gotEOL {
			if i > 1 {
				return false, startingBitPos, errRTCCorrupt
			} else {
				bitPos = startingBitPos

				break
			}
		}
	}

	return gotEOL, bitPos, nil
}

// tryFetchEOFB tries to fetch the EOFB code (000000000001 X 2) for Group4 encoding from
// the encoded data. Returns the moved bit position if the code was found. The other way returns the
// the original bit position. The `errEOFBCorrupt` is returned if the EOFB code is corrupt. The EOFB code is considered
// corrupt if there is a single EOL code (000000000001).
func tryFetchEOFB(encoded []byte, bitPos int) (bool, int, error) {
	startingBitPos := bitPos

	var gotEOL bool
	gotEOL, bitPos = tryFetchEOL(encoded, bitPos)
	if gotEOL {
		// 1 EOL to fill EOFB
		gotEOL, bitPos = tryFetchEOL(encoded, bitPos)

		if gotEOL {
			return true, bitPos, nil
		} else {
			return false, startingBitPos, errEOFBCorrupt
		}
	}

	return false, startingBitPos, nil
}

// bitFromUint16 fetches the bit value from the `num` at the `bitPos` position.
func bitFromUint16(num uint16, bitPos int) byte {
	if bitPos < 8 {
		num >>= 8
	}

	bitPos %= 8

	mask := byte(0x01 << (7 - uint(bitPos)))

	return (byte(num) & mask) >> (7 - uint(bitPos))
}

// fetchNextRunLen fetches the next Group3 1 dimensional code from the encoded data.
// Returns the corresponding pixel run length and the moved bit position. The returned
// bit position is not moved if no valid code is met.
func fetchNextRunLen(data []byte, bitPos int, isWhite bool) (int, int) {
	var (
		codeNum        uint16
		codeBitPos     int
		startingBitPos int
	)

	startingBitPos = bitPos
	codeNum, codeBitPos, bitPos = fetchNextCode(data, bitPos)

	runLen, code := decodeRunLenFromUint16(codeNum, codeBitPos, isWhite)
	if runLen == -1 {
		return -1, startingBitPos
	}

	return runLen, startingBitPos + code.BitsWritten
}

// fetchNext2DCode fetches the next 2-dimensional code from the encoded data.
// Returns the code and the moved bit position. The returned bit position is not moved
// if no valid code is met.
func fetchNext2DCode(data []byte, bitPos int) (code, int, bool) {
	var (
		codeNum        uint16
		codeBitPos     int
		startingBitPos int
	)

	startingBitPos = bitPos
	codeNum, codeBitPos, bitPos = fetchNextCode(data, bitPos)

	codeStruct, ok := get2DCodeFromUint16(codeNum, codeBitPos)
	if !ok {
		return code{}, startingBitPos, false
	}

	return codeStruct, startingBitPos + codeStruct.BitsWritten, true
}

// get2DCodeFromUint16 finds the 2-dimensional code from the encoded data presented as
// uint16.
func get2DCodeFromUint16(encoded uint16, bitPos int) (code, bool) {
	_, codePtr := findRunLen(twoDimTree, encoded, bitPos)

	if codePtr == nil {
		return code{}, false
	}

	return *codePtr, true
}

// decodeRunLenFromUint16 finds the 1-dimensional code from the encoded data presented as
// uint16.
func decodeRunLenFromUint16(encoded uint16, bitPos int, isWhite bool) (int, code) {
	var runLenPtr *int
	var codePtr *code

	if isWhite {
		runLenPtr, codePtr = findRunLen(whiteTree, encoded, bitPos)
	} else {
		runLenPtr, codePtr = findRunLen(blackTree, encoded, bitPos)
	}

	if runLenPtr == nil {
		return -1, code{}
	}

	return *runLenPtr, *codePtr
}

// fetchNextCode assembles the next at most 16 bits starting from the `bitPos` into
// a single uint16 value. Returns the moved bit position.
func fetchNextCode(data []byte, bitPos int) (uint16, int, int) {
	startingBitPos := bitPos

	bytePos := bitPos / 8
	bitPos %= 8

	if bytePos >= len(data) {
		return 0, 16, startingBitPos
	}

	// take the rest bits from the current byte
	mask := byte(0xFF >> uint(bitPos))
	code := uint16((data[bytePos]&mask)<<uint(bitPos)) << 8

	bitsWritten := 8 - bitPos

	bytePos++
	bitPos = 0

	if bytePos >= len(data) {
		return code >> (16 - uint(bitsWritten)), 16 - bitsWritten, startingBitPos + bitsWritten
	}

	// take the whole next byte
	code |= uint16(data[bytePos]) << (8 - uint(bitsWritten))

	bitsWritten += 8

	bytePos++
	bitPos = 0

	if bytePos >= len(data) {
		return code >> (16 - uint(bitsWritten)), 16 - bitsWritten, startingBitPos + bitsWritten
	}

	if bitsWritten == 16 {
		return code, 0, startingBitPos + bitsWritten
	}

	// take the needed bits from the next byte
	leftToWrite := 16 - bitsWritten

	code |= uint16(data[bytePos] >> (8 - uint(leftToWrite)))

	return code, 0, startingBitPos + 16
}

// decodeRow1D decodes the next pixels row using the Group3 1-dimensional CCITT facsimile (fax)
// decoding algorithm.
func (e *Encoder) decodeRow1D(encoded []byte, bitPos int) ([]byte, int) {
	var pixelsRow []byte

	isWhite := true

	var runLen int
	runLen, bitPos = fetchNextRunLen(encoded, bitPos, isWhite)
	for runLen != -1 {
		pixelsRow = drawPixels(pixelsRow, isWhite, runLen)

		// got terminal code, switch color
		if runLen < 64 {
			if len(pixelsRow) >= e.Columns {
				break
			}

			isWhite = !isWhite
		}

		runLen, bitPos = fetchNextRunLen(encoded, bitPos, isWhite)
	}

	return pixelsRow, bitPos
}

// drawPixels appends the length pixels of the specified color to the `row`.
func drawPixels(row []byte, isWhite bool, length int) []byte {
	if length < 0 {
		return row
	}

	runLenPixels := make([]byte, length)
	if isWhite {
		for i := 0; i < len(runLenPixels); i++ {
			runLenPixels[i] = white
		}
	} else {
		for i := 0; i < len(runLenPixels); i++ {
			runLenPixels[i] = black
		}
	}

	row = append(row, runLenPixels...)

	return row
}

// tryFetchEOL tries to fetch the EOL code (000000000001) from the encoded data starting
// from the `bitPos` position. Returns the moved bit position if the code was met, and the original
// position otherwise.
func tryFetchEOL(encoded []byte, bitPos int) (bool, int) {
	startingBitPos := bitPos

	var (
		code       uint16
		codeBitPos int
	)

	code, codeBitPos, bitPos = fetchNextCode(encoded, bitPos)

	// when this is true, the fetched code if less than 12 bits long,
	// which doesn't fit the EOL code length
	if codeBitPos > 4 {
		return false, startingBitPos
	}

	// format the fetched code
	code >>= uint(4 - codeBitPos)
	code <<= 4

	if code != eol.Code {
		return false, startingBitPos
	} else {
		return true, bitPos - 4 + codeBitPos
	}
}

// tryFetchExtendedEOL tries to fetch the extended EOL code (0000000000011 or 0000000000010) from
// the encoded data. Returns the moved bit position if the code was met, and the original
// position otherwise.
func tryFetchExtendedEOL(encoded []byte, bitPos int, eolCode code) (bool, int) {
	startingBitPos := bitPos

	var (
		code       uint16
		codeBitPos int
	)

	code, codeBitPos, bitPos = fetchNextCode(encoded, bitPos)

	// when this is true, the fetched code if less than 13 bits long,
	// which doesn't fit the EOL0/EOL1 code length
	if codeBitPos > 3 {
		return false, startingBitPos
	}

	// format the fetched code
	code >>= uint(3 - codeBitPos)
	code <<= 3

	if code != eolCode.Code {
		return false, startingBitPos
	} else {
		return true, bitPos - 3 + codeBitPos
	}
}

// tryFetchEOL0 tries to fetch the EOL0 code (0000000000010) from the encoded data. Returns the moved bit
// position if the code was met, and the original position otherwise.
func tryFetchEOL0(encoded []byte, bitPos int) (bool, int) {
	return tryFetchExtendedEOL(encoded, bitPos, eol0)
}

// tryFetchEOL1 tries to fetch the EOL1 code (0000000000011) from the encoded data. Returns the moved bit
// position if the code was met, and the original position otherwise.
func tryFetchEOL1(encoded []byte, bitPos int) (bool, int) {
	return tryFetchExtendedEOL(encoded, bitPos, eol1)
}
