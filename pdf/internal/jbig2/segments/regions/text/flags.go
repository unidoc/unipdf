package text

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
)

const (
	SbHuff       = "SB_HUFF"
	SBRefine     = "SB_REFINE"
	LogSbStripes = "LOG_SB_STRIPES"
	RefCorner    = "REF_CORNER"
	Transposed   = "TRANSPOSED"
	SbCombOp     = "SB_COMB_OP"
	SbDefPixel   = "SB_DEF_PIXEL"
	SbDsOffset   = "SB_DS_OFFSET"
	SbRTemplate  = "SB_R_TEMPLATE"
)

// Flags is the Text Region Segment flags container
type Flags struct {
	*flags.Flags
}

func newFlags() *Flags {
	return &Flags{
		Flags: flags.New(),
	}
}

func (f *Flags) SetValue(flag int) {
	f.Map[SbHuff] = flag & 1
	f.Map[SBRefine] = (flag >> 1) & 1
	f.Map[LogSbStripes] = (flag >> 2) & 3
	f.Map[RefCorner] = (flag >> 4) & 3
	f.Map[Transposed] = (flag >> 6) & 1
	f.Map[SbCombOp] = (flag >> 7) & 3
	f.Map[SbDefPixel] = (flag >> 9) & 1

	sOffset := flag >> 10 & 0x1f
	if sOffset&0x10 != 0 {
		sOffset |= -1 - 0x0f
	}

	f.Map[SbDsOffset] = sOffset
	f.Map[SbRTemplate] = (flag >> 15) & 1

	common.Log.Debug("Flags: %v", f.Map)
	common.Log.Debug("%016b", flag)

}

const (
	SBHUFFFS    = "SB_HUFF_FS"
	SBHUFFDS    = "SB_HUFF_DS"
	SBHUFFDT    = "SB_HUFF_DT"
	SBHUFFRDW   = "SB_HUFF_RDW"
	SBHUFFRDH   = "SB_HUFF_RDH"
	SBHUFFRDX   = "SB_HUFF_RDX"
	SBHUFFRDY   = "SB_HUFF_RDY"
	SBHUFFRSIZE = "SB_HUFF_RSIZE"
)

// HuffmanFlags contains Text Region Segment huffman flags
type HuffmanFlags struct {
	*flags.Flags
}

func (h *HuffmanFlags) SetValue(flag int) {
	h.Map[SBHUFFFS] = flag & 3

	h.Map[SBHUFFDS] = (flag >> 2) & 3
	h.Map[SBHUFFDT] = (flag >> 4) & 3
	h.Map[SBHUFFRDW] = (flag >> 6) & 3
	h.Map[SBHUFFRDH] = (flag >> 8) & 3
	h.Map[SBHUFFRDX] = (flag >> 10) & 3
	h.Map[SBHUFFRDY] = (flag >> 12) & 3
	h.Map[SBHUFFRSIZE] = (flag >> 14) & 1
	common.Log.Debug("HuffmanFlags: %v", h.Map)
	common.Log.Debug("%016b", flag)
}

// newHuffmanFlags is the creator for the HuffmanFlags
func newHuffmanFlags() *HuffmanFlags {
	return &HuffmanFlags{
		Flags: flags.New(),
	}
}
