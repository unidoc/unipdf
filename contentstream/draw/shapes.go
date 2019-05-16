package draw

import (
	"math"

	pdfcontent "github.com/unidoc/unipdf/v3/contentstream"
	pdfcore "github.com/unidoc/unipdf/v3/core"
	pdf "github.com/unidoc/unipdf/v3/model"
)

// Circle represents a circle shape with fill and border properties that can be drawn to a PDF content stream.
type Circle struct {
	X             float64
	Y             float64
	Width         float64
	Height        float64
	FillEnabled   bool // Show fill?
	FillColor     *pdf.PdfColorDeviceRGB
	BorderEnabled bool // Show border?
	BorderWidth   float64
	BorderColor   *pdf.PdfColorDeviceRGB
	Opacity       float64 // Alpha value (0-1).
}

// Draw draws the circle. Can specify a graphics state (gsName) for setting opacity etc.  Otherwise leave empty ("").
// Returns the content stream as a byte array, the bounding box and an error on failure.
func (c Circle) Draw(gsName string) ([]byte, *pdf.PdfRectangle, error) {
	xRad := c.Width / 2
	yRad := c.Height / 2
	if c.BorderEnabled {
		xRad -= c.BorderWidth / 2
		yRad -= c.BorderWidth / 2
	}

	magic := 0.551784
	xMagic := xRad * magic
	yMagic := yRad * magic

	bpath := NewCubicBezierPath()
	bpath = bpath.AppendCurve(NewCubicBezierCurve(-xRad, 0, -xRad, yMagic, -xMagic, yRad, 0, yRad))
	bpath = bpath.AppendCurve(NewCubicBezierCurve(0, yRad, xMagic, yRad, xRad, yMagic, xRad, 0))
	bpath = bpath.AppendCurve(NewCubicBezierCurve(xRad, 0, xRad, -yMagic, xMagic, -yRad, 0, -yRad))
	bpath = bpath.AppendCurve(NewCubicBezierCurve(0, -yRad, -xMagic, -yRad, -xRad, -yMagic, -xRad, 0))
	bpath = bpath.Offset(xRad, yRad)
	if c.BorderEnabled {
		bpath = bpath.Offset(c.BorderWidth/2, c.BorderWidth/2)
	}
	if c.X != 0 || c.Y != 0 {
		bpath = bpath.Offset(c.X, c.Y)
	}

	creator := pdfcontent.NewContentCreator()

	creator.Add_q()

	if c.FillEnabled {
		creator.Add_rg(c.FillColor.R(), c.FillColor.G(), c.FillColor.B())
	}
	if c.BorderEnabled {
		creator.Add_RG(c.BorderColor.R(), c.BorderColor.G(), c.BorderColor.B())
		creator.Add_w(c.BorderWidth)
	}
	if len(gsName) > 1 {
		// If a graphics state is provided, use it. (Used for transparency settings here).
		creator.Add_gs(pdfcore.PdfObjectName(gsName))
	}

	DrawBezierPathWithCreator(bpath, creator)
	creator.Add_h() // Close the path.

	if c.FillEnabled && c.BorderEnabled {
		creator.Add_B() // fill and stroke.
	} else if c.FillEnabled {
		creator.Add_f() // Fill.
	} else if c.BorderEnabled {
		creator.Add_S() // Stroke.
	}
	creator.Add_Q()

	// Get bounding box.
	pathBbox := bpath.GetBoundingBox()
	if c.BorderEnabled {
		// Account for stroke width.
		pathBbox.Height += c.BorderWidth
		pathBbox.Width += c.BorderWidth
		pathBbox.X -= c.BorderWidth / 2
		pathBbox.Y -= c.BorderWidth / 2
	}

	// Bounding box - global coordinate system.
	bbox := &pdf.PdfRectangle{}
	bbox.Llx = pathBbox.X
	bbox.Lly = pathBbox.Y
	bbox.Urx = pathBbox.X + pathBbox.Width
	bbox.Ury = pathBbox.Y + pathBbox.Height

	return creator.Bytes(), bbox, nil
}

