/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import "fmt"

// InvoiceAddress.
type InvoiceAddress struct {
	Name    string
	Street  string
	Zip     string
	City    string
	Country string
	Phone   string
	Email   string
}

// InvoiceCellProps.
type InvoiceCellProps struct {
	TextStyle       TextStyle
	Alignment       CellHorizontalAlignment
	BackgroundColor Color

	BorderColor Color
	BorderWidth float64
	BorderSides []CellBorderSide
}

// InvoiceCell.
type InvoiceCell struct {
	InvoiceCellProps

	Value string
}

// Invoice represents a configurable invoice template.
type Invoice struct {
	// Invoice title.
	title string

	// Invoice logo.
	logo *Image

	// Invoice addresses.
	buyerAddress  *InvoiceAddress
	sellerAddress *InvoiceAddress

	// Invoice information.
	number  [2]*InvoiceCell
	date    [2]*InvoiceCell
	dueDate [2]*InvoiceCell
	info    [][2]*InvoiceCell

	// Invoice lines.
	columns []*InvoiceCell
	lines   [][]*InvoiceCell

	// Invoice totals.
	subtotal [2]*InvoiceCell
	total    [2]*InvoiceCell
	totals   [][2]*InvoiceCell

	// Invoice note sections.
	notes    [2]string
	terms    [2]string
	sections [][2]string

	// Invoice styles.
	defaultStyle TextStyle
	headingStyle TextStyle
	titleStyle   TextStyle

	infoProps  InvoiceCellProps
	colProps   InvoiceCellProps
	itemProps  InvoiceCellProps
	totalProps InvoiceCellProps

	// Positioning: relative/absolute.
	positioning positioning
}

// newInvoice returns an instance of an empty invoice.
func newInvoice(defaultStyle, headingStyle TextStyle) *Invoice {
	i := &Invoice{
		// Title.
		title: "INVOICE",

		// Addresses.
		sellerAddress: &InvoiceAddress{},
		buyerAddress:  &InvoiceAddress{},

		// Styles.
		defaultStyle: defaultStyle,
		headingStyle: headingStyle,
	}

	// Default colors.
	lightGrey := ColorRGBFrom8bit(245, 245, 245)
	mediumGrey := ColorRGBFrom8bit(155, 155, 155)

	// Default title style.
	i.titleStyle = headingStyle
	i.titleStyle.Color = mediumGrey
	i.titleStyle.FontSize = 20

	// Invoice information default properties.
	i.infoProps = i.NewCellProps()
	i.infoProps.BackgroundColor = lightGrey
	i.infoProps.TextStyle = headingStyle

	// Invoice line items default properties.
	i.colProps = i.NewCellProps()
	i.colProps.TextStyle = headingStyle
	i.colProps.BackgroundColor = lightGrey
	i.colProps.BorderColor = lightGrey

	i.itemProps = i.NewCellProps()
	i.itemProps.BorderColor = lightGrey
	i.itemProps.BorderSides = []CellBorderSide{CellBorderSideBottom}
	i.itemProps.Alignment = CellHorizontalAlignmentRight

	// Invoice totals default properties.
	i.totalProps = i.NewCellProps()
	i.totalProps.Alignment = CellHorizontalAlignmentRight

	// Invoice information fields.
	i.number = [2]*InvoiceCell{
		i.newCell("Invoice number", i.infoProps),
		i.newCell("", i.infoProps),
	}

	i.date = [2]*InvoiceCell{
		i.newCell("Date", i.infoProps),
		i.newCell("", i.infoProps),
	}

	i.dueDate = [2]*InvoiceCell{
		i.newCell("Date", i.infoProps),
		i.newCell("", i.infoProps),
	}

	// Invoice totals fields.
	i.subtotal = [2]*InvoiceCell{
		i.newCell("Subtotal", i.totalProps),
		i.newCell("", i.totalProps),
	}

	totalProps := i.totalProps
	totalProps.TextStyle = headingStyle
	totalProps.BackgroundColor = lightGrey
	totalProps.BorderColor = lightGrey

	i.total = [2]*InvoiceCell{
		i.newCell("Total", totalProps),
		i.newCell("", totalProps),
	}

	// Invoice notes fields.
	i.notes = [2]string{"Notes", ""}
	i.terms = [2]string{"Terms and conditions", ""}

	// Default item columns.
	i.columns = []*InvoiceCell{
		i.newColumn("Description", CellHorizontalAlignmentLeft),
		i.newColumn("Quantity", CellHorizontalAlignmentRight),
		i.newColumn("Unit price", CellHorizontalAlignmentRight),
		i.newColumn("Amount", CellHorizontalAlignmentRight),
	}

	return i
}

