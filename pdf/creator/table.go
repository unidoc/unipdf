/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"

	"github.com/unidoc/unidoc/common"
)

// Table allows organizing content in an rows X columns matrix, which can spawn across multiple pages.
type Table struct {
	// Number of rows and columns.
	rows int
	cols int

	// Current cell.  Current cell in the table.
	// For 4x4 table, if in the 2nd row, 3rd column, then
	// curCell = 4+3 = 7
	curCell int

	// Column width fractions: should add up to 1.
	colWidths []float64

	// Row heights.
	rowHeights []float64

	// Default row height.
	defaultRowHeight float64

	// Content cells.
	cells []*tableCell

	// Positioning: relative / absolute.
	positioning positioning

	// Absolute coordinates (when in absolute mode).
	xPos, yPos float64

	// Margins to be applied around the block when drawing on Page.
	margins margins
}

// Create a new table with a fixed rows and column size.
func NewTable(cols int) *Table {
	t := &Table{}
	t.rows = 0
	t.cols = cols

	t.curCell = 0

	// Initialize column widths as all equal.
	t.colWidths = []float64{}
	colWidth := float64(1.0) / float64(cols)
	for i := 0; i < cols; i++ {
		t.colWidths = append(t.colWidths, colWidth)
	}

	t.rowHeights = []float64{}

	// Default row height
	// XXX/TODO: Base on contents instead?
	t.defaultRowHeight = 10.0

	t.cells = []*tableCell{}

	return t
}

// Set the fractional column widths. Number of width inputs must match number of columns.
// Each width should be in the range 0-1 and is a fraction of the table width.
func (table *Table) SetColumnWidths(widths ...float64) error {
	if len(widths) != table.cols {
		common.Log.Debug("Mismatching number of widths and columns")
		return errors.New("Range check error")
	}

	table.colWidths = widths

	return nil
}

// Total height of all rows.
func (table *Table) Height() float64 {
	sum := float64(0.0)
	for _, h := range table.rowHeights {
		sum += h
	}

	return sum
}

// Set the left, right, top, bottom Margins.
func (table *Table) SetMargins(left, right, top, bottom float64) {
	table.margins.left = left
	table.margins.right = right
	table.margins.top = top
	table.margins.bottom = bottom
}

// Get the left, right, top, bottom Margins.
func (table *Table) GetMargins() (float64, float64, float64, float64) {
	return table.margins.left, table.margins.right, table.margins.top, table.margins.bottom
}

// Set the height for a specified row.
func (table *Table) SetRowHeight(row int, h float64) error {
	if row < 1 || row > len(table.rowHeights) {
		return errors.New("Range check error")
	}

	table.rowHeights[row-1] = h
	return nil
}

// Get row of the current cell position.
func (table *Table) CurRow() int {
	curRow := (table.curCell-1)/table.cols + 1
	return curRow
}

// Get column of the current cell position.
func (table *Table) CurCol() int {
	curCol := (table.curCell-1)%(table.cols) + 1
	return curCol
}

// Set absolute coordinates.
// Note that this is only sensible to use when the table does not wrap over multiple pages.
// XXX/TODO: Should be able to set width too (not just based on context/relative positioning mode).
func (table *Table) SetPos(x, y float64) {
	table.positioning = positionAbsolute
	table.xPos = x
	table.yPos = y
}

