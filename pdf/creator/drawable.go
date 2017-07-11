/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// All widgets that can be used to draw with the creator need to implement the Drawable interface.
type Drawable interface {
	// Draw onto blocks representing Page contents. As the content can wrap over many pages, multiple
	// templates are returned, one per Page.  The function also takes a draw context containing information
	// where to draw (if relative positioning) and the available height to draw on accounting for Margins etc.
	GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error)
}

// Drawing context.  Continuously used when drawing the page contents.  Keeps track of current X, Y position,
// available height as well as other page parameters such as margins and dimensions.
type DrawContext struct {
	// Current page number.
	Page int

	// Current position.  In a relative positioning mode, a drawable will be placed at these coordinates.
	X, Y float64

	// Context dimensions.  Available width and height.
	Width, Height float64

	// Page Margins.
	Margins margins

	// Absolute Page size, widths and height.
	PageWidth  float64
	PageHeight float64
}
