/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"math"
	"sort"
	"unicode"
)

// TOL is the tolerance for coordinates to be consideted equal. It is big enough to cover all
// rounding errors and small enough that TOL point differences on a page aren't visible.
const TOL = 1.0e-6

// isZero returns true if x is with TOL of 0.0
func isZero(x float64) bool {
	return math.Abs(x) < TOL
}

// minInt return the lesser of `a` and `b`.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt return the greater of `a` and `b`.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// addNeighbours fills out the below and right fields of the paras in `paras`.
// For each para `a`:
//    a.below is the unique highest para completely below `a` that overlaps it in the x-direction
//    a.right is the unique leftmost para completely to the right of `a` that overlaps it in the y-direction
func (paras paraList) addNeighbours() {
	paraNeighbours := paras.yNeighbours()
	for _, para := range paras {
		var left *textPara
		dup := false
		for _, k := range paraNeighbours[para] {
			b := paras[k]
			if b.Urx <= para.Llx {
				if left == nil {
					left = b
				} else {
					if b.Llx > left.Llx {
						left = b
						dup = false
					} else if b.Llx == left.Llx {
						dup = true
					}
				}
			}
		}
		if !dup {
			para.left = left
		}
	}
	for _, para := range paras {
		var right *textPara
		dup := false
		for _, k := range paraNeighbours[para] {
			b := paras[k]
			if b.Llx >= para.Urx {
				if right == nil {
					right = b
				} else {
					if b.Llx < right.Llx {
						right = b
						dup = false
					} else if b.Llx == right.Llx {
						dup = true
					}
				}
			}
		}
		if !dup {
			para.right = right
		}
	}

	paraNeighbours = paras.xNeighbours()
	for _, para := range paras {
		var above *textPara
		dup := false
		for _, i := range paraNeighbours[para] {
			b := paras[i]
			if b.Lly >= para.Ury {
				if above == nil {
					above = b
				} else {
					if b.Ury < above.Ury {
						above = b
						dup = false
					} else if b.Ury == above.Ury {
						dup = true
					}
				}
			}
		}
		if !dup {
			para.above = above
		}
	}
	for _, para := range paras {
		var below *textPara
		dup := false
		for _, i := range paraNeighbours[para] {
			b := paras[i]
			if b.Ury <= para.Lly {
				if below == nil {
					below = b
				} else {
					if b.Ury > below.Ury {
						below = b
						dup = false
					} else if b.Ury == below.Ury {
						dup = true
					}
				}
			}
		}
		if !dup {
			para.below = below
		}
	}
}

// xNeighbours returns a map {para: indexes of paras that x-overlap para}.
func (paras paraList) xNeighbours() map[*textPara][]int {
	events := make([]event, 2*len(paras))
	for i, para := range paras {
		events[2*i] = event{para.Llx, true, i}
		events[2*i+1] = event{para.Urx, false, i}
	}
	return paras.eventNeighbours(events)
}

// yNeighbours returns a map {para: indexes of paras that y-overlap para}.
func (paras paraList) yNeighbours() map[*textPara][]int {
	events := make([]event, 2*len(paras))
	for i, para := range paras {
		events[2*i] = event{para.Lly, true, i}
		events[2*i+1] = event{para.Ury, false, i}
	}
	return paras.eventNeighbours(events)
}

// event is an entry or exit from an interval while scanning.
type event struct {
	z     float64 // Coordinate in the scanning direction.
	enter bool    // True if entering the interval, false it leaving.
	i     int     // Index of the interval
}

