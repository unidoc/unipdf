/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package render

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/unidoc/unipdf/v3/model"
	imagectx "github.com/unidoc/unipdf/v3/render/context/image"
)

// ImageDevice is used to render PDF pages to image targets.
type ImageDevice struct {
	renderer
}

// NewImageDevice returns a new image device.
func NewImageDevice() *ImageDevice {
	return &ImageDevice{}
}

// Render converts the specified PDF page into an image and returns the result.
func (d *ImageDevice) Render(page *model.PdfPage) (image.Image, error) {
	// Get page dimensions.
	mbox, err := page.GetMediaBox()
	if err != nil {
		return nil, err
	}

	// Render page.
	width, height := mbox.Llx+mbox.Width(), mbox.Lly+mbox.Height()

	ctx := imagectx.NewContext(int(width), int(height))
	if err := d.renderPage(ctx, page); err != nil {
		return nil, err
	}

	// Apply crop box, if one exists.
	img := ctx.Image()
	if box := page.CropBox; box != nil {
		// Calculate crop bounds and crop start position.
		cropBounds := image.Rect(0, 0, int(box.Width()), int(box.Height()))
		cropStart := image.Pt(int(box.Llx), int(height-box.Ury))

		// Crop image.
		cropImg := image.NewRGBA(cropBounds)
		draw.Draw(cropImg, cropBounds, img, cropStart, draw.Src)
		img = cropImg
	}

	return img, nil
}

// RenderToPath converts the specified PDF page into an image and saves the
// result at the specified location.
func (d *ImageDevice) RenderToPath(page *model.PdfPage, outputPath string) error {
	image, err := d.Render(page)
	if err != nil {
		return err
	}

	extension := strings.ToLower(filepath.Ext(outputPath))
	if extension == "" {
		return errors.New("could not recognize output file type")
	}

	switch extension {
	case ".png":
		return savePNG(outputPath, image)
	case ".jpg", ".jpeg":
		return saveJPG(outputPath, image, 100)
	}

	return fmt.Errorf("unrecognized output file type: %s", extension)
}

func savePNG(path string, image image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, image)
}

func saveJPG(path string, image image.Image, quality int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, image, &jpeg.Options{Quality: quality})
}
