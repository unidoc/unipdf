/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"bytes"
	"math"
	"strconv"
	"strings"

	cssSelector "github.com/ericchiang/css"
	"github.com/vanng822/css"
	"golang.org/x/net/html"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/model"
)

type htmlMargin struct {
	marginLeft   float64
	marginRight  float64
	marginTop    float64
	marginBottom float64
}

type htmlPadding struct {
	paddingLeft   float64
	paddingRight  float64
	paddingTop    float64
	paddingBottom float64
}

// htmlElementStyle is one element only style
type htmlElementStyle struct {
	// block style
	htmlMargin
	htmlPadding

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

func (style *htmlBlockStyle) addCSSStyle(tag, property, value string) {
	switch property {
	case "text-align":
		switch value {
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
		w, _ := strconv.ParseFloat(value, 64)
		es.borderWidthBottom = w
		es.borderWidthLeft = w
		es.borderWidthRight = w
		es.borderWidthTop = w
	case "border-color":
		es := style.getOrCreateElementStyle()
		c := getRGBColorFromHTML(value)
		es.borderColorBottom = model.NewPdfColorDeviceRGB(c.ToRGB())
		es.borderColorLeft = es.borderColorBottom
		es.borderColorRight = es.borderColorBottom
		es.borderColorTop = es.borderColorBottom
	case "border-style":
		es := style.getOrCreateElementStyle()
		switch value {
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
		style.Color = getRGBColorFromHTML(value)
	case "background-color":
		es := style.getOrCreateElementStyle()
		c := getRGBColorFromHTML(value)
		es.backgroundColor = model.NewPdfColorDeviceRGB(c.ToRGB())
	case "padding-left":
		es := style.getOrCreateElementStyle()
		es.paddingLeft, _ = strconv.ParseFloat(value, 64)
	case "padding-right":
		es := style.getOrCreateElementStyle()
		es.paddingRight, _ = strconv.ParseFloat(value, 64)
	case "padding-top":
		es := style.getOrCreateElementStyle()
		es.paddingTop, _ = strconv.ParseFloat(value, 64)
	case "padding-bottom":
		es := style.getOrCreateElementStyle()
		es.paddingBottom, _ = strconv.ParseFloat(value, 64)
	case "margin-left":
		es := style.getOrCreateElementStyle()
		es.marginLeft, _ = strconv.ParseFloat(value, 64)
	case "margin-right":
		es := style.getOrCreateElementStyle()
		es.marginRight, _ = strconv.ParseFloat(value, 64)
	case "margin-top":
		es := style.getOrCreateElementStyle()
		es.marginTop, _ = strconv.ParseFloat(value, 64)
	case "margin-bottom":
		es := style.getOrCreateElementStyle()
		es.marginBottom, _ = strconv.ParseFloat(value, 64)
	}
}

func (style *htmlBlockStyle) addEmbeddedCSS(tag string, csstext string) {
	ss := css.ParseBlock(csstext)
	for _, s := range ss {
		style.addCSSStyle(tag, s.Property, s.Value)
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
	if t.table != nil {
		return
	}
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
	owner      *HTMLContent
	parent     *htmlBlock
	tableStack *htmlTableStack
	styleStack *htmlStyleStack

	elements         []VectorDrawable
	currentParagraph *StyledParagraph
	style            htmlBlockStyle
}

func newHTMLBlock(parent *htmlBlock, style htmlBlockStyle) *htmlBlock {
	return &htmlBlock{
		owner:            parent.owner,
		parent:           parent,
		tableStack:       parent.tableStack,
		styleStack:       parent.styleStack,
		elements:         nil,
		currentParagraph: nil,
		style:            style,
	}
}

var ignoreReplacer = strings.NewReplacer("\r", "", "\n", "", "\t", " ")

func getAttributeValue(node *html.Node, attr string) string {
	for _, a := range node.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}

func (b *htmlBlock) parseNodeStyle(node *html.Node) htmlBlockStyle {
	style := b.styleStack.currentStyle()

	if cssStyles, found := b.owner.cssStylessByNode[node]; found {
		for _, cssStyle := range cssStyles {
			style.addCSSStyle(node.Data, cssStyle.Property, cssStyle.Value)
		}
	}

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

func isElementNode(node *html.Node, tag string) bool {
	if node == nil {
		return false
	}
	if node.Type != html.ElementNode {
		return false
	}
	return node.Data == tag
}

func (b *htmlBlock) processNode(node *html.Node) error {
	newB := b
	switch node.Type {
	case html.TextNode:
		text := ignoreReplacer.Replace(node.Data)
		if text == "" {
			return nil
		}

		isSpaces := strings.Count(text, " ") == len(text)
		if isSpaces && isElementNode(node.NextSibling, "p") {
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
		case "img":
			//imgB := newHTMLBlock(b, b.styleStack.currentStyle())
			//b.elements = append(b.elements, imgB)
			b.currentParagraph = nil
			//style := imgB.parseNodeStyle(node)
			//newB.style = style
			src := getAttributeValue(node, "src")
			if img := b.owner.getImage(src); img != nil {
				imgCopy := *img
				//newB.elements = []VectorDrawable{&imgCopy}
				b.elements = append(b.elements, &imgCopy)
			}
		case "br":
			p, created := b.getCurrentOrCreateParagraph()
			if created {
				b.currentParagraph.alignment = b.styleStack.currentStyle().TextAlignment
			} else {
				p.Append("\n")
			}
		case "b":
			style := b.parseNodeStyle(node)
			b.styleStack.pushStyle(style)
			defer b.styleStack.popStyle()
			b.styleStack.pushStyle(b.styleStack.addBold())
			defer b.styleStack.popStyle()
		case "i":
			style := b.parseNodeStyle(node)
			b.styleStack.pushStyle(style)
			defer b.styleStack.popStyle()
			b.styleStack.pushStyle(b.styleStack.addItalic())
			defer b.styleStack.popStyle()
		default:
			style := b.parseNodeStyle(node)
			b.styleStack.pushStyle(style)
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

type htmlCSSRule struct {
	rule     *css.CSSRule
	selector *cssSelector.Selector
}

// HTMLContent  allow to use simple HTML markup to generate content.
type HTMLContent struct {
	blocks            []*htmlBlock
	tableStack        htmlTableStack
	styleStack        htmlStyleStack
	cssRules          []htmlCSSRule
	cssStylessByNode  map[*html.Node][]*css.CSSStyleDeclaration
	images            map[string]*Image
	loadImageCallback func(string) *Image
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
	if es := b.style.ElementStyle; es != nil {
		w += es.marginLeft + es.marginRight + es.paddingLeft + es.paddingRight
	}
	return w
}

// Height returns the height of the Drawable.
func (b *htmlBlock) Height() float64 {
	if es := b.style.ElementStyle; es != nil && es.height != nil {
		return *es.height
	}
	var h float64
	for _, e := range b.elements {
		h += e.Height()
	}
	if es := b.style.ElementStyle; es != nil {
		h += es.marginTop + es.marginBottom + es.paddingTop + es.paddingBottom
	}
	return h
}

// SetWidth sets the width of the block.
func (b *htmlBlock) SetWidth(w float64) {
	b.style.getOrCreateElementStyle().width = &w
}

// SetHeight sets the height of the block.
func (b *htmlBlock) SetHeight(h float64) {
	b.style.getOrCreateElementStyle().height = &h
}

// GeneratePageBlocks generates the page blocks.  Multiple blocks are generated if the contents wrap
// over multiple pages. Implements the Drawable interface.
func (b *htmlBlock) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	var blocks []*Block
	origCtx := ctx
	if es := b.style.ElementStyle; es != nil {
		ctx.Y += es.marginTop
		ctx.Height -= es.marginTop + es.marginBottom
		ctx.X += es.marginLeft
		ctx.Width -= es.marginLeft + es.marginRight

		blockCtx := ctx
		blockCtx.Height = es.paddingTop + es.paddingBottom
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

		ctx.Y += es.paddingTop
		ctx.Height -= es.paddingTop + es.paddingBottom
		ctx.X += es.paddingLeft
		ctx.Width -= es.paddingLeft + es.paddingRight
	}

	for _, e := range b.elements {
		var newBlocks []*Block
		var err error
		var addY float64
		switch v := e.(type) {
		case *Image:
			v.SetPos(ctx.X, ctx.Y)
			v.ScaleToWidth(ctx.Width)
			addY = v.Height()
		}
		newBlocks, ctx, err = e.GeneratePageBlocks(ctx)
		if err != nil {
			return nil, ctx, err
		}
		ctx.Y += addY
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
	ctx.X = origCtx.X
	ctx.Width = origCtx.Width
	if es := b.style.ElementStyle; es != nil {
		ctx.Y += es.marginBottom
	}

	return blocks, ctx, nil
}

func newHTMLParagraph(baseStyle TextStyle) *HTMLContent {
	hp := HTMLContent{}
	hp.styleStack.RegularStyle = baseStyle
	hp.SetBoldFont(baseStyle.Font)
	hp.SetBoldItalicFont(baseStyle.Font)
	hp.SetItalicFont(baseStyle.Font)
	return &hp
}

// SetImage sets an image for the path.
func (h *HTMLContent) SetImage(path string, img *Image) {
	if h.images == nil {
		h.images = make(map[string]*Image)
	}
	h.images[path] = img
}

// SetGetImageCallback set a callback for load images.
func (h *HTMLContent) SetImageLoadCallback(callback func(string) *Image) {
	h.loadImageCallback = callback
}

func (h *HTMLContent) getImage(path string) *Image {
	if img, ok := h.images[path]; ok {
		return img
	}
	if h.loadImageCallback != nil {
		img := h.loadImageCallback(path)
		h.SetImage(path, img)
		return img
	}
	return nil
}

// SetRegularStyle sets the default text style.
func (h *HTMLContent) SetRegularStyle(style TextStyle) {
	h.styleStack.RegularStyle = style
}

// SetRegularFont sets the font for the default text style.
func (h *HTMLContent) SetRegularFont(font *model.PdfFont) {
	h.styleStack.RegularStyle.Font = font
}

// SetBoldFont sets the font for the bold text style.
func (h *HTMLContent) SetBoldFont(font *model.PdfFont) {
	h.styleStack.BoldFont = font
}

// SetItalicFont sets the font for the italic text style.
func (h *HTMLContent) SetItalicFont(font *model.PdfFont) {
	h.styleStack.ItalicFont = font
}

// SetBoldItalicFont sets the font for the bold and italic together text style.
func (h *HTMLContent) SetBoldItalicFont(font *model.PdfFont) {
	h.styleStack.BoldItalicFont = font
}

// Append adds html to paragraph.
func (h *HTMLContent) Append(htmlCode string) error {
	doc, err := html.Parse(bytes.NewBufferString(htmlCode))
	if err != nil {
		return err
	}
	h.addDocumentCSS(doc)
	h.processCSSRules(doc)
	newB := htmlBlock{
		owner:      h,
		parent:     nil,
		tableStack: &h.tableStack,
		styleStack: &h.styleStack,
	}
	h.blocks = append(h.blocks, &newB)
	return newB.processNode(doc)
}

// AddCSS adds CSS to the paragraph.
func (h *HTMLContent) AddCSS(cssText string) error {
	ss := css.Parse(cssText)
	rules := ss.GetCSSRuleList()
	for _, r := range rules {
		sel, err := cssSelector.Compile(r.Style.SelectorText)
		if err != nil {
			return err
		}
		h.cssRules = append(h.cssRules, htmlCSSRule{
			rule:     r,
			selector: sel,
		})
	}
	return nil
}

func (h *HTMLContent) addDocumentCSS(node *html.Node) error {
	if node.Type == html.ElementNode && node.Data == "style" && node.FirstChild != nil {
		if err := h.AddCSS(node.FirstChild.Data); err != nil {
			return err
		}
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		if err := h.addDocumentCSS(next); err != nil {
			return err
		}
	}
	return nil
}

func (h *HTMLContent) processCSSRules(doc *html.Node) {
	h.cssStylessByNode = make(map[*html.Node][]*css.CSSStyleDeclaration)
	for _, r := range h.cssRules {
		for _, node := range r.selector.Select(doc) {
			for _, style := range r.rule.Style.Styles {
				h.cssStylessByNode[node] = append(h.cssStylessByNode[node], style)
			}
		}
	}
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
func (h *HTMLContent) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	var blocks []*Block
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
	return blocks, ctx, nil
}

// Width returns the width of the Drawable.
func (h *HTMLContent) Width() float64 {
	var width float64
	for _, b := range h.blocks {
		width = math.Max(width, b.Width())
	}
	return width
}

// Height returns the height of the Drawable.
func (h *HTMLContent) Height() float64 {
	var height float64
	for _, b := range h.blocks {
		height += b.Height()
	}
	return height
}

// SetWidth sets the width of the block.
func (h *HTMLContent) SetWidth(w float64) {
	for _, b := range h.blocks {
		b.SetWidth(w)
	}
}
