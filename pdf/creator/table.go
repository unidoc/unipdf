/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model"
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
	cells []*TableCell

	// Positioning: relative / absolute.
	positioning positioning

	// Absolute coordinates (when in absolute mode).
	xPos, yPos float64

	// Margins to be applied around the block when drawing on Page.
	margins margins
}

// NewTable create a new Table with a specified number of columns.
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

	t.cells = []*TableCell{}

	return t
}

// SetColumnWidths sets the fractional column widths.
// Each width should be in the range 0-1 and is a fraction of the table width.
// The number of width inputs must match number of columns, otherwise an error is returned.
func (table *Table) SetColumnWidths(widths ...float64) error {
	if len(widths) != table.cols {
		common.Log.Debug("Mismatching number of widths and columns")
		return errors.New("Range check error")
	}

	table.colWidths = widths

	return nil
}

// Height returns the total height of all rows.
func (table *Table) Height() float64 {
	sum := float64(0.0)
	for _, h := range table.rowHeights {
		sum += h
	}

	return sum
}

// SetMargins sets the Table's left, right, top, bottom margins.
func (table *Table) SetMargins(left, right, top, bottom float64) {
	table.margins.left = left
	table.margins.right = right
	table.margins.top = top
	table.margins.bottom = bottom
}

// GetMargins returns the left, right, top, bottom Margins.
func (table *Table) GetMargins() (float64, float64, float64, float64) {
	return table.margins.left, table.margins.right, table.margins.top, table.margins.bottom
}

// SetRowHeight sets the height for a specified row.
func (table *Table) SetRowHeight(row int, h float64) error {
	if row < 1 || row > len(table.rowHeights) {
		return errors.New("Range check error")
	}

	table.rowHeights[row-1] = h
	return nil
}

// CurRow returns the currently active cell's row number.
func (table *Table) CurRow() int {
	curRow := (table.curCell-1)/table.cols + 1
	return curRow
}

// CurCol returns the currently active cell's column number.
func (table *Table) CurCol() int {
	curCol := (table.curCell-1)%(table.cols) + 1
	return curCol
}

// SetPos sets the Table's positioning to absolute mode and specifies the upper-left corner coordinates as (x,y).
// Note that this is only sensible to use when the table does not wrap over multiple pages.
// TODO: Should be able to set width too (not just based on context/relative positioning mode).
func (table *Table) SetPos(x, y float64) {
	table.positioning = positionAbsolute
	table.xPos = x
	table.yPos = y
}

