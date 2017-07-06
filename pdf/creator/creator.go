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

	context DrawContext

	pageMargins margins

	pageWidth, pageHeight float64

	// Keep track of number of chapters for indexing.
	chapters int

	genFrontPageFunc      func(pageNum int, totPages int)
	genTableOfContentFunc func(toc *TableOfContents) *Chapter
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

// Returns the current Page width.
func (c *Creator) Width() float64 {
	//return c.context.Width
	return c.pageWidth
}

// Returns the current Page height.
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

// Set a new Page size.  Pages that are added after this will be created with this Page size.
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

// Set a function to generate a front Page.
func (c *Creator) CreateFrontPage(genFrontPageFunc func(pageNum int, numPages int)) {
	c.genFrontPageFunc = genFrontPageFunc
}

// Seta function to generate table of contents.
func (c *Creator) CreateTableOfContents(genTOCFunc func(toc *TableOfContents) *Chapter) {
	c.genTableOfContentFunc = genTOCFunc
}

// Create a new Page with current parameters.
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
}

func (c *Creator) AddPage(page *model.PdfPage) {
	c.pages = append(c.pages, page)
}

// Call before writing out.  Takes care of adding headers and footers, as well as generating front Page and
// table of contents.
func (c *Creator) finalize() {
	totPages := len(c.pages)

	// Estimate number of additional generated pages and update TOC.
	genpages := 0
	if c.genFrontPageFunc != nil {
		genpages++
	}
	if c.genTableOfContentFunc != nil {
		c.initContext()
		c.context.Page = genpages + 1
		ch := c.genTableOfContentFunc(c.toc)
		if ch != nil {
			// Make an estimate of the number of pages.
			blocks, _, _ := ch.GeneratePageBlocks(c.context)
			genpages += len(blocks)

			// Update the table of content Page numbers, accounting for front Page and TOC.
			for idx, _ := range c.toc.entries {
				c.toc.entries[idx].PageNumber += genpages
			}

			// Remove the TOC chapter entry.
			c.toc.entries = c.toc.entries[:len(c.toc.entries)-1]
		}
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
		ch := c.genTableOfContentFunc(c.toc)
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
		c.context.Page = 1
	}

	blocks, ctx, err := d.GeneratePageBlocks(c.context)
	if err != nil {
		return err
	}

	for idx, blk := range blocks {
		if idx > 0 {
			c.NewPage()
			c.context.Page++
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

// Write output of creator to file.
func (c *Creator) WriteToFile(outputPath string) error {
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
