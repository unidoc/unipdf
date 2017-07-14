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

	// Column width fractions: should add up to 1.
	colWidths []float64

	// Row heights.
	rowHeights []float64

	// Default row height.
	rowHeight float64

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
func NewTable(rows, cols int) *Table {
	t := &Table{}
	t.rows = rows
	t.cols = cols

	// Initialize column widths as all equal.
	t.colWidths = []float64{}
	colWidth := float64(1.0) / float64(cols)
	for i := 0; i < cols; i++ {
		t.colWidths = append(t.colWidths, colWidth)
	}

	// Initialize row heights all to 10.
	t.rowHeights = []float64{}
	for i := 0; i < rows; i++ {
		t.rowHeights = append(t.rowHeights, 10.0)
	}
	// Default row height
	t.rowHeight = 10.0

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

// Table occupies whatever width is available. Returns -1.
func (table *Table) Width() float64 {
	return -1 // Occupy whatever width is available.
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

// Set absolute coordinates.
// Note that this is only sensible to use when the table does not wrap over multiple pages.
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
			h += table.rowHeights[cell.col+i-1]
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
func (table *Table) NewCell(row, col int) *tableCell {
	cell := &tableCell{}
	cell.row = row
	cell.col = col

	cell.rowspan = 1
	cell.colspan = 1

	table.cells = append(table.cells, cell)

	// Keep reference to the table.
	cell.table = table

	return cell
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
