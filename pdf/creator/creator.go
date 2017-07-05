/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"os"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model"
)

//
// The content creator is a wrapper around functionality for creating PDF reports and/or adding new
// content onto imported PDF pages.
//
type Creator struct {
	pages      []*model.PdfPage
	activePage *model.PdfPage

	pagesize PageSize

	context drawContext

	pageMargins margins

	pageWidth, pageHeight float64

	// Keep track of number of chapters for indexing.
	chapters int

	genFrontPageFunc      func(pageNum int, totPages int)
	genTableOfContentFunc func(pageNum int, totPages int)
	drawHeaderFunc        func(pageNum int, totPages int)
	drawFooterFunc        func(pageNum int, totPages int)

	finalized bool
}

type margins struct {
	left   float64
	right  float64
	top    float64
	bottom float64
}

// Create a new instance of the PDF creator.
func New() *Creator {
	c := &Creator{}
	c.pages = []*model.PdfPage{}
	c.SetPageSize(PageSizeLetter)

	m := 0.1 * c.pageWidth
	c.pageMargins.left = m
	c.pageMargins.right = m
	c.pageMargins.top = m
	c.pageMargins.bottom = m

	return c
}

// Returns the current page width.
func (c *Creator) Width() float64 {
	//return c.context.Width
	return c.pageWidth
}

// Returns the current page height.
func (c *Creator) Height() float64 {
	//return c.context.Height
	return c.pageHeight
}

func (c *Creator) setActivePage(p *model.PdfPage) {
	c.activePage = p
}

func (c *Creator) getActivePage() *model.PdfPage {
	if c.activePage == nil {
		if len(c.pages) == 0 {
			return nil
		}
		return c.pages[len(c.pages)-1]
	} else {
		return c.activePage
	}
}

// Set a new page size.  Pages that are added after this will be created with this page size.
// Does not affect pages already created.
func (c *Creator) SetPageSize(size PageSize) {
	c.pagesize = size

	dimensions := dimensionsMMtoPoints(pageSizesMM[size], float64(72.0))
	c.pageWidth = dimensions[0]
	c.pageHeight = dimensions[1]
}

// Set a function to draw a header on created output pages.
func (c *Creator) DrawHeader(drawHeaderFunc func(int, int)) {
	c.drawHeaderFunc = drawHeaderFunc
}

// Set a function to draw a footer on created output pages.
func (c *Creator) DrawFooter(drawFooterFunc func(int, int)) {
	c.drawFooterFunc = drawFooterFunc
}

// Set a function to generate a front page.
func (c *Creator) CreateFrontPage(genFrontPageFunc func(pageNum int, numPages int)) {
	c.genFrontPageFunc = genFrontPageFunc
}

// Seta function to generate table of contents.
func (c *Creator) CreateTableOfContents(genTOCFunc func(pageNum int, numPages int)) {
	c.genTableOfContentFunc = genTOCFunc
}

// Create a new page with current parameters.
func (c *Creator) newPage() *model.PdfPage {
	page := model.NewPdfPage()

	// Default: 72 points per inch.
	ppi := float64(72.0)
	dimensions := dimensionsMMtoPoints(pageSizesMM[c.pagesize], ppi)
	width := dimensions[0]
	height := dimensions[1]

	bbox := model.PdfRectangle{0, 0, width, height}
	page.MediaBox = &bbox

	c.pageWidth = width
	c.pageHeight = height

	// Update context, move to upper left corner.
	c.context.X = c.pageMargins.left
	c.context.Y = c.pageMargins.top
	c.context.Width = c.pageWidth - c.pageMargins.right - c.pageMargins.left
	c.context.Height = c.pageHeight - c.pageMargins.bottom - c.pageMargins.top
	c.context.pageHeight = c.pageHeight
	c.context.pageWidth = c.pageWidth
	c.context.margins = c.pageMargins

	return page
}

// Adds a new page to the creator and sets as the active page.
func (c *Creator) NewPage() {
	page := c.newPage()
	c.pages = append(c.pages, page)
}

func (c *Creator) AddPage(page *model.PdfPage) {
	c.pages = append(c.pages, page)
}

// Call before writing out.  Takes care of adding headers and footers, as well as generating front page and
// table of contents.
func (c *Creator) finalize() {
	totPages := len(c.pages)

	hasFrontPage := false
	if c.genFrontPageFunc != nil {
		totPages++
		p := c.newPage()
		// Place at front.
		c.pages = append([]*model.PdfPage{p}, c.pages...)
		c.setActivePage(p)

		c.genFrontPageFunc(1, totPages)
		hasFrontPage = true
	}

	if c.genTableOfContentFunc != nil {
		totPages++
		p := c.newPage()
		// Place at front.
		pageNum := 1
		if hasFrontPage {
			c.pages = append([]*model.PdfPage{c.pages[0], p}, c.pages[1:]...)
			pageNum = 2
		} else {
			c.pages = append([]*model.PdfPage{p}, c.pages...)
		}
		c.setActivePage(p)
		c.genTableOfContentFunc(pageNum, totPages)
	}

	for idx, page := range c.pages {
		c.setActivePage(page)
		if c.drawHeaderFunc != nil {
			c.drawHeaderFunc(idx+1, totPages)
		}
		if c.drawFooterFunc != nil {
			c.drawFooterFunc(idx+1, totPages)
		}
	}

	c.finalized = true
}

// Move absolute position to x, y.
func (c *Creator) MoveTo(x, y float64) {
	c.context.X = x
	c.context.Y = y
}

// Move absolute position x.
func (c *Creator) MoveX(x float64) {
	c.context.X = x
}

// Move absolute position x.
func (c *Creator) MoveY(y float64) {
	c.context.Y = y
}

// Move relative position x.
func (c *Creator) MoveXRel(dx float64) {
	c.context.X += dx
}

// Move relative position y.
func (c *Creator) MoveYRel(dy float64) {
	c.context.Y += dy
}

// Draw the drawable widget to the document.  This can span over 1 or more pages. Additional pages are added if
// the contents go over the current page.
func (c *Creator) Draw(d Drawable) error {
	if c.getActivePage() == nil {
		// Add a new page if none added already.
		c.NewPage()
	}

	blocks, ctx, err := d.GeneratePageBlocks(c.context)
	if err != nil {
		return err
	}

	for idx, blk := range blocks {
		if idx > 0 {
			c.NewPage()
		}

		p := c.getActivePage()
		err := blk.drawToPage(p)
		if err != nil {
			return err
		}
	}

	// Inner elements can affect X, Y position and available height.
	c.context.X = ctx.X
	c.context.Y = ctx.Y
	c.context.Height = ctx.pageHeight - ctx.Y - ctx.margins.bottom

	return nil
}

// Write output of creator to file.
func (c *Creator) WriteToFile(outputPath string) error {
	if !c.finalized {
		c.finalize()
	}
	pdfWriter := model.NewPdfWriter()

	for _, page := range c.pages {
		err := pdfWriter.AddPage(page)
		if err != nil {
			common.Log.Error("Failed to add page: %s", err)
			return err
		}
	}

	fWrite, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer fWrite.Close()

	err = pdfWriter.Write(fWrite)
	if err != nil {
		return err
	}

	return nil
}
