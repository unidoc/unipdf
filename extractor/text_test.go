/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/text/unicode/norm"
)

const (
	// markupPDFs should be set to true to save PDF pages with their bounding boxes marked.
	// See markupList below.
	markupPDFs = false
	// markupDir is the directory where the marked-up PDFs are saved.
	// The PDFs in markupDir can be viewed in a PDF viewer to check that they correct.
	markupDir = "marked.up"
)

// NOTE: We do a best effort at finding the PDF file because we don't keep PDF test files in this
// repo so you will need to setup UNIDOC_EXTRACT_TESTDATA to point at the corpus directory.

// forceTest should be set to true to force running all tests.
// NOTE: Setting environment variable UNIDOC_EXTRACT_FORCETEST = 1 sets this to true.
var forceTest = os.Getenv("UNIDOC_EXTRACT_FORCETEST") == "1"

var corpusFolder = os.Getenv("UNIDOC_EXTRACT_TESTDATA")

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelInfo))
	if flag.Lookup("test.v") != nil {
		isTesting = true
	}
}

// TestTextExtractionFragments tests text extraction on the PDF fragments in `fragmentTests`.
func TestTextExtractionFragments(t *testing.T) {
	fragmentTests := []struct {
		name     string
		contents string
		text     string
	}{
		{
			name: "portrait",
			contents: `
        BT
        /UniDocCourier 24 Tf
        (Hello World!)Tj
        0 -10 Td
        (Doink)Tj
        ET
        `,
			text: "Hello World!\nDoink",
		},
		{
			name: "landscape",
			contents: `
        BT
        /UniDocCourier 24 Tf
        0 1 -1 0 0 0 Tm
        (Hello World!)Tj
        0 -10 Td
        (Doink)Tj
        ET
        `,
			text: "Hello World!\nDoink",
		},
		{
			name: "180 degree rotation",
			contents: `
        BT
        /UniDocCourier 24 Tf
        -1 0 0 -1 0 0 Tm
        (Hello World!)Tj
        0 -10 Td
        (Doink)Tj
        ET
        `,
			text: "Hello World!\nDoink",
		},
		{
			name: "Helvetica",
			contents: `
        BT
        /UniDocHelvetica 24 Tf
        0 -1 1 0 0 0 Tm
        (Hello World!)Tj
        0 -10 Td
        (Doink)Tj
        ET
        `,
			text: "Hello World!\nDoink",
		},
	}

	// Setup mock resources.
	resources := model.NewPdfPageResources()
	{
		courier := model.NewStandard14FontMustCompile(model.CourierName)
		helvetica := model.NewStandard14FontMustCompile(model.HelveticaName)
		resources.SetFontByName("UniDocHelvetica", helvetica.ToPdfObject())
		resources.SetFontByName("UniDocCourier", courier.ToPdfObject())
	}

	for _, f := range fragmentTests {
		t.Run(f.name, func(t *testing.T) {
			e := Extractor{resources: resources, contents: f.contents}
			text, err := e.ExtractText()
			if err != nil {
				t.Fatalf("Error extracting text: %q err=%v", f.name, err)
				return
			}
			if text != f.text {
				t.Fatalf("Text mismatch: %q Got %q. Expected %q", f.name, text, f.text)
				return
			}
		})
	}
}

// TestTextExtractionFiles tests text extraction on a set of PDF files.
// It checks for the existence of specified strings of words on specified pages.
// We currently only check within lines as our line order is still improving.
func TestTextExtractionFiles(t *testing.T) {
	if len(corpusFolder) == 0 && !forceTest {
		t.Log("Corpus folder not set - skipping")
		return
	}
	for _, test := range fileExtractionTests {
		testExtractFileOptions(t, test.filename, test.pageTerms, false)
		testExtractFileOptions(t, test.filename, test.pageTerms, true)
	}
}

// TestTextLocations tests locations of text marks.
func TestTextLocations(t *testing.T) {
	if len(corpusFolder) == 0 && !forceTest {
		t.Log("Corpus folder not set - skipping")
		return
	}
	for _, e := range textLocTests {
		e.testTextComponent(t, false)
		e.testTextComponent(t, true)
	}
}

// TestTermMarksFiles stress tests testTermMarksMulti() by running it on all files in the corpus.
// It can take several minutes to run.
func TestTermMarksFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test")
	}
	common.Log.Info("Running text stress tests. go test --short to skip these.")
	if len(corpusFolder) == 0 && !forceTest {
		t.Log("Corpus folder not set - skipping")
		return
	}
	testTermMarksFiles(t)
}

