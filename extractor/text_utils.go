/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"math"
	"sort"
)

// serial is used to add serial numbers to all text* instances.
var serial serialState

// serialState keeps serial number for text* structs.
type serialState struct {
	mark    int // textMark
	word    int // textWord
	wordBag int // wordBag
	line    int // textLine
	para    int // textPara
}

// reset resets `serial` to all zeros.
func (serial *serialState) reset() {
	var empty serialState
	*serial = empty
}

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
