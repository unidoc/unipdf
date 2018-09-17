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
	"github.com/unidoc/unidoc/pdf/contentstream/draw"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/textencoding"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

const testPdfFile1 = "./testdata/minimal.pdf"
const testPdfLoremIpsumFile = "./testdata/lorem.pdf"
const testPdfTemplatesFile1 = "./testdata/templates1.pdf"
const testImageFile1 = "./testdata/logo.png"
const testImageFile2 = "./testdata/signature.png"
const testRobotoRegularTTFFile = "./testdata/roboto/Roboto-Regular.ttf"
const testRobotoBoldTTFFile = "./testdata/roboto/Roboto-Bold.ttf"
const testWts11TTFFile = "./testdata/wts11.ttf"

// XXX: /tmp/2_p_multi.pdf which is created in this test gives an error message when opened in
//      Adobe Reader: The font FreeSans contains bad Widths.
//      This problem did not occur when I replaced FreeSans.ttf with LiberationSans-Regular.ttf
const testFreeSansTTFFile = "./testdata/FreeSans.ttf"

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

// Example drawing image and line shape on a block and applying to pages, also demonstrating block
// rotation.
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
// TODO: In the future we would like the paragraph to split up between pages.  Split up on line,
// never allowing less than 2 lines to go over (common practice).
// TODO: In the future we would like to implement Donald Knuth's line wrapping algorithm or
// something similar.
func TestParagraphWrapping(t *testing.T) {
	creator := New()

	p := NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum.")

	p.SetMargins(0, 0, 5, 0)

	alignments := []TextAlignment{TextAlignmentLeft, TextAlignmentJustify, TextAlignmentCenter,
		TextAlignmentRight}
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

	alignments := []TextAlignment{TextAlignmentLeft, TextAlignmentJustify, TextAlignmentCenter,
		TextAlignmentRight}
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

	helvetica := model.NewStandard14FontMustCompile(model.Helvetica)

	fonts := []*model.PdfFont{roboto, robotoBold, helvetica, roboto, robotoBold, helvetica}
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

	names := []model.Standard14Font{
		model.Courier,
		model.CourierBold,
		model.CourierBoldOblique,
		model.CourierOblique,
		model.Helvetica,
		model.HelveticaBold,
		model.HelveticaBoldOblique,
		model.HelveticaOblique,
		model.TimesRoman,
		model.TimesBold,
		model.TimesBoldItalic,
		model.TimesItalic,
		model.Symbol,
		model.ZapfDingbats,
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

	for idx, name := range names {
		p := NewParagraph(texts[idx])
		font := model.NewStandard14FontMustCompile(name)
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

// Test paragraph with Chinese characters.
func TestParagraphChinese(t *testing.T) {
	creator := New()

	lines := []string{
		"你好",
		"你好你好你好你好",
		"河上白云",
	}

	for _, line := range lines {
		p := NewParagraph(line)

		font, err := model.NewCompositePdfFontFromTTFFile(testWts11TTFFile)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}

		p.SetFont(font)

		err = creator.Draw(p)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err := creator.WriteToFile("/tmp/2_p_nihao.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// Test paragraph with composite font and various unicode characters.
func TestParagraphUnicode(t *testing.T) {
	creator := New()

	font, err := model.NewCompositePdfFontFromTTFFile(testFreeSansTTFFile)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	texts := []string{
		"Testing of letters \u010c,\u0106,\u0160,\u017d,\u0110",
		"Vous \u00eates d'o\u00f9?",
		"\u00c0 tout \u00e0 l'heure. \u00c0 bient\u00f4t.",
		"Je me pr\u00e9sente.",
		"C'est un \u00e9tudiant.",
		"\u00c7a va?",
		"Il est ing\u00e9nieur. Elle est m\u00e9decin.",
		"C'est une fen\u00eatre.",
		"R\u00e9p\u00e9tez, s'il vous pla\u00eet.",
		"Odkud jste?",
		"Uvid\u00edme se za chvilku. M\u011bj se.",
		"Dovolte, abych se p\u0159edstavil.",
		"To je studentka.",
		"V\u0161echno v po\u0159\u00e1dku?",
		"On je in\u017een\u00fdr. Ona je l\u00e9ka\u0159.",
		"Toto je okno.",
		"Zopakujte to pros\u00edm.",
		"\u041e\u0442\u043a\u0443\u0434\u0430 \u0442\u044b?",
		"\u0423\u0432\u0438\u0434\u0438\u043c\u0441\u044f \u0432 \u043d\u0435\u043c\u043d\u043e\u0433\u043e. \u0423\u0432\u0438\u0434\u0438\u043c\u0441\u044f.",
		"\u041f\u043e\u0437\u0432\u043e\u043b\u044c\u0442\u0435 \u043c\u043d\u0435 \u043f\u0440\u0435\u0434\u0441\u0442\u0430\u0432\u0438\u0442\u044c\u0441\u044f.",
		"\u042d\u0442\u043e \u0441\u0442\u0443\u0434\u0435\u043d\u0442.",
		"\u0425\u043e\u0440\u043e\u0448\u043e?",
		"\u041e\u043d \u0438\u043d\u0436\u0435\u043d\u0435\u0440. \u041e\u043d\u0430 \u0434\u043e\u043a\u0442\u043e\u0440.",
		"\u042d\u0442\u043e \u043e\u043a\u043d\u043e.",
		"\u041f\u043e\u0432\u0442\u043e\u0440\u0438\u0442\u0435, \u043f\u043e\u0436\u0430\u043b\u0443\u0439\u0441\u0442\u0430.",
		`Lorem Ipsum - это текст-"рыба", часто используемый в печати и вэб-дизайне. Lorem Ipsum является стандартной "рыбой" для текстов на латинице с начала XVI века. В то время некий безымянный печатник создал большую коллекцию размеров и форм шрифтов, используя Lorem Ipsum для распечатки образцов. Lorem Ipsum не только успешно пережил без заметных изменений пять веков, но и перешагнул в электронный дизайн. Его популяризации в новое время послужили публикация листов Letraset с образцами Lorem Ipsum в 60-х годах и, в более недавнее время, программы электронной вёрстки типа Aldus PageMaker, в шаблонах которых используется Lorem Ipsum.`,
	}

	for _, text := range texts {
		fmt.Printf("Text: %s\n", text)

		p := NewParagraph(text)
		p.SetFont(font)

		err = creator.Draw(p)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}

	err = creator.WriteToFile("/tmp/2_p_multi.pdf")
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
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	p.SetEnableWrap(true)
	p.SetWidth(cell.Width(c.Context()))
	p.SetTextAlignment(TextAlignmentJustify)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	p = NewParagraph("B Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentRight)
	cell.SetContent(p)

	cell = table.NewCell()
	p = NewParagraph("C Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	p.SetEnableWrap(true)
	cell.SetContent(p)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	cell = table.NewCell()
	p = NewParagraph("1,4")
	cell.SetContent(p)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	cell = table.NewCell()
	p = NewParagraph("2,1")
	cell.SetContent(p)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	cell = table.NewCell()
	p = NewParagraph("2,2")
	cell.SetContent(p)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	cell = table.NewCell()
	p = NewParagraph("2,2")
	cell.SetContent(p)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	//table.SkipCells(1) // Skip over 2,3.

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	//p = NewParagraph("D Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	p = NewParagraph("X")
	p.SetEnableWrap(true)
	cell.SetContent(p)

	// Skip over two rows.
	table.SkipRows(2)
	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
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
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
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
func TestBorderedTable1(t *testing.T) {
	table := NewTable(4) // Mx4 table
	// Default, equal column sizes (4x0.25)...
	table.SetColumnWidths(0.5, 0.2, 0.2, 0.1)

	cell1 := table.NewCell()
	p := NewParagraph("A")
	cell1.SetContent(p)
	cell1.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1) // border will be on left
	cell1.SetBorderLineStyle(draw.LineStyleDashed)

	table.SkipCells(1)

	cell2 := table.NewCell()
	p = NewParagraph("B")
	cell2.SetContent(p)
	cell2.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1) // border will be around
	cell2.SetBorderLineStyle(draw.LineStyleSolid)
	cell2.SetBackgroundColor(ColorRed)

	table.SkipCells(1) // Skip over 2,3.

	// Skip over two rows.
	table.SkipRows(2)
	cell8 := table.NewCell()
	p = NewParagraph("H")
	cell8.SetContent(p)
	cell8.SetBorder(CellBorderSideRight, CellBorderStyleSingle, 1) // border will be on right
	cell8.SetBorderLineStyle(draw.LineStyleSolid)

	c := New()
	c.Draw(table)

	err := c.WriteToFile("/tmp/4_table_bordered.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// Test creating and drawing a table.
func TestBorderedTable2(t *testing.T) {
	table := NewTable(4) // Mx4 table
	// Default, equal column sizes (4x0.25)...
	table.SetColumnWidths(0.5, 0.2, 0.2, 0.1)

	cell1 := table.NewCell()
	p := NewParagraph("A")
	cell1.SetContent(p)
	cell1.SetBorder(CellBorderSideLeft, CellBorderStyleSingle, 1) // border will be on left
	cell1.SetBorderLineStyle(draw.LineStyleSolid)

	cell2 := table.NewCell()
	p = NewParagraph("B")
	cell2.SetContent(p)
	cell2.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1) // border will be around
	cell2.SetBorderLineStyle(draw.LineStyleSolid)

	table.SkipCells(1)

	cell4 := table.NewCell()
	p = NewParagraph("D")
	cell4.SetContent(p)
	cell4.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1) // border will be around
	cell4.SetBorderLineStyle(draw.LineStyleSolid)

	table.SkipCells(1)

	cell6 := table.NewCell()
	p = NewParagraph("F")
	cell6.SetContent(p)
	cell6.SetBorder(CellBorderSideLeft, CellBorderStyleSingle, 1) // border will be on left
	cell6.SetBorderLineStyle(draw.LineStyleSolid)

	table.SkipCells(1) // Skip over 2,3.

	cell7 := table.NewCell()
	p = NewParagraph("G")
	cell7.SetContent(p)
	cell7.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1) // border will be around
	cell7.SetBorderLineStyle(draw.LineStyleSolid)

	// Skip over two rows.
	table.SkipRows(2)
	cell8 := table.NewCell()
	p = NewParagraph("H")
	cell8.SetContent(p)
	cell8.SetBorder(CellBorderSideRight, CellBorderStyleSingle, 1) // border will be on right
	cell8.SetBorderLineStyle(draw.LineStyleSolid)

	c := New()
	c.Draw(table)

	err := c.WriteToFile("/tmp/4_table_bordered.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func newContent(text string, alignment TextAlignment, font *model.PdfFont, fontSize float64, color Color) *Paragraph {
	p := NewParagraph(text)
	p.SetFontSize(fontSize)
	p.SetTextAlignment(alignment)
	p.SetFont(font)
	p.SetColor(color)
	return p
}

func newBillItem(t *Table, no, date, notes, amount, con, retApplied, ret, netBill string) {
	timesBold := model.NewStandard14FontMustCompile(model.TimesBold)

	billNo := t.NewCell()
	billNo.SetContent(newContent(no, TextAlignmentLeft, timesBold, 8, ColorBlack))
	billDate := t.NewCell()
	billDate.SetContent(newContent(date, TextAlignmentCenter, timesBold, 8, ColorBlack))
	billNotes := t.NewCell()
	billNotes.SetContent(newContent(notes, TextAlignmentLeft, timesBold, 8, ColorBlack))
	billAmount := t.NewCell()
	billAmount.SetContent(newContent(amount, TextAlignmentRight, timesBold, 8, ColorBlack))
	billCon := t.NewCell()
	billCon.SetContent(newContent(con, TextAlignmentLeft, timesBold, 8, ColorBlack))
	billRetApplied := t.NewCell()
	billRetApplied.SetContent(newContent(retApplied, TextAlignmentRight, timesBold, 8, ColorBlack))
	billRet := t.NewCell()
	billRet.SetContent(newContent(ret, TextAlignmentLeft, timesBold, 8, ColorBlack))
	billNetBill := t.NewCell()
	billNetBill.SetContent(newContent(netBill, TextAlignmentRight, timesBold, 8, ColorBlack))
}

// Test creating and drawing a table.
func TestCreatorHendricksReq1(t *testing.T) {
	timesRoman := model.NewStandard14FontMustCompile(model.TimesRoman)
	timesBold := model.NewStandard14FontMustCompile(model.TimesBold)
	table := NewTable(3) // Mx4 table
	// Default, equal column sizes (4x0.25)...
	table.SetColumnWidths(0.35, 0.30, 0.35)

	projectColorOne := ColorBlue
	projectColorTwo := ColorRed

	companyTitle := table.NewCell()
	companyTitle.SetContent(newContent("Hendricks Consulting LLC", TextAlignmentLeft, timesBold, 12, projectColorOne))

	table.SkipCells(1)

	pageHeader := table.NewCell()
	pageHeader.SetContent(newContent("Billing Schedule by Project", TextAlignmentCenter, timesBold, 12, ColorBlack))
	pageHeader.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 3)
	pageHeader.SetBorderLineStyle(draw.LineStyleSolid)

	companyAddress := table.NewCell()
	companyAddress.SetContent(newContent("2666 Airport Drive, Apt. 309", TextAlignmentLeft, timesRoman, 8, ColorBlack))

	table.SkipCells(2)

	companyLocation := table.NewCell()
	companyLocation.SetContent(newContent("Portland, Oregon, 92019", TextAlignmentLeft, timesRoman, 8, ColorBlack))

	table.SkipCells(1)

	printingDate := table.NewCell()
	printingDate.SetContent(newContent("Printed on: 22/02/2011", TextAlignmentRight, timesRoman, 8, ColorBlack))

	companyTelAndFax := table.NewCell()
	companyTelAndFax.SetContent(newContent("Tel: (999) 609-4032  Fax: (999) 999-9922", TextAlignmentLeft, timesRoman, 8, ColorBlack))

	table.SkipCells(1)

	pageOf := table.NewCell()
	pageOf.SetContent(newContent("Page 10 of 10", TextAlignmentRight, timesRoman, 8, ColorBlack))

	email := table.NewCell()
	email.SetContent(newContent("admin@hendricks.com", TextAlignmentLeft, timesRoman, 8, ColorBlack))

	table.SkipCells(2)

	website := table.NewCell()
	website.SetContent(newContent("www.hendricks.com", TextAlignmentLeft, timesRoman, 8, ColorBlack))

	table2 := NewTable(5)
	table2.SetColumnWidths(0.20, 0.20, 0.20, 0.20, 0.20)
	table2.SkipCells(5)

	projectName := table2.NewCell()
	projectName.SetContent(newContent("Project Name (ID):", TextAlignmentLeft, timesBold, 8, projectColorOne))

	projectNameValue := table2.NewCell()
	projectNameValue.SetContent(newContent("Biggi Group", TextAlignmentLeft, timesBold, 8, ColorBlack))

	table2.SkipCells(3)

	projectID := table2.NewCell()
	projectID.SetContent(newContent("Project ID:", TextAlignmentLeft, timesBold, 8, projectColorOne))

	projectIDValue := table2.NewCell()
	projectIDValue.SetContent(newContent("BG:01", TextAlignmentLeft, timesBold, 8, ColorBlack))

	table2.SkipCells(1)

	contractType := table2.NewCell()
	contractType.SetContent(newContent("Contract Type:", TextAlignmentRight, timesBold, 8, projectColorOne))

	contractTypeValue := table2.NewCell()
	contractTypeValue.SetContent(newContent("Percentage", TextAlignmentLeft, timesRoman, 8, ColorBlack))

	projectManager := table2.NewCell()
	projectManager.SetContent(newContent("Manager:", TextAlignmentLeft, timesBold, 8, projectColorOne))

	projectManagerValue := table2.NewCell()
	projectManagerValue.SetContent(newContent("SHH", TextAlignmentLeft, timesBold, 8, ColorBlack))

	table2.SkipCells(1)

	contractAmount := table2.NewCell()
	contractAmount.SetContent(newContent("Contract Amount:", TextAlignmentRight, timesBold, 8, projectColorOne))

	contractAmountValue := table2.NewCell()
	contractAmountValue.SetContent(newContent("$2,975.00", TextAlignmentLeft, timesRoman, 8, ColorBlack))

	clientID := table2.NewCell()
	clientID.SetContent(newContent("Client ID:", TextAlignmentLeft, timesBold, 8, projectColorOne))

	clientIDValue := table2.NewCell()
	clientIDValue.SetContent(newContent("Baggi ehf", TextAlignmentLeft, timesBold, 8, ColorBlack))

	table2.SkipCells(1)

	retainerAmount := table2.NewCell()
	retainerAmount.SetContent(newContent("Retainer Amount:", TextAlignmentRight, timesBold, 8, projectColorOne))

	retainerAmountValue := table2.NewCell()
	retainerAmountValue.SetContent(newContent("", TextAlignmentLeft, timesRoman, 8, ColorBlack))

	table3 := NewTable(8)
	table3.SetColumnWidths(0.05, 0.10, 0.35, 0.10, 0.10, 0.10, 0.10, 0.10)
	table3.SkipCells(8)

	billNo := table3.NewCell()
	billNo.SetContent(newContent("Bill #", TextAlignmentLeft, timesBold, 8, projectColorOne))
	billNo.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 2)
	billNo.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	billNo.SetBorderColor(projectColorOne)

	billDate := table3.NewCell()
	billDate.SetContent(newContent("Date", TextAlignmentLeft, timesBold, 8, projectColorOne))
	billDate.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 2)
	billDate.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	billDate.SetBorderColor(projectColorOne)

	billNotes := table3.NewCell()
	billNotes.SetContent(newContent("Notes", TextAlignmentLeft, timesBold, 8, projectColorOne))
	billNotes.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 2)
	billNotes.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	billNotes.SetBorderColor(projectColorOne)

	billAmount := table3.NewCell()
	billAmount.SetContent(newContent("Bill Amount", TextAlignmentLeft, timesBold, 8, projectColorOne))
	billAmount.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 2)
	billAmount.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	billAmount.SetBorderColor(projectColorOne)

	billCon := table3.NewCell()
	billCon.SetContent(newContent("% Con", TextAlignmentLeft, timesBold, 8, projectColorOne))
	billCon.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 2)
	billCon.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	billCon.SetBorderColor(projectColorOne)

	billRetApplied := table3.NewCell()
	billRetApplied.SetContent(newContent("Ret Applied", TextAlignmentLeft, timesBold, 8, projectColorOne))
	billRetApplied.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 2)
	billRetApplied.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	billRetApplied.SetBorderColor(projectColorOne)

	billRet := table3.NewCell()
	billRet.SetContent(newContent("% Ret", TextAlignmentLeft, timesBold, 8, projectColorOne))
	billRet.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 2)
	billRet.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	billRet.SetBorderColor(projectColorOne)

	billNetBill := table3.NewCell()
	billNetBill.SetContent(newContent("Net Bill Amt", TextAlignmentLeft, timesBold, 8, projectColorOne))
	billNetBill.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 2)
	billNetBill.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	billNetBill.SetBorderColor(projectColorOne)

	newBillItem(table3, "1", "1/2/2012", "", "$297.50", "", "$0.00", "", "$297.50")
	newBillItem(table3, "2", "1/2/2012", "", "$595.00", "", "$0.00", "", "$595.00")
	newBillItem(table3, "3", "1/3/2012", "", "$446.25", "", "$0.00", "", "$446.25")
	newBillItem(table3, "4", "1/4/2012", "", "$595.00", "", "$0.00", "", "$595.00")
	newBillItem(table3, "5", "1/5/2012", "", "$446.25", "", "$0.00", "", "$446.25")
	newBillItem(table3, "6", "1/6/2012", "", "$892.50", "", "$0.00", "", "$892.50")

	table3.SkipCells(2 + 8)

	totalBill := table3.NewCell()
	totalBill.SetContent(newContent("Total:     ", TextAlignmentRight, timesBold, 8, projectColorTwo))

	totalBillAmount := table3.NewCell()
	totalBillAmount.SetContent(newContent("$3,272.50", TextAlignmentRight, timesBold, 8, projectColorTwo))
	totalBillAmount.SetBorder(CellBorderSideTop, CellBorderStyleDouble, 1)
	totalBillAmount.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)

	table3.SkipCells(1)

	totalRetAmount := table3.NewCell()
	totalRetAmount.SetContent(newContent("$0.00", TextAlignmentRight, timesBold, 8, projectColorTwo))
	totalRetAmount.SetBorder(CellBorderSideTop, CellBorderStyleDouble, 1)
	totalRetAmount.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)

	table3.SkipCells(1)

	totalNetAmount := table3.NewCell()
	totalNetAmount.SetContent(newContent("$3,272.50", TextAlignmentRight, timesBold, 8, projectColorTwo))
	totalNetAmount.SetBorder(CellBorderSideTop, CellBorderStyleDouble, 1)
	totalNetAmount.SetBorder(CellBorderSideBottom, CellBorderStyleSingle, 1)
	totalNetAmount.SetBorderLineStyle(draw.LineStyleSolid)

	c := New()
	c.Draw(table)
	c.Draw(table2)
	c.Draw(table3)

	err := c.WriteToFile("/tmp/hendricks.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestCreatorTableBorderReq1(t *testing.T) {
	timesRoman := model.NewStandard14FontMustCompile(model.TimesRoman)
	table := NewTable(1) // Mx4 table
	table.SetColumnWidths(1)

	fullLengthCell := table.NewCell()
	fullLengthCell.SetContent(newContent("boxed, solid, default width", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	fullLengthCell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	table2 := NewTable(4) // Mx4 table
	table2.SetColumnWidths(.25, .25, .25, .25)

	table2.SkipCells(4)

	a := table2.NewCell()
	a.SetContent(newContent("A", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	a.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	b := table2.NewCell()
	b.SetContent(newContent("B", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	b.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	cc := table2.NewCell()
	cc.SetContent(newContent("C", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	cc.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	d := table2.NewCell()
	d.SetContent(newContent("D", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	d.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	e := table2.NewCell()
	e.SetContent(newContent("E", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	e.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	f := table2.NewCell()
	f.SetContent(newContent("F", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	f.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	g := table2.NewCell()
	g.SetContent(newContent("G", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	g.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	h := table2.NewCell()
	h.SetContent(newContent("H", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	h.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	i := table2.NewCell()
	i.SetContent(newContent("I", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	i.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	j := table2.NewCell()
	j.SetContent(newContent("J", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	j.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

	table3 := NewTable(1) // Mx4 table
	table3.SetColumnWidths(1)

	table3.SkipCells(1)

	dash := table3.NewCell()
	dash.SetContent(newContent("boxed, dashed, default width", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	dash.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	dash.SetBorderLineStyle(draw.LineStyleDashed)

	table4 := NewTable(4) // Mx4 table
	table4.SetColumnWidths(.25, .25, .25, .25)

	table4.SkipCells(4)

	ad := table4.NewCell()
	ad.SetContent(newContent("A", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	ad.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	ad.SetBorderLineStyle(draw.LineStyleDashed)

	bd := table4.NewCell()
	bd.SetContent(newContent("B", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	bd.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	bd.SetBorderLineStyle(draw.LineStyleDashed)

	table4.SkipCells(2)

	ccd := table4.NewCell()
	ccd.SetContent(newContent("C", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	ccd.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	ccd.SetBorderLineStyle(draw.LineStyleDashed)

	dd := table4.NewCell()
	dd.SetContent(newContent("D", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	dd.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	dd.SetBorderLineStyle(draw.LineStyleDashed)

	table4.SkipCells(2)

	ed := table4.NewCell()
	ed.SetContent(newContent("E", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	ed.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	ed.SetBorderLineStyle(draw.LineStyleDashed)

	fd := table4.NewCell()
	fd.SetContent(newContent("F", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	fd.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	fd.SetBorderLineStyle(draw.LineStyleDashed)

	gd := table4.NewCell()
	gd.SetContent(newContent("G", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	gd.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	gd.SetBorderLineStyle(draw.LineStyleDashed)

	hd := table4.NewCell()
	hd.SetContent(newContent("H", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	hd.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	hd.SetBorderLineStyle(draw.LineStyleDashed)

	id := table4.NewCell()
	id.SetContent(newContent("I", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	id.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	id.SetBorderLineStyle(draw.LineStyleDashed)

	jd := table4.NewCell()
	jd.SetContent(newContent("J", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	jd.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	jd.SetBorderLineStyle(draw.LineStyleDashed)

	kd := table4.NewCell()
	kd.SetContent(newContent("K", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	kd.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	kd.SetBorderLineStyle(draw.LineStyleDashed)

	ld := table4.NewCell()
	ld.SetContent(newContent("L", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	ld.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	ld.SetBorderLineStyle(draw.LineStyleDashed)

	md := table4.NewCell()
	md.SetContent(newContent("M", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	md.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	md.SetBorderLineStyle(draw.LineStyleDashed)

	table5 := NewTable(1) // Mx4 table
	table5.SetColumnWidths(1)

	table5.SkipCells(1)

	doubled := table5.NewCell()
	doubled.SetContent(newContent("boxed, double, default width", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	doubled.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	table6 := NewTable(4) // Mx4 table
	table6.SetColumnWidths(.25, .25, .25, .25)

	table6.SkipCells(4)

	add := table6.NewCell()
	add.SetContent(newContent("A", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	add.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	bdd := table6.NewCell()
	bdd.SetContent(newContent("B", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	bdd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	ccdd := table6.NewCell()
	ccdd.SetContent(newContent("C", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	ccdd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	ddd := table6.NewCell()
	ddd.SetContent(newContent("D", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	ddd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	edd := table6.NewCell()
	edd.SetContent(newContent("E", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	edd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	fdd := table6.NewCell()
	fdd.SetContent(newContent("F", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	fdd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	gdd := table6.NewCell()
	gdd.SetContent(newContent("G", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	gdd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	hdd := table6.NewCell()
	hdd.SetContent(newContent("H", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	hdd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	idd := table6.NewCell()
	idd.SetContent(newContent("I", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	idd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	jdd := table6.NewCell()
	jdd.SetContent(newContent("J", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	jdd.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	table7 := NewTable(1) // Mx4 table
	table7.SetColumnWidths(1)

	table7.SkipCells(1)

	fullLengthCell7 := table7.NewCell()
	fullLengthCell7.SetContent(newContent("boxed, solid, thick", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	fullLengthCell7.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	table8 := NewTable(4) // Mx4 table
	table8.SetColumnWidths(.25, .25, .25, .25)

	table8.SkipCells(4)

	a8 := table8.NewCell()
	a8.SetContent(newContent("A", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	a8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	b8 := table8.NewCell()
	b8.SetContent(newContent("B", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	b8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	cc8 := table8.NewCell()
	cc8.SetContent(newContent("C", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	cc8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	d8 := table8.NewCell()
	d8.SetContent(newContent("D", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	d8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	e8 := table8.NewCell()
	e8.SetContent(newContent("E", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	e8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	f8 := table8.NewCell()
	f8.SetContent(newContent("F", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	f8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	g8 := table8.NewCell()
	g8.SetContent(newContent("G", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	g8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	h8 := table8.NewCell()
	h8.SetContent(newContent("H", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	h8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	i8 := table8.NewCell()
	i8.SetContent(newContent("I", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	i8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	j8 := table8.NewCell()
	j8.SetContent(newContent("J", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	j8.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)

	table9 := NewTable(1) // Mx4 table
	table9.SetColumnWidths(1)

	table9.SkipCells(1)

	fullLengthCell9 := table9.NewCell()
	fullLengthCell9.SetContent(newContent("boxed, dashed, thick", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	fullLengthCell9.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	fullLengthCell9.SetBorderLineStyle(draw.LineStyleDashed)

	table10 := NewTable(4) // Mx4 table
	table10.SetColumnWidths(.25, .25, .25, .25)

	table10.SkipCells(4)

	a10 := table10.NewCell()
	a10.SetContent(newContent("A", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	a10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	a10.SetBorderLineStyle(draw.LineStyleDashed)

	b10 := table10.NewCell()
	b10.SetContent(newContent("B", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	b10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	b10.SetBorderLineStyle(draw.LineStyleDashed)

	cc10 := table10.NewCell()
	cc10.SetContent(newContent("C", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	cc10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	cc10.SetBorderLineStyle(draw.LineStyleDashed)

	d10 := table10.NewCell()
	d10.SetContent(newContent("D", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	d10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	d10.SetBorderLineStyle(draw.LineStyleDashed)

	e10 := table10.NewCell()
	e10.SetContent(newContent("E", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	e10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	e10.SetBorderLineStyle(draw.LineStyleDashed)

	f10 := table10.NewCell()
	f10.SetContent(newContent("F", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	f10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	f10.SetBorderLineStyle(draw.LineStyleDashed)

	g10 := table10.NewCell()
	g10.SetContent(newContent("G", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	g10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	g10.SetBorderLineStyle(draw.LineStyleDashed)

	h10 := table10.NewCell()
	h10.SetContent(newContent("H", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	h10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	h10.SetBorderLineStyle(draw.LineStyleDashed)

	i10 := table10.NewCell()
	i10.SetContent(newContent("I", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	i10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	i10.SetBorderLineStyle(draw.LineStyleDashed)

	j10 := table10.NewCell()
	j10.SetContent(newContent("J", TextAlignmentLeft, timesRoman, 10, ColorBlack))
	j10.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 2)
	j10.SetBorderLineStyle(draw.LineStyleDashed)

	c := New()
	c.Draw(table)
	c.Draw(table2)
	c.Draw(table3)
	c.Draw(table4)
	c.Draw(table5)
	c.Draw(table6)
	c.Draw(table7)
	c.Draw(table8)
	c.Draw(table9)
	c.Draw(table10)

	err := c.WriteToFile("/tmp/table_border_req1_test.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestCellBorder(t *testing.T) {
	timesBold := model.NewStandard14FontMustCompile(model.TimesBold)

	table := NewTable(2)
	table.SetColumnWidths(0.50, 0.50)

	cell1 := table.NewCell()
	cell1.SetContent(newContent("Cell 1", TextAlignmentLeft, timesBold, 8, ColorRed))
	cell1.SetBorder(CellBorderSideAll, CellBorderStyleDouble, 1)

	c := New()
	c.Draw(table)

	err := c.WriteToFile("/tmp/cell.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestTableInSubchapter(t *testing.T) {
	c := New()

	fontRegular := model.NewStandard14FontMustCompile(model.Helvetica)
	fontBold := model.NewStandard14FontMustCompile(model.HelveticaBold)

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
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1.0)
	cell.SetIndent(5)

	p = NewParagraph("Company Inc.")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(ColorGreen)
	cell = issuerTable.NewCell()
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorRed)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1.0)
	cell.SetIndent(5)

	p = NewParagraph("Belongs to")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorBlack)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1.0)
	cell.SetIndent(5)

	p = NewParagraph("Bezt business bureu")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(ColorGreen)
	cell = issuerTable.NewCell()
	cell.SetContent(p)
	cell.SetBackgroundColor(ColorRed)
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1.0)
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
