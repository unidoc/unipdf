package extractor

import (
	"os"
	"reflect"
	"testing"

	"github.com/unidoc/unidoc/pdf/model"
)

// TODO(gunnsth): Create test cases, specifically:
// 1. Basic test (one XObject image at specific position).
// 2. Test transform matrix handling (nested cms).
// 3. Inline image extraction.
// 4. Check caching.
// 5. Benchmark.

func loadPageFromPDFFile(filePath string, pageNum int) (*model.PdfPage, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, err
	}

	return pdfReader.GetPage(pageNum)
}

func TestImageExtractionBasic(t *testing.T) {
	type expectedImage struct {
		X      float64
		Y      float64
		Width  float64
		Height float64
		Angle  int
	}

	testcases := []struct {
		Name     string
		PageNum  int
		Path     string
		Expected []ImageMark
	}{
		{
			"basic xobject",
			1,
			"./testdata/basic_xobject.pdf",
			[]ImageMark{
				{
					Image:  nil,
					X:      0,
					Y:      294.865385,
					Width:  612,
					Height: 197.134615,
					Angle:  0,
				},
			},
		},
	}

	for _, tcase := range testcases {
		page, err := loadPageFromPDFFile(tcase.Path, tcase.PageNum)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		pageExtractor, err := New(page)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		pageImages, err := pageExtractor.ExtractPageImages()
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if len(pageImages.Images) != len(tcase.Expected) {
			t.Fatalf("%d != %d", len(pageImages.Images), len(tcase.Expected))
		}

		for i, img := range pageImages.Images {
			img.Image = nil // Discard image data.
			if !reflect.DeepEqual(img, tcase.Expected[i]) {
				t.Fatalf("i: %d - %#v != %#v", i, img, tcase.Expected[i])
			}
		}
	}
}

// Test position extraction with nest transform matrices.
func TestImageExtractionNestedCM(t *testing.T) {
}

func TestImageExtractionInline(t *testing.T) {
}

func TestImageExtractionCaching(t *testing.T) {
}

func BenchmarkImageExtraction(b *testing.B) {
}
