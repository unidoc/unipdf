/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"bytes"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/vanng822/css"
	"golang.org/x/net/html"
)

// htmlElementStyle is one element only style
type htmlElementStyle struct {
	// block style

	// Background
	backgroundColor *model.PdfColorDeviceRGB

	borderLineStyle draw.LineStyle

	// border
	borderStyleLeft   CellBorderStyle
	borderColorLeft   *model.PdfColorDeviceRGB
	borderWidthLeft   float64
	borderStyleBottom CellBorderStyle
	borderColorBottom *model.PdfColorDeviceRGB
	borderWidthBottom float64
	borderStyleRight  CellBorderStyle
	borderColorRight  *model.PdfColorDeviceRGB
	borderWidthRight  float64
	borderStyleTop    CellBorderStyle
	borderColorTop    *model.PdfColorDeviceRGB
	borderWidthTop    float64

	width  *float64
	height *float64
}

type htmlBlockStyle struct {
	TextStyle
	Bold          bool
	Italic        bool
	TextAlignment TextAlignment
	ElementStyle  *htmlElementStyle
}

func (style *htmlBlockStyle) getOrCreateElementStyle() *htmlElementStyle {
	if style.ElementStyle == nil {
		style.ElementStyle = &htmlElementStyle{}
	}
	return style.ElementStyle
}

func (style *htmlBlockStyle) addEmbeddedCSS(tag string, csstext string) {
	ss := css.ParseBlock(csstext)
	for _, s := range ss {
		switch s.Property {
		case "text-align":
			switch s.Value {
			case "center":
				style.TextAlignment = TextAlignmentCenter
			case "left":
				style.TextAlignment = TextAlignmentLeft
			case "right":
				style.TextAlignment = TextAlignmentRight
			case "justify":
				style.TextAlignment = TextAlignmentJustify
			}
		case "border-width":
			es := style.getOrCreateElementStyle()
			w, _ := strconv.ParseFloat(s.Value, 64)
			es.borderWidthBottom = w
			es.borderWidthLeft = w
			es.borderWidthRight = w
			es.borderWidthTop = w
		case "border-color":
			es := style.getOrCreateElementStyle()
			c := getRGBColorFromHTML(s.Value)
			es.borderColorBottom = model.NewPdfColorDeviceRGB(c.ToRGB())
			es.borderColorLeft = es.borderColorBottom
			es.borderColorRight = es.borderColorBottom
			es.borderColorTop = es.borderColorBottom
		case "border-style":
			es := style.getOrCreateElementStyle()
			switch s.Value {
			case "solid":
				es.borderLineStyle = draw.LineStyleSolid
				es.borderStyleTop = CellBorderStyleSingle
				es.borderStyleRight = CellBorderStyleSingle
				es.borderStyleBottom = CellBorderStyleSingle
				es.borderStyleLeft = CellBorderStyleSingle
			case "dashed":
				es.borderLineStyle = draw.LineStyleDashed
				es.borderStyleTop = CellBorderStyleSingle
				es.borderStyleRight = CellBorderStyleSingle
				es.borderStyleBottom = CellBorderStyleSingle
				es.borderStyleLeft = CellBorderStyleSingle
			case "double":
				es.borderLineStyle = draw.LineStyleSolid
				es.borderStyleTop = CellBorderStyleDouble
				es.borderStyleRight = CellBorderStyleDouble
				es.borderStyleBottom = CellBorderStyleDouble
				es.borderStyleLeft = CellBorderStyleDouble
			default:
				es.borderStyleTop = CellBorderStyleNone
				es.borderStyleRight = CellBorderStyleNone
				es.borderStyleBottom = CellBorderStyleNone
				es.borderStyleLeft = CellBorderStyleNone
			}
		case "color":
			style.Color = getRGBColorFromHTML(s.Value)
		case "background-color":
			es := style.getOrCreateElementStyle()
			c := getRGBColorFromHTML(s.Value)
			es.backgroundColor = model.NewPdfColorDeviceRGB(c.ToRGB())
		}
	}
}

