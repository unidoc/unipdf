/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"math"
	"sort"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
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

	// Specifies whether the table has a header.
	hasHeader bool

	// Header rows.
	headerStartRow int
	headerEndRow   int
}

// newTable create a new Table with a specified number of columns.
func newTable(cols int) *Table {
	t := &Table{
		cols:             cols,
		defaultRowHeight: 10.0,
		colWidths:        []float64{},
		rowHeights:       []float64{},
		cells:            []*TableCell{},
	}

	t.resetColumnWidths()
	return t
}

// SetColumnWidths sets the fractional column widths.
// Each width should be in the range 0-1 and is a fraction of the table width.
// The number of width inputs must match number of columns, otherwise an error is returned.
func (table *Table) SetColumnWidths(widths ...float64) error {
	if len(widths) != table.cols {
		common.Log.Debug("Mismatching number of widths and columns")
		return errors.New("range check error")
	}

	table.colWidths = widths
	return nil
}

func (table *Table) resetColumnWidths() {
	table.colWidths = []float64{}
	colWidth := float64(1.0) / float64(table.cols)

	// Initialize column widths as all equal.
	for i := 0; i < table.cols; i++ {
		table.colWidths = append(table.colWidths, colWidth)
	}
}

// Height returns the total height of all rows.
func (table *Table) Height() float64 {
	sum := float64(0.0)
	for _, h := range table.rowHeights {
		sum += h
	}

	return sum
}

