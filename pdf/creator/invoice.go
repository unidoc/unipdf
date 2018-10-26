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
	totals [][2]*InvoiceCell

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
	i.colProps.Style = headingStyle

	i.itemProps = i.newCellProps()

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

func (i *Invoice) AddSubtotal(value string) *InvoiceCell {
	return nil
}

func (i *Invoice) AddTotal(value string) *InvoiceCell {
	return nil
}

func (i *Invoice) AddTotalsLine(desc, value string) *InvoiceCell {
	return nil
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

// GeneratePageBlocks generate the Page blocks. Multiple blocks are generated
// if the contents wrap over multiple pages.
func (i *Invoice) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	origCtx := ctx

	// Generate title blocks.
	blocks, ctx, err := i.generateHeaderBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}

	// Generate address and information blocks.
	newBlocks, c, err := i.generateInformationBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}

	blocks[len(blocks)-1].mergeBlocks(newBlocks[0])
	blocks = append(blocks, newBlocks[1:]...)
	ctx = c

	// Generate line items blocks.
	newBlocks, c, err = i.generateLineBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}

	blocks[len(blocks)-1].mergeBlocks(newBlocks[0])
	blocks = append(blocks, newBlocks[1:]...)
	ctx = c

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
