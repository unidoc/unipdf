/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package testutils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Errors.
var (
	ErrRenderNotSupported = errors.New("rendering PDF files is not supported on this system")
)

// CopyFile copies the `src` file to `dst`.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// ReadPNG reads and returns the specified PNG file.
func ReadPNG(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return png.Decode(f)
}

// HashFile generates an MD5 hash from the contents of the specified file.
func HashFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// CompareImages compares the specified images at pixel level.
func CompareImages(img1, img2 image.Image) (bool, error) {
	rect := img1.Bounds()
	diff := 0
	for x := 0; x < rect.Size().X; x++ {
		for y := 0; y < rect.Size().Y; y++ {
			r1, g1, b1, _ := img1.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 {
				diff++
			}
		}
	}

	diffFraction := float64(diff) / float64(rect.Dx()*rect.Dy())
	if diffFraction > 0.0001 {
		fmt.Printf("diff fraction: %v (%d)\n", diffFraction, diff)
		return false, nil
	}

	return true, nil
}

// ComparePNGFiles compares the specified PNG files.
func ComparePNGFiles(file1, file2 string) (bool, error) {
	// Fast path - compare hashes.
	h1, err := HashFile(file1)
	if err != nil {
		return false, err
	}
	h2, err := HashFile(file2)
	if err != nil {
		return false, err
	}
	if h1 == h2 {
		return true, nil
	}

	// Slow path - compare pixel by pixel.
	img1, err := ReadPNG(file1)
	if err != nil {
		return false, err
	}
	img2, err := ReadPNG(file2)
	if err != nil {
		return false, err
	}
	if img1.Bounds() != img2.Bounds() {
		return false, nil
	}

	return CompareImages(img1, img2)
}

// RenderPDFToPNGs uses ghostscript (gs) to render the specified PDF file into
// a set of PNG images (one per page). PNG images will be named xxx-N.png where
// N is the number of page, starting from 1.
func RenderPDFToPNGs(pdfPath string, dpi int, outpathTpl string) error {
	if dpi <= 0 {
		dpi = 100
	}
	if _, err := exec.LookPath("gs"); err != nil {
		return ErrRenderNotSupported
	}

	return exec.Command("gs", "-sDEVICE=pngalpha", "-o", outpathTpl, fmt.Sprintf("-r%d", dpi), pdfPath).Run()
}

// RunRenderTest renders the PDF file specified by `pdfPath` to the `outputDir`
// and compares the output PNG files to the golden files found at the
// `baselineRenderPath` location. If the specified PDF file is new (there are
// no golden files) and the `saveBaseline` parameter is set to true, the
// output render files are saved to the `baselineRenderPath`.
func RunRenderTest(t *testing.T, pdfPath, outputDir, baselineRenderPath string, saveBaseline bool) {
	tplName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	t.Run("render", func(t *testing.T) {
		imgPathPrefix := filepath.Join(outputDir, tplName)
		imgPathTpl := imgPathPrefix + "-%d.png"
		if err := RenderPDFToPNGs(pdfPath, 0, imgPathTpl); err != nil {
			t.Skip(err)
		}

		for i := 1; true; i++ {
			imgPath := fmt.Sprintf("%s-%d.png", imgPathPrefix, i)
			expImgPath := filepath.Join(baselineRenderPath, fmt.Sprintf("%s-%d_exp.png", tplName, i))
			if _, err := os.Stat(imgPath); err != nil {
				break
			}
			t.Logf("%s", expImgPath)
			if _, err := os.Stat(expImgPath); os.IsNotExist(err) {
				if saveBaseline {
					t.Logf("Copying %s -> %s", imgPath, expImgPath)
					CopyFile(imgPath, expImgPath)
					continue
				}
				break
			}

			t.Run(fmt.Sprintf("page%d", i), func(t *testing.T) {
				t.Logf("Comparing %s vs %s", imgPath, expImgPath)
				ok, err := ComparePNGFiles(imgPath, expImgPath)
				if os.IsNotExist(err) {
					t.Fatal("image file missing")
				} else if !ok {
					t.Fatal("wrong page rendered")
				}
			})
		}
	})
}