// fileExtractionTests are PDF file names and terms we expect to find on specified pages of those
// PDF files.
// `pageTerms`[pageNum] are  the terms we expect to find on (1-offset) page number pageNum of
// the PDF file `filename`.
var fileExtractionTests = []struct {
	filename  string
	pageTerms map[int][]string
}{
	{filename: "reader.pdf",
		pageTerms: map[int][]string{
			1: []string{"A Research UNIX Reader:",
				"Annotated Excerpts from the Programmer’s Manual,",
				"1. Introduction",
				"To keep the size of this report",
				"last common ancestor of a radiative explosion",
			},
		},
	},
	{filename: "000026.pdf",
		pageTerms: map[int][]string{
			1: []string{"Fresh Flower",
				"Care & Handling ",
			},
		},
	},
	{filename: "search_sim_key.pdf",
		pageTerms: map[int][]string{
			2: []string{"A cryptographic scheme which enables searching",
				"Untrusted server should not be able to search for a word without authorization",
			},
		},
	},
	{filename: "Theil_inequality.pdf",
		pageTerms: map[int][]string{
			1: []string{"London School of Economics and Political Science"},
			4: []string{"The purpose of this paper is to set Theil’s approach"},
		},
	},
	{filename: "8207.pdf",
		pageTerms: map[int][]string{
			1: []string{"In building graphic systems for use with raster devices,"},
			2: []string{"The imaging model specifies how geometric shapes and colors are"},
			3: []string{"The transformation matrix T that maps application defined"},
		},
	},
	{filename: "ling-2013-0040ad.pdf",
		pageTerms: map[int][]string{
			1: []string{"Although the linguistic variation among texts is continuous"},
			2: []string{"distinctions. For example, much of the research on spoken/written"},
		},
	},
	{filename: "26-Hazard-Thermal-environment.pdf",
		pageTerms: map[int][]string{
			1: []string{"OHS Body of Knowledge"},
			2: []string{"Copyright notice and licence terms"},
		},
	},
	{filename: "Threshold_survey.pdf",
		pageTerms: map[int][]string{
			1: []string{"clustering, entropy, object attributes, spatial correlation, and local"},
		},
	},
	{filename: "circ2.pdf",
		pageTerms: map[int][]string{
			1: []string{"Understanding and complying with copyright law can be a challenge"},
		},
	},
	{filename: "rare_word.pdf",
		pageTerms: map[int][]string{
			6: []string{"words in the test set, we increase the BLEU score"},
		},
	},
	{filename: "Planck_Wien.pdf",
		pageTerms: map[int][]string{
			1: []string{"entropy of a system of n identical resonators in a stationary radiation field"},
		},
	},
	// Case where combineDiacritics was combining ' and " with preceeding letters.
	// NOTE(peterwilliams97): Part of the reason this test fails is that we don't currently read
	// Type0:CIDFontType0 font metrics and assume zero displacemet so that we place the ' and " too
	// close to the preceeding letters.
	{filename: "/rfc6962.txt.pdf",
		pageTerms: map[int][]string{
			4: []string{
				"timestamps for certificates they then don’t log",
				`The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",`},
		},
	},
	// TODO(peterwilliams97): Reinstate these 2 tests when diacritic combination is fixed.
	// {filename: "Ito_Formula.pdf",
	// 	pageTerms: map[int][]string{
	// 		1: []string{
	// 			"In the Itô stochastic calculus",
	// 			"In standard, non-stochastic calculus, one computes a derivative"},
	// 		2: []string{"Financial Economics Itô’s Formula"},
	// 	},
	// },
	// {filename: "thanh.pdf",
	// 	pageTerms: map[int][]string{
	// 		1: []string{"Hàn Thé̂ Thành"},
	// 	},
	// },
}

// testExtractFile tests the ExtractTextWithStats text extractor on `filename` and compares the
// extracted text to `pageTerms`. If `lazy` is true, the PDF is lazily loaded.
func testExtractFileOptions(t *testing.T, filename string, pageTerms map[int][]string, lazy bool) {
	filepath := filepath.Join(corpusFolder, filename)
	exists := checkFileExists(filepath)
	if !exists {
		if forceTest {
			t.Fatalf("filepath=%q does not exist", filepath)
		}
		t.Logf("%s not found", filepath)
		return
	}

	_, actualPageText := extractPageTexts(t, filepath, lazy)
	for _, pageNum := range sortedKeys(pageTerms) {
		expectedTerms, ok := pageTerms[pageNum]
		actualText, ok := actualPageText[pageNum]
		if !ok {
			t.Fatalf("%q doesn't have page %d", filepath, pageNum)
		}
		actualText = norm.NFKC.String(actualText)
		if !containsTerms(t, expectedTerms, actualText) {
			t.Fatalf("Text mismatch filepath=%q page=%d", filepath, pageNum)
		}
	}
}

