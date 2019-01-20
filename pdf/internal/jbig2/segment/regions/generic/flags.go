package generic

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
)

const (
	MMR        = "MMR"
	GBTemplate = "GB_TEMPLATE"
	TPGDOn     = "TPGDON"
)

// GenericRegionFlags is the GenericRegionSegment flagger container
type GenericRegionFlags struct {
	*flags.Flags
}

func newFlags() *GenericRegionFlags {
	return &GenericRegionFlags{
		Flags: flags.New(),
	}
}

// SetValue implements the method for the SegmentFlagger interface
// Sets the flag with provided value
func (g *GenericRegionFlags) SetValue(flag int) {
	g.IntFlags = flag

	// Extract MMR
	g.Flags.Map[MMR] = (flag & 1)

	// Extract GB_TEMPLATE
	g.Flags.Map[GBTemplate] = int((flag >> 1) & 3)

	// Extract TPGDON
	g.Flags.Map[TPGDOn] = int((flag >> 3) & 1)
}
