/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package textencoding

import "github.com/unidoc/unidoc/pdf/core"

// CID represents a character identifier.
type CID uint16

// CMap maps character codes to CIDs.
type CMap interface {
	CharacterCodesToCID(charcodes []byte) ([]CID, error)
}

// CMapIdentityH is a representation of the /Identity-H cmap.
type CMapIdentityH struct {
}

// CharacterCodesToCID converts charcodes to CIDs for the Identity CMap, which maps
// 2-byte character codes (from the raw data) from 0-65535 to the same 2-byte CID value.
func (cmap CMapIdentityH) CharacterCodesToCID(raw []byte) ([]CID, error) {
	if len(raw)%2 != 0 {
		return nil, core.ErrRangeError
	}

	cids := []CID{}
	for i := 0; i < len(raw); i += 2 {
		b1 := CID(raw[i])
		b2 := CID(raw[i+1])
		cids = append(cids, (b1<<8)|b2)
	}
	return cids, nil
}
