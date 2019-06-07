/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"fmt"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
)

// RegionSegment is the model representing base jbig2 segment region - see 7.4.1.
type RegionSegment struct {
	r reader.StreamReader

	// Region segment bitmap width, 7.4.1.1
	BitmapWidth int

	// Region segment bitmap height, 7.4.1.2
	BitmapHeight int

	// Region segment bitmap X location, 7.4.1.3
	XLocation int

	// Region segment bitmap Y location, 7.4.1.4
	YLocation int

	// Region segment flags, 7.4.1.5
	CombinaionOperator bitmap.CombinationOperator
}

// NewRegionSegment creates new Region segment model.
func NewRegionSegment(r reader.StreamReader) *RegionSegment {
	rs := &RegionSegment{r: r}
	return rs
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
	common.Log.Debug("[REGION][PARSE-HEADER] Begin")
	defer func() {
		common.Log.Debug("[REGION][PARSE-HEADER] Finished")
	}()

	temp, err := r.r.ReadBits(32)
	if err != nil {
		return err
	}
	r.BitmapWidth = int(temp & 0xffffffff)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return err
	}
	r.BitmapHeight = int(temp & 0xffffffff)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return err
	}
	r.XLocation = int(temp & 0xffffffff)

	temp, err = r.r.ReadBits(32)
	if err != nil {
		return err
	}
	r.YLocation = int(temp & 0xffffffff)

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
