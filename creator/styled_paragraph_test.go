/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

func TestParagraphRegularVsStyled(t *testing.T) {
	fontRegular := newStandard14Font(t, model.HelveticaName)
	fontBold := newStandard14Font(t, model.HelveticaBoldName)

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
	fontRegular := newStandard14Font(t, model.CourierName)
	fontBold := newStandard14Font(t, model.CourierBoldName)
	fontHelvetica := newStandard14Font(t, model.HelveticaName)

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

func TestStyledParagraphLinks(t *testing.T) {
	c := New()

	// First page.
	c.NewPage()

	p := c.NewStyledParagraph()
	p.Append("Paragraph links are useful for going to remote places like ")
	p.AddExternalLink("Google", "https://google.com")
	p.Append(", or maybe ")
	p.AddExternalLink("Github", "https://github.com")
	p.Append("\nHowever, you can also use them to move go to the ")
	p.AddInternalLink("start", 2, 0, 0, 0).Style.Color = ColorRGBFrom8bit(255, 0, 0)
	p.Append(" of the second page, the ")
	p.AddInternalLink("middle", 2, 0, 250, 0).Style.Color = ColorRGBFrom8bit(0, 255, 0)
	p.Append(" or the ")
	p.AddInternalLink("end", 2, 0, 500, 0).Style.Color = ColorRGBFrom8bit(0, 0, 255)
	p.Append(" of the second page.\nOr maybe go to the third ")
	p.AddInternalLink("page", 3, 0, 0, 0)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Second page.
	c.NewPage()

	p = c.NewStyledParagraph()

	p.Append(`Lorem ipsum dolor sit amet, consectetur adipiscing elit. In maximus id purus vitae faucibus. Proin at egestas ex. Mauris id luctus nulla, et scelerisque odio. Praesent scelerisque a erat non ullamcorper. Donec at est nec nunc tempor bibendum at eget quam. Aliquam bibendum est vel ultrices condimentum. Sed augue sapien, commodo et ligula a, consequat consectetur diam. Donec in justo dui. Proin quis aliquam magna. Fusce vel enim ut leo sagittis vehicula vel sed magna. Curabitur lacinia condimentum laoreet. Maecenas venenatis, sapien a hendrerit viverra, arcu odio blandit nulla, a varius sem nisl in magna. Fusce aliquam nec urna nec congue. Phasellus metus quam, hendrerit ac laoreet non, bibendum quis augue. Donec quam ex, aliquam sed rutrum a, lobortis at turpis. Pellentesque pellentesque vitae augue at faucibus.
Nullam porttitor scelerisque mauris. Aenean nunc nunc, facilisis ut arcu eget, dignissim euismod justo. Curabitur lobortis ut augue sit amet pellentesque. Donec interdum lobortis quam, eget lacinia nunc sagittis sed. Nunc tristique consectetur convallis. Fusce tincidunt consequat tincidunt. Phasellus a faucibus metus. Vestibulum eu facilisis sem. Quisque vulputate eros in quam vulputate, id faucibus nibh aliquet. Etiam dictum laoreet urna, sed ultricies nulla volutpat vel. In volutpat nisl nisl, eu suscipit risus feugiat eu. Duis egestas, ante quis pellentesque pulvinar, purus urna imperdiet metus, nec commodo libero sem in dolor.
Phasellus semper, ipsum sollicitudin iaculis dapibus, justo leo interdum ipsum, id feugiat enim lacus id nisl. Sed cursus lacinia laoreet. Cras cursus risus ex, id dapibus mauris lacinia et. Nam eget metus nec ex iaculis laoreet a eu mauris. Vivamus porta ut lacus nec suscipit. Vivamus eu elit in ante consectetur condimentum. Vivamus iaculis tristique lacus, id iaculis arcu maximus pellentesque. Duis commodo nisi turpis, et gravida libero mattis id. Nullam pretium arcu metus, at auctor neque tincidunt vel. Morbi sagittis massa sed arcu dictum, eget ornare nisi ullamcorper. Sed ac lacus ex. Aliquam ornare vehicula interdum. Nulla vehicula est vel turpis ullamcorper iaculis.
Suspendisse potenti. Aenean pellentesque eros nulla, sed tempor tellus hendrerit tristique. Etiam nec enim et ligula sollicitudin faucibus ut eget libero. Suspendisse eget blandit lacus. Suspendisse consequat orci risus. Curabitur id libero quam. Ut pellentesque tristique porta. Phasellus leo augue, porttitor id suscipit eleifend, elementum ut diam. Ut non ipsum in orci consectetur posuere. Nulla facilisi. In laoreet, nunc fringilla feugiat dapibus, augue diam cursus felis, eu efficitur dui ipsum vestibulum orci. Maecenas leo leo, sagittis pharetra venenatis at, porttitor ut risus.
Nunc euismod facilisis venenatis. Donec diam enim, sollicitudin ac vestibulum ultrices, malesuada eget ipsum. Morbi et sem vel metus convallis scelerisque. Vivamus justo felis, ullamcorper nec arcu eget, pharetra fringilla diam. Praesent ut mauris leo. Quisque sollicitudin sodales justo vel ornare. Proin sollicitudin suscipit risus, vel aliquam nisl ultrices a. Nulla facilisi. Sed eget facilisis dui. Duis maximus tortor eget massa varius sollicitudin. Cras interdum ornare nulla, pulvinar sagittis elit gravida non. Nulla consequat arcu gravida ante commodo, non tempus turpis porta. Quisque tincidunt quam et nisl maximus, nec hendrerit libero feugiat. Sed vel vestibulum leo. Mauris quis efficitur ligula, quis facilisis nibh. Suspendisse commodo elit id vehicula viverra.
Donec auctor tempor ante vel eleifend. Cras laoreet in lacus sit amet tristique. Donec porta, mi non dignissim consectetur, magna urna gravida lectus, in mattis nisi felis id odio. Sed sem ligula, feugiat et lectus tincidunt, condimentum sollicitudin dolor. Donec pulvinar, nibh ultricies tristique aliquam, lorem massa laoreet purus, eu ultrices lorem turpis eget erat. Maecenas auctor tempus dignissim. Pellentesque ut consequat magna. Vestibulum ante velit, feugiat id lectus pellentesque, congue consectetur sapien. Quisque mattis, nisi et facilisis pulvinar, nulla tellus dignissim tortor, eget convallis sem dolor vel lorem. Pellentesque pharetra tortor odio, et egestas elit scelerisque et. Maecenas suscipit lorem ut purus porta dictum.
Donec placerat finibus leo, quis aliquet ipsum dignissim sed. Duis semper vulputate rutrum. Suspendisse egestas, magna in lacinia vulputate, massa sapien lobortis quam, consequat interdum tellus enim ac nulla. Praesent non risus ut nulla tincidunt posuere quis vitae dui. Cras sem lectus, efficitur eget nunc et, gravida accumsan massa. Nam egestas laoreet nunc, eu posuere dolor iaculis sed. In fringilla cursus lectus sit amet ullamcorper.
Curabitur iaculis elit id neque sollicitudin, non dapibus nisi bibendum. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Quisque nec dui tempor, convallis felis in, fermentum urna. In ut egestas lacus, quis mollis turpis. Vestibulum finibus metus vel turpis maximus pharetra. Duis tempus aliquam leo eu feugiat. In tincidunt lectus dolor. Mauris id tristique enim, vitae pellentesque elit. Nam mattis vestibulum molestie. Quisque aliquam lacus vel porttitor euismod. Donec rhoncus erat orci. Curabitur venenatis augue vitae metus facilisis, bibendum bibendum ligula elementum.
Sed imperdiet sodales lacus sed sollicitudin. In porta tortor quis augue tempor, eget laoreet tortor tempor. Phasellus in elit et risus interdum tincidunt a ut ante. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Duis volutpat nisl id molestie finibus. Suspendisse et nunc aliquet, elementum metus et, bibendum dui. Cras aliquam nunc est, et sagittis nibh tristique sed. Phasellus porta lectus vel sapien elementum, in finibus orci sodales. Mauris orci felis, porta et dapibus eu, dignissim sed tortor. Nullam faucibus sit amet magna ut pellentesque. Etiam non purus non lacus auctor faucibus.`)
	p.Append("\n\nYou can also go back to ").Style.FontSize = 14
	p.AddInternalLink("page 1", 1, 0, 0, 0).Style.FontSize = 14

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Third page.
	c.NewPage()

	p = c.NewStyledParagraph()
	p.Append("This is the third page.\nGo to ")
	p.AddInternalLink("page 1", 1, 0, 0, 0)
	p.Append("\nGo to ")
	p.AddInternalLink("page 2", 2, 0, 0, 0)

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("styled_paragraph_links.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestStyledParagraphRenderingModes(t *testing.T) {
	fontRegular := newStandard14Font(t, model.HelveticaName)

	c := New()
	c.NewPage()

	// Showcase rendering modes.
	tmodes := []TextRenderingMode{
		TextRenderingModeFill,
		TextRenderingModeStroke,
		TextRenderingModeFillStroke,
		TextRenderingModeInvisible,
		TextRenderingModeFillClip,
		TextRenderingModeStrokeClip,
		TextRenderingModeFillStrokeClip,
		TextRenderingModeClip,
	}

	tmodesDesc := []string{
		"- Text rendering mode fill: ",
		"- Text rendering mode stroke: ",
		"- Text rendering mode fill and stroke: ",
		"- Text rendering mode invisible: ",
		"- Text rendering mode fill and clip: ",
		"- Text rendering mode stroke and clip: ",
		"- Text rendering mode fill, stroke and clip: ",
		"- Text rendering mode clip: ",
	}

	p := c.NewStyledParagraph()
	p.SetLineHeight(1.5)

	for i, tmode := range tmodes {
		chunk := p.Append(tmodesDesc[i])
		chunk.Style.Font = fontRegular
		chunk.Style.FontSize = 12

		chunk = p.Append("This is some sample text\n")
		chunk.Style.RenderingMode = tmode
		chunk.Style.Font = fontRegular
		chunk.Style.FontSize = 22
	}

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Invisible, manually positioned paragraph.
	p = c.NewStyledParagraph()
	p.SetPos(150, 500)
	chunk := p.Append("Invisible text >>> ")
	chunk.Style.Font = fontRegular
	chunk.Style.FontSize = 12

	chunk = p.Append("Some invisible text manually positioned")
	chunk.Style.Font = fontRegular
	chunk.Style.FontSize = 15
	chunk.Style.RenderingMode = TextRenderingModeInvisible

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("styled_paragraph_rendering_mode.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestStyledParagraphCharacterSpacing(t *testing.T) {
	fontRegular := newStandard14Font(t, model.HelveticaName)
	fontBold := newStandard14Font(t, model.HelveticaBoldName)

	c := New()
	c.NewPage()

	// Draw section title.
	s := c.NewStyledParagraph()
	chunk := s.Append("Text align: left (regular vs character spacing)")
	chunk.Style.Font = fontBold

	err := c.Draw(s)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Text align left - regular.
	p := c.NewStyledParagraph()
	p.SetMargins(0, 0, 20, 0)
	p.SetTextAlignment(TextAlignmentLeft)

	chunk = p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Sed enim ut sem viverra aliquet eget sit amet. Sed viverra ipsum nunc aliquet bibendum enim facilisis. Consectetur libero id faucibus nisl. Cras adipiscing enim eu turpis egestas pretium aenean pharetra. Viverra suspendisse potenti nullam ac tortor vitae purus. Ultrices in iaculis nunc sed augue lacus viverra. Sed pulvinar proin gravida hendrerit lectus. Tempus quam pellentesque nec nam aliquam sem et tortor consequat. Magna ac placerat vestibulum lectus mauris ultrices eros in. Non diam phasellus vestibulum lorem. Lorem donec massa sapien faucibus et molestie ac. Est placerat in egestas erat. Massa id neque aliquam vestibulum morbi blandit cursus.")
	chunk.Style.Font = fontRegular

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Text align left - character spacing.
	p = c.NewStyledParagraph()
	p.SetMargins(0, 0, 20, 50)
	p.SetTextAlignment(TextAlignmentLeft)

	chunk = p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Sed enim ut sem viverra aliquet eget sit amet. Sed viverra ipsum nunc aliquet bibendum enim facilisis. Consectetur libero id faucibus nisl. Cras adipiscing enim eu turpis egestas pretium aenean pharetra. Viverra suspendisse potenti nullam ac tortor vitae purus. Ultrices in iaculis nunc sed augue lacus viverra. Sed pulvinar proin gravida hendrerit lectus. Tempus quam pellentesque nec nam aliquam sem et tortor consequat. Magna ac placerat vestibulum lectus mauris ultrices eros in. Non diam phasellus vestibulum lorem. Lorem donec massa sapien faucibus et molestie ac. Est placerat in egestas erat. Massa id neque aliquam vestibulum morbi blandit cursus.")
	chunk.Style.Font = fontRegular
	chunk.Style.CharSpacing = 0.75

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Draw section title.
	s = c.NewStyledParagraph()
	chunk = s.Append("Text align: center (regular vs character spacing)")
	chunk.Style.Font = fontBold

	err = c.Draw(s)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Text alignment center - regular.
	p = c.NewStyledParagraph()
	p.SetMargins(0, 0, 20, 0)
	p.SetTextAlignment(TextAlignmentCenter)

	chunk = p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Sed enim ut sem viverra aliquet eget sit amet. Sed viverra ipsum nunc aliquet bibendum enim facilisis. Consectetur libero id faucibus nisl. Cras adipiscing enim eu turpis egestas pretium aenean pharetra. Viverra suspendisse potenti nullam ac tortor vitae purus. Ultrices in iaculis nunc sed augue lacus viverra. Sed pulvinar proin gravida hendrerit lectus. Tempus quam pellentesque nec nam aliquam sem et tortor consequat. Magna ac placerat vestibulum lectus mauris ultrices eros in. Non diam phasellus vestibulum lorem. Lorem donec massa sapien faucibus et molestie ac. Est placerat in egestas erat. Massa id neque aliquam vestibulum morbi blandit cursus.")
	chunk.Style.Font = fontRegular

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Text alignment center - character spacing.
	p = c.NewStyledParagraph()
	p.SetMargins(0, 0, 20, 50)
	p.SetTextAlignment(TextAlignmentCenter)

	chunk = p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Sed enim ut sem viverra aliquet eget sit amet. Sed viverra ipsum nunc aliquet bibendum enim facilisis. Consectetur libero id faucibus nisl. Cras adipiscing enim eu turpis egestas pretium aenean pharetra. Viverra suspendisse potenti nullam ac tortor vitae purus. Ultrices in iaculis nunc sed augue lacus viverra. Sed pulvinar proin gravida hendrerit lectus. Tempus quam pellentesque nec nam aliquam sem et tortor consequat. Magna ac placerat vestibulum lectus mauris ultrices eros in. Non diam phasellus vestibulum lorem. Lorem donec massa sapien faucibus et molestie ac. Est placerat in egestas erat. Massa id neque aliquam vestibulum morbi blandit cursus.")
	chunk.Style.Font = fontRegular
	chunk.Style.CharSpacing = 1

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	c.NewPage()

	// Draw section title.
	s = c.NewStyledParagraph()
	chunk = s.Append("Text align: right (regular vs character spacing)")
	chunk.Style.Font = fontBold

	err = c.Draw(s)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Text alignment right - regular.
	p = c.NewStyledParagraph()
	p.SetMargins(0, 0, 20, 0)
	p.SetTextAlignment(TextAlignmentRight)

	chunk = p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Sed enim ut sem viverra aliquet eget sit amet. Sed viverra ipsum nunc aliquet bibendum enim facilisis. Consectetur libero id faucibus nisl. Cras adipiscing enim eu turpis egestas pretium aenean pharetra. Viverra suspendisse potenti nullam ac tortor vitae purus. Ultrices in iaculis nunc sed augue lacus viverra. Sed pulvinar proin gravida hendrerit lectus. Tempus quam pellentesque nec nam aliquam sem et tortor consequat. Magna ac placerat vestibulum lectus mauris ultrices eros in. Non diam phasellus vestibulum lorem. Lorem donec massa sapien faucibus et molestie ac. Est placerat in egestas erat. Massa id neque aliquam vestibulum morbi blandit cursus.")
	chunk.Style.Font = fontRegular

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Text alignment right - character spacing.
	p = c.NewStyledParagraph()
	p.SetMargins(0, 0, 20, 50)
	p.SetTextAlignment(TextAlignmentRight)

	chunk = p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Sed enim ut sem viverra aliquet eget sit amet. Sed viverra ipsum nunc aliquet bibendum enim facilisis. Consectetur libero id faucibus nisl. Cras adipiscing enim eu turpis egestas pretium aenean pharetra. Viverra suspendisse potenti nullam ac tortor vitae purus. Ultrices in iaculis nunc sed augue lacus viverra. Sed pulvinar proin gravida hendrerit lectus. Tempus quam pellentesque nec nam aliquam sem et tortor consequat. Magna ac placerat vestibulum lectus mauris ultrices eros in. Non diam phasellus vestibulum lorem. Lorem donec massa sapien faucibus et molestie ac. Est placerat in egestas erat. Massa id neque aliquam vestibulum morbi blandit cursus.")
	chunk.Style.Font = fontRegular
	chunk.Style.CharSpacing = 1

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Draw section title.
	s = c.NewStyledParagraph()
	chunk = s.Append("Text align: justify (regular vs character spacing)")
	chunk.Style.Font = fontBold

	err = c.Draw(s)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Text alignment justify - regular.
	p = c.NewStyledParagraph()
	p.SetMargins(0, 0, 20, 0)
	p.SetTextAlignment(TextAlignmentJustify)

	chunk = p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Sed enim ut sem viverra aliquet eget sit amet. Sed viverra ipsum nunc aliquet bibendum enim facilisis. Consectetur libero id faucibus nisl. Cras adipiscing enim eu turpis egestas pretium aenean pharetra. Viverra suspendisse potenti nullam ac tortor vitae purus. Ultrices in iaculis nunc sed augue lacus viverra. Sed pulvinar proin gravida hendrerit lectus. Tempus quam pellentesque nec nam aliquam sem et tortor consequat. Magna ac placerat vestibulum lectus mauris ultrices eros in. Non diam phasellus vestibulum lorem. Lorem donec massa sapien faucibus et molestie ac. Est placerat in egestas erat. Massa id neque aliquam vestibulum morbi blandit cursus.")
	chunk.Style.Font = fontRegular

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Text alignment justify - character spacing.
	p = c.NewStyledParagraph()
	p.SetMargins(0, 0, 20, 0)
	p.SetTextAlignment(TextAlignmentJustify)

	chunk = p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Sed enim ut sem viverra aliquet eget sit amet. Sed viverra ipsum nunc aliquet bibendum enim facilisis. Consectetur libero id faucibus nisl. Cras adipiscing enim eu turpis egestas pretium aenean pharetra. Viverra suspendisse potenti nullam ac tortor vitae purus. Ultrices in iaculis nunc sed augue lacus viverra. Sed pulvinar proin gravida hendrerit lectus. Tempus quam pellentesque nec nam aliquam sem et tortor consequat. Magna ac placerat vestibulum lectus mauris ultrices eros in. Non diam phasellus vestibulum lorem. Lorem donec massa sapien faucibus et molestie ac. Est placerat in egestas erat. Massa id neque aliquam vestibulum morbi blandit cursus.")
	chunk.Style.Font = fontRegular
	chunk.Style.CharSpacing = 0.75

	err = c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("styled_paragraph_character_spacing.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestStyledLinkRotation(t *testing.T) {
	c := New()
	c.NewPage()

	// Rotation 0 degrees.
	p := c.NewStyledParagraph()
	p.AddExternalLink("Link with no rotation", "https://google.com")
	p.SetPos(50, 50)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotation 45 degrees.
	p = c.NewStyledParagraph()
	p.AddExternalLink("Link rotated 45 degrees", "https://google.com")
	p.SetAngle(45)
	p.SetPos(200, 100)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotation -45 degrees.
	p = c.NewStyledParagraph()
	p.AddExternalLink("Link rotated -45 degrees", "https://google.com")
	p.SetAngle(-45)
	p.SetPos(350, 25)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotation 90 degrees.
	p = c.NewStyledParagraph()
	p.AddExternalLink("Link rotated 90 degrees", "https://google.com")
	p.SetAngle(90)
	p.SetPos(50, 250)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotation -90 degrees.
	p = c.NewStyledParagraph()
	p.AddExternalLink("Link rotated -90 degrees", "https://google.com")
	p.SetAngle(-90)
	p.SetPos(100, 145)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotation 180 degrees.
	p = c.NewStyledParagraph()
	p.AddExternalLink("Link rotated 180 degrees", "https://google.com")
	p.SetAngle(180)
	p.SetPos(250, 150)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotation 225 degrees.
	p = c.NewStyledParagraph()
	p.AddExternalLink("Link rotated 225 degrees", "https://google.com")
	p.SetAngle(225)
	p.SetPos(375, 150)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotation 135 degrees.
	p = c.NewStyledParagraph()
	p.AddExternalLink("Link rotated 135 degrees", "https://google.com")
	p.SetAngle(135)
	p.SetPos(525, 230)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotate paragraph with text 90 degrees.
	p = c.NewStyledParagraph()
	p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.AddExternalLink("Link rotated 90 degrees", "https://google.com")
	p.SetAngle(90)
	p.SetWidth(200)
	p.SetPos(50, 500)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotate paragraph with text -90 degrees.
	p = c.NewStyledParagraph()
	p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.AddExternalLink("Link rotated -90 degrees", "https://google.com")
	p.SetAngle(-90)
	p.SetWidth(200)
	p.SetPos(340, 310)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Rotate paragraph with text -45 degrees.
	p = c.NewStyledParagraph()
	p.Append("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt. Dis parturient montes nascetur ridiculus mus. Pharetra diam sit amet nisl suscipit adipiscing. Proin fermentum leo vel orci porta. Id diam vel quam elementum pulvinar.")
	p.AddExternalLink("Link rotated -45 degrees", "https://google.com")
	p.SetAngle(-45)
	p.SetWidth(200)
	p.SetPos(460, 310)

	if err := c.Draw(p); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err := c.WriteToFile(tempFile("styled_paragraph_rotated_links.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestStyledParagraphTableVerticalAlignment(t *testing.T) {
	c := New()

	fontRegular := newStandard14Font(t, model.CourierName)

	createTable := func(c *Creator, text string, align CellVerticalAlignment, fontSize float64) {
		textStyle := c.NewTextStyle()
		textStyle.Font = fontRegular
		textStyle.FontSize = fontSize

		table := c.NewTable(1)
		table.SetMargins(0, 0, 5, 5)

		cell := table.NewCell()
		sp := c.NewStyledParagraph()
		textChunk := sp.Append(text)
		textChunk.Style = textStyle

		cell.SetVerticalAlignment(align)
		cell.SetContent(sp)
		cell.SetBorder(CellBorderSideAll, CellBorderStyleSingle, 1)

		if err := c.Draw(table); err != nil {
			t.Fatalf("Error drawing: %v", err)
		}
	}

	chunks := []string{
		"TR",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Lacus viverra vitae congue eu consequat. Cras adipiscing enim eu turpis. Lectus magna fringilla urna porttitor. Condimentum id venenatis a condimentum. Quis ipsum suspendisse ultrices gravida dictum fusce. In fermentum posuere urna nec tincidunt.",
	}

	alignments := []struct {
		Label     string
		Alignment CellVerticalAlignment
	}{
		{"Top alignment", CellVerticalAlignmentTop},
		{"Middle alignment", CellVerticalAlignmentMiddle},
		{"Bottom alignment", CellVerticalAlignmentBottom},
	}

	for _, chunk := range chunks {
		for _, alignment := range alignments {
			c.NewPage()

			sp := c.NewStyledParagraph()
			sp.Append(alignment.Label).Style.FontSize = 16
			sp.SetMargins(0, 0, 0, 5)

			if err := c.Draw(sp); err != nil {
				t.Fatalf("Error drawing: %v", err)
			}

			for i := 4; i <= 20; i += 2 {
				createTable(c, chunk, alignment.Alignment, float64(i))
			}
		}
	}

	// Write output file.
	testWriteAndRender(t, c, "styled_paragraph_table_vertical_align.pdf")
}