// GeneratePageBlocks generate the page blocks.  Multiple blocks are generated if the contents wrap over multiple pages.
// Implements the Drawable interface.
func (table *Table) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	blocks := []*Block{}
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	origCtx := ctx
	if table.positioning.isAbsolute() {
		ctx.X = table.xPos
		ctx.Y = table.yPos
	} else {
		// Relative mode: add margins.
		ctx.X += table.margins.left
		ctx.Y += table.margins.top
		ctx.Width -= table.margins.left + table.margins.right
		ctx.Height -= table.margins.bottom + table.margins.top
	}
	tableWidth := ctx.Width

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

		if cell.backgroundColor != nil {
			// Draw background (fill)
			rect := NewRectangle(ctx.X, ctx.Y, w, h)
			r := cell.backgroundColor.R()
			g := cell.backgroundColor.G()
			b := cell.backgroundColor.B()
			rect.SetFillColor(ColorRGBFromArithmetic(r, g, b))
			if cell.borderStyle != CellBorderStyleNone {
				// and border.
				rect.SetBorderWidth(cell.borderWidth)
				r := cell.borderColor.R()
				g := cell.borderColor.G()
				b := cell.borderColor.B()
				rect.SetBorderColor(ColorRGBFromArithmetic(r, g, b))
			} else {
				rect.SetBorderWidth(0)
			}
			err := block.Draw(rect)
			if err != nil {
				common.Log.Debug("Error: %v\n", err)
			}
		} else if cell.borderStyle != CellBorderStyleNone {
			// Draw border (no fill).
			rect := NewRectangle(ctx.X, ctx.Y, w, h)
			rect.SetBorderWidth(cell.borderWidth)
			r := cell.borderColor.R()
			g := cell.borderColor.G()
			b := cell.borderColor.B()
			rect.SetBorderColor(ColorRGBFromArithmetic(r, g, b))
			err := block.Draw(rect)
			if err != nil {
				common.Log.Debug("Error: %v\n", err)
			}
		}

		if cell.content != nil {
			// Account for horizontal alignment:
			cw := cell.content.Width() // content width.
			switch cell.horizontalAlignment {
			case CellHorizontalAlignmentLeft:
				// Account for indent.
				ctx.X += cell.indent
				ctx.Width -= cell.indent
			case CellHorizontalAlignmentCenter:
				// Difference between available space and content space.
				dw := w - cw
				if dw > 0 {
					ctx.X += dw / 2
					ctx.Width -= dw / 2
				}
			case CellHorizontalAlignmentRight:
				if w > cw {
					ctx.X = ctx.X + w - cw - cell.indent
					ctx.Width = cw
				}
			}

			// Account for vertical alignment.
			ch := cell.content.Height() // content height.
			switch cell.verticalAlignment {
			case CellVerticalAlignmentTop:
				// Default: do nothing.
			case CellVerticalAlignmentMiddle:
				dh := h - ch
				if dh > 0 {
					ctx.Y += dh / 2
					ctx.Height -= dh / 2
				}
			case CellVerticalAlignmentBottom:
				if h > ch {
					ctx.Y = ctx.Y + h - ch
					ctx.Height = ch
				}
			}

			err := block.DrawWithContext(cell.content, ctx)
			if err != nil {
				common.Log.Debug("Error: %v\n", err)
			}
		}

		ctx.Y += h
	}
	blocks = append(blocks, block)

	if table.positioning.isAbsolute() {
		return blocks, origCtx, nil
	} else {
		// Move back X after.
		ctx.X = origCtx.X
		// Return original width
		ctx.Width = origCtx.Width
		// Add the bottom margin
		ctx.Y += table.margins.bottom
	}

	return blocks, ctx, nil
}

// CellBorderStyle defines the table cell's border style.
type CellBorderStyle int

// Currently supported table styles are: None (no border) and boxed (line along each side).
const (
	// No border
	CellBorderStyleNone CellBorderStyle = iota

	// Borders along all sides (boxed).
	CellBorderStyleBox
)

// CellHorizontalAlignment defines the table cell's horizontal alignment.
type CellHorizontalAlignment int

// Table cells have three horizontal alignment modes: left, center and right.
const (
	// Align cell content on the left (with specified indent); unused space on the right.
	CellHorizontalAlignmentLeft CellHorizontalAlignment = iota

	// Align cell content in the middle (unused space divided equally on the left/right).
	CellHorizontalAlignmentCenter

	// Align the cell content on the right; unsued space on the left.
	CellHorizontalAlignmentRight
)

// CellVerticalAlignment defines the table cell's vertical alignment.
type CellVerticalAlignment int

// Table cells have three vertical alignment modes: top, middle and bottom.
const (
	// Align cell content vertically to the top; unused space below.
	CellVerticalAlignmentTop CellVerticalAlignment = iota

	// Align cell content in the middle; unused space divided equally above and below.
	CellVerticalAlignmentMiddle

	// Align cell content on the bottom; unused space above.
	CellVerticalAlignmentBottom
)

// TableCell defines a table cell which can contain a Drawable as content.
type TableCell struct {
	// Background
	backgroundColor *model.PdfColorDeviceRGB

	// Border
	borderStyle CellBorderStyle
	borderColor *model.PdfColorDeviceRGB
	borderWidth float64

	// The row and column which the cell starts from.
	row, col int

	// Row, column span.
	rowspan int
	colspan int

	// Each cell can contain 1 drawable.
	content VectorDrawable

	// Alignment
	horizontalAlignment CellHorizontalAlignment
	verticalAlignment   CellVerticalAlignment

	// Left indent.
	indent float64

	// Table reference
	table *Table
}

