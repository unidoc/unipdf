/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"errors"
	goimage "image"
	"io"
	"os"
	"strconv"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
)

// Creator is a wrapper around functionality for creating PDF reports and/or adding new
// content onto imported PDF pages, etc.
type Creator struct {
	pages      []*model.PdfPage
	pageBlocks map[*model.PdfPage]*Block

	activePage *model.PdfPage

	pagesize PageSize

	context DrawContext

	pageMargins margins

	pageWidth, pageHeight float64

	// Keep track of number of chapters for indexing.
	chapters int

	// Hooks.
	genFrontPageFunc      func(args FrontpageFunctionArgs)
	genTableOfContentFunc func(toc *TOC) error
	drawHeaderFunc        func(header *Block, args HeaderFunctionArgs)
	drawFooterFunc        func(footer *Block, args FooterFunctionArgs)
	pdfWriterAccessFunc   func(writer *model.PdfWriter) error

	finalized bool

	// Controls whether a table of contents will be generated.
	AddTOC bool

	// The table of contents.
	toc *TOC

	// Controls whether outlines will be generated.
	AddOutlines bool

	// Outline.
	outline *model.Outline

	// External outlines.
	externalOutline *model.PdfOutlineTreeNode

	// Forms.
	acroForm *model.PdfAcroForm

	optimizer model.Optimizer

	// Default fonts used by all components instantiated through the creator.
	defaultFontRegular *model.PdfFont
	defaultFontBold    *model.PdfFont
}

// SetForms adds an Acroform to a PDF file.  Sets the specified form for writing.
func (c *Creator) SetForms(form *model.PdfAcroForm) error {
	c.acroForm = form
	return nil
}

// SetOutlineTree adds the specified outline tree to the PDF file generated
// by the creator. Adding an external outline tree disables the automatic
// generation of outlines done by the creator for the relevant components.
func (c *Creator) SetOutlineTree(outlineTree *model.PdfOutlineTreeNode) {
	c.externalOutline = outlineTree
}

// FrontpageFunctionArgs holds the input arguments to a front page drawing function.
// It is designed as a struct, so additional parameters can be added in the future with backwards
// compatibility.
type FrontpageFunctionArgs struct {
	PageNum    int
	TotalPages int
}

// HeaderFunctionArgs holds the input arguments to a header drawing function.
// It is designed as a struct, so additional parameters can be added in the future with backwards
// compatibility.
type HeaderFunctionArgs struct {
	PageNum    int
	TotalPages int
}

// FooterFunctionArgs holds the input arguments to a footer drawing function.
// It is designed as a struct, so additional parameters can be added in the future with backwards
// compatibility.
type FooterFunctionArgs struct {
	PageNum    int
	TotalPages int
}

// Margins.  Can be page margins, or margins around an element.
type margins struct {
	left   float64
	right  float64
	top    float64
	bottom float64
}

// New creates a new instance of the PDF Creator.
func New() *Creator {
	c := &Creator{}
	c.pages = []*model.PdfPage{}
	c.pageBlocks = map[*model.PdfPage]*Block{}
	c.SetPageSize(PageSizeLetter)

	m := 0.1 * c.pageWidth
	c.pageMargins.left = m
	c.pageMargins.right = m
	c.pageMargins.top = m
	c.pageMargins.bottom = m

	// Initialize default fonts.
	var err error

	c.defaultFontRegular, err = model.NewStandard14Font(model.HelveticaName)
	if err != nil {
		c.defaultFontRegular = model.DefaultFont()
	}

	c.defaultFontBold, err = model.NewStandard14Font(model.HelveticaBoldName)
	if err != nil {
		c.defaultFontRegular = model.DefaultFont()
	}

	// Initialize creator table of contents.
	c.toc = c.NewTOC("Table of Contents")

	// Initialize outline.
	c.AddOutlines = true
	c.outline = model.NewOutline()

	return c
}

// SetOptimizer sets the optimizer to optimize PDF before writing.
func (c *Creator) SetOptimizer(optimizer model.Optimizer) {
	c.optimizer = optimizer
}

// GetOptimizer returns current PDF optimizer.
func (c *Creator) GetOptimizer() model.Optimizer {
	return c.optimizer
}

