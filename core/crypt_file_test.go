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

	pdfcontent "github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	pdf "github.com/unidoc/unipdf/v3/model"
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
		{
			file: "i-9.pdf", R: 4, pages: 5,
			pass:  "",
			page1: "Department of Homeland Security \nU.S. Citizenship and Immigration Services\tForm I-9, Employment \nEligibility Verification\nAnti-Discrimination Notice.\t It is illegal to discriminate against \nany individual (other than an alien not authorized to work in the  \nUnited States) in hiring, discharging, or recruiting or referring for a \nfee because of that individual's national origin or citizenship status. \nIt is illegal to discriminate against work-authorized individuals. \nEmployers CANNOT specify which document(s) they will accept \nfrom an employee. The refusal to hire an individual because the \ndocuments presented have a future expiration date may also \nconstitute illegal discrimination. For more information, call the \nOffice of Special Counsel for Immigration Related Unfair \nEmployment Practices at 1-800-255-8155.\nAll employees (citizens and noncitizens) hired after November \n6, 1986, and working in the United States must complete \nForm I-9.\tOMB No. 1615-0047; Expires 08/31/12The Preparer/Translator Certification must be completed if\n \nSection 1 is prepared by a person other than the employee. A \npreparer/translator may be used only when the employee is \nunable to complete \nSection 1 on his or her own. However, the \nemployee must still sign \nSection 1 personally.\nForm I-9 (Rev. 08/07/09) Y Read all instructions carefully before completing this form.  InstructionsWhen Should Form I-9 Be Used?What Is the Purpose of This Form?The purpose of this form is to document that each new \nemployee (both citizen and noncitizen) hired after November \n6, 1986, is authorized to work in the United States.\nFor the purpose of completing this form, the term \"employer\" \nmeans all employers including those recruiters and referrers \nfor a fee who are agricultural associations, agricultural \nemployers, or farm labor contractors.  Employers must \ncomplete \nSection 2 by examining evidence of identity and \nemployment authorization within three business days of the \ndate employment begins. However, if an employer hires an \nindividual for less than three business days, \nSection 2 must be \ncompleted at the time employment begins. Employers cannot \nspecify which document(s) listed on the last page of Form I-9 \nemployees present to establish identity and employment \nauthorization. Employees may present any List A document \nOR a combination of a List B and a List C document.Filling Out Form I-9This part of the form must be completed no later than the time \nof hire, which is the actual beginning of employment. \nProviding the Social Security Number is voluntary, except for \nemployees hired by employers participating in the USCIS \nElectronic Employment Eligibility Verification Program (E-\nVerify). The employer is responsible for ensuring that \nSection 1 is timely and properly completed.\n1.  Document title;\n2.  Issuing authority;\n3.  Document number;\n4.  Expiration date, if any; and \n5.  The date employment begins. \nEmployers must sign and date the certification in \nSection 2. \nEmployees must present original documents. Employers may, \nbut are not required to, photocopy the document(s) presented. \nIf photocopies are made, they must be made for all new hires. \nPhotocopies may only be used for the verification process and \nmust be retained with Form I-9. \nEmployers are still \nresponsible for completing and retaining Form I-9.Noncitizen nationals of the United States are persons born in \nAmerican Samoa, certain former citizens of the former Trust \nTerritory of the Pacific Islands, and certain children of \nnoncitizen nationals born abroad.\nEmployers should note the work authorization expiration \ndate (if any) shown in \nSection 1. For employees who indicate \nan employment authorization expiration date in \nSection 1, \nemployers are required to reverify employment authorization \nfor employment on or before the date shown. Note that some \nemployees may leave the expiration date blank if they are \naliens whose work authorization does not expire (e.g., asylees, \nrefugees, certain citizens of the Federated States of Micronesia \nor the Republic of the Marshall Islands). For such employees, \nreverification does not apply unless they choose to present\nIf an employee is unable to present a required document (or \ndocuments), the employee must present an acceptable receipt \nin lieu of a document listed on the last page of this form. \nReceipts showing that a person has applied for an initial grant \nof employment authorization, or for renewal of employment \nauthorization, are not acceptable. Employees must present \nreceipts within three business days of the date employment \nbegins and must present valid replacement documents within \n90 days or other specified time.\nEmployers must record in Section 2:\nPreparer/Translator Certification\nSection 2, Employer Section 1, Employee\nin Section 2 evidence of employment authorization that \ncontains an expiration date (e.g., Employment Authorization \nDocument (Form I-766)).",
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
				t.Errorf("wrong text: %q", txt)
			}

			for _, objNum := range p.GetObjectNums() {
				obj, err := p.GetIndirectObjectByNumber(objNum)
				if err != nil {
					t.Fatal(objNum, err)
				}
				if stream, is := obj.(*core.PdfObjectStream); is {
					_, err := core.DecodeStream(stream)
					if err != nil {
						t.Fatal(err)
					}
				} else if indObj, is := obj.(*core.PdfIndirectObject); is {
					_ = indObj.PdfObject.String()
				}
			}
		})
	}
}
