/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core/security"
)

// Regular Expressions for parsing and identifying object signatures.
var rePdfVersion = regexp.MustCompile(`%PDF-(\d)\.(\d)`)
var reEOF = regexp.MustCompile("%%EOF?")
var reXrefTable = regexp.MustCompile(`\s*xref\s*`)
var reStartXref = regexp.MustCompile(`startx?ref\s*(\d+)`)
var reNumeric = regexp.MustCompile(`^[\+-.]*([0-9.]+)`)
var reExponential = regexp.MustCompile(`^[\+-.]*([0-9.]+)[eE][\+-.]*([0-9.]+)`)
var reReference = regexp.MustCompile(`^\s*[-]*(\d+)\s+(\d+)\s+R`)
var reIndirectObject = regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)
var reXrefSubsection = regexp.MustCompile(`(\d+)\s+(\d+)\s*$`)
var reXrefEntry = regexp.MustCompile(`(\d+)\s+(\d+)\s+([nf])\s*$`)

// PdfParser parses a PDF file and provides access to the object structure of the PDF.
type PdfParser struct {
	version Version

	rs               io.ReadSeeker
	reader           *bufio.Reader
	fileSize         int64
	xrefs            XrefTable
	xrefOffset       int64     // Offset of first xref object.
	xrefType         *xrefType // Type of first xref object.
	objstms          objectStreams
	trailer          *PdfObjectDictionary
	crypter          *PdfCrypt
	repairsAttempted bool // Avoid multiple attempts for repair.

	ObjCache objectCache

	// Tracker for reference lookups when looking up Length entry of stream objects.
	// The Length entries of stream objects are a special case, as they can require recursive parsing, i.e. look up
	// the length reference (if not object) prior to reading the actual stream.  This has risks of endless looping.
	// Tracking is necessary to avoid recursive loops.
	streamLengthReferenceLookupInProgress map[int64]bool
}

// Version represents a version of a PDF standard.
type Version struct {
	Major int
	Minor int
}

// String returns the PDF version as a string. Implements interface fmt.Stringer.
func (v Version) String() string {
	return fmt.Sprintf("%0d.%0d", v.Major, v.Minor)
}

// PdfVersion returns version of the PDF file.
func (parser *PdfParser) PdfVersion() Version {
	return parser.version
}

// GetCrypter returns the PdfCrypt instance which has information about the PDFs encryption.
func (parser *PdfParser) GetCrypter() *PdfCrypt {
	return parser.crypter
}

// IsAuthenticated returns true if the PDF has already been authenticated for accessing.
func (parser *PdfParser) IsAuthenticated() bool {
	return parser.crypter.authenticated
}

// GetTrailer returns the PDFs trailer dictionary. The trailer dictionary is typically the starting point for a PDF,
// referencing other key objects that are important in the document structure.
func (parser *PdfParser) GetTrailer() *PdfObjectDictionary {
	return parser.trailer
}

// GetXrefTable returns the PDFs xref table.
func (parser *PdfParser) GetXrefTable() XrefTable {
	return parser.xrefs
}

// GetXrefOffset returns the offset of the xref table.
func (parser *PdfParser) GetXrefOffset() int64 {
	return parser.xrefOffset
}

// GetXrefType returns the type of the first xref object (table or stream).
func (parser *PdfParser) GetXrefType() *xrefType {
	return parser.xrefType
}

// Skip over any spaces.
func (parser *PdfParser) skipSpaces() (int, error) {
	cnt := 0
	for {
		b, err := parser.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		if IsWhiteSpace(b) {
			cnt++
		} else {
			parser.reader.UnreadByte()
			break
		}
	}

	return cnt, nil
}

