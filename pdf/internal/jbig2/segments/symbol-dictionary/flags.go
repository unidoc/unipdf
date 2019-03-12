package symboldict

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
)

const (
	SD_HUFF            = "SD_HUFF"
	SD_REF_AGG         = "SD_REF_AGG"
	SD_HUFF_DH         = "SD_HUFF_DH"
	SD_HUFF_DW         = "SD_HUFF_DW"
	SD_HUFF_BM_SIZE    = "SD_HUFF_BM_SIZE"
	SD_HUFF_AGG_INST   = "SD_HUFF_AGG_INST"
	BITMAP_CC_USED     = "BITMAP_CC_USED"
	BITMAP_CC_RETAINED = "BITMAP_CC_RETAINED"
	SD_TEMPLATE        = "SD_TEMPLATE"
	SD_R_TEMPLATE      = "SD_R_TEMPLATE"
)

type SymbolDictFlags struct {
	*flags.Flags
}

// SetValue implements flags SegmentFlager interface 'SetValue' method
func (s *SymbolDictFlags) SetValue(flagValue int) {
	// Extract SD_HUFF flag
	s.Map[SD_HUFF] = flagValue & 1

	// Extract SD_REF_AGG flag
	s.Map[SD_REF_AGG] = ((flagValue >> 1) & 1)

	// Extract SD_HUFF_DH
	s.Map[SD_HUFF_DH] = ((flagValue >> 2) & 3)

	// Extract SD_HUFF_DW
	s.Map[SD_HUFF_DW] = ((flagValue >> 4) & 3)

	// Extract SD_HUFF_BM_SIZE
	s.Map[SD_HUFF_BM_SIZE] = ((flagValue >> 6) & 1)

	// Extract SD_HUFF_AGG_INST
	s.Map[SD_HUFF_AGG_INST] = ((flagValue >> 7) & 1)

	// Extract BITMAP_CC_USED
	s.Map[BITMAP_CC_USED] = ((flagValue >> 8) & 1)

	// Extract BITMAP_CC_RETAINED
	s.Map[BITMAP_CC_RETAINED] = ((flagValue >> 9) & 1)

	// Extract SD_TEMPLATE
	s.Map[SD_TEMPLATE] = ((flagValue >> 10) & 3)

	// Extract SD_R_TEMPLATE
	s.Map[SD_R_TEMPLATE] = ((flagValue >> 12) & 1)

	common.Log.Debug("Symbol Dictionary Flags: %v", s.Map)
}
