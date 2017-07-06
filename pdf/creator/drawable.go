/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// All widgets that can be used to draw with the creator need to implement the Drawable interface.
type Drawable interface {
	// Sizing can either by occupy available space or a specified size.
	GetSizingMechanism() Sizing

	// Set absolute position of the widget on the Page/template to be drawn onto.
	SetPos(x, y float64)

	// Set the left, right, top, bottom Margins.
	SetMargins(float64, float64, float64, float64)

	// Get the left, right, top, bottom Margins.
	GetMargins() (float64, float64, float64, float64)

	// Returns the width/height of the drawable.
	Width() float64
	Height() float64

	// Draw onto blocks representing Page contents. As the content can wrap over many pages, multiple
	// templates are returned, one per Page.  The function also takes a draw context containing information
	// where to draw (if relative positioning) and the available height to draw on accounting for Margins etc.
	GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error)
}

// Some drawables can be scaled. Mostly objects that fit on a single Page.  E.g. image, block.
type Scalable interface {
	// Scale the drawable.  Does not actually influence the object contents but rather how it is represented
	// when drawn to the screen.  I.e. a coordinate transform.
	// Does change the Width and Height properties.
	Scale(float64, float64)
	ScaleToHeight(float64)
	ScaleToWidth(float64)
}

// Some drawables can be rotated. Mostly vector graphics that fit on a single Page.  E.g. image, block.
type Rotatable interface {
	// Set the rotation angle of the drawable in degrees.
	// The rotation does not change the dimensions of the Drawable and is only applied at the time of drawing.
	SetAngle(angleDeg float64)
}

// Drawing context.  Continuously used when drawing the Page contents.  Keeps track of current X, Y position,
// available height as well as other Page parameters such as Margins and dimensions.
type DrawContext struct {
	Page int

	// Current position.  In a relative positioning mode, a drawable will be placed at these coordinates.
	X, Y float64
	// Context dimensions.  Available width and height.
	Width, Height float64

	// Page Margins...
	Margins margins

	// Absolute Page size, widths and height.
	PageWidth  float64
	PageHeight float64
}