// Width is not used. Not used as a Table element is designed to fill into
// available width depending on the context. Returns 0.
func (table *Table) Width() float64 {
	return 0
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

// GetRowHeight returns the height of the specified row.
func (table *Table) GetRowHeight(row int) (float64, error) {
	if row < 1 || row > len(table.rowHeights) {
		return 0, errors.New("range check error")
	}

	return table.rowHeights[row-1], nil
}

// SetRowHeight sets the height for a specified row.
func (table *Table) SetRowHeight(row int, h float64) error {
	if row < 1 || row > len(table.rowHeights) {
		return errors.New("range check error")
	}

	table.rowHeights[row-1] = h
	return nil
}

// Rows returns the total number of rows the table has.
func (table *Table) Rows() int {
	return table.rows
}

// Cols returns the total number of columns the table has.
func (table *Table) Cols() int {
	return table.cols
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

// SetPos sets the Table's positioning to absolute mode and specifies the upper-left corner
// coordinates as (x,y).
// Note that this is only sensible to use when the table does not wrap over multiple pages.
// TODO: Should be able to set width too (not just based on context/relative positioning mode).
func (table *Table) SetPos(x, y float64) {
	table.positioning = positionAbsolute
	table.xPos = x
	table.yPos = y
}

// SetHeaderRows turns the selected table rows into headers that are repeated
// for every page the table spans. startRow and endRow are inclusive.
func (table *Table) SetHeaderRows(startRow, endRow int) error {
	if startRow <= 0 {
		return errors.New("header start row must be greater than 0")
	}
	if endRow <= 0 {
		return errors.New("header end row must be greater than 0")
	}
	if startRow > endRow {
		return errors.New("header start row  must be less than or equal to the end row")
	}

	table.hasHeader = true
	table.headerStartRow = startRow
	table.headerEndRow = endRow
	return nil
}

// AddSubtable copies the cells of the subtable in the table, starting with the
// specified position. The table row and column indices are 1-based, which
// makes the position of the first cell of the first row of the table 1,1.
// The table is automatically extended if the subtable exceeds its columns.
// This can happen when the subtable has more columns than the table or when
// one or more columns of the subtable starting from the specified position
// exceed the last column of the table.
func (table *Table) AddSubtable(row, col int, subtable *Table) {
	for _, cell := range subtable.cells {
		c := &TableCell{}
		*c = *cell
		c.table = table

		// Adjust added cell column. Add extra columns to the table to
		// accomodate the new cell, if needed.
		c.col += col - 1
		if colsLeft := table.cols - (c.col - 1); colsLeft < c.colspan {
			table.cols += c.colspan - colsLeft
			table.resetColumnWidths()
			common.Log.Debug("Table: subtable exceeds destination table. Expanding table to %d columns.", table.cols)
		}

		// Extend number of rows, if needed.
		c.row += row - 1

		subRowHeight := subtable.rowHeights[cell.row-1]
		if c.row > table.rows {
			for c.row > table.rows {
				table.rows++
				table.rowHeights = append(table.rowHeights, table.defaultRowHeight)
			}

			table.rowHeights[c.row-1] = subRowHeight
		} else {
			table.rowHeights[c.row-1] = math.Max(table.rowHeights[c.row-1], subRowHeight)
		}

		table.cells = append(table.cells, c)
	}

	// Sort cells by row, column.
	sort.Slice(table.cells, func(i, j int) bool {
		rowA := table.cells[i].row
		rowB := table.cells[j].row
		if rowA < rowB {
			return true
		}
		if rowA > rowB {
			return false
		}

		return table.cells[i].col < table.cells[j].col
	})
}

// GeneratePageBlocks generate the page blocks.  Multiple blocks are generated if the contents wrap
// over multiple pages.
// Implements the Drawable interface.
func (table *Table) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	var blocks []*Block
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

	// Indices of the first and the last header cells.
	startHeaderCell := -1
	endHeaderCell := -1

	// Prepare for drawing: Calculate cell dimensions, row, cell heights.
	for cellIdx, cell := range table.cells {
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

		// Calculate header cell range.
		if table.hasHeader {
			if cell.row >= table.headerStartRow && cell.row <= table.headerEndRow {
				if startHeaderCell < 0 {
					startHeaderCell = cellIdx
				}
				endHeaderCell = cellIdx
			}
		}

		// For text: Calculate width, height, wrapping within available space if specified.
		switch t := cell.content.(type) {
		case *Paragraph:
			p := t
			if p.enableWrap {
				p.SetWidth(w - cell.indent)
			}

			newh := p.Height() + p.margins.bottom + p.margins.bottom
			newh += 0.5 * p.fontSize * p.lineHeight // TODO: Make the top margin configurable?
			if newh > h {
				diffh := newh - h
				// Add diff to last row.
				table.rowHeights[cell.row+cell.rowspan-2] += diffh
			}
		case *StyledParagraph:
			sp := t
			if sp.enableWrap {
				sp.SetWidth(w - cell.indent)
			}

			newh := sp.Height() + sp.margins.top + sp.margins.bottom
			newh += 0.5 * sp.getTextHeight() // TODO: Make the top margin configurable?
			if newh > h {
				diffh := newh - h
				// Add diff to last row.
				table.rowHeights[cell.row+cell.rowspan-2] += diffh
			}
		case *Image:
			img := t
			newh := img.Height() + img.margins.top + img.margins.bottom
			if newh > h {
				diffh := newh - h
				// Add diff to last row.
				table.rowHeights[cell.row+cell.rowspan-2] += diffh
			}
		case *Table:
			tbl := t
			newh := tbl.Height() + tbl.margins.top + tbl.margins.bottom
			if newh > h {
				diffh := newh - h
				// Add diff to last row.
				table.rowHeights[cell.row+cell.rowspan-2] += diffh
			}
		case *List:
			lst := t
			newh := lst.tableHeight(w-cell.indent) + lst.margins.top + lst.margins.bottom
			if newh > h {
				diffh := newh - h
				// Add diff to last row.
				table.rowHeights[cell.row+cell.rowspan-2] += diffh
			}
		case *Division:
			div := t

			ctx := DrawContext{
				X:     xrel,
				Y:     yrel,
				Width: w,
			}

			// Mock call to generate page blocks.
			divBlocks, updCtx, err := div.GeneratePageBlocks(ctx)
			if err != nil {
				return nil, ctx, err
			}

			if len(divBlocks) > 1 {
				// Wraps across page, make cell reach all the way to bottom of current page.
				newh := ctx.Height - h
				if newh > h {
					diffh := newh - h
					// Add diff to last row.
					table.rowHeights[cell.row+cell.rowspan-2] += diffh
				}
			}

			newh := div.Height() + div.margins.top + div.margins.bottom
			_ = updCtx

			// Get available width and height.
			if newh > h {
				diffh := newh - h
				// Add diff to last row.
				table.rowHeights[cell.row+cell.rowspan-2] += diffh
			}
		}
	}

	// Draw cells.
	// row height, cell height
	var drawingHeaders bool
	var resumeIdx, resumeStartRow int

	for cellIdx := 0; cellIdx < len(table.cells); cellIdx++ {
		cell := table.cells[cellIdx]

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
			ctx.Page++
			origHeight = ctx.Height

			startrow = cell.row - 1
			yrel = 0

			// Save state and jump back to the first header cell.
			if table.hasHeader && startHeaderCell >= 0 {
				resumeIdx = cellIdx
				cellIdx = startHeaderCell - 1

				resumeStartRow = startrow
				startrow = table.headerStartRow - 1

				drawingHeaders = true
				continue
			}
		}

		// Height should be how much space there is left of the page.
		ctx.Width = w
		ctx.X = ulX + xrel
		ctx.Y = ulY + yrel

		// Creating border
		border := newBorder(ctx.X, ctx.Y, w, h)

		if cell.backgroundColor != nil {
			r := cell.backgroundColor.R()
			g := cell.backgroundColor.G()
			b := cell.backgroundColor.B()

			border.SetFillColor(ColorRGBFromArithmetic(r, g, b))
		}

		border.LineStyle = cell.borderLineStyle

		border.styleLeft = cell.borderStyleLeft
		border.styleRight = cell.borderStyleRight
		border.styleTop = cell.borderStyleTop
		border.styleBottom = cell.borderStyleBottom

		if cell.borderColorLeft != nil {
			border.SetColorLeft(ColorRGBFromArithmetic(cell.borderColorLeft.R(), cell.borderColorLeft.G(), cell.borderColorLeft.B()))
		}
		if cell.borderColorBottom != nil {
			border.SetColorBottom(ColorRGBFromArithmetic(cell.borderColorBottom.R(), cell.borderColorBottom.G(), cell.borderColorBottom.B()))
		}
		if cell.borderColorRight != nil {
			border.SetColorRight(ColorRGBFromArithmetic(cell.borderColorRight.R(), cell.borderColorRight.G(), cell.borderColorRight.B()))
		}
		if cell.borderColorTop != nil {
			border.SetColorTop(ColorRGBFromArithmetic(cell.borderColorTop.R(), cell.borderColorTop.G(), cell.borderColorTop.B()))
		}

		border.SetWidthBottom(cell.borderWidthBottom)
		border.SetWidthLeft(cell.borderWidthLeft)
		border.SetWidthRight(cell.borderWidthRight)
		border.SetWidthTop(cell.borderWidthTop)

		err := block.Draw(border)
		if err != nil {
			common.Log.Debug("ERROR: %v", err)
		}

		if cell.content != nil {
			cw := cell.content.Width()  // content width.
			ch := cell.content.Height() // content height.
			vertOffset := 0.0

			switch t := cell.content.(type) {
			case *Paragraph:
				if t.enableWrap {
					cw = t.getMaxLineWidth() / 1000.0
				}
			case *StyledParagraph:
				if t.enableWrap {
					cw = t.getMaxLineWidth() / 1000.0
				}

				// Calculate the height of the paragraph.
				lineCapHeight, lineHeight := t.getLineHeight(0)
				if len(t.lines) == 1 {
					ch = lineCapHeight
				} else {
					ch = ch - lineHeight + lineCapHeight
				}

				// Account for the top offset the paragraph adds.
				vertOffset = lineCapHeight - t.defaultStyle.FontSize*t.lineHeight

				switch cell.verticalAlignment {
				case CellVerticalAlignmentTop:
					// Add a bit of space from the top border of the cell.
					vertOffset += lineCapHeight * 0.5
				case CellVerticalAlignmentBottom:
					// Add a bit of space from the bottom border of the cell.
					vertOffset -= lineCapHeight * 0.5
				}
			case *Table:
				cw = w
			case *List:
				cw = w
			}

			// Account for horizontal alignment:
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

			ctx.Y += vertOffset

			// Account for vertical alignment.
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
					ctx.Height = h
				}
			}

			err := block.DrawWithContext(cell.content, ctx)
			if err != nil {
				common.Log.Debug("ERROR: %v", err)
			}

			ctx.Y -= vertOffset
		}

		ctx.Y += h

		// Resume previous state after headers have been rendered.
		if drawingHeaders && cellIdx+1 > endHeaderCell {
			// Account for the height of the rendered headers.
			ulY += yrel + h
			origHeight -= h + yrel

			startrow = resumeStartRow
			cellIdx = resumeIdx - 1

			drawingHeaders = false
		}
	}
	blocks = append(blocks, block)

	if table.positioning.isAbsolute() {
		return blocks, origCtx, nil
	}
	// Relative mode.
	// Move back X after.
	ctx.X = origCtx.X
	// Return original width.
	ctx.Width = origCtx.Width
	// Add the bottom margin.
	ctx.Y += table.margins.bottom

	return blocks, ctx, nil
}

