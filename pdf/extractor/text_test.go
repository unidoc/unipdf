/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model"
)

// NOTE: We do a best effort at finding the PDF file because we don't keep PDF test files in this repo so you
// will need to setup UNIDOC_EXTRACT_TESTDATA to point at the corpus directory.

// forceTest should be set to true to force running all tests.
// NOTE: Setting environment variable UNIDOC_EXTRACT_FORCETEST = 1 sets this to true.
var forceTest = os.Getenv("UNIDOC_EXTRACT_FORCETEST") == "1"

var corpusFolder = os.Getenv("UNIDOC_EXTRACT_TESTDATA")

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelError))
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
		courier := model.NewStandard14FontMustCompile(model.Courier)
		helvetica := model.NewStandard14FontMustCompile(model.Helvetica)
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
		t.Run(test.filename, func(t *testing.T) {
			testExtractFile(t, test.filename, test.expectedPageText)
		})
	}
}

// fileExtractionTests are the PDFs and texts we are looking for on specified pages.
var fileExtractionTests = []struct {
	filename         string
	expectedPageText map[int][]string
}{
	{filename: "reader.pdf",
		expectedPageText: map[int][]string{
			1: []string{"A Research UNIX Reader:",
				"Annotated Excerpts from the Programmer’s Manual,",
				"1. Introduction",
				"To keep the size of this report",
				"last common ancestor of a radiative explosion",
			},
		},
	},
	{filename: "000026.pdf",
		expectedPageText: map[int][]string{
			1: []string{"Fresh Flower",
				"Care & Handling ",
			},
		},
	},
	{filename: "search_sim_key.pdf",
		expectedPageText: map[int][]string{
			2: []string{"A cryptographic scheme which enables searching",
				"Untrusted server should not be able to search for a word without authorization",
			},
		},
	},
	{filename: "Theil_inequality.pdf",
		expectedPageText: map[int][]string{
			1: []string{"London School of Economics and Political Science"},
			4: []string{"The purpose of this paper is to set Theil’s approach"},
		},
	},
	{filename: "8207.pdf",
		expectedPageText: map[int][]string{
			1: []string{"In building graphic systems for use with raster devices,"},
			2: []string{"The imaging model specifies how geometric shapes and colors are"},
			3: []string{"The transformation matrix T that maps application defined"},
		},
	},
	{filename: "ling-2013-0040ad.pdf",
		expectedPageText: map[int][]string{
			1: []string{"Although the linguistic variation among texts is continuous"},
			2: []string{"distinctions. For example, much of the research on spoken/written"},
		},
	},
	{filename: "26-Hazard-Thermal-environment.pdf",
		expectedPageText: map[int][]string{
			1: []string{"OHS Body of Knowledge"},
			2: []string{"Copyright notice and licence terms"},
		},
	},
	{filename: "Threshold_survey.pdf",
		expectedPageText: map[int][]string{
			1: []string{"clustering, entropy, object attributes, spatial correlation, and local"},
		},
	},
	{filename: "Ito_Formula.pdf",
		expectedPageText: map[int][]string{
			// 1: []string{"In the Itô stochastic calculus"},
			1: []string{"In standard, non-stochastic calculus, one computes a derivative"},
		},
	},
	{filename: "circ2.pdf",
		expectedPageText: map[int][]string{
			1: []string{"Understanding and complying with copyright law can be a challenge"},
		},
	},
	{filename: "rare_word.pdf",
		expectedPageText: map[int][]string{
			6: []string{"words in the test set, we increase the BLEU score"},
		},
	},
	{filename: "Planck_Wien.pdf",
		expectedPageText: map[int][]string{
			1: []string{"entropy of a system of n identical resonators in a stationary radiation field"},
		},
	},
}

// testExtractFile tests the ExtractTextWithStats text extractor on `filename` and compares the extracted
// text to `expectedPageText`.
//
// NOTE: We do a best effort at finding the PDF file because we don't keep PDF test files in this repo
// so you will need to set the environment variable UNIDOC_EXTRACT_TESTDATA to point at
// the corpus directory.
//
// If `filename` cannot be found in `corpusFolders` then the test is skipped unless `forceTest` global
// variable is true (e.g. setting environment variable UNIDOC_EXTRACT_FORCETESTS = 1).
func testExtractFile(t *testing.T, filename string, expectedPageText map[int][]string) {
	filepath := filepath.Join(corpusFolder, filename)
	exists := checkFileExists(filepath)
	if !exists {
		if forceTest {
			t.Fatalf("filename=%q does not exist", filename)
		}
		t.Logf("%s not found", filename)
		return
	}

	_, actualPageText := extractPageTexts(t, filepath)
	for _, pageNum := range sortedKeys(expectedPageText) {
		expectedSentences, ok := expectedPageText[pageNum]
		actualText, ok := actualPageText[pageNum]
		if !ok {
			t.Fatalf("%q doesn't have page %d", filename, pageNum)
		}
		if !containsSentences(t, expectedSentences, actualText) {
			t.Fatalf("Text mismatch filepath=%q page=%d", filepath, pageNum)
		}
	}
}

// extractPageTexts runs ExtractTextWithStats on all pages in PDF `filename` and returns the result as a map
// {page number: page text}
func extractPageTexts(t *testing.T, filename string) (int, map[int]string) {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Couldn't open filename=%q err=%v", filename, err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Fatalf("NewPdfReader failed. filename=%q err=%v", filename, err)
	}
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
			t.Fatalf("extractor.New failed. filename=%q page=%d err=%v", filename, pageNum, err)
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

// containsSentences returns true if all strings `expectedSentences` are contained in `actualText`.
func containsSentences(t *testing.T, expectedSentences []string, actualText string) bool {
	for _, e := range expectedSentences {
		if !strings.Contains(actualText, e) {
			t.Errorf("No match for %#q", e)
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
