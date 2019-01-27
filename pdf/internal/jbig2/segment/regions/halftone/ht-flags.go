package halftone

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
)

const (
	H_MMR         string = "H_MMR"
	H_TEMPLATE    string = "H_TEMPLATE"
	H_ENABLE_SKIP string = "H_ENABLE_SKIP"
	H_COMB_OP     string = "H_COMB_OP"
	H_DEF_PIXEL   string = "H_DEF_PIXEL"
)

// HalftoneSegmentFlags are the flags used in the halftone region segment
type HalftoneSegmentFlags struct {
	*flags.Flags
}

func newFlags() *HalftoneSegmentFlags {
	return &HalftoneSegmentFlags{flags.New()}
}

// SetValue sets the HalfonteSegment
func (h *HalftoneSegmentFlags) SetValue(flag int) {
	h.IntFlags = flag

	// Extract H_MMR on 0th bit
	h.Map[H_MMR] = flag & 1

	// Extract H_TEMPLATE 1st and 2nd bit
	h.Map[H_TEMPLATE] = (flag >> 1) & 3

	// Extract H_ENABLE_SKIP 3rd bit
	h.Map[H_TEMPLATE] = (flag >> 3) & 1

	// Extract H_COMB_OP on the 4,5,6 bits
	h.Map[H_COMB_OP] = (flag >> 4) & 7

	// Extract H_DEF_PIXEL on the 7th bit
	h.Map[H_DEF_PIXEL] = (flag >> 7) & 1

	common.Log.Debug("Halfton Segment Flags set with values: %v", h.Map)
}