type htmlStyleStack struct {
	RegularStyle   TextStyle
	BoldFont       *model.PdfFont
	ItalicFont     *model.PdfFont
	BoldItalicFont *model.PdfFont
	styleStack     []htmlBlockStyle
}

func (s *htmlStyleStack) currentStyle() htmlBlockStyle {
	if len(s.styleStack) == 0 {
		return htmlBlockStyle{
			TextStyle: s.RegularStyle,
			Bold:      false,
			Italic:    false,
		}
	}
	return s.styleStack[len(s.styleStack)-1]
}

func (s *htmlStyleStack) pushStyle(style htmlBlockStyle) {
	style.ElementStyle = nil
	s.styleStack = append(s.styleStack, style)
}

func (s *htmlStyleStack) popStyle() htmlBlockStyle {
	style := s.currentStyle()
	if len(s.styleStack) < 1 {
		return style
	}
	s.styleStack = s.styleStack[:len(s.styleStack)-1]
	return style
}

func (s *htmlStyleStack) addBold() htmlBlockStyle {
	style := s.currentStyle()
	style.Bold = true
	if s.BoldFont != nil {
		style.Font = s.BoldFont
	}
	if style.Italic && s.BoldItalicFont != nil {
		style.Font = s.BoldItalicFont
	}
	return style
}

func (s *htmlStyleStack) addItalic() htmlBlockStyle {
	style := s.currentStyle()
	style.Italic = true
	if s.ItalicFont != nil {
		style.Font = s.ItalicFont
	}
	if style.Bold && s.BoldItalicFont != nil {
		style.Font = s.BoldItalicFont
	}
	return style
}

type htmlTableCell struct {
	block *htmlBlock
}

type htmlTableRow struct {
	cells []*htmlTableCell
}

type htmlTable struct {
	table       *Table
	maxColIndex int
	rows        []*htmlTableRow
}

func (t *htmlTable) generateContent() {
	if t.table == nil {
		t.table = newTable(t.maxColIndex)
		for _, row := range t.rows {
			for _, cell := range row.cells {
				c := t.table.NewCell()
				if cell.block != nil {
					c.SetContent(cell.block)
				}
			}
		}
	}
}

// Width returns the width of the Drawable.
func (t *htmlTable) Width() float64 {
	t.generateContent()
	return t.table.Width()
}

// Height returns the height of the Drawable.
func (t *htmlTable) Height() float64 {
	t.generateContent()
	return t.table.Height()
}

// GeneratePageBlocks generates the page blocks.  Multiple blocks are generated if the contents wrap
// over multiple pages. Implements the Drawable interface.
func (t *htmlTable) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	t.generateContent()
	return t.table.GeneratePageBlocks(ctx)
}

type htmlTableStack struct {
	tableStack []*htmlTable
}

func (st *htmlTableStack) createAndPushTable() *htmlTable {
	t := &htmlTable{}
	st.pushTable(t)
	return t
}

func (st *htmlTableStack) currentTable() *htmlTable {
	if len(st.tableStack) == 0 {
		return nil
	}
	return st.tableStack[len(st.tableStack)-1]
}

func (st *htmlTableStack) pushTable(table *htmlTable) {
	st.tableStack = append(st.tableStack, table)
}

func (st *htmlTableStack) popTable() *htmlTable {
	t := st.currentTable()
	if len(st.tableStack) < 1 {
		return t
	}
	st.tableStack = st.tableStack[:len(st.tableStack)-1]
	return t
}

type htmlBlock struct {
	owner      *HTMLParagraph
	parent     *htmlBlock
	tableStack *htmlTableStack
	styleStack *htmlStyleStack

	elements         []VectorDrawable
	currentParagraph *StyledParagraph
	style            htmlBlockStyle
}

