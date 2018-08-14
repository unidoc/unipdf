/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func TestTableMultiParagraphWrapped(t *testing.T) {
	c := New()

	pageHistoryTable := NewTable(4)
	pageHistoryTable.SetColumnWidths(0.1, 0.6, 0.15, 0.15)
	content := [][]string{
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
	}
	fontHelvetica := fonts.NewFontHelvetica()
	for _, rows := range content {
		for _, txt := range rows {
			p := NewParagraph(txt)
			p.SetFontSize(12)
			p.SetFont(fontHelvetica)
			p.SetColor(ColorBlack)
			p.SetEnableWrap(true)

			cell := pageHistoryTable.NewCell()
			cell.SetBorder(CellBorderStyleBox, 1)
			cell.SetContent(p)
		}
	}

	err := c.Draw(pageHistoryTable)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile("/tmp/table_pagehist.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableWithImage(t *testing.T) {
	c := New()

	pageHistoryTable := NewTable(4)
	pageHistoryTable.SetColumnWidths(0.1, 0.6, 0.15, 0.15)
	content := [][]string{
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
		{"1", "FullText Search Highlight the Term in Results \n\nissues 60", "120", "130"},
	}
	fontHelvetica := fonts.NewFontHelvetica()
	for _, rows := range content {
		for _, txt := range rows {
			p := NewParagraph(txt)
			p.SetFontSize(12)
			p.SetFont(fontHelvetica)
			p.SetColor(ColorBlack)
			p.SetEnableWrap(true)

			cell := pageHistoryTable.NewCell()
			cell.SetBorder(CellBorderStyleBox, 1)
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
	img, err := NewImageFromData(imgData)
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
	cell.SetBorder(CellBorderStyleBox, 1)

	err = c.Draw(pageHistoryTable)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile("/tmp/table_pagehist_with_img.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableWithDiv(t *testing.T) {
	c := New()

	pageHistoryTable := NewTable(4)
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
	fontHelvetica := fonts.NewFontHelvetica()
	fontHelveticaBold := fonts.NewFontHelveticaBold()
	for _, rows := range content {
		for colIdx, txt := range rows {
			p := NewParagraph(txt)
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
			div := NewDivision()

			if len(headings[colIdx]) > 0 {
				heading := NewParagraph(headings[colIdx])
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
			cell.SetBorder(CellBorderStyleBox, 1)
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
	img, err := NewImageFromData(imgData)
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
	cell.SetBorder(CellBorderStyleBox, 1)

	err = c.Draw(pageHistoryTable)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile("/tmp/table_pagehist_with_div.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
