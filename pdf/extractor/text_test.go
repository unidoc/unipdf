/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"flag"
	"testing"
)

func init() {
	if flag.Lookup("test.v") != nil {
		isTesting = true
	}
}

const testContents1 = `
BT
/F1 24 Tf
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
		t.Errorf("Text mismatch (%s)", s)
		t.Errorf("Text mismatch (% X vs % X)", s, testExpected1)
		return
	}
}
