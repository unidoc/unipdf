/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package pdf

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/unidoc/unidoc/common"
)

// Regular Expressions for parsing and identifying object signatures.
var rePdfVersion = regexp.MustCompile(`%PDF-(\d\.\d)`)
var reEOF = regexp.MustCompile("%%EOF")
var reXrefTable = regexp.MustCompile(`\s*xref\s*`)
var reStartXref = regexp.MustCompile(`startx?ref\s*(\d+)`)
var reNumeric = regexp.MustCompile(`^[\+-.]*([0-9.]+)`)
var reExponential = regexp.MustCompile(`^[\+-.]*([0-9.]+)e[\+-.]*([0-9.]+)`)
var reReference = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+R`)
var reIndirectObject = regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)
var reXrefSubsection = regexp.MustCompile(`(\d+)\s+(\d+)\s*$`)
var reXrefEntry = regexp.MustCompile(`(\d+)\s+(\d+)\s+([nf])\s*$`)

type PdfParser struct {
	rs               io.ReadSeeker
	reader           *bufio.Reader
	xrefs            XrefTable
	objstms          ObjectStreams
	trailer          *PdfObjectDictionary
	ObjCache         ObjectCache
	crypter          *PdfCrypt
	repairsAttempted bool // Avoid multiple attempts for repair.
}

func isWhiteSpace(ch byte) bool {
	// Table 1 white-space characters (7.2.2 Character Set)
	// spaceCharacters := string([]byte{0x00, 0x09, 0x0A, 0x0C, 0x0D, 0x20})
	if (ch == 0x00) || (ch == 0x09) || (ch == 0x0A) || (ch == 0x0C) || (ch == 0x0D) || (ch == 0x20) {
		return true
	} else {
		return false
	}
}

func isDecimalDigit(c byte) bool {
	if c >= '0' && c <= '9' {
		return true
	} else {
		return false
	}
}

func isOctalDigit(c byte) bool {
	if c >= '0' && c <= '7' {
		return true
	} else {
		return false
	}
}

// Skip over any spaces.
func (this *PdfParser) skipSpaces() (int, error) {
	cnt := 0
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			return 0, err
		}
		if isWhiteSpace(bb[0]) {
			this.reader.ReadByte()
			cnt++
		} else {
			break
		}
	}

	return cnt, nil
}

// Skip over comments and spaces. Can handle multi-line comments.
func (this *PdfParser) skipComments() error {
	if _, err := this.skipSpaces(); err != nil {
		return err
	}

	isFirst := true
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			common.Log.Debug("Error %s", err.Error())
			return err
		}
		if isFirst && bb[0] != '%' {
			// Not a comment clearly.
			return nil
		} else {
			isFirst = false
		}
		if (bb[0] != '\r') && (bb[0] != '\n') {
			this.reader.ReadByte()
		} else {
			break
		}
	}

	// Call recursively to handle multiline comments.
	return this.skipComments()
}

// Read a comment starting with '%'.
func (this *PdfParser) readComment() (string, error) {
	commentText := ""

	_, err := this.skipSpaces()
	if err != nil {
		return commentText, err
	}

	isFirst := true
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			common.Log.Debug("Error %s", err.Error())
			return commentText, err
		}
		if isFirst && bb[0] != '%' {
			return commentText, errors.New("Comment should start with %")
		} else {
			isFirst = false
		}
		if (bb[0] != '\r') && (bb[0] != '\n') {
			b, _ := this.reader.ReadByte()
			commentText += string(b)
		} else {
			break
		}
	}
	return commentText, nil
}

// Read a single line of text from current position.
func (this *PdfParser) readTextLine() (string, error) {
	lineStr := ""
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			common.Log.Debug("Error %s", err.Error())
			return lineStr, err
		}
		if (bb[0] != '\r') && (bb[0] != '\n') {
			b, _ := this.reader.ReadByte()
			lineStr += string(b)
		} else {
			break
		}
	}
	return lineStr, nil
}

// Parse a name starting with '/'.
func (this *PdfParser) parseName() (PdfObjectName, error) {
	name := ""
	nameStarted := false
	for {
		bb, err := this.reader.Peek(1)
		if err == io.EOF {
			break // Can happen when loading from object stream.
		}
		if err != nil {
			return PdfObjectName(name), err
		}

		if !nameStarted {
			// Should always start with '/', otherwise not valid.
			if bb[0] == '/' {
				nameStarted = true
				this.reader.ReadByte()
			} else if bb[0] == '%' {
				this.readComment()
				this.skipSpaces()
			} else {
				common.Log.Debug("ERROR Name starting with %s (% x)", bb, bb)
				return PdfObjectName(name), fmt.Errorf("Invalid name: (%c)", bb[0])
			}
		} else {
			if isWhiteSpace(bb[0]) {
				break
			} else if (bb[0] == '/') || (bb[0] == '[') || (bb[0] == '(') || (bb[0] == ']') || (bb[0] == '<') || (bb[0] == '>') {
				break // Looks like start of next statement.
			} else if bb[0] == '#' {
				hexcode, err := this.reader.Peek(3)
				if err != nil {
					return PdfObjectName(name), err
				}
				this.reader.Discard(3)

				code, err := hex.DecodeString(string(hexcode[1:3]))
				if err != nil {
					return PdfObjectName(name), err
				}
				name += string(code)
			} else {
				b, _ := this.reader.ReadByte()
				name += string(b)
			}
		}
	}
	return PdfObjectName(name), nil
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
// Nontheless, we sometimes get numbers with exponential format, so
// we will support it in the reader (no confusion with other types, so
// no compromise).
func (this *PdfParser) parseNumber() (PdfObject, error) {
	isFloat := false
	allowSigns := true
	numStr := ""
	for {
		common.Log.Debug("Parsing number \"%s\"", numStr)
		bb, err := this.reader.Peek(1)
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
			b, _ := this.reader.ReadByte()
			numStr += string(b)
			allowSigns = false // Only allowed in beginning, and after e (exponential).
		} else if isDecimalDigit(bb[0]) {
			b, _ := this.reader.ReadByte()
			numStr += string(b)
		} else if bb[0] == '.' {
			b, _ := this.reader.ReadByte()
			numStr += string(b)
			isFloat = true
		} else if bb[0] == 'e' {
			// Exponential number format.
			b, _ := this.reader.ReadByte()
			numStr += string(b)
			isFloat = true
			allowSigns = true
		} else {
			break
		}
	}

	if isFloat {
		fVal, err := strconv.ParseFloat(numStr, 64)
		o := PdfObjectFloat(fVal)
		return &o, err
	} else {
		intVal, err := strconv.ParseInt(numStr, 10, 64)
		o := PdfObjectInteger(intVal)
		return &o, err
	}
}

// A string starts with '(' and ends with ')'.
func (this *PdfParser) parseString() (PdfObjectString, error) {
	this.reader.ReadByte()

	bytes := []byte{}
	count := 1
	for {
		bb, err := this.reader.Peek(1)
		if err != nil {
			return PdfObjectString(bytes), err
		}

		if bb[0] == '\\' { // Escape sequence.
			this.reader.ReadByte() // Skip the escape \ byte.
			b, err := this.reader.ReadByte()
			if err != nil {
				return PdfObjectString(bytes), err
			}

			// Octal '\ddd' number (base 8).
			if isOctalDigit(b) {
				bb, err := this.reader.Peek(2)
				if err != nil {
					return PdfObjectString(bytes), err
				}

				numeric := []byte{}
				numeric = append(numeric, b)
				for _, val := range bb {
					if isOctalDigit(val) {
						numeric = append(numeric, val)
					} else {
						break
					}
				}
				this.reader.Discard(len(numeric) - 1)

				common.Log.Debug("Numeric string \"%s\"", numeric)
				code, err := strconv.ParseUint(string(numeric), 8, 32)
				if err != nil {
					return PdfObjectString(bytes), err
				}
				bytes = append(bytes, byte(code))
				continue
			}

			switch b {
			case 'n':
				bytes = append(bytes, '\n')
			case 'r':
				bytes = append(bytes, '\r')
			case 't':
				bytes = append(bytes, '\t')
			case 'b':
				bytes = append(bytes, '\b')
			case 'f':
				bytes = append(bytes, '\f')
			case '(':
				bytes = append(bytes, '(')
			case ')':
				bytes = append(bytes, ')')
			case '\\':
				bytes = append(bytes, '\\')
			}

			continue
		} else if bb[0] == '(' {
			count++
		} else if bb[0] == ')' {
			count--
			if count == 0 {
				this.reader.ReadByte()
				break
			}
		}

		b, _ := this.reader.ReadByte()
		bytes = append(bytes, b)
	}

	return PdfObjectString(bytes), nil
}

// Starts with '<' ends with '>'.
// Currently not converting the hex codes to characters.
func (this *PdfParser) parseHexString() (PdfObjectString, error) {
	this.reader.ReadByte()

	hextable := []byte("0123456789abcdefABCDEF")

	tmp := []byte{}
	for {
		this.skipSpaces()

		bb, err := this.reader.Peek(1)
		if err != nil {
			return PdfObjectString(""), err
		}

		if bb[0] == '>' {
			this.reader.ReadByte()
			break
		}

		b, _ := this.reader.ReadByte()
		if bytes.IndexByte(hextable, b) >= 0 {
			tmp = append(tmp, b)
		}
	}

	if len(tmp)%2 == 1 {
		tmp = append(tmp, '0')
	}

	buf, _ := hex.DecodeString(string(tmp))
	return PdfObjectString(buf), nil
}

// Starts with '[' ends with ']'.  Can contain any kinds of direct objects.
func (this *PdfParser) parseArray() (PdfObjectArray, error) {
	arr := make(PdfObjectArray, 0)

	this.reader.ReadByte()

	for {
		this.skipSpaces()

		bb, err := this.reader.Peek(1)
		if err != nil {
			return arr, err
		}

		if bb[0] == ']' {
			this.reader.ReadByte()
			break
		}

		obj, err := this.parseObject()
		if err != nil {
			return arr, err
		}
		arr = append(arr, obj)
	}

	return arr, nil
}

// Parse bool object.
func (this *PdfParser) parseBool() (PdfObjectBool, error) {
	bb, err := this.reader.Peek(4)
	if err != nil {
		return PdfObjectBool(false), err
	}
	if (len(bb) >= 4) && (string(bb[:4]) == "true") {
		this.reader.Discard(4)
		return PdfObjectBool(true), nil
	}

	bb, err = this.reader.Peek(5)
	if err != nil {
		return PdfObjectBool(false), err
	}
	if (len(bb) >= 5) && (string(bb[:5]) == "false") {
		this.reader.Discard(5)
		return PdfObjectBool(false), nil
	}

	return PdfObjectBool(false), errors.New("Unexpected boolean string")
}

// Parse reference to an indirect object.
func parseReference(refStr string) (PdfObjectReference, error) {
	objref := PdfObjectReference{}

	result := reReference.FindStringSubmatch(string(refStr))
	if len(result) < 3 {
		common.Log.Debug("Error parsing reference")
		return objref, errors.New("Unable to parse reference")
	}

	objNum, _ := strconv.Atoi(result[1])
	genNum, _ := strconv.Atoi(result[2])
	objref.ObjectNumber = int64(objNum)
	objref.GenerationNumber = int64(genNum)

	return objref, nil
}

// Parse null object.
func (this *PdfParser) parseNull() (PdfObjectNull, error) {
	_, err := this.reader.Discard(4)
	return PdfObjectNull{}, err
}

// Detect the signature at the current file position and parse
// the corresponding object.
func (this *PdfParser) parseObject() (PdfObject, error) {
	common.Log.Debug("Read direct object")
	this.skipSpaces()
	for {
		bb, err := this.reader.Peek(2)
		if err != nil {
			return nil, err
		}

		common.Log.Debug("Peek string: %s", string(bb))
		// Determine type.
		if bb[0] == '/' {
			name, err := this.parseName()
			common.Log.Debug("->Name: '%s'", name)
			return &name, err
		} else if bb[0] == '(' {
			common.Log.Debug("->String!")
			str, err := this.parseString()
			return &str, err
		} else if bb[0] == '[' {
			common.Log.Debug("->Array!")
			arr, err := this.parseArray()
			return &arr, err
		} else if (bb[0] == '<') && (bb[1] == '<') {
			common.Log.Debug("->Dict!")
			dict, err := this.parseDict()
			return dict, err
		} else if bb[0] == '<' {
			common.Log.Debug("->Hex string!")
			str, err := this.parseHexString()
			return &str, err
		} else if bb[0] == '%' {
			this.readComment()
			this.skipSpaces()
		} else {
			common.Log.Debug("->Number or ref?")
			// Reference or number?
			// Let's peek farther to find out.
			bb, _ = this.reader.Peek(15)
			peekStr := string(bb)
			common.Log.Debug("Peek str: %s", peekStr)

			if (len(peekStr) > 3) && (peekStr[:4] == "null") {
				null, err := this.parseNull()
				return &null, err
			} else if (len(peekStr) > 4) && (peekStr[:5] == "false") {
				b, err := this.parseBool()
				return &b, err
			} else if (len(peekStr) > 3) && (peekStr[:4] == "true") {
				b, err := this.parseBool()
				return &b, err
			}

			// Match reference.
			result1 := reReference.FindStringSubmatch(string(peekStr))
			if len(result1) > 1 {
				bb, _ = this.reader.ReadBytes('R')
				common.Log.Debug("-> !Ref: '%s'", string(bb[:len(bb)]))
				ref, err := parseReference(string(bb))
				return &ref, err
			}

			result2 := reNumeric.FindStringSubmatch(string(peekStr))
			if len(result2) > 1 {
				// Number object.
				common.Log.Debug("-> Number!")
				num, err := this.parseNumber()
				return num, err
			}

			result2 = reExponential.FindStringSubmatch(string(peekStr))
			if len(result2) > 1 {
				// Number object (exponential)
				common.Log.Debug("-> Exponential Number!")
				common.Log.Debug("% s", result2)
				num, err := this.parseNumber()
				return num, err
			}

			common.Log.Debug("ERROR Unknown (peek \"%s\")", peekStr)
			return nil, errors.New("Object parsing error - unexpected pattern")
		}
	}

	return nil, errors.New("Object parsing error - unexpected pattern")
}

// Reads and parses a PDF dictionary object enclosed with '<<' and '>>'
func (this *PdfParser) parseDict() (*PdfObjectDictionary, error) {
	common.Log.Debug("Reading PDF Dict!")

	dict := make(PdfObjectDictionary)

	// Pass the '<<'
	c, _ := this.reader.ReadByte()
	if c != '<' {
		return nil, errors.New("Invalid dict")
	}
	c, _ = this.reader.ReadByte()
	if c != '<' {
		return nil, errors.New("Invalid dict")
	}

	for {
		this.skipSpaces()
		this.skipComments()

		bb, err := this.reader.Peek(2)
		if err != nil {
			return nil, err
		}

		common.Log.Debug("Dict peek: %s (% x)!", string(bb), string(bb))
		if (bb[0] == '>') && (bb[1] == '>') {
			common.Log.Debug("EOF dictionary")
			this.reader.ReadByte()
			this.reader.ReadByte()
			break
		}
		common.Log.Debug("Parse the name!")

		keyName, err := this.parseName()
		common.Log.Debug("Key: %s", keyName)
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
			this.skipSpaces()
			bb, _ := this.reader.Peek(1)
			if bb[0] == '/' {
				var nullObj PdfObjectNull
				dict[newKey] = &nullObj
				continue
			}
		}

		this.skipSpaces()

		val, err := this.parseObject()
		if err != nil {
			return nil, err
		}
		dict[keyName] = val

		common.Log.Debug("dict[%s] = %s", keyName, val.String())
	}

	return &dict, nil
}

// Parse the pdf version from the beginning of the file.
func (this *PdfParser) parsePdfVersion() (float64, error) {
	this.rs.Seek(0, os.SEEK_SET)
	var offset int64 = 20
	b := make([]byte, offset)
	this.rs.Read(b)

	result1 := rePdfVersion.FindStringSubmatch(string(b))
	if len(result1) < 2 {
		common.Log.Debug("Error: PDF Version not found!")
		return -1, errors.New("PDF version not found")
	}

	version, err := strconv.ParseFloat(result1[1], 64)
	if err != nil {
		return 0, err
	}

	//version, _ := strconv.Atoi(result1[1])
	common.Log.Debug("Pdf version %f", version)

	return version, nil
}

// Conventional xref table starting with 'xref'.
func (this *PdfParser) parseXrefTable() (*PdfObjectDictionary, error) {
	var trailer *PdfObjectDictionary

	txt, err := this.readTextLine()
	if err != nil {
		return nil, err
	}

	common.Log.Debug("xref first line: %s", txt)
	curObjNum := -1
	secObjects := 0
	insideSubsection := false
	for {
		this.skipSpaces()
		_, err := this.reader.Peek(1)
		if err != nil {
			return nil, err
		}

		txt, err = this.readTextLine()
		if err != nil {
			return nil, err
		}

		result1 := reXrefSubsection.FindStringSubmatch(txt)
		if len(result1) == 3 {
			// Match
			first, _ := strconv.Atoi(result1[1])
			second, _ := strconv.Atoi(result1[2])
			curObjNum = first
			secObjects = second
			insideSubsection = true
			common.Log.Debug("xref subsection: first object: %d objects: %d", curObjNum, secObjects)
			continue
		}
		result2 := reXrefEntry.FindStringSubmatch(txt)
		if len(result2) == 4 {
			if insideSubsection == false {
				common.Log.Debug("ERROR Xref invalid format!\n")
				return nil, errors.New("Xref invalid format")
			}

			first, _ := strconv.ParseInt(result2[1], 10, 64)
			gen, _ := strconv.Atoi(result2[2])
			third := result2[3]

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
				x, ok := this.xrefs[curObjNum]
				if !ok || gen > x.generation {
					obj := XrefObject{objectNumber: curObjNum,
						xtype:  XREF_TABLE_ENTRY,
						offset: first, generation: gen}
					this.xrefs[curObjNum] = obj
				}
			}

			curObjNum++
			continue
		}
		if (len(txt) > 6) && (txt[:7] == "trailer") {
			// Sometimes get "trailer << ...."
			// Need to rewind to end of trailer text.
			if len(txt) > 9 {
				offset := this.GetFileOffset()
				this.SetFileOffset(offset - int64(len(txt)) + 7)
			}

			this.skipSpaces()
			common.Log.Debug("Reading trailer dict!")
			common.Log.Debug("peek: \"%s\"", txt)
			trailer, err = this.parseDict()
			common.Log.Debug("EOF reading trailer dict!")
			if err != nil {
				common.Log.Debug("Error parsing trailer dict (%s)", err)
				return nil, err
			}
			break
		}

		if txt == "%%EOF" {
			common.Log.Debug("ERROR: end of file - trailer not found - error!")
			return nil, errors.New("End of file - trailer not found!")
		}

		common.Log.Debug("xref more : %s", txt)
	}
	common.Log.Debug("EOF parsing xref table!")

	return trailer, nil
}

// Load the cross references from an xref stream object (XRefStm).
// Also load the dictionary information (trailer dictionary).
func (this *PdfParser) parseXrefStream(xstm *PdfObjectInteger) (*PdfObjectDictionary, error) {
	if xstm != nil {
		common.Log.Debug("XRefStm xref table object at %d", xstm)
		this.rs.Seek(int64(*xstm), os.SEEK_SET)
		this.reader = bufio.NewReader(this.rs)
	}

	xrefObj, err := this.parseIndirectObject()
	if err != nil {
		common.Log.Debug("ERROR: Failed to read xref object")
		return nil, errors.New("Failed to read xref object")
	}

	common.Log.Debug("XRefStm object: %s", xrefObj)
	xs, ok := xrefObj.(*PdfObjectStream)
	if !ok {
		common.Log.Debug("ERROR: XRefStm pointing to non-stream object!")
		return nil, errors.New("XRefStm pointing to a non-stream object!")
	}

	trailerDict := xs.PdfObjectDictionary

	sizeObj, ok := (*(xs.PdfObjectDictionary))["Size"].(*PdfObjectInteger)
	if !ok {
		common.Log.Debug("ERROR: Missing size from xref stm")
		return nil, errors.New("Missing Size from xref stm")
	}

	wObj := (*(xs.PdfObjectDictionary))["W"]
	wArr, ok := wObj.(*PdfObjectArray)
	if !ok {
		return nil, errors.New("Invalid W in xref stream")
	}

	wLen := len(*wArr)
	if wLen != 3 {
		common.Log.Debug("ERROR: Unsupported xref stm (len(W) != 3 - %d)", wLen)
		return nil, errors.New("Unsupported xref stm len(W) != 3")
	}

	var b []int64
	for i := 0; i < 3; i++ {
		w, ok := (*wArr)[i].(PdfObject)
		if !ok {
			return nil, errors.New("Invalid W")
		}
		wVal, ok := w.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("Invalid w object type")
		}

		b = append(b, int64(*wVal))
	}

	ds, err := this.decodeStream(xs)
	if err != nil {
		common.Log.Debug("ERROR: Unable to decode stream")
		return nil, err
	}

	s0 := int(b[0])
	s1 := int(b[0] + b[1])
	s2 := int(b[0] + b[1] + b[2])
	deltab := int(b[0] + b[1] + b[2])

	// Calculate expected entries.
	entries := len(ds) / deltab

	// Get the object indices.

	objCount := 0
	indexObj := (*(xs.PdfObjectDictionary))["Index"]
	// Table 17 (7.5.8.2 Cross-Reference Stream Dictionary)
	// (Optional) An array containing a pair of integers for each
	// subsection in this section. The first integer shall be the first
	// object number in the subsection; the second integer shall be the
	// number of entries in the subsection.
	// The array shall be sorted in ascending order by object number.
	// Subsections cannot overlap; an object number may have at most
	// one entry in a section.
	// Default value: [0 Size].
	indexList := []int{}
	if indexObj != nil {
		common.Log.Debug("Index: %b", indexObj)
		indices, ok := indexObj.(*PdfObjectArray)
		if !ok {
			common.Log.Debug("Invalid Index object (should be an array)")
			return nil, errors.New("Invalid Index object")
		}

		// Expect indLen to be a multiple of 2.
		if len(*indices)%2 != 0 {
			common.Log.Debug("WARNING Failure loading xref stm index not multiple of 2.")
			return nil, err
		}

		objCount = 0
		for i := 0; i < len(*indices); i += 2 {
			// add the indices to the list..
			startIdx := int(*(*indices)[i].(*PdfObjectInteger))
			numObjs := int(*(*indices)[i+1].(*PdfObjectInteger))
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
		common.Log.Debug("BAD file: allowing compatibility (append one object to xref stm)")
		indexList = append(indexList, objCount)
		objCount++
	}

	if entries != len(indexList) {
		// If mismatch -> error (already allowing mismatch of 1 if Index not specified).
		common.Log.Debug("ERROR: xref stm: num entries != len(indices) (%d != %d)", entries, len(indexList))
		return nil, errors.New("Xref stm num entries != len(indices)")
	}

	common.Log.Debug("Objects count %d", objCount)
	common.Log.Debug("Indices: % d", indexList)

	// Convert byte array to a larger integer, little-endian.
	convertBytes := func(v []byte) int64 {
		var tmp int64 = 0
		for i := 0; i < len(v); i++ {
			tmp += int64(v[i]) * (1 << uint(8*(len(v)-i-1)))

		}
		return tmp
	}

	common.Log.Debug("Decoded stream length: %d", len(ds))
	objIndex := 0
	for i := 0; i < len(ds); i += deltab {
		p1 := ds[i : i+s0]
		p2 := ds[i+s0 : i+s1]
		p3 := ds[i+s1 : i+s2]
		ftype := convertBytes(p1)
		n2 := convertBytes(p2)
		n3 := convertBytes(p3)

		if b[0] == 0 {
			// If first entry in W is 0, then default to to type 1.
			// (uncompressed object via offset).
			ftype = 1
		}

		objNum := indexList[objIndex]
		objIndex++

		common.Log.Debug("%d. p1: % x", objNum, p1)
		common.Log.Debug("%d. p2: % x", objNum, p2)
		common.Log.Debug("%d. p3: % x", objNum, p3)

		common.Log.Debug("%d. xref: %d %d %d", objNum, ftype, n2, n3)
		if ftype == 0 {
			common.Log.Debug("- Free object - can probably ignore")
		} else if ftype == 1 {
			common.Log.Debug("- In use - uncompressed via offset %b", p2)
			// Object type 1: Objects that are in use but are not
			// compressed, i.e. defined by an offset (normal entry)
			if xr, ok := this.xrefs[objNum]; !ok || int(n3) > xr.generation {
				// Only overload if not already loaded!
				// or has a newer generation number. (should not happen)
				obj := XrefObject{objectNumber: objNum,
					xtype: XREF_TABLE_ENTRY, offset: n2, generation: int(n3)}
				this.xrefs[objNum] = obj
			}
		} else if ftype == 2 {
			// Object type 2: Compressed object.
			common.Log.Debug("- In use - compressed object")
			if _, ok := this.xrefs[objNum]; !ok {
				obj := XrefObject{objectNumber: objNum,
					xtype: XREF_OBJECT_STREAM, osObjNumber: int(n2), osObjIndex: int(n3)}
				this.xrefs[objNum] = obj
				common.Log.Debug("entry: %s", this.xrefs[objNum])
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

	return trailerDict, nil
}

// Parse xref table at the current file position.  Can either be a
// standard xref table, or an xref stream.
func (this *PdfParser) parseXref() (*PdfObjectDictionary, error) {
	var err error
	var trailerDict *PdfObjectDictionary

	// Points to xref table or xref stream object?
	bb, _ := this.reader.Peek(20)
	if reIndirectObject.MatchString(string(bb)) {
		common.Log.Debug("xref points to an object.  Probably xref object")
		common.Log.Debug("starting with \"%s\"", string(bb))
		trailerDict, err = this.parseXrefStream(nil)
		if err != nil {
			return nil, err
		}
	} else if reXrefTable.MatchString(string(bb)) {
		common.Log.Debug("Standard xref section table!")
		var err error
		trailerDict, err = this.parseXrefTable()
		if err != nil {
			return nil, err
		}
	} else {
		common.Log.Debug("ERROR: Invalid xref.... starting with \"%s\"", string(bb))
		return nil, errors.New("Invalid xref format")
	}

	return trailerDict, err
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
// The earlier xrefs have higher precedance.  If objects already
// loaded will ignore older versions.
//
func (this *PdfParser) loadXrefs() (*PdfObjectDictionary, error) {
	this.xrefs = make(XrefTable)
	this.objstms = make(ObjectStreams)

	// Look for EOF marker and seek to its beginning.
	// Define an offset position from the end of the file.
	var offset int64 = 1000
	// Get the file size.
	fSize, err := this.rs.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}
	common.Log.Debug("fsize: %d", fSize)
	if fSize <= offset {
		offset = fSize
	}
	_, err = this.rs.Seek(-offset, os.SEEK_END)
	if err != nil {
		return nil, err
	}
	b1 := make([]byte, offset)
	this.rs.Read(b1)
	common.Log.Debug("Looking for EOF marker: \"%s\"", string(b1))
	ind := reEOF.FindAllStringIndex(string(b1), -1)
	if ind == nil {
		common.Log.Debug("Error: EOF marker not found!")
		return nil, errors.New("EOF marker not found")
	}
	lastInd := ind[len(ind)-1]
	common.Log.Debug("Ind: % d", ind)
	this.rs.Seek(-offset+int64(lastInd[0]), os.SEEK_END)

	// Look for startxref and get the xref offset.
	offset = 64
	this.rs.Seek(-offset, os.SEEK_CUR)
	b2 := make([]byte, offset)
	this.rs.Read(b2)

	result := reStartXref.FindStringSubmatch(string(b2))
	if len(result) < 2 {
		common.Log.Debug("Error: startxref not found!")
		return nil, errors.New("Startxref not found")
	}
	if len(result) > 2 {
		// GH: Take the last one?  Make a test case.
		common.Log.Debug("ERROR: Multiple startxref (%s)!", b2)
		return nil, errors.New("Multiple startxref entries?")
	}
	offsetXref, _ := strconv.ParseInt(result[1], 10, 64)
	common.Log.Debug("startxref at %d", offsetXref)

	if offsetXref > fSize {
		common.Log.Debug("ERROR: Xref offset outside of file")
		common.Log.Debug("Attempting repair")
		offsetXref, err = this.repairLocateXref()
		if err != nil {
			common.Log.Debug("ERROR: Repair attempt failed (%s)")
			return nil, err
		}
	}
	// Read the xref.
	this.rs.Seek(int64(offsetXref), os.SEEK_SET)
	this.reader = bufio.NewReader(this.rs)

	trailerDict, err := this.parseXref()
	if err != nil {
		return nil, err
	}

	// Check the XrefStm object also from the trailer.
	xx, present := (*trailerDict)["XRefStm"]
	if present {
		xo, ok := xx.(*PdfObjectInteger)
		if !ok {
			return nil, errors.New("XRefStm != int")
		}
		_, err = this.parseXrefStream(xo)
		if err != nil {
			return nil, err
		}
	}

	// Load old objects also.  Only if not already specified.
	prevList := []int64{}
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
	xx, present = (*trailerDict)["Prev"]
	for present {
		off := *(xx.(*PdfObjectInteger))
		common.Log.Debug("Another Prev xref table object at %d", off)

		// Can be either regular table, or an xref object...
		this.rs.Seek(int64(off), os.SEEK_SET)
		this.reader = bufio.NewReader(this.rs)

		ptrailerDict, err := this.parseXref()
		if err != nil {
			common.Log.Debug("ERROR: Failed loading another (Prev) trailer")
			return nil, err
		}

		xx, present = (*ptrailerDict)["Prev"]
		if present {
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

// Parse an indirect object from the input stream.
// Can also be an object stream.
func (this *PdfParser) parseIndirectObject() (PdfObject, error) {
	indirect := PdfIndirectObject{}

	common.Log.Debug("-Read indirect obj")
	bb, err := this.reader.Peek(20)
	if err != nil {
		common.Log.Debug("ERROR: Fail to read indirect obj")
		return &indirect, err
	}
	common.Log.Debug("(indirect obj peek \"%s\"", string(bb))

	indices := reIndirectObject.FindStringSubmatchIndex(string(bb))
	if len(indices) < 6 {
		common.Log.Debug("ERROR: Unable to find object signature (%s)", string(bb))
		return &indirect, errors.New("Unable to detect indirect object signature")
	}
	this.reader.Discard(indices[0]) // Take care of any small offset.
	common.Log.Debug("Offsets % d", indices)

	// Read the object header.
	hlen := indices[1] - indices[0]
	hb := make([]byte, hlen)
	_, err = this.ReadAtLeast(hb, hlen)
	if err != nil {
		common.Log.Debug("ERROR: unable to read - %s", err)
		return nil, err
	}
	common.Log.Debug("textline: %s", hb)

	result := reIndirectObject.FindStringSubmatch(string(hb))
	if len(result) < 3 {
		common.Log.Debug("ERROR: Unable to find object signature (%s)", string(hb))
		return &indirect, errors.New("Unable to detect indirect object signature")
	}

	on, _ := strconv.Atoi(result[1])
	gn, _ := strconv.Atoi(result[2])
	indirect.ObjectNumber = int64(on)
	indirect.GenerationNumber = int64(gn)

	for {
		bb, err := this.reader.Peek(2)
		if err != nil {
			return &indirect, err
		}

		if isWhiteSpace(bb[0]) {
			this.skipSpaces()
		} else if (bb[0] == '<') && (bb[1] == '<') {
			indirect.PdfObject, err = this.parseDict()
			if err != nil {
				return &indirect, err
			}
		} else if (bb[0] == '/') || (bb[0] == '(') || (bb[0] == '[') || (bb[0] == '<') {
			indirect.PdfObject, err = this.parseObject()
			if err != nil {
				return &indirect, err
			}
		} else {
			if bb[0] == 'e' {
				lineStr, err := this.readTextLine()
				if err != nil {
					return nil, err
				}
				if len(lineStr) >= 6 && lineStr[0:6] == "endobj" {
					break
				}
			} else if bb[0] == 's' {
				bb, _ = this.reader.Peek(10)
				if string(bb[:6]) == "stream" {
					discardBytes := 6
					if len(bb) > 6 {
						if bb[discardBytes] == '\r' {
							discardBytes++
							if bb[discardBytes] == '\n' {
								discardBytes++
							}
						} else if bb[discardBytes] == '\n' {
							discardBytes++
						}
					}

					this.reader.Discard(discardBytes)

					dict, isDict := indirect.PdfObject.(*PdfObjectDictionary)
					if !isDict {
						return nil, errors.New("Stream object missing dictionary")
					}
					common.Log.Debug("Stream dict %s", dict)

					slo, err := this.Trace((*dict)["Length"])
					if err != nil {
						return nil, err
					}

					common.Log.Debug("Stream length? %s", slo)

					pstreamLength, ok := slo.(*PdfObjectInteger)
					if !ok {
						return nil, errors.New("Stream length needs to be an integer")
					}
					streamLength := *pstreamLength
					if streamLength < 0 {
						return nil, errors.New("Stream needs to be longer than 0")
					}

					stream := make([]byte, streamLength)
					_, err = this.ReadAtLeast(stream, int(streamLength))
					if err != nil {
						common.Log.Debug("ERROR stream (%d): %X", len(stream), stream)
						return nil, err
					}

					streamobj := PdfObjectStream{}
					streamobj.Stream = stream
					streamobj.PdfObjectDictionary = indirect.PdfObject.(*PdfObjectDictionary)
					streamobj.ObjectNumber = indirect.ObjectNumber
					streamobj.GenerationNumber = indirect.GenerationNumber

					this.skipSpaces()
					this.reader.Discard(9) // endstream
					this.skipSpaces()
					return &streamobj, nil
				}
			}

			indirect.PdfObject, err = this.parseObject()
			return &indirect, err
		}
	}
	return &indirect, nil
}

// Creates a new parser for a PDF file via ReadSeeker.  Loads the
// cross reference stream and trailer.
func NewParser(rs io.ReadSeeker) (*PdfParser, error) {
	parser := &PdfParser{}

	parser.rs = rs
	parser.ObjCache = make(ObjectCache)

	// Start by reading xrefs from bottom
	trailer, err := parser.loadXrefs()
	if err != nil {
		common.Log.Debug("ERROR: Failed to load xref table! %s", err)
		// Try to rebuild entire xref table?
		return nil, err
	}

	common.Log.Debug("Trailer: %s", trailer)

	if len(parser.xrefs) == 0 {
		return nil, fmt.Errorf("Empty XREF table. Invalid.")
	}

	// printXrefTable(parser.xrefs)

	_, err = parser.parsePdfVersion()
	if err != nil {
		return nil, fmt.Errorf("Unable to parse version (%s)", err)
	}

	parser.trailer = trailer

	return parser, nil
}

// Check if the document is encrypted.  First time when called, will
// check if the Encrypt dictionary is accessible through the trailer
// dictionary.
// If encrypted, prepares a crypt datastructure which can be used to
// authenticate and decrypt the document.
func (this *PdfParser) IsEncrypted() (bool, error) {
	if this.crypter != nil {
		return true, nil
	}

	if this.trailer != nil {
		common.Log.Debug("Checking encryption dictionary!")
		encDictRef, isEncrypted := (*(this.trailer))["Encrypt"].(*PdfObjectReference)
		if isEncrypted {
			common.Log.Debug("Is encrypted!")
			common.Log.Debug("0: Look up ref %q", encDictRef)
			encObj, err := this.LookupByReference(*encDictRef)
			common.Log.Debug("1: %q", encObj)
			if err != nil {
				return false, err
			}
			encDict, ok := encObj.(*PdfIndirectObject).PdfObject.(*PdfObjectDictionary)
			common.Log.Debug("2: %q", encDict)
			if !ok {
				return false, errors.New("Trailer Encrypt object non dictionary")
			}
			crypter, err := PdfCryptMakeNew(encDict, this.trailer)
			if err != nil {
				return false, err
			}

			this.crypter = &crypter
			common.Log.Debug("Crypter object %b", crypter)
			return true, nil
		}
	}
	return false, nil
}

// Decrypt the PDF file with a specified password.  Also tries to
// decrypt with an empty password.  Returns true if successful,
// false otherwise.
func (this *PdfParser) Decrypt(password []byte) (bool, error) {
	// Also build the encryption/decryption key.
	if this.crypter == nil {
		return false, errors.New("Check encryption first")
	}

	authenticated, err := this.crypter.authenticate(password)
	if err != nil {
		return false, err
	}

	if !authenticated {
		authenticated, err = this.crypter.authenticate([]byte(""))
	}

	return authenticated, err
}
