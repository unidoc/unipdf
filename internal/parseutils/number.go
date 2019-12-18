package parseutils

import (
	"bufio"
	"bytes"
	"io"
	"strconv"

	"github.com/unidoc/unipdf/v3/common"
)

// ParseNumber parses a numeric objects from a buffered stream.
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
func ParseNumber(bufr *bufio.Reader) (interface{}, error) {
	isFloat := false
	allowSigns := true
	var r bytes.Buffer
	for {
		if common.Log.IsLogLevel(common.LogLevelTrace) {
			common.Log.Trace("Parsing number \"%s\"", r.String())
		}
		bb, err := bufr.Peek(1)
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
			b, _ := bufr.ReadByte()
			r.WriteByte(b)
			allowSigns = false // Only allowed in beginning, and after e (exponential).
		} else if IsDecimalDigit(bb[0]) {
			b, _ := bufr.ReadByte()
			r.WriteByte(b)
		} else if bb[0] == '.' {
			b, _ := bufr.ReadByte()
			r.WriteByte(b)
			isFloat = true
		} else if bb[0] == 'e' || bb[0] == 'E' {
			// Exponential number format.
			b, _ := bufr.ReadByte()
			r.WriteByte(b)
			isFloat = true
			allowSigns = true
		} else {
			break
		}
	}

	var o interface{}
	if isFloat {
		fVal, err := strconv.ParseFloat(r.String(), 64)
		if err != nil {
			common.Log.Debug("Error parsing number %v err=%v. Using 0.0. Output may be incorrect", r.String(), err)
			fVal = 0.0
			err = nil
		}
		o = fVal
	} else {
		intVal, err := strconv.ParseInt(r.String(), 10, 64)
		if err != nil {
			common.Log.Debug("Error parsing number %v err=%v. Using 0. Output may be incorrect", r.String(), err)
			intVal = 0
			err = nil
		}
		o = intVal
	}

	return o, nil
}

// IsDecimalDigit checks if the character is a part of a decimal number string.
func IsDecimalDigit(c byte) bool {
	if c >= '0' && c <= '9' {
		return true
	}

	return false
}
