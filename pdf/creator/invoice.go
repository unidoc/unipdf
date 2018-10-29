/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import "fmt"

// InvoiceAddress
type InvoiceAddress struct {
	Name    string
	Street  string
	Zip     string
	City    string
	Country string
	Phone   string
	Email   string
}

// InvoiceCellProps
type InvoiceCellProps struct {
	Alignment       CellHorizontalAlignment
	BackgroundColor Color
	BorderColor     Color
	Style           TextStyle
}

// InvoiceColumn
type InvoiceColumn struct {
	InvoiceCellProps

	Description string
}

// InvoiceCell
type InvoiceCell struct {
	InvoiceCellProps

	Value string
}

// Invoice represents a configurable template for an invoice.
type Invoice struct {
	// Invoice styles.
	defaultStyle TextStyle
	headingStyle TextStyle
	titleStyle   TextStyle

	infoDescProps InvoiceCellProps
	infoValProps  InvoiceCellProps
	colProps      InvoiceCellProps
	itemProps     InvoiceCellProps
	totalProps    InvoiceCellProps

	// The title of the invoice.
	title string

	// The logo of the invoice.
	logo *Image

	// Buyer address.
	buyerAddress InvoiceAddress

	// Seller address.
	sellerAddress InvoiceAddress

	// Invoice information.
	number       string
	date         string
	paymentTerms string
	dueDate      string

	additionalInfo [][2]string

	// Invoice columns.
	columns []*InvoiceColumn

	// Invoice lines.
	lines [][]*InvoiceCell

	// Invoice totals.
	subtotal *InvoiceCell
	total    *InvoiceCell
	totals   [][2]*InvoiceCell

	// Invoice notes.
	notes string
	terms string

	// Positioning: relative/absolute.
	positioning positioning
}

// newInvoice returns an instance of an empty invoice.
func newInvoice(defaultStyle, headingStyle TextStyle) *Invoice {
	i := &Invoice{
		// Styles.
		defaultStyle: defaultStyle,
		headingStyle: headingStyle,

		// Title.
		title: "INVOICE",

		// Information.
		additionalInfo: [][2]string{},

		// Addresses.
		sellerAddress: InvoiceAddress{},
		buyerAddress:  InvoiceAddress{},
	}

	// Default style properties.
	lightBlue := ColorRGBFrom8bit(217, 240, 250)
	lightGrey := ColorRGBFrom8bit(245, 245, 245)
	mediumGrey := ColorRGBFrom8bit(155, 155, 155)

	i.titleStyle = headingStyle
	i.titleStyle.Color = mediumGrey
	i.titleStyle.FontSize = 20

	i.infoDescProps = i.newCellProps()
	i.infoDescProps.BackgroundColor = lightBlue
	i.infoDescProps.Style = headingStyle

	i.infoValProps = i.newCellProps()
	i.infoValProps.BackgroundColor = lightGrey

	i.colProps = i.newCellProps()
	i.colProps.BackgroundColor = lightBlue
	i.colProps.BorderColor = lightBlue
	i.colProps.Style = headingStyle

	i.itemProps = i.newCellProps()
	i.itemProps.Alignment = CellHorizontalAlignmentRight

	i.totalProps = i.newCellProps()
	i.totalProps.Alignment = CellHorizontalAlignmentRight

	i.subtotal = i.NewCell("")
	i.subtotal.Alignment = CellHorizontalAlignmentRight

	i.total = i.NewCell("")
	i.total.BackgroundColor = lightBlue
	i.total.Style = headingStyle
	i.total.Alignment = CellHorizontalAlignmentRight
	i.total.BorderColor = lightBlue

	// Default item columns.
	quantityCol := i.NewColumn("Quantity")
	quantityCol.Alignment = CellHorizontalAlignmentRight

	unitPriceCol := i.NewColumn("Unit price")
	unitPriceCol.Alignment = CellHorizontalAlignmentRight

	amountCol := i.NewColumn("Amount")
	amountCol.Alignment = CellHorizontalAlignmentRight

	i.columns = []*InvoiceColumn{
		i.NewColumn("Description"),
		quantityCol,
		unitPriceCol,
		amountCol,
	}

	return i
}

