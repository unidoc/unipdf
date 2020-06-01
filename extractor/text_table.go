/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"fmt"
	"math"
	"sort"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

type textTable struct {
	model.PdfRectangle
	w, h  int
	cells cellList
}

func (t textTable) bbox() model.PdfRectangle {
	return t.PdfRectangle
}

type cellList paraList

const DBL_MIN, DBL_MAX = -1.0e10, +1.0e10

// extractTables converts the`paras` that are table cells to tables containing those cells.
func (paras paraList) extractTables() paraList {
	common.Log.Debug("extractTables=%d ===========x=============", len(paras))
	if len(paras) < 4 {
		return nil
	}
	show := func(title string) {
		common.Log.Info("%8s: %d=========----------=====", title, len(paras))
		for i, para := range paras {
			text := para.text()
			tabl := "  "
			if para.table != nil {
				tabl = fmt.Sprintf("[%dx%d]", para.table.w, para.table.h)
			}
			fmt.Printf("%4d: %6.2f %s %q\n", i, para.PdfRectangle, tabl, truncate(text, 50))
			if len(text) == 0 {
				panic("empty")
			}
			if para.table != nil && len(para.table.cells) == 0 {
				panic(para)
			}
		}
	}
	tables := paras.extractTableAtoms()
	tables = combineTables(tables)
	common.Log.Info("combined tables %d ================", len(tables))
	for i, t := range tables {
		t.log(fmt.Sprintf("combined %d", i))
	}
	// if len(tables) == 0 {panic("NO TABLES")}
	show("tables extracted")
	paras = paras.applyTables(tables)
	show("tables applied")
	paras = paras.trimTables()
	show("tables trimmed")

	return paras
}

func (paras paraList) trimTables() paraList {
	var recycledParas paraList
	seen := map[*textPara]bool{}
	for _, para := range paras {
		for _, p := range paras {
			if p == para {
				continue
			}
			table := para.table
			if table != nil && overlapped(table, p) {
				table.log("REMOVE")
				for _, cell := range table.cells {
					if _, ok := seen[cell]; ok {
						continue
					}
					recycledParas = append(recycledParas, cell)
					seen[cell] = true
				}
				para.table.cells = nil
			}
		}
	}

	for _, p := range paras {
		if p.table != nil && p.table.cells == nil {
			continue
		}
		recycledParas = append(recycledParas, p)
	}
	return recycledParas
}

