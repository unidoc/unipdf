/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model"
)

// XXX(peterwilliams97) NOTE: We do a best effort at finding the PDF file because we don't keep PDF
// test files in this repo so you will need to setup `corpusFolders` to point at the corpus directory.

// forceTest should be set to true to force running all tests.
const forceTest = false

// corpusFolders is where we search for test files.
var corpusFolders = []string{
	"./testdata",
	"~/testdata",
	".",
}

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelError))
	if flag.Lookup("test.v") != nil {
		isTesting = true
	}
}

// TestTextExtraction1 tests text extraction on the PDF fragments in `fragmentTests`.
func TestTextExtraction1(t *testing.T) {
	for _, f := range fragmentTests {
		f.testExtraction(t)
	}
}

type fragment struct {
	name     string
	contents string
	text     string
}

var fragmentTests = []fragment{

	{name: "portrait",
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
	{name: "landscape",
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
	{name: "180 degree rotation",
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
	{name: "Helvetica",
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

// testExtraction checks that ExtractText() works on fragment `f`.
func (f fragment) testExtraction(t *testing.T) {
	e := Extractor{contents: f.contents}
	text, err := e.ExtractText()
	if err != nil {
		t.Fatalf("Error extracting text: %q err=%v", f.name, err)
		return
	}
	if text != f.text {
		t.Fatalf("Text mismatch: %q Got %q. Expected %q", f.name, text, f.text)
		return
	}
}

// TestTextExtraction2 tests text extraction on set of PDF files.
// It checks for the existence of specified strings of words on specified pages.
// We currently only check within lines as our line order is still improving.
func TestTextExtraction2(t *testing.T) {
	for _, test := range extract2Tests {
		testExtract2(t, test.filename, test.expectedPageText)
	}
}

// extract2Tests are the PDFs and texts we are looking for on specified pages.
var extract2Tests = []struct {
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

// testExtract2 tests the ExtractText2 text extractor on `filename` and compares the extracted
// text to `expectedPageText`.
// XXX(peterwilliams97) NOTE: We do a best effort at finding the PDF file because we don't keep PDF
// test files in this repo so you will need to setup `corpusFolders` to point at the corpus directory.
// If `filename` cannot be found in `corpusFolders` then the test is skipped.
func testExtract2(t *testing.T, filename string, expectedPageText map[int][]string) {
	homeDir, hasHome := getHomeDir()
	path, ok := searchDirectories(homeDir, hasHome, corpusFolders, filename)
	if !ok {
		if forceTest {
			t.Fatalf("filename=%q does not exist", filename)
		}
		return
	}
	_, actualPageText := extractPageTexts(t, path)
	for _, pageNum := range sortedKeys(expectedPageText) {
		expectedSentences, ok := expectedPageText[pageNum]
		actualText, ok := actualPageText[pageNum]
		if !ok {
			t.Fatalf("%q doesn't have page %d", filename, pageNum)
		}
		if !containsSentences(t, expectedSentences, actualText) {
			t.Fatalf("Text mismatch filename=%q page=%d", path, pageNum)
		}
	}
}

// extractPageTexts runs ExtractText2 on all pages in PDF `filename` and returns the result as a map
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
		text, _, _, err := ex.ExtractText2()
		if err != nil {
			t.Fatalf("ExtractText2 failed. filename=%q page=%d err=%v", filename, pageNum, err)
		}
		// XXX(peterwilliams97)TODO: Improve text extraction space insertion so we don't need reduceSpaces.
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

// searchDirectories searches `directories` for `filename` and returns the full file path if it is
// found. `homeDir` and `hasHome` are used for home directory substitution.
func searchDirectories(homeDir string, hasHome bool, directories []string, filename string) (string, bool) {
	for _, direct := range directories {
		if hasHome {
			direct = strings.Replace(direct, "~", homeDir, 1)
		}
		path := filepath.Join(direct, filename)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

// getHomeDir returns the current user's home directory if it is defined and a bool to tell if it
// is defined.
func getHomeDir() (string, bool) {
	usr, err := user.Current()
	if err != nil {
		common.Log.Error("No current user. err=%v", err)
		return "", false
	}
	return usr.HomeDir, true
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