// CellBorderStyle defines the table cell's border style.
type CellBorderStyle int

// Currently supported table styles are: None (no border) and boxed (line along each side).
const (
	CellBorderStyleNone CellBorderStyle = iota // no border

	// Borders along all sides (boxed).

	CellBorderStyleSingle
	CellBorderStyleDouble
)

// CellBorderSide defines the table cell's border side.
type CellBorderSide int

const (
	// CellBorderSideLeft adds border on the left side of the table.
	CellBorderSideLeft CellBorderSide = iota

	// CellBorderSideRight adds a border on the right side of the table.
	CellBorderSideRight

	// CellBorderSideTop adds a border on the top side of the table.
	CellBorderSideTop

	// CellBorderSideBottom adds a border on the bottom side of the table.
	CellBorderSideBottom

	// CellBorderSideAll adds borders on all sides of the table.
	CellBorderSideAll
)

// CellHorizontalAlignment defines the table cell's horizontal alignment.
type CellHorizontalAlignment int

// Table cells have three horizontal alignment modes: left, center and right.
const (
	// CellHorizontalAlignmentLeft aligns cell content on the left (with specified indent); unused space on the right.
	CellHorizontalAlignmentLeft CellHorizontalAlignment = iota

	// CellHorizontalAlignmentCenter aligns cell content in the middle (unused space divided equally on the left/right).
	CellHorizontalAlignmentCenter

	// CellHorizontalAlignmentRight aligns the cell content on the right; unsued space on the left.
	CellHorizontalAlignmentRight
)