// NewCell makes a new cell and inserts into the table at current position in the table.
func (table *Table) NewCell() *TableCell {
	table.curCell++

	curRow := (table.curCell-1)/table.cols + 1
	for curRow > table.rows {
		table.rows++
		table.rowHeights = append(table.rowHeights, table.defaultRowHeight)
	}
	curCol := (table.curCell-1)%(table.cols) + 1

	cell := &TableCell{}
	cell.row = curRow
	cell.col = curCol

	// Default left indent
	cell.indent = 5

	cell.borderStyle = CellBorderStyleNone
	cell.borderColor = model.NewPdfColorDeviceRGB(0, 0, 0)

	// Alignment defaults.
	cell.horizontalAlignment = CellHorizontalAlignmentLeft
	cell.verticalAlignment = CellVerticalAlignmentTop

	cell.rowspan = 1
	cell.colspan = 1

	table.cells = append(table.cells, cell)

	// Keep reference to the table.
	cell.table = table

	return cell
}

// SkipCells skips over a specified number of cells in the table.
func (table *Table) SkipCells(num int) {
	if num < 0 {
		common.Log.Debug("Table: cannot skip back to previous cells")
		return
	}
	table.curCell += num
}

// SkipRows skips over a specified number of rows in the table.
func (table *Table) SkipRows(num int) {
	ncells := num*table.cols - 1
	if ncells < 0 {
		common.Log.Debug("Table: cannot skip back to previous cells")
		return
	}
	table.curCell += ncells
}

// SkipOver skips over a specified number of rows and cols.
func (table *Table) SkipOver(rows, cols int) {
	ncells := rows*table.cols + cols - 1
	if ncells < 0 {
		common.Log.Debug("Table: cannot skip back to previous cells")
		return
	}
	table.curCell += ncells
}

// SetIndent sets the cell's left indent.
func (cell *TableCell) SetIndent(indent float64) {
	cell.indent = indent
}

// SetHorizontalAlignment sets the cell's horizontal alignment of content.
// Can be one of:
// - CellHorizontalAlignmentLeft
// - CellHorizontalAlignmentCenter
// - CellHorizontalAlignmentRight
func (cell *TableCell) SetHorizontalAlignment(halign CellHorizontalAlignment) {
	cell.horizontalAlignment = halign
}

// SetVerticalAlignment set the cell's vertical alignment of content.
// Can be one of:
// - CellHorizontalAlignmentTop
// - CellHorizontalAlignmentMiddle
// - CellHorizontalAlignmentBottom
func (cell *TableCell) SetVerticalAlignment(valign CellVerticalAlignment) {
	cell.verticalAlignment = valign
}

// SetBorder sets the cell's border style.
func (cell *TableCell) SetBorder(style CellBorderStyle, width float64) {
	cell.borderStyle = style
	cell.borderWidth = width
}

// SetBorderColor sets the cell's border color.
func (cell *TableCell) SetBorderColor(color rgbColor) {
	cell.borderColor = model.NewPdfColorDeviceRGB(color.r, color.g, color.b)
}

// SetBackgroundColor sets the cell's background color.
func (cell *TableCell) SetBackgroundColor(col Color) {
	cell.backgroundColor = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// Width returns the cell's width based on the input draw context.
func (cell *TableCell) Width(ctx DrawContext) float64 {
	fraction := float64(0.0)
	for j := 0; j < cell.colspan; j++ {
		fraction += cell.table.colWidths[cell.col+j-1]
	}
	w := ctx.Width * fraction
	return w
}

// SetContent sets the cell's content.  The content is a VectorDrawable, i.e. a Drawable with a known height and width.
// The currently supported VectorDrawable is: *Paragraph.
// TODO: Add support for *Image, *Block.
func (cell *TableCell) SetContent(vd VectorDrawable) error {
	switch t := vd.(type) {
	case *Paragraph:
		// Default paragraph settings in table:
		t.SetEnableWrap(false) // No wrapping.
		h := cell.table.rowHeights[cell.row-1]
		nh := t.Height() * 1.5 // Default multiplier 1.5.
		// Increase height if needed.
		if nh > h {
			cell.table.SetRowHeight(cell.row, nh)
		}
		cell.content = vd
	default:
		common.Log.Debug("Error: unsupported cell content type %T\n", vd)
		return errors.New("Type check error")
	}

	return nil
}
