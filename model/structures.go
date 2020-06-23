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
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/unidoc/unipdf/v3/core"
)

// PdfRectangle is a definition of a rectangle.
type PdfRectangle struct {
	Llx float64 // Lower left corner (ll).
	Lly float64
	Urx float64 // Upper right corner (ur).
	Ury float64
}

// NewPdfRectangle creates a PDF rectangle object based on an input array of 4 integers.
// Defining the lower left (LL) and upper right (UR) corners with
// floating point numbers.
func NewPdfRectangle(arr core.PdfObjectArray) (*PdfRectangle, error) {
	rect := PdfRectangle{}
	if arr.Len() != 4 {
		return nil, errors.New("invalid rectangle array, len != 4")
	}

	var err error
	rect.Llx, err = core.GetNumberAsFloat(arr.Get(0))
	if err != nil {
		return nil, err
	}

	rect.Lly, err = core.GetNumberAsFloat(arr.Get(1))
	if err != nil {
		return nil, err
	}

	rect.Urx, err = core.GetNumberAsFloat(arr.Get(2))
	if err != nil {
		return nil, err
	}

	rect.Ury, err = core.GetNumberAsFloat(arr.Get(3))
	if err != nil {
		return nil, err
	}

	return &rect, nil
}

// Height returns the height of `rect`.
func (rect *PdfRectangle) Height() float64 {
	return math.Abs(rect.Ury - rect.Lly)
}

// Width returns the width of `rect`.
func (rect *PdfRectangle) Width() float64 {
	return math.Abs(rect.Urx - rect.Llx)
}

// ToPdfObject converts rectangle to a PDF object.
func (rect *PdfRectangle) ToPdfObject() core.PdfObject {
	return core.MakeArray(
		core.MakeFloat(rect.Llx),
		core.MakeFloat(rect.Lly),
		core.MakeFloat(rect.Urx),
		core.MakeFloat(rect.Ury),
	)
}

// PdfDate represents a date, which is a PDF string of the form:
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

// ToGoTime returns the date in time.Time format.
func (d PdfDate) ToGoTime() time.Time {
	utcOffset := int(d.utOffsetHours*60*60 + d.utOffsetMins*60)
	switch d.utOffsetSign {
	case '-':
		utcOffset = -utcOffset
	case 'Z':
		utcOffset = 0
	}
	tzName := fmt.Sprintf("UTC%c%.2d%.2d", d.utOffsetSign, d.utOffsetHours, d.utOffsetMins)
	tz := time.FixedZone(tzName, utcOffset)

	return time.Date(int(d.year), time.Month(d.month), int(d.day), int(d.hour), int(d.minute), int(d.second), 0, tz)
}

var reDate = regexp.MustCompile(`\s*D\s*:\s*(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})([+-Z])?(\d{2})?'?(\d{2})?`)

// NewPdfDate returns a new PdfDate object from a PDF date string (see 7.9.4 Dates).
// format: "D: YYYYMMDDHHmmSSOHH'mm"
func NewPdfDate(dateStr string) (PdfDate, error) {
	d := PdfDate{}

	matches := reDate.FindAllStringSubmatch(dateStr, 1)
	if len(matches) < 1 {
		return d, fmt.Errorf("invalid date string (%s)", dateStr)
	}
	if len(matches[0]) != 10 {
		return d, errors.New("invalid regexp group match length != 10")
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

// NewPdfDateFromTime will create a PdfDate based on the given time
func NewPdfDateFromTime(timeObj time.Time) (PdfDate, error) {
	timezone := timeObj.Format("-07:00")
	utOffsetHours, _ := strconv.ParseInt(timezone[1:3], 10, 32)
	utOffsetMins, _ := strconv.ParseInt(timezone[4:6], 10, 32)

	return PdfDate{
		year:          int64(timeObj.Year()),
		month:         int64(timeObj.Month()),
		day:           int64(timeObj.Day()),
		hour:          int64(timeObj.Hour()),
		minute:        int64(timeObj.Minute()),
		second:        int64(timeObj.Second()),
		utOffsetSign:  timezone[0],
		utOffsetHours: utOffsetHours,
		utOffsetMins:  utOffsetMins,
	}, nil
}

// ToPdfObject converts date to a PDF string object.
func (d *PdfDate) ToPdfObject() core.PdfObject {
	str := fmt.Sprintf("D:%.4d%.2d%.2d%.2d%.2d%.2d%c%.2d'%.2d'",
		d.year, d.month, d.day, d.hour, d.minute, d.second,
		d.utOffsetSign, d.utOffsetHours, d.utOffsetMins)
	return core.MakeString(str)
}