// extractPageTexts runs ExtractTextWithStats on all pages in PDF `filename` and returns the result
// as a map {page number: page text}
func extractPageTexts(t *testing.T, filename string, lazy bool) (int, map[int]string) {
	f, pdfReader := openPdfReader(t, filename, lazy)
	defer f.Close()

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		t.Fatalf("GetNumPages failed. filename=%q err=%v", filename, err)
	}
	pageText := map[int]string{}
	for pageNum := 1; pageNum <= numPages; pageNum++ {

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			t.Fatalf("GetPage failed. filename=%q page=%d err=%v", filename, pageNum, err)
		}
		ex, err := New(page)
		if err != nil {
			t.Fatalf("extractor.New failed. filename=%q lazy=%t page=%d err=%v",
				filename, lazy, pageNum, err)
		}
		text, _, _, err := ex.ExtractTextWithStats()
		if err != nil {
			t.Fatalf("ExtractTextWithStats failed. filename=%q page=%d err=%v", filename, pageNum, err)
		}
		// TODO(peterwilliams97): Improve text extraction space insertion so we don't need reduceSpaces.
		pageText[pageNum] = reduceSpaces(text)
	}
	return numPages, pageText
}

// textLocTest is a text extraction locations test.
type textLocTest struct {
	filename string
	numPages int
	contents map[int]pageContents
}

// pageNums returns the (1-offset) page numbers that are to be tested in `e`.
func (e textLocTest) pageNums() []int {
	var nums []int
	for pageNum := range e.contents {
		nums = append(nums, pageNum)
	}
	sort.Ints(nums)
	return nums
}

// String returns a description of `e`.
func (e textLocTest) String() string {
	return fmt.Sprintf("{TEXTLOCTEST: filename=%q}", e.filename)
}

// pageContents are the substrings and TextMark that we expect to find in the text extracted from a
// PDF page
type pageContents struct {
	terms    []string                      // Substrings of the extracted text.
	marks    []TextMark                    // TextMarks in the extracted text.
	termBBox map[string]model.PdfRectangle // {term: bounding box of term on PDF page}
}

// matchTerms returns the keys of `c`.termBBox.
func (c pageContents) matchTerms() []string {
	var terms []string
	for w := range c.termBBox {
		terms = append(terms, w)
	}
	sort.Strings(terms)
	return terms
}