func (i *Invoice) Title() string {
	return i.title
}

func (i *Invoice) SetTitle(title string) {
	i.title = title
}

func (i *Invoice) Logo() *Image {
	return i.logo
}

func (i *Invoice) SetLogo(logo *Image) {
	i.logo = logo
}

func (i *Invoice) SellerAddress() *InvoiceAddress {
	return i.sellerAddress
}

func (i *Invoice) SetSellerAddress(address *InvoiceAddress) {
	i.sellerAddress = address
}

func (i *Invoice) BuyerAddress() *InvoiceAddress {
	return i.buyerAddress
}

func (i *Invoice) SetBuyerAddress(address *InvoiceAddress) {
	i.buyerAddress = address
}

func (i *Invoice) Number() (*InvoiceCell, *InvoiceCell) {
	return i.number[0], i.number[1]
}

func (i *Invoice) SetNumber(number string) (*InvoiceCell, *InvoiceCell) {
	i.number[1].Value = number
	return i.number[0], i.number[1]
}

func (i *Invoice) Date() (*InvoiceCell, *InvoiceCell) {
	return i.date[0], i.date[1]
}

func (i *Invoice) SetDate(date string) (*InvoiceCell, *InvoiceCell) {
	i.date[1].Value = date
	return i.date[0], i.date[1]
}

func (i *Invoice) DueDate() (*InvoiceCell, *InvoiceCell) {
	return i.dueDate[0], i.dueDate[1]
}

func (i *Invoice) SetDueDate(dueDate string) (*InvoiceCell, *InvoiceCell) {
	i.dueDate[1].Value = dueDate
	return i.dueDate[0], i.dueDate[1]
}

func (i *Invoice) InfoLines() [][2]*InvoiceCell {
	info := [][2]*InvoiceCell{
		i.number,
		i.date,
		i.dueDate,
	}

	return append(info, i.info...)
}

func (i *Invoice) AddInfo(description, value string) (*InvoiceCell, *InvoiceCell) {
	info := [2]*InvoiceCell{
		i.newCell(description, i.infoProps),
		i.newCell(value, i.infoProps),
	}

	i.info = append(i.info, info)
	return info[0], info[1]
}

func (i *Invoice) Columns() []*InvoiceCell {
	return i.columns
}

func (i *Invoice) AppendColumn(description string) *InvoiceCell {
	col := i.NewColumn(description)
	i.columns = append(i.columns, col)
	return col
}

func (i *Invoice) InsertColumn(index uint, description string) *InvoiceCell {
	l := uint(len(i.columns))
	if index > l {
		index = l
	}

	col := i.NewColumn(description)
	i.columns = append(i.columns[:index], append([]*InvoiceCell{col}, i.columns[index:]...)...)
	return col
}

func (i *Invoice) Lines() [][]*InvoiceCell {
	return i.lines
}

func (i *Invoice) AddLine(values ...string) []*InvoiceCell {
	lenCols := len(i.columns)

	var line []*InvoiceCell
	for j, value := range values {
		itemCell := i.newCell(value, i.itemProps)
		if j < lenCols {
			itemCell.Alignment = i.columns[j].Alignment
		}

		line = append(line, itemCell)
	}

	i.lines = append(i.lines, line)
	return line
}

func (i *Invoice) Subtotal() (*InvoiceCell, *InvoiceCell) {
	return i.subtotal[0], i.subtotal[1]
}

func (i *Invoice) SetSubtotal(value string) {
	i.subtotal[1].Value = value
}

func (i *Invoice) Total() (*InvoiceCell, *InvoiceCell) {
	return i.total[0], i.total[1]
}

