/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/unidoc/unidoc/pdf/model/fonts"
)

var seed = rand.New(rand.NewSource(time.Now().UnixNano()))

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

func TestDivVertical(t *testing.T) {
	fontRegular := fonts.NewFontCourier()
	fontBold := fonts.NewFontCourierBold()

	c := New()
	c.NewPage()

	// Draw section title.
	p := NewParagraph("Regular division component")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Draw division.
	div := NewDivision()

	if div.Inline() {
		t.Fatal("Fail: Incorrect inline mode value")
	}

	p = NewParagraph("Components are stacked vertically ")
	p.SetFont(fontRegular)
	div.Add(p)

	p = NewParagraph("but not horizontally")
	p.SetFont(fontBold)
	div.Add(p)

	// Add styled paragraph
	style := NewTextStyle()
	style.Color = ColorRGBFrom8bit(0, 0, 255)

	s := NewStyledParagraph("Not even with a styled ", style)

	style.Color = ColorRGBFrom8bit(255, 0, 0)
	s.Append("paragraph", style)

	div.Add(s)

	err = c.Draw(div)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/division_vertical.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestDivInline(t *testing.T) {
	fontRegular := fonts.NewFontCourier()
	fontBold := fonts.NewFontCourierBold()

	c := New()
	c.NewPage()

	// Draw section title.
	p := NewParagraph("Inline division component")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Draw division.
	div := NewDivision()
	div.SetInline(true)

	if !div.Inline() {
		t.Fatal("Fail: Incorrect inline mode value")
	}

	p = NewParagraph("Components are stacked both vertically ")
	p.SetEnableWrap(false)
	p.SetFont(fontRegular)
	div.Add(p)

	p = NewParagraph("and horizontally. ")
	p.SetEnableWrap(false)
	p.SetFont(fontBold)
	div.Add(p)

	p = NewParagraph("Only if they fit right!")
	p.SetEnableWrap(false)
	p.SetFont(fontRegular)
	div.Add(p)

	p = NewParagraph("This one did not fit in the available line space. ")
	p.SetEnableWrap(false)
	p.SetFont(fontBold)
	div.Add(p)

	// Add styled paragraph
	style := NewTextStyle()
	style.Color = ColorRGBFrom8bit(0, 0, 255)

	s := NewStyledParagraph("This styled paragraph should ", style)

	style.Color = ColorRGBFrom8bit(255, 0, 0)
	s.Append("fit", style)

	style.Color = ColorRGBFrom8bit(0, 255, 0)
	style.Font = fontBold
	s.Append(" right in.", style)

	div.Add(s)

	err = c.Draw(div)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/division_inline.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestDivNumberMatrix(t *testing.T) {
	fontRegular := fonts.NewFontCourier()
	fontBold := fonts.NewFontCourierBold()

	c := New()
	c.NewPage()

	// Draw section title.
	p := NewParagraph("A list of numbers in an inline division")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Draw division.
	div := NewDivision()
	div.SetInline(true)

	for i := 0; i < 100; i++ {
		r := byte(seed.Intn(200))
		g := byte(seed.Intn(200))
		b := byte(seed.Intn(200))

		p := NewParagraph(strconv.Itoa(i) + " ")
		p.SetEnableWrap(false)
		p.SetColor(ColorRGBFrom8bit(r, g, b))

		if seed.Intn(2)%2 == 0 {
			p.SetFont(fontRegular)
		} else {
			p.SetFont(fontBold)
		}

		div.Add(p)
	}

	err = c.Draw(div)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/division_number_matrix.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestDivRandomSequences(t *testing.T) {
	fontRegular := fonts.NewFontHelvetica()
	fontBold := fonts.NewFontHelveticaBold()

	c := New()
	c.NewPage()

	// Draw section title.
	p := NewParagraph("Inline division of random sequences on multiple pages")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Draw division.
	div := NewDivision()
	div.SetInline(true)

	style := NewTextStyle()

	for i := 0; i < 350; i++ {
		r := byte(seed.Intn(200))
		g := byte(seed.Intn(200))
		b := byte(seed.Intn(200))

		word := RandString(seed.Intn(10)+5) + " "
		fontSize := float64(11 + seed.Intn(3))

		if seed.Intn(2)%2 == 0 {
			p := NewParagraph(word)
			p.SetEnableWrap(false)
			p.SetColor(ColorRGBFrom8bit(r, g, b))
			p.SetFontSize(fontSize)

			if seed.Intn(2)%2 == 0 {
				p.SetFont(fontBold)
			} else {
				p.SetFont(fontRegular)
			}

			div.Add(p)
		} else {
			style.Color = ColorRGBFrom8bit(r, g, b)
			style.FontSize = fontSize

			if seed.Intn(2)%2 == 0 {
				style.Font = fontBold
			} else {
				style.Font = fontRegular
			}

			s := NewStyledParagraph(word, style)
			s.SetEnableWrap(false)
			div.Add(s)
		}
	}

	err = c.Draw(div)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/division_random_sequences.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestTableDivisions(t *testing.T) {
	fontRegular := fonts.NewFontHelvetica()
	fontBold := fonts.NewFontHelveticaBold()

	c := New()
	c.NewPage()

	// Draw section title.
	p := NewParagraph("Table containing division components")
	p.SetMargins(0, 0, 20, 10)
	p.SetFont(fontBold)

	err := c.Draw(p)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	table := NewTable(2)
	table.SetColumnWidths(0.35, 0.65)

	// Add regular division to table.
	divRegular := NewDivision()

	p = NewParagraph("Components are stacked vertically ")
	p.SetFont(fontRegular)
	divRegular.Add(p)

	p = NewParagraph("but not horizontally")
	p.SetFont(fontBold)
	divRegular.Add(p)

	cell := table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(divRegular)

	// Add inline division to table.
	divInline := NewDivision()
	divInline.SetInline(true)

	p = NewParagraph("Components are stacked vertically ")
	p.SetEnableWrap(false)
	p.SetFont(fontRegular)
	divInline.Add(p)

	p = NewParagraph("and horizontally. ")
	p.SetEnableWrap(false)
	p.SetFont(fontBold)
	divInline.Add(p)

	p = NewParagraph("Only if they fit!")
	p.SetEnableWrap(false)
	p.SetFont(fontRegular)
	divInline.Add(p)

	p = NewParagraph("This one did not fit in the available line space")
	p.SetEnableWrap(false)
	p.SetFont(fontBold)
	divInline.Add(p)

	cell = table.NewCell()
	cell.SetBorder(CellBorderStyleBox, 1)
	cell.SetContent(divInline)

	// Draw table.
	err = c.Draw(table)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/division_table.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