// textLocTests are the extracted text location tests. All coordinates are multiples of 0.5 points.
var textLocTests = []textLocTest{
	textLocTest{
		filename: "prop-price-list-2017.pdf",
		numPages: 1,
		contents: map[int]pageContents{
			1: pageContents{
				terms: []string{
					"PRICE LIST",
					"THING ONE", "$99",
					"THING TWO", "$314",
					"THING THREE", "$499",
					"THING FOUR", "$667",
				},
				marks: []TextMark{
					l(0, "P", 165, 725.2, 197.2, 773.2),
					l(1, "R", 197.2, 725.2, 231.9, 773.2),
					l(2, "I", 231.9, 725.2, 245.2, 773.2),
					l(3, "C", 245.2, 725.2, 279.9, 773.2),
					l(4, "E", 279.9, 725.2, 312.0, 773.2),
					l(5, " ", 312.0, 725.2, 325.3, 773.2),
					l(6, "L", 325.3, 725.2, 354.6, 773.2),
					l(7, "I", 354.6, 725.2, 368.0, 773.2),
					l(8, "S", 368.0, 725.2, 400.0, 773.2),
					l(9, "T", 400.0, 725.2, 429.4, 773.2),
				},
				termBBox: map[string]model.PdfRectangle{
					"THING ONE": r(72.0, 534.5, 197.0, 558.5),
				},
			},
		},
	},
	textLocTest{
		filename: "pol_e.pdf",
		numPages: 2,
		contents: map[int]pageContents{
			1: pageContents{
				marks: []TextMark{
					l(3602, "W", 152.5, 185.5, 163.0, 196.5),
					l(3603, "T", 163.0, 185.5, 169.5, 196.5),
					l(3604, "O", 169.5, 185.5, 177.5, 196.5),
				},
				termBBox: map[string]model.PdfRectangle{
					"global public good": r(244.0, 398.5, 332.5, 410.0),
					"international":      r(323.5, 611.0, 377.5, 622.0),
				},
			},
		},
	},
	textLocTest{
		filename: "thanh.pdf",
		numPages: 6,
		contents: map[int]pageContents{
			1: pageContents{
				terms: []string{
					"result is a set of Type 1 fonts that is similar to the Blue Sky fonts",
					"provide Vietnamese letters with the same quality of outlines and hints",
					"Vietnamese letters and VNR fonts",
					"Vietnamese accents can be divided into three the Czech and Polish version of CMR fonts",
					"kinds of diacritic marks: tone, vowel and consonant. about 2 years until the ﬁrst version",
				},
				termBBox: map[string]model.PdfRectangle{
					"the Blue Sky fonts":                       r(358.0, 532.5, 439.0, 542.5),
					"Vietnamese letters with the same quality": r(165.5, 520.5, 344.5, 530.5),
				},
			},
			2: pageContents{
				terms: []string{
					"number of glyphs needed for each font is 47",
					"which 22 are Vietnamese accents and letters.",
					"I would like to point out that I am not a type",
					"shows all the glyphs that need to be converted",
					"designer and therefore the aesthetic aspect of",
					"to Type 1 format.",
				},
				marks: []TextMark{
					l(286, "T", 334.0, 674.5, 341.2, 684.5),
					l(287, "a", 340.5, 674.5, 345.5, 684.5),
					l(288, "k", 345.5, 674.5, 350.5, 684.5),
					l(289, "e", 350.5, 674.5, 355.0, 684.5),
				},
				termBBox: map[string]model.PdfRectangle{
					"glyphs needed for each font": r(382.0, 443.0, 501.0, 453.0),
					"22 are Vietnamese accents":   r(343.5, 431.0, 461.0, 441.0),
				},
			},
		},
	},
	textLocTest{
		filename: "unicodeexample.pdf",
		numPages: 6,
		contents: map[int]pageContents{
			2: pageContents{
				terms: []string{
					"Österreich", "Johann Strauß",
					"Azərbaycan", "Vaqif Səmədoğlu",
					"Азәрбајҹан", "Вагиф Сәмәдоғлу",
				},
				marks: []TextMark{
					l(447, "Ö", 272.0, 521.0, 281.0, 533.0),
					l(448, "s", 281.0, 521.0, 287.0, 533.0),
					l(449, "t", 287.0, 521.0, 290.5, 533.0),
					l(450, "e", 290.5, 521.0, 297.0, 533.0),
					l(451, "r", 297.0, 521.0, 301.0, 533.0),
					l(452, "r", 301.0, 521.0, 305.0, 533.0),
					l(453, "e", 305.0, 521.0, 312.0, 533.0),
					l(454, "i", 312.0, 521.0, 314.5, 533.0),
					l(455, "c", 314.5, 521.0, 320.5, 533.0),
					l(456, "h", 320.5, 521.0, 327.0, 533.0),
				},
				termBBox: map[string]model.PdfRectangle{
					"Österreich": r(272.0, 521.0, 327.0, 533.0), "Johann Strauß": r(400.5, 521.0, 479.5, 533.0),
					"Azərbaycan": r(272.0, 490.5, 335.0, 502.5), "Vaqif Səmədoğlu": r(400.5, 490.5, 492.0, 502.5),
					"Азәрбајҹан": r(272.0, 460.5, 334.5, 472.5), "Вагиф Сәмәдоғлу": r(400.5, 460.5, 501.0, 472.5),
				},
			},
		},
	},
	textLocTest{
		filename: "AF+handout+scanned.pdf",
		numPages: 3,
		contents: map[int]pageContents{
			1: pageContents{
				termBBox: map[string]model.PdfRectangle{
					"reserved": r(505.0, 488.5, 538.5, 497.0),
				},
			},
			2: pageContents{
				termBBox: map[string]model.PdfRectangle{
					"atrium": r(414.5, 113.5, 435.5, 121.0),
				},
			},
			3: pageContents{
				termBBox: map[string]model.PdfRectangle{
					"treatment": r(348.0, 302.0, 388.0, 311.5),
				},
			},
		},
	},
}

