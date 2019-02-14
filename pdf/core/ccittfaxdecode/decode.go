package ccittfaxdecode

import (
	"errors"
)

var (
	ErrEOFBCorrupt               = errors.New("EOFB code is corrupted")
	ErrRTCCorrupt                = errors.New("RTC code is corrupted")
	ErrWrongCodeInHorizontalMode = errors.New("wrong code in horizontal mode")
	ErrNoEOLFound                = errors.New("no EOL found while the EndOfLine parameter is true")
	ErrInvalidEOL                = errors.New("invalid EOL")
	ErrInvalid2DCode             = errors.New("invalid 2D code")

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
	for runLen, code := range WTerms {
		addNode(whiteTree, code, 0, runLen)
	}

	for runLen, code := range WMakeups {
		addNode(whiteTree, code, 0, runLen)
	}

	for runLen, code := range BTerms {
		addNode(blackTree, code, 0, runLen)
	}

	for runLen, code := range BMakeups {
		addNode(blackTree, code, 0, runLen)
	}

	for runLen, code := range CommonMakeups {
		addNode(whiteTree, code, 0, runLen)
		addNode(blackTree, code, 0, runLen)
	}

	addNode(twoDimTree, P, 0, 0)
	addNode(twoDimTree, H, 0, 0)
	addNode(twoDimTree, V0, 0, 0)
	addNode(twoDimTree, V1R, 0, 0)
	addNode(twoDimTree, V2R, 0, 0)
	addNode(twoDimTree, V3R, 0, 0)
	addNode(twoDimTree, V1L, 0, 0)
	addNode(twoDimTree, V2L, 0, 0)
	addNode(twoDimTree, V3L, 0, 0)
}

func (e *Encoder) Decode(encoded []byte) ([][]byte, error) {
	if e.BlackIs1 {
		white = 0
		black = 1
	} else {
		white = 1
		black = 0
	}

	if e.K == 0 {
		return e.decodeG31D(encoded)
	}

	if e.K > 0 {
		return e.decodeG32D(encoded)
	}

	if e.K < 4 {
		return e.decodeG4(encoded)
	}

	return nil, nil
}

func (e *Encoder) decodeG31D(encoded []byte) ([][]byte, error) {
	var pixels [][]byte

	// do g31d decoding
	var bitPos int
	for (bitPos / 8) < len(encoded) {
		var gotEOL bool

		gotEOL, bitPos = tryFetchEOL(encoded, bitPos)
		if !gotEOL {
			if e.EndOfLine {
				return nil, ErrNoEOLFound
			}
		} else {
			// 5 EOLs left to fill RTC
			for i := 0; i < 5; i++ {
				gotEOL, bitPos = tryFetchEOL(encoded, bitPos)

				if !gotEOL {
					if i == 0 {
						break
					}

					return nil, ErrInvalidEOL
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
			bitPos += 8 - bitPos%8
		}

		pixels = append(pixels, row)

		if e.Rows > 0 && !e.EndOfBlock && len(pixels) >= e.Rows {
			break
		}
	}

	return pixels, nil
}

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
				return nil, ErrNoEOLFound
			}
		}

		// decode 1 of K rows as 1D
		var row []byte
		row, bitPos = e.decodeRow1D(encoded, bitPos)

		if e.EncodedByteAlign && bitPos%8 != 0 {
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
				gotEOL, bitPos, err = tryFetchRTC2D(encoded, bitPos)
				if err != nil {
					return nil, err
				}

				if gotEOL {
					break byteLoop
				} else {
					if e.EndOfLine {
						return nil, ErrNoEOLFound
					}
				}
			}

			var (
				twoDimCode Code
				ok         bool
			)

			isWhite := true
			var pixelsRow []byte
			a0 := -1
			for twoDimCode, bitPos, ok = fetchNext2DCode(encoded, bitPos); ok; twoDimCode, bitPos, ok = fetchNext2DCode(encoded, bitPos) {
				switch twoDimCode {
				case P:
					// do pass mode decoding
					pixelsRow, a0 = decodePassMode(pixels, pixelsRow, isWhite, a0)
				case H:
					// do horizontal mode decoding
					pixelsRow, bitPos, a0, err = decodeHorizontalMode(encoded, pixelsRow, bitPos, isWhite, a0)
					if err != nil {
						return nil, err
					}
				case V0:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 0)

					isWhite = !isWhite
				case V1R:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 1)

					isWhite = !isWhite
				case V2R:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 2)

					isWhite = !isWhite
				case V3R:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 3)

					isWhite = !isWhite
				case V1L:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -1)

					isWhite = !isWhite
				case V2L:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -2)

					isWhite = !isWhite
				case V3L:
					pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -3)

					isWhite = !isWhite
				}

				// TODO: additionally check for errors. EOL should be fetched when a0 == len(pixels[len(pixels)-1])
				if len(pixelsRow) >= e.Columns {
					break
				}
			}

			if e.EncodedByteAlign && bitPos%8 != 0 {
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

func (e *Encoder) decodeG4(encoded []byte) ([][]byte, error) {
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
			twoDimCode Code
			ok         bool
		)

		isWhite := true
		var pixelsRow []byte
		a0 := -1
		for a0 < e.Columns {
			twoDimCode, bitPos, ok = fetchNext2DCode(encoded, bitPos)
			if !ok {
				return nil, ErrInvalid2DCode
			}

			switch twoDimCode {
			case P:
				// do pass mode decoding
				pixelsRow, a0 = decodePassMode(pixels, pixelsRow, isWhite, a0)
			case H:
				// do horizontal mode decoding
				pixelsRow, bitPos, a0, err = decodeHorizontalMode(encoded, pixelsRow, bitPos, isWhite, a0)
				if err != nil {
					return nil, err
				}
			case V0:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 0)

				isWhite = !isWhite
			case V1R:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 1)

				isWhite = !isWhite
			case V2R:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 2)

				isWhite = !isWhite
			case V3R:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, 3)

				isWhite = !isWhite
			case V1L:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -1)

				isWhite = !isWhite
			case V2L:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -2)

				isWhite = !isWhite
			case V3L:
				pixelsRow, a0 = decodeVerticalMode(pixels, pixelsRow, isWhite, a0, -3)

				isWhite = !isWhite
			}

			if len(pixelsRow) >= e.Columns {
				break
			}
		}

		if e.EncodedByteAlign && bitPos%8 != 0 {
			bitPos += 8 - bitPos%8
		}

		pixels = append(pixels, pixelsRow)

		if e.Rows > 0 && !e.EndOfBlock && len(pixels) >= (e.Rows+1) {
			break
		}
	}

	pixels = pixels[1:]

	return pixels, nil
}

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
		return pixelsRow, startingBitPos, ErrWrongCodeInHorizontalMode
	}

	return pixelsRow, bitPos, nil
}

