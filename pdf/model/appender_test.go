/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model_test

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/pkcs12"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/annotator"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/sighandler"
)

// This test file contains multiple tests to generate PDFs from existing Pdf files. The outputs are written
// into TMPDIR as files.  The files themselves need to be observed to check for correctness as we don't have
// a good way to automatically check if every detail is correct.

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

const testPKS12Key = "./testdata/certificate.p12"
const testPKS12KeyPassword = "password"
const testSampleSignatureFile = "./testdata/sample_signature"

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
	pdf1, err := model.NewPdfReader(f1)
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
	pdf2, err := model.NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
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
	pdf1, err := model.NewPdfReader(f1)
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
	pdf2, err := model.NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
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
	pdf1, err := model.NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
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
	pdf1, err := model.NewPdfReader(f1)
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
	pdf2, err := model.NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
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
	pdf1, err := model.NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	page := pdf1.PageList[0]
	annotation := model.NewPdfAnnotationSquare()
	rect := model.PdfRectangle{Ury: 250.0, Urx: 150.0, Lly: 50.0, Llx: 50.0}
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
	pdf1, err := model.NewPdfReader(f1)
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
	pdf2, err := model.NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
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

	pdf1, err := model.NewPdfReader(f1)
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

	pdf2, err := model.NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
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
	pdf1, err := model.NewPdfReader(f1)
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

	pdf2, err := model.NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
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
	t.Logf("Validating %s", fileName)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	reader, err := model.NewPdfReader(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	handler, _ := sighandler.NewAdobeX509RSASHA1(nil, nil)
	handler2, _ := sighandler.NewAdobePKCS7Detached(nil, nil)
	handlers := []model.SignatureHandler{handler, handler2}

	res, err := reader.ValidateSignatures(handlers)
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

	for i, item := range res {
		t.Logf("== Signature %d", i+1)
		t.Logf("%s", item.String())
	}
}