// testTextComponent tests TextMark and TextByComponents() functionality. If `lazy` is true then
// PDFs are lazily loaded.
func (e textLocTest) testTextComponent(t *testing.T, lazy bool) {
	desc := fmt.Sprintf("%s lazy=%t", e, lazy)
	common.Log.Debug("textLocTest.testTextComponent: %s", desc)

	filename := filepath.Join(corpusFolder, e.filename)
	f, pdfReader := openPdfReader(t, filename, lazy)
	defer f.Close()
	l := createMarkupList(t, filename, pdfReader)
	defer l.saveOutputPdf()

	n, err := pdfReader.GetNumPages()
	if err != nil {
		t.Fatalf("GetNumPages failed. %s err=%v", desc, err)
	}
	if n != e.numPages {
		t.Fatalf("Wrong number of pages. Expected %d. Got %d. %s",
			e.numPages, n, desc)
	}

	for _, pageNum := range e.pageNums() {
		c := e.contents[pageNum]
		pageDesc := fmt.Sprintf("%s pageNum=%d", desc, pageNum)
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			t.Fatalf("GetPage failed. %s err=%v", pageDesc, err)
		}
		l.setPageNum(pageNum)
		c.testPageTextAndMarks(t, l, pageDesc, page)
	}
}

// testPageTextAndMarks tests that pageTextAndMarks returns extracted page text and TextMarks
// that match the expected results in `c`.
func (c pageContents) testPageTextAndMarks(t *testing.T, l *markupList, desc string,
	page *model.PdfPage) {
	text, textMarks := pageTextAndMarks(t, desc, page)

	// 1) Check that all expected terms are found in `text`.
	for i, term := range c.terms {
		common.Log.Debug("%d: %q", i, term)
		if !strings.Contains(text, term) {
			t.Fatalf("text doesn't contain %q. %s", term, desc)
		}
	}

	// 2) Check that all expected TextMarks are in `textMarks`.
	offsetMark := marksMap(textMarks)
	for i, tm := range c.marks {
		common.Log.Debug("%d: %v", i, tm)
		checkContains(t, desc, offsetMark, tm)
	}

	// 3) Check that locationsIndex() finds TextMarks in `textMarks` corresponding to some
	//   substrings of `text`.
	//   We do this before testing getBBox() below so can narrow down why getBBox() has failed
	//   if it fails.
	testTermMarksMulti(t, text, textMarks)

	// 4) Check that longer terms are matched and found in their expected locations.
	for _, term := range c.matchTerms() {
		expectedBBox := c.termBBox[term]
		bbox, err := getBBox(text, textMarks, term)
		if err != nil {
			t.Fatalf("textMarks doesn't contain term %q. %s", term, desc)
		}
		l.addMatch(term, bbox)
		if !rectEquals(expectedBBox, bbox) {
			t.Fatalf("bbox is wrong - %s\n"+
				"\t    term: %q\n"+
				"\texpected: %v\n"+
				"\t     got: %v",
				desc, term, expectedBBox, bbox)
		}
	}

	// 5) Check out or range cases
	if spanMarks, err := textMarks.RangeOffset(-1, 0); err == nil {
		t.Fatalf("textMarks.RangeOffset(-1, 0) succeeded. %s\n\tspanMarks=%s", desc, spanMarks)
	}
	if spanMarks, err := textMarks.RangeOffset(0, 1e10); err == nil {
		t.Fatalf("textMarks.RangeOffset(0, 1e10) succeeded. %s\n\tspanMarks=%s", desc, spanMarks)
	}
	if spanMarks, err := textMarks.RangeOffset(1, 0); err == nil {
		t.Fatalf("textMarks.RangeOffset(1, 0) succeeded. %s\n\tspanMarks=%s", desc, spanMarks)
	}
}

// testTermMarksFiles stress tests testTermMarksMulti() by running it on all files in the corpus.
func testTermMarksFiles(t *testing.T) {
	pattern := filepath.Join(corpusFolder, "*.pdf")
	pathList, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Glob(%q) failed. err=%v", pattern, err)
	}
	for i, filename := range pathList {
		for _, lazy := range []bool{false, true} {
			common.Log.Info("%4d of %d: %q lazy=%t", i+1, len(pathList), filename, lazy)
			f, pdfReader := openPdfReader(t, filename, lazy)
			defer f.Close()
			numPages, err := pdfReader.GetNumPages()
			if err != nil {
				t.Fatalf("GetNumPages failed. filename=%q err=%v", filename, err)
			}

			for pageNum := 1; pageNum < numPages; pageNum++ {
				desc := fmt.Sprintf("filename=%q pageNum=%d", filename, pageNum)
				page, err := pdfReader.GetPage(pageNum)
				if err != nil {
					t.Fatalf("GetNumPages failed. %s err=%v", desc, err)
				}
				text, textMarks := pageTextAndMarks(t, desc, page)
				testTermMarksMulti(t, text, textMarks)
			}
		}
	}
}

