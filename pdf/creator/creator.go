/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	"io"
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

	context DrawContext

	pageMargins margins

	pageWidth, pageHeight float64

	// Keep track of number of chapters for indexing.
	chapters int

	genFrontPageFunc      func(pageNum int, totPages int)
	genTableOfContentFunc func(toc *TableOfContents) (*Chapter, error)
	drawHeaderFunc        func(pageNum int, totPages int)
	drawFooterFunc        func(pageNum int, totPages int)

	finalized bool

	toc *TableOfContents
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

	c.toc = newTableOfContents()

	return c
}

// Set the page margins: left, right, top, bottom.
// The default page margins are 10% of document width.
func (c *Creator) SetPageMargins(left, right, top, bottom float64) {
	c.pageMargins.left = left
	c.pageMargins.right = right
	c.pageMargins.top = top
	c.pageMargins.bottom = bottom
}

// Returns the current Page width.
func (c *Creator) Width() float64 {
	return c.pageWidth
}

// Returns the current Page height.
func (c *Creator) Height() float64 {
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

// Set a new Page size.  Pages that are added after this will be created with this Page size.
// Does not affect pages already created.
//
// Common page sizes are defined as constants.
// Examples:
// 1. c.SetPageSize(creator.PageSizeA4)
// 2. c.SetPageSize(creator.PageSizeA3)
// 3. c.SetPageSize(creator.PageSizeLegal)
// 4. c.SetPageSize(creator.PageSizeLetter)
//
// For custom sizes: Use the PPMM (points per mm) and PPI (points per inch) when defining those based on
// physical page sizes:
//
// Examples:
// 1. 10x15 sq. mm: SetPageSize(PageSize{10*creator.PPMM, 15*creator.PPMM}) where PPMM is points per mm.
// 2. 3x2 sq. inches: SetPageSize(PageSize{3*creator.PPI, 2*creator.PPI}) where PPI is points per inch.
//
func (c *Creator) SetPageSize(size PageSize) {
	c.pagesize = size

	c.pageWidth = size[0]
	c.pageHeight = size[1]

	// Update default margins to 10% of width.
	m := 0.1 * c.pageWidth
	c.pageMargins.left = m
	c.pageMargins.right = m
	c.pageMargins.top = m
	c.pageMargins.bottom = m
}

// Set a function to draw a header on created output pages.
func (c *Creator) DrawHeader(drawHeaderFunc func(int, int)) {
	c.drawHeaderFunc = drawHeaderFunc
}

// Set a function to draw a footer on created output pages.
func (c *Creator) DrawFooter(drawFooterFunc func(int, int)) {
	c.drawFooterFunc = drawFooterFunc
}

// Set a function to generate a front Page.
func (c *Creator) CreateFrontPage(genFrontPageFunc func(pageNum int, numPages int)) {
	c.genFrontPageFunc = genFrontPageFunc
}

// Seta function to generate table of contents.
func (c *Creator) CreateTableOfContents(genTOCFunc func(toc *TableOfContents) (*Chapter, error)) {
	c.genTableOfContentFunc = genTOCFunc
}

// Create a new Page with current parameters.
func (c *Creator) newPage() *model.PdfPage {
	page := model.NewPdfPage()

	width := c.pagesize[0]
	height := c.pagesize[1]

	bbox := model.PdfRectangle{0, 0, width, height}
	page.MediaBox = &bbox

	c.pageWidth = width
	c.pageHeight = height

	c.initContext()

	return page
}

// Initialize the drawing context, moving to upper left corner.
func (c *Creator) initContext() {
	// Update context, move to upper left corner.
	c.context.X = c.pageMargins.left
	c.context.Y = c.pageMargins.top
	c.context.Width = c.pageWidth - c.pageMargins.right - c.pageMargins.left
	c.context.Height = c.pageHeight - c.pageMargins.bottom - c.pageMargins.top
	c.context.PageHeight = c.pageHeight
	c.context.PageWidth = c.pageWidth
	c.context.Margins = c.pageMargins
}

// Adds a new Page to the creator and sets as the active Page.
func (c *Creator) NewPage() {
	page := c.newPage()
	c.pages = append(c.pages, page)
	c.context.Page++
}

func (c *Creator) AddPage(page *model.PdfPage) error {
	mbox, err := page.GetMediaBox()
	if err != nil {
		common.Log.Debug("Failed to get page mediabox: %v", err)
		return err
	}

	c.context.X = mbox.Llx + c.pageMargins.left
	c.context.Y = c.pageMargins.top
	c.context.PageHeight = mbox.Ury - mbox.Lly
	c.context.PageWidth = mbox.Urx - mbox.Llx

	c.pages = append(c.pages, page)
	c.context.Page++

	return nil
}

// Rotate the current active page by angle degrees.
// NOTE: angleDeg must be a multiple of 90.
func (c *Creator) RotateDeg(angleDeg int64) error {
	page := c.getActivePage()
	if page == nil {
		common.Log.Debug("Fail to rotate: no page currently active")
		return errors.New("No page active")
	}
	if angleDeg%90 != 0 {
		common.Log.Debug("Error: Page rotation angle not a multiple of 90")
		return errors.New("Range check error")
	}

	// Do the rotation.
	var rotation int64 = 0
	if page.Rotate != nil {
		rotation = *(page.Rotate)
	}
	rotation += angleDeg // Rotate by angleDeg degrees.
	page.Rotate = &rotation

	return nil
}

// Get current context.
func (c *Creator) Context() DrawContext {
	return c.context
}

// Call before writing out.  Takes care of adding headers and footers, as well as generating front Page and
// table of contents.
func (c *Creator) finalize() error {
	totPages := len(c.pages)

	// Estimate number of additional generated pages and update TOC.
	genpages := 0
	if c.genFrontPageFunc != nil {
		genpages++
	}
	if c.genTableOfContentFunc != nil {
		c.initContext()
		c.context.Page = genpages + 1
		ch, err := c.genTableOfContentFunc(c.toc)
		if err != nil {
			return err
		}

		// Make an estimate of the number of pages.
		blocks, _, err := ch.GeneratePageBlocks(c.context)
		if err != nil {
			common.Log.Debug("Failed to generate blocks: %v", err)
			return err
		}
		genpages += len(blocks)

		// Update the table of content Page numbers, accounting for front Page and TOC.
		for idx, _ := range c.toc.entries {
			c.toc.entries[idx].PageNumber += genpages
		}

		// Remove the TOC chapter entry.
		c.toc.entries = c.toc.entries[:len(c.toc.entries)-1]
	}

	hasFrontPage := false
	// Generate the front Page.
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
		c.initContext()
		ch, err := c.genTableOfContentFunc(c.toc)
		if err != nil {
			common.Log.Debug("Error generating TOC: %v", err)
			return err
		}
		ch.SetShowNumbering(false)
		ch.SetIncludeInTOC(false)

		blocks, _, _ := ch.GeneratePageBlocks(c.context)
		tocpages := []*model.PdfPage{}
		for _, block := range blocks {
			block.SetPos(0, 0)
			totPages++
			p := c.newPage()
			// Place at front.
			tocpages = append(tocpages, p)
			c.setActivePage(p)
			c.Draw(block)
		}

		if hasFrontPage {
			front := c.pages[0]
			rest := c.pages[1:]
			c.pages = append([]*model.PdfPage{front}, tocpages...)
			c.pages = append(c.pages, rest...)
		} else {
			c.pages = append(tocpages, c.pages...)
		}

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

	return nil
}

// Move absolute position to x, y.
func (c *Creator) MoveTo(x, y float64) {
	c.context.X = x
	c.context.Y = y
}

// Move draw context to absolute position x.
func (c *Creator) MoveX(x float64) {
	c.context.X = x
}

// Move draw context to absolute position y.
func (c *Creator) MoveY(y float64) {
	c.context.Y = y
}

// Move draw context right by relative position dx (negative goes left).
func (c *Creator) MoveRight(dx float64) {
	c.context.X += dx
}

// Move draw context down by relative position dy (negative goes up).
func (c *Creator) MoveDown(dy float64) {
	c.context.Y += dy
}

// Draw the drawable widget to the document.  This can span over 1 or more pages. Additional pages are added if
// the contents go over the current Page.
func (c *Creator) Draw(d Drawable) error {
	if c.getActivePage() == nil {
		// Add a new Page if none added already.
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
	c.context.Height = ctx.PageHeight - ctx.Y - ctx.Margins.bottom

	return nil
}

// Write output of creator to io.WriteSeeker interface.
func (c *Creator) Write(ws io.WriteSeeker) error {
	if !c.finalized {
		c.finalize()
	}

	pdfWriter := model.NewPdfWriter()

	for _, page := range c.pages {
		err := pdfWriter.AddPage(page)
		if err != nil {
			common.Log.Error("Failed to add Page: %s", err)
			return err
		}
	}

	err := pdfWriter.Write(ws)
	if err != nil {
		return err
	}

	return nil
}

// Write output of creator to file.
func (c *Creator) WriteToFile(outputPath string) error {
	fWrite, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer fWrite.Close()

	err = c.Write(fWrite)
	if err != nil {
		return err
	}

	return nil
}
