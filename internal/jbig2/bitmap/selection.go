/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

// SelectionValue is the enum value used for the Selection data.
type SelectionValue int

// SelectionValue enums values.
const (
	SelDontCare SelectionValue = iota
	SelHit
	SelMiss
)

// Selection is the structure used for matching the bitmaps.
type Selection struct {
	Height, Width int
	// Location of the Selection origin.
	Cx, Cy int
	Name   string
	Data   [][]SelectionValue
}

// findMaxTranslations finds the maximum shifts for the erosion operation.
func (s *Selection) findMaxTranslations() (pxp, pyp, pxn, pyn int) {
	for i := 0; i < s.Height; i++ {
		for j := 0; j < s.Width; j++ {
			if s.Data[i][j] == SelHit {
				pxp = max(pxp, s.Cx-j)
				pyp = max(pyp, s.Cy-i)
				pxn = max(pxn, j-s.Cx)
				pyn = max(pyn, i-s.Cy)
			}
		}
	}
	return pxp, pyp, pxn, pyn
}

func (s *Selection) setOrigin(cy, cx int) {
	s.Cy, s.Cx = cy, cx
}

// SelCreateBrick creates a rectangular selection of all hits, misses or don't cares.
func SelCreateBrick(h, w int, cy, cx int, tp SelectionValue) *Selection {
	sel := selCreate(h, w, "")
	sel.setOrigin(cy, cx)
	var i, j int
	for i = 0; i < h; i++ {
		for j = 0; j < w; j++ {
			sel.Data[i][j] = tp
		}
	}
	return sel
}

func selCreate(h, w int, name string) *Selection {
	sel := &Selection{Height: h, Width: w, Name: name}
	sel.Data = make([][]SelectionValue, h)
	for i := 0; i < h; i++ {
		sel.Data[i] = make([]SelectionValue, w)
	}
	return sel
}