// Rectangle is a shape with a specified Width and Height and a lower left corner at (X,Y) that can be
// drawn to a PDF content stream.  The rectangle can optionally have a border and a filling color.
// The Width/Height includes the border (if any specified), i.e. is positioned inside.
type Rectangle struct {
	X             float64
	Y             float64
	Width         float64
	Height        float64
	FillEnabled   bool // Show fill?
	FillColor     *pdf.PdfColorDeviceRGB
	BorderEnabled bool // Show border?
	BorderWidth   float64
	BorderColor   *pdf.PdfColorDeviceRGB
	Opacity       float64 // Alpha value (0-1).
}

// Draw draws the rectangle. Can specify a graphics state (gsName) for setting opacity etc.
// Otherwise leave empty (""). Returns the content stream as a byte array, bounding box and an error on failure.
func (rect Rectangle) Draw(gsName string) ([]byte, *pdf.PdfRectangle, error) {
	path := NewPath()

	path = path.AppendPoint(NewPoint(0, 0))
	path = path.AppendPoint(NewPoint(0, rect.Height))
	path = path.AppendPoint(NewPoint(rect.Width, rect.Height))
	path = path.AppendPoint(NewPoint(rect.Width, 0))
	path = path.AppendPoint(NewPoint(0, 0))

	if rect.X != 0 || rect.Y != 0 {
		path = path.Offset(rect.X, rect.Y)
	}

	creator := pdfcontent.NewContentCreator()

	creator.Add_q()
	if rect.FillEnabled {
		creator.Add_rg(rect.FillColor.R(), rect.FillColor.G(), rect.FillColor.B())
	}
	if rect.BorderEnabled {
		creator.Add_RG(rect.BorderColor.R(), rect.BorderColor.G(), rect.BorderColor.B())
		creator.Add_w(rect.BorderWidth)
	}
	if len(gsName) > 1 {
		// If a graphics state is provided, use it. (Used for transparency settings here).
		creator.Add_gs(pdfcore.PdfObjectName(gsName))
	}
	DrawPathWithCreator(path, creator)
	creator.Add_h() // Close the path.

	if rect.FillEnabled && rect.BorderEnabled {
		creator.Add_B() // fill and stroke.
	} else if rect.FillEnabled {
		creator.Add_f() // Fill.
	} else if rect.BorderEnabled {
		creator.Add_S() // Stroke.
	}
	creator.Add_Q()

	// Get bounding box.
	pathBbox := path.GetBoundingBox()

	// Bounding box - global coordinate system.
	bbox := &pdf.PdfRectangle{}
	bbox.Llx = pathBbox.X
	bbox.Lly = pathBbox.Y
	bbox.Urx = pathBbox.X + pathBbox.Width
	bbox.Ury = pathBbox.Y + pathBbox.Height

	return creator.Bytes(), bbox, nil
}

// LineEndingStyle defines the line ending style for lines.
// The currently supported line ending styles are None, Arrow (ClosedArrow) and Butt.
type LineEndingStyle int

// Line ending styles.
const (
	LineEndingStyleNone  LineEndingStyle = 0
	LineEndingStyleArrow LineEndingStyle = 1
	LineEndingStyleButt  LineEndingStyle = 2
)

// LineStyle refers to how the line will be created.
type LineStyle int

// Line styles.
const (
	LineStyleSolid  LineStyle = 0
	LineStyleDashed LineStyle = 1
)

// Line defines a line shape between point 1 (X1,Y1) and point 2 (X2,Y2).  The line ending styles can be none (regular line),
// or arrows at either end.  The line also has a specified width, color and opacity.
type Line struct {
	X1               float64
	Y1               float64
	X2               float64
	Y2               float64
	LineColor        *pdf.PdfColorDeviceRGB
	Opacity          float64 // Alpha value (0-1).
	LineWidth        float64
	LineEndingStyle1 LineEndingStyle // Line ending style of point 1.
	LineEndingStyle2 LineEndingStyle // Line ending style of point 2.
	LineStyle        LineStyle
}

