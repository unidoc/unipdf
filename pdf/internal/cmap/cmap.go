/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmap

import (
	"bytes"
	"errors"
	"io"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

// CMap represents a character code to unicode mapping used in PDF files.
type CMap struct {
	*cMapParser

	// Text encoder to look up runes from input glyph names.
	encoder textencoding.TextEncoder

	// map of character code to string (sequence of runes) for 1-4 byte codes separately.
	codeMap [4]map[uint64]string

	name       string
	ctype      int
	codespaces []codespace
}

// codespace represents a single codespace range used in the CMap.
type codespace struct {
	numBytes int
	low      uint64
	high     uint64
}

// Name returns the name of the CMap.
func (cmap *CMap) Name() string {
	return cmap.name
}

// Type returns the type of the CMap.
func (cmap *CMap) Type() int {
	return cmap.ctype
}

// CharcodeBytesToUnicode converts a byte array of charcodes to a unicode string representation.
func (cmap *CMap) CharcodeBytesToUnicode(src []byte) string {
	var buf bytes.Buffer

	// Maximum number of possible bytes per code.
	maxLen := 4

	i := 0
	for i < len(src) {
		var code uint64
		var j int
		for j = 0; j < maxLen && i+j < len(src); j++ {
			b := src[i+j]

			code <<= 8
			code |= uint64(b)

			tgt, has := cmap.codeMap[j][code]
			if has {
				buf.WriteString(tgt)
				break
			} else if j == maxLen-1 || i+j == len(src)-1 {
				break
			}
		}
		i += j + 1
	}

	return buf.String()
}

// CharcodeToUnicode converts a single character code to unicode string.
// Note that CharcodeBytesToUnicode is typically more efficient.
func (cmap *CMap) CharcodeToUnicode(srcCode uint64) string {
	// Search through different code lengths.
	for numBytes := 1; numBytes <= 4; numBytes++ {
		if c, has := cmap.codeMap[numBytes-1][srcCode]; has {
			return c
		}
	}

	// Not found.
	return "?"
}

// newCMap returns an initialized CMap.
func newCMap() *CMap {
	cmap := &CMap{}
	cmap.codespaces = []codespace{}
	cmap.codeMap = [4]map[uint64]string{}
	// Maps for 1-4 bytes are initialized. Minimal overhead if not used (most commonly used are 1-2 bytes).
	cmap.codeMap[0] = map[uint64]string{}
	cmap.codeMap[1] = map[uint64]string{}
	cmap.codeMap[2] = map[uint64]string{}
	cmap.codeMap[3] = map[uint64]string{}
	return cmap
}

// LoadCmapFromData parses CMap data in memory through a byte vector and returns a CMap which
// can be used for character code to unicode conversion.
func LoadCmapFromData(data []byte) (*CMap, error) {
	cmap := newCMap()
	cmap.cMapParser = newCMapParser(data)

	err := cmap.parse()
	if err != nil {
		return cmap, err
	}

	return cmap, nil
}

// parse parses the CMap file and loads into the CMap structure.
func (cmap *CMap) parse() error {
	for {
		o, err := cmap.parseObject()
		if err != nil {
			if err == io.EOF {
				break
			}

			common.Log.Debug("Error parsing CMap: %v", err)
			return err
		}

		if op, isOp := o.(cmapOperand); isOp {
			common.Log.Trace("Operand: %s", op.Operand)

			if op.Operand == begincodespacerange {
				err := cmap.parseCodespaceRange()
				if err != nil {
					return err
				}
			} else if op.Operand == beginbfchar {
				err := cmap.parseBfchar()
				if err != nil {
					return err
				}
			} else if op.Operand == beginbfrange {
				err := cmap.parseBfrange()
				if err != nil {
					return err
				}
			}
		} else if n, isName := o.(cmapName); isName {
			if n.Name == cmapname {
				o, err := cmap.parseObject()
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				name, ok := o.(cmapName)
				if !ok {
					return errors.New("CMap name not a name")
				}
				cmap.name = name.Name
			} else if n.Name == cmaptype {
				o, err := cmap.parseObject()
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				typeInt, ok := o.(cmapInt)
				if !ok {
					return errors.New("CMap type not an integer")
				}
				cmap.ctype = int(typeInt.val)
			}
		} else {
			common.Log.Trace("Unhandled object: %T %#v", o, o)
		}
	}

	return nil
}

// parseCodespaceRange parses the codespace range section of a CMap.
func (cmap *CMap) parseCodespaceRange() error {
	for {
		o, err := cmap.parseObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		hexLow, isHex := o.(cmapHexString)
		if !isHex {
			if op, isOperand := o.(cmapOperand); isOperand {
				if op.Operand == endcodespacerange {
					return nil
				}
				return errors.New("Unexpected operand")
			}
		}

		o, err = cmap.parseObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		hexHigh, ok := o.(cmapHexString)
		if !ok {
			return errors.New("Non-hex high")
		}

		if hexLow.numBytes != hexHigh.numBytes {
			return errors.New("Unequal number of bytes in range")
		}

		low := hexToUint64(hexLow)
		high := hexToUint64(hexHigh)
		numBytes := hexLow.numBytes

		cspace := codespace{numBytes: numBytes, low: low, high: high}
		cmap.codespaces = append(cmap.codespaces, cspace)

		common.Log.Trace("Codespace low: 0x%X, high: 0x%X", low, high)
	}

	return nil
}

// parseBfchar parses a bfchar section of a CMap file.
func (cmap *CMap) parseBfchar() error {
	for {
		// Src code.
		o, err := cmap.parseObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		var srcCode uint64
		var numBytes int

		switch v := o.(type) {
		case cmapOperand:
			if v.Operand == endbfchar {
				return nil
			}
			return errors.New("Unexpected operand")
		case cmapHexString:
			srcCode = hexToUint64(v)
			numBytes = v.numBytes
		default:
			return errors.New("Unexpected type")
		}

		// Target code.
		o, err = cmap.parseObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		var toCode string

		switch v := o.(type) {
		case cmapOperand:
			if v.Operand == endbfchar {
				return nil
			}
			return errors.New("Unexpected operand")
		case cmapHexString:
			toCode = hexToString(v)
		case cmapName:
			toCode = "?"
			if cmap.encoder != nil {
				if r, found := cmap.encoder.GlyphToRune(v.Name); found {
					toCode = string(r)
				}
			}
		default:
			return errors.New("Unexpected type")
		}

		if numBytes <= 0 || numBytes > 4 {
			return errors.New("Invalid code length")
		}

		cmap.codeMap[numBytes-1][srcCode] = toCode
	}

	return nil
}

// parseBfrange parses a bfrange section of a CMap file.
func (cmap *CMap) parseBfrange() error {
	for {
		// The specifications are in pairs of 3.
		// <srcCodeFrom> <srcCodeTo> <target>
		// where target can be either <destFrom> as a hex code, or a list.

		// Src code from.
		var srcCodeFrom uint64
		var numBytes int
		{
			o, err := cmap.parseObject()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			switch v := o.(type) {
			case cmapOperand:
				if v.Operand == endbfrange {
					return nil
				}
				return errors.New("Unexpected operand")
			case cmapHexString:
				srcCodeFrom = hexToUint64(v)
				numBytes = v.numBytes
			default:
				return errors.New("Unexpected type")
			}
		}

		// Src code to.
		var srcCodeTo uint64
		{
			o, err := cmap.parseObject()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			switch v := o.(type) {
			case cmapOperand:
				if v.Operand == endbfrange {
					return nil
				}
				return errors.New("Unexpected operand")
			case cmapHexString:
				srcCodeTo = hexToUint64(v)
			default:
				return errors.New("Unexpected type")
			}
		}

		// target(s).
		o, err := cmap.parseObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if numBytes <= 0 || numBytes > 4 {
			return errors.New("Invalid code length")
		}

		switch v := o.(type) {
		case cmapArray:
			sc := srcCodeFrom
			for _, o := range v.Array {
				hexs, ok := o.(cmapHexString)
				if !ok {
					return errors.New("Non-hex string in array")
				}
				cmap.codeMap[numBytes-1][sc] = hexToString(hexs)
				sc++
			}
			if sc != srcCodeTo+1 {
				return errors.New("Invalid number of items in array")
			}
		case cmapHexString:
			// <srcCodeFrom> <srcCodeTo> <dstCode>, maps [from,to] to [dstCode,dstCode+to-from].
			// in hex format.
			target := hexToUint64(v)
			i := uint64(0)
			for sc := srcCodeFrom; sc <= srcCodeTo; sc++ {
				r := target + i
				cmap.codeMap[numBytes-1][sc] = string(r)
				i++
			}
		default:
			return errors.New("Unexpected type")
		}
	}

	return nil
}
