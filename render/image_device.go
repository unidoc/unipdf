package render

import (
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"strings"

	imagectx "github.com/unidoc/unipdf/render/context/image"
	"github.com/unidoc/unipdf/v3/model"
)

type ImageDevice struct {
	renderer
}

func NewImageDevice() *ImageDevice {
	return &ImageDevice{}
}

func (d *ImageDevice) render(page *model.PdfPage) (*imagectx.Context, error) {
	mbox, err := page.GetMediaBox()
	if err != nil {
		return nil, err
	}

	ctx := imagectx.NewContext(int(mbox.Width()), int(mbox.Height()))
	if err := d.renderPage(page, ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (d *ImageDevice) Render(page *model.PdfPage) (image.Image, error) {
	ctx, err := d.render(page)
	if err != nil {
		return nil, err
	}

	return ctx.Image(), nil
}

func (d *ImageDevice) RenderToPath(page *model.PdfPage, outputPath string) error {
	ctx, err := d.render(page)
	if err != nil {
		return err
	}

	extension := strings.ToLower(filepath.Ext(outputPath))
	if extension == "" {
		return errors.New("could not recognize output file type")
	}

	switch extension {
	case ".png":
		return ctx.SavePNG(outputPath)
	case ".jpg", ".jpeg":
		return ctx.SaveJPG(outputPath, 100)
	}

	return fmt.Errorf("unrecognized output file type: %s", extension)
}
