package creator

import (
	"github.com/unidoc/unidoc/pdf/contentstream/draw"
	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	pdfcore "github.com/unidoc/unidoc/pdf/core"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

type FilledCurve struct {
	curves        []draw.CubicBezierCurve
	FillEnabled   bool // Show fill?
	fillColor     *pdf.PdfColorDeviceRGB
	BorderEnabled bool // Show border?
	BorderWidth   float64
	borderColor   *pdf.PdfColorDeviceRGB
	Opacity       float64 // Alpha value (0-1).
}

// NewFilledCurve returns a instance of filled curve
func NewFilledCurve() *FilledCurve {
	curve := FilledCurve{}
	curve.curves = []draw.CubicBezierCurve{}
	return &curve
}

// AppendCurve appends curve to filled curve
func (this *FilledCurve) AppendCurve(curve draw.CubicBezierCurve) *FilledCurve {
	this.curves = append(this.curves, curve)
	return this
}

func (this *FilledCurve) SetFillColor(color Color) {
	this.fillColor = pdf.NewPdfColorDeviceRGB(color.ToRGB())
}

func (this *FilledCurve) SetBorderColor(color Color) {
	this.borderColor = pdf.NewPdfColorDeviceRGB(color.ToRGB())
}

// Draw a circle. Can specify a graphics state (gsName) for setting opacity etc.  Otherwise leave empty ("").
// Returns the content stream as a byte array, the bounding box and an error on failure.
func (this *FilledCurve) draw(gsName string) ([]byte, *pdf.PdfRectangle, error) {
	bpath := draw.NewCubicBezierPath()
	for _, c := range this.curves {
		bpath = bpath.AppendCurve(c)
	}

	creator := pdfcontent.NewContentCreator()
	creator.Add_q()

	if this.FillEnabled {
		creator.Add_rg(this.fillColor.R(), this.fillColor.G(), this.fillColor.B())
	}
	if this.BorderEnabled {
		creator.Add_RG(this.borderColor.R(), this.borderColor.G(), this.borderColor.B())
		creator.Add_w(this.BorderWidth)
	}
	if len(gsName) > 1 {
		// If a graphics state is provided, use it. (Used for transparency settings here).
		creator.Add_gs(pdfcore.PdfObjectName(gsName))
	}

	draw.DrawBezierPathWithCreator(bpath, creator)
	creator.Add_h() // Close the path.

	if this.FillEnabled && this.BorderEnabled {
		creator.Add_B() // fill and stroke.
	} else if this.FillEnabled {
		creator.Add_f() // Fill.
	} else if this.BorderEnabled {
		creator.Add_S() // Stroke.
	}
	creator.Add_Q()

	// Get bounding box.
	pathBbox := bpath.GetBoundingBox()
	if this.BorderEnabled {
		// Account for stroke width.
		pathBbox.Height += this.BorderWidth
		pathBbox.Width += this.BorderWidth
		pathBbox.X -= this.BorderWidth / 2
		pathBbox.Y -= this.BorderWidth / 2
	}

	// Bounding box - global coordinate system.
	bbox := &pdf.PdfRectangle{}
	bbox.Llx = pathBbox.X
	bbox.Lly = pathBbox.Y
	bbox.Urx = pathBbox.X + pathBbox.Width
	bbox.Ury = pathBbox.Y + pathBbox.Height
	return creator.Bytes(), bbox, nil
}

// GeneratePageBlocks generates page blocks
func (this *FilledCurve) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	block := NewBlock(ctx.PageWidth, ctx.PageHeight)

	contents, _, err := this.draw("")
	err = block.addContentsByString(string(contents))
	if err != nil {
		return nil, ctx, err
	}
	return []*Block{block}, ctx, nil
}
