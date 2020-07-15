/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"sort"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

// textTable is a table of `w` x `h` textPara cells.
type textTable struct {
	model.PdfRectangle                      // Bounding rectangle.
	w, h               int                  // w=number of columns. h=number of rows.
	cells              map[uint64]*textPara // The cells
}

// String returns a description of `t`.
func (t *textTable) String() string {
	return fmt.Sprintf("%d x %d", t.w, t.h)
}

// bbox makes textLine implement the `bounded` interface.
func (t *textTable) bbox() model.PdfRectangle {
	return t.PdfRectangle
}

// extractTables converts the`paras` that are table cells to tables containing those cells.
func (paras paraList) extractTables() paraList {
	if verboseTable {
		common.Log.Debug("extractTables=%d ===========x=============", len(paras))
	}
	if len(paras) < minTableParas {
		return paras
	}
	tables := paras.findTables()
	if verboseTable {
		common.Log.Info("combined tables %d ================", len(tables))
		for i, t := range tables {
			t.log(fmt.Sprintf("combined %d", i))
		}
	}
	return paras.applyTables(tables)
}

// findTables returns all the tables  in `paras`.
func (paras paraList) findTables() []*textTable {
	paras.addNeighbours()
	// Pre-sort by reading direction then depth
	sort.Slice(paras, func(i, j int) bool {
		return diffReadingDepth(paras[i], paras[j]) < 0
	})

	var tables []*textTable
	for _, para := range paras {
		if para.isCell {
			continue
		}
		table := para.isAtom()
		if table == nil {
			continue
		}

		table.growTable()
		if table.w*table.h < minTableParas {
			continue
		}
		table.markCells()
		table.log("grown")
		tables = append(tables, table)

	}
	return tables
}

// isAtom atempts to build the smallest possible table fragment of 2 x 2 cells.
// If a table can be built then it is returned. Otherwise nil is returned.
// The smallest possible table is
//   a b
//   c d
// where
//   a is `para`.
//   b is immediately to the right of a and overlaps it in the y axis.
//   c is immediately below a and overlaps it in the x axis.
//   d is immediately to the right of c and overlaps it in the y axis and
//        immediately below b and ooverlaps it in the s axis.
//   None of a, b, c or d are cells in existing tables.
func (para *textPara) isAtom() *textTable {
	a := para
	b := para.right
	c := para.below
	if !(b != nil && !b.isCell && c != nil && !c.isCell) {
		return nil
	}
	d := b.below
	if !(d != nil && !d.isCell && d == c.right) {
		return nil
	}

	if b.left != a || c.above != a || d.left != c || d.above != b {
		return nil
	}
	return newTableAtom(a, b, c, d)
}

// newTable returns a table containing the a, b, c, d elements from isAtom().
func newTableAtom(a, b, c, d *textPara) *textTable {
	t := &textTable{w: 2, h: 2, cells: map[uint64]*textPara{}}
	t.put(0, 0, a)
	t.put(1, 0, b)
	t.put(0, 1, c)
	t.put(1, 1, d)
	return t
}

// growTable grows `t` to the largest w x h it can while remaining a valid table.
// It repeatedly tries to extend by one row and/or column
//    - down and right, then
//    - down, then
//    - right.
func (t *textTable) growTable() {
	growDown := func(down paraList) {
		t.h++
		for x := 0; x < t.w; x++ {
			cell := down[x]
			t.put(x, t.h-1, cell)
		}
	}
	growRight := func(right paraList) {
		t.w++
		for y := 0; y < t.h; y++ {
			cell := right[y]
			t.put(t.w-1, y, cell)
		}
	}

	for {
		changed := false
		down := t.getDown()
		right := t.getRight()
		if down != nil && right != nil {
			downRight := down[len(down)-1]
			if downRight != nil && !downRight.isCell && downRight == right[len(right)-1] {
				growDown(down)
				growRight(right)
				t.put(t.w-1, t.h-1, downRight)
				changed = true
			}
		}
		if !changed && down != nil {
			growDown(down)
			changed = true
		}
		if !changed && right != nil {
			growRight(right)
			changed = true
		}
		if !changed {
			break
		}
	}
}

