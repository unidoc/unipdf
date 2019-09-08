/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"errors"
	"image"
	"math"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// PatternDictionary is the jbig2 model for the pattern dictionary segment - 7.4.4.
type PatternDictionary struct {
	r reader.StreamReader

	DataHeaderOffset int64
	DataHeaderLength int64
	DataOffset       int64
	DataLength       int64

	GBAtX []int8
	GBAtY []int8

	// Flags 7.4.4.1.1
	IsMMREncoded bool
	HDTemplate   byte

	// Width of the patterns in the pattern dictionary
	HdpWidth byte
	// Height of the patterns in the pattern dictionary
	HdpHeight byte

	// Decoded bitmaps stored to be used by segments that refer to it
	Patterns []*bitmap.Bitmap

	// Largest gray-scale value 7.4.4.1.4
	GrayMax uint32
}

// GetDictionary gets the PatternDictionary segment Dictionary bitmaps.
func (p *PatternDictionary) GetDictionary() ([]*bitmap.Bitmap, error) {
	if p.Patterns != nil {
		return p.Patterns, nil
	}

	if !p.IsMMREncoded {
		p.setGbAtPixels()
	}

	genericRegion := NewGenericRegion(p.r)
	genericRegion.setParametersMMR(p.IsMMREncoded, p.DataOffset, p.DataLength, uint32(p.HdpHeight), (p.GrayMax+1)*uint32(p.HdpWidth), p.HDTemplate, false, false, p.GBAtX, p.GBAtY)

	collectiveBitmap, err := genericRegion.GetRegionBitmap()
	if err != nil {
		return nil, err
	}

	if err = p.extractPatterns(collectiveBitmap); err != nil {
		return nil, err
	}
	return p.Patterns, nil
}

// Init implements Segmenter interface.
func (p *PatternDictionary) Init(h *Header, r reader.StreamReader) error {
	p.r = r
	return p.parseHeader()
}

func (p *PatternDictionary) checkInput() error {
	if p.HdpHeight < 1 || p.HdpWidth < 1 {
		return errors.New("invalid Header Value: Width/Height must be greater than zero")
	}

	if p.IsMMREncoded {
		if p.HDTemplate != 0 {
			common.Log.Debug("variable HDTemplate should not contain the value 0")
		}
	}
	return nil
}

func (p *PatternDictionary) computeSegmentDataStructure() error {
	p.DataOffset = p.r.StreamPosition()
	p.DataHeaderLength = p.DataOffset - p.DataHeaderOffset
	p.DataLength = int64(p.r.Length()) - p.DataHeaderLength
	return nil
}

func (p *PatternDictionary) extractPatterns(collectiveBitmap *bitmap.Bitmap) error {
	// 3)
	var gray int
	patterns := make([]*bitmap.Bitmap, p.GrayMax+1)

	// 4
	for gray <= int(p.GrayMax) {
		// 4 a)
		x0 := int(p.HdpWidth) * gray
		roi := image.Rect(x0, 0, x0+int(p.HdpWidth), int(p.HdpHeight))
		patternBitmap, err := bitmap.Extract(roi, collectiveBitmap)
		if err != nil {
			return err
		}
		patterns[gray] = patternBitmap

		// 4 b)
		gray++
	}

	p.Patterns = patterns
	return nil
}

func (p *PatternDictionary) parseHeader() error {
	common.Log.Trace("[PATTERN-DICTIONARY][parseHeader] begin")
	defer func() {
		common.Log.Trace("[PATTERN-DICTIONARY][parseHeader] finished")
	}()

	// Bit 3-7 dirty read
	_, err := p.r.ReadBits(5)
	if err != nil {
		return err
	}
	// Bit 1-2
	if err = p.readTemplate(); err != nil {
		return err
	}

	// Bit 0
	if err = p.readIsMMREncoded(); err != nil {
		return err
	}

	if err = p.readPatternWidthAndHeight(); err != nil {
		return err
	}

	if err = p.readGrayMax(); err != nil {
		return err
	}

	if err = p.computeSegmentDataStructure(); err != nil {
		return err
	}
	return p.checkInput()
}

func (p *PatternDictionary) readTemplate() error {
	temp, err := p.r.ReadBits(2)
	if err != nil {
		return err
	}
	p.HDTemplate = byte(temp)
	return nil
}

func (p *PatternDictionary) readIsMMREncoded() error {
	bit, err := p.r.ReadBit()
	if err != nil {
		return err
	}
	if bit != 0 {
		p.IsMMREncoded = true
	}
	return nil
}

func (p *PatternDictionary) readPatternWidthAndHeight() error {
	temp, err := p.r.ReadByte()
	if err != nil {
		return err
	}
	p.HdpWidth = temp

	temp, err = p.r.ReadByte()
	if err != nil {
		return err
	}
	p.HdpHeight = temp
	return nil
}

func (p *PatternDictionary) readGrayMax() error {
	temp, err := p.r.ReadBits(32)
	if err != nil {
		return err
	}
	p.GrayMax = uint32(temp & math.MaxUint32)
	return nil
}

func (p *PatternDictionary) setGbAtPixels() {
	if p.HDTemplate == 0 {
		p.GBAtX = make([]int8, 4)
		p.GBAtY = make([]int8, 4)

		p.GBAtX[0] = -int8(p.HdpWidth)
		p.GBAtY[0] = 0

		p.GBAtX[1] = -3
		p.GBAtY[1] = -1

		p.GBAtX[2] = 2
		p.GBAtY[2] = -2

		p.GBAtX[3] = -2
		p.GBAtY[3] = -2
	} else {
		p.GBAtX = []int8{-int8(p.HdpWidth)}
		p.GBAtY = []int8{0}
	}
}
