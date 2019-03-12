package regions

import (
	"encoding/binary"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/model"
	"io"
)

const (
	ExternalCombinationOperator = "EXTERNAL_COMBINATION_OPERATOR"
)

// RegionFlag is the container for the RegionSegment flags
// Implements SegmentFlager interface
type RegionFlags struct {
	*flags.Flags
}

// SetValue sets the value for given flags
func (r *RegionFlags) SetValue(flag int) {
	r.IntFlags = flag
	r.Map[ExternalCombinationOperator] = (flag & 7)
}

// hardcoded check for SegmentFlager interface
var _ flags.SegmentFlager = &RegionFlags{}

// RegionSegment is the segment that takes
type RegionSegment struct {
	*model.Segment
	BMWidth, BMHeight        int
	BMXLocation, BMYLocation int
	RegionFlags              *RegionFlags
}

func NewRegionSegment(decoder *container.Decoder, h *header.Header) *RegionSegment {
	rs := &RegionSegment{
		Segment:     model.New(decoder, h),
		RegionFlags: &RegionFlags{Flags: flags.New()},
	}
	return rs
}

// Decode decodes the Region segment's basic paramters
func (rs *RegionSegment) Decode(r io.Reader) (err error) {
	buf := make([]byte, 4)

	if _, err = r.Read(buf); err != nil {
		return
	}

	rs.BMWidth = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)
	if _, err = r.Read(buf); err != nil {
		return
	}

	rs.BMHeight = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)
	if _, err = r.Read(buf); err != nil {
		return
	}

	rs.BMXLocation = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)
	if _, err = r.Read(buf); err != nil {
		return
	}

	rs.BMYLocation = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 1)
	if _, err = r.Read(buf); err != nil {
		return
	}

	rs.RegionFlags.SetValue(int(buf[0]))
	return
}