// getDown returns the row of cells below `t` if they are a valid extension to `t` or nil if they aren't.
func (t *textTable) getDown() paraList {
	cells := make(paraList, t.w)
	for x := 0; x < t.w; x++ {
		cell := t.get(x, t.h-1).below
		if cell == nil || cell.isCell {
			return nil
		}
		cells[x] = cell
	}
	for x := 0; x < t.w-1; x++ {
		if cells[x].right != cells[x+1] {
			return nil
		}
	}
	return cells
}

// getRight returns the column of cells to the right `t` if they are a valid extension to `t` or nil
// if they aren't.
func (t *textTable) getRight() paraList {
	cells := make(paraList, t.h)
	for y := 0; y < t.h; y++ {
		cell := t.get(t.w-1, y).right
		if cell == nil || cell.isCell {
			return nil
		}
		cells[y] = cell
	}
	for y := 0; y < t.h-1; y++ {
		if cells[y].below != cells[y+1] {
			return nil
		}
	}
	return cells
}

// applyTables replaces the paras that are cells in `tables` with paras containing the tables in
//`tables`. This, of course, reduces the number of paras.
func (paras paraList) applyTables(tables []*textTable) paraList {
	consumed := map[*textPara]struct{}{}
	var tabled paraList
	for _, table := range tables {
		for _, para := range table.cells {
			consumed[para] = struct{}{}
		}
		tabled = append(tabled, table.newTablePara())
	}
	for _, para := range paras {
		if _, ok := consumed[para]; !ok {
			tabled = append(tabled, para)
		}
	}
	return tabled
}

// markCells marks the paras that are cells in `t` with isCell=true so that the won't be considered
// as cell candidates for tables in the future.
func (t *textTable) markCells() {
	for y := 0; y < t.h; y++ {
		for x := 0; x < t.w; x++ {
			para := t.get(x, y)
			para.isCell = true
		}
	}
}

// newTablePara returns a textPara containing `t`.
func (t *textTable) newTablePara() *textPara {
	bbox := t.computeBbox()
	return &textPara{
		PdfRectangle: bbox,
		eBBox:        bbox,
		table:        t,
	}
}

// computeBbox computes and returns the bounding box of `t`.
func (t *textTable) computeBbox() model.PdfRectangle {
	r := t.get(0, 0).PdfRectangle
	for x := 1; x < t.w; x++ {
		r = rectUnion(r, t.get(x, 0).PdfRectangle)
	}
	for y := 1; y < t.h; y++ {
		for x := 0; x < t.w; x++ {
			r = rectUnion(r, t.get(x, y).PdfRectangle)
		}
	}
	return r
}

// toTextTable returns the TextTable corresponding to `t`.
func (t *textTable) toTextTable() TextTable {
	cells := make([][]TableCell, t.h)
	for y := 0; y < t.h; y++ {
		cells[y] = make([]TableCell, t.w)
		for x := 0; x < t.w; x++ {
			c := t.get(x, y)
			cells[y][x].Text = c.text()
			offset := 0
			cells[y][x].Marks.marks = c.toTextMarks(&offset)
		}
	}
	return TextTable{W: t.w, H: t.h, Cells: cells}
}

// get returns the cell at `x`, `y`.
func (t *textTable) get(x, y int) *textPara {
	return t.cells[cellIndex(x, y)]
}

// put sets the cell at `x`, `y` to `cell`.
func (t *textTable) put(x, y int, cell *textPara) {
	t.cells[cellIndex(x, y)] = cell
}

// cellIndex returns a number that will be different for different `x` and `y` for any table found
// in a PDF which will less than 2^32 wide and hight.
func cellIndex(x, y int) uint64 {
	return uint64(x)*0x1000000 + uint64(y)
}

func (t *textTable) log(title string) {
	if !verboseTable {
		return
	}
	common.Log.Info("~~~ %s: %d x %d\n      %6.2f", title,
		t.w, t.h, t.PdfRectangle)
	for y := 0; y < t.h; y++ {
		for x := 0; x < t.w; x++ {
			p := t.get(x, y)
			fmt.Printf("%4d %2d: %6.2f %q\n", x, y, p.PdfRectangle, truncate(p.text(), 50))
		}
	}
}
