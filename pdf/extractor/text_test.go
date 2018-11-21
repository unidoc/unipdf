/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"flag"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelError))
	if flag.Lookup("test.v") != nil {
		isTesting = true
	}
}

const testContents1 = `
    BT
    /UniDocCourier 24 Tf
    (Hello World!)Tj
    0 -10 Td
    (Doink)Tj
    ET
`

const testExpected1 = "Hello World!\nDoink"

func TestTextExtraction1(t *testing.T) {
	e := Extractor{}
	e.contents = testContents1

	s, err := e.ExtractText()
	if err != nil {
		t.Errorf("Error extracting text: %v", err)
		return
	}
	if s != testExpected1 {
		t.Errorf("Text mismatch. Got %q. Expected %q", s, testExpected1)
		return
	}
}

func TestTextExtraction2(t *testing.T) {
	for _, test := range extract2Tests {
		testExtract2(t, test.filename, test.expectedPageText)
	}
}

var extract2Tests = []struct {
	filename         string
	expectedPageText map[int][]string
}{
	{filename: "testdata/reader.pdf",
		expectedPageText: map[int][]string{
			1: []string{"A Research UNIX Reader:",
				"Annotated Excerpts from the Programmer’s Manual,",
				"1. Introduction",
				"To keep the size of this report",
				"last common ancestor of a radiative explosion",
			},
		},
	},
	{filename: "testdata/000026.pdf",
		expectedPageText: map[int][]string{
			1: []string{"Fresh Flower",
				"Care & Handling ",
			},
		},
	},
	{filename: "testdata/search_sim_key.pdf",
		expectedPageText: map[int][]string{
			2: []string{"A cryptographic scheme which enables searching",
				"Untrusted server should not be able to search for a word without authorization",
			},
		},
	},
	{filename: "testdata/Theil_inequality.pdf",
		expectedPageText: map[int][]string{
			1: []string{"London School of Economics and Political Science"},
			4: []string{"The purpose of this paper is to set Theil’s approach"},
		},
	},
	{filename: "testdata/8207.pdf",
		expectedPageText: map[int][]string{
			1: []string{"In building graphic systems for use with raster devices,"},
			2: []string{"The imaging model specifies how geometric shapes and colors are"},
			3: []string{"The transformation matrix T that maps application defined"},
		},
	},
	{filename: "testdata/ling-2013-0040ad.pdf",
		expectedPageText: map[int][]string{
			1: []string{"Although the linguistic variation among texts is continuous"},
			2: []string{"distinctions. For example, much of the research on spoken/written"},
		},
	},
	{filename: "testdata/26-Hazard-Thermal-environment.pdf",
		expectedPageText: map[int][]string{
			1: []string{"OHS Body of Knowledge"},
			2: []string{"Copyright notice and licence terms"},
		},
	},
}

func testExtract2(t *testing.T, filename string, expectedPageText map[int][]string) {
	_, actualPageText := extractPageTexts(t, filename)
	for _, pageNum := range sortedKeys(expectedPageText) {
		expectedSentences, ok := expectedPageText[pageNum]
		actualText, ok := actualPageText[pageNum]
		if !ok {
			t.Fatalf("%q doesn't have page %d", filename, pageNum)
		}
		if !containsSentences(t, expectedSentences, actualText) {
			t.Fatalf("Text mismatch filename=%q page=%d", filename, pageNum)
		}
	}
}

func containsSentences(t *testing.T, expectedSentences []string, actualText string) bool {
	for _, e := range expectedSentences {
		if !strings.Contains(actualText, e) {
			t.Errorf("No match for %+q", e)
			return false
		}
	}
	return true
}

func sortedKeys(m map[int][]string) []int {
	keys := []int{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

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
		pageText[pageNum] = reduceSpaces(text)
	}
	return numPages, pageText
}

// reduceSpaces returns `text` with runs of spaces of any kind (spaces, tabs, line breaks, etc)
// reduced to a single space.
func reduceSpaces(text string) string {
	text = reSpace.ReplaceAllString(text, " ")
	return strings.Trim(text, " \t\n\r\v")
}

var reSpace = regexp.MustCompile(`(?m)\s+`)
