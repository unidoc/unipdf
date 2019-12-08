/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// RegionSegment is the model representing base jbig2 segment region - see 7.4.1.
type RegionSegment struct {
	r reader.StreamReader
	// Region segment bitmap width, 7.4.1.1
	BitmapWidth uint32
	// Region segment bitmap height, 7.4.1.2
	BitmapHeight uint32
	// Region segment bitmap X location, 7.4.1.3
	XLocation uint32
	// Region segment bitmap Y location, 7.4.1.4
	YLocation uint32
	// Region segment flags, 7.4.1.5
	CombinaionOperator bitmap.CombinationOperator
}

// NewRegionSegment creates new Region segment model.
func NewRegionSegment(r reader.StreamReader) *RegionSegment {
	return &RegionSegment{r: r}
}

// compile time check for the SegmentEncoder interface.
var _ SegmentEncoder = &RegionSegment{}

// Encode implements the SegmentEncoder interface.
func (r *RegionSegment) Encode(w writer.BinaryWriter) (n int, err error) {
	const processName = "RegionSegment.Encode"
	// the region segment encodes as follows:
	// segment bitmap width 		- big endian uint32
	// segment bitmap height 		- big endian uint32
	// segment bitmap x location 	- big endian uint32
	// segment bitmap y location	- big endian uint32
	// segment flags 				- 1 byte - 5 empty bits + 3 combination operator

	// prepare 4 bytes temporary slice.
	temp := make([]byte, 4)

	// encode region segment width
	binary.BigEndian.PutUint32(temp, r.BitmapWidth)
	n, err = w.Write(temp)
	if err != nil {
		return 0, errors.Wrap(err, processName, "Width")
	}

	// encode region segment height
	binary.BigEndian.PutUint32(temp, r.BitmapHeight)
	var tempCount int
	tempCount, err = w.Write(temp)
	if err != nil {
		return 0, errors.Wrap(err, processName, "Height")
	}
	n += tempCount

	// encode region segment x location
	binary.BigEndian.PutUint32(temp, r.XLocation)
	tempCount, err = w.Write(temp)
	if err != nil {
		return 0, errors.Wrap(err, processName, "XLocation")
	}
	n += tempCount

	// encode region segment y location
	binary.BigEndian.PutUint32(temp, r.YLocation)
	tempCount, err = w.Write(temp)
	if err != nil {
		return 0, errors.Wrap(err, processName, "YLocation")
	}
	n += tempCount

	// the region segment flags is composed of:
	// 5 zero bits + 3 bits for the external combination operator
	if err = w.WriteByte(byte(r.CombinaionOperator) & 0x07); err != nil {
		return 0, errors.Wrap(err, processName, "combination operator")
	}
	n++
	return n, nil
}

// Size returns the bytewise size of the region segment.
func (r *RegionSegment) Size() int {
	// width + height + xlocation + ylocation + flags = 17
	// 4 + 4 + 4 + 4 + 1 = 17
	return 17
}

// String implements the Stringer interface.
func (r *RegionSegment) String() string {
	sb := &strings.Builder{}

	sb.WriteString("\t[REGION SEGMENT]\n")
	sb.WriteString(fmt.Sprintf("\t\t- Bitmap (width, height) [%dx%d]\n", r.BitmapWidth, r.BitmapHeight))
	sb.WriteString(fmt.Sprintf("\t\t- Location (x,y): [%d,%d]\n", r.XLocation, r.YLocation))
	sb.WriteString(fmt.Sprintf("\t\t- CombinationOperator: %s", r.CombinaionOperator))
	return sb.String()
}

// parseHeader parses the RegionSegment header.
func (r *RegionSegment) parseHeader() error {
	const processName = "parseHeader"
	common.Log.Trace("[REGION][PARSE-HEADER] Begin")
	defer func() {
		common.Log.Trace("[REGION][PARSE-HEADER] Finished")
	}()

	temp, err := r.r.ReadBits(32)
	if err != nil {
		return errors.Wrap(err, processName, "width")
	}
	r.BitmapWidth = uint32(temp & math.MaxUint32)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return errors.Wrap(err, processName, "height")
	}
	r.BitmapHeight = uint32(temp & math.MaxUint32)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return errors.Wrap(err, processName, "x location")
	}
	r.XLocation = uint32(temp & math.MaxUint32)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return errors.Wrap(err, processName, "y location")
	}
	r.YLocation = uint32(temp & math.MaxUint32)

	// Bit 3-7
	if _, err = r.r.ReadBits(5); err != nil {
		return errors.Wrap(err, processName, "diry read")
	}

	// Bit 0-2
	if err = r.readCombinationOperator(); err != nil {
		return errors.Wrap(err, processName, "combination operator")
	}
	return nil
}

func (r *RegionSegment) readCombinationOperator() error {
	temp, err := r.r.ReadBits(3)
	if err != nil {
		return err
	}

	r.CombinaionOperator = bitmap.CombinationOperator(temp & 0xF)
	return nil
}
