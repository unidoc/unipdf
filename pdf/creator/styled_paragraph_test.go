/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func TestParagraphRegularVsStyled(t *testing.T) {
	fontRegular := fonts.NewFontHelvetica()
	fontBold := fonts.NewFontHelveticaBold()

	c := New()
	c.NewPage()

	// Draw section title.
	p := NewParagraph("Regular paragraph vs styled paragraph (should be identical)")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	table := NewTable(2)
	table.SetColumnWidths(0.5, 0.5)

	// Add regular paragraph to table.
	p = NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetMargins(10, 10, 5, 10)
	p.SetFont(fontBold)
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentLeft)

	cell := table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(p)

	// Add styled paragraph to table.
	style := NewTextStyle()
	style.Font = fontBold

	s := NewStyledParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.", style)
	s.SetMargins(10, 10, 5, 10)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentLeft)

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(s)

	// Add regular paragraph to table.
	p = NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetMargins(10, 10, 5, 10)
	p.SetFont(fontRegular)
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentJustify)
	p.SetColor(ColorRGBFrom8bit(0, 0, 255))

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(p)

	// Add styled paragraph to table.
	style.Font = fontRegular
	style.Color = ColorRGBFrom8bit(0, 0, 255)

	s = NewStyledParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.", style)
	s.SetMargins(10, 10, 5, 10)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentJustify)

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(s)

	// Add regular paragraph to table.
	p = NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetMargins(10, 10, 5, 10)
	p.SetFont(fontRegular)
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentRight)

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(p)

	// Add styled paragraph to table.
	style.Font = fontRegular
	style.Color = ColorRGBFrom8bit(0, 0, 0)

	s = NewStyledParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.", style)
	s.SetMargins(10, 10, 5, 10)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentRight)

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(s)

	// Add regular paragraph to table.
	p = NewParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.SetMargins(10, 10, 5, 10)
	p.SetFont(fontBold)
	p.SetEnableWrap(true)
	p.SetTextAlignment(TextAlignmentCenter)

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(p)

	// Add styled paragraph to table.
	style.Font = fontBold
	style.Color = ColorRGBFrom8bit(0, 0, 0)

	s = NewStyledParagraph("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.", style)
	s.SetMargins(10, 10, 5, 10)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentCenter)

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(s)

	// Draw table.
	err = c.Draw(table)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/paragraphs_regular_vs_styled.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestStyledParagraph(t *testing.T) {
	fontRegular := fonts.NewFontCourier()
	fontBold := fonts.NewFontCourierBold()
	fontHelvetica := fonts.NewFontHelvetica()

	c := New()
	c.NewPage()

	// Draw section title.
	p := NewParagraph("Styled paragraph")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	style := NewTextStyle()
	style.Font = fontRegular

	s := NewStyledParagraph("This is a paragraph ", style)
	s.SetEnableWrap(true)
	s.SetTextAlignment(TextAlignmentJustify)
	s.SetMargins(0, 0, 10, 0)

	style.Color = ColorRGBFrom8bit(255, 0, 0)
	s.Append("with different colors ", style)

	style.Color = ColorRGBFrom8bit(0, 0, 0)
	style.FontSize = 14
	s.Append("and with different font sizes ", style)

	style.FontSize = 10
	style.Font = fontBold
	s.Append("and with different font styles ", style)

	style.Font = fontHelvetica
	style.FontSize = 13
	s.Append("and with different fonts ", style)

	style.Font = fontBold
	style.Color = ColorRGBFrom8bit(0, 0, 255)
	style.FontSize = 15
	s.Append("and with the changed properties all at once. ", style)

	style.Color = ColorRGBFrom8bit(127, 255, 0)
	style.FontSize = 12
	style.Font = fontHelvetica
	s.Append("And maybe try a different color again.", style)

	err = c.Draw(s)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Test the reset function and also pagination
	style.Color = ColorRGBFrom8bit(255, 0, 0)
	style.Font = fontRegular

	s.Reset("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar. ", style)
	s.SetTextAlignment(TextAlignmentJustify)

	style.Color = ColorRGBFrom8bit(0, 0, 255)
	style.FontSize = 15
	style.Font = fontHelvetica
	s.Append("And maybe try a different color again.", style)

	err = c.Draw(s)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/styled_paragraph.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
