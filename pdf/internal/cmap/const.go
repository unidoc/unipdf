/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmap

import (
	"errors"
	"regexp"
)

var (
	ErrBadCMap        = errors.New("Bad cmap")
	ErrBadCMapComment = errors.New("Comment should start with %")
	ErrBadCMapDict    = errors.New("Invalid dict")
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

var reNumeric = regexp.MustCompile(`^[\+-.]*([0-9.]+)`)
