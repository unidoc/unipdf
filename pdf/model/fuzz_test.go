package model

import (
	"testing"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
}

// Test for an endless recursive loop in
// func (this *PdfReader) buildPageList(node *PdfIndirectObject, parent *PdfIndirectObject) error
func TestFuzzReaderBuildPageLoop(t *testing.T) {
	/*
		The problem is when there are Pages entries pointing forward and backward (illegal), causing endless
		recursive looping.

		Example problem data:
			3 0 obj
			<< /Type /Pages /MediaBox [0 0 595 842] /Count 2 /Kids [ 2 0 R 12 0 R ] >>
			endobj


			2 0 obj
			<< /Type /Pages
			   /Kids [3 0 R]
			   /Count 1
			   /MediaBox [0 0 300 144]
			>>
			endobj

			12 0 obj
			<<
				/Type /Page
				/Parent 3 0 R
				/Resources 15 0 R
				/Contents 13 0 R
				/MediaBox [0 0 595 842]
			>>
			endobj
	*/

	pageDict := core.MakeDict()
	pageDict.Set("Type", core.MakeName("Pages"))
	page := core.MakeIndirectObject(pageDict)

	pagesDict := core.MakeDict()
	pages := core.MakeIndirectObject(pagesDict)
	pagesDict.Set("Type", core.MakeName("Pages"))
	pagesDict.Set("Kids", core.MakeArray(page))

	pageDict.Set("Kids", core.MakeArray(pages))

	// Make a dummy reader to test
	dummyPdfReader := PdfReader{}
	dummyPdfReader.traversed = map[core.PdfObject]bool{}
	dummyPdfReader.modelManager = NewModelManager()

	traversedPageNodes := map[core.PdfObject]bool{}
	err := dummyPdfReader.buildPageList(pages, nil, traversedPageNodes)

	// Current behavior is to avoid the recursive endless loop and simply return nil.  Logs a debug message.

	if err != nil {
		t.Errorf("Fail: %v", err)
	}

}