// Draw draws the line to PDF contentstream. Generates the content stream which can be used in page contents or
// appearance stream of annotation. Returns the stream content, XForm bounding box (local), bounding box and an error
// if one occurred.
func (line Line) Draw(gsName string) ([]byte, *pdf.PdfRectangle, error) {
	x1, x2 := line.X1, line.X2
	y1, y2 := line.Y1, line.Y2

	dy := y2 - y1
	dx := x2 - x1
	theta := math.Atan2(dy, dx)

	L := math.Sqrt(math.Pow(dx, 2.0) + math.Pow(dy, 2.0))
	w := line.LineWidth

	pi := math.Pi

	mul := 1.0
	if dx < 0 {
		mul *= -1.0
	}
	if dy < 0 {
		mul *= -1.0
	}

	// Vs.
	VsX := mul * (-w / 2 * math.Cos(theta+pi/2))
	VsY := mul * (-w/2*math.Sin(theta+pi/2) + w*math.Sin(theta+pi/2))

	// V1.
	V1X := VsX + w/2*math.Cos(theta+pi/2)
	V1Y := VsY + w/2*math.Sin(theta+pi/2)

	// P2.
	V2X := VsX + w/2*math.Cos(theta+pi/2) + L*math.Cos(theta)
	V2Y := VsY + w/2*math.Sin(theta+pi/2) + L*math.Sin(theta)

	// P3.
	V3X := VsX + w/2*math.Cos(theta+pi/2) + L*math.Cos(theta) + w*math.Cos(theta-pi/2)
	V3Y := VsY + w/2*math.Sin(theta+pi/2) + L*math.Sin(theta) + w*math.Sin(theta-pi/2)

	// P4.
	V4X := VsX + w/2*math.Cos(theta-pi/2)
	V4Y := VsY + w/2*math.Sin(theta-pi/2)

	path := NewPath()
	path = path.AppendPoint(NewPoint(V1X, V1Y))
	path = path.AppendPoint(NewPoint(V2X, V2Y))
	path = path.AppendPoint(NewPoint(V3X, V3Y))
	path = path.AppendPoint(NewPoint(V4X, V4Y))

	lineEnding1 := line.LineEndingStyle1
	lineEnding2 := line.LineEndingStyle2

	// TODO: Allow custom height/widths.
	arrowHeight := 3 * w
	arrowWidth := 3 * w
	arrowExtruding := (arrowWidth - w) / 2

	if lineEnding2 == LineEndingStyleArrow {
		// Convert P2, P3
		p2 := path.GetPointNumber(2)

		va1 := NewVectorPolar(arrowHeight, theta+pi)
		pa1 := p2.AddVector(va1)

		bVec := NewVectorPolar(arrowWidth/2, theta+pi/2)
		aVec := NewVectorPolar(arrowHeight, theta)

		va2 := NewVectorPolar(arrowExtruding, theta+pi/2)
		pa2 := pa1.AddVector(va2)

		va3 := aVec.Add(bVec.Flip())
		pa3 := pa2.AddVector(va3)

		va4 := bVec.Scale(2).Flip().Add(va3.Flip())
		pa4 := pa3.AddVector(va4)

		pa5 := pa1.AddVector(NewVectorPolar(w, theta-pi/2))

		newpath := NewPath()
		newpath = newpath.AppendPoint(path.GetPointNumber(1))
		newpath = newpath.AppendPoint(pa1)
		newpath = newpath.AppendPoint(pa2)
		newpath = newpath.AppendPoint(pa3)
		newpath = newpath.AppendPoint(pa4)
		newpath = newpath.AppendPoint(pa5)
		newpath = newpath.AppendPoint(path.GetPointNumber(4))

		path = newpath
	}
	if lineEnding1 == LineEndingStyleArrow {
		// Get the first and last points.
		p1 := path.GetPointNumber(1)
		pn := path.GetPointNumber(path.Length())

		// First three points on arrow.
		v1 := NewVectorPolar(w/2, theta+pi+pi/2)
		pa1 := p1.AddVector(v1)

		v2 := NewVectorPolar(arrowHeight, theta).Add(NewVectorPolar(arrowWidth/2, theta+pi/2))
		pa2 := pa1.AddVector(v2)

		v3 := NewVectorPolar(arrowExtruding, theta-pi/2)
		pa3 := pa2.AddVector(v3)

		// Last three points
		v5 := NewVectorPolar(arrowHeight, theta)
		pa5 := pn.AddVector(v5)

		v6 := NewVectorPolar(arrowExtruding, theta+pi+pi/2)
		pa6 := pa5.AddVector(v6)

		pa7 := pa1

		newpath := NewPath()
		newpath = newpath.AppendPoint(pa1)
		newpath = newpath.AppendPoint(pa2)
		newpath = newpath.AppendPoint(pa3)
		for _, p := range path.Points[1 : len(path.Points)-1] {
			newpath = newpath.AppendPoint(p)
		}
		newpath = newpath.AppendPoint(pa5)
		newpath = newpath.AppendPoint(pa6)
		newpath = newpath.AppendPoint(pa7)

		path = newpath
	}

	creator := pdfcontent.NewContentCreator()

	// Draw line with arrow
	creator.
		Add_q().
		Add_rg(line.LineColor.R(), line.LineColor.G(), line.LineColor.B())
	if len(gsName) > 1 {
		// If a graphics state is provided, use it. (Used for transparency settings here).
		creator.Add_gs(pdfcore.PdfObjectName(gsName))
	}

	path = path.Offset(line.X1, line.Y1)

	pathBbox := path.GetBoundingBox()

	DrawPathWithCreator(path, creator)

	if line.LineStyle == LineStyleDashed {
		creator.
			Add_d([]int64{1, 1}, 0).
			Add_S().
			Add_f().
			Add_Q()
	} else {
		creator.
			Add_f().
			Add_Q()
	}

	// Bounding box - global coordinate system.
	bbox := &pdf.PdfRectangle{}
	bbox.Llx = pathBbox.X
	bbox.Lly = pathBbox.Y
	bbox.Urx = pathBbox.X + pathBbox.Width
	bbox.Ury = pathBbox.Y + pathBbox.Height

	return creator.Bytes(), bbox, nil
}

