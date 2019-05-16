/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fdf

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// Regular Expressions for parsing and identifying object signatures.
var reFdfVersion = regexp.MustCompile(`%FDF-(\d)\.(\d)`)
var reEOF = regexp.MustCompile("%%EOF")
var reNumeric = regexp.MustCompile(`^[\+-.]*([0-9.]+)`)
var reExponential = regexp.MustCompile(`^[\+-.]*([0-9.]+)e[\+-.]*([0-9.]+)`)
var reReference = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+R`)
var reIndirectObject = regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)

// fdfParser parses a FDF file and provides access to the object structure of the FDF.
type fdfParser struct {
	majorVersion int
	minorVersion int

	objCache map[int64]core.PdfObject

	rs       io.ReadSeeker
	reader   *bufio.Reader
	fileSize int64

	trailerDict *core.PdfObjectDictionary
}

// Skip over any spaces.
func (parser *fdfParser) skipSpaces() (int, error) {
	cnt := 0
	for {
		b, err := parser.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		if core.IsWhiteSpace(b) {
			cnt++
		} else {
			parser.reader.UnreadByte()
			break
		}
	}

	return cnt, nil
}

// Skip over comments and spaces. Can handle multi-line comments.
func (parser *fdfParser) skipComments() error {
	if _, err := parser.skipSpaces(); err != nil {
		return err
	}

	isFirst := true
	for {
		bb, err := parser.reader.Peek(1)
		if err != nil {
			common.Log.Debug("Error %s", err.Error())
			return err
		}
		if isFirst && bb[0] != '%' {
			// Not a comment clearly.
			return nil
		}
		isFirst = false
		if (bb[0] != '\r') && (bb[0] != '\n') {
			parser.reader.ReadByte()
		} else {
			break
		}
	}

	// Call recursively to handle multiline comments.
	return parser.skipComments()
}

// Read a comment starting with '%'.
func (parser *fdfParser) readComment() (string, error) {
	var r bytes.Buffer

	_, err := parser.skipSpaces()
	if err != nil {
		return r.String(), err
	}

	isFirst := true
	for {
		bb, err := parser.reader.Peek(1)
		if err != nil {
			common.Log.Debug("Error %s", err.Error())
			return r.String(), err
		}
		if isFirst && bb[0] != '%' {
			return r.String(), errors.New("comment should start with %")
		}
		isFirst = false
		if (bb[0] != '\r') && (bb[0] != '\n') {
			b, _ := parser.reader.ReadByte()
			r.WriteByte(b)
		} else {
			break
		}
	}
	return r.String(), nil
}

// Read a single line of text from current position.
func (parser *fdfParser) readTextLine() (string, error) {
	var r bytes.Buffer
	for {
		bb, err := parser.reader.Peek(1)
		if err != nil {
			common.Log.Debug("Error %s", err.Error())
			return r.String(), err
		}
		if (bb[0] != '\r') && (bb[0] != '\n') {
			b, _ := parser.reader.ReadByte()
			r.WriteByte(b)
		} else {
			break
		}
	}
	return r.String(), nil
}

// Parse a name starting with '/'.
func (parser *fdfParser) parseName() (core.PdfObjectName, error) {
	var r bytes.Buffer
	nameStarted := false
	for {
		bb, err := parser.reader.Peek(1)
		if err == io.EOF {
			break // Can happen when loading from object stream.
		}
		if err != nil {
			return core.PdfObjectName(r.String()), err
		}

		if !nameStarted {
			// Should always start with '/', otherwise not valid.
			if bb[0] == '/' {
				nameStarted = true
				parser.reader.ReadByte()
			} else if bb[0] == '%' {
				parser.readComment()
				parser.skipSpaces()
			} else {
				common.Log.Debug("ERROR Name starting with %s (% x)", bb, bb)
				return core.PdfObjectName(r.String()), fmt.Errorf("invalid name: (%c)", bb[0])
			}
		} else {
			if core.IsWhiteSpace(bb[0]) {
				break
			} else if (bb[0] == '/') || (bb[0] == '[') || (bb[0] == '(') || (bb[0] == ']') || (bb[0] == '<') || (bb[0] == '>') {
				break // Looks like start of next statement.
			} else if bb[0] == '#' {
				hexcode, err := parser.reader.Peek(3)
				if err != nil {
					return core.PdfObjectName(r.String()), err
				}
				parser.reader.Discard(3)

				code, err := hex.DecodeString(string(hexcode[1:3]))
				if err != nil {
					return core.PdfObjectName(r.String()), err
				}
				r.Write(code)
			} else {
				b, _ := parser.reader.ReadByte()
				r.WriteByte(b)
			}
		}
	}
	return core.PdfObjectName(r.String()), nil
}

// Numeric objects.
// Section 7.3.3.
// Integer or Float.
//
// An integer shall be written as one or more decimal digits optionally
// preceded by a sign. The value shall be interpreted as a signed
// decimal integer and shall be converted to an integer object.
//
// A real value shall be written as one or more decimal digits with an
// optional sign and a leading, trailing, or embedded PERIOD (2Eh)
// (decimal point). The value shall be interpreted as a real number
// and shall be converted to a real object.
//
// Regarding exponential numbers: 7.3.3 Numeric Objects:
// A conforming writer shall not use the PostScript syntax for numbers
// with non-decimal radices (such as 16#FFFE) or in exponential format
// (such as 6.02E23).
// Nonetheless, we sometimes get numbers with exponential format, so
// we will support it in the reader (no confusion with other types, so
// no compromise).
func (parser *fdfParser) parseNumber() (core.PdfObject, error) {
	isFloat := false
	allowSigns := true
	var r bytes.Buffer
	for {
		common.Log.Trace("Parsing number \"%s\"", r.String())
		bb, err := parser.reader.Peek(1)
		if err == io.EOF {
			// GH: EOF handling.  Handle EOF like end of line.  Can happen with
			// encoded object streams that the object is at the end.
			// In other cases, we will get the EOF error elsewhere at any rate.
			break // Handle like EOF
		}
		if err != nil {
			common.Log.Debug("ERROR %s", err)
			return nil, err
		}
		if allowSigns && (bb[0] == '-' || bb[0] == '+') {
			// Only appear in the beginning, otherwise serves as a delimiter.
			b, _ := parser.reader.ReadByte()
			r.WriteByte(b)
			allowSigns = false // Only allowed in beginning, and after e (exponential).
		} else if core.IsDecimalDigit(bb[0]) {
			b, _ := parser.reader.ReadByte()
			r.WriteByte(b)
		} else if bb[0] == '.' {
			b, _ := parser.reader.ReadByte()
			r.WriteByte(b)
			isFloat = true
		} else if bb[0] == 'e' {
			// Exponential number format.
			b, _ := parser.reader.ReadByte()
			r.WriteByte(b)
			isFloat = true
			allowSigns = true
		} else {
			break
		}
	}

	if isFloat {
		fVal, err := strconv.ParseFloat(r.String(), 64)
		o := core.PdfObjectFloat(fVal)
		return &o, err
	} else {
		intVal, err := strconv.ParseInt(r.String(), 10, 64)
		o := core.PdfObjectInteger(intVal)
		return &o, err
	}
}

// A string starts with '(' and ends with ')'.
func (parser *fdfParser) parseString() (*core.PdfObjectString, error) {
	parser.reader.ReadByte()

	var r bytes.Buffer
	count := 1
	for {
		bb, err := parser.reader.Peek(1)
		if err != nil {
			return core.MakeString(r.String()), err
		}

		if bb[0] == '\\' { // Escape sequence.
			parser.reader.ReadByte() // Skip the escape \ byte.
			b, err := parser.reader.ReadByte()
			if err != nil {
				return core.MakeString(r.String()), err
			}

			// Octal '\ddd' number (base 8).
			if core.IsOctalDigit(b) {
				bb, err := parser.reader.Peek(2)
				if err != nil {
					return core.MakeString(r.String()), err
				}

				var numeric []byte
				numeric = append(numeric, b)
				for _, val := range bb {
					if core.IsOctalDigit(val) {
						numeric = append(numeric, val)
					} else {
						break
					}
				}
				parser.reader.Discard(len(numeric) - 1)

				common.Log.Trace("Numeric string \"%s\"", numeric)
				code, err := strconv.ParseUint(string(numeric), 8, 32)
				if err != nil {
					return core.MakeString(r.String()), err
				}
				r.WriteByte(byte(code))
				continue
			}

			switch b {
			case 'n':
				r.WriteRune('\n')
			case 'r':
				r.WriteRune('\r')
			case 't':
				r.WriteRune('\t')
			case 'b':
				r.WriteRune('\b')
			case 'f':
				r.WriteRune('\f')
			case '(':
				r.WriteRune('(')
			case ')':
				r.WriteRune(')')
			case '\\':
				r.WriteRune('\\')
			}

			continue
		} else if bb[0] == '(' {
			count++
		} else if bb[0] == ')' {
			count--
			if count == 0 {
				parser.reader.ReadByte()
				break
			}
		}

		b, _ := parser.reader.ReadByte()
		r.WriteByte(b)
	}

	return core.MakeString(r.String()), nil
}

// Starts with '<' ends with '>'.
// Currently not converting the hex codes to characters.
func (parser *fdfParser) parseHexString() (*core.PdfObjectString, error) {
	parser.reader.ReadByte()

	var r bytes.Buffer
	for {
		bb, err := parser.reader.Peek(1)
		if err != nil {
			return core.MakeHexString(""), err
		}

		if bb[0] == '>' {
			parser.reader.ReadByte()
			break
		}

		b, _ := parser.reader.ReadByte()
		if !core.IsWhiteSpace(b) {
			r.WriteByte(b)
		}
	}

	if r.Len()%2 == 1 {
		r.WriteRune('0')
	}

	buf, err := hex.DecodeString(r.String())
	if err != nil {
		common.Log.Debug("ERROR Parsing hex string: '%s' - returning an empty string", r.String())
		return core.MakeHexString(""), nil
	}
	return core.MakeHexString(string(buf)), nil
}

// Starts with '[' ends with ']'.  Can contain any kinds of direct objects.
func (parser *fdfParser) parseArray() (*core.PdfObjectArray, error) {
	arr := core.MakeArray()

	parser.reader.ReadByte()

	for {
		parser.skipSpaces()

		bb, err := parser.reader.Peek(1)
		if err != nil {
			return arr, err
		}

		if bb[0] == ']' {
			parser.reader.ReadByte()
			break
		}

		obj, err := parser.parseObject()
		if err != nil {
			return arr, err
		}
		arr.Append(obj)
	}

	return arr, nil
}

// Parse bool object.
func (parser *fdfParser) parseBool() (core.PdfObjectBool, error) {
	bb, err := parser.reader.Peek(4)
	if err != nil {
		return core.PdfObjectBool(false), err
	}
	if (len(bb) >= 4) && (string(bb[:4]) == "true") {
		parser.reader.Discard(4)
		return core.PdfObjectBool(true), nil
	}

	bb, err = parser.reader.Peek(5)
	if err != nil {
		return core.PdfObjectBool(false), err
	}
	if (len(bb) >= 5) && (string(bb[:5]) == "false") {
		parser.reader.Discard(5)
		return core.PdfObjectBool(false), nil
	}

	return core.PdfObjectBool(false), errors.New("unexpected boolean string")
}

// Parse reference to an indirect object.
func parseReference(refStr string) (core.PdfObjectReference, error) {
	objref := core.PdfObjectReference{}

	result := reReference.FindStringSubmatch(string(refStr))
	if len(result) < 3 {
		common.Log.Debug("Error parsing reference")
		return objref, errors.New("unable to parse reference")
	}

	objNum, err := strconv.Atoi(result[1])
	if err != nil {
		common.Log.Debug("Error parsing object number '%s' - Using object num = 0", result[1])
		return objref, nil
	}
	objref.ObjectNumber = int64(objNum)

	genNum, err := strconv.Atoi(result[2])
	if err != nil {
		common.Log.Debug("Error parsing generation number '%s' - Using gen = 0", result[2])
		return objref, nil
	}
	objref.GenerationNumber = int64(genNum)

	return objref, nil
}

// Parse null object.
func (parser *fdfParser) parseNull() (core.PdfObjectNull, error) {
	_, err := parser.reader.Discard(4)
	return core.PdfObjectNull{}, err
}

// Detect the signature at the current file position and parse
// the corresponding object.
func (parser *fdfParser) parseObject() (core.PdfObject, error) {
	common.Log.Trace("Read direct object")
	parser.skipSpaces()
	for {
		bb, err := parser.reader.Peek(2)
		if err != nil {
			return nil, err
		}

		common.Log.Trace("Peek string: %s", string(bb))
		// Determine type.
		if bb[0] == '/' {
			name, err := parser.parseName()
			common.Log.Trace("->Name: '%s'", name)
			return &name, err
		} else if bb[0] == '(' {
			common.Log.Trace("->String!")
			return parser.parseString()
		} else if bb[0] == '[' {
			common.Log.Trace("->Array!")
			return parser.parseArray()
		} else if (bb[0] == '<') && (bb[1] == '<') {
			common.Log.Trace("->Dict!")
			return parser.parseDict()
		} else if bb[0] == '<' {
			common.Log.Trace("->Hex string!")
			return parser.parseHexString()
		} else if bb[0] == '%' {
			parser.readComment()
			parser.skipSpaces()
		} else {
			common.Log.Trace("->Number or ref?")
			// Reference or number?
			// Let's peek farther to find out.
			bb, _ = parser.reader.Peek(15)
			peekStr := string(bb)
			common.Log.Trace("Peek str: %s", peekStr)

			if (len(peekStr) > 3) && (peekStr[:4] == "null") {
				null, err := parser.parseNull()
				return &null, err
			} else if (len(peekStr) > 4) && (peekStr[:5] == "false") {
				b, err := parser.parseBool()
				return &b, err
			} else if (len(peekStr) > 3) && (peekStr[:4] == "true") {
				b, err := parser.parseBool()
				return &b, err
			}

			// Match reference.
			result1 := reReference.FindStringSubmatch(string(peekStr))
			if len(result1) > 1 {
				bb, _ = parser.reader.ReadBytes('R')
				common.Log.Trace("-> !Ref: '%s'", string(bb[:]))
				ref, err := parseReference(string(bb))
				return &ref, err
			}

			result2 := reNumeric.FindStringSubmatch(string(peekStr))
			if len(result2) > 1 {
				// Number object.
				common.Log.Trace("-> Number!")
				return parser.parseNumber()
			}

			result2 = reExponential.FindStringSubmatch(string(peekStr))
			if len(result2) > 1 {
				// Number object (exponential)
				common.Log.Trace("-> Exponential Number!")
				common.Log.Trace("% s", result2)
				return parser.parseNumber()
			}

			common.Log.Debug("ERROR Unknown (peek \"%s\")", peekStr)
			return nil, errors.New("object parsing error - unexpected pattern")
		}
	}
}

// Reads and parses a FDF dictionary object enclosed with '<<' and '>>'
func (parser *fdfParser) parseDict() (*core.PdfObjectDictionary, error) {
	common.Log.Trace("Reading FDF Dict!")

	dict := core.MakeDict()

	// Pass the '<<'
	c, _ := parser.reader.ReadByte()
	if c != '<' {
		return nil, errors.New("invalid dict")
	}
	c, _ = parser.reader.ReadByte()
	if c != '<' {
		return nil, errors.New("invalid dict")
	}

	for {
		parser.skipSpaces()
		parser.skipComments()

		bb, err := parser.reader.Peek(2)
		if err != nil {
			return nil, err
		}

		common.Log.Trace("Dict peek: %s (% x)!", string(bb), string(bb))
		if (bb[0] == '>') && (bb[1] == '>') {
			common.Log.Trace("EOF dictionary")
			parser.reader.ReadByte()
			parser.reader.ReadByte()
			break
		}
		common.Log.Trace("Parse the name!")

		keyName, err := parser.parseName()
		common.Log.Trace("Key: %s", keyName)
		if err != nil {
			common.Log.Debug("ERROR Returning name err %s", err)
			return nil, err
		}

		if len(keyName) > 4 && keyName[len(keyName)-4:] == "null" {
			// Some writers have a bug where the null is appended without
			// space.  For example "\Boundsnull"
			newKey := keyName[0 : len(keyName)-4]
			common.Log.Debug("Taking care of null bug (%s)", keyName)
			common.Log.Debug("New key \"%s\" = null", newKey)
			parser.skipSpaces()
			bb, _ := parser.reader.Peek(1)
			if bb[0] == '/' {
				dict.Set(newKey, core.MakeNull())
				continue
			}
		}

		parser.skipSpaces()

		val, err := parser.parseObject()
		if err != nil {
			return nil, err
		}
		dict.Set(keyName, val)

		common.Log.Trace("dict[%s] = %s", keyName, val.String())
	}
	common.Log.Trace("returning FDF Dict!")

	return dict, nil
}

// Parse the FDF version from the beginning of the file.
// Returns the major and minor parts of the version.
// E.g. for "FDF-1.4" would return 1 and 4.
func (parser *fdfParser) parseFdfVersion() (int, int, error) {
	parser.rs.Seek(0, io.SeekStart)
	offset := 20
	b := make([]byte, offset)
	parser.rs.Read(b)

	result1 := reFdfVersion.FindStringSubmatch(string(b))
	if len(result1) < 3 {
		major, minor, err := parser.seekFdfVersionTopDown()
		if err != nil {
			common.Log.Debug("Failed recovery - unable to find version")
			return 0, 0, err
		}

		return major, minor, nil
	}

	majorVersion, err := strconv.Atoi(result1[1])
	if err != nil {
		return 0, 0, err
	}

	minorVersion, err := strconv.Atoi(result1[2])
	if err != nil {
		return 0, 0, err
	}

	common.Log.Debug("Fdf version %d.%d", majorVersion, minorVersion)

	return int(majorVersion), int(minorVersion), nil
}

// Look for EOF marker and seek to its beginning.
// Define an offset position from the end of the file.
func (parser *fdfParser) seekToEOFMarker(fSize int64) error {
	// Define the starting point (from the end of the file) to search from.
	offset := int64(0)

	// Define an buffer length in terms of how many bytes to read from the end of the file.
	buflen := int64(1000)

	for offset < fSize {
		if fSize <= (buflen + offset) {
			buflen = fSize - offset
		}

		// Move back enough (as we need to read forward).
		_, err := parser.rs.Seek(-offset-buflen, io.SeekEnd)
		if err != nil {
			return err
		}

		// Read the data.
		b1 := make([]byte, buflen)
		parser.rs.Read(b1)
		common.Log.Trace("Looking for EOF marker: \"%s\"", string(b1))
		ind := reEOF.FindAllStringIndex(string(b1), -1)
		if ind != nil {
			// Found it.
			lastInd := ind[len(ind)-1]
			common.Log.Trace("Ind: % d", ind)
			parser.rs.Seek(-offset-buflen+int64(lastInd[0]), io.SeekEnd)
			return nil
		}
		common.Log.Debug("Warning: EOF marker not found! - continue seeking")
		offset += buflen
	}

	common.Log.Debug("Error: EOF marker was not found.")
	return errors.New("EOF not found")
}

// Parse an indirect object from the input stream. Can also be an object stream.
// Returns the indirect object (*PdfIndirectObject) or the stream object (*PdfObjectStream).
func (parser *fdfParser) parseIndirectObject() (core.PdfObject, error) {
	indirect := core.PdfIndirectObject{}

	common.Log.Trace("-Read indirect obj")
	bb, err := parser.reader.Peek(20)
	if err != nil {
		common.Log.Debug("ERROR: Fail to read indirect obj")
		return &indirect, err
	}
	common.Log.Trace("(indirect obj peek \"%s\"", string(bb))

	indices := reIndirectObject.FindStringSubmatchIndex(string(bb))
	if len(indices) < 6 {
		common.Log.Debug("ERROR: Unable to find object signature (%s)", string(bb))
		return &indirect, errors.New("unable to detect indirect object signature")
	}
	parser.reader.Discard(indices[0]) // Take care of any small offset.
	common.Log.Trace("Offsets % d", indices)

	// Read the object header.
	hlen := indices[1] - indices[0]
	hb := make([]byte, hlen)
	_, err = parser.readAtLeast(hb, hlen)
	if err != nil {
		common.Log.Debug("ERROR: unable to read - %s", err)
		return nil, err
	}
	common.Log.Trace("textline: %s", hb)

	result := reIndirectObject.FindStringSubmatch(string(hb))
	if len(result) < 3 {
		common.Log.Debug("ERROR: Unable to find object signature (%s)", string(hb))
		return &indirect, errors.New("unable to detect indirect object signature")
	}

	on, _ := strconv.Atoi(result[1])
	gn, _ := strconv.Atoi(result[2])
	indirect.ObjectNumber = int64(on)
	indirect.GenerationNumber = int64(gn)

	for {
		bb, err := parser.reader.Peek(2)
		if err != nil {
			return &indirect, err
		}
		common.Log.Trace("Ind. peek: %s (% x)!", string(bb), string(bb))

		if core.IsWhiteSpace(bb[0]) {
			parser.skipSpaces()
		} else if bb[0] == '%' {
			parser.skipComments()
		} else if (bb[0] == '<') && (bb[1] == '<') {
			common.Log.Trace("Call ParseDict")
			indirect.PdfObject, err = parser.parseDict()
			common.Log.Trace("EOF Call ParseDict: %v", err)
			if err != nil {
				return &indirect, err
			}
			common.Log.Trace("Parsed dictionary... finished.")
		} else if (bb[0] == '/') || (bb[0] == '(') || (bb[0] == '[') || (bb[0] == '<') {
			indirect.PdfObject, err = parser.parseObject()
			if err != nil {
				return &indirect, err
			}
			common.Log.Trace("Parsed object ... finished.")
		} else {
			if bb[0] == 'e' {
				lineStr, err := parser.readTextLine()
				if err != nil {
					return nil, err
				}
				if len(lineStr) >= 6 && lineStr[0:6] == "endobj" {
					break
				}
			} else if bb[0] == 's' {
				bb, _ = parser.reader.Peek(10)
				if string(bb[:6]) == "stream" {
					discardBytes := 6
					if len(bb) > 6 {
						if core.IsWhiteSpace(bb[discardBytes]) && bb[discardBytes] != '\r' && bb[discardBytes] != '\n' {
							// If any other white space character... should not happen!
							// Skip it..
							common.Log.Debug("Non-conformant FDF not ending stream line properly with EOL marker")
							discardBytes++
						}
						if bb[discardBytes] == '\r' {
							discardBytes++
							if bb[discardBytes] == '\n' {
								discardBytes++
							}
						} else if bb[discardBytes] == '\n' {
							discardBytes++
						}
					}

					parser.reader.Discard(discardBytes)

					dict, isDict := indirect.PdfObject.(*core.PdfObjectDictionary)
					if !isDict {
						return nil, errors.New("stream object missing dictionary")
					}
					common.Log.Trace("Stream dict %s", dict)

					pstreamLength, ok := dict.Get("Length").(*core.PdfObjectInteger)
					if !ok {
						return nil, errors.New("stream length needs to be an integer")
					}
					streamLength := *pstreamLength
					if streamLength < 0 {
						return nil, errors.New("stream needs to be longer than 0")
					}

					// Make sure is less than actual file size.
					if int64(streamLength) > parser.fileSize {
						common.Log.Debug("ERROR: Stream length cannot be larger than file size")
						return nil, errors.New("invalid stream length, larger than file size")
					}

					stream := make([]byte, streamLength)
					_, err = parser.readAtLeast(stream, int(streamLength))
					if err != nil {
						common.Log.Debug("ERROR stream (%d): %X", len(stream), stream)
						common.Log.Debug("ERROR: %v", err)
						return nil, err
					}

					streamobj := core.PdfObjectStream{}
					streamobj.Stream = stream
					streamobj.PdfObjectDictionary = indirect.PdfObject.(*core.PdfObjectDictionary)
					streamobj.ObjectNumber = indirect.ObjectNumber
					streamobj.GenerationNumber = indirect.GenerationNumber

					parser.skipSpaces()
					parser.reader.Discard(9) // endstream
					parser.skipSpaces()
					return &streamobj, nil
				}
			}

			indirect.PdfObject, err = parser.parseObject()
			return &indirect, err
		}
	}
	common.Log.Trace("Returning indirect!")
	return &indirect, nil
}

// newParserFromString parses an FDF from a string.
// Useful for testing purposes.
func newParserFromString(txt string) (*fdfParser, error) {
	parser := fdfParser{}
	buf := []byte(txt)

	bufReader := bytes.NewReader(buf)
	parser.rs = bufReader
	parser.objCache = map[int64]core.PdfObject{}

	bufferedReader := bufio.NewReader(bufReader)
	parser.reader = bufferedReader

	parser.fileSize = int64(len(txt))

	return &parser, parser.parse()
}

// Root returns the Root of the FDF document.
func (parser *fdfParser) Root() (*core.PdfObjectDictionary, error) {
	if parser.trailerDict != nil {
		if rootDict, ok := parser.trace(parser.trailerDict.Get("Root")).(*core.PdfObjectDictionary); ok {
			if fdfDict, ok := parser.trace(rootDict.Get("FDF")).(*core.PdfObjectDictionary); ok {
				return fdfDict, nil
			}
		}
	}

	var keys []int64
	for objNum := range parser.objCache {
		keys = append(keys, objNum)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, objNum := range keys {
		obj := parser.objCache[objNum]
		if rootDict, ok := parser.trace(obj).(*core.PdfObjectDictionary); ok {
			if fdfDict, ok := parser.trace(rootDict.Get("FDF")).(*core.PdfObjectDictionary); ok {
				return fdfDict, nil
			}
		}
	}

	return nil, errors.New("FDF not found")
}

// newParser creates a new parser for a FDF file via ReadSeeker. Loads the cross reference stream and trailer.
// An error is returned on failure.
func newParser(rs io.ReadSeeker) (*fdfParser, error) {
	parser := &fdfParser{}

	parser.rs = rs
	parser.objCache = map[int64]core.PdfObject{}

	// Read from top to bottom...
	// 1. Get the version
	// 2. Sequentially parse indirect objects, until does not match

	majorVersion, minorVersion, err := parser.parseFdfVersion()
	if err != nil {
		common.Log.Error("Unable to parse version: %v", err)
		return nil, err
	}
	parser.majorVersion = majorVersion
	parser.minorVersion = minorVersion

	err = parser.parse()
	return parser, err
}

// trace resolves a PdfObject to direct object, looking up and resolving references as needed.
func (parser *fdfParser) trace(obj core.PdfObject) core.PdfObject {
	switch t := obj.(type) {
	case *core.PdfObjectReference:
		indObj, ok := parser.objCache[t.ObjectNumber].(*core.PdfIndirectObject)
		if ok {
			return indObj.PdfObject
		}
		common.Log.Debug("Type error")
		return nil
	case *core.PdfIndirectObject:
		return t.PdfObject
	}

	return obj
}

// parse runs through the file and parses indirect objects and loads into cache.
func (parser *fdfParser) parse() error {
	// Go to beginning, reset reader.
	parser.rs.Seek(0, io.SeekStart)
	parser.reader = bufio.NewReader(parser.rs)

	// Parse indirect objects sequentially.
	for {
		parser.skipComments()

		bb, err := parser.reader.Peek(20)
		if err != nil {
			common.Log.Debug("ERROR: Fail to read indirect obj")
			return err
		}

		if strings.HasPrefix(string(bb), "trailer") {
			// End
			parser.reader.Discard(7)
			parser.skipSpaces()
			parser.skipComments()
			trailerDict, _ := parser.parseDict()
			parser.trailerDict = trailerDict
			break
		}

		indices := reIndirectObject.FindStringSubmatchIndex(string(bb))
		if len(indices) < 6 {
			common.Log.Debug("ERROR: Unable to find object signature (%s)", string(bb))
			return errors.New("unable to detect indirect object signature")
		}

		indObj, err := parser.parseIndirectObject()
		if err != nil {
			return err
		}

		switch o := indObj.(type) {
		case *core.PdfIndirectObject:
			parser.objCache[o.ObjectNumber] = o
		case *core.PdfObjectStream:
			parser.objCache[o.ObjectNumber] = o
		default:
			return errors.New("type error")
		}

	}

	return nil
}

// Called when Fdf version not found normally.  Looks for the PDF version by scanning top-down.
// %FDF-1.4
func (parser *fdfParser) seekFdfVersionTopDown() (int, int, error) {
	// Go to beginning, reset reader.
	parser.rs.Seek(0, io.SeekStart)
	parser.reader = bufio.NewReader(parser.rs)

	// Keep a running buffer of last bytes.
	bufLen := 20
	last := make([]byte, bufLen)

	for {
		b, err := parser.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return 0, 0, err
			}
		}

		// Format:
		// object number - whitespace - generation number - obj
		// e.g. "12 0 obj"
		if core.IsDecimalDigit(b) && last[bufLen-1] == '.' && core.IsDecimalDigit(last[bufLen-2]) && last[bufLen-3] == '-' &&
			last[bufLen-4] == 'F' && last[bufLen-5] == 'D' && last[bufLen-6] == 'P' {
			major := int(last[bufLen-2] - '0')
			minor := int(b - '0')
			return major, minor, nil
		}

		last = append(last[1:bufLen], b)
	}

	return 0, 0, errors.New("version not found")
}