func (i *Invoice) SetTotal(value string) {
	i.total[1].Value = value
}

func (i *Invoice) TotalLines() [][2]*InvoiceCell {
	totals := [][2]*InvoiceCell{i.subtotal}
	totals = append(totals, i.totals...)
	return append(totals, i.total)
}

func (i *Invoice) AddTotalLine(desc, value string) (*InvoiceCell, *InvoiceCell) {
	descCell := &InvoiceCell{
		i.totalProps,
		desc,
	}
	valueCell := &InvoiceCell{
		i.totalProps,
		value,
	}

	i.totals = append(i.totals, [2]*InvoiceCell{descCell, valueCell})
	return descCell, valueCell
}

func (i *Invoice) Notes() (string, string) {
	return i.notes[0], i.notes[1]
}

func (i *Invoice) SetNotes(title, content string) {
	i.notes = [2]string{
		title,
		content,
	}
}

func (i *Invoice) Terms() (string, string) {
	return i.terms[0], i.terms[1]
}

func (i *Invoice) SetTerms(title, content string) {
	i.terms = [2]string{
		title,
		content,
	}
}

func (i *Invoice) Sections() [][2]string {
	return i.sections
}

func (i *Invoice) AddSection(title, content string) {
	i.sections = append(i.sections, [2]string{
		title,
		content,
	})
}

func (i *Invoice) NewCellProps() InvoiceCellProps {
	white := ColorRGBFrom8bit(255, 255, 255)

	return InvoiceCellProps{
		TextStyle:       i.defaultStyle,
		Alignment:       CellHorizontalAlignmentLeft,
		BackgroundColor: white,
		BorderColor:     white,
		BorderWidth:     1,
		BorderSides:     []CellBorderSide{CellBorderSideAll},
	}
}

func (i *Invoice) NewCell(value string) *InvoiceCell {
	return i.newCell(value, i.NewCellProps())
}

func (i *Invoice) newCell(value string, props InvoiceCellProps) *InvoiceCell {
	return &InvoiceCell{
		props,
		value,
	}
}

func (i *Invoice) NewColumn(description string) *InvoiceCell {
	return i.newColumn(description, CellHorizontalAlignmentLeft)
}

func (i *Invoice) newColumn(description string, alignment CellHorizontalAlignment) *InvoiceCell {
	col := &InvoiceCell{i.colProps, description}
	col.Alignment = alignment

	return col
}

func (i *Invoice) setTitleStyle(style TextStyle) {
	i.titleStyle = style
}

func (i *Invoice) setCellBorder(cell *TableCell, invoiceCell *InvoiceCell) {
	for _, side := range invoiceCell.BorderSides {
		cell.SetBorder(side, CellBorderStyleSingle, invoiceCell.BorderWidth)
	}

	cell.SetBorderColor(invoiceCell.BorderColor)
}

func (i *Invoice) drawAddress(title, name string, addr *InvoiceAddress) []*StyledParagraph {
	var paragraphs []*StyledParagraph

	// Address title.
	if title != "" {
		titleParagraph := newStyledParagraph(i.headingStyle)
		titleParagraph.SetMargins(0, 0, 0, 7)
		titleParagraph.Append(title)

		paragraphs = append(paragraphs, titleParagraph)
	}

	// Address information.
	addressParagraph := newStyledParagraph(i.defaultStyle)
	addressParagraph.SetLineHeight(1.2)

	city := addr.City
	if addr.Zip != "" {
		if city != "" {
			city += ", "
		}

		city += addr.Zip
	}

	if name != "" {
		addressParagraph.Append(addr.Name + "\n")
	}
	if addr.Street != "" {
		addressParagraph.Append(addr.Street + "\n")
	}
	if city != "" {
		addressParagraph.Append(city + "\n")
	}
	if addr.Country != "" {
		addressParagraph.Append(addr.Country + "\n")
	}

	// Contact information.
	contactParagraph := newStyledParagraph(i.defaultStyle)
	contactParagraph.SetLineHeight(1.2)
	contactParagraph.SetMargins(0, 0, 7, 0)

	if addr.Phone != "" {
		contactParagraph.Append(fmt.Sprintf("Phone: %s\n", addr.Phone))
	}
	if addr.Email != "" {
		contactParagraph.Append(fmt.Sprintf("Email: %s\n", addr.Email))
	}

	paragraphs = append(paragraphs, addressParagraph, contactParagraph)
	return paragraphs
}

