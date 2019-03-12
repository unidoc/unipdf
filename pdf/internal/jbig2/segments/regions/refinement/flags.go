package refinement

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
)

const (
	GR_TEMPLATE string = "GR_TEMPLATE"
	TPGDON      string = "TPGDON"
)

// RefinementRegionFlags is the model for the flags encoded in the RefinementRegionSegment
type RefinementRegionFlags struct {
	*flags.Flags
}

func newFlags() *RefinementRegionFlags {
	return &RefinementRegionFlags{
		Flags: flags.New(),
	}
}

// SetValue sets the flags value from the provided ints
func (r *RefinementRegionFlags) SetValue(flag int) {
	// save the flag
	r.IntFlags = flag

	// Read the GR_TEMPLATE flag 0th bit
	r.Map[GR_TEMPLATE] = flag & 1

	// Read the TPGDON flag at 1st bit
	r.Map[TPGDON] = (flag >> 1) & 1

	common.Log.Debug("RefinementRegionFlags set with values: %v", r.Map)
}