func tryFetchRTC2D(encoded []byte, bitPos int) (bool, int, error) {
	startingBitPos := bitPos
	var gotEOL = false

	// 6 EOL1s to fill RTC
	for i := 0; i < 6; i++ {
		gotEOL, bitPos = tryFetchEOL1(encoded, bitPos)

		if !gotEOL {
			if i > 1 {
				return false, startingBitPos, ErrRTCCorrupt
			} else {
				bitPos = startingBitPos

				break
			}
		}
	}

	return gotEOL, bitPos, nil
}

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
			return false, startingBitPos, ErrEOFBCorrupt
		}
	}

	return false, startingBitPos, nil
}

func bitFromUint16(num uint16, bitPos int) byte {
	if bitPos < 8 {
		num >>= 8
	}

	bitPos %= 8

	mask := byte(0x01 << (7 - uint(bitPos)))

	return (byte(num) & mask) >> (7 - uint(bitPos))
}

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

func fetchNext2DCode(data []byte, bitPos int) (Code, int, bool) {
	var (
		codeNum        uint16
		codeBitPos     int
		startingBitPos int
	)

	startingBitPos = bitPos
	codeNum, codeBitPos, bitPos = fetchNextCode(data, bitPos)

	code, ok := get2DCodeFromUint16(codeNum, codeBitPos)
	if !ok {
		return Code{}, startingBitPos, false
	}

	return code, startingBitPos + code.BitsWritten, true
}

func get2DCodeFromUint16(encoded uint16, bitPos int) (Code, bool) {
	_, codePtr := findRunLen(twoDimTree, encoded, bitPos)

	if codePtr == nil {
		return Code{}, false
	}

	return *codePtr, true
}

func decodeRunLenFromUint16(encoded uint16, bitPos int, isWhite bool) (int, Code) {
	var runLenPtr *int
	var codePtr *Code

	if isWhite {
		runLenPtr, codePtr = findRunLen(whiteTree, encoded, bitPos)
	} else {
		runLenPtr, codePtr = findRunLen(blackTree, encoded, bitPos)
	}

	if runLenPtr == nil {
		return -1, Code{}
	}

	return *runLenPtr, *codePtr
}

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

	if code != EOL.Code {
		return false, startingBitPos
	} else {
		return true, bitPos - 4 + codeBitPos
	}
}

func tryFetchExtendedEOL(encoded []byte, bitPos int, eolCode Code) (bool, int) {
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

func tryFetchEOL0(encoded []byte, bitPos int) (bool, int) {
	return tryFetchExtendedEOL(encoded, bitPos, EOL0)
}

func tryFetchEOL1(encoded []byte, bitPos int) (bool, int) {
	return tryFetchExtendedEOL(encoded, bitPos, EOL1)
}
