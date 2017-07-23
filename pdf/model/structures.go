/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

// Common basic data structures: PdfRectangle, PdfDate, etc.
// These kinds of data structures can be copied, do not need a unique copy of each object.

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	. "github.com/unidoc/unidoc/pdf/core"
)

// Definition of a rectangle.
type PdfRectangle struct {
	Llx float64 // Lower left corner (ll).
	Lly float64
	Urx float64 // Upper right corner (ur).
	Ury float64
}

// Create a PDF rectangle object based on an input array of 4 integers.
// Defining the lower left (LL) and upper right (UR) corners with
// floating point numbers.
func NewPdfRectangle(arr PdfObjectArray) (*PdfRectangle, error) {
	rect := PdfRectangle{}
	if len(arr) != 4 {
		return nil, errors.New("Invalid rectangle array, len != 4")
	}

	var err error
	rect.Llx, err = getNumberAsFloat(arr[0])
	if err != nil {
		return nil, err
	}

	rect.Lly, err = getNumberAsFloat(arr[1])
	if err != nil {
		return nil, err
	}

	rect.Urx, err = getNumberAsFloat(arr[2])
	if err != nil {
		return nil, err
	}

	rect.Ury, err = getNumberAsFloat(arr[3])
	if err != nil {
		return nil, err
	}

	return &rect, nil
}

// Convert to a PDF object.
func (rect *PdfRectangle) ToPdfObject() PdfObject {
	arr := PdfObjectArray{}
	arr = append(arr, MakeFloat(rect.Llx))
	arr = append(arr, MakeFloat(rect.Lly))
	arr = append(arr, MakeFloat(rect.Urx))
	arr = append(arr, MakeFloat(rect.Ury))
	return &arr
}

// A date is a PDF string of the form:
// (D:YYYYMMDDHHmmSSOHH'mm)
type PdfDate struct {
	year          int64 // YYYY
	month         int64 // MM (01-12)
	day           int64 // DD (01-31)
	hour          int64 // HH (00-23)
	minute        int64 // mm (00-59)
	second        int64 // SS (00-59)
	utOffsetSign  byte  // O ('+' / '-' / 'Z')
	utOffsetHours int64 // HH' (00-23 followed by ')
	utOffsetMins  int64 // mm (00-59)
}

var reDate = regexp.MustCompile(`\s*D\s*:\s*(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})([+-Z])?(\d{2})?'?(\d{2})?`)

// Make a new PdfDate object from a PDF date string (see 7.9.4 Dates).
// format: "D: YYYYMMDDHHmmSSOHH'mm"
func NewPdfDate(dateStr string) (PdfDate, error) {
	d := PdfDate{}

	matches := reDate.FindAllStringSubmatch(dateStr, 1)
	if len(matches) < 1 {
		return d, fmt.Errorf("Invalid date string (%s)", dateStr)
	}
	if len(matches[0]) != 10 {
		return d, errors.New("Invalid regexp group match length != 10")
	}

	// No need to handle err from ParseInt, as pre-validated via regexp.
	d.year, _ = strconv.ParseInt(matches[0][1], 10, 32)
	d.month, _ = strconv.ParseInt(matches[0][2], 10, 32)
	d.day, _ = strconv.ParseInt(matches[0][3], 10, 32)
	d.hour, _ = strconv.ParseInt(matches[0][4], 10, 32)
	d.minute, _ = strconv.ParseInt(matches[0][5], 10, 32)
	d.second, _ = strconv.ParseInt(matches[0][6], 10, 32)
	// Some poor implementations do not include the offset.
	if len(matches[0][7]) > 0 {
		d.utOffsetSign = matches[0][7][0]
	} else {
		d.utOffsetSign = '+'
	}
	if len(matches[0][8]) > 0 {
		d.utOffsetHours, _ = strconv.ParseInt(matches[0][8], 10, 32)
	} else {
		d.utOffsetHours = 0
	}
	if len(matches[0][9]) > 0 {
		d.utOffsetMins, _ = strconv.ParseInt(matches[0][9], 10, 32)
	} else {
		d.utOffsetMins = 0
	}

	return d, nil
}

// Convert to a PDF string object.
func (date *PdfDate) ToPdfObject() PdfObject {
	str := fmt.Sprintf("D:%.4d%.2d%.2d%.2d%.2d%.2d%c%.2d'%.2d'",
		date.year, date.month, date.day, date.hour, date.minute, date.second,
		date.utOffsetSign, date.utOffsetHours, date.utOffsetMins)
	pdfStr := PdfObjectString(str)
	return &pdfStr
}