// eventNeighbours returns a map {para: indexes of paras that overlap para in `events`}.
func (paras paraList) eventNeighbours(events []event) map[*textPara][]int {
	sort.Slice(events, func(i, j int) bool {
		ei, ej := events[i], events[j]
		zi, zj := ei.z, ej.z
		if zi != zj {
			return zi < zj
		}
		if ei.enter != ej.enter {
			return ei.enter
		}
		return i < j
	})

	overlaps := map[int]map[int]struct{}{}
	olap := map[int]struct{}{}
	for _, e := range events {
		if e.enter {
			overlaps[e.i] = map[int]struct{}{}
			for i := range olap {
				if i != e.i {
					overlaps[e.i][i] = struct{}{}
					overlaps[i][e.i] = struct{}{}
				}
			}
			olap[e.i] = struct{}{}
		} else {
			delete(olap, e.i)
		}
	}

	paraNeighbors := map[*textPara][]int{}
	for i, olap := range overlaps {
		para := paras[i]
		neighbours := make([]int, len(olap))
		k := 0
		for j := range olap {
			neighbours[k] = j
			k++
		}
		paraNeighbors[para] = neighbours
	}
	return paraNeighbors
}

// isTextSpace returns true if `text` contains nothing but space code points.
func isTextSpace(text string) bool {
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// combiningDiacritic returns the combining version of `text` if text contains a single uncombined
// diacritic rune.
func combiningDiacritic(text string) (string, bool) {
	runes := []rune(text)
	if len(runes) != 1 {
		return "", false
	}
	combining, isDiacritic := diacriticsToCombining[runes[0]]
	return combining, isDiacritic
}

var (
	// diacriticsToCombining is a map of diacritic runes to their combining diacritic equivalents.
	// These values were  copied from  (https://svn.apache.org/repos/asf/pdfbox/trunk/pdfbox/src/main/java/org/apache/pdfbox/text/TextPosition.java)
	diacriticsToCombining = map[rune]string{
		0x0060: "\u0300", //   ` -> ò
		0x02CB: "\u0300", //   ˋ -> ò
		0x0027: "\u0301", //   ' -> ó
		0x00B4: "\u0301", //   ´ -> ó
		0x02B9: "\u0301", //   ʹ -> ó
		0x02CA: "\u0301", //   ˊ -> ó
		0x005E: "\u0302", //   ^ -> ô
		0x02C6: "\u0302", //   ˆ -> ô
		0x007E: "\u0303", //   ~ -> õ
		0x02DC: "\u0303", //   ˜ -> õ
		0x00AF: "\u0304", //   ¯ -> ō
		0x02C9: "\u0304", //   ˉ -> ō
		0x02D8: "\u0306", //   ˘ -> ŏ
		0x02D9: "\u0307", //   ˙ -> ȯ
		0x00A8: "\u0308", //   ¨ -> ö
		0x00B0: "\u030A", //   ° -> o̊
		0x02DA: "\u030A", //   ˚ -> o̊
		0x02BA: "\u030B", //   ʺ -> ő
		0x02DD: "\u030B", //   ˝ -> ő
		0x02C7: "\u030C", //   ˇ -> ǒ
		0x02C8: "\u030D", //   ˈ -> o̍
		0x0022: "\u030E", //   " -> o̎
		0x02BB: "\u0312", //   ʻ -> o̒
		0x02BC: "\u0313", //   ʼ -> o̓
		0x0486: "\u0313", //   ҆ -> o̓
		0x055A: "\u0313", //   ՚ -> o̓
		0x02BD: "\u0314", //   ʽ -> o̔
		0x0485: "\u0314", //   ҅ -> o̔
		0x0559: "\u0314", //   ՙ -> o̔
		0x02D4: "\u031D", //   ˔ -> o̝
		0x02D5: "\u031E", //   ˕ -> o̞
		0x02D6: "\u031F", //   ˖ -> o̟
		0x02D7: "\u0320", //   ˗ -> o̠
		0x02B2: "\u0321", //   ʲ -> o̡
		0x00B8: "\u0327", //   ¸ -> o̧
		0x02CC: "\u0329", //   ˌ -> o̩
		0x02B7: "\u032B", //   ʷ -> o̫
		0x02CD: "\u0331", //   ˍ -> o̱
		0x005F: "\u0332", //   _ -> o̲
		0x204E: "\u0359", //   ⁎ -> o͙
	}
)
