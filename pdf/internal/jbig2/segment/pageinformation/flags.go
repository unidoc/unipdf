package pageinformation

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/flags"
)

const (
	DefaultPixelValue          = "DEFAULT_PIXEL_VALUE"
	DefaultCombinationOperator = "DEFAULT_COMBINATION_OPERATOR"
)

type Flags struct {
	*flags.Flags
}

func (f *Flags) SetValue(flag int) {
	f.IntFlags = flag
	f.Map[DefaultPixelValue] = ((flag >> 2) & 1)
	f.Map[DefaultCombinationOperator] = ((flag >> 3) & 3)

	common.Log.Debug("Set Pageinformation flags: %v", f.Map)
}