// CellVerticalAlignment defines the table cell's vertical alignment.
type CellVerticalAlignment int

// Table cells have three vertical alignment modes: top, middle and bottom.
const (
	// CellVerticalAlignmentTop aligns cell content vertically to the top; unused space below.
	CellVerticalAlignmentTop CellVerticalAlignment = iota

	// CellVerticalAlignmentMiddle aligns cell content in the middle; unused space divided equally above and below.
	CellVerticalAlignmentMiddle

	// CellVerticalAlignmentBottom aligns cell content on the bottom; unused space above.
	CellVerticalAlignmentBottom
)

// TableCell defines a table cell which can contain a Drawable as content.
type TableCell struct {
	// Background
	backgroundColor *model.PdfColorDeviceRGB

	borderLineStyle draw.LineStyle

	// border
	borderStyleLeft   CellBorderStyle
	borderColorLeft   *model.PdfColorDeviceRGB
	borderWidthLeft   float64
	borderStyleBottom CellBorderStyle
	borderColorBottom *model.PdfColorDeviceRGB
	borderWidthBottom float64
	borderStyleRight  CellBorderStyle
	borderColorRight  *model.PdfColorDeviceRGB
	borderWidthRight  float64
	borderStyleTop    CellBorderStyle
	borderColorTop    *model.PdfColorDeviceRGB
	borderWidthTop    float64

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

// NewCell makes a new cell and inserts it into the table at the current position.
func (table *Table) NewCell() *TableCell {
	return table.newCell(1)
}

// MultiColCell makes a new cell with the specified column span and inserts it
// into the table at the current position.
func (table *Table) MultiColCell(colspan int) *TableCell {
	return table.newCell(colspan)
}

func (table *Table) newCell(colspan int) *TableCell {
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
	cell.rowspan = 1

	// Default left indent
	cell.indent = 5

	cell.borderStyleLeft = CellBorderStyleNone
	cell.borderLineStyle = draw.LineStyleSolid

	// Alignment defaults.
	cell.horizontalAlignment = CellHorizontalAlignmentLeft
	cell.verticalAlignment = CellVerticalAlignmentTop

	cell.borderWidthLeft = 0
	cell.borderWidthBottom = 0
	cell.borderWidthRight = 0
	cell.borderWidthTop = 0

	col := ColorBlack
	cell.borderColorLeft = model.NewPdfColorDeviceRGB(col.ToRGB())
	cell.borderColorBottom = model.NewPdfColorDeviceRGB(col.ToRGB())
	cell.borderColorRight = model.NewPdfColorDeviceRGB(col.ToRGB())
	cell.borderColorTop = model.NewPdfColorDeviceRGB(col.ToRGB())

	// Set column span.
	if colspan < 1 {
		common.Log.Debug("Table: cell colspan less than 1 (%d). Setting cell colspan to 1.", colspan)
		colspan = 1
	}

	remainingCols := table.cols - (cell.col - 1)
	if colspan > remainingCols {
		common.Log.Debug("Table: cell colspan (%d) exceeds remaining row cols (%d). Adjusting colspan.", colspan, remainingCols)
		colspan = remainingCols
	}
	cell.colspan = colspan
	table.curCell += colspan - 1

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
func (cell *TableCell) SetBorder(side CellBorderSide, style CellBorderStyle, width float64) {
	if style == CellBorderStyleSingle && side == CellBorderSideAll {
		cell.borderStyleLeft = CellBorderStyleSingle
		cell.borderWidthLeft = width
		cell.borderStyleBottom = CellBorderStyleSingle
		cell.borderWidthBottom = width
		cell.borderStyleRight = CellBorderStyleSingle
		cell.borderWidthRight = width
		cell.borderStyleTop = CellBorderStyleSingle
		cell.borderWidthTop = width
	} else if style == CellBorderStyleDouble && side == CellBorderSideAll {
		cell.borderStyleLeft = CellBorderStyleDouble
		cell.borderWidthLeft = width
		cell.borderStyleBottom = CellBorderStyleDouble
		cell.borderWidthBottom = width
		cell.borderStyleRight = CellBorderStyleDouble
		cell.borderWidthRight = width
		cell.borderStyleTop = CellBorderStyleDouble
		cell.borderWidthTop = width
	} else if (style == CellBorderStyleSingle || style == CellBorderStyleDouble) && side == CellBorderSideLeft {
		cell.borderStyleLeft = style
		cell.borderWidthLeft = width
	} else if (style == CellBorderStyleSingle || style == CellBorderStyleDouble) && side == CellBorderSideBottom {
		cell.borderStyleBottom = style
		cell.borderWidthBottom = width
	} else if (style == CellBorderStyleSingle || style == CellBorderStyleDouble) && side == CellBorderSideRight {
		cell.borderStyleRight = style
		cell.borderWidthRight = width
	} else if (style == CellBorderStyleSingle || style == CellBorderStyleDouble) && side == CellBorderSideTop {
		cell.borderStyleTop = style
		cell.borderWidthTop = width
	}
}

// SetBorderColor sets the cell's border color.
func (cell *TableCell) SetBorderColor(col Color) {
	cell.borderColorLeft = model.NewPdfColorDeviceRGB(col.ToRGB())
	cell.borderColorBottom = model.NewPdfColorDeviceRGB(col.ToRGB())
	cell.borderColorRight = model.NewPdfColorDeviceRGB(col.ToRGB())
	cell.borderColorTop = model.NewPdfColorDeviceRGB(col.ToRGB())
}

// SetBorderLineStyle sets border style (currently dashed or plain).
func (cell *TableCell) SetBorderLineStyle(style draw.LineStyle) {
	cell.borderLineStyle = style
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
// The currently supported VectorDrawable is: *Paragraph, *StyledParagraph.
func (cell *TableCell) SetContent(vd VectorDrawable) error {
	switch t := vd.(type) {
	case *Paragraph:
		if t.defaultWrap {
			// Enable wrapping by default.
			t.enableWrap = true
		}

		cell.content = vd
	case *StyledParagraph:
		if t.defaultWrap {
			// Enable wrapping by default.
			t.enableWrap = true
		}

		cell.content = vd
	case *Image:
		cell.content = vd
	case *Table:
		cell.content = vd
	case *List:
		cell.content = vd
	case *Division:
		cell.content = vd
	default:
		common.Log.Debug("ERROR: unsupported cell content type %T", vd)
		return core.ErrTypeError
	}

	return nil
}
