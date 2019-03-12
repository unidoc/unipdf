package patterndict

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
)

const (
	HD_MMR      = "HD_MMR"
	HD_TEMPLATE = "HD_TEMPLATE"
)

// PatternDictionaryFlags are the flags defined for the PatternDictionarySegment
// Implement SegmentFlagger interface
type PatternDictionaryFlags struct {
	*flags.Flags
}

// SetValue sets the flag value in the structure
// Implements the SegmentFlagger interface method
func (p *PatternDictionaryFlags) SetValue(flag int) {
	p.IntFlags = flag

	// Extract HD_MMR from the first bit
	p.Map[HD_MMR] = (flag & 1)

	// Extratct HD_TEMPLATE from the second and third bit
	p.Map[HD_TEMPLATE] = ((flag >> 1) & 3)

	common.Log.Debug("PatternDictionaryFlags SetValue finished. Flags: %v", p.Map)
}

// NewPatternDictionaryFlags creates the pattern dictionary flags
func NewPatternDictionaryFlags() *PatternDictionaryFlags {
	return &PatternDictionaryFlags{Flags: flags.New()}
}