// SetPageMargins sets the page margins: left, right, top, bottom.
// The default page margins are 10% of document width.
func (c *Creator) SetPageMargins(left, right, top, bottom float64) {
	c.pageMargins.left = left
	c.pageMargins.right = right
	c.pageMargins.top = top
	c.pageMargins.bottom = bottom
}

// Width returns the current page width.
func (c *Creator) Width() float64 {
	return c.pageWidth
}

// Height returns the current page height.
func (c *Creator) Height() float64 {
	return c.pageHeight
}

// TOC returns the table of contents component of the creator.
func (c *Creator) TOC() *TOC {
	return c.toc
}

// SetTOC sets the table of content component of the creator.
// This method should be used when building a custom table of contents.
func (c *Creator) SetTOC(toc *TOC) {
	if toc == nil {
		return
	}

	c.toc = toc
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
	}
	return c.activePage
}

// SetPageSize sets the Creator's page size.  Pages that are added after this will be created with
// this Page size.
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

// DrawHeader sets a function to draw a header on created output pages.
func (c *Creator) DrawHeader(drawHeaderFunc func(header *Block, args HeaderFunctionArgs)) {
	c.drawHeaderFunc = drawHeaderFunc
}

// DrawFooter sets a function to draw a footer on created output pages.
func (c *Creator) DrawFooter(drawFooterFunc func(footer *Block, args FooterFunctionArgs)) {
	c.drawFooterFunc = drawFooterFunc
}

// CreateFrontPage sets a function to generate a front Page.
func (c *Creator) CreateFrontPage(genFrontPageFunc func(args FrontpageFunctionArgs)) {
	c.genFrontPageFunc = genFrontPageFunc
}

// CreateTableOfContents sets a function to generate table of contents.
func (c *Creator) CreateTableOfContents(genTOCFunc func(toc *TOC) error) {
	c.genTableOfContentFunc = genTOCFunc
}

