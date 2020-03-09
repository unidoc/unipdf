/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

var (
	fontHelvetica     = model.NewStandard14FontMustCompile(model.HelveticaName)
	fontHelveticaBold = model.NewStandard14FontMustCompile(model.HelveticaBoldName)
)

func TestTableMultiParagraphWrapped(t *testing.T) {
	c := New()

	pageHistoryTable := c.NewTable(4)
	pageHistoryTable.SetColumnWidths(0.1, 0.6, 0.15, 0.15)
	content := [][]string{
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
	}

	for _, rows := range content {
		for _, txt := range rows {
			p := c.NewParagraph(txt)
			p.SetFontSize(12)
			p.SetFont(fontHelvetica)
			p.SetColor(ColorBlack)
			p.SetEnableWrap(true)

			cell := pageHistoryTable.NewCell()
			cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
			cell.SetContent(p)
		}
	}

	err := c.Draw(pageHistoryTable)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile(tempFile("table_pagehist.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableWithImage(t *testing.T) {
	c := New()

	pageHistoryTable := c.NewTable(4)
	pageHistoryTable.SetColumnWidths(0.1, 0.6, 0.15, 0.15)
	content := [][]string{
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
	}
	for _, rows := range content {
		for _, txt := range rows {
			p := c.NewParagraph(txt)
			p.SetFontSize(12)
			p.SetFont(fontHelvetica)
			p.SetColor(ColorBlack)
			p.SetEnableWrap(true)

			cell := pageHistoryTable.NewCell()
			cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
			err := cell.SetContent(p)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
		}
	}

	pageHistoryTable.SkipCells(1)

	// Add image.
	imgData, err := ioutil.ReadFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	img, err := c.NewImageFromData(imgData)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	img.margins.top = 2.0
	img.margins.bottom = 2.0
	img.margins.left = 2.0
	img.margins.bottom = 2.0
	img.ScaleToWidth(0.3 * c.Width())
	fmt.Printf("Scaling image to width: %v\n", 0.5*c.Width())

	cell := pageHistoryTable.NewCell()
	cell.SetContent(img)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	err = c.Draw(pageHistoryTable)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile(tempFile("table_pagehist_with_img.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableWithDiv(t *testing.T) {
	c := New()

	pageHistoryTable := c.NewTable(4)
	pageHistoryTable.SetColumnWidths(0.1, 0.6, 0.15, 0.15)

	headings := []string{
		"", "Description", "Passing", "Total",
	}
	content := [][]string{
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"2", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"3", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"3", "FullText Search Highlight the Term in Results. Going hunting in the winter can be fruitful, especially if it has not been too cold and the deer are well fed. \n\nissues 60", "120 90 30", "130 1"},
		{"4", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"5", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"6", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130 a b c d e f g"},
		{"7", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"8", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130 gogogoggogoogogo"},
		{"9", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"10", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"11", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"12", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"13", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"14", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"15", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"16", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"17", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"18", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"19", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"20", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
	}
	for _, rows := range content {
		for colIdx, txt := range rows {
			p := c.NewParagraph(txt)
			p.SetFontSize(12)
			p.SetFont(fontHelvetica)
			p.SetColor(ColorBlack)
			p.SetMargins(0, 5, 10.0, 10.0)
			if len(txt) > 10 {
				p.SetTextAlignment(TextAlignmentJustify)
			} else {
				p.SetTextAlignment(TextAlignmentCenter)
			}

			// Place cell contents (header and text) inside a div.
			div := c.NewDivision()

			if len(headings[colIdx]) > 0 {
				heading := c.NewParagraph(headings[colIdx])
				heading.SetFontSize(14)
				heading.SetFont(fontHelveticaBold)
				heading.SetColor(ColorRed)
				heading.SetTextAlignment(TextAlignmentCenter)
				err := div.Add(heading)
				if err != nil {
					t.Fatalf("Error: %v", err)
				}
			}

			err := div.Add(p)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}

			cell := pageHistoryTable.NewCell()
			cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
			err = cell.SetContent(div)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
		}
	}

	pageHistoryTable.SkipCells(1)

	// Add image.
	imgData, err := ioutil.ReadFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	img, err := c.NewImageFromData(imgData)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	img.margins.top = 2.0
	img.margins.bottom = 2.0
	img.margins.left = 2.0
	img.margins.bottom = 2.0
	img.ScaleToWidth(0.2 * c.Width())
	fmt.Printf("Scaling image to width: %v\n", 0.5*c.Width())

	cell := pageHistoryTable.NewCell()
	cell.SetContent(img)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	err = c.Draw(pageHistoryTable)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile(tempFile("table_pagehist_with_div.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableColSpan(t *testing.T) {
	c := New()

	table := c.NewTable(4)
	table.SetColumnWidths(0.25, 0.25, 0.25, 0.25)

	p := c.NewStyledParagraph()
	p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")

	// Colspan 1 + 1 + 1 + 1
	cell := table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	// Colspan 2 + 2
	cell = table.MultiColCell(2)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	cell = table.MultiColCell(2)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	// Colspan 3 + 1
	cell = table.MultiColCell(3)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	// Colspan 4
	cell = table.MultiColCell(4)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	// Invalid colspans -1 and 5. Will be adjusted to 1 and 3.
	cell = table.MultiColCell(-1)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	cell = table.MultiColCell(5)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	err := c.Draw(table)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile(tempFile("table_col_span.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableHeaderTest(t *testing.T) {
	c := New()

	p := c.NewStyledParagraph()
	p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")

	table := c.NewTable(4)
	table.SetColumnWidths(0.25, 0.25, 0.25, 0.25)
	table.SetHeaderRows(1, 3)

	// Add header
	for i := 0; i < 3; i++ {
		p := c.NewParagraph(fmt.Sprintf("Table header %d", i+1))
		p.SetColor(ColorRGBFrom8bit(
			byte((i+1)*50), byte((i+1)*50), byte((i+1)*50),
		))

		cell := table.MultiColCell(1)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(p)
		cell.SetBackgroundColor(ColorRGBFrom8bit(
			byte(100*(i+1)), byte(100*(i+1)), byte(100*(i+1)),
		))

		for j := 0; j < 3; j++ {
			p := c.NewParagraph(fmt.Sprintf("Header column %d-%d", i+1, j+1))
			p.SetColor(ColorRGBFrom8bit(255, 255, 255))

			cell = table.MultiColCell(1)
			cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
			cell.SetContent(p)
			cell.SetBackgroundColor(ColorRGBFrom8bit(
				byte(100*(i+1)), byte(50*(j+1)), byte(50*(i+j+1)),
			))
		}
	}

	// Add content
	for i := 0; i < 50; i++ {
		j := i * 4

		// Colspan 4
		cell := table.MultiColCell(4)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 1", j+1)))

		// Colspan 1 + 1 + 1 + 1
		cell = table.NewCell()
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 1", j+2)))

		cell = table.NewCell()
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 2", j+2)))

		cell = table.NewCell()
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 3", j+2)))

		cell = table.NewCell()
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 4", j+2)))

		// Colspan 2 + 2
		cell = table.MultiColCell(2)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 1", j+3)))

		cell = table.MultiColCell(2)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 2", j+3)))

		// Colspan 3 + 1
		cell = table.MultiColCell(3)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 1", j+4)))

		cell = table.NewCell()
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetContent(c.NewParagraph(fmt.Sprintf("Line %d - Col 2", j+4)))

		if i > 0 && i%5 == 0 {
			cell := table.MultiColCell(4)
			cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
			cell.SetContent(p)
		}
	}

	err := c.Draw(table)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile(tempFile("table_headers.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableSubtables(t *testing.T) {
	c := New()
	headerColor := ColorRGBFrom8bit(255, 255, 0)
	footerColor := ColorRGBFrom8bit(0, 255, 0)

	generateSubtable := func(rows, cols, index int, rightBorder bool) *Table {
		subtable := c.NewTable(cols)

		// Add header row.
		sp := c.NewStyledParagraph()
		sp.Append(fmt.Sprintf("Header of subtable %d", index))

		cell := subtable.MultiColCell(cols)
		cell.SetContent(sp)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetHorizontalAlignment(CellHorizontalAlignmentCenter)
		cell.SetVerticalAlignment(CellVerticalAlignmentMiddle)
		cell.SetBackgroundColor(headerColor)

		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				sp = c.NewStyledParagraph()
				sp.Append(fmt.Sprintf("%d-%d", i+1, j+1))
				cell = subtable.NewCell()
				cell.SetContent(sp)

				if j == 0 {
					cell.SetBorder(CellBorderSideLeft, CellBorderStyleSingle, 1)
				}
				if rightBorder && j == cols-1 {
					cell.SetBorder(CellBorderSideRight, CellBorderStyleSingle, 1)
				}
			}
		}

		// Add footer row.
		sp = c.NewStyledParagraph()
		sp.Append(fmt.Sprintf("Footer of subtable %d", index))

		cell = subtable.MultiColCell(cols)
		cell.SetContent(sp)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetHorizontalAlignment(CellHorizontalAlignmentCenter)
		cell.SetVerticalAlignment(CellVerticalAlignmentMiddle)
		cell.SetBackgroundColor(footerColor)

		subtable.SetRowHeight(1, 30)
		subtable.SetRowHeight(subtable.Rows(), 40)
		return subtable
	}

	table := c.NewTable(6)

	// Test number of rows and columns.
	if rows := table.Rows(); rows != 0 {
		t.Errorf("Table error: expected 0 rows. Got %d rows", rows)
	}
	if cols := table.Cols(); cols != 6 {
		t.Errorf("Table error: expected 6 cols. Got %d cols", cols)
	}

	// Add subtable 1 on row 1, col 1 (4x4)
	table.AddSubtable(1, 1, generateSubtable(4, 4, 1, false))

	// Add subtable 2 on row 1, col 5 (4x4)
	// Table will be expanded to 8 columns because the subtable does not fit.
	table.AddSubtable(1, 5, generateSubtable(4, 4, 2, true))

	// Add subtable 3 on row 7, col 1 (4x4)
	table.AddSubtable(7, 1, generateSubtable(4, 4, 3, false))

	// Add subtable 4 on row 7, col 5 (4x4)
	table.AddSubtable(7, 5, generateSubtable(4, 4, 4, true))

	// Add subtable 5 on row 13, col 3 (4x4)
	table.AddSubtable(13, 3, generateSubtable(4, 4, 5, true))

	// Add subtable 6 on row 13, col 1 (3x2)
	table.AddSubtable(13, 1, generateSubtable(3, 2, 6, false))

	// Add subtable 7 on row 13, col 7 (3x2)
	table.AddSubtable(13, 7, generateSubtable(3, 2, 7, true))

	// Add subtable 8 on row 18, col 1 (3x2)
	table.AddSubtable(18, 1, generateSubtable(3, 2, 8, false))

	// Add subtable 9 on row 19, col 3 (2x4)
	table.AddSubtable(19, 3, generateSubtable(2, 4, 9, true))

	// Add subtable 10 on row 18, col 7 (3x2)
	table.AddSubtable(18, 7, generateSubtable(3, 2, 10, true))

	// Test number of rows and columns.
	if rows := table.Rows(); rows != 22 {
		t.Errorf("Table error: expected 22 rows. Got %d rows", rows)
	}
	if cols := table.Cols(); cols != 8 {
		t.Errorf("Table error: expected 8 cols. Got %d cols", cols)
	}

	// Test row height
	if rowHeight, err := table.GetRowHeight(1); err != nil || rowHeight != 30.0 {
		t.Errorf("Table error: expected row height 30.0. Got %f", rowHeight)
	}
	if rowHeight, err := table.GetRowHeight(18); err != nil || rowHeight != 40.0 {
		t.Errorf("Table error: expected row height 40.0. Got %f", rowHeight)
	}

	err := c.Draw(table)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile(tempFile("table_add_subtables.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableParagraphLinks(t *testing.T) {
	c := New()

	// First page.
	c.NewPage()
	table := c.NewTable(2)

	// Add internal link cells.
	cell := table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	p := c.NewStyledParagraph()
	p.Append("Internal link")
	cell.SetContent(p)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	p = c.NewStyledParagraph()
	p.AddInternalLink("link to second page", 2, 0, 0, 0)
	cell.SetContent(p)

	// Add external link cells.
	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	p = c.NewStyledParagraph()
	p.Append("External link")
	cell.SetContent(p)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	p = c.NewStyledParagraph()
	p.AddExternalLink("link to UniPDF", "https://github.com/unidoc/unipdf")
	cell.SetContent(p)

	if err := c.Draw(table); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Second page.
	c.NewPage()

	p = c.NewStyledParagraph()
	p.Append("Page 2").Style.FontSize = 24

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	if err := c.WriteToFile(tempFile("table_paragraph_links.pdf")); err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
