package model

import (
	"sort"
	"testing"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model/internal/fonts"
)

func TestCIDWidthArr(t *testing.T) {
	widths := map[rune]int{
		'a': 1,
		'b': 1,
		'c': 1,
		'd': 2,
		'e': 3,
		'f': 3,
		'g': 4,
	}
	gids := map[rune]fonts.GID{
		'a': 1,
		'b': 2,
		'c': 3,
		'd': 4,
		'e': 5,
		'f': 6,
		'g': 7,
	}
	var runes []rune
	for r := range widths {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool {
		return gids[runes[i]] < gids[runes[j]]
	})

	arr := makeCIDWidthArr(runes, widths, gids)

	var out []int64
	for i := 0; i < arr.Len(); i++ {
		v := arr.Get(i).(*core.PdfObjectInteger)
		out = append(out, int64(*v))
	}

	exp := []int64{
		// gid1, gid2, width
		1, 3, 1,
		4, 4, 2,
		5, 6, 3,
		7, 7, 4,
	}
	if len(out) != len(exp) {
		t.Fatalf("\n%v\nvs\n%v", out, exp)
	}
	for i := range exp {
		if exp[i] != out[i] {
			t.Fatalf("\n%v\nvs\n%v", out, exp)
		}
	}
}