func (i *Invoice) drawSection(title, content string) []*StyledParagraph {
	var paragraphs []*StyledParagraph

	// Title paragraph.
	if title != "" {
		titleParagraph := newStyledParagraph(i.headingStyle)
		titleParagraph.SetMargins(0, 0, 0, 5)
		titleParagraph.Append(title)

		paragraphs = append(paragraphs, titleParagraph)
	}

	// Content paragraph.
	if content != "" {
		contentParagraph := newStyledParagraph(i.defaultStyle)
		contentParagraph.Append(content)

		paragraphs = append(paragraphs, contentParagraph)
	}

	return paragraphs
}

func (i *Invoice) drawInformation() *Table {
	table := newTable(2)

	info := append([][2]*InvoiceCell{
		i.number,
		i.date,
		i.dueDate,
	}, i.info...)

	for _, v := range info {
		description, value := v[0], v[1]
		if value.Value == "" {
			continue
		}

		// Add description.
		cell := table.NewCell()
		cell.SetBackgroundColor(description.BackgroundColor)
		i.setCellBorder(cell, description)

		p := newStyledParagraph(description.TextStyle)
		p.Append(description.Value)
		p.SetMargins(0, 0, 2, 1)
		cell.SetContent(p)

		// Add value.
		cell = table.NewCell()
		cell.SetBackgroundColor(value.BackgroundColor)
		i.setCellBorder(cell, value)

		p = newStyledParagraph(value.TextStyle)
		p.Append(value.Value)
		p.SetMargins(0, 0, 2, 1)
		cell.SetContent(p)
	}

	return table
}

func (i *Invoice) drawTotals() *Table {
	table := newTable(2)

	totals := [][2]*InvoiceCell{i.subtotal}
	totals = append(totals, i.totals...)
	totals = append(totals, i.total)

	for _, total := range totals {
		description, value := total[0], total[1]
		if value.Value == "" {
			continue
		}

		// Add description.
		cell := table.NewCell()
		cell.SetBackgroundColor(description.BackgroundColor)
		cell.SetHorizontalAlignment(value.Alignment)
		i.setCellBorder(cell, description)

		p := newStyledParagraph(description.TextStyle)
		p.SetMargins(0, 0, 2, 1)
		p.Append(description.Value)
		cell.SetContent(p)

		// Add value.
		cell = table.NewCell()
		cell.SetBackgroundColor(value.BackgroundColor)
		cell.SetHorizontalAlignment(value.Alignment)
		i.setCellBorder(cell, description)

		p = newStyledParagraph(value.TextStyle)
		p.SetMargins(0, 0, 2, 1)
		p.Append(value.Value)
		cell.SetContent(p)
	}

	return table
}

func (i *Invoice) generateHeaderBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	// Create title paragraph.
	titleParagraph := newStyledParagraph(i.titleStyle)
	titleParagraph.SetEnableWrap(true)
	titleParagraph.Append(i.title)

	// Add invoice logo.
	table := newTable(2)

	if i.logo != nil {
		cell := table.NewCell()
		cell.SetHorizontalAlignment(CellHorizontalAlignmentLeft)
		cell.SetVerticalAlignment(CellVerticalAlignmentMiddle)
		cell.SetContent(i.logo)

		i.logo.ScaleToHeight(titleParagraph.Height() + 20)
	} else {
		table.SkipCells(1)
	}

	// Add invoice title.
	cell := table.NewCell()
	cell.SetHorizontalAlignment(CellHorizontalAlignmentRight)
	cell.SetVerticalAlignment(CellVerticalAlignmentMiddle)
	cell.SetContent(titleParagraph)

	// Generate blocks.
	return table.GeneratePageBlocks(ctx)
}

