/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unidoc/unipdf/v3/model"
)

func TestTextChunkWrap(t *testing.T) {
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	tc := NewTextChunk(text, TextStyle{
		Font:     model.DefaultFont(),
		FontSize: 10,
	})

	// Check wrap when width <= 0.
	expectedLines := []string{text}

	lines, err := tc.Wrap(0)
	require.NoError(t, err)
	require.Equal(t, len(lines), len(expectedLines))
	require.Equal(t, lines, expectedLines)

	// Check wrap for width = 500.
	expectedLines = []string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore",
		"magna aliqua.\n",
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
		"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint",
		"occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
	}

	lines, err = tc.Wrap(500)
	require.NoError(t, err)
	require.Equal(t, len(lines), 5)
	require.Equal(t, lines, expectedLines)

	// Check wrap for width = 100.
	expectedLines = []string{
		"Lorem ipsum dolor sit",
		"amet, consectetur",
		"adipiscing elit, sed do",
		"eiusmod tempor",
		"incididunt ut labore et",
		"dolore magna aliqua.\n",
		"Ut enim ad minim",
		"veniam, quis nostrud",
		"exercitation ullamco",
		"laboris nisi ut aliquip",
		"ex ea commodo",
		"consequat. Duis aute",
		"irure dolor in",
		"reprehenderit in",
		"voluptate velit esse",
		"cillum dolore eu",
		"fugiat nulla pariatur.",
		"Excepteur sint",
		"occaecat cupidatat",
		"non proident, sunt in",
		"culpa qui officia",
		"deserunt mollit anim",
		"id est laborum.",
	}

	lines, err = tc.Wrap(100)
	require.NoError(t, err)
	require.Equal(t, len(lines), len(expectedLines))
	require.Equal(t, lines, expectedLines)

	// Check wrap for width = 2000.
	expectedLines = []string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n",
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
	}

	lines, err = tc.Wrap(2000)
	require.NoError(t, err)
	require.Equal(t, len(lines), len(expectedLines))
	require.Equal(t, lines, expectedLines)
}

func TestTextChunkFit(t *testing.T) {
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	tc := NewTextChunk(text, TextStyle{
		Font:     model.DefaultFont(),
		FontSize: 10,
	})

	expected := [][2]string{
		[2]string{
			"Lorem ipsum dolor sit",
			"amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		},
		[2]string{
			"amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et",
			"dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		},
		[2]string{
			"dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in",
			"reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		},
		[2]string{
			"reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"",
		},
	}

	for i := 1; i < 10; i++ {
		tc2, err := tc.Fit(float64(i*100), float64(i*10))
		require.NoError(t, err)

		remainder := ""
		if tc2 != nil {
			remainder = tc2.Text
		}
		require.Equal(t, tc.Text, expected[i-1][0])
		require.Equal(t, remainder, expected[i-1][1])

		if tc2 == nil {
			break
		}
		tc = tc2
	}
}
