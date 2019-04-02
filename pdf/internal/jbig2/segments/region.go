package segments

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// RegionSegment is the segment that takes
type RegionSegment struct {
	r reader.StreamReader

	/** Region segment bitmap width, 7.4.1.1 */
	BitmapWidth int

	/** Region segment bitmap height, 7.4.1.2 */
	BitmapHeight int

	/** Region segment bitmap X location, 7.4.1.3 */
	XLocation int

	/** Region segment bitmap Y location, 7.4.1.4 */
	YLocation int

	/** Region segment flags, 7.4.1.5 */
	CombinaionOperator bitmap.CombinationOperator
}

// NewRegionSegment creates new Region segment model
func NewRegionSegment(r reader.StreamReader) *RegionSegment {
	rs := &RegionSegment{r: r}
	return rs
}

// parseHeader parses the RegionSegment Header
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

// // Decode decodes the Region segment's basic paramters
// func (rs *RegionSegment) Decode(r io.Reader) (err error) {
// 	buf := make([]byte, 4)

// 	if _, err = r.Read(buf); err != nil {
// 		return
// 	}

// 	rs.BMWidth = int(binary.BigEndian.Uint32(buf))

// 	buf = make([]byte, 4)
// 	if _, err = r.Read(buf); err != nil {
// 		return
// 	}

// 	rs.BMHeight = int(binary.BigEndian.Uint32(buf))

// 	buf = make([]byte, 4)
// 	if _, err = r.Read(buf); err != nil {
// 		return
// 	}

// 	rs.BMXLocation = int(binary.BigEndian.Uint32(buf))

// 	buf = make([]byte, 4)
// 	if _, err = r.Read(buf); err != nil {
// 		return
// 	}

// 	rs.BMYLocation = int(binary.BigEndian.Uint32(buf))

// 	buf = make([]byte, 1)
// 	if _, err = r.Read(buf); err != nil {
// 		return
// 	}

// 	rs.RegionFlags.SetValue(int(buf[0]))
// 	return
// }