// Skip over comments and spaces. Can handle multi-line comments.
func (parser *PdfParser) skipComments() error {
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
func (parser *PdfParser) readComment() (string, error) {
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
func (parser *PdfParser) readTextLine() (string, error) {
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
func (parser *PdfParser) parseName() (PdfObjectName, error) {
	var r bytes.Buffer
	nameStarted := false
	for {
		bb, err := parser.reader.Peek(1)
		if err == io.EOF {
			break // Can happen when loading from object stream.
		}
		if err != nil {
			return PdfObjectName(r.String()), err
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
				return PdfObjectName(r.String()), fmt.Errorf("invalid name: (%c)", bb[0])
			}
		} else {
			if IsWhiteSpace(bb[0]) {
				break
			} else if (bb[0] == '/') || (bb[0] == '[') || (bb[0] == '(') || (bb[0] == ']') || (bb[0] == '<') || (bb[0] == '>') {
				break // Looks like start of next statement.
			} else if bb[0] == '#' {
				hexcode, err := parser.reader.Peek(3)
				if err != nil {
					return PdfObjectName(r.String()), err
				}

				code, err := hex.DecodeString(string(hexcode[1:3]))
				if err != nil {
					common.Log.Debug("ERROR: Invalid hex following '#', continuing using literal - Output may be incorrect")

					// Treat as literal '#' rather than hex code.
					r.WriteByte('#')

					// Discard just the '#' byte and continue parsing the name.
					parser.reader.Discard(1)
					continue
				}

				// Hex decoding succeeded. Safe to discard all peeked bytes.
				parser.reader.Discard(3)
				r.Write(code)
			} else {
				b, _ := parser.reader.ReadByte()
				r.WriteByte(b)
			}
		}
	}
	return PdfObjectName(r.String()), nil
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
func (parser *PdfParser) parseNumber() (PdfObject, error) {
	return ParseNumber(parser.reader)
}

// A string starts with '(' and ends with ')'.
func (parser *PdfParser) parseString() (*PdfObjectString, error) {
	parser.reader.ReadByte()

	var r bytes.Buffer
	count := 1
	for {
		bb, err := parser.reader.Peek(1)
		if err != nil {
			return MakeString(r.String()), err
		}

		if bb[0] == '\\' { // Escape sequence.
			parser.reader.ReadByte() // Skip the escape \ byte.
			b, err := parser.reader.ReadByte()
			if err != nil {
				return MakeString(r.String()), err
			}

			// Octal '\ddd' number (base 8).
			if IsOctalDigit(b) {
				bb, err := parser.reader.Peek(2)
				if err != nil {
					return MakeString(r.String()), err
				}

				var numeric []byte
				numeric = append(numeric, b)
				for _, val := range bb {
					if IsOctalDigit(val) {
						numeric = append(numeric, val)
					} else {
						break
					}
				}
				parser.reader.Discard(len(numeric) - 1)

				common.Log.Trace("Numeric string \"%s\"", numeric)
				code, err := strconv.ParseUint(string(numeric), 8, 32)
				if err != nil {
					return MakeString(r.String()), err
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

	return MakeString(r.String()), nil
}

// Starts with '<' ends with '>'.
// Currently not converting the hex codes to characters.
func (parser *PdfParser) parseHexString() (*PdfObjectString, error) {
	parser.reader.ReadByte()

	var r bytes.Buffer
	for {
		bb, err := parser.reader.Peek(1)
		if err != nil {
			return MakeString(""), err
		}

		if bb[0] == '>' {
			parser.reader.ReadByte()
			break
		}

		b, _ := parser.reader.ReadByte()
		if !IsWhiteSpace(b) {
			r.WriteByte(b)
		}
	}

	if r.Len()%2 == 1 {
		r.WriteRune('0')
	}

	buf, _ := hex.DecodeString(r.String())
	return MakeHexString(string(buf)), nil
}

// Starts with '[' ends with ']'.  Can contain any kinds of direct objects.
func (parser *PdfParser) parseArray() (*PdfObjectArray, error) {
	arr := MakeArray()

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
func (parser *PdfParser) parseBool() (PdfObjectBool, error) {
	bb, err := parser.reader.Peek(4)
	if err != nil {
		return PdfObjectBool(false), err
	}
	if (len(bb) >= 4) && (string(bb[:4]) == "true") {
		parser.reader.Discard(4)
		return PdfObjectBool(true), nil
	}

	bb, err = parser.reader.Peek(5)
	if err != nil {
		return PdfObjectBool(false), err
	}
	if (len(bb) >= 5) && (string(bb[:5]) == "false") {
		parser.reader.Discard(5)
		return PdfObjectBool(false), nil
	}

	return PdfObjectBool(false), errors.New("unexpected boolean string")
}

// Parse reference to an indirect object.
func parseReference(refStr string) (PdfObjectReference, error) {
	objref := PdfObjectReference{}

	result := reReference.FindStringSubmatch(string(refStr))
	if len(result) < 3 {
		common.Log.Debug("Error parsing reference")
		return objref, errors.New("unable to parse reference")
	}

	objNum, _ := strconv.Atoi(result[1])
	genNum, _ := strconv.Atoi(result[2])
	objref.ObjectNumber = int64(objNum)
	objref.GenerationNumber = int64(genNum)

	return objref, nil
}

// Parse null object.
func (parser *PdfParser) parseNull() (PdfObjectNull, error) {
	_, err := parser.reader.Discard(4)
	return PdfObjectNull{}, err
}

// Detect the signature at the current file position and parse
// the corresponding object.
func (parser *PdfParser) parseObject() (PdfObject, error) {
	common.Log.Trace("Read direct object")
	parser.skipSpaces()
	for {
		bb, err := parser.reader.Peek(2)
		if err != nil {
			// If EOFs after 1 byte then should still try to continue parsing.
			if err != io.EOF || len(bb) == 0 {
				return nil, err
			}
			if len(bb) == 1 {
				// Add space as code below is expecting 2 bytes.
				bb = append(bb, ' ')
			}
		}

		common.Log.Trace("Peek string: %s", string(bb))
		// Determine type.
		if bb[0] == '/' {
			name, err := parser.parseName()
			common.Log.Trace("->Name: '%s'", name)
			return &name, err
		} else if bb[0] == '(' {
			common.Log.Trace("->String!")
			str, err := parser.parseString()
			return str, err
		} else if bb[0] == '[' {
			common.Log.Trace("->Array!")
			arr, err := parser.parseArray()
			return arr, err
		} else if (bb[0] == '<') && (bb[1] == '<') {
			common.Log.Trace("->Dict!")
			dict, err := parser.ParseDict()
			return dict, err
		} else if bb[0] == '<' {
			common.Log.Trace("->Hex string!")
			str, err := parser.parseHexString()
			return str, err
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
				ref.parser = parser
				return &ref, err
			}

			result2 := reNumeric.FindStringSubmatch(string(peekStr))
			if len(result2) > 1 {
				// Number object.
				common.Log.Trace("-> Number!")
				num, err := parser.parseNumber()
				return num, err
			}

			result2 = reExponential.FindStringSubmatch(string(peekStr))
			if len(result2) > 1 {
				// Number object (exponential)
				common.Log.Trace("-> Exponential Number!")
				common.Log.Trace("% s", result2)
				num, err := parser.parseNumber()
				return num, err
			}

			common.Log.Debug("ERROR Unknown (peek \"%s\")", peekStr)
			return nil, errors.New("object parsing error - unexpected pattern")
		}
	}
}

// ParseDict reads and parses a PDF dictionary object enclosed with '<<' and '>>'
func (parser *PdfParser) ParseDict() (*PdfObjectDictionary, error) {
	common.Log.Trace("Reading PDF Dict!")

	dict := MakeDict()
	dict.parser = parser

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
				dict.Set(newKey, MakeNull())
				continue
			}
		}

		parser.skipSpaces()

		val, err := parser.parseObject()
		if err != nil {
			return nil, err
		}
		dict.Set(keyName, val)

		if common.Log.IsLogLevel(common.LogLevelTrace) {
			// Avoid calling unless needed as the String() can be heavy for large objects.
			common.Log.Trace("dict[%s] = %s", keyName, val.String())
		}
	}
	common.Log.Trace("returning PDF Dict!")

	return dict, nil
}

// Parse the pdf version from the beginning of the file.
// Returns the major and minor parts of the version.
// E.g. for "PDF-1.7" would return 1 and 7.
func (parser *PdfParser) parsePdfVersion() (int, int, error) {
	var offset int64 = 20
	b := make([]byte, offset)
	parser.rs.Seek(0, os.SEEK_SET)
	parser.rs.Read(b)

	// Try matching the PDF version at the start of the file, within the
	// first 20 bytes. If the PDF version is not found, search for it
	// starting from the top of the file.
	var err error
	var major, minor int

	if match := rePdfVersion.FindStringSubmatch(string(b)); len(match) < 3 {
		if major, minor, err = parser.seekPdfVersionTopDown(); err != nil {
			common.Log.Debug("Failed recovery - unable to find version")
			return 0, 0, err
		}

		// Create a new offset reader that ignores the invalid data before
		// the PDF version. Sets reader offset at the start of the PDF
		// version string.
		parser.rs, err = newOffsetReader(parser.rs, parser.GetFileOffset()-8)
		if err != nil {
			return 0, 0, err
		}
	} else {
		if major, err = strconv.Atoi(match[1]); err != nil {
			return 0, 0, err
		}
		if minor, err = strconv.Atoi(match[2]); err != nil {
			return 0, 0, err
		}

		// Reset parser reader offset.
		parser.SetFileOffset(0)
	}
	parser.reader = bufio.NewReader(parser.rs)

	common.Log.Debug("Pdf version %d.%d", major, minor)
	return major, minor, nil
}

// Conventional xref table starting with 'xref'.
func (parser *PdfParser) parseXrefTable() (*PdfObjectDictionary, error) {
	var trailer *PdfObjectDictionary

	txt, err := parser.readTextLine()
	if err != nil {
		return nil, err
	}

	common.Log.Trace("xref first line: %s", txt)
	curObjNum := -1
	secObjects := 0
	insideSubsection := false
	unmatchedContent := ""
	for {
		parser.skipSpaces()
		_, err := parser.reader.Peek(1)
		if err != nil {
			return nil, err
		}

		txt, err = parser.readTextLine()
		if err != nil {
			return nil, err
		}

		result1 := reXrefSubsection.FindStringSubmatch(txt)
		if len(result1) == 0 {
			// Try to match invalid subsection beginning lines from previously
			// read, unidentified lines. Covers cases in which the object number
			// and the number of entries in the subsection are not on the same line.
			tryMatch := len(unmatchedContent) > 0
			unmatchedContent += txt + "\n"
			if tryMatch {
				result1 = reXrefSubsection.FindStringSubmatch(unmatchedContent)
			}
		}
		if len(result1) == 3 {
			// Match
			first, _ := strconv.Atoi(result1[1])
			second, _ := strconv.Atoi(result1[2])
			curObjNum = first
			secObjects = second
			insideSubsection = true
			unmatchedContent = ""
			common.Log.Trace("xref subsection: first object: %d objects: %d", curObjNum, secObjects)
			continue
		}
		result2 := reXrefEntry.FindStringSubmatch(txt)
		if len(result2) == 4 {
			if insideSubsection == false {
				common.Log.Debug("ERROR Xref invalid format!\n")
				return nil, errors.New("xref invalid format")
			}

			first, _ := strconv.ParseInt(result2[1], 10, 64)
			gen, _ := strconv.Atoi(result2[2])
			third := result2[3]
			unmatchedContent = ""

			if strings.ToLower(third) == "n" && first > 1 {
				// Object in use in the file!  Load it.
				// Ignore free objects ('f').
				//
				// Some malformed writers mark the offset as 0 to
				// indicate that the object is free, and still mark as 'n'
				// Fairly safe to assume is free if offset is 0.
				//
				// Some malformed writers even seem to have values such as
				// 1.. Assume null object for those also. That is referring
				// to within the PDF version in the header clearly.
				//
				// Load if not existing or higher generation number than previous.
				// Usually should not happen, lower generation numbers
				// would be marked as free.  But can still happen!
				x, ok := parser.xrefs.ObjectMap[curObjNum]
				if !ok || gen > x.Generation {
					obj := XrefObject{ObjectNumber: curObjNum,
						XType:  XrefTypeTableEntry,
						Offset: first, Generation: gen}
					parser.xrefs.ObjectMap[curObjNum] = obj
				}
			}

			curObjNum++
			continue
		}

		if (len(txt) > 6) && (txt[:7] == "trailer") {
			common.Log.Trace("Found trailer - %s", txt)
			// Sometimes get "trailer << ...."
			// Need to rewind to end of trailer text.
			if len(txt) > 9 {
				offset := parser.GetFileOffset()
				parser.SetFileOffset(offset - int64(len(txt)) + 7)
			}

			parser.skipSpaces()
			parser.skipComments()
			common.Log.Trace("Reading trailer dict!")
			common.Log.Trace("peek: \"%s\"", txt)
			trailer, err = parser.ParseDict()
			common.Log.Trace("EOF reading trailer dict!")
			if err != nil {
				common.Log.Debug("Error parsing trailer dict (%s)", err)
				return nil, err
			}
			break
		}

		if txt == "%%EOF" {
			common.Log.Debug("ERROR: end of file - trailer not found - error!")
			return nil, errors.New("end of file - trailer not found")
		}

		common.Log.Trace("xref more : %s", txt)
	}
	common.Log.Trace("EOF parsing xref table!")
	if parser.xrefType == nil {
		t := XrefTypeTableEntry
		parser.xrefType = &t
	}

	return trailer, nil
}

// Load the cross references from an xref stream object (XRefStm).
// Also load the dictionary information (trailer dictionary).
func (parser *PdfParser) parseXrefStream(xstm *PdfObjectInteger) (*PdfObjectDictionary, error) {
	if xstm != nil {
		common.Log.Trace("XRefStm xref table object at %d", xstm)
		parser.rs.Seek(int64(*xstm), io.SeekStart)
		parser.reader = bufio.NewReader(parser.rs)
	}

	xsOffset := parser.GetFileOffset()

	xrefObj, err := parser.ParseIndirectObject()
	if err != nil {
		common.Log.Debug("ERROR: Failed to read xref object")
		return nil, errors.New("failed to read xref object")
	}

	common.Log.Trace("XRefStm object: %s", xrefObj)
	xs, ok := xrefObj.(*PdfObjectStream)
	if !ok {
		common.Log.Debug("ERROR: XRefStm pointing to non-stream object!")
		return nil, errors.New("XRefStm pointing to a non-stream object")
	}

	trailerDict := xs.PdfObjectDictionary

	sizeObj, ok := xs.PdfObjectDictionary.Get("Size").(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("ERROR: Missing size from xref stm")
		return nil, errors.New("missing Size from xref stm")
	}
	// Sanity check to avoid DoS attacks. Maximum number of indirect objects on 32 bit system.
	if int64(*sizeObj) > 8388607 {
		common.Log.Debug("ERROR: xref Size exceeded limit, over 8388607 (%d)", *sizeObj)
		return nil, errors.New("range check error")
	}

	wObj := xs.PdfObjectDictionary.Get("W")
	wArr, ok := wObj.(*PdfObjectArray)
	if !ok {
		return nil, errors.New("invalid W in xref stream")
	}

	wLen := wArr.Len()
	if wLen != 3 {
		common.Log.Debug("ERROR: Unsupported xref stm (len(W) != 3 - %d)", wLen)
		return nil, errors.New("unsupported xref stm len(W) != 3")
	}

	var b []int64
	for i := 0; i < 3; i++ {
		wVal, ok := GetInt(wArr.Get(i))
		if !ok {
			return nil, errors.New("invalid w object type")
		}

		b = append(b, int64(*wVal))
	}

	ds, err := DecodeStream(xs)
	if err != nil {
		common.Log.Debug("ERROR: Unable to decode stream: %v", err)
		return nil, err
	}

	s0 := int(b[0])
	s1 := int(b[0] + b[1])
	s2 := int(b[0] + b[1] + b[2])
	deltab := int(b[0] + b[1] + b[2])

	if s0 < 0 || s1 < 0 || s2 < 0 {
		common.Log.Debug("Error s value < 0 (%d,%d,%d)", s0, s1, s2)
		return nil, errors.New("range check error")
	}
	if deltab == 0 {
		common.Log.Debug("No xref objects in stream (deltab == 0)")
		return trailerDict, nil
	}

	// Calculate expected entries.
	entries := len(ds) / deltab

	// Get the object indices.

	objCount := 0
	indexObj := xs.PdfObjectDictionary.Get("Index")
	// Table 17 (7.5.8.2 Cross-Reference Stream Dictionary)
	// (Optional) An array containing a pair of integers for each
	// subsection in this section. The first integer shall be the first
	// object number in the subsection; the second integer shall be the
	// number of entries in the subsection.
	// The array shall be sorted in ascending order by object number.
	// Subsections cannot overlap; an object number may have at most
	// one entry in a section.
	// Default value: [0 Size].
	var indexList []int
	if indexObj != nil {
		common.Log.Trace("Index: %b", indexObj)
		indicesArray, ok := indexObj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Invalid Index object (should be an array)")
			return nil, errors.New("invalid Index object")
		}

		// Expect indLen to be a multiple of 2.
		if indicesArray.Len()%2 != 0 {
			common.Log.Debug("WARNING Failure loading xref stm index not multiple of 2.")
			return nil, errors.New("range check error")
		}

		objCount = 0

		indices, err := indicesArray.ToIntegerArray()
		if err != nil {
			common.Log.Debug("Error getting index array as integers: %v", err)
			return nil, err
		}

		for i := 0; i < len(indices); i += 2 {
			// add the indices to the list..

			startIdx := indices[i]
			numObjs := indices[i+1]
			for j := 0; j < numObjs; j++ {
				indexList = append(indexList, startIdx+j)
			}
			objCount += numObjs
		}
	} else {
		// If no Index, then assume [0 Size]
		for i := 0; i < int(*sizeObj); i++ {
			indexList = append(indexList, i)
		}
		objCount = int(*sizeObj)
	}

	if entries == objCount+1 {
		// For compatibility, expand the object count.
		common.Log.Debug("Incompatibility: Index missing coverage of 1 object - appending one - May lead to problems")
		maxIndex := objCount - 1
		for _, ind := range indexList {
			if ind > maxIndex {
				maxIndex = ind
			}
		}
		indexList = append(indexList, maxIndex+1)
		objCount++
	}

	if entries != len(indexList) {
		// If mismatch -> error (already allowing mismatch of 1 if Index not specified).
		common.Log.Debug("ERROR: xref stm: num entries != len(indices) (%d != %d)", entries, len(indexList))
		return nil, errors.New("xref stm num entries != len(indices)")
	}

	common.Log.Trace("Objects count %d", objCount)
	common.Log.Trace("Indices: % d", indexList)

	// Convert byte array to a larger integer, little-endian.
	convertBytes := func(v []byte) int64 {
		var tmp int64
		for i := 0; i < len(v); i++ {
			tmp += int64(v[i]) * (1 << uint(8*(len(v)-i-1)))
		}
		return tmp
	}

	common.Log.Trace("Decoded stream length: %d", len(ds))
	objIndex := 0
	for i := 0; i < len(ds); i += deltab {
		err := checkBounds(len(ds), i, i+s0)
		if err != nil {
			common.Log.Debug("Invalid slice range: %v", err)
			return nil, err
		}
		p1 := ds[i : i+s0]

		err = checkBounds(len(ds), i+s0, i+s1)
		if err != nil {
			common.Log.Debug("Invalid slice range: %v", err)
			return nil, err
		}
		p2 := ds[i+s0 : i+s1]

		err = checkBounds(len(ds), i+s1, i+s2)
		if err != nil {
			common.Log.Debug("Invalid slice range: %v", err)
			return nil, err
		}
		p3 := ds[i+s1 : i+s2]

		ftype := convertBytes(p1)
		n2 := convertBytes(p2)
		n3 := convertBytes(p3)

		if b[0] == 0 {
			// If first entry in W is 0, then default to to type 1.
			// (uncompressed object via offset).
			ftype = 1
		}

		if objIndex >= len(indexList) {
			common.Log.Debug("XRef stream - Trying to access index out of bounds - breaking")
			break
		}
		objNum := indexList[objIndex]
		objIndex++

		common.Log.Trace("%d. p1: % x", objNum, p1)
		common.Log.Trace("%d. p2: % x", objNum, p2)
		common.Log.Trace("%d. p3: % x", objNum, p3)

		common.Log.Trace("%d. xref: %d %d %d", objNum, ftype, n2, n3)
		if ftype == 0 {
			common.Log.Trace("- Free object - can probably ignore")
		} else if ftype == 1 {
			common.Log.Trace("- In use - uncompressed via offset %b", p2)
			// If offset (n2) is same as the XRefs table offset, then update the Object number with the
			// one that was parsed.  Fixes problem where the object number is incorrectly or not specified
			// in the Index.
			if n2 == xsOffset {
				common.Log.Debug("Updating object number for XRef table %d -> %d", objNum, xs.ObjectNumber)
				objNum = int(xs.ObjectNumber)
			}

			// Object type 1: Objects that are in use but are not
			// compressed, i.e. defined by an offset (normal entry)
			if xr, ok := parser.xrefs.ObjectMap[objNum]; !ok || int(n3) > xr.Generation {
				// Only overload if not already loaded!
				// or has a newer generation number. (should not happen)
				obj := XrefObject{ObjectNumber: objNum,
					XType: XrefTypeTableEntry, Offset: n2, Generation: int(n3)}
				parser.xrefs.ObjectMap[objNum] = obj
			}
		} else if ftype == 2 {
			// Object type 2: Compressed object.
			common.Log.Trace("- In use - compressed object")
			if _, ok := parser.xrefs.ObjectMap[objNum]; !ok {
				obj := XrefObject{ObjectNumber: objNum,
					XType: XrefTypeObjectStream, OsObjNumber: int(n2), OsObjIndex: int(n3)}
				parser.xrefs.ObjectMap[objNum] = obj
				common.Log.Trace("entry: %+v", obj)
			}
		} else {
			common.Log.Debug("ERROR: --------INVALID TYPE XrefStm invalid?-------")
			// Continue, we do not define anything -> null object.
			// 7.5.8.3:
			//
			// In PDF 1.5 through PDF 1.7, only types 0, 1, and 2 are
			// allowed. Any other value shall be interpreted as a
			// reference to the null object, thus permitting new entry
			// types to be defined in the future.
			continue
		}
	}

	if parser.xrefType == nil {
		t := XrefTypeObjectStream
		parser.xrefType = &t
	}

	return trailerDict, nil
}

// Parse xref table at the current file position. Can either be a standard xref
// table, or an xref stream.
func (parser *PdfParser) parseXref() (*PdfObjectDictionary, error) {
	// Search xrefs within 20 bytes of the current location. If the first
	// iteration of the loop is unable to find a match, peek another 20 bytes
	// left of the current location, add them to the previously read buffer
	// and try again.
	const bufLen = 20
	bb, _ := parser.reader.Peek(bufLen)
	for i := 0; i < 2; i++ {
		if parser.xrefOffset == 0 {
			parser.xrefOffset = parser.GetFileOffset()
		}
		if reIndirectObject.Match(bb) {
			common.Log.Trace("xref points to an object. Probably xref object")
			common.Log.Debug("starting with \"%s\"", string(bb))
			return parser.parseXrefStream(nil)
		}
		if reXrefTable.Match(bb) {
			common.Log.Trace("Standard xref section table!")
			return parser.parseXrefTable()
		}

		// xref match failed. Peek 20 bytes to the left of the current offset,
		// append them to the previously read buffer and try again. Reset to the
		// original offset after reading.
		offset := parser.GetFileOffset()
		if parser.xrefOffset == 0 {
			parser.xrefOffset = offset
		}
		parser.SetFileOffset(offset - bufLen)
		defer parser.SetFileOffset(offset)

		lbb, _ := parser.reader.Peek(bufLen)
		bb = append(lbb, bb...)
	}

	common.Log.Debug("Warning: Unable to find xref table or stream. Repair attempted: Looking for earliest xref from bottom.")
	if err := parser.repairSeekXrefMarker(); err != nil {
		common.Log.Debug("Repair failed - %v", err)
		return nil, err
	}
	return parser.parseXrefTable()
}

// Look for EOF marker and seek to its beginning.
// Define an offset position from the end of the file.
func (parser *PdfParser) seekToEOFMarker(fSize int64) error {
	// Define the starting point (from the end of the file) to search from.
	var offset int64

	// Define an buffer length in terms of how many bytes to read from the end of the file.
	var buflen int64 = 2048

	for offset < fSize-4 {
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
		offset += buflen - 4
	}

	common.Log.Debug("Error: EOF marker was not found.")
	return errors.New("EOF not found")
}

//
// Load the xrefs from the bottom of file prior to parsing the file.
// 1. Look for %%EOF marker, then
// 2. Move up to find startxref
// 3. Then move to that position (slight offset)
// 4. Move until find "startxref"
// 5. Load the xref position
// 6. Move to the xref position and parse it.
// 7. Load each xref into a table.
//
// Multiple xref table handling:
// 1. Check main xref table (primary)
// 2. Check the Xref stream object (PDF >=1.5)
// 3. Check the Prev xref
// 4. Continue looking for Prev until not found.
//
// The earlier xrefs have higher precedence.  If objects already
// loaded will ignore older versions.
//
func (parser *PdfParser) loadXrefs() (*PdfObjectDictionary, error) {
	parser.xrefs.ObjectMap = make(map[int]XrefObject)
	parser.objstms = make(objectStreams)

	// Get the file size.
	fSize, err := parser.rs.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	common.Log.Trace("fsize: %d", fSize)
	parser.fileSize = fSize

	// Seek the EOF marker.
	err = parser.seekToEOFMarker(fSize)
	if err != nil {
		common.Log.Debug("Failed seek to eof marker: %v", err)
		return nil, err
	}

	// Look for startxref and get the xref offset.
	curOffset, err := parser.rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	// Seek 64 bytes (numBytes) back from EOF marker start.
	var numBytes int64 = 64
	offset := curOffset - numBytes
	if offset < 0 {
		offset = 0
	}
	_, err = parser.rs.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	b2 := make([]byte, numBytes)
	_, err = parser.rs.Read(b2)
	if err != nil {
		common.Log.Debug("Failed reading while looking for startxref: %v", err)
		return nil, err
	}

	result := reStartXref.FindStringSubmatch(string(b2))
	if len(result) < 2 {
		common.Log.Debug("Error: startxref not found!")
		return nil, errors.New("startxref not found")
	}
	if len(result) > 2 {
		common.Log.Debug("ERROR: Multiple startxref (%s)!", b2)
		return nil, errors.New("multiple startxref entries?")
	}
	offsetXref, _ := strconv.ParseInt(result[1], 10, 64)
	common.Log.Trace("startxref at %d", offsetXref)

	if offsetXref > fSize {
		common.Log.Debug("ERROR: Xref offset outside of file")
		common.Log.Debug("Attempting repair")
		offsetXref, err = parser.repairLocateXref()
		if err != nil {
			common.Log.Debug("ERROR: Repair attempt failed (%s)")
			return nil, err
		}
	}
	// Read the xref.
	parser.rs.Seek(int64(offsetXref), io.SeekStart)
	parser.reader = bufio.NewReader(parser.rs)

	trailerDict, err := parser.parseXref()
	if err != nil {
		return nil, err
	}

	// Check the XrefStm object also from the trailer.
	xx := trailerDict.Get("XRefStm")
	if xx != nil {
		xo, ok := xx.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("XRefStm != int")
		}
		_, err = parser.parseXrefStream(xo)
		if err != nil {
			return nil, err
		}
	}

	// Load old objects also.  Only if not already specified.
	var prevList []int64
	intInSlice := func(val int64, list []int64) bool {
		for _, b := range list {
			if b == val {
				return true
			}
		}
		return false
	}

	// Load any Previous xref tables (old versions), which can
	// refer to objects also.
	xx = trailerDict.Get("Prev")
	for xx != nil {
		prevInt, ok := xx.(*PdfObjectInteger)
		if !ok {
			// For compatibility: If Prev is invalid, just go with whatever xrefs are loaded already.
			// i.e. not returning an error.  A debug message is logged.
			common.Log.Debug("Invalid Prev reference: Not a *PdfObjectInteger (%T)", xx)
			return trailerDict, nil
		}

		off := *prevInt
		common.Log.Trace("Another Prev xref table object at %d", off)

		// Can be either regular table, or an xref object...
		parser.rs.Seek(int64(off), os.SEEK_SET)
		parser.reader = bufio.NewReader(parser.rs)

		ptrailerDict, err := parser.parseXref()
		if err != nil {
			common.Log.Debug("Warning: Error - Failed loading another (Prev) trailer")
			common.Log.Debug("Attempting to continue by ignoring it")
			break
		}

		xx = ptrailerDict.Get("Prev")
		if xx != nil {
			prevoff := *(xx.(*PdfObjectInteger))
			if intInSlice(int64(prevoff), prevList) {
				// Prevent circular reference!
				common.Log.Debug("Preventing circular xref referencing")
				break
			}
			prevList = append(prevList, int64(prevoff))
		}
	}

	return trailerDict, nil
}

// Return the closest object following offset from the xrefs table.
func (parser *PdfParser) xrefNextObjectOffset(offset int64) int64 {
	nextOffset := int64(0)

	if len(parser.xrefs.ObjectMap) == 0 {
		return 0
	}

	if len(parser.xrefs.sortedObjects) == 0 {
		count := 0
		for _, xref := range parser.xrefs.ObjectMap {
			if xref.Offset > 0 {
				count++
			}
		}
		if count == 0 {
			// No objects with offset.
			return 0
		}
		parser.xrefs.sortedObjects = make([]XrefObject, count)

		i := 0
		for _, xref := range parser.xrefs.ObjectMap {
			if xref.Offset > 0 {
				parser.xrefs.sortedObjects[i] = xref
				i++
			}
		}

		// Sort by offset, ascending.
		sort.Slice(parser.xrefs.sortedObjects, func(i, j int) bool {
			return parser.xrefs.sortedObjects[i].Offset < parser.xrefs.sortedObjects[j].Offset
		})
	}

	i := sort.Search(len(parser.xrefs.sortedObjects), func(i int) bool {
		return parser.xrefs.sortedObjects[i].Offset >= offset
	})
	if i < len(parser.xrefs.sortedObjects) {
		nextOffset = parser.xrefs.sortedObjects[i].Offset
	}

	return nextOffset
}

// Get stream length, avoiding recursive loops.
// The input is the PdfObject that is to be traced to a direct object.
func (parser *PdfParser) traceStreamLength(lengthObj PdfObject) (PdfObject, error) {
	lengthRef, isRef := lengthObj.(*PdfObjectReference)
	if isRef {
		lookupInProgress, has := parser.streamLengthReferenceLookupInProgress[lengthRef.ObjectNumber]
		if has && lookupInProgress {
			common.Log.Debug("Stream Length reference unresolved (illegal)")
			return nil, errors.New("illegal recursive loop")
		}
		// Mark lookup as in progress.
		parser.streamLengthReferenceLookupInProgress[lengthRef.ObjectNumber] = true
	}

	slo, err := parser.Resolve(lengthObj)
	if err != nil {
		return nil, err
	}
	common.Log.Trace("Stream length? %s", slo)

	if isRef {
		// Mark as completed lookup
		parser.streamLengthReferenceLookupInProgress[lengthRef.ObjectNumber] = false
	}

	return slo, nil
}

// ParseIndirectObject parses an indirect object from the input stream. Can also be an object stream.
// Returns the indirect object (*PdfIndirectObject) or the stream object (*PdfObjectStream).
func (parser *PdfParser) ParseIndirectObject() (PdfObject, error) {
	indirect := PdfIndirectObject{}
	indirect.parser = parser
	common.Log.Trace("-Read indirect obj")
	bb, err := parser.reader.Peek(20)
	if err != nil {
		if err != io.EOF {
			common.Log.Debug("ERROR: Fail to read indirect obj")
			return &indirect, err
		}
	}
	common.Log.Trace("(indirect obj peek \"%s\"", string(bb))

	indices := reIndirectObject.FindStringSubmatchIndex(string(bb))
	if len(indices) < 6 {
		if err == io.EOF {
			// If an EOF error occurred above and the object signature was not found, then return
			// with the EOF error.
			return nil, err
		}
		common.Log.Debug("ERROR: Unable to find object signature (%s)", string(bb))
		return &indirect, errors.New("unable to detect indirect object signature")
	}
	parser.reader.Discard(indices[0]) // Take care of any small offset.
	common.Log.Trace("Offsets % d", indices)

	// Read the object header.
	hlen := indices[1] - indices[0]
	hb := make([]byte, hlen)
	_, err = parser.ReadAtLeast(hb, hlen)
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

		if IsWhiteSpace(bb[0]) {
			parser.skipSpaces()
		} else if bb[0] == '%' {
			parser.skipComments()
		} else if (bb[0] == '<') && (bb[1] == '<') {
			common.Log.Trace("Call ParseDict")
			indirect.PdfObject, err = parser.ParseDict()
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
		} else if bb[0] == ']' {
			// ']' not used as an array object ending marker, or array object
			// terminated multiple times. Discarding the character.
			common.Log.Debug("WARNING: ']' character not being used as an array ending marker. Skipping.")
			parser.reader.Discard(1)
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
						if IsWhiteSpace(bb[discardBytes]) && bb[discardBytes] != '\r' && bb[discardBytes] != '\n' {
							// If any other white space character... should not happen!
							// Skip it..
							common.Log.Debug("Non-conformant PDF not ending stream line properly with EOL marker")
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

					dict, isDict := indirect.PdfObject.(*PdfObjectDictionary)
					if !isDict {
						return nil, errors.New("stream object missing dictionary")
					}
					common.Log.Trace("Stream dict %s", dict)

					// Special stream length tracing function used to avoid endless recursive looping.
					slo, err := parser.traceStreamLength(dict.Get("Length"))
					if err != nil {
						common.Log.Debug("Fail to trace stream length: %v", err)
						return nil, err
					}
					common.Log.Trace("Stream length? %s", slo)

					pstreamLength, ok := slo.(*PdfObjectInteger)
					if !ok {
						return nil, errors.New("stream length needs to be an integer")
					}
					streamLength := *pstreamLength
					if streamLength < 0 {
						return nil, errors.New("stream needs to be longer than 0")
					}

					// Validate the stream length based on the cross references.
					// Find next object with closest offset to current object and calculate
					// the expected stream length based on that.
					streamStartOffset := parser.GetFileOffset()
					nextObjectOffset := parser.xrefNextObjectOffset(streamStartOffset)
					if streamStartOffset+int64(streamLength) > nextObjectOffset && nextObjectOffset > streamStartOffset {
						common.Log.Debug("Expected ending at %d", streamStartOffset+int64(streamLength))
						common.Log.Debug("Next object starting at %d", nextObjectOffset)
						// endstream + "\n" endobj + "\n" (17)
						newLength := nextObjectOffset - streamStartOffset - 17
						if newLength < 0 {
							return nil, errors.New("invalid stream length, going past boundaries")
						}

						common.Log.Debug("Attempting a length correction to %d...", newLength)
						streamLength = PdfObjectInteger(newLength)
						dict.Set("Length", MakeInteger(newLength))
					}

					// Make sure is less than actual file size.
					if int64(streamLength) > parser.fileSize {
						common.Log.Debug("ERROR: Stream length cannot be larger than file size")
						return nil, errors.New("invalid stream length, larger than file size")
					}

					stream := make([]byte, streamLength)
					_, err = parser.ReadAtLeast(stream, int(streamLength))
					if err != nil {
						common.Log.Debug("ERROR stream (%d): %X", len(stream), stream)
						common.Log.Debug("ERROR: %v", err)
						return nil, err
					}

					streamobj := PdfObjectStream{}
					streamobj.Stream = stream
					streamobj.PdfObjectDictionary = indirect.PdfObject.(*PdfObjectDictionary)
					streamobj.ObjectNumber = indirect.ObjectNumber
					streamobj.GenerationNumber = indirect.GenerationNumber
					streamobj.PdfObjectReference.parser = parser

					parser.skipSpaces()
					parser.reader.Discard(9) // endstream
					parser.skipSpaces()
					return &streamobj, nil
				}
			}

			indirect.PdfObject, err = parser.parseObject()
			if indirect.PdfObject == nil {
				common.Log.Debug("INCOMPATIBILITY: Indirect object not containing an object - assuming null object")
				indirect.PdfObject = MakeNull()
			}
			return &indirect, err
		}
	}
	if indirect.PdfObject == nil {
		common.Log.Debug("INCOMPATIBILITY: Indirect object not containing an object - assuming null object")
		indirect.PdfObject = MakeNull()
	}
	common.Log.Trace("Returning indirect!")
	return &indirect, nil
}

