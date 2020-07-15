/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package imageutil

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCMYK_ColorAt(t *testing.T) {
	g, err := NewImage(10, 10, 8, 4, nil, nil, nil)
	require.NoError(t, err)

	c := color.CMYK{C: 255, M: 240, Y: 200, K: 100}
	g.Set(5, 5, c)

	cmyk, ok := g.(CMYK)
	require.True(t, ok)

	assert.Equal(t, c, cmyk.CMYKAt(5, 5))

	cc, err := g.ColorAt(5, 5)
	require.NoError(t, err)

	cCMYK, ok := cc.(color.CMYK)
	require.True(t, ok)

	assert.Equal(t, c, cCMYK)
}
