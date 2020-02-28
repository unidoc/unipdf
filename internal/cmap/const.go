/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmap

import (
	"errors"
)

// CMap parser errors.
var (
	ErrBadCMap        = errors.New("bad cmap")
	ErrBadCMapComment = errors.New("comment should start with %")
	ErrBadCMapDict    = errors.New("invalid dict")
)

const (
	cidSystemInfo       = "CIDSystemInfo"
	begincmap           = "begincmap"
	endcmap             = "endcmap"
	begincodespacerange = "begincodespacerange"
	endcodespacerange   = "endcodespacerange"
	beginbfchar         = "beginbfchar"
	endbfchar           = "endbfchar"
	beginbfrange        = "beginbfrange"
	endbfrange          = "endbfrange"
	begincidrange       = "begincidrange"
	endcidrange         = "endcidrange"
	usecmap             = "usecmap"

	cmapname    = "CMapName"
	cmaptype    = "CMapType"
	cmapversion = "CMapVersion"
)
