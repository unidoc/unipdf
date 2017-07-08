/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// This test file contains multiple tests to generate PDFs. The outputs are written into /tmp as files.  The files
// themselves need to be observed to check for correctness as we don't have a good way to automatically check
// if every detail is correct.

import (
	"fmt"
	goimage "image"
	"io/ioutil"
	"math"
	"testing"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

const testPdfFile1 = "../../testfiles/minimal.pdf"
const testPdfLoremIpsumFile = "../../testfiles/lorem.pdf"
const testPdfTemplatesFile1 = "../../testfiles/templates1.pdf"
const testImageFile1 = "../../testfiles/logo.png"
const testImageFile2 = "../../testfiles/signature.png"

func TestTemplate1(t *testing.T) {
	creator := New()

	pages, err := loadPagesFromFile(testPdfFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	template, err := NewBlockFromPage(pages[0])
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	template.SetPos(0, 0)
	creator.Draw(template)

	template.SetAngle(45)
	creator.Draw(template)

	template.Scale(0.5, 0.5)
	creator.Draw(template)

	template.Scale(4, 4)
	creator.Draw(template)

	template.SetAngle(90)
	template.SetPos(100, 200)
	creator.Draw(template)

	creator.WriteToFile("/tmp/template_1.pdf")

	return
}

func TestImage1(t *testing.T) {
	creator := New()

	imgData, err := ioutil.ReadFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	img, err := NewImage(imgData)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	img.SetPos(0, 100)
	img.ScaleToWidth(1.0 * creator.Width())

	err = creator.Draw(img)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = creator.WriteToFile("/tmp/1.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestImageWrapping(t *testing.T) {
	creator := New()

	imgData, err := ioutil.ReadFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	img, err := NewImage(imgData)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	for j := 0; j < 40; j++ {
		img.ScaleToWidth(100 + 10*float64(j+1))

		err = creator.Draw(img)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err = creator.WriteToFile("/tmp/1_wrap.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestImageRotation(t *testing.T) {
	creator := New()

	imgData, err := ioutil.ReadFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	img, err := NewImage(imgData)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	//creator.MoveTo(0, 0)
	img.ScaleToWidth(100)

	angles := []float64{0, 90, 180, 270}
	for _, angle := range angles {
		creator.NewPage()

		creator.MoveTo(100, 100)

		img.SetAngle(angle)
		err = creator.Draw(img)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err = creator.WriteToFile("/tmp/1_rotate.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestImageRotationAndWrap(t *testing.T) {
	creator := New()

	imgData, err := ioutil.ReadFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	img, err := NewImage(imgData)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	img.ScaleToWidth(100)

	creator.NewPage()

	creator.MoveTo(100, 100)

	angles := []float64{0, 90, 180, 270, 0, 90, 180, 270}
	//angles := []float64{0, 0, 45, 90}

	for _, angle := range angles {
		img.SetAngle(angle)

		err = creator.Draw(img)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err = creator.WriteToFile("/tmp/rotate_2.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestParagraph1(t *testing.T) {
	creator := New()

	p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")

	err := creator.Draw(p)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = creator.WriteToFile("/tmp/2_p1.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestParagraphWrapping(t *testing.T) {
	creator := New()

	p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")

	p.SetMargins(0, 0, 5, 0)

	alignments := []TextAlignment{TextAlignmentLeft, TextAlignmentJustify, TextAlignmentCenter, TextAlignmentRight}
	for j := 0; j < 25; j++ {
		//p.SetAlignment(alignments[j%4])
		p.SetTextAlignment(alignments[1])

		err := creator.Draw(p)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err := creator.WriteToFile("/tmp/2_pwrap.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestParagraphFonts(t *testing.T) {
	creator := New()

	verdana, err := model.NewPdfFontFromTTFFile("/Library/Fonts/Verdana.ttf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	arialBold, err := model.NewPdfFontFromTTFFile("/Library/Fonts/Arial Bold.ttf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	fonts := []fonts.Font{verdana, arialBold, fonts.NewFontHelvetica(), verdana, arialBold, fonts.NewFontHelvetica()}
	for _, font := range fonts {
		p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
			"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
			"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
			"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
			"mollit anim id est laborum.")
		p.SetFont(font)
		p.SetFontSize(14)
		//p.SetWrapWidth(0.8 * creator.Width())
		p.SetTextAlignment(TextAlignmentJustify)
		p.SetLineHeight(1.2)
		p.SetMargins(0, 0, 5, 0)

		err = creator.Draw(p)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err = creator.WriteToFile("/tmp/2_pArial.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestParagraphStandardFonts(t *testing.T) {
	creator := New()

	names := []string{
		"Courier",
		"Courier-Bold",
		"Courier-BoldOblique",
		"Courier-Oblique",
		"Helvetica",
		"Helvetica-Bold",
		"Helvetica-BoldOblique",
		"Helvetica-Oblique",
		"Times-Bold",
		"Times-BoldItalic",
		"Times-Italic",
		"Times-Roman",
		"Symbol",
		"ZapfDingbats",
	}
	fonts := []fonts.Font{
		fonts.NewFontCourier(),
		fonts.NewFontCourierBold(),
		fonts.NewFontCourierBoldOblique(),
		fonts.NewFontCourierOblique(),
		fonts.NewFontHelvetica(),
		fonts.NewFontHelveticaBold(),
		fonts.NewFontHelveticaBoldOblique(),
		fonts.NewFontHelveticaOblique(),
		fonts.NewFontTimesBold(),
		fonts.NewFontTimesBoldItalic(),
		fonts.NewFontTimesItalic(),
		fonts.NewFontTimesRoman(),
		fonts.NewFontSymbol(),
		fonts.NewFontZapfDingbats(),
	}

	for idx, font := range fonts {
		p := NewParagraph(names[idx] + ": Lorem ipsum dolor sit amet, consectetur adipiscing elit...")
		p.SetFont(font)
		p.SetFontSize(12)
		p.SetLineHeight(1.2)
		p.SetMargins(0, 0, 5, 0)

		err := creator.Draw(p)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err := creator.WriteToFile("/tmp/2_standard14fonts.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestChapter(t *testing.T) {
	c := New()

	ch1 := c.NewChapter("Introduction")

	//subCh1 := NewSubchapter(ch1, "Workflow")

	p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt " +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore " +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")
	p.SetMargins(0, 0, 10, 0)

	for j := 0; j < 55; j++ {
		ch1.Add(p) // Can add any drawable..
	}

	c.Draw(ch1)

	err := c.WriteToFile("/tmp/3_chapters.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestSubchaptersSimple(t *testing.T) {
	c := New()

	ch1 := c.NewChapter("Introduction")
	subchap1 := c.NewSubchapter(ch1, "The fundamentals of the mastery of the most genious experiment of all times in modern world history. The story of the maker and the maker bot and the genius cow.")
	subchap1.SetMargins(0, 0, 5, 0)

	//subCh1 := NewSubchapter(ch1, "Workflow")

	p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt " +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore " +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")
	p.SetTextAlignment(TextAlignmentJustify)
	p.SetMargins(0, 0, 5, 0)
	for j := 0; j < 1; j++ {
		subchap1.Add(p)
	}

	subchap2 := c.NewSubchapter(ch1, "Mechanism")
	subchap2.SetMargins(0, 0, 5, 0)
	for j := 0; j < 1; j++ {
		subchap2.Add(p)
	}

	subchap3 := c.NewSubchapter(ch1, "Discussion")
	subchap3.SetMargins(0, 0, 5, 0)
	for j := 0; j < 1; j++ {
		subchap3.Add(p)
	}

	subchap4 := c.NewSubchapter(ch1, "Conclusion")
	subchap4.SetMargins(0, 0, 5, 0)
	for j := 0; j < 1; j++ {
		subchap4.Add(p)
	}
	c.Draw(ch1)

	ch2 := c.NewChapter("References")
	for j := 0; j < 13; j++ {
		ch2.Add(p)
	}
	c.Draw(ch2)

	// Set a function to create the front Page.
	c.CreateFrontPage(func(pageNum int, numPages int) {
		p := NewParagraph("Example Report")
		p.SetWidth(c.Width())
		p.SetTextAlignment(TextAlignmentCenter)
		p.SetFontSize(32)
		p.SetPos(0, 300)
		c.Draw(p)

		p.SetFontSize(22)
		p.SetText("Example Report Data Results")
		p.SetPos(0, 340)
		c.Draw(p)
	})

	// Set a function to create the table of contents.
	// Should be able to wrap..
	c.CreateTableOfContents(func(toc *TableOfContents) (*Chapter, error) {
		ch := c.NewChapter("Table of contents")
		ch.GetHeading().SetColor(0.5, 0.5, 0.5)
		ch.GetHeading().SetFontSize(28)
		ch.GetHeading().SetMargins(0, 0, 0, 30)

		numRows := len(toc.entries)
		table := NewTable(numRows, 2) // Nx3 table
		// Default, equal column sizes (4x0.25)...
		table.SetColumnWidths(0.9, 0.1)

		row := 1
		for _, entry := range toc.entries {
			// Col 1. Chapter number, title.
			var str string
			if entry.Subchapter == 0 {
				str = fmt.Sprintf("%d. %s", entry.Chapter, entry.Title)
			} else {
				str = fmt.Sprintf("        %d.%d. %s", entry.Chapter, entry.Subchapter, entry.Title)
			}
			p := NewParagraph(str)
			p.SetFontSize(14)
			cell := table.NewCell(row, 1)
			cell.SetContent(p)
			// Set the width so the height can be determined correctly.
			p.SetWidth(c.Width() - c.pageMargins.left - c.pageMargins.right)
			table.SetRowHeight(row, p.Height()+5)

			// Col 1. Page number.
			p = NewParagraph(fmt.Sprintf("%d", entry.PageNumber))
			p.SetFontSize(14)
			cell = table.NewCell(row, 2)
			cell.SetContent(p)

			row++
		}
		err := ch.Add(table)
		if err != nil {
			fmt.Printf("Error adding table: %v\n", err)
			return nil, err
		}

		return ch, nil
	})

	err := c.WriteToFile("/tmp/3_subchapters_simple.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestSubchapters(t *testing.T) {
	c := New()

	ch1 := c.NewChapter("Introduction")
	subchap1 := c.NewSubchapter(ch1, "The fundamentals")
	subchap1.SetMargins(0, 0, 5, 0)

	//subCh1 := NewSubchapter(ch1, "Workflow")

	p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt " +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore " +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")
	p.SetTextAlignment(TextAlignmentJustify)
	p.SetMargins(0, 0, 5, 0)
	for j := 0; j < 18; j++ {
		subchap1.Add(p)
	}

	subchap2 := c.NewSubchapter(ch1, "Mechanism")
	subchap2.SetMargins(0, 0, 5, 0)
	for j := 0; j < 15; j++ {
		subchap2.Add(p)
	}

	subchap3 := c.NewSubchapter(ch1, "Discussion")
	subchap3.SetMargins(0, 0, 5, 0)
	for j := 0; j < 19; j++ {
		subchap3.Add(p)
	}

	subchap4 := c.NewSubchapter(ch1, "Conclusion")
	subchap4.SetMargins(0, 0, 5, 0)
	for j := 0; j < 23; j++ {
		subchap4.Add(p)
	}

	c.Draw(ch1)

	for i := 0; i < 50; i++ {
		ch2 := c.NewChapter("References")
		for j := 0; j < 13; j++ {
			ch2.Add(p)
		}

		c.Draw(ch2)
	}

	// Set a function to create the front Page.
	c.CreateFrontPage(func(pageNum int, numPages int) {
		p := NewParagraph("Example Report")
		p.SetWidth(c.Width())
		p.SetTextAlignment(TextAlignmentCenter)
		p.SetFontSize(32)
		p.SetPos(0, 300)
		c.Draw(p)

		p.SetFontSize(22)
		p.SetText("Example Report Data Results")
		p.SetPos(0, 340)
		c.Draw(p)
	})

	// Set a function to create the table of contents.
	c.CreateTableOfContents(func(toc *TableOfContents) (*Chapter, error) {
		ch := c.NewChapter("Table of contents")
		ch.GetHeading().SetColor(0.5, 0.5, 0.5)
		ch.GetHeading().SetFontSize(28)
		ch.GetHeading().SetMargins(0, 0, 0, 30)

		numRows := len(toc.entries)
		table := NewTable(numRows, 2) // Nx3 table
		// Default, equal column sizes (4x0.25)...
		table.SetColumnWidths(0.9, 0.1)

		row := 1
		for _, entry := range toc.entries {
			// Col 1. Chapter number, title.
			var str string
			if entry.Subchapter == 0 {
				str = fmt.Sprintf("%d. %s", entry.Chapter, entry.Title)
			} else {
				str = fmt.Sprintf("        %d.%d. %s", entry.Chapter, entry.Subchapter, entry.Title)
			}
			p := NewParagraph(str)
			p.SetFontSize(14)
			cell := table.NewCell(row, 1)
			cell.SetContent(p)
			// Set the width so the height can be determined correctly.
			p.SetWidth(c.Width() - c.pageMargins.left - c.pageMargins.right)
			table.SetRowHeight(row, p.Height()+5)

			// Col 1. Page number.
			p = NewParagraph(fmt.Sprintf("%d", entry.PageNumber))
			p.SetFontSize(14)
			cell = table.NewCell(row, 2)
			cell.SetContent(p)

			row++
		}
		err := ch.Add(table)
		if err != nil {
			fmt.Printf("Error adding table: %v\n", err)
			return nil, err
		}

		return ch, nil
	})

	addHeadersAndFooters(c)

	err := c.WriteToFile("/tmp/3_subchapters.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestTable(t *testing.T) {
	table := NewTable(4, 4) // 4x4 table
	// Default, equal column sizes (4x0.25)...
	table.SetColumnWidths(0.5, 0.2, 0.2, 0.1)

	cell := table.NewCell(1, 1)
	p := NewParagraph("1,1")
	cell.SetContent(p)

	cell = table.NewCell(1, 2)
	p = NewParagraph("1,2")
	cell.SetContent(p)

	cell = table.NewCell(1, 3)
	p = NewParagraph("1,3")
	cell.SetContent(p)

	cell = table.NewCell(1, 4)
	p = NewParagraph("1,4")
	cell.SetContent(p)

	cell = table.NewCell(2, 1)
	p = NewParagraph("2,1")
	cell.SetContent(p)

	cell = table.NewCell(2, 2)
	p = NewParagraph("2,2")
	cell.SetContent(p)

	cell = table.NewCell(2, 4)
	p = NewParagraph("2,4")
	cell.SetContent(p)

	cell = table.NewCell(4, 4)
	p = NewParagraph("4,4")
	cell.SetContent(p)

	c := New()
	c.Draw(table)

	err := c.WriteToFile("/tmp/4_table.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func addHeadersAndFooters(c *Creator) {
	c.DrawHeader(func(pageNum int, totPages int) {
		/*
			if pageNum == 1 {
				// Skip on front Page.
				return
			}
		*/

		// Add Page number
		p := NewParagraph(fmt.Sprintf("Page %d / %d", pageNum, totPages))
		p.SetPos(0.8*c.pageWidth, 20)
		c.Draw(p)

		// Draw on the template...
		img, err := NewImageFromFile(testImageFile1)
		if err != nil {
			fmt.Printf("ERROR : %v\n", err)
		}
		img.ScaleToHeight(0.4 * c.pageMargins.top)
		img.SetPos(20, 10)

		c.Draw(img)
	})

	c.DrawFooter(func(pageNum int, totPages int) {
		/*
			if pageNum == 1 {
				// Skip on front Page.
				return
			}
		*/

		// Add company name.
		companyName := "Company inc."
		p := NewParagraph(companyName)
		p.SetPos(0.1*c.pageWidth, c.pageHeight-c.pageMargins.bottom+10)
		c.Draw(p)

		p = NewParagraph("July 2017")
		p.SetPos(0.8*c.pageWidth, c.pageHeight-c.pageMargins.bottom+10)
		c.Draw(p)
	})
}

func TestHeadersAndFooters(t *testing.T) {
	c := New()

	ch1 := c.NewChapter("Introduction")

	p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt " +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore " +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")
	p.SetMargins(0, 0, 10, 0)

	for j := 0; j < 55; j++ {
		ch1.Add(p) // Can add any drawable..
	}

	c.Draw(ch1)

	// Make unidoc headers and footers.
	addHeadersAndFooters(c)

	err := c.WriteToFile("/tmp/4_headers.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func makeQrCodeImage(text string, width float64, oversampling int) (goimage.Image, error) {
	qrCode, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return nil, err
	}

	pixelWidth := 5 * int(math.Ceil(width))
	qrCode, err = barcode.Scale(qrCode, pixelWidth, pixelWidth)
	if err != nil {
		return nil, err
	}

	return qrCode, nil
}

func TestQRCodeOnNewPage(t *testing.T) {
	creator := New()

	qrCode, err := makeQrCodeImage("HELLO", 40, 5)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	img, err := NewImageFromGoImage(qrCode)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	// Sets positioning to absolute coordinates.
	img.SetWidth(40)
	img.SetHeight(40)

	for i := 0; i < 5; i++ {
		img.SetPos(100.0*float64(i), 100.0*float64(i))
		creator.Draw(img)
	}

	creator.WriteToFile("/tmp/3_barcode_qr_newpage.pdf")
}

// Example of using a template Page, generating and applying QR
func TestQRCodeOnTemplate(t *testing.T) {
	pages, err := loadPagesFromFile(testPdfTemplatesFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	if len(pages) < 2 {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Load Page 1 as template.
	tpl, err := NewBlockFromPage(pages[1])
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	tpl.SetPos(0, 0)

	// Generate QR code.
	qrCode, err := makeQrCodeImage("HELLO", 50, 5)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Prepare content image.
	image, err := NewImageFromGoImage(qrCode)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	image.SetWidth(50)
	image.SetHeight(50)
	image.SetPos(480, 100)

	tpl.Draw(image)

	creator := New()
	creator.NewPage()
	creator.Draw(tpl)

	// Add another Page where the template has been rotated.
	creator.NewPage()
	tpl.SetAngle(90)
	tpl.SetPos(-50, 750)

	creator.Draw(tpl)

	// Add another Page where the template is rotated 90 degrees.
	loremPages, err := loadPagesFromFile(testPdfLoremIpsumFile)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	if len(loremPages) != 1 {
		t.Errorf("Pages != 1")
		return
	}

	// Add another Page where another Page is embedded on the Page.  The other Page is scaled and shifted to fit
	// on the right of the template.
	loremTpl, err := NewBlockFromPage(loremPages[0])
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	loremTpl.ScaleToWidth(0.8 * creator.Width())
	loremTpl.SetPos(100, 100)

	creator.Draw(loremTpl)

	// Write the example to file.
	creator.WriteToFile("/tmp/4_barcode_on_tpl.pdf")
}
