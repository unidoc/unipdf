/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"crypto/rsa"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
	"golang.org/x/crypto/pkcs12"
)

// This test file contains multiple tests to generate PDFs from existing Pdf files. The outputs are written into /tmp as files.  The files
// themselves need to be observed to check for correctness as we don't have a good way to automatically check
// if every detail is correct.

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
}

const testPdfFile1 = "./testdata/minimal.pdf"
const testPdfLoremIpsumFile = "./testdata/lorem.pdf"
const testPdf3pages = "./testdata/pages3.pdf"

const imgPdfFile1 = "./testdata/img1-1.pdf"
const imgPdfFile2 = "./testdata/img1-2.pdf"

// source http://foersom.com/net/HowTo/data/OoPdfFormExample.pdf
const testPdfAcroFormFile1 = "./testdata/OoPdfFormExample.pdf"

const testPdfSignedPDFDocument = "./testdata/SampleSignedPDFDocument.pdf"

const testPKS12Key = "./testdata/ks12"
const testPKS12KeyPassword = "password"

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}

func TestAppenderAddPage(t *testing.T) {
	f1, err := os.Open(testPdfLoremIpsumFile)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	f2, err := os.Open(testPdfFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f2.Close()
	pdf2, err := NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender.AddPages(pdf1.PageList...)

	appender.AddPages(pdf2.PageList...)
	appender.AddPages(pdf2.PageList...)

	err = appender.WriteToFile(tempFile("appender_add_page_1.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestAppenderAddPage2(t *testing.T) {
	f1, err := os.Open(testPdfLoremIpsumFile)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	f2, err := os.Open(testPdfAcroFormFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f2.Close()
	pdf2, err := NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender.AddPages(pdf2.PageList...)

	appender.ReplaceAcroForm(pdf2.AcroForm)

	err = appender.WriteToFile(tempFile("appender_add_page_2.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestAppenderRemovePage(t *testing.T) {
	f1, err := os.Open(testPdf3pages)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender.RemovePage(1)
	appender.RemovePage(2)

	err = appender.WriteToFile(tempFile("appender_remove_page_1.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestAppenderReplacePage(t *testing.T) {
	f1, err := os.Open(testPdf3pages)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	f2, err := os.Open(testPdfFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f2.Close()
	pdf2, err := NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender.ReplacePage(1, pdf1.PageList[1])
	appender.ReplacePage(3, pdf2.PageList[0])

	err = appender.WriteToFile(tempFile("appender_replace_page_1.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestAppenderAddAnnotation(t *testing.T) {
	f1, err := os.Open(testPdf3pages)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	page := pdf1.PageList[0]
	annotation := NewPdfAnnotationSquare()
	rect := PdfRectangle{Ury: 250.0, Urx: 150.0, Lly: 50.0, Llx: 50.0}
	annotation.Rect = rect.ToPdfObject()
	annotation.IC = core.MakeArrayFromFloats([]float64{4.0, 0.0, 0.3})
	annotation.CA = core.MakeFloat(0.5)

	page.Annotations = append(page.Annotations, annotation.PdfAnnotation)

	appender.ReplacePage(1, page)

	err = appender.WriteToFile(tempFile("appender_add_annotation_1.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestAppenderMergePage(t *testing.T) {
	f1, err := os.Open(testPdf3pages)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	f2, err := os.Open(testPdfFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f2.Close()
	pdf2, err := NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = appender.MergePageWith(1, pdf2.PageList[0])
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = appender.WriteToFile(tempFile("appender_merge_page_1.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestAppenderMergePage2(t *testing.T) {
	f1, err := os.Open(imgPdfFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()

	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	f2, err := os.Open(imgPdfFile2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f2.Close()

	pdf2, err := NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = appender.MergePageWith(1, pdf2.PageList[0])
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender.AddPages(pdf2.PageList...)

	err = appender.WriteToFile(tempFile("appender_merge_page_2.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func TestAppenderMergePage3(t *testing.T) {
	f1, err := os.Open(testPdfFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	f2, err := os.Open(testPdfLoremIpsumFile)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f2.Close()

	pdf2, err := NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = appender.MergePageWith(1, pdf2.PageList[0])
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = appender.WriteToFile(tempFile("appender_merge_page_3.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

func validateFile(t *testing.T, fileName string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	reader, err := NewPdfReader(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	handler, _ := NewAdobeX509RSASHA1SignatureHandler(nil, nil)
	handler2, _ := NewAdobePKCS7DetachedSignatureHandler(nil, nil)
	handlers := []SignatureHandler{handler, handler2}

	res, err := reader.Validate(handlers)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	if len(res) == 0 {
		t.Errorf("Fail: signature fields not found")
		return
	}

	if !res[0].IsSigned || !res[0].IsVerified {
		t.Errorf("Fail: validation failed")
		return
	}
}

func TestAppenderSignPage4(t *testing.T) {
	// TODO move to reader_test.go
	validateFile(t, testPdfSignedPDFDocument)

	f1, err := os.Open(testPdfFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	f, _ := ioutil.ReadFile(testPKS12Key)
	privateKey, cert, err := pkcs12.Decode(f, testPKS12KeyPassword)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	handler, err := NewAdobePKCS7DetachedSignatureHandler(privateKey.(*rsa.PrivateKey), cert)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	_, appearance, err := appender.Sign(1, handler)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appearance.Signature.Reason = core.MakeString("TestAppenderSignPage4")
	appearance.Signature.Name = core.MakeString("Test Appender")

	err = appender.WriteToFile(tempFile("appender_sign_page_4.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	validateFile(t, tempFile("appender_sign_page_4.pdf"))
}