// BasicLine defines a line between point 1 (X1,Y1) and point 2 (X2,Y2). The line has a specified width, color and opacity.
type BasicLine struct {
	X1        float64
	Y1        float64
	X2        float64
	Y2        float64
	LineColor *pdf.PdfColorDeviceRGB
	Opacity   float64 // Alpha value (0-1).
	LineWidth float64
	LineStyle LineStyle
}

// Draw draws the basic line to PDF. Generates the content stream which can be used in page contents or appearance
// stream of annotation. Returns the stream content, XForm bounding box (local), bounding box and an error if
// one occurred.
func (line BasicLine) Draw(gsName string) ([]byte, *pdf.PdfRectangle, error) {
	w := line.LineWidth

	path := NewPath()
	path = path.AppendPoint(NewPoint(line.X1, line.Y1))
	path = path.AppendPoint(NewPoint(line.X2, line.Y2))

	cc := pdfcontent.NewContentCreator()

	pathBbox := path.GetBoundingBox()

	DrawPathWithCreator(path, cc)

	if line.LineStyle == LineStyleDashed {
		cc.Add_d([]int64{1, 1}, 0)
	}
	cc.Add_RG(line.LineColor.R(), line.LineColor.G(), line.LineColor.B()).
		Add_w(w).
		Add_S().
		Add_Q()

	// Bounding box - global coordinate system.
	bbox := &pdf.PdfRectangle{}
	bbox.Llx = pathBbox.X
	bbox.Lly = pathBbox.Y
	bbox.Urx = pathBbox.X + pathBbox.Width
	bbox.Ury = pathBbox.Y + pathBbox.Height

	return cc.Bytes(), bbox, nil
}
