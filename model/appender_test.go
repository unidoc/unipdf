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

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/pkcs12"

	"github.com/unidoc/unipdf/v3/annotator"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/model/sighandler"
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

func tempFile(name string) string {
	return filepath.Join(os.TempDir(), name)
}

// Appender with no data added should output an equivalent file.
func TestAppenderNoop(t *testing.T) {
	f, err := os.Open("./testdata/minimal.pdf")
	require.NoError(t, err)
	defer f.Close()

	reader, err := model.NewPdfReader(f)
	require.NoError(t, err)

	appender, err := model.NewPdfAppender(reader)
	require.NoError(t, err)

	model.SetPdfProducer("UniPDF")
	model.SetPdfCreator("UniDoc UniPDF")
	err = appender.WriteToFile(tempFile("appender_noop.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// The first part of the appended file should be same as in the original.
	minimal, err := ioutil.ReadFile("./testdata/minimal.pdf")
	require.NoError(t, err)
	out, err := ioutil.ReadFile(tempFile("appender_noop.pdf"))
	require.NoError(t, err)
	require.Equal(t, minimal, out[0:len(minimal)])

	origReader := reader
	origObjNums := origReader.GetObjectNums()
	require.Equal(t, []int{1, 2, 3, 4}, origObjNums)

	// Check cross references table of original.
	page, err := origReader.GetPage(1)
	require.NoError(t, err)
	origParser := page.GetPageAsIndirectObject().GetParser()
	origXref := origParser.GetXrefTable()
	expected := map[int]core.XrefObject{
		1: core.XrefObject{ObjectNumber: 1, XType: 0, Offset: 18},
		2: core.XrefObject{ObjectNumber: 2, XType: 0, Offset: 77},
		3: core.XrefObject{ObjectNumber: 3, XType: 0, Offset: 178},
		4: core.XrefObject{ObjectNumber: 4, XType: 0, Offset: 457},
	}
	require.Equal(t, expected, origXref.ObjectMap)

	// Check the output.
	{
		f, err := os.Open(tempFile("appender_noop.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		objNums := reader.GetObjectNums()

		// Number of objects should be equal.
		// The appended version should only add 2 objects (new Info and Catalog).
		require.Equal(t, len(origObjNums)+2, len(objNums))
		require.Equal(t, []int{1, 2, 3, 4, 5, 6}, objNums)

		// Check cross references table of appended version.
		page, err := reader.GetPage(1)
		require.NoError(t, err)
		parser := page.GetPageAsIndirectObject().GetParser()
		xrefs := parser.GetXrefTable()
		expected := map[int]core.XrefObject{
			1: core.XrefObject{ObjectNumber: 1, XType: 0, Offset: 18},
			2: core.XrefObject{ObjectNumber: 2, XType: 0, Offset: 77},
			3: core.XrefObject{ObjectNumber: 3, XType: 0, Offset: 178},
			4: core.XrefObject{ObjectNumber: 4, XType: 0, Offset: 457},
			5: core.XrefObject{ObjectNumber: 5, XType: 0, Offset: 740},
			6: core.XrefObject{ObjectNumber: 6, XType: 0, Offset: 802},
		}
		require.Equal(t, expected, xrefs.ObjectMap)
	}
}

func TestAppenderRemovePage(t *testing.T) {
	f, err := os.Open(testPdf3pages)
	require.NoError(t, err)
	defer f.Close()

	reader, err := model.NewPdfReader(f)
	require.NoError(t, err)

	appender, err := model.NewPdfAppender(reader)
	require.NoError(t, err)

	appender.RemovePage(1)
	appender.RemovePage(2)

	err = appender.WriteToFile(tempFile("appender_remove_page_1.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	origReader := reader
	origObjNums := origReader.GetObjectNums()

	// Check the output.
	{
		f, err := os.Open(tempFile("appender_remove_page_1.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		objNums := reader.GetObjectNums()

		// Number of objects should be equal.
		// The appended version should only add 2 objects (new Info and Catalog).
		require.Equal(t, len(origObjNums)+2, len(objNums))

		obj2, err := reader.GetIndirectObjectByNumber(2)
		require.NoError(t, err)

		pagesDict, ok := core.GetDict(obj2)
		require.True(t, ok)

		kidsArr, ok := core.GetArray(pagesDict.Get("Kids"))
		require.True(t, ok)
		require.Len(t, kidsArr.Elements(), 1)
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

	// if the page is already in the same file, expect that the Pages object is only updated
	appender.ReplacePage(1, pdf1.PageList[1])
	appender.ReplacePage(3, pdf2.PageList[0])

	err = appender.WriteToFile(tempFile("appender_replace_page_1.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	origReader := pdf1
	origObjNums := origReader.GetObjectNums()

	// Check the output.
	{
		f, err := os.Open(tempFile("appender_replace_page_1.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		objNums := reader.GetObjectNums()

		// Number of objects should be equal.
		// The appended version adds 8 new objects: Related to the new page and new Info, Catalog, Pages.
		// As well as updating some previous objects.
		// TODO: Check specifically the xrefs table regarding updates, new objects etc.
		require.Equal(t, len(origObjNums)+9, len(objNums))

		obj2, err := reader.GetIndirectObjectByNumber(2)
		require.NoError(t, err)

		pagesDict, ok := core.GetDict(obj2)
		require.True(t, ok)

		kidsArr, ok := core.GetArray(pagesDict.Get("Kids"))
		require.True(t, ok)
		require.Len(t, kidsArr.Elements(), 3)

		// The first page is a copy of the second one.  And the third one is from another file.
		// All new objects.
		require.Equal(t, "[IObject:39, IObject:21, IObject:42]", kidsArr.String())
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

	page.AddAnnotation(annotation.PdfAnnotation)

	//appender.ReplacePage(1, page)
	appender.UpdatePage(page)

	err = appender.WriteToFile(tempFile("appender_add_annotation_1.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Check the output.
	{
		f, err := os.Open(tempFile("appender_add_annotation_1.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)
		require.NotNil(t, reader)
		require.Nil(t, reader.AcroForm)

		page, err := reader.GetPage(1)
		require.NoError(t, err)
		require.NotNil(t, page.Annots)

		annots, err := page.GetAnnotations()
		require.NoError(t, err)
		require.NotNil(t, annots)
		require.Len(t, annots, 1)
	}
}

// Append annotation to page which already has annotations.
func TestAppenderAddAnnotation2(t *testing.T) {
	f1, err := os.Open("testdata/OoPdfFormExample.pdf")
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

	page.AddAnnotation(annotation.PdfAnnotation)

	appender.UpdatePage(page)

	err = appender.WriteToFile(tempFile("appender_add_annotation_2.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// Check the output.
	{
		f, err := os.Open(tempFile("appender_add_annotation_2.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)
		require.NotNil(t, reader)
		require.NotNil(t, reader.AcroForm)

		require.Len(t, reader.AcroForm.AllFields(), 17)
	}
}

// Implements example H.7 Updating example (PDF32000_2008).
// The original file is the minimal pdf file (H.2). The updates are in 4 stages with the file saved after
// each stage:
// a) Four text annotations are added.
// b) The text of one of the annotations is altered.
// c) Two of the text annotations are deleted.
// d) Three text annotations are added.
func TestAppenderUpdatingExample(t *testing.T) {
	stage1 := func() {
		t.Logf("------- Stage 1")
		f, err := os.Open("./testdata/minimal.pdf")
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		appender, err := model.NewPdfAppender(reader)
		require.NoError(t, err)

		// Stage 1: Add four text annotations.
		annot1 := model.NewPdfAnnotationText()
		annot1.Rect = core.MakeArrayFromIntegers([]int{44, 616, 162, 735})
		annot1.Contents = core.MakeString("Text #1")
		annot1.Open = core.MakeBool(true)

		annot2 := model.NewPdfAnnotationText()
		annot2.Rect = core.MakeArrayFromIntegers([]int{224, 668, 457, 735})
		annot2.Contents = core.MakeString("Text #2")
		annot2.Open = core.MakeBool(false)

		annot3 := model.NewPdfAnnotationText()
		annot3.Rect = core.MakeArrayFromIntegers([]int{293, 393, 328, 622})
		annot3.Contents = core.MakeString("Text #3")
		annot3.Open = core.MakeBool(true)

		annot4 := model.NewPdfAnnotationText()
		annot4.Rect = core.MakeArrayFromIntegers([]int{34, 398, 225, 575})
		annot4.Contents = core.MakeString("Text #4")
		annot4.Open = core.MakeBool(false)

		page, err := reader.GetPage(1)
		require.NoError(t, err)

		page.AddAnnotation(annot1.PdfAnnotation)
		page.AddAnnotation(annot2.PdfAnnotation)
		page.AddAnnotation(annot3.PdfAnnotation)
		page.AddAnnotation(annot4.PdfAnnotation)

		appender.UpdatePage(page)

		err = appender.WriteToFile(tempFile("appender_h7_stage1.pdf"))
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}
	stage1()

	// Check the output of stage 1.
	func() {
		f, err := os.Open("./testdata/minimal.pdf")
		require.NoError(t, err)
		defer f.Close()
		origReader, err := model.NewPdfReader(f)
		origObjNums := origReader.GetObjectNums()

		f, err = os.Open(tempFile("appender_h7_stage1.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		objNums := reader.GetObjectNums()

		// The new revision should contain:
		// - Updated Page
		// - 4 New annotation objects
		// - New Info object (refrenced from the trailer).
		// - New Catalog
		require.Equal(t, len(origObjNums)+6, len(objNums))

		// Check that the Pages object number is unchanged.
		obj2, err := reader.GetIndirectObjectByNumber(2)
		require.NoError(t, err)

		pagesDict, ok := core.GetDict(obj2)
		require.True(t, ok)

		kidsArr, ok := core.GetArray(pagesDict.Get("Kids"))
		require.True(t, ok)
		require.Len(t, kidsArr.Elements(), 1)
		require.Equal(t, "[IObject:3]", kidsArr.String())
	}()

	stage2 := func() {
		t.Logf("------- Stage 2")
		f, err := os.Open(tempFile("appender_h7_stage1.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		appender, err := model.NewPdfAppender(reader)
		require.NoError(t, err)

		page, err := reader.GetPage(1)
		require.NoError(t, err)

		annots, err := page.GetAnnotations()
		require.NoError(t, err)
		require.Len(t, annots, 4)

		annot3 := annots[2]
		annot3.Contents = core.MakeString("Modified Text #3")

		appender.UpdateObject(annot3.GetContext().ToPdfObject())

		err = appender.WriteToFile(tempFile("appender_h7_stage2.pdf"))
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}
	stage2()
	// Check the output of stage 2.
	func() {
		f, err := os.Open(tempFile("appender_h7_stage1.pdf"))
		require.NoError(t, err)
		defer f.Close()
		origReader, err := model.NewPdfReader(f)
		origObjNums := origReader.GetObjectNums()

		f, err = os.Open(tempFile("appender_h7_stage2.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		objNums := reader.GetObjectNums()

		// The new revision should contain:
		// - Updated Annots object.
		// - New Info
		// - New Catalog
		// - Updated Pages
		// - No other changes
		require.Equal(t, len(origObjNums)+2, len(objNums))

		// Check that the Pages object number is unchanged.
		obj2, err := reader.GetIndirectObjectByNumber(2)
		require.NoError(t, err)

		pagesDict, ok := core.GetDict(obj2)
		require.True(t, ok)

		kidsArr, ok := core.GetArray(pagesDict.Get("Kids"))
		require.True(t, ok)
		require.Len(t, kidsArr.Elements(), 1)
		require.Equal(t, "[IObject:3]", kidsArr.String())

		annots, err := reader.PageList[0].GetAnnotations()
		require.NoError(t, err)
		str, ok := core.GetString(annots[2].Contents)
		require.True(t, ok)
		require.Equal(t, "Modified Text #3", str.String())
		textAnnot, ok := annots[2].GetContext().(*model.PdfAnnotationText)
		require.True(t, ok)
		open, ok := core.GetBool(textAnnot.Open)
		require.True(t, ok)
		require.Equal(t, "true", open.String())
	}()

	stage3 := func() {
		t.Logf("------- Stage 3")
		f, err := os.Open(tempFile("appender_h7_stage2.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		appender, err := model.NewPdfAppender(reader)
		require.NoError(t, err)

		page, err := reader.GetPage(1)
		require.NoError(t, err)

		annots, err := page.GetAnnotations()
		require.NoError(t, err)
		require.Len(t, annots, 4)

		// Remove two annotations.
		annots = annots[2:]
		page.SetAnnotations(annots)

		appender.UpdatePage(page)

		err = appender.WriteToFile(tempFile("appender_h7_stage3.pdf"))
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}
	stage3()
	// Check output of stage 3. Expected is just new Annots array + xrefs.
	func() {
		f, err := os.Open(tempFile("appender_h7_stage2.pdf"))
		require.NoError(t, err)
		defer f.Close()
		origReader, err := model.NewPdfReader(f)
		origObjNums := origReader.GetObjectNums()

		f, err = os.Open(tempFile("appender_h7_stage3.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		objNums := reader.GetObjectNums()

		// The new revision should contain:
		// - Updated Page object (including the Annots).
		// - New Info
		// - New Catalog
		// - No other changes
		require.Equal(t, len(origObjNums)+2, len(objNums))

		// Check that the Pages object number is unchanged.
		obj2, err := reader.GetIndirectObjectByNumber(2)
		require.NoError(t, err)

		pagesDict, ok := core.GetDict(obj2)
		require.True(t, ok)

		kidsArr, ok := core.GetArray(pagesDict.Get("Kids"))
		require.True(t, ok)
		require.Len(t, kidsArr.Elements(), 1)
		require.Equal(t, "[IObject:3]", kidsArr.String())

		annots, err := reader.PageList[0].GetAnnotations()
		require.NoError(t, err)
		require.Len(t, annots, 2)
	}()

	stage4 := func() {
		t.Logf("------- Stage 4")
		f, err := os.Open(tempFile("appender_h7_stage3.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		appender, err := model.NewPdfAppender(reader)
		require.NoError(t, err)

		// Stage 4: Add 3 new annotations.
		newAnnot1 := model.NewPdfAnnotationText()
		newAnnot1.Rect = core.MakeArrayFromIntegers([]int{58, 657, 172, 742})
		newAnnot1.Contents = core.MakeString("New Text #1")
		newAnnot1.Open = core.MakeBool(true)

		newAnnot2 := model.NewPdfAnnotationText()
		newAnnot2.Rect = core.MakeArrayFromIntegers([]int{389, 459, 570, 537})
		newAnnot2.Contents = core.MakeString("New Text #2")
		newAnnot2.Open = core.MakeBool(false)

		newAnnot3 := model.NewPdfAnnotationText()
		newAnnot3.Rect = core.MakeArrayFromIntegers([]int{44, 253, 473, 337})
		newAnnot3.Contents = core.MakeString("New Text #3")
		newAnnot3.Open = core.MakeBool(true)

		page, err := reader.GetPage(1)
		require.NoError(t, err)

		page.AddAnnotation(newAnnot1.PdfAnnotation)
		page.AddAnnotation(newAnnot2.PdfAnnotation)
		page.AddAnnotation(newAnnot3.PdfAnnotation)

		appender.UpdatePage(page)

		err = appender.WriteToFile(tempFile("appender_h7_stage4.pdf"))
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}
	}
	stage4()

	// Check output of stage 4.
	func() {
		f, err := os.Open(tempFile("appender_h7_stage3.pdf"))
		require.NoError(t, err)
		defer f.Close()
		origReader, err := model.NewPdfReader(f)
		origObjNums := origReader.GetObjectNums()

		f, err = os.Open(tempFile("appender_h7_stage4.pdf"))
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		objNums := reader.GetObjectNums()

		// The new revision should contain:
		// - Updated Page object (including the Annots).
		// - New Info
		// - New Catalog
		// - 3 New Annots
		// - No other changes
		require.Equal(t, len(origObjNums)+5, len(objNums))

		annots, err := reader.PageList[0].GetAnnotations()
		require.NoError(t, err)
		require.Len(t, annots, 5)
	}()
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
	require.NoError(t, err)
	defer f1.Close()

	pdf1, err := model.NewPdfReader(f1)
	require.NoError(t, err)

	f2, err := os.Open(testPdfAcroFormFile1)
	require.NoError(t, err)
	defer f2.Close()

	pdf2, err := model.NewPdfReader(f2)
	defer f2.Close()

	appender, err := model.NewPdfAppender(pdf1)
	defer f2.Close()

	appender.AddPages(pdf2.PageList...)

	// AcroForm from a different file.
	appender.ReplaceAcroForm(pdf2.AcroForm)

	err = appender.WriteToFile(tempFile("appender_add_page_2.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}

	// TODO: Check outputs.
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

// Multiple revisions of signing.
func TestAppenderSignMultiple(t *testing.T) {
	inputPath := "./testdata/minimal.pdf"

	for i := 0; i < 3; i++ {
		t.Logf("======================================")
		t.Logf("--> Signature revision %d", i+1)
		t.Logf("Input %s", inputPath)
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

		annotations, err := pdfReader.PageList[0].GetAnnotations()
		require.NoError(t, err)
		t.Logf("Annotations: %d", len(annotations))
		if len(annotations) != i {
			t.Fatalf("page annotations != %d (got %d)", i, len(annotations))
		}
		for j, annot := range annotations {
			annotPage := core.ResolveReference(annot.P)
			t.Logf("i=%d Annots page object equal? %v == %v?", j, pdfReader.PageList[0].GetContainingPdfObject(), annotPage)
			require.Equal(t, pdfReader.PageList[0].GetContainingPdfObject(), annotPage)
		}

		appender, err := model.NewPdfAppender(pdfReader)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		pfxData, _ := ioutil.ReadFile("./testdata/JohnSmith.pfx")
		privateKey, cert, err := pkcs12.Decode(pfxData, "password")
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
		t.Logf("Signing to %s", outPath)

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

// Multiple revisions of signing with appearances.
func TestAppenderSignMultipleAppearances(t *testing.T) {
	inputPath := testPdf3pages

	for i := 0; i < 3; i++ {
		t.Logf("======================================")
		t.Logf("--> Signature revision %d", i+1)
		t.Logf("Input %s", inputPath)
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

		numPages, err := pdfReader.GetNumPages()
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			return
		}

		t.Logf("Fields: %d", len(pdfReader.AcroForm.AllFields()))
		if len(pdfReader.AcroForm.AllFields()) != i*numPages {
			t.Fatalf("fields != %d (got %d)", i*numPages, len(pdfReader.AcroForm.AllFields()))
		}

		annotations, err := pdfReader.PageList[0].GetAnnotations()
		require.NoError(t, err)
		t.Logf("Annotations: %d", len(annotations))
		if len(annotations) != i {
			t.Fatalf("page annotations != %d (got %d)", i, len(annotations))
		}
		for j, annot := range annotations {
			annotPage := core.ResolveReference(annot.P)
			t.Logf("i=%d Annots page object equal? %v == %v?", j, pdfReader.PageList[0].GetContainingPdfObject(), annotPage)
			require.Equal(t, pdfReader.PageList[0].GetContainingPdfObject(), annotPage)
		}

		appender, err := model.NewPdfAppender(pdfReader)
		if err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		pfxData, _ := ioutil.ReadFile("./testdata/JohnSmith.pfx")
		privateKey, cert, err := pkcs12.Decode(pfxData, "password")
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

		for j := 0; j < numPages; j++ {
			pageNum := j + 1

			opts := annotator.NewSignatureFieldOpts()
			opts.BorderSize = 1
			opts.FontSize = 10
			opts.Rect = []float64{float64(200*i) + 50, 25, float64(200*i) + 150, 80}
			opts.FillColor = model.NewPdfColorDeviceRGB(255, 255, 0)
			opts.TextColor = model.NewPdfColorDeviceRGB(0, 0, 200)

			sigField, err := annotator.NewSignatureField(
				signature,
				[]*annotator.SignatureLine{
					annotator.NewSignatureLine("Name", fmt.Sprintf("John Smith %d", i+1)),
					annotator.NewSignatureLine("Date", fmt.Sprintf("2019.0%d.%d", i+1, i+1)),
					annotator.NewSignatureLine("Reason", fmt.Sprintf("Reason %d", i+1)),
					annotator.NewSignatureLine("Location", "New York"),
					annotator.NewSignatureLine("DN", fmt.Sprintf("authority%d:name%d", i+1, i+1)),
				},
				opts,
			)
			sigField.T = core.MakeString(fmt.Sprintf("Signature %d-%d", i+1, j+1))

			if err = appender.Sign(pageNum, sigField); err != nil {
				t.Errorf("Fail: %v\n", err)
				f.Close()
				return
			}
		}

		outPath := tempFile(fmt.Sprintf("appender_sign_multiple_appearances_%d.pdf", i+1))
		t.Logf("Signing to %s", outPath)

		if err = appender.WriteToFile(outPath); err != nil {
			t.Errorf("Fail: %v\n", err)
			f.Close()
			return
		}

		//validateFile(t, outPath)
		inputPath = outPath

		f.Close()
	}
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
