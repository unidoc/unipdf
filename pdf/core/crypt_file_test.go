/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Integration tests for the PDF crypt support.

package core_test

import (
	"os"
	"path/filepath"
	"testing"

	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

const aes3Dir = `./testdata`

func TestDecryptAES3(t *testing.T) {
	cases := []struct {
		file  string
		pass  string
		R     int
		pages int
		page1 string
	}{
		// See https://github.com/mozilla/pdf.js/issues/6010
		{
			file: "issue6010_1.pdf", pass: "abc", R: 6, pages: 1,
			page1: "\nIssue 6010",
		},
		{
			file: "issue6010_2.pdf", pass: "æøå", R: 6, pages: 10,
			page1: "\nSample PDF Document\nRobert Maron\nGrzegorz Grudzi\n\xb4\nnski\nFebruary 20, 1999",
		},
		// See https://github.com/mozilla/pdf.js/pull/6531
		{
			file: "pr6531_1.pdf", pass: "asdfasdf", R: 6, pages: 1,
		},
		{
			file: "pr6531_2.pdf", pass: "asdfasdf", R: 6, pages: 1,
		},
		// See https://github.com/sumatrapdfreader/sumatrapdf/issues/294
		{
			file: "testcase_encry.pdf", pass: "123", R: 5, pages: 1, // owner pass
			page1: "\n\x00\x01\x00\x02\x00\x03\x00\x04\x00\x05\x00\x06\x00\a\x00\b\n\x00\x01\n\x00\t\x00\n\x00\v",
		},
		{
			file: "testcase_encry.pdf", pass: "456", R: 5, pages: 1, // user pass
			page1: "\n\x00\x01\x00\x02\x00\x03\x00\x04\x00\x05\x00\x06\x00\a\x00\b\n\x00\x01\n\x00\t\x00\n\x00\v",
		},
		{
			file: "x300.pdf", R: 5, pages: 1,
			pass:  "rnofajrcudiaplhafbqrkrafphehjlvctmwftvpzvachsulmfkjltliftbfpgabustkjfybeqvwgdfawyghoijxgwuxkkrywybpapsswxcnigwwnpttgvfxtrlnbqzberhrnelvcqjaasothqhtzjoxqttlqrmxfqawyhizoslazxhdqffiweruqjrmpdsxutvevceaormydxhregsadphblbaziucrnsbntzptdzfkzfzlwmxhslywusuajwspvabqwopbxdttwbjappgiaxrkgmsuodkzhbqvqiwummcdu",
			page1: " \nTemplate form for pdf_form_add.go\t \nThis PDF is explicitly created as a template\t \tfor adding\t \ta PDF interactive form to.\t \n \nFull \tName: _________________________________________\t \nAddress\t \tLine 1\t: \t__________________\t________________\t____\t \nAddress\t \tLine \t2\t: ________________\t_______\t___________\t____\t \nAge: ______\t \nGender: \t  \t[ ] Male    [ ] Female\t \nCity: ______________\t \nCountry: ______________\t \nFavorite Color:\t \t \t___________________\t \n \n \n ",
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.file, func(t *testing.T) {
			f, err := os.Open(filepath.Join(aes3Dir, c.file))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			p, err := pdf.NewPdfReader(f)
			if err != nil {
				t.Fatal(err)
			}
			if ok, err := p.IsEncrypted(); err != nil {
				t.Fatal(err)
			} else if !ok {
				t.Fatal("document is not encrypted")
			}
			ok, err := p.Decrypt([]byte(c.pass))
			if err != nil {
				t.Fatal(err)
			} else if !ok {
				t.Fatal("wrong password")
			}

			numPages, err := p.GetNumPages()
			if err != nil {
				t.Fatal(err)
			} else if numPages != c.pages {
				t.Errorf("wrong number of pages: %d", numPages)
			}

			page, err := p.GetPage(1)
			if err != nil {
				t.Fatal(err)
			}

			streams, err := page.GetContentStreams()
			if err != nil {
				t.Fatal(err)
			}

			content := ""
			for _, cstream := range streams {
				content += cstream
			}

			cstreamParser := pdfcontent.NewContentStreamParser(content)
			txt, err := cstreamParser.ExtractText()
			if err != nil {
				t.Fatal(err)
			} else if txt != c.page1 {
				t.Fatalf("wrong text: %q", txt)
			}
		})
	}
}
