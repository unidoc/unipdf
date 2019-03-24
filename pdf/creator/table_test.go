/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/unidoc/unidoc/pdf/model"
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

	err := c.Draw(table)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile(tempFile("table_col_span.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