// testTermMarksMulti checks that textMarks.RangeOffset() finds the TextMarks in `textMarks`
// corresponding to some substrings of `text` with lengths 1-20.
func testTermMarksMulti(t *testing.T, text string, textMarks *TextMarkArray) {
	m := len([]rune(text))
	if m > 20 {
		m = 20
	}
	for n := 1; n <= m; n++ {
		testTermMarks(t, text, textMarks, n)
	}
}

// testTermMarks checks that textMarks.RangeOffset() finds the TextMarks in `textMarks`
// corresponding to some substrings of `text` with length `n`.
func testTermMarks(t *testing.T, text string, textMarks *TextMarkArray, n int) {
	if len(text) < 2 {
		return
	}
	common.Log.Debug("testTermMarks: text=%d n=%d", len(text), n)
	runes := []rune(text)
	if n > len(runes)/2 {
		n = len(runes) / 2
	}
	for ofs := 0; ofs < len(runes)-n; ofs++ {
		term := string(runes[ofs : ofs+n])

		// Get TextMarks spanned `term` with RangeOffset().
		spanArray, err := textMarks.RangeOffset(ofs, ofs+n)
		if err != nil {
			t.Fatalf("textMarks.RangeOffset failed term=%q=runes[%d:%d]=%02x err=%v",
				term, ofs, ofs+n, runes[ofs:ofs+n], err)
		}
		if spanArray.Len() == 0 {
			t.Fatalf("No matches. term=%q=runes[%d:%d]=%02x err=%v",
				term, ofs, ofs+n, runes[ofs:ofs+n], err)
		}

		spanMarks := spanArray.Elements()
		mark0 := spanMarks[0]
		mark1 := spanMarks[spanArray.Len()-1]

		if !strings.HasPrefix(term, mark0.Text) {
			t.Fatalf("mark0 is not a prefix for term=%q=runes[%d:%d]=%02x mark0=%v",
				term, ofs, ofs+n, runes[ofs:ofs+n], mark0)
		}
		if !strings.HasSuffix(term, mark1.Text) {
			t.Fatalf("mark1 is not a suffix for term=%q=runes[%d:%d]=%v mark1=%v",
				term, ofs, ofs+n, runes[ofs:ofs+n], mark1)
		}
	}
}

// checkContains checks that `offsetMark` contains `expectedMark`.
// Contains means: `expectedMark`.Offset is in `offsetMark` and for this element (call it tm)
//   tm.Text == expectedMark.Text and the bounding boxes of
//   tm and expectedMark are within `tol` of each other.
func checkContains(t *testing.T, desc string, offsetMark map[int]TextMark, expectedMark TextMark) {
	tm, ok := offsetMark[expectedMark.Offset]
	if !ok {
		t.Fatalf("offsetMark doesn't contain %v - %s", expectedMark, desc)
	}
	if tm.Text != expectedMark.Text {
		t.Fatalf("text doesn't match expected=%q got=%q - %s\n"+
			"\texpected %v\n"+
			"\t     got %v",
			expectedMark.Text, tm.Text, desc, expectedMark, tm)
	}
	if !rectEquals(expectedMark.BBox, tm.BBox) {
		t.Fatalf("Bounding boxes doesn't match  - %s\n"+
			"\texpected %v\n"+
			"\t     got %v",
			desc, expectedMark, tm)
	}
}

// getBBox returns the minimum bounding box around the TextMarks in `textMarks` that correspond to
// the first instance of `term` in `text`, where `text` and `textMarks` are the extracted text
// returned by TextByComponents().
// NOTE: This is how you would use TextByComponents in an application.
func getBBox(text string, textMarks *TextMarkArray, term string) (model.PdfRectangle, error) {
	start := indexRunes(text, term)
	if start < 0 {
		return model.PdfRectangle{}, fmt.Errorf("indexRunes has no match for term=%q", term)
	}
	end := start + len([]rune(term))

	spanMarks, err := textMarks.RangeOffset(start, end)
	if err != nil {
		return model.PdfRectangle{}, err
	}

	bbox, ok := spanMarks.BBox()
	if !ok {
		return model.PdfRectangle{}, fmt.Errorf("spanMarks.BBox has no bounding box. spanMarks=%s",
			spanMarks)
	}
	return bbox, nil
}

// indexRunes returns the index in `text` of the first instance of `term` if `term` is a substring
// of `text`, or -1 if it is not a substring.
// This index is over runes, unlike strings.Index.
func indexRunes(text, term string) int {
	runes := []rune(text)
	substr := []rune(term)
	for i := 0; i < len(runes)-len(substr); i++ {
		matched := true
		for j, r := range substr {
			if runes[i+j] != r {
				matched = false
				break
			}
		}
		if matched {
			return i
		}
	}
	return -1
}

