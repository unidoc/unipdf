/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fonts

import (
	"fmt"

	"github.com/unidoc/unipdf/v3/internal/textencoding"
)

// Font represents a font which is a series of glyphs. Character codes from PDF strings can be
// mapped to and from glyphs. Each glyph has metrics.
type Font interface {
	Encoder() textencoding.TextEncoder
	GetRuneMetrics(r rune) (CharMetrics, bool)
}

// CharMetrics represents width and height metrics of a glyph.
type CharMetrics struct {
	Wx float64
	Wy float64 // TODO(dennwc): none of code paths sets this to anything except 0
}

func (m CharMetrics) String() string {
	return fmt.Sprintf("<%.1f,%.1f>", m.Wx, m.Wy)
}
