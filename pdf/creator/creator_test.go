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

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

const testPdfFile1 = "../../testfiles/minimal.pdf"
const testPdfLoremIpsumFile = "../../testfiles/lorem.pdf"
const testPdfTemplatesFile1 = "../../testfiles/templates1.pdf"
const testImageFile1 = "../../testfiles/logo.png"
const testImageFile2 = "../../testfiles/signature.png"
const testRobotoRegularTTFFile = "../../testfiles/roboto/Roboto-Regular.ttf"
const testRobotoBoldTTFFile = "../../testfiles/roboto/Roboto-Bold.ttf"

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

// TestImage1 tests loading an image and adding to file at an absolute position.
func TestImage1(t *testing.T) {
	creator := New()

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

// TestImageWithEncoder tests loading inserting an image with a specified encoder.
func TestImageWithEncoder(t *testing.T) {
	creator := New()

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

	// JPEG encoder (DCT) with quality factor 70.
	encoder := core.NewDCTEncoder()
	encoder.Quality = 70
	encoder.Width = int(img.Width())
	encoder.Height = int(img.Height())
	img.SetEncoder(encoder)

	img.SetPos(0, 100)
	img.ScaleToWidth(1.0 * creator.Width())

	err = creator.Draw(img)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = creator.WriteToFile("/tmp/1_dct.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestShapes1(t *testing.T) {
	creator := New()

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

	img.SetPos(0, 100)
	img.ScaleToWidth(1.0 * creator.Width())

	err = creator.Draw(img)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Add line.
	line := NewLine(0, 0, 100, 100)
	line.SetLineWidth(3.0)
	line.SetColor(ColorRGBFromHex("#ff0000"))
	err = creator.Draw(line)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Add rect with default params.
	rect := NewRectangle(100, 100, 100, 100)
	err = creator.Draw(rect)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Add rect with fill and large border
	rect = NewRectangle(100, 500, 100, 100)
	rect.SetBorderColor(ColorRGBFromHex("#00ff00")) // Green border
	rect.SetBorderWidth(15.0)
	rect.SetFillColor(ColorRGBFromHex("#0000ff")) // Blue fill
	err = creator.Draw(rect)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Draw a circle. (inscribed inside the previous rectangle).
	ell := NewEllipse(100, 100, 100, 100)
	err = creator.Draw(ell)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Draw a circle around upper right page corner.
	ell = NewEllipse(creator.Width(), 0, 100, 100)
	err = creator.Draw(ell)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Draw an ellipse with fill and border.
	ell = NewEllipse(500, 100, 100, 200)
	ell.SetFillColor(ColorRGBFromHex("#ccc")) // Gray fill
	ell.SetBorderWidth(10.0)
	err = creator.Draw(ell)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = creator.WriteToFile("/tmp/1_shapes.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// Example drawing image and line shape on a block and applying to pages, also demonstrating block rotation.
func TestShapesOnBlock(t *testing.T) {
	creator := New()

	block := NewBlock(creator.Width(), 200)

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

	img.SetPos(50, 75)
	img.ScaleToHeight(100.0)
	block.Draw(img)

	// Add line.
	line := NewLine(0, 180, creator.Width(), 180)
	line.SetLineWidth(10.0)
	line.SetColor(ColorRGBFromHex("#ff0000"))
	block.Draw(line)

	creator.NewPage()
	creator.MoveTo(0, 0)
	creator.Draw(block)

	creator.NewPage()
	creator.MoveTo(0, 200)
	creator.Draw(block)

	creator.NewPage()
	creator.MoveTo(0, 700)
	block.SetAngle(90)
	creator.Draw(block)

	err = creator.WriteToFile("/tmp/1_shapes_on_block.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// Test image wrapping between pages when using relative context mode.
func TestImageWrapping(t *testing.T) {
	creator := New()

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

// Test rotating image.  Rotating about upper left corner.
func TestImageRotation(t *testing.T) {
	creator := New()

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

// Test image, rotation and page wrapping.  Disadvantage here is that content is overlapping.  May be reconsidered
// in the future.  And actually reconsider overall how images are used in the relative context mode.
func TestImageRotationAndWrap(t *testing.T) {
	creator := New()

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

// Test basic paragraph with default font.
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

// Test paragraph and page and text wrapping with left, justify, center and right modes.
// TODO: In the future we would like the paragraph to split up between pages.  Split up on line, never allowing
// less than 2 lines to go over (common practice).
// TODO: In the future we would like to implement Donald Knuth's line wrapping algorithm or something similar.
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

func TestParagraphWrapping2(t *testing.T) {
	creator := New()

	p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")

	alignments := []TextAlignment{TextAlignmentLeft, TextAlignmentJustify, TextAlignmentCenter, TextAlignmentRight}
	for j := 0; j < 25; j++ {
		//p.SetAlignment(alignments[j%4])
		p.SetMargins(50, 50, 50, 50)
		p.SetTextAlignment(alignments[1])

		err := creator.Draw(p)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err := creator.WriteToFile("/tmp/2_pwrap2.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// Test writing with TTF fonts.
func TestParagraphFonts(t *testing.T) {
	creator := New()

	roboto, err := model.NewPdfFontFromTTFFile(testRobotoRegularTTFFile)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	robotoBold, err := model.NewPdfFontFromTTFFile(testRobotoBoldTTFFile)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	fonts := []fonts.Font{roboto, robotoBold, fonts.NewFontHelvetica(), roboto, robotoBold, fonts.NewFontHelvetica()}
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

// Test writing with the 14 built in fonts.
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
		"Times-Roman",
		"Times-Bold",
		"Times-BoldItalic",
		"Times-Italic",
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
		fonts.NewFontTimesRoman(),
		fonts.NewFontTimesBold(),
		fonts.NewFontTimesBoldItalic(),
		fonts.NewFontTimesItalic(),
		fonts.NewFontSymbol(),
		fonts.NewFontZapfDingbats(),
	}
	texts := []string{
		"Courier: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Courier-Bold: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Courier-BoldOblique: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Courier-Oblique: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",

		"Helvetica: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Helvetica-Bold: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Helvetica-BoldOblique: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Helvetica-Oblique: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",

		"Times-Roman: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Times-Bold: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Times-BoldItalic: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"Times-Italic: Lorem ipsum dolor sit amet, consectetur adipiscing elit...",
		"\u2206\u0393\u0020\u2192\u0020\u0030", // Delta Gamma space arrowright space zero (demonstrate Symbol font)
		"\u2702\u0020\u2709\u261e\u2711\u2714", // a2 (scissors) space a117 (mail) a12 (finger) a17 (pen) a20 (checkmark)
	}

	for idx, font := range fonts {
		p := NewParagraph(texts[idx])
		p.SetFont(font)
		p.SetFontSize(12)
		p.SetLineHeight(1.2)
		p.SetMargins(0, 0, 5, 0)

		if names[idx] == "Symbol" {
			// For Symbol font, need to use Symbol encoder.
			p.SetEncoder(textencoding.NewSymbolEncoder())
		} else if names[idx] == "ZapfDingbats" {
			// Font ZapfDingbats font, need to use ZapfDingbats encoder.
			p.SetEncoder(textencoding.NewZapfDingbatsEncoder())
		}

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

// Tests creating a chapter with paragraphs.
func TestChapter(t *testing.T) {
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

	err := c.WriteToFile("/tmp/3_chapters.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// Tests creating a chapter with paragraphs.
func TestChapterMargins(t *testing.T) {
	c := New()

	for j := 0; j < 20; j++ {
		ch := c.NewChapter(fmt.Sprintf("Chapter %d", j+1))
		if j < 5 {
			ch.SetMargins(3*float64(j), 3*float64(j), 5+float64(j), 0)
		}

		p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt " +
			"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
			"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore " +
			"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
			"mollit anim id est laborum.")
		p.SetTextAlignment(TextAlignmentJustify)
		ch.Add(p)
		c.Draw(ch)
	}

	err := c.WriteToFile("/tmp/3_chapters_margins.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// Test creating and drawing subchapters with text content.
// Also generates a front page, and a table of contents.
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
	c.CreateFrontPage(func(args FrontpageFunctionArgs) {
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
		ch.GetHeading().SetColor(ColorRGBFromArithmetic(0.5, 0.5, 0.5))
		ch.GetHeading().SetFontSize(28)
		ch.GetHeading().SetMargins(0, 0, 0, 30)

		table := NewTable(2) // 2 column table.
		// Default, equal column sizes (4x0.25)...
		table.SetColumnWidths(0.9, 0.1)

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
			cell := table.NewCell()
			cell.SetContent(p)
			// Set the paragraph width to the cell width.
			p.SetWidth(cell.Width(c.Context()))
			table.SetRowHeight(table.CurRow(), p.Height()*1.2)

			// Col 1. Page number.
			p = NewParagraph(fmt.Sprintf("%d", entry.PageNumber))
			p.SetFontSize(14)
			cell = table.NewCell()
			cell.SetContent(p)
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
	c.CreateFrontPage(func(args FrontpageFunctionArgs) {
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
		ch.GetHeading().SetColor(ColorRGBFromArithmetic(0.5, 0.5, 0.5))
		ch.GetHeading().SetFontSize(28)
		ch.GetHeading().SetMargins(0, 0, 0, 30)

		table := NewTable(2)
		// Default, equal column sizes (4x0.25)...
		table.SetColumnWidths(0.9, 0.1)

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
			cell := table.NewCell()
			cell.SetContent(p)
			// Set the paragraph width to the cell width.
			p.SetWidth(cell.Width(c.Context()))
			table.SetRowHeight(table.CurRow(), p.Height()*1.2)

			// Col 1. Page number.
			p = NewParagraph(fmt.Sprintf("%d", entry.PageNumber))
			p.SetFontSize(14)
			cell = table.NewCell()
			cell.SetContent(p)
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

// Test creating and drawing a table.
func TestTable(t *testing.T) {
	table := NewTable(4) // Mx4 table
	// Default, equal column sizes (4x0.25)...
	table.SetColumnWidths(0.5, 0.2, 0.2, 0.1)

	cell := table.NewCell()
	p := NewParagraph("1,1")
	cell.SetContent(p)

	cell = table.NewCell()
	p = NewParagraph("1,2")
	cell.SetContent(p)

	cell = table.NewCell()
	p = NewParagraph("1,3")
	cell.SetContent(p)

	cell = table.NewCell()
	p = NewParagraph("1,4")
	cell.SetContent(p)

	cell = table.NewCell()
	p = NewParagraph("2,1")
	cell.SetContent(p)

	cell = table.NewCell()
	p = NewParagraph("2,2")
	cell.SetContent(p)

	table.SkipCells(1) // Skip over 2,3.

	cell = table.NewCell()
	p = NewParagraph("2,4")
	cell.SetContent(p)

	// Skip over two rows.
	table.SkipRows(2)
	cell = table.NewCell()
	p = NewParagraph("4,4")
	cell.SetContent(p)

	// Move down 3 rows, 2 to the left.
	table.SkipOver(3, -2)
	cell = table.NewCell()
	p = NewParagraph("7,2")
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorRGBFrom8bit(255, 0, 0))

	c := New()
	c.Draw(table)

	err := c.WriteToFile("/tmp/4_table.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestTableCellWrapping(t *testing.T) {
	c := New()
	c.NewPage()

	table := NewTable(4) // Mx4 table
	// Default, equal column sizes (4x0.25)...
	table.SetColumnWidths(0.5, 0.2, 0.2, 0.1)

	cell := table.NewCell()
	p := NewParagraph("A Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)
	p.SetEnableWrap(true)
	p.SetWidth(cell.Width(c.Context()))
	p.SetTextAlignment(TextAlignmentJustify)

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	p = NewParagraph("B Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentRight)
	cell.SetContent(p)

	cell = table.NewCell()
	p = NewParagraph("C Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	p.SetEnableWrap(true)
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("1,4")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("2,1")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("2,2")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("2,2")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	//table.SkipCells(1) // Skip over 2,3.

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	//p = NewParagraph("D Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	p = NewParagraph("X")
	p.SetEnableWrap(true)
	cell.SetContent(p)

	// Skip over two rows.
	table.SkipRows(2)
	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	p = NewParagraph("4,4")
	cell.SetContent(p)

	// Move down 3 rows, 2 to the left.
	table.SkipOver(3, -2)
	cell = table.NewCell()
	p = NewParagraph("7,2")
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorRGBFrom8bit(255, 0, 0))

	table.SkipRows(1)
	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	p = NewParagraph("This is\nnewline\nwrapped\n\nmulti")
	p.SetEnableWrap(true)
	cell.SetContent(p)

	err := c.Draw(table)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile("/tmp/tablecell_wrap.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

// Test creating and drawing a table.
func TestBorderedTable(t *testing.T) {
	table := NewTable(4) // Mx4 table
	// Default, equal column sizes (4x0.25)...
	table.SetColumnWidths(0.5, 0.2, 0.2, 0.1)

	cell := table.NewCell()
	p := NewParagraph("1,1")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("1,2")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("1,3")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("1,4")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("2,1")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	cell = table.NewCell()
	p = NewParagraph("2,2")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	table.SkipCells(1) // Skip over 2,3.

	cell = table.NewCell()
	p = NewParagraph("2,4")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	// Skip over two rows.
	table.SkipRows(2)
	cell = table.NewCell()
	p = NewParagraph("4,4")
	cell.SetContent(p)
	cell.SetBorder(CellBorderStyleBox, 1)

	// Move down 3 rows, 2 to the left.
	table.SkipOver(3, -2)
	cell = table.NewCell()
	p = NewParagraph("7,2")
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorRGBFrom8bit(255, 0, 0))
	cell.SetBorder(CellBorderStyleBox, 1)

	c := New()
	c.Draw(table)

	err := c.WriteToFile("/tmp/4_table_bordered.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestTableInSubchapter(t *testing.T) {
	c := New()

	fontRegular := fonts.NewFontHelvetica()
	fontBold := fonts.NewFontHelveticaBold()

	ch := c.NewChapter("Document control")
	ch.SetMargins(0, 0, 40, 0)
	ch.GetHeading().SetFont(fontRegular)
	ch.GetHeading().SetFontSize(18)
	ch.GetHeading().SetColor(ColorRGBFrom8bit(72, 86, 95))

	sc := c.NewSubchapter(ch, "Issuer details")
	sc.SetMargins(0, 0, 5, 0)
	sc.GetHeading().SetFont(fontRegular)
	sc.GetHeading().SetFontSize(18)
	sc.GetHeading().SetColor(ColorRGBFrom8bit(72, 86, 95))

	issuerTable := NewTable(2)

	p := NewParagraph("Non-Disclosure")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(ColorWhite)
	cell := issuerTable.NewCell()
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorBlack)
	cell.SetBorder(CellBorderStyleBox, 1.0)
	cell.SetIndent(5)

	p = NewParagraph("Company Inc.")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(ColorGreen)
	cell = issuerTable.NewCell()
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorRed)
	cell.SetBorder(CellBorderStyleBox, 1.0)
	cell.SetIndent(5)

	p = NewParagraph("Belongs to")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorBlack)
	cell.SetBorder(CellBorderStyleBox, 1.0)
	cell.SetIndent(5)

	p = NewParagraph("Bezt business bureu")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(ColorGreen)
	cell = issuerTable.NewCell()
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorRed)
	cell.SetBorder(CellBorderStyleBox, 1.0)
	cell.SetIndent(5)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentCenter)
	//cell.SetVerticalAlignment(CellVerticalAlignmentMiddle)

	issuerTable.SetMargins(5, 5, 10, 10)

	ch.Add(issuerTable)

	sc = c.NewSubchapter(ch, "My Statement")
	//sc.SetMargins(0, 0, 5, 0)
	sc.GetHeading().SetFont(fontRegular)
	sc.GetHeading().SetFontSize(18)
	sc.GetHeading().SetColor(ColorRGBFrom8bit(72, 86, 95))

	myText := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt " +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore " +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum."

	myPara := NewParagraph(myText)
	myPara.SetFont(fontRegular)
	myPara.SetFontSize(10)
	myPara.SetColor(ColorRGBFrom8bit(72, 86, 95))
	myPara.SetTextAlignment(TextAlignmentJustify)
	myPara.SetLineHeight(1.5)
	sc.Add(myPara)

	err := c.Draw(ch)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = c.WriteToFile("/tmp/4_tables_in_subchap.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// Add headers and footers via creator.
func addHeadersAndFooters(c *Creator) {
	c.DrawHeader(func(header *Block, args HeaderFunctionArgs) {
		/*
			if pageNum == 1 {
				// Skip on front Page.
				return
			}
		*/

		// Add Page number
		p := NewParagraph(fmt.Sprintf("Page %d / %d", args.PageNum, args.TotalPages))
		p.SetPos(0.8*header.Width(), 20)
		header.Draw(p)

		// Draw on the template...
		img, err := NewImageFromFile(testImageFile1)
		if err != nil {
			fmt.Printf("ERROR : %v\n", err)
		}
		img.ScaleToHeight(0.4 * c.pageMargins.top)
		img.SetPos(20, 10)

		header.Draw(img)
	})

	c.DrawFooter(func(footer *Block, args FooterFunctionArgs) {
		/*
			if pageNum == 1 {
				// Skip on front Page.
				return
			}
		*/

		// Add company name.
		companyName := "Company inc."
		p := NewParagraph(companyName)
		p.SetPos(0.1*footer.Width(), 10)
		footer.Draw(p)

		p = NewParagraph("July 2017")
		p.SetPos(0.8*footer.Width(), 10)
		footer.Draw(p)
	})
}

// Test creating headers and footers.
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

	pixelWidth := oversampling * int(math.Ceil(width))
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

// Test adding encryption to output.
func TestEncrypting1(t *testing.T) {
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

	c.SetPdfWriterAccessFunc(func(w *model.PdfWriter) error {
		userPass := []byte("password")
		ownerPass := []byte("password")
		err := w.Encrypt(userPass, ownerPass, nil)
		return err
	})

	err := c.WriteToFile("/tmp/6_chapters_encrypted_password.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}
