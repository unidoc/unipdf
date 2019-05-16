/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

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
		{
			"inline image",
			1,
			"./testdata/inline.pdf",
			[]ImageMark{
				{
					Image:  nil,
					X:      0,
					Y:      -0.000000358,
					Width:  12,
					Height: 12,
					Angle:  0,
				},
			},
		},
	}

	for _, tcase := range testcases {
		f, err := os.Open(tcase.Path)
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		page, err := reader.GetPage(tcase.PageNum)
		require.NoError(t, err)

		pageExtractor, err := New(page)
		require.NoError(t, err)

		pageImages, err := pageExtractor.ExtractPageImages(nil)
		require.NoError(t, err)

		assert.Equal(t, len(tcase.Expected), len(pageImages.Images))

		for i, img := range pageImages.Images {
			img.Image = nil // Discard image data.
			assert.Equalf(t, tcase.Expected[i], img, "i = %d", i)
		}
	}
}

// Test position extraction with nested transform matrices.
func TestImageExtractionNestedCM(t *testing.T) {
	testcases := []struct {
		Name      string
		PageNum   int
		Path      string
		PrependCS string
		AppendCS  string
		Expected  []ImageMark
	}{
		{
			"basic xobject - translate (100,50)",
			1,
			"./testdata/basic_xobject.pdf",
			"1 0 0 1 100.0 50.0 cm q",
			"Q",
			[]ImageMark{
				{
					Image:  nil,
					X:      0 + 100.0,
					Y:      294.865385 + 50.0,
					Width:  612,
					Height: 197.134615,
					Angle:  0,
				},
			},
		},
		{
			"basic xobject - scale (1.5,2)X",
			1,
			"./testdata/basic_xobject.pdf",
			"1.5 0 0 2.0 0 0 cm q",
			"Q",
			[]ImageMark{
				{
					Image:  nil,
					X:      0,
					Y:      294.865385 * 2.0,
					Width:  612 * 1.5,
					Height: 197.134615 * 2.0,
					Angle:  0,
				},
			},
		},
		{
			"basic xobject - translate (100,50) scale (1.5,2)X",
			1,
			"./testdata/basic_xobject.pdf",
			"1.5 0 0 2.0 0 0 cm q 1 0 0 1 100.0 50.0 cm q",
			"Q Q",
			[]ImageMark{
				{
					Image:  nil,
					X:      100.0 * 1.5,
					Y:      (294.865385 + 50.0) * 2.0,
					Width:  612 * 1.5,
					Height: 197.134615 * 2.0,
					Angle:  0,
				},
			},
		},
	}

	for _, tcase := range testcases {
		f, err := os.Open(tcase.Path)
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		page, err := reader.GetPage(tcase.PageNum)
		require.NoError(t, err)

		contentstr, err := page.GetAllContentStreams()
		require.NoError(t, err)

		// Modify the contentstream to alter the position by way of nested transform matrices.
		contentstr = tcase.PrependCS + " " + contentstr + " " + tcase.AppendCS
		err = page.SetContentStreams([]string{contentstr}, core.NewFlateEncoder())
		require.NoError(t, err)

		pageExtractor, err := New(page)
		require.NoError(t, err)

		pageImages, err := pageExtractor.ExtractPageImages(nil)
		require.NoError(t, err)

		assert.Equal(t, len(tcase.Expected), len(pageImages.Images))

		for i, img := range pageImages.Images {
			img.Image = nil // Discard image data.
			assert.Equalf(t, tcase.Expected[i], img, "i = %d", i)
		}
	}
}

// Test multiple copies of same image XObject with different scales.
func TestImageExtractionMulti(t *testing.T) {
	testcases := []struct {
		PageNum       int
		Path          string
		NumImages     int
		DimensionFunc func(i int) (dy float64, w float64, h float64)
		NumSamples    int
	}{
		{
			1,
			"./testdata/multi.pdf",
			12,
			func(i int) (dy float64, w float64, h float64) {
				w = 100 + 10*float64(i+1)
				ar := 35.432692 / 110.0
				h = w * ar

				dy = h

				return dy, w, h
			},
			416 * 134 * 3,
		},
	}

	for _, tcase := range testcases {
		f, err := os.Open(tcase.Path)
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		page, err := reader.GetPage(tcase.PageNum)
		require.NoError(t, err)

		pageExtractor, err := New(page)
		require.NoError(t, err)

		pageImages, err := pageExtractor.ExtractPageImages(nil)
		require.NoError(t, err)

		assert.Equal(t, tcase.NumImages, len(pageImages.Images))

		for i, img := range pageImages.Images {
			dy, w, h := tcase.DimensionFunc(i)

			assert.Equalf(t, tcase.NumSamples, len(img.Image.GetSamples()), "i = %d", i)

			// Comparison with tolerance.
			assert.Truef(t, math.Abs(w-img.Width) < 0.00001, "i = %d", i)
			assert.Truef(t, math.Abs(h-img.Height) < 0.00001, "i = %d", i)

			if i > 0 {
				measDY := pageImages.Images[i-1].Y - pageImages.Images[i].Y
				assert.Truef(t, math.Abs(dy-measDY) < 0.00001, "i = %d", i)
			}
		}
	}
}

func TestImageExtractionRealWorld(t *testing.T) {
	if len(corpusFolder) == 0 && !forceTest {
		t.Log("Corpus folder not set - skipping")
		return
	}

	testcases := []struct {
		Name     string
		PageNum  int
		Path     string
		Expected []ImageMark
	}{
		{
			"ICC color space",
			3,
			"icnp12-qinghua.pdf",
			[]ImageMark{
				{
					Image:  nil,
					Width:  2.877,
					Height: 22.344,
					X:      236.508,
					Y:      685.248,
					Angle:  0.0,
				},
				{
					Image:  nil,
					Width:  247.44,
					Height: 0.48,
					X:      313.788,
					Y:      715.248,
					Angle:  0.0,
				},
				{
					Image:  nil,
					Width:  247.44,
					Height: 0.48,
					X:      313.788,
					Y:      594.648,
					Angle:  0.0,
				},
			},
		},
		{
			"Indexed color space",
			1,
			"MondayAM.pdf",
			[]ImageMark{},
		},
	}

	for _, tcase := range testcases {
		inputPath := filepath.Join(corpusFolder, tcase.Path)

		f, err := os.Open(inputPath)
		require.NoError(t, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(t, err)

		page, err := reader.GetPage(tcase.PageNum)
		require.NoError(t, err)

		pageExtractor, err := New(page)
		require.NoError(t, err)

		pageImages, err := pageExtractor.ExtractPageImages(nil)
		require.NoError(t, err)

		if len(tcase.Expected) == 0 {
			// This is to test that images parse without error only.
			continue
		}
		assert.Equal(t, len(tcase.Expected), len(pageImages.Images))

		for i, img := range pageImages.Images {
			img.Image = nil // Discard image data.
			assert.Equalf(t, tcase.Expected[i], img, "i = %d", i)
		}
	}
}

func BenchmarkImageExtraction(b *testing.B) {
	cnt := 0
	for i := 0; i < b.N; i++ {
		f, err := os.Open("./testdata/basic_xobject.pdf")
		require.NoError(b, err)
		defer f.Close()

		reader, err := model.NewPdfReader(f)
		require.NoError(b, err)

		page, err := reader.GetPage(1)
		require.NoError(b, err)

		pageExtractor, err := New(page)
		require.NoError(b, err)

		pageImages, err := pageExtractor.ExtractPageImages(nil)
		require.NoError(b, err)

		cnt += len(pageImages.Images)
	}

	assert.Equal(b, b.N, cnt)
}
