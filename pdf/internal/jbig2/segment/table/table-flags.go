package table

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
)

// TableSegmentFlags
type TableSegmentFlags struct {
	*flags.Flags
}

func newFlags() *TableSegmentFlags {
	return &TableSegmentFlags{
		Flags: flags.New(),
	}
}

func (f *TableSegmentFlags) SetValue(flag int) {
	// Set flags
	f.IntFlags = flag

}