func (i *Invoice) newCellProps() InvoiceCellProps {
	white := ColorRGBFrom8bit(255, 255, 255)

	return InvoiceCellProps{
		Alignment:       CellHorizontalAlignmentLeft,
		BackgroundColor: white,
		BorderColor:     white,
		Style:           i.defaultStyle,
	}
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

func (i *Invoice) SetSellerAddress(address InvoiceAddress) {
	i.sellerAddress = address
}

func (i *Invoice) SetBuyerAddress(address InvoiceAddress) {
	i.buyerAddress = address
}

func (i *Invoice) Number() string {
	return i.number
}

func (i *Invoice) SetNumber(number string) {
	i.number = number
}

func (i *Invoice) Date() string {
	return i.date
}

func (i *Invoice) SetDate(date string) {
	i.date = date
}

func (i *Invoice) PaymentTerms() string {
	return i.paymentTerms
}

func (i *Invoice) SetPaymentTerms(paymentTerms string) {
	i.paymentTerms = paymentTerms
}

func (i *Invoice) DueDate() string {
	return i.dueDate
}

func (i *Invoice) SetDueDate(dueDate string) {
	i.dueDate = dueDate
}

func (i *Invoice) AddInvoiceInfo(description, value string) {
	i.additionalInfo = append(i.additionalInfo, [2]string{description, value})
}

func (i *Invoice) AppendColumn(description string) *InvoiceColumn {
	return nil
}

func (i *Invoice) InsertColumn(description string) *InvoiceColumn {
	return nil
}

func (i *Invoice) Lines() [][]*InvoiceCell {
	return i.lines
}

func (i *Invoice) AddLine(values ...string) {
	lenCols := len(i.columns)

	var line []*InvoiceCell
	for j, value := range values {
		itemCell := i.NewCell(value)
		if j < lenCols {
			itemCell.Alignment = i.columns[j].Alignment
		}

		line = append(line, itemCell)
	}

	i.lines = append(i.lines, line)
}

func (i *Invoice) Subtotal() *InvoiceCell {
	return i.subtotal
}

func (i *Invoice) SetSubtotal(value string) *InvoiceCell {
	i.subtotal.Value = value
	return i.subtotal
}

func (i *Invoice) Total() *InvoiceCell {
	return i.total
}

func (i *Invoice) SetTotal(value string) *InvoiceCell {
	i.total.Value = value
	return i.total
}

func (i *Invoice) TotalLines() [][2]*InvoiceCell {
	return i.totals
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

func (i *Invoice) Notes() string {
	return i.notes
}

func (i *Invoice) SetNotes(notes string) {
	i.notes = notes
}

func (i *Invoice) Terms() string {
	return i.terms
}

func (i *Invoice) SetTerms(terms string) {
	i.terms = terms
}

func (i *Invoice) NewCell(value string) *InvoiceCell {
	return &InvoiceCell{
		i.itemProps,
		value,
	}
}

func (i *Invoice) NewColumn(description string) *InvoiceColumn {
	return &InvoiceColumn{
		i.colProps,
		description,
	}
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

	// Contact information
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

	info := [][2]string{
		[2]string{"Invoice number", i.number},
		[2]string{"Invoice date", i.date},
		[2]string{"Payment terms", i.paymentTerms},
		[2]string{"Due date", i.dueDate},
	}
	info = append(info, i.additionalInfo...)

	for _, v := range info {
		description, value := v[0], v[1]
		if len(value) == 0 {
			continue
		}

		// Add description.
		cell := table.NewCell()
		cell.SetBackgroundColor(i.infoDescProps.BackgroundColor)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetBorderColor(i.infoDescProps.BorderColor)

		p := newStyledParagraph(i.infoDescProps.Style)
		p.Append(description)
		p.SetMargins(0, 0, 2, 0)
		cell.SetContent(p)

		// Add value.
		cell = table.NewCell()
		cell.SetBackgroundColor(i.infoValProps.BackgroundColor)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetBorderColor(i.infoValProps.BorderColor)

		p = newStyledParagraph(i.infoValProps.Style)
		p.Append(value)
		p.SetMargins(0, 0, 2, 0)
		cell.SetContent(p)
	}

	return table
}

func (i *Invoice) drawTotals() *Table {
	table := newTable(2)

	totals := [][2]*InvoiceCell{}
	if i.subtotal.Value != "" {
		subtotalDesc := *i.subtotal
		subtotalDesc.Value = "Subtotal"

		totals = append(totals, [2]*InvoiceCell{&subtotalDesc, i.subtotal})
	}

	totals = append(totals, i.totals...)

	if i.total.Value != "" {
		totalDesc := *i.total
		totalDesc.Value = "Total"

		totals = append(totals, [2]*InvoiceCell{&totalDesc, i.total})
	}

	for _, total := range totals {
		description, value := total[0], total[1]

		// Add description.
		cell := table.NewCell()
		cell.SetBackgroundColor(description.BackgroundColor)
		cell.SetHorizontalAlignment(value.Alignment)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetBorderColor(value.BorderColor)

		p := newStyledParagraph(description.Style)
		p.SetMargins(0, 0, 1, 1)
		p.Append(description.Value)
		cell.SetContent(p)

		// Add value.
		cell = table.NewCell()
		cell.SetBackgroundColor(value.BackgroundColor)
		cell.SetHorizontalAlignment(value.Alignment)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetBorderColor(value.BorderColor)

		p = newStyledParagraph(value.Style)
		p.SetMargins(0, 0, 1, 1)
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

	addrParagraphs := i.drawAddress(i.sellerAddress.Name, "", &i.sellerAddress)
	addrParagraphs = append(addrParagraphs, separatorParagraph)
	addrParagraphs = append(addrParagraphs,
		i.drawAddress("Bill to", i.buyerAddress.Name, &i.buyerAddress)...)

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
	table := newTable(4)
	table.SetMargins(0, 0, 25, 0)

	// Draw item columns
	for _, col := range i.columns {
		paragraph := newStyledParagraph(col.Style)
		paragraph.SetMargins(0, 0, 1, 0)
		paragraph.Append(col.Description)

		cell := table.NewCell()
		cell.SetHorizontalAlignment(col.Alignment)
		cell.SetBackgroundColor(col.BackgroundColor)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
		cell.SetBorderColor(col.BorderColor)
		cell.SetContent(paragraph)
	}

	// Draw item lines
	for _, line := range i.lines {
		for _, itemCell := range line {
			paragraph := newStyledParagraph(itemCell.Style)
			paragraph.SetMargins(0, 0, 1, 0)
			paragraph.Append(itemCell.Value)

			cell := table.NewCell()
			cell.SetHorizontalAlignment(itemCell.Alignment)
			cell.SetBackgroundColor(itemCell.BackgroundColor)
			cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
			cell.SetBorderColor(itemCell.BorderColor)
			cell.SetContent(paragraph)
		}
	}

	return table.GeneratePageBlocks(ctx)
}

func (i *Invoice) generateTotalBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	table := newTable(2)
	table.SetMargins(0, 0, 5, 35)

	mediumGrey := ColorRGBFrom8bit(195, 195, 195)

	cell := table.NewCell()
	cell.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 1)
	cell.SetBorderColor(mediumGrey)

	if i.notes != "" {
		noteParagraphs := i.drawSection("Notes", i.notes)

		noteDivision := newDivision()
		for _, noteParagraph := range noteParagraphs {
			noteParagraph.SetMargins(0, 0, 5, 0)
			noteDivision.Add(noteParagraph)
		}

		cell.SetContent(noteDivision)
	}

	totalsTable := i.drawTotals()
	totalsTable.SetMargins(0, 0, 5, 0)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideTop, CellBorderStyleSingle, 1)
	cell.SetBorderColor(mediumGrey)
	cell.SetContent(totalsTable)

	return table.GeneratePageBlocks(ctx)
}

func (i *Invoice) generateNoteBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	division := newDivision()

	if i.terms != "" {
		termParagraphs := i.drawSection("Terms and conditions", i.terms)
		for _, termParagraph := range termParagraphs {
			termParagraph.SetMargins(0, 0, 5, 0)
			division.Add(termParagraph)
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