func newHTMLBlock(parent *htmlBlock, style htmlBlockStyle) *htmlBlock {
	b := htmlBlock{
		owner:            parent.owner,
		parent:           parent,
		tableStack:       parent.tableStack,
		styleStack:       parent.styleStack,
		elements:         nil,
		currentParagraph: nil,
		style:            style,
	}
	return &b
}

var ignoreReplacer = strings.NewReplacer("\r", "", "\n", "", "\t", " ")

func (b *htmlBlock) parseNodeStyle(node *html.Node) htmlBlockStyle {
	style := b.styleStack.currentStyle()
	for _, attr := range node.Attr {
		switch attr.Key {
		case "style":
			style.addEmbeddedCSS(node.Data, attr.Val)
		case "align":
			switch attr.Val {
			case "center":
				style.TextAlignment = TextAlignmentCenter
			case "left":
				style.TextAlignment = TextAlignmentLeft
			case "right":
				style.TextAlignment = TextAlignmentRight
			case "justify":
				style.TextAlignment = TextAlignmentJustify
			}
		}
	}
	return style
}

func (b *htmlBlock) processNode(node *html.Node) error {
	newB := b

	switch node.Type {
	case html.TextNode:
		text := ignoreReplacer.Replace(node.Data)
		if text == "" {
			return nil
		}
		p, created := b.getCurrentOrCreateParagraph()
		if created {
			b.currentParagraph.alignment = b.styleStack.currentStyle().TextAlignment
		}
		p.Append(text).Style = b.styleStack.currentStyle().TextStyle
		return nil
	case html.ElementNode:

		switch node.Data {
		case "style":
			log.Println(node)
			return nil
		case "script":
			return nil
		case "table":
			t := b.tableStack.createAndPushTable()
			style := b.parseNodeStyle(node)
			b.styleStack.pushStyle(style)
			defer b.styleStack.popStyle()
			b.elements = append(b.elements, t)
			defer b.tableStack.popTable()
		case "tr":
			if t := b.tableStack.currentTable(); t != nil {
				t.rows = append(t.rows, &htmlTableRow{})
			}
		case "td", "th":
			if t := b.tableStack.currentTable(); t != nil && len(t.rows) > 0 {
				newB = newHTMLBlock(b, b.styleStack.currentStyle())
				style := newB.parseNodeStyle(node)
				newB.style = style
				newB.styleStack.pushStyle(style)
				defer newB.styleStack.popStyle()

				row := t.rows[len(t.rows)-1]
				cell := htmlTableCell{block: newB}
				row.cells = append(row.cells, &cell)
				if l := len(row.cells); l > t.maxColIndex {
					t.maxColIndex = l
				}
			}
			if node.Data == "th" {
				newB.styleStack.pushStyle(newB.styleStack.addBold())
				defer newB.styleStack.popStyle()
			}
		case "div":
			newB = newHTMLBlock(b, b.styleStack.currentStyle())
			b.elements = append(b.elements, newB)
			b.currentParagraph = nil
			style := newB.parseNodeStyle(node)
			newB.style = style
			newB.styleStack.pushStyle(style)
			defer newB.styleStack.popStyle()
		case "p":
			style := b.parseNodeStyle(node)
			b.styleStack.pushStyle(style)
			defer b.styleStack.popStyle()
			b.currentParagraph = newStyledParagraph(b.styleStack.currentStyle().TextStyle)
			b.currentParagraph.alignment = b.styleStack.currentStyle().TextAlignment
			b.elements = append(b.elements, b.currentParagraph)
		case "br":
			//b.currentParagraph = newStyledParagraph(b.styleStack.currentStyle().TextStyle)
			//b.currentParagraph.alignment = b.styleStack.currentStyle().TextAlignment
			//b.elements = append(b.elements, b.currentParagraph)
			//b.currentParagraph.Append("\n")
			p, created := b.getCurrentOrCreateParagraph()
			if created {
				b.currentParagraph.alignment = b.styleStack.currentStyle().TextAlignment
			} else {
				p.Append("\n")
			}
		case "b":
			b.styleStack.pushStyle(b.styleStack.addBold())
			defer b.styleStack.popStyle()
		case "i":
			b.styleStack.pushStyle(b.styleStack.addItalic())
			defer b.styleStack.popStyle()
		}
	}

	for next := node.FirstChild; next != nil; next = next.NextSibling {
		if err := newB.processNode(next); err != nil {
			return err
		}
	}
	return nil
}

