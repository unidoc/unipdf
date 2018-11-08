/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/unidoc/unidoc/pdf/model"
)

func TestParagraphRegularVsStyled(t *testing.T) {
	fontRegular := newStandard14Font(t, model.Helvetica)
	fontBold := newStandard14Font(t, model.HelveticaBold)

	c := New()
	c.NewPage()

	// Draw section title.
	p := c.NewParagraph("Regular paragraph vs styled paragraph (should be identical)")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	table := c.NewTable(2)
	table.SetColumnWidths(0.5, 0.5)

	// Add regular paragraph to table.
	p = c.NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetMargins(10, 10, 5, 10)
	p.SetFont(fontBold)
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentLeft)

	cell := table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	// Add styled paragraph to table.
	style := c.NewTextStyle()
	style.Font = fontBold

	s := c.NewStyledParagraph()
	s.SetMargins(10, 10, 5, 10)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentLeft)

	chunk := s.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(s)

	// Add regular paragraph to table.
	p = c.NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetMargins(10, 10, 5, 10)
	p.SetFont(fontRegular)
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentJustify)
	p.SetColor(ColorRGBFrom8bit(0, 0, 255))

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	// Add styled paragraph to table.
	style.Font = fontRegular
	style.Color = ColorRGBFrom8bit(0, 0, 255)

	s = c.NewStyledParagraph()
	s.SetMargins(10, 10, 5, 10)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentJustify)

	chunk = s.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(s)

	// Add regular paragraph to table.
	p = c.NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetMargins(10, 10, 5, 10)
	p.SetFont(fontRegular)
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentRight)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	// Add styled paragraph to table.
	style.Font = fontRegular
	style.Color = ColorRGBFrom8bit(0, 0, 0)

	s = c.NewStyledParagraph()
	s.SetMargins(10, 10, 5, 10)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentRight)

	chunk = s.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(s)

	// Add regular paragraph to table.
	p = c.NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetMargins(10, 10, 5, 10)
	p.SetFont(fontBold)
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentCenter)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(p)

	// Add styled paragraph to table.
	style.Font = fontBold
	style.Color = ColorRGBFrom8bit(0, 0, 0)

	s = c.NewStyledParagraph()
	s.SetMargins(10, 10, 5, 10)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentCenter)

	chunk = s.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetContent(s)

	// Test table cell alignment.
	style = c.NewTextStyle()

	// Test left alignment with paragraph wrapping enabled.
	p = c.NewParagraph("Wrap enabled. This text should be left aligned.")
	p.SetEnableWrap(true)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentLeft)
	cell.SetContent(p)

	s = c.NewStyledParagraph()
	s.SetEnableWrap(true)

	chunk = s.Append("Wrap enabled. This text should be left aligned.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentLeft)
	cell.SetContent(s)

	// Test left alignment with paragraph wrapping disabled.
	p = c.NewParagraph("Wrap disabled. This text should be left aligned.")
	p.SetEnableWrap(false)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentLeft)
	cell.SetContent(p)

	s = c.NewStyledParagraph()
	s.SetEnableWrap(false)

	chunk = s.Append("Wrap disabled. This text should be left aligned.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentLeft)
	cell.SetContent(s)

	// Test center alignment with paragraph wrapping enabled.
	p = c.NewParagraph("Wrap enabled. This text should be center aligned.")
	p.SetEnableWrap(true)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentCenter)
	cell.SetContent(p)

	s = c.NewStyledParagraph()
	s.SetEnableWrap(true)

	chunk = s.Append("Wrap enabled. This text should be center aligned.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentCenter)
	cell.SetContent(s)

	// Test center alignment with paragraph wrapping disabled.
	p = c.NewParagraph("Wrap disabled. This text should be center aligned.")
	p.SetEnableWrap(false)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentCenter)
	cell.SetContent(p)

	s = c.NewStyledParagraph()
	s.SetEnableWrap(false)

	chunk = s.Append("Wrap disabled. This text should be center aligned.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentCenter)
	cell.SetContent(s)

	// Test right alignment with paragraph wrapping enabled.
	p = c.NewParagraph("Wrap enabled. This text should be right aligned.")
	p.SetEnableWrap(true)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentRight)
	cell.SetContent(p)

	s = c.NewStyledParagraph()
	s.SetEnableWrap(true)

	chunk = s.Append("Wrap enabled. This text should be right aligned.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentRight)
	cell.SetContent(s)

	// Test right alignment with paragraph wrapping disabled.
	p = c.NewParagraph("Wrap disabled. This text should be right aligned.")
	p.SetEnableWrap(false)

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentRight)
	cell.SetContent(p)

	s = c.NewStyledParagraph()
	s.SetEnableWrap(false)

	chunk = s.Append("Wrap disabled. This text should be right aligned.")
	chunk.Style = style

	cell = table.NewCell()
	cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)
	cell.SetHorizontalAlignment(CellHorizontalAlignmentRight)
	cell.SetContent(s)

	// Draw table.
	err = c.Draw(table)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("paragraphs_regular_vs_styled.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestStyledParagraph(t *testing.T) {
	fontRegular := newStandard14Font(t, model.Courier)
	fontBold := newStandard14Font(t, model.CourierBold)
	fontHelvetica := newStandard14Font(t, model.Helvetica)

	c := New()
	c.NewPage()

	// Draw section title.
	p := c.NewParagraph("Styled paragraph")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	style := c.NewTextStyle()
	style.Font = fontRegular

	s := c.NewStyledParagraph()
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentJustify)
	s.SetMargins(0, 0, 10, 0)

	chunk := s.Append("This is a paragraph ")
	chunk.Style = style

	style.Color = ColorRGBFrom8bit(255, 0, 0)
	chunk = s.Append("with different colors ")
	chunk.Style = style

	style.Color = ColorRGBFrom8bit(0, 0, 0)
	style.FontSize = 14
	chunk = s.Append("and with different font sizes ")
	chunk.Style = style

	style.FontSize = 10
	style.Font = fontBold
	chunk = s.Append("and with different font styles ")
	chunk.Style = style

	style.Font = fontHelvetica
	style.FontSize = 13
	chunk = s.Append("and with different fonts ")
	chunk.Style = style

	style.Font = fontBold
	style.Color = ColorRGBFrom8bit(0, 0, 255)
	style.FontSize = 15
	chunk = s.Append("and with the changed properties all at once. ")
	chunk.Style = style

	style.Color = ColorRGBFrom8bit(127, 255, 0)
	style.FontSize = 12
	style.Font = fontHelvetica
	chunk = s.Append("And maybe try a different color again.")
	chunk.Style = style

	err = c.Draw(s)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Test the reset function and also pagination
	style.Color = ColorRGBFrom8bit(255, 0, 0)
	style.Font = fontRegular

	s.Reset()
	s.SetTextAlignment(TextAlignmentJustify)

	chunk = s.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar. ")
	chunk.Style = style

	style.Color = ColorRGBFrom8bit(0, 0, 255)
	style.FontSize = 15
	style.Font = fontHelvetica
	chunk = s.Append("And maybe try a different color again.")
	chunk.Style = style

	err = c.Draw(s)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("styled_paragraph.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
