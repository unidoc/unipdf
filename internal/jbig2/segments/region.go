/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"fmt"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
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
	common.Log.Trace("[REGION][PARSE-HEADER] Begin")
	defer func() {
		common.Log.Trace("[REGION][PARSE-HEADER] Finished")
	}()

	temp, err := r.r.ReadBits(32)
	if err != nil {
		return err
	}
	r.BitmapWidth = uint32(temp & math.MaxUint32)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return err
	}
	r.BitmapHeight = uint32(temp & math.MaxUint32)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return err
	}
	r.XLocation = uint32(temp & math.MaxUint32)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return err
	}
	r.YLocation = uint32(temp & math.MaxUint32)

	// Bit 3-7
	r.r.ReadBits(5) // dirty read

	// Bit 0-2
	if err = r.readCombinationOperator(); err != nil {
		return err
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
