/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package imageutil

import (
	"github.com/unidoc/unipdf/v3/internal/bitwise"
)

// AddDataPadding adds the row bit padding to the given data slice if it is required by the image parameters.
func AddDataPadding(width, height, bitsPerComponent, colorComponents int, data []byte) ([]byte, error) {
	bytesPerLine := BytesPerLine(width, bitsPerComponent, colorComponents)
	if bytesPerLine == width*colorComponents*bitsPerComponent/8 {
		return data, nil
	}
	// Compute the number of bits in the line.
	bitsPerLineOld := width * colorComponents * bitsPerComponent
	bitsPerLineNew := bytesPerLine * 8
	diffBitsPerLine := 8 - (bitsPerLineNew - bitsPerLineOld)

	r := bitwise.NewReader(data)

	fullBytesNumber := bytesPerLine - 1
	fullBytesPerLine := make([]byte, fullBytesNumber)
	output := make([]byte, height*bytesPerLine)
	w := bitwise.NewWriterMSB(output)
	var bits uint64
	var err error
	for y := 0; y < height; y++ {
		// Read and write full bytes.
		_, err = r.Read(fullBytesPerLine)
		if err != nil {
			return nil, err
		}
		_, err = w.Write(fullBytesPerLine)
		if err != nil {
			return nil, err
		}
		// Read and write only padding bits.
		bits, err = r.ReadBits(byte(diffBitsPerLine))
		if err != nil {
			return nil, err
		}
		_, err = w.WriteBits(bits, diffBitsPerLine)
		if err != nil {
			return nil, err
		}
		w.FinishByte()
	}
	return output, nil
}
