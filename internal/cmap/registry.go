/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmap

import (
	"github.com/unidoc/unipdf/v3/internal/cmap/bcmaps"
)

var cmapCache = map[string]*CMap{}

func LoadPredefinedCMap(name string) (*CMap, error) {
	if cmap, ok := cmapCache[name]; ok {
		return cmap, nil
	}

	cmapData, err := bcmaps.Asset(name)
	if err != nil {
		return nil, err
	}

	cmap, err := LoadCmapFromDataCID(cmapData)
	if err != nil {
		return nil, err
	}
	if cmap.usecmap == "" {
		return cmap, nil
	}

	// TODO: load base cmap if one is used.
	return cmap, nil
}

func IsPredefinedCMap(name string) bool {
	return bcmaps.AssetExists(name)
}
