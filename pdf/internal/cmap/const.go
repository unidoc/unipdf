package cmap

import "regexp"

const (
	cisSystemInfo       = "/CIDSystemInfo"
	begincodespacerange = "begincodespacerange"
	endcodespacerange   = "endcodespacerange"
	beginbfchar         = "beginbfchar"
	endbfchar           = "endbfchar"
	beginbfrange        = "beginbfrange"
	endbfrange          = "endbfrange"

	cmapname = "CMapName"
	cmaptype = "CMapType"
)

var reNumeric = regexp.MustCompile(`^[\+-.]*([0-9.]+)`)
