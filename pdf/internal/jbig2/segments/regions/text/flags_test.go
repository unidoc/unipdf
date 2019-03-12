package text

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeFlags(t *testing.T) {

	t.Run("Main", func(t *testing.T) {
		f := newFlags()

		f.SetValue(0x0C09)

		t.Logf("%016b", 0x0c09)

		assert.Equal(t, true, f.GetValue(SbHuff) != 0)
		assert.Equal(t, false, f.GetValue(SBRefine) != 0)
		assert.Equal(t, 2, f.GetValue(LogSbStripes))

	})
}