func (i *Invoice) generateInformationBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	// Draw addresses.
	separatorParagraph := newStyledParagraph(i.defaultStyle)
	separatorParagraph.SetMargins(0, 0, 0, 20)

	addrParagraphs := i.drawAddress(i.sellerAddress.Name, "", i.sellerAddress)
	addrParagraphs = append(addrParagraphs, separatorParagraph)
	addrParagraphs = append(addrParagraphs,
		i.drawAddress("Bill to", i.buyerAddress.Name, i.buyerAddress)...)

	addrDivision := newDivision()
	for _, addrParagraph := range addrParagraphs {
		addrDivision.Add(addrParagraph)
	}

	// Draw invoice information.
	information := i.drawInformation()

	// Generate blocks.
	table := newTable(2)
	table.SetMargins(0, 0, 25, 0)

	cell := table.NewCell()
	cell.SetContent(addrDivision)

	cell = table.NewCell()
	cell.SetContent(information)

	return table.GeneratePageBlocks(ctx)
}

func (i *Invoice) generateLineBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	table := newTable(len(i.columns))
	table.SetMargins(0, 0, 25, 0)

	// Draw item columns.
	for _, col := range i.columns {
		paragraph := newStyledParagraph(col.TextStyle)
		paragraph.SetMargins(0, 0, 1, 0)
		paragraph.Append(col.Value)

		cell := table.NewCell()
		cell.SetHorizontalAlignment(col.Alignment)
		cell.SetBackgroundColor(col.BackgroundColor)
		i.setCellBorder(cell, col)
		cell.SetContent(paragraph)
	}

	// Draw item lines.
	for _, line := range i.lines {
		for _, itemCell := range line {
			paragraph := newStyledParagraph(itemCell.TextStyle)
			paragraph.SetMargins(0, 0, 3, 2)
			paragraph.Append(itemCell.Value)

			cell := table.NewCell()
			cell.SetHorizontalAlignment(itemCell.Alignment)
			cell.SetBackgroundColor(itemCell.BackgroundColor)
			i.setCellBorder(cell, itemCell)
			cell.SetContent(paragraph)
		}
	}

	return table.GeneratePageBlocks(ctx)
}

func (i *Invoice) generateTotalBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	table := newTable(2)
	table.SetMargins(0, 0, 5, 40)
	table.SkipCells(1)

	totalsTable := i.drawTotals()
	totalsTable.SetMargins(0, 0, 5, 0)

	cell := table.NewCell()
	cell.SetContent(totalsTable)

	return table.GeneratePageBlocks(ctx)
}

func (i *Invoice) generateNoteBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	division := newDivision()

	sections := [][2]string{
		i.notes,
		i.terms,
	}
	sections = append(sections, i.sections...)

	for _, section := range sections {
		if section[1] != "" {
			paragraphs := i.drawSection(section[0], section[1])
			for _, paragraph := range paragraphs {
				division.Add(paragraph)
			}

			sepParagraph := newStyledParagraph(i.defaultStyle)
			sepParagraph.SetMargins(0, 0, 10, 0)
			division.Add(sepParagraph)
		}
	}

	return division.GeneratePageBlocks(ctx)
}

// GeneratePageBlocks generate the Page blocks. Multiple blocks are generated
// if the contents wrap over multiple pages.
func (i *Invoice) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	blockFuncs := []func(ctx DrawContext) ([]*Block, DrawContext, error){
		i.generateHeaderBlocks,
		i.generateInformationBlocks,
		i.generateLineBlocks,
		i.generateTotalBlocks,
		i.generateNoteBlocks,
	}

	var blocks []*Block
	for _, blockFunc := range blockFuncs {
		newBlocks, c, err := blockFunc(ctx)
		if err != nil {
			return blocks, ctx, err
		}

		if len(blocks) == 0 {
			blocks = newBlocks
		} else if len(newBlocks) > 0 {
			blocks[len(blocks)-1].mergeBlocks(newBlocks[0])
			blocks = append(blocks, newBlocks[1:]...)
		}

		ctx = c
	}

	if i.positioning.isRelative() {
		// Move back X to same start of line.
		ctx.X = origCtx.X
	}

	if i.positioning.isAbsolute() {
		// If absolute: return original context.
		return blocks, origCtx, nil
	}

	return blocks, ctx, nil
}