// HTMLParagraph  allow to use simple HTML markup to generate content.
type HTMLParagraph struct {
	blocks     []*htmlBlock
	tableStack htmlTableStack
	styleStack htmlStyleStack
}

func (b *htmlBlock) getCurrentOrCreateParagraph() (*StyledParagraph, bool) {
	if b.currentParagraph == nil {
		b.currentParagraph = newStyledParagraph(b.styleStack.RegularStyle)
		b.elements = append(b.elements, b.currentParagraph)
		return b.currentParagraph, true
	}
	return b.currentParagraph, false
}

// Width returns the width of the Drawable.
func (b *htmlBlock) Width() float64 {
	if es := b.style.ElementStyle; es != nil && es.width != nil {
		return *es.width
	}
	var w float64
	for _, e := range b.elements {
		w = math.Max(w, e.Width())
	}
	return w
}

// Height returns the height of the Drawable.
func (b *htmlBlock) Height() float64 {
	var h float64
	for _, e := range b.elements {
		h = math.Max(h, e.Height())
	}
	return h
}

func (b *htmlBlock) SetWidth(w float64) {
	b.style.getOrCreateElementStyle().width = &w
}

// GeneratePageBlocks generates the page blocks.  Multiple blocks are generated if the contents wrap
// over multiple pages. Implements the Drawable interface.
func (b *htmlBlock) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	var blocks []*Block

	if es := b.style.ElementStyle; es != nil {
		blockCtx := ctx
		blockCtx.Height = 0.0
		for _, e := range b.elements {
			blockCtx.Height += e.Height()
		}

		block := NewBlock(blockCtx.PageWidth, blockCtx.PageHeight)
		block.xPos = ctx.X
		block.yPos = ctx.Y
		blocks = append(blocks, block)
		border := newBorder(blockCtx.X, blockCtx.Y, blockCtx.Width, blockCtx.Height)

		if es.backgroundColor != nil {
			r := es.backgroundColor.R()
			g := es.backgroundColor.G()
			b := es.backgroundColor.B()

			border.SetFillColor(ColorRGBFromArithmetic(r, g, b))
		}

		border.LineStyle = es.borderLineStyle

		border.styleLeft = es.borderStyleLeft
		border.styleRight = es.borderStyleRight
		border.styleTop = es.borderStyleTop
		border.styleBottom = es.borderStyleBottom

		if es.borderColorLeft != nil {
			border.SetColorLeft(ColorRGBFromArithmetic(es.borderColorLeft.R(), es.borderColorLeft.G(), es.borderColorLeft.B()))
		}
		if es.borderColorBottom != nil {
			border.SetColorBottom(ColorRGBFromArithmetic(es.borderColorBottom.R(), es.borderColorBottom.G(), es.borderColorBottom.B()))
		}
		if es.borderColorRight != nil {
			border.SetColorRight(ColorRGBFromArithmetic(es.borderColorRight.R(), es.borderColorRight.G(), es.borderColorRight.B()))
		}
		if es.borderColorTop != nil {
			border.SetColorTop(ColorRGBFromArithmetic(es.borderColorTop.R(), es.borderColorTop.G(), es.borderColorTop.B()))
		}

		border.SetWidthBottom(es.borderWidthBottom)
		border.SetWidthLeft(es.borderWidthLeft)
		border.SetWidthRight(es.borderWidthRight)
		border.SetWidthTop(es.borderWidthTop)

		err := block.Draw(border)
		if err != nil {
			common.Log.Debug("ERROR: %v", err)
		}
	}

	for _, e := range b.elements {
		var newBlocks []*Block
		var err error
		newBlocks, ctx, err = e.GeneratePageBlocks(ctx)
		if err != nil {
			return nil, ctx, err
		}
		if len(newBlocks) < 1 {
			continue
		}
		if len(blocks) == 0 {
			blocks = newBlocks[0:1]
		} else {
			blocks[len(blocks)-1].mergeBlocks(newBlocks[0])
		}

		blocks = append(blocks, newBlocks[1:]...)
	}
	return blocks, ctx, nil
}