// Create a new Page with current parameters.
func (c *Creator) newPage() *model.PdfPage {
	page := model.NewPdfPage()

	width := c.pagesize[0]
	height := c.pagesize[1]

	bbox := model.PdfRectangle{Llx: 0, Lly: 0, Urx: width, Ury: height}
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

// NewPage adds a new Page to the Creator and sets as the active Page.
func (c *Creator) NewPage() *model.PdfPage {
	page := c.newPage()
	c.pages = append(c.pages, page)
	c.context.Page++
	return page
}

// AddPage adds the specified page to the creator.
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

// RotateDeg rotates the current active page by angle degrees.  An error is returned on failure,
// which can be if there is no currently active page, or the angleDeg is not a multiple of 90 degrees.
func (c *Creator) RotateDeg(angleDeg int64) error {
	page := c.getActivePage()
	if page == nil {
		common.Log.Debug("Fail to rotate: no page currently active")
		return errors.New("no page active")
	}
	if angleDeg%90 != 0 {
		common.Log.Debug("ERROR: Page rotation angle not a multiple of 90")
		return errors.New("range check error")
	}

	// Do the rotation.
	var rotation int64
	if page.Rotate != nil {
		rotation = *(page.Rotate)
	}
	rotation += angleDeg // Rotate by angleDeg degrees.
	page.Rotate = &rotation

	return nil
}

// Context returns the current drawing context.
func (c *Creator) Context() DrawContext {
	return c.context
}

// Call before writing out. Takes care of adding headers and footers, as well
// as generating front Page and table of contents.
func (c *Creator) finalize() error {
	totPages := len(c.pages)

	// Estimate number of additional generated pages and update TOC.
	genpages := 0
	if c.genFrontPageFunc != nil {
		genpages++
	}
	if c.AddTOC {
		c.initContext()
		c.context.Page = genpages + 1

		if c.genTableOfContentFunc != nil {
			if err := c.genTableOfContentFunc(c.toc); err != nil {
				return err
			}
		}

		// Make an estimate of the number of pages.
		blocks, _, err := c.toc.GeneratePageBlocks(c.context)
		if err != nil {
			common.Log.Debug("Failed to generate blocks: %v", err)
			return err
		}
		genpages += len(blocks)

		// Update the table of content Page numbers, accounting for front Page and TOC.
		lines := c.toc.Lines()
		for _, line := range lines {
			pageNum, err := strconv.Atoi(line.Page.Text)
			if err != nil {
				continue
			}

			line.Page.Text = strconv.Itoa(pageNum + genpages)
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

		args := FrontpageFunctionArgs{
			PageNum:    1,
			TotalPages: totPages,
		}
		c.genFrontPageFunc(args)
		hasFrontPage = true
	}

	if c.AddTOC {
		c.initContext()

		if c.genTableOfContentFunc != nil {
			if err := c.genTableOfContentFunc(c.toc); err != nil {
				common.Log.Debug("Error generating TOC: %v", err)
				return err
			}
		}

		// Account for the front page and the table of content pages.
		lines := c.toc.Lines()
		for _, line := range lines {
			line.linkPage += int64(genpages)
		}

		// Create TOC pages.
		var tocpages []*model.PdfPage
		blocks, _, _ := c.toc.GeneratePageBlocks(c.context)

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

	// Account for the front page and the table of content pages.
	if c.outline != nil && c.AddOutlines {
		var adjustOutlineDest func(item *model.OutlineItem)
		adjustOutlineDest = func(item *model.OutlineItem) {
			item.Dest.Page += int64(genpages)

			// Reverse the Y axis of the destination coordinates.
			// The user passes in the annotation coordinates as if
			// position 0, 0 is at the top left of the page.
			// However, position 0, 0 in the PDF is at the bottom
			// left of the page.
			item.Dest.Y = c.pageHeight - item.Dest.Y

			outlineItems := item.Items()
			for _, outlineItem := range outlineItems {
				adjustOutlineDest(outlineItem)
			}
		}

		outlineItems := c.outline.Items()
		for _, outlineItem := range outlineItems {
			adjustOutlineDest(outlineItem)
		}

		// Add outline TOC item.
		if c.AddTOC {
			var tocPage int64
			if hasFrontPage {
				tocPage = 1
			}

			c.outline.Insert(0, model.NewOutlineItem(
				"Table of Contents",
				model.NewOutlineDest(tocPage, 0, c.pageHeight),
			))
		}
	}

	for idx, page := range c.pages {
		c.setActivePage(page)

		// Draw page header.
		if c.drawHeaderFunc != nil {
			// Prepare a block to draw on.
			// Header is drawn on the top of the page. Has width of the page, but height limited to
			// the page margin top height.
			headerBlock := NewBlock(c.pageWidth, c.pageMargins.top)
			args := HeaderFunctionArgs{
				PageNum:    idx + 1,
				TotalPages: totPages,
			}
			c.drawHeaderFunc(headerBlock, args)
			headerBlock.SetPos(0, 0)

			if err := c.Draw(headerBlock); err != nil {
				common.Log.Debug("ERROR: drawing header: %v", err)
				return err
			}
		}

		// Draw page footer.
		if c.drawFooterFunc != nil {
			// Prepare a block to draw on.
			// Footer is drawn on the bottom of the page. Has width of the page, but height limited
			// to the page margin bottom height.
			footerBlock := NewBlock(c.pageWidth, c.pageMargins.bottom)
			args := FooterFunctionArgs{
				PageNum:    idx + 1,
				TotalPages: totPages,
			}
			c.drawFooterFunc(footerBlock, args)
			footerBlock.SetPos(0, c.pageHeight-footerBlock.height)

			if err := c.Draw(footerBlock); err != nil {
				common.Log.Debug("ERROR: drawing footer: %v", err)
				return err
			}
		}

		// Draw page blocks.
		block, ok := c.pageBlocks[page]
		if !ok {
			continue
		}
		if err := block.drawToPage(page); err != nil {
			common.Log.Debug("ERROR: drawing page %d blocks: %v", idx+1, err)
			return err
		}
	}

	c.finalized = true
	return nil
}

// MoveTo moves the drawing context to absolute coordinates (x, y).
func (c *Creator) MoveTo(x, y float64) {
	c.context.X = x
	c.context.Y = y
}

// MoveX moves the drawing context to absolute position x.
func (c *Creator) MoveX(x float64) {
	c.context.X = x
}

// MoveY moves the drawing context to absolute position y.
func (c *Creator) MoveY(y float64) {
	c.context.Y = y
}

// MoveRight moves the drawing context right by relative displacement dx (negative goes left).
func (c *Creator) MoveRight(dx float64) {
	c.context.X += dx
}

// MoveDown moves the drawing context down by relative displacement dy (negative goes up).
func (c *Creator) MoveDown(dy float64) {
	c.context.Y += dy
}

// Draw draws the Drawable widget to the document.  This can span over 1 or more pages. Additional
// pages are added if the contents go over the current Page.
func (c *Creator) Draw(d Drawable) error {
	if c.getActivePage() == nil {
		// Add a new Page if none added already.
		c.NewPage()
	}

	blocks, ctx, err := d.GeneratePageBlocks(c.context)
	if err != nil {
		return err
	}

	for idx, block := range blocks {
		if idx > 0 {
			c.NewPage()
		}

		page := c.getActivePage()
		if pageBlock, ok := c.pageBlocks[page]; ok {
			if err := pageBlock.mergeBlocks(block); err != nil {
				return err
			}
			if err := mergeResources(block.resources, pageBlock.resources); err != nil {
				return err
			}
		} else {
			c.pageBlocks[page] = block
		}
	}

	// Inner elements can affect X, Y position and available height.
	c.context.X = ctx.X
	c.context.Y = ctx.Y
	c.context.Height = ctx.PageHeight - ctx.Y - ctx.Margins.bottom

	return nil
}

// Write output of creator to io.Writer interface.
func (c *Creator) Write(ws io.Writer) error {
	if !c.finalized {
		if err := c.finalize(); err != nil {
			return err
		}
	}

	pdfWriter := model.NewPdfWriter()
	pdfWriter.SetOptimizer(c.optimizer)

	// Form fields.
	if c.acroForm != nil {
		err := pdfWriter.SetForms(c.acroForm)
		if err != nil {
			common.Log.Debug("Failure: %v", err)
			return err
		}
	}

	// Outlines.
	if c.externalOutline != nil {
		pdfWriter.AddOutlineTree(c.externalOutline)
	} else if c.outline != nil && c.AddOutlines {
		pdfWriter.AddOutlineTree(&c.outline.ToPdfOutline().PdfOutlineTreeNode)
	}

	// Pdf Writer access hook. Can be used to encrypt, etc. via the PdfWriter instance.
	if c.pdfWriterAccessFunc != nil {
		err := c.pdfWriterAccessFunc(&pdfWriter)
		if err != nil {
			common.Log.Debug("Failure: %v", err)
			return err
		}
	}

	for _, page := range c.pages {
		err := pdfWriter.AddPage(page)
		if err != nil {
			common.Log.Error("Failed to add Page: %v", err)
			return err
		}
	}

	err := pdfWriter.Write(ws)
	if err != nil {
		return err
	}

	return nil
}

// SetPdfWriterAccessFunc sets a PdfWriter access function/hook.
// Exposes the PdfWriter just prior to writing the PDF.  Can be used to encrypt the output PDF, etc.
//
// Example of encrypting with a user/owner password "password"
// Prior to calling c.WriteFile():
//
// c.SetPdfWriterAccessFunc(func(w *model.PdfWriter) error {
//	userPass := []byte("password")
//	ownerPass := []byte("password")
//	err := w.Encrypt(userPass, ownerPass, nil)
//	return err
// })
//
func (c *Creator) SetPdfWriterAccessFunc(pdfWriterAccessFunc func(writer *model.PdfWriter) error) {
	c.pdfWriterAccessFunc = pdfWriterAccessFunc
}

// WriteToFile writes the Creator output to file specified by path.
func (c *Creator) WriteToFile(outputPath string) error {
	fWrite, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer fWrite.Close()

	return c.Write(fWrite)
}

/*
Component creation methods.
*/

// NewTextStyle creates a new text style object which can be used to style
// chunks of text.
// Default attributes:
// Font: Helvetica
// Font size: 10
// Encoding: WinAnsiEncoding
// Text color: black
func (c *Creator) NewTextStyle() TextStyle {
	return newTextStyle(c.defaultFontRegular)
}

// NewParagraph creates a new text paragraph.
// Default attributes:
// Font: Helvetica,
// Font size: 10
// Encoding: WinAnsiEncoding
// Wrap: enabled
// Text color: black
func (c *Creator) NewParagraph(text string) *Paragraph {
	return newParagraph(text, c.NewTextStyle())
}

// NewStyledParagraph creates a new styled paragraph.
// Default attributes:
// Font: Helvetica,
// Font size: 10
// Encoding: WinAnsiEncoding
// Wrap: enabled
// Text color: black
func (c *Creator) NewStyledParagraph() *StyledParagraph {
	return newStyledParagraph(c.NewTextStyle())
}

// NewTable create a new Table with a specified number of columns.
func (c *Creator) NewTable(cols int) *Table {
	return newTable(cols)
}

// NewDivision returns a new Division container component.
func (c *Creator) NewDivision() *Division {
	return newDivision()
}

// NewTOC creates a new table of contents.
func (c *Creator) NewTOC(title string) *TOC {
	headingStyle := c.NewTextStyle()
	headingStyle.Font = c.defaultFontBold

	return newTOC(title, c.NewTextStyle(), headingStyle)
}

// NewTOCLine creates a new table of contents line with the default style.
func (c *Creator) NewTOCLine(number, title, page string, level uint) *TOCLine {
	return newTOCLine(number, title, page, level, c.NewTextStyle())
}

// NewStyledTOCLine creates a new table of contents line with the provided style.
func (c *Creator) NewStyledTOCLine(number, title, page TextChunk, level uint, style TextStyle) *TOCLine {
	return newStyledTOCLine(number, title, page, level, style)
}

// NewChapter creates a new chapter with the specified title as the heading.
func (c *Creator) NewChapter(title string) *Chapter {
	c.chapters++
	style := c.NewTextStyle()
	style.FontSize = 16

	return newChapter(nil, c.toc, c.outline, title, c.chapters, style)
}

// NewInvoice returns an instance of an empty invoice.
func (c *Creator) NewInvoice() *Invoice {
	headingStyle := c.NewTextStyle()
	headingStyle.Font = c.defaultFontBold

	return newInvoice(c.NewTextStyle(), headingStyle)
}

// NewList creates a new list.
func (c *Creator) NewList() *List {
	return newList(c.NewTextStyle())
}

// NewRectangle creates a new Rectangle with default parameters
// with left corner at (x,y) and width, height as specified.
func (c *Creator) NewRectangle(x, y, width, height float64) *Rectangle {
	return newRectangle(x, y, width, height)
}

// NewPageBreak create a new page break.
func (c *Creator) NewPageBreak() *PageBreak {
	return newPageBreak()
}

// NewLine creates a new Line with default parameters between (x1,y1) to (x2,y2).
func (c *Creator) NewLine(x1, y1, x2, y2 float64) *Line {
	return newLine(x1, y1, x2, y2)
}

// NewFilledCurve returns a instance of filled curve.
func (c *Creator) NewFilledCurve() *FilledCurve {
	return newFilledCurve()
}

// NewEllipse creates a new ellipse centered at (xc,yc) with a width and height specified.
func (c *Creator) NewEllipse(xc, yc, width, height float64) *Ellipse {
	return newEllipse(xc, yc, width, height)
}

// NewCurve returns new instance of Curve between points (x1,y1) and (x2, y2) with control point (cx,cy).
func (c *Creator) NewCurve(x1, y1, cx, cy, x2, y2 float64) *Curve {
	return newCurve(x1, y1, cx, cy, x2, y2)
}

// NewImage create a new image from a unidoc image (model.Image).
func (c *Creator) NewImage(img *model.Image) (*Image, error) {
	return newImage(img)
}

// NewImageFromData creates an Image from image data.
func (c *Creator) NewImageFromData(data []byte) (*Image, error) {
	return newImageFromData(data)
}

// NewImageFromFile creates an Image from a file.
func (c *Creator) NewImageFromFile(path string) (*Image, error) {
	return newImageFromFile(path)
}

// NewImageFromGoImage creates an Image from a go image.Image data structure.
func (c *Creator) NewImageFromGoImage(goimg goimage.Image) (*Image, error) {
	return newImageFromGoImage(goimg)
}