// NewParserFromString is used for testing purposes.
func NewParserFromString(txt string) *PdfParser {
	bufReader := bytes.NewReader([]byte(txt))

	parser := &PdfParser{
		ObjCache:                              objectCache{},
		rs:                                    bufReader,
		reader:                                bufio.NewReader(bufReader),
		fileSize:                              int64(len(txt)),
		streamLengthReferenceLookupInProgress: map[int64]bool{},
	}
	parser.xrefs.ObjectMap = make(map[int]XrefObject)

	return parser
}

// NewParser creates a new parser for a PDF file via ReadSeeker. Loads the cross reference stream and trailer.
// An error is returned on failure.
func NewParser(rs io.ReadSeeker) (*PdfParser, error) {
	parser := &PdfParser{
		rs:                                    rs,
		ObjCache:                              make(objectCache),
		streamLengthReferenceLookupInProgress: map[int64]bool{},
	}

	// Parse PDF version.
	majorVersion, minorVersion, err := parser.parsePdfVersion()
	if err != nil {
		common.Log.Error("Unable to parse version: %v", err)
		return nil, err
	}
	parser.version.Major = majorVersion
	parser.version.Minor = minorVersion

	// Start by reading the xrefs (from bottom).
	if parser.trailer, err = parser.loadXrefs(); err != nil {
		common.Log.Debug("ERROR: Failed to load xref table! %s", err)
		return nil, err
	}
	common.Log.Trace("Trailer: %s", parser.trailer)

	if len(parser.xrefs.ObjectMap) == 0 {
		return nil, fmt.Errorf("empty XREF table - Invalid")
	}

	return parser, nil
}