func (paras paraList) applyTables(tables []textTable) paraList {
	// if len(tables) == 0 {panic("no tables")}
	consumed := map[*textPara]bool{}
	for _, table := range tables {
		if len(table.cells) == 0 {
			panic("no cells")
		}
		for _, para := range table.cells {
			consumed[para] = true
		}
	}
	// if len(consumed) == 0 {panic("no paras consumed")}

	var tabled paraList
	for _, table := range tables {
		if table.cells == nil {
			panic(table)
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

// extractTableAtome returns all the 2x2 table candidateds in `paras`.
func (paras paraList) extractTableAtoms() []textTable {
	// Pre-sort by reading direction then depth
	sort.Slice(paras, func(i, j int) bool {
		return diffReadingDepth(paras[i], paras[j]) < 0
	})

	var llx0, lly0, llx1, lly1 float64
	var tables []textTable

	for _, para1 := range paras {
		llx0, lly0 = DBL_MAX, DBL_MIN
		llx1, lly1 = DBL_MAX, DBL_MIN

		// Build a table fragment of 4 cells
		//   0 1
		//   2 3
		// where
		//   0 is `para1`
		//   1 is on the right of 0 and overlaps with 0 in y axis
		//   2 is under 0 and overlaps with 0 in x axis
		//   3 is under 1 and on the right of 1 and closest to 0
		cells := make(cellList, 4)
		cells[0] = para1

		for _, para2 := range paras {
			if para1 == para2 {
				continue
			}
			if yOverlap(para1, para2) && toRight(para2, para1) && para2.Llx < llx0 {
				llx0 = para2.Llx
				cells[1] = para2
			} else if xOverlap(para1, para2) && below(para2, para1) && para2.Ury > lly0 {
				lly0 = para2.Ury
				cells[2] = para2
			} else if toRight(para2, para1) && para2.Llx < llx1 && below(para2, para1) && para2.Ury > lly1 {
				llx1 = para2.Llx
				lly1 = para2.Ury
				cells[3] = para2
			}
		}
		// if we found any then look whether they form a table  !@#$
		if !(cells[1] != nil && cells[2] != nil && cells[3] != nil) {
			continue
		}
		// 1 cannot overlap with 2 in x and y
		// 3 cannot overlap with 2 in x and with 1 in y
		// 3 has to overlap with 2 in y and with 1 in x

		if (xOverlap(cells[2], cells[3]) || yOverlap(cells[1], cells[3]) ||
			xOverlap(cells[1], cells[2]) || yOverlap(cells[1], cells[2])) ||
			!(xOverlap(cells[1], cells[3]) && yOverlap(cells[2], cells[3])) {
			continue
		}

		// common.Log.Info("@@10 ip=%d %s", ip, truncate(para1.text(), 40))

		deltaX := cells.fontsize()
		deltaY := deltaX
		//       deltaX *= minColSpacing1;  !@#$
		//       deltaY *= maxIntraLineDelta;
		deltaX *= maxIntraReadingGapR
		deltaY *= lineDepthR

		correspondenceX := cells.alignedX(cells.fontsize() * maxIntraReadingGapR)
		correspondenceY := cells.alignedY(cells.fontsize() * lineDepthR)

		// are blocks aligned in x and y ?
		if correspondenceX > 0 && correspondenceY > 0 {
			table := newTable(cells, 2, 2)
			tables = append(tables, table)
			table.log("New textTable")
			// common.Log.Info("New textTable\n      %6.2f", table.PdfRectangle)
			// for i, p := range cells {
			// 	fmt.Printf("%4d: %6.2f %q\n", i, p.PdfRectangle, truncate(p.text(), 50))
			// }
		}
	}
	return tables
}

func (table textTable) log(title string) {
	common.Log.Info("~~~ %s: %s: %d x %d\n      %6.2f", title, fileLine(1, false),
		table.w, table.h, table.PdfRectangle)
	for i, p := range table.cells {
		fmt.Printf("%4d: %6.2f %q\n", i, p.PdfRectangle, truncate(p.text(), 50))
	}
}

// 0 1
// 2 3
// A B
// C
// Extensions:
//   A[1] == B[0] right
//   A[2] == C[0] down
func combineTables(tables []textTable) []textTable {
	// if len(tables) == 0 {panic("tables")}
	tablesY := combineTablesY(tables)
	// if len(tablesY) == 0 {	panic("tablesY")}
	heightTables := map[int][]textTable{}
	for _, table := range tablesY {
		heightTables[table.h] = append(heightTables[table.h], table)
	}
	// if len(heightTables) == 0 {panic("heightTables")}
	var heights []int
	for h := range heightTables {
		heights = append(heights, h)
	}
	// Try to extend tallest tables to the right
	sort.Slice(heights, func(i, j int) bool { return heights[i] > heights[j] })
	// for _, h := range heights {
	// 	columns := heightTables[h]
	// 	if len(columns) < 2 {
	// 		continue
	// 	}
	// 	heightTables[h] = combineTablesX(columns)
	// }

	var combined []textTable
	for _, h := range heights {
		combined = append(combined, heightTables[h]...)
	}
	for i, table := range combined {
		table.log(fmt.Sprintf("Combined %d", i))
	}
	return combined
}

func combineTablesY(tables []textTable) []textTable {
	sort.Slice(tables, func(i, j int) bool { return tables[i].Ury > tables[j].Ury })
	removed := map[int]bool{}

	var combinedTables []textTable
	common.Log.Info("combineTablesY ------------------\n\t ------------------")
	for i1, t1 := range tables {
		if _, ok := removed[i1]; ok {
			continue
		}
		fontsize := t1.cells.fontsize()
		c1 := t1.corners()
		var combo *textTable
		for i2, t2 := range tables {
			if _, ok := removed[i2]; ok {
				continue
			}
			if t1.w != t2.w {
				continue
			}
			c2 := t2.corners()
			if c1[2] != c2[0] {
				continue
			}
			// common.Log.Info("Comparing i1=%d i2=%d", i1, i2)
			// t1.log("t1")
			// t2.log("t2")
			cells := cellList{
				c1[0], c1[1],
				c2[2], c2[3],
			}
			alX := cells.alignedX(fontsize * maxIntraReadingGapR)
			alY := cells.alignedY(fontsize * lineDepthR)
			common.Log.Info("alX=%d alY=%d", alX, alY)
			if !(alX > 0 && alY > 0) {
				if combo != nil {
					combinedTables = append(combinedTables, *combo)
				}
				combo = nil
				continue
			}
			if combo == nil {
				combo = &t1
				removed[i1] = true
			}

			w := combo.w
			h := combo.h + t2.h - 1
			common.Log.Info("COMBINE! %dx%d", w, h)
			combined := make(cellList, w*h)
			for y := 0; y < t1.h; y++ {
				for x := 0; x < w; x++ {
					combined[y*w+x] = combo.cells[y*w+x]
				}
			}
			for y := 1; y < t2.h; y++ {
				yy := y + combo.h - 1
				for x := 0; x < w; x++ {
					combined[yy*w+x] = t2.cells[y*w+x]
				}
			}
			combo.cells = combined
			combo.h = h
			combo.log("combo")
			removed[i2] = true
			fontsize = combo.cells.fontsize()
			c1 = combo.corners()
		}
		if combo != nil {
			combinedTables = append(combinedTables, *combo)
		}
	}

	common.Log.Info("combineTablesY a: combinedTables=%d", len(combinedTables))
	for i, t := range tables {
		if _, ok := removed[i]; ok {
			continue
		}
		combinedTables = append(combinedTables, t)
	}
	common.Log.Info("combineTablesY b: combinedTables=%d", len(combinedTables))

	return combinedTables
}

func combineTablesX(tables []textTable) []textTable {
	sort.Slice(tables, func(i, j int) bool { return tables[i].Llx < tables[j].Llx })
	removed := map[int]bool{}
	for i1, t1 := range tables {
		if _, ok := removed[i1]; ok {
			continue
		}
		fontsize := t1.cells.fontsize()
		c1 := t1.corners()
		for i2, t2 := range tables {
			if _, ok := removed[i2]; ok {
				continue
			}
			if t1.w != t2.w {
				continue
			}
			c2 := t2.corners()
			if c1[1] != c2[0] {
				continue
			}
			cells := cellList{
				c1[0], c2[1],
				c1[2], c2[3],
			}
			if !(cells.alignedX(fontsize*maxIntraReadingGapR) > 0 &&
				cells.alignedY(fontsize*lineDepthR) > 0) {
				continue
			}
			w := t1.w + t2.w
			h := t1.h
			combined := make(cellList, w*h)
			for y := 0; y < h; y++ {
				for x := 0; x < t1.w; x++ {
					combined[y*w+x] = t1.cells[y*w+x]
				}
				for x := 0; x < t2.w; x++ {
					xx := x + t1.w
					combined[y*w+xx] = t1.cells[y*w+x]
				}
			}
			removed[i2] = true
			fontsize = t1.cells.fontsize()
			c1 = t1.corners()
		}
	}
	var reduced []textTable
	for i, t := range tables {
		if _, ok := removed[i]; ok {
			continue
		}
		reduced = append(reduced, t)
	}
	return reduced
}

func yOverlap(para1, para2 *textPara) bool {
	//  blk2->yMin <= blk1->yMax &&blk2->yMax >= blk1->yMin
	return para2.Lly <= para1.Ury && para1.Lly <= para2.Ury
}
func xOverlap(para1, para2 *textPara) bool {
	//  blk2->yMin <= blk1->yMax &&blk2->yMax >= blk1->yMin
	return para2.Llx <= para1.Urx && para1.Llx <= para2.Urx
}
func toRight(para2, para1 *textPara) bool {
	//  blk2->yMin <= blk1->yMax &&blk2->yMax >= blk1->yMin
	return para2.Llx > para1.Urx
}
func below(para2, para1 *textPara) bool {
	//  blk2->yMin <= blk1->yMax &&blk2->yMax >= blk1->yMin
	return para2.Ury < para1.Lly
}

func (paras cellList) cellDepths() []float64 {
	topF := func(p *textPara) float64 { return p.Ury }
	botF := func(p *textPara) float64 { return p.Lly }
	top := paras.calcCellDepths(topF)
	bottom := paras.calcCellDepths(botF)
	if len(bottom) < len(top) {
		return bottom
	}
	return top
}

func (paras cellList) calcCellDepths(getY func(*textPara) float64) []float64 {
	depths := []float64{getY(paras[0])}
	delta := paras.fontsize() * maxIntraDepthGapR
	for _, para := range paras {
		newDepth := true
		y := getY(para)
		for _, d := range depths {
			if math.Abs(d-getY(para)) < delta {
				newDepth = false
				break
			}
		}
		if newDepth {
			depths = append(depths, y)
		}
	}
	return depths
}

func (c *textTable) corners() paraList {
	w, h := c.w, c.h
	if w == 0 || h == 0 {
		panic(c)
	}
	cnrs := paraList{
		c.cells[0],
		c.cells[w-1],
		c.cells[w*(h-1)],
		c.cells[w*h-1],
	}
	for i0, c0 := range cnrs {
		for _, c1 := range cnrs[:i0] {
			if c0.serial == c1.serial {
				panic("dup")
			}
		}
	}
	return cnrs
}

func newTable(cells cellList, w, h int) textTable {
	if w == 0 || h == 0 {
		panic("emprty")
	}
	for i0, c0 := range cells {
		for _, c1 := range cells[:i0] {
			if c0.serial == c1.serial {
				panic("dup")
			}
		}
	}
	rect := cells[0].PdfRectangle
	for _, c := range cells[1:] {
		rect = rectUnion(rect, c.PdfRectangle)
	}
	return textTable{
		PdfRectangle: rect,
		w:            w,
		h:            h,
		cells:        cells,
	}
}

func (table textTable) newTablePara() *textPara {
	cells := table.cells
	sort.Slice(cells, func(i, j int) bool { return diffDepthReading(cells[i], cells[j]) < 0 })
	table.cells = cells
	para := textPara{
		serial:       serial.para,
		PdfRectangle: table.PdfRectangle,
		eBBox:        table.PdfRectangle,
		table:        &table,
	}
	table.log(fmt.Sprintf("newTablePara: serial=%d", para.serial))

	serial.para++
	return &para
}

func (cells cellList) alignedX(delta float64) int {
	matches := 0
	for _, get := range gettersX {
		if cells.aligned(0, 2, delta, get) && cells.aligned(1, 3, delta, get) {
			matches++
		}
	}
	return matches
}

func (cells cellList) alignedY(delta float64) int {
	matches := 0
	for _, get := range gettersY {
		if cells.aligned(0, 1, delta, get) && cells.aligned(2, 3, delta, get) {
			matches++
		}
	}
	return matches
}

func (cells cellList) aligned(i, j int, delta float64, get getter) bool {
	return parasAligned(cells[i], cells[j], delta, get)
}

type getter func(*textPara) float64

var (
	gettersX = []getter{getXCe, getXLl, getXUr}
	gettersY = []getter{getYCe, getYLl, getYUr}
)

func getXCe(para *textPara) float64 { return 0.5 * (para.Llx + para.Urx) }
func getXLl(para *textPara) float64 { return para.Llx }
func getXUr(para *textPara) float64 { return para.Urx }
func getYCe(para *textPara) float64 { return 0.5 * (para.Lly + para.Ury) }
func getYLl(para *textPara) float64 { return para.Lly }
func getYUr(para *textPara) float64 { return para.Ury }

func parasAligned(para1, para2 *textPara, delta float64, get func(*textPara) float64) bool {
	z1 := get(para1)
	z2 := get(para2)
	return math.Abs(z1-z2) <= delta
}

// fontsize for a paraList is the minimum font size of the paras.
func (paras cellList) fontsize() float64 {
	size := paras[0].fontsize()
	for _, p := range paras[1:] {
		size = math.Min(size, p.fontsize())
	}
	return size
}