// marksMap returns `textMarks` as a map of TextMarks keyed by TextMark.Offset.
func marksMap(textMarks *TextMarkArray) map[int]TextMark {
	offsetMark := make(map[int]TextMark, textMarks.Len())
	for _, tm := range textMarks.Elements() {
		offsetMark[tm.Offset] = tm
	}
	return offsetMark
}

// tol is the tolerance for matching coordinates. We are specifying coordinates to the nearest 0.5
// point so the tolerance should be just over 0.5
const tol = 0.5001

// l is a shorthand for writing TextMark literals, which get verbose in Go,
func l(o int, t string, llx, lly, urx, ury float64) TextMark {
	return TextMark{Offset: o, BBox: r(llx, lly, urx, ury), Text: t}
}

// r is a shorthand for writing model.PdfRectangle literals, which get verbose in Go,
func r(llx, lly, urx, ury float64) model.PdfRectangle {
	return model.PdfRectangle{Llx: llx, Lly: lly, Urx: urx, Ury: ury}
}

// pageTextAndMarks returns the extracted page text and TextMarks for PDF page `page`.
func pageTextAndMarks(t *testing.T, desc string, page *model.PdfPage) (string, *TextMarkArray) {
	ex, err := New(page)
	if err != nil {
		t.Fatalf("extractor.New failed. %s err=%v", desc, err)
	}
	common.Log.Debug("pageTextAndMarks: %s", desc)
	pageText, _, _, err := ex.ExtractPageText()
	if err != nil {
		t.Fatalf("ExtractPageText failed. %s err=%v", desc, err)
	}

	text := pageText.Text()
	textMarks := pageText.Marks()

	{ // Some extra debugging to see how the code works. Not needed by test.
		common.Log.Debug("text=>>>%s<<<\n", text)
		common.Log.Debug("textMarks=%s %q", textMarks, desc)
		for i, tm := range textMarks.Elements() {
			common.Log.Debug("%6d: %d %q=%02x %v", i, tm.Offset, tm.Text, []rune(tm.Text), tm.BBox)
		}
	}

	return text, textMarks
}

// openPdfReader returns a file handle and a PdfReader for file `filename`. If `lazy` is true, it
// will be lazy reader.
// The file handle needs to be returned so the caller can hold it open while the PdfReader is being
// used in the lazy case.
func openPdfReader(t *testing.T, filename string, lazy bool) (*os.File, *model.PdfReader) {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Couldn't open filename=%q err=%v", filename, err)
	}

	var pdfReader *model.PdfReader
	if lazy {
		pdfReader, err = model.NewPdfReaderLazy(f)
		if err != nil {
			t.Fatalf("NewPdfReaderLazy failed. filename=%q err=%v", filename, err)
		}
	} else {
		pdfReader, err = model.NewPdfReader(f)
		if err != nil {
			t.Fatalf("NewPdfReader failed. filename=%q err=%v", filename, err)
		}
	}
	return f, pdfReader
}

// containsTerms returns true if all strings `terms` are contained in `actualText`.
func containsTerms(t *testing.T, terms []string, actualText string) bool {
	for _, w := range terms {
		w = norm.NFKC.String(w)
		if !strings.Contains(actualText, w) {
			t.Errorf("No match for %q", w)
			return false
		}
	}
	return true
}

// reduceSpaces returns `text` with runs of spaces of any kind (spaces, tabs, line breaks, etc)
// reduced to a single space.
func reduceSpaces(text string) string {
	text = reSpace.ReplaceAllString(text, " ")
	return strings.Trim(text, " \t\n\r\v")
}

var reSpace = regexp.MustCompile(`(?m)\s+`)

// checkFileExists returns true if `filepath` exists.
func checkFileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