// Resolves a reference, returning the object and indicates whether or not it was cached.
func (parser *PdfParser) resolveReference(ref *PdfObjectReference) (PdfObject, bool, error) {
	cachedObj, isCached := parser.ObjCache[int(ref.ObjectNumber)]
	if isCached {
		return cachedObj, true, nil
	}
	obj, err := parser.LookupByReference(*ref)
	if err != nil {
		return nil, false, err
	}
	parser.ObjCache[int(ref.ObjectNumber)] = obj
	return obj, false, nil
}

// IsEncrypted checks if the document is encrypted. A bool flag is returned indicating the result.
// First time when called, will check if the Encrypt dictionary is accessible through the trailer dictionary.
// If encrypted, prepares a crypt datastructure which can be used to authenticate and decrypt the document.
// On failure, an error is returned.
func (parser *PdfParser) IsEncrypted() (bool, error) {
	if parser.crypter != nil {
		return true, nil
	} else if parser.trailer == nil {
		return false, nil
	}

	common.Log.Trace("Checking encryption dictionary!")
	e := parser.trailer.Get("Encrypt")
	if e == nil {
		return false, nil
	}
	common.Log.Trace("Is encrypted!")
	var (
		dict *PdfObjectDictionary
	)
	switch e := e.(type) {
	case *PdfObjectDictionary:
		dict = e
	case *PdfObjectReference:
		common.Log.Trace("0: Look up ref %q", e)
		encObj, err := parser.LookupByReference(*e)
		common.Log.Trace("1: %q", encObj)
		if err != nil {
			return false, err
		}

		encIndObj, ok := encObj.(*PdfIndirectObject)
		if !ok {
			common.Log.Debug("Encryption object not an indirect object")
			return false, errors.New("type check error")
		}
		encDict, ok := encIndObj.PdfObject.(*PdfObjectDictionary)

		common.Log.Trace("2: %q", encDict)
		if !ok {
			return false, errors.New("trailer Encrypt object non dictionary")
		}
		dict = encDict
	case *PdfObjectNull:
		common.Log.Debug("Encrypt is a null object. File should not be encrypted.")
		return false, nil
	default:
		return false, fmt.Errorf("unsupported type: %T", e)
	}

	crypter, err := PdfCryptNewDecrypt(parser, dict, parser.trailer)
	if err != nil {
		return false, err
	}
	// list objects that should never be decrypted
	for _, key := range []string{"Info", "Encrypt"} {
		f := parser.trailer.Get(PdfObjectName(key))
		if f == nil {
			continue
		}
		switch f := f.(type) {
		case *PdfObjectReference:
			crypter.decryptedObjNum[int(f.ObjectNumber)] = struct{}{}
		case *PdfIndirectObject:
			crypter.decryptedObjects[f] = true
			crypter.decryptedObjNum[int(f.ObjectNumber)] = struct{}{}
		}
	}
	parser.crypter = crypter
	common.Log.Trace("Crypter object %b", crypter)
	return true, nil
}