// Generate the Page blocks.  Multiple blocks are generated if the contents wrap over multiple pages.
func (table *Table) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	// Iterate through all the cells
	tableWidth := ctx.Width

	blocks := []*Block{}
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	origCtx := ctx
	if table.positioning.isAbsolute() {
		ctx.X = table.xPos
		ctx.Y = table.yPos
	}

	// Store table's upper left corner.
	ulX := ctx.X
	ulY := ctx.Y

	ctx.Height = ctx.PageHeight - ctx.Y - ctx.Margins.bottom
	origHeight := ctx.Height

	// Start row keeps track of starting row (wraps to 0 on new page).
	startrow := 0

	// row height, cell height
	for _, cell := range table.cells {
		// Get total width fraction
		wf := float64(0.0)
		for i := 0; i < cell.colspan; i++ {
			wf += table.colWidths[cell.col+i-1]
		}
		// Get x pos relative to table upper left corner.
		xrel := float64(0.0)
		for i := 0; i < cell.col-1; i++ {
			xrel += table.colWidths[i] * tableWidth
		}
		// Get y pos relative to table upper left corner.
		yrel := float64(0.0)
		for i := startrow; i < cell.row-1; i++ {
			yrel += table.rowHeights[i]
		}

		// Calculate the width out of available width.
		w := wf * tableWidth

		// Get total height.
		h := float64(0.0)
		for i := 0; i < cell.rowspan; i++ {
			h += table.rowHeights[cell.row+i-1]
		}

		ctx.Height = origHeight - yrel
		if h > ctx.Height {
			// Go to next page.
			blocks = append(blocks, block)
			block = NewBlock(ctx.PageWidth, ctx.PageHeight)
			ulX = ctx.Margins.left
			ulY = ctx.Margins.top
			ctx.Height = ctx.PageHeight - ctx.Margins.top - ctx.Margins.bottom

			startrow = cell.row - 1
			yrel = 0
		}

		// Height should be how much space there is left of the page.
		ctx.Width = w
		ctx.X = ulX + xrel
		ctx.Y = ulY + yrel

		err := block.DrawWithContext(cell.content, ctx)
		if err != nil {
			common.Log.Debug("Error: %v\n", err)
		}
	}
	blocks = append(blocks, block)

	if table.positioning.isAbsolute() {
		return blocks, origCtx, nil
	}

	return blocks, ctx, nil
}

// Table cell
type tableCell struct {
	// The row and column which the cell starts from.
	row, col int

	// Row, column span.
	rowspan int
	colspan int

	// Each cell can contain 1 drawable.
	content Drawable

	// Table reference
	table *Table
}

// Make a new cell and insert into the table at specified row and column.
func (table *Table) NewCell() *tableCell {
	table.curCell++

	curRow := (table.curCell-1)/table.cols + 1
	for curRow > table.rows {
		table.rows++
		table.rowHeights = append(table.rowHeights, table.defaultRowHeight)
	}
	curCol := (table.curCell-1)%(table.cols) + 1

	cell := &tableCell{}
	cell.row = curRow
	cell.col = curCol

	cell.rowspan = 1
	cell.colspan = 1

	table.cells = append(table.cells, cell)

	// Keep reference to the table.
	cell.table = table

	return cell
}

// Skip over a specified number of cells.
func (table *Table) SkipCells(num int) {
	if num < 0 {
		common.Log.Debug("Table: cannot skip back to previous cells")
		return
	}
	table.curCell += num
}

// Skip over a specified number of rows.
func (table *Table) SkipRows(num int) {
	ncells := num*table.cols - 1
	if ncells < 0 {
		common.Log.Debug("Table: cannot skip back to previous cells")
		return
	}
	table.curCell += ncells
}

// Skip over rows, cols.
func (table *Table) SkipOver(rows, cols int) {
	ncells := rows*table.cols + cols - 1
	if ncells < 0 {
		common.Log.Debug("Table: cannot skip back to previous cells")
		return
	}
	table.curCell += ncells
}

// Get cell width based on input draw context.
func (cell *tableCell) Width(ctx DrawContext) float64 {
	fraction := float64(0.0)
	for j := 0; j < cell.colspan; j++ {
		fraction += cell.table.colWidths[cell.col+j-1]
	}
	w := ctx.Width * fraction
	return w
}

// Set cell content.
func (cell *tableCell) SetContent(d Drawable) error {
	switch d.(type) {
	case *paragraph:
		cell.content = d
	default:
		common.Log.Debug("Error: unsupported cell content type %T\n", d)
		return errors.New("Type check error")
	}

	return nil
}