// sortedKeys returns the keys of `m` as a sorted slice.
func sortedKeys(m map[int][]string) []int {
	keys := []int{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// rectEquals returns true if `b1` and `b2` corners are within `tol` of each other.
// NOTE: All the coordinates in this source file are in points.
func rectEquals(b1, b2 model.PdfRectangle) bool {
	return math.Abs(b1.Llx-b2.Llx) <= tol &&
		math.Abs(b1.Lly-b2.Lly) <= tol &&
		math.Abs(b1.Urx-b2.Urx) <= tol &&
		math.Abs(b1.Ury-b2.Ury) <= tol
}

// markupList saves the results of text searches so they can be used to mark-up a PDF with search
// matches and show the search term that was matched.
// Marked up results are saved in markupDir if markupPDFs is true.
// The PDFs in markupDir can be viewed in a PDF viewer to check that they correct.
type markupList struct {
	inPath      string           // Name of input PDF that was searched
	pageMatches map[int][]match  // {pageNum: matches on page}
	t           *testing.T       // Testing context
	pdfReader   *model.PdfReader // Reader for input PDF
	pageNum     int              // (1-offset) Page number being worked on.
}

// match is a match of search term `Term` on a page. `BBox` is the bounding box around the matched
// term on the PDF page
type match struct {
	Term string
	BBox model.PdfRectangle
}

// String returns a description of `l`.
func (l markupList) String() string {
	return fmt.Sprintf("%q: %d pages. Input pageNums=%v", l.inPath, len(l.pageMatches), l.pageNums())
}

// createMarkupList returns an initialized markupList for saving match results to so the bounding
// boxes can be checked for accuracy in a PDF viewer.
func createMarkupList(t *testing.T, inPath string, pdfReader *model.PdfReader) *markupList {
	return &markupList{
		t:           t,
		inPath:      inPath,
		pdfReader:   pdfReader,
		pageMatches: map[int][]match{},
	}
}

// setPageNum sets the page number that markupList.addMatch() will use to add to matches to.
func (l *markupList) setPageNum(pageNum int) {
	l.pageNum = pageNum
}

// addMatch added a match on search term `term` that was found to have bounding box `bbox` to
// for `l`.pageNum. l.pageNum is set with markupList.setPageNum()
func (l *markupList) addMatch(term string, bbox model.PdfRectangle) {
	m := match{Term: term, BBox: bbox}
	l.pageMatches[l.pageNum] = append(l.pageMatches[l.pageNum], m)
}

// pageNums returns the (1-offset) page numbers in `l` of pages that have searc matches
func (l *markupList) pageNums() []int {
	var nums []int
	for pageNum, matches := range l.pageMatches {
		if len(matches) == 0 {
			continue
		}
		nums = append(nums, pageNum)
	}
	sort.Ints(nums)
	return nums
}

// saveOutputPdf is called to mark-up a PDF file with the locations of text.
// `l` contains the input PDF, the pages, search terms and bounding boxes to mark.
func (l *markupList) saveOutputPdf() {
	if !markupPDFs {
		return
	}
	if len(l.pageNums()) == 0 {
		common.Log.Info("No marked-up PDFs to save")
		return
	}
	common.Log.Info("Saving marked-up PDFs. %s", l)

	os.Mkdir(markupDir, 0777)
	outPath := filepath.Join(markupDir, filepath.Base(l.inPath))
	ext := path.Ext(outPath)
	metaPath := outPath[:len(outPath)-len(ext)] + ".json"

	// Make a new PDF creator.
	c := creator.New()

	for _, pageNum := range l.pageNums() {

		common.Log.Debug("saveOutputPdf: %a pageNum=%d", l.inPath, pageNum)
		page, err := l.pdfReader.GetPage(pageNum)
		if err != nil {
			l.t.Fatalf("saveOutputPdf: Could not get page  pageNum=%d. err=%v", pageNum, err)
		}
		mediaBox, err := page.GetMediaBox()
		if err == nil && page.MediaBox == nil {
			// Deal with MediaBox inherited from Parent.
			common.Log.Info("MediaBox: %v -> %v", page.MediaBox, mediaBox)
			page.MediaBox = mediaBox
		}
		h := mediaBox.Ury

		if err := c.AddPage(page); err != nil {
			l.t.Fatalf("AddPage failed %s:%d err=%v ", l, pageNum, err)
		}

		for _, m := range l.pageMatches[pageNum] {
			r := m.BBox
			rect := c.NewRectangle(r.Llx, h-r.Lly, r.Urx-r.Llx, -(r.Ury - r.Lly))
			rect.SetBorderColor(creator.ColorRGBFromHex("#0000ff")) // Blue border.
			rect.SetBorderWidth(1.0)
			if err := c.Draw(rect); err != nil {
				l.t.Fatalf("Draw failed. pageNum=%d match=%v err=%v", pageNum, m, err)
			}
		}
	}

	c.SetOutlineTree(l.pdfReader.GetOutlineTree())
	if err := c.WriteToFile(outPath); err != nil {
		l.t.Fatalf("WriteToFile failed. err=%v", err)
	}

	b, err := json.MarshalIndent(l.pageMatches, "", "\t")
	if err != nil {
		l.t.Fatalf("MarshalIndent failed. err=%v", err)
	}
	err = ioutil.WriteFile(metaPath, b, 0666)
	if err != nil {
		l.t.Fatalf("WriteFile failed. metaPath=%q err=%v", metaPath, err)
	}
}