// Decrypt attempts to decrypt the PDF file with a specified password.  Also tries to
// decrypt with an empty password.  Returns true if successful, false otherwise.
// An error is returned when there is a problem with decrypting.
func (parser *PdfParser) Decrypt(password []byte) (bool, error) {
	// Also build the encryption/decryption key.
	if parser.crypter == nil {
		return false, errors.New("check encryption first")
	}

	authenticated, err := parser.crypter.authenticate(password)
	if err != nil {
		return false, err
	}

	if !authenticated {
		// TODO(dennwc): R6 handler will try it automatically, make R4 do the same
		authenticated, err = parser.crypter.authenticate([]byte(""))
	}

	return authenticated, err
}

// CheckAccessRights checks access rights and permissions for a specified password. If either user/owner password is
// specified, full rights are granted, otherwise the access rights are specified by the Permissions flag.
//
// The bool flag indicates that the user can access and view the file.
// The AccessPermissions shows what access the user has for editing etc.
// An error is returned if there was a problem performing the authentication.
func (parser *PdfParser) CheckAccessRights(password []byte) (bool, security.Permissions, error) {
	// Also build the encryption/decryption key.
	if parser.crypter == nil {
		// If the crypter is not set, the file is not encrypted and we can assume full access permissions.
		return true, security.PermOwner, nil
	}
	return parser.crypter.checkAccessRights(password)
}
