/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import "testing"

func TestCMapIdentityH_CharacterCodesToCID(t *testing.T) {
	identityCMap := CMapIdentityH{}

	type dataPair struct {
		raw      []byte
		expected []CID
		errs     bool
	}

	dataPairs := []dataPair{
		{[]byte{0x00, 0x00, 0x04, 0xff}, []CID{0x0000, 0x04ff}, false},
		{[]byte{0x00, 0x00, 0x04}, []CID{0x0000, 0x04ff}, true},
	}

	for _, data := range dataPairs {
		cids, err := identityCMap.CharacterCodesToCID(data.raw)
		if err != nil {
			if data.errs {
				continue
			}
			t.Errorf("Failed: %v", err)
			return
		}

		if len(data.expected) != len(cids) {
			t.Errorf("Length mismatch")
			return
		}

		for i := 0; i < len(data.expected); i++ {
			if cids[i] != data.expected[i] {
				t.Errorf("Not equal")
			}
		}
	}
}