// NewHTMLParagraph creates a new html paragraph.
// Default attributes:
// Font: Helvetica,
// Font size: 10
// Encoding: WinAnsiEncoding
// Wrap: enabled
// Text color: black
func (c *Creator) NewHTMLParagraph() *HTMLParagraph {
	return newHTMLParagraph(c.NewTextStyle())
}

func newHTMLParagraph(baseStyle TextStyle) *HTMLParagraph {
	hp := HTMLParagraph{}
	hp.styleStack.RegularStyle = baseStyle
	return &hp
}

// SetRegularStyle sets the default text style.
func (h *HTMLParagraph) SetRegularStyle(style TextStyle) {
	h.styleStack.RegularStyle = style
}

// SetRegularFont sets the font for the default text style.
func (h *HTMLParagraph) SetRegularFont(font *model.PdfFont) {
	h.styleStack.RegularStyle.Font = font
}

// SetBoldFont sets the font for the bold text style.
func (h *HTMLParagraph) SetBoldFont(font *model.PdfFont) {
	h.styleStack.BoldFont = font
}

// SetItalicFont sets the font for the italic text style.
func (h *HTMLParagraph) SetItalicFont(font *model.PdfFont) {
	h.styleStack.ItalicFont = font
}

// SetBoldItalicFont sets the font for the bold and italic together text style.
func (h *HTMLParagraph) SetBoldItalicFont(font *model.PdfFont) {
	h.styleStack.BoldItalicFont = font
}

// Append adds html to paragraph.
func (h *HTMLParagraph) Append(htmlCode string) error {
	doc, err := html.Parse(bytes.NewBufferString(htmlCode))
	if err != nil {
		return err
	}
	newB := htmlBlock{
		owner:      h,
		parent:     nil,
		tableStack: &h.tableStack,
		styleStack: &h.styleStack,
	}
	h.blocks = append(h.blocks, &newB)
	return newB.processNode(doc)
}

var stdHTMLColors = map[string]Color{
	"blue":   ColorBlue,
	"black":  ColorBlack,
	"green":  ColorGreen,
	"red":    ColorRed,
	"white":  ColorWhite,
	"yellow": ColorYellow,
}

func getRGBColorFromHTML(color string) Color {
	if c, ok := stdHTMLColors[color]; ok {
		return c
	}
	return ColorRGBFromHex(color)
}

// GeneratePageBlocks generates the page blocks.  Multiple blocks are generated if the contents wrap
// over multiple pages. Implements the Drawable interface.
func (h *HTMLParagraph) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	var blocks []*Block
	origCtx := ctx
	for _, e := range h.blocks {
		var newBlocks []*Block
		var err error
		newBlocks, ctx, err = e.GeneratePageBlocks(ctx)
		if err != nil {
			return nil, ctx, err
		}
		if len(newBlocks) < 1 {
			continue
		}
		if len(blocks) == 0 {
			blocks = newBlocks[0:1]
		} else {
			blocks[len(blocks)-1].mergeBlocks(newBlocks[0])
		}

		blocks = append(blocks, newBlocks[1:]...)
	}
	return blocks, origCtx, nil
}
