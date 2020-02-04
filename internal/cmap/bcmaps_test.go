/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmap

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unidoc/unipdf/v3/internal/cmap/bcmaps"
)

func TestIsPredefinedCMap(t *testing.T) {
	cmaps := bcmaps.AssetNames()
	for _, cmap := range cmaps {
		require.True(t, IsPredefinedCMap(cmap))
	}
}

func TestLoadPredefinedCMap(t *testing.T) {
	// Load base charcode to CID CMap.
	cmap, err := LoadPredefinedCMap("90ms-RKSJ-H")
	require.NoError(t, err)

	require.Equal(t, cmap.name, "90ms-RKSJ-H")
	require.Equal(t, cmap.ctype, 1)
	require.Equal(t, cmap.nbits, 16)
	require.Equal(t, cmap.usecmap, "")
	require.Equal(t, cmap.systemInfo.String(), "Adobe-Japan1-002")
	require.Equal(t, cmap.codespaces, []Codespace{
		Codespace{NumBytes: 1, Low: 0x0, High: 0x80},
		Codespace{NumBytes: 1, Low: 0xa0, High: 0xdf},
		Codespace{NumBytes: 2, Low: 0x8140, High: 0x9ffc},
		Codespace{NumBytes: 2, Low: 0xe040, High: 0xfcfc},
	})

	expCIDs := map[CharCode]CharCode{
		32:    231,
		125:   324,
		37952: 3287,
		38014: 3349,
		38016: 3350,
		38140: 3474,
		160:   326,
		223:   389,
		64576: 8706,
		64587: 8717,
		34670: 8047,
		33972: 7554,
	}

	for code, expCID := range expCIDs {
		cid, ok := cmap.CharcodeToCID(code)
		require.True(t, ok)
		require.Equal(t, cid, expCID)
	}

	// Load charcode to CID CMap using base.
	cmap, err = LoadPredefinedCMap("90ms-RKSJ-V")
	require.NoError(t, err)

	require.Equal(t, cmap.name, "90ms-RKSJ-V")
	require.Equal(t, cmap.ctype, 1)
	require.Equal(t, cmap.nbits, 16)
	require.Equal(t, cmap.usecmap, "90ms-RKSJ-H")
	require.Equal(t, cmap.systemInfo.String(), "Adobe-Japan1-002")
	require.Equal(t, cmap.codespaces, []Codespace{
		Codespace{NumBytes: 1, Low: 0x0, High: 0x80},
		Codespace{NumBytes: 1, Low: 0xa0, High: 0xdf},
		Codespace{NumBytes: 2, Low: 0x8140, High: 0x9ffc},
		Codespace{NumBytes: 2, Low: 0xe040, High: 0xfcfc},
	})

	// Test CID inheritance and CID override.
	expCIDs[34670] = 8349
	expCIDs[33972] = 7554
	expCIDs[33089] = 7887
	expCIDs[33090] = 7888

	for code, expCID := range expCIDs {
		cid, ok := cmap.CharcodeToCID(code)
		require.True(t, ok)
		require.Equal(t, cid, expCID)
	}
}