func parseByteRange(byteRange *core.PdfObjectArray) ([]int64, error) {
	if byteRange == nil {
		return nil, errors.New("byte range cannot be nil")
	}
	if byteRange.Len() != 4 {
		return nil, errors.New("invalid byte range length")
	}

	s1, err := core.GetNumberAsInt64(byteRange.Get(0))
	if err != nil {
		return nil, errors.New("invalid byte range value")
	}
	l1, err := core.GetNumberAsInt64(byteRange.Get(1))
	if err != nil {
		return nil, errors.New("invalid byte range value")
	}

	s2, err := core.GetNumberAsInt64(byteRange.Get(2))
	if err != nil {
		return nil, errors.New("invalid byte range value")
	}
	l2, err := core.GetNumberAsInt64(byteRange.Get(3))
	if err != nil {
		return nil, errors.New("invalid byte range value")
	}

	return []int64{s1, s1 + l1, s2, s2 + l2}, nil
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
	pdf1, err := model.NewPdfReader(f1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
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

	handler, err := sighandler.NewAdobePKCS7Detached(privateKey.(*rsa.PrivateKey), cert)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Create signature field and appearance.
	signature := model.NewPdfSignature(handler)
	signature.SetName("Test Appender")
	signature.SetReason("TestAppenderSignPage4")
	signature.SetDate(time.Now(), "")

	if err := signature.Initialize(); err != nil {
		return
	}

	sigField := model.NewPdfFieldSignature(signature)
	sigField.T = core.MakeString("Signature1")
	sigField.Rect = core.MakeArray(
		core.MakeInteger(0),
		core.MakeInteger(0),
		core.MakeInteger(0),
		core.MakeInteger(0),
	)

	if err = appender.Sign(1, sigField); err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	err = appender.WriteToFile(tempFile("appender_sign_page_4.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	validateFile(t, tempFile("appender_sign_page_4.pdf"))
}

func TestAppenderSignMultiple(t *testing.T) {
	inputPath := testPdfFile1

	for i := 0; i < 3; i++ {
		f, err := os.Open(inputPath)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}

		pdfReader, err := model.NewPdfReader(f)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		t.Logf("Fields: %d", len(pdfReader.AcroForm.AllFields()))

		if len(pdfReader.AcroForm.AllFields()) != i {
			t.Fatalf("fields != %d (got %d)", i, len(pdfReader.AcroForm.AllFields()))
		}

		t.Logf("Annotations: %d", len(pdfReader.PageList[0].Annotations))
		if len(pdfReader.PageList[0].Annotations) != i {
			t.Fatalf("page annotations != %d (got %d)", i, len(pdfReader.PageList[0].Annotations))
		}

		appender, err := model.NewPdfAppender(pdfReader)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		pfxData, _ := ioutil.ReadFile(testPKS12Key)
		privateKey, cert, err := pkcs12.Decode(pfxData, testPKS12KeyPassword)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		handler, err := sighandler.NewAdobePKCS7Detached(privateKey.(*rsa.PrivateKey), cert)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		// Create signature field and appearance.
		signature := model.NewPdfSignature(handler)
		signature.SetName(fmt.Sprintf("Test Appender - Round %d", i+1))
		signature.SetReason(fmt.Sprintf("Test Appender - Round %d", i+1))
		signature.SetDate(time.Now(), "")

		if err := signature.Initialize(); err != nil {
			return
		}

		sigField := model.NewPdfFieldSignature(signature)
		sigField.T = core.MakeString(fmt.Sprintf("Signature %d", i+1))
		sigField.Rect = core.MakeArray(
			core.MakeInteger(0),
			core.MakeInteger(0),
			core.MakeInteger(0),
			core.MakeInteger(0),
		)

		if err = appender.Sign(1, sigField); err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		outPath := tempFile(fmt.Sprintf("appender_sign_multiple_%d.pdf", i+1))

		err = appender.WriteToFile(outPath)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		validateFile(t, outPath)
		inputPath = outPath

		f.Close()
	}
}

func TestSignatureAppearance(t *testing.T) {
	f, err := os.Open(testPdf3pages)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	t.Logf("Fields: %d", len(pdfReader.AcroForm.AllFields()))

	appender, err := model.NewPdfAppender(pdfReader)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	pfxData, _ := ioutil.ReadFile(testPKS12Key)
	privateKey, cert, err := pkcs12.Decode(pfxData, testPKS12KeyPassword)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	handler, err := sighandler.NewAdobePKCS7Detached(privateKey.(*rsa.PrivateKey), cert)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Create signature.
	signature := model.NewPdfSignature(handler)
	signature.SetName("Test Signature Appearance Name")
	signature.SetReason("TestSignatureAppearance Reason")
	signature.SetDate(time.Now(), "")

	if err := signature.Initialize(); err != nil {
		return
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		// Annot1
		opts := annotator.NewSignatureFieldOpts()
		opts.FontSize = 10
		opts.Rect = []float64{10, 25, 75, 60}

		sigField, err := annotator.NewSignatureField(
			signature,
			[]*annotator.SignatureLine{
				annotator.NewSignatureLine("Name", "Jane Doe"),
				annotator.NewSignatureLine("Date", "2019.01.03"),
				annotator.NewSignatureLine("Reason", "Some reason"),
				annotator.NewSignatureLine("Location", "New York"),
				annotator.NewSignatureLine("DN", "authority1:name1"),
			},
			opts,
		)
		sigField.T = core.MakeString(fmt.Sprintf("Signature %d", pageNum))

		if err = appender.Sign(pageNum, sigField); err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}

		// Annot2
		opts = annotator.NewSignatureFieldOpts()
		opts.FontSize = 8
		opts.Rect = []float64{250, 25, 325, 70}
		opts.TextColor = model.NewPdfColorDeviceRGB(255, 0, 0)

		sigField, err = annotator.NewSignatureField(
			signature,
			[]*annotator.SignatureLine{
				annotator.NewSignatureLine("Name", "John Doe"),
				annotator.NewSignatureLine("Date", "2019.03.14"),
				annotator.NewSignatureLine("Reason", "No reason"),
				annotator.NewSignatureLine("Location", "London"),
				annotator.NewSignatureLine("DN", "authority2:name2"),
			},
			opts,
		)
		sigField.T = core.MakeString(fmt.Sprintf("Signature2 %d", pageNum))

		if err = appender.Sign(pageNum, sigField); err != nil {
			log.Fatalf("Fail: %v\n", err)
		}

		// Annot3
		opts = annotator.NewSignatureFieldOpts()
		opts.BorderSize = 1
		opts.FontSize = 10
		opts.Rect = []float64{475, 25, 590, 80}
		opts.FillColor = model.NewPdfColorDeviceRGB(255, 255, 0)
		opts.TextColor = model.NewPdfColorDeviceRGB(0, 0, 200)

		sigField, err = annotator.NewSignatureField(
			signature,
			[]*annotator.SignatureLine{
				annotator.NewSignatureLine("Name", "John Smith"),
				annotator.NewSignatureLine("Date", "2019.02.19"),
				annotator.NewSignatureLine("Reason", "Another reason"),
				annotator.NewSignatureLine("Location", "Paris"),
				annotator.NewSignatureLine("DN", "authority3:name3"),
			},
			opts,
		)
		sigField.T = core.MakeString(fmt.Sprintf("Signature3 %d", pageNum))

		if err = appender.Sign(pageNum, sigField); err != nil {
			log.Fatalf("Fail: %v\n", err)
		}
	}

	outPath := tempFile("appender_signature_appearance.pdf")
	if err = appender.WriteToFile(outPath); err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	validateFile(t, outPath)
}

func TestAppenderExternalSignature(t *testing.T) {
	validateFile(t, testPdfSignedPDFDocument)

	// Function to generate signed PDF using the specified signature handler.
	generateSignedFile := func(handler model.SignatureHandler) ([]byte, *model.PdfSignature, error) {
		file, err := os.Open(testPdfFile1)
		if err != nil {
			return nil, nil, err
		}
		defer file.Close()

		reader, err := model.NewPdfReader(file)
		if err != nil {
			return nil, nil, err
		}

		appender, err := model.NewPdfAppender(reader)
		if err != nil {
			return nil, nil, err
		}

		// Create signature.
		signature := model.NewPdfSignature(handler)
		signature.SetName("Test External Signature")
		signature.SetReason("TestAppenderExternalSignature")
		signature.SetDate(time.Date(2019, 3, 24, 7, 30, 24, 0, time.UTC), "")

		if err := signature.Initialize(); err != nil {
			return nil, nil, err
		}

		// Create signature field and appearance.
		opts := annotator.NewSignatureFieldOpts()
		opts.FontSize = 10
		opts.Rect = []float64{10, 25, 75, 60}

		field, err := annotator.NewSignatureField(
			signature,
			[]*annotator.SignatureLine{
				annotator.NewSignatureLine("Name", "John Doe"),
				annotator.NewSignatureLine("Date", "2019.15.03"),
				annotator.NewSignatureLine("Reason", "External signature test"),
			},
			opts,
		)
		field.T = core.MakeString("External signature")

		if err = appender.Sign(1, field); err != nil {
			return nil, nil, err
		}

		// Write PDF file to buffer.
		pdfBuf := bytes.NewBuffer(nil)
		if err = appender.Write(pdfBuf); err != nil {
			return nil, nil, err
		}

		return pdfBuf.Bytes(), signature, nil
	}

	// Function which signs PDF and returns the signature data.
	getExternalSignature := func() ([]byte, error) {
		certFile, err := ioutil.ReadFile(testPKS12Key)
		if err != nil {
			return nil, err
		}

		privateKey, cert, err := pkcs12.Decode(certFile, testPKS12KeyPassword)
		if err != nil {
			return nil, err
		}

		handler, err := sighandler.NewAdobePKCS7Detached(privateKey.(*rsa.PrivateKey), cert)
		if err != nil {
			return nil, err
		}

		// Get external signature data.
		_, signature, err := generateSignedFile(handler)
		if err != nil {
			return nil, err
		}

		return signature.Contents.Bytes(), nil
	}

	// Generate PDF file signed with empty signature.
	handler, err := sighandler.NewEmptyAdobePKCS7Detached(8192)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	pdfData, signature, err := generateSignedFile(handler)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Parse signature byte range.
	byteRange, err := parseByteRange(signature.ByteRange)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// This would be the time to send the PDF buffer to a signing device or
	// signing web service and get back the signature. We will simulate this by
	// signing the PDF using UniDoc and returning the signature data.
	signatureData, err := getExternalSignature()
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Apply external signature to the PDF data buffer.
	sigBytes := make([]byte, 8192)
	copy(sigBytes, signatureData)

	sig := core.MakeHexString(string(sigBytes)).WriteString()
	copy(pdfData[byteRange[1]:byteRange[2]], []byte(sig))

	// Write output file.
	outputPath := tempFile("appender_sign_external_signature.pdf")
	if err := ioutil.WriteFile(outputPath, pdfData, os.ModePerm); err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Validate output file.
	validateFile(t, outputPath)
}

// Each Appender can only be written out once, further invokations of Write should result in an error.
func TestAppenderAttemptMultiWrite(t *testing.T) {
	f1, err := os.Open(testPdfLoremIpsumFile)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
	defer f1.Close()
	pdf1, err := model.NewPdfReader(f1)
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
	pdf2, err := model.NewPdfReader(f2)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender, err := model.NewPdfAppender(pdf1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	appender.AddPages(pdf1.PageList...)
	appender.AddPages(pdf2.PageList...)
	appender.AddPages(pdf2.PageList...)

	// Write twice to buffer and compare results.
	var buf1, buf2 bytes.Buffer
	err = appender.Write(&buf1)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	err = appender.Write(&buf2)
	if err == nil {
		t.Fatalf("Second invokation of appender.Write should yield an error")
	}
}
