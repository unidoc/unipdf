/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import "testing"

// This test file contains multiple tests to generate PDFs. The outputs are written into /tmp as files.  The files
// themselves need to be observed to check for correctness as we don't have a good way to automatically check
// if every detail is correct.

// TestSingleBorder tests drawing a border with double border on all sides.
func TestSingleBorder(t *testing.T) {
	border := newBorder(100, 100, 100, 100)
	border.SetColorBottom(ColorGreen)
	border.SetColorTop(ColorGreen)
	border.SetColorLeft(ColorRed)
	border.SetColorRight(ColorRed)

	border.SetWidthBottom(3)
	border.SetWidthTop(3)
	border.SetWidthLeft(3)
	border.SetWidthRight(3)

	border.SetStyleBottom(CellBorderStyleDouble)
	border.SetStyleTop(CellBorderStyleDouble)
	border.SetStyleLeft(CellBorderStyleDouble)
	border.SetStyleRight(CellBorderStyleDouble)

	c := New()
	c.Draw(border)

	err := c.WriteToFile(tempFile("border_single.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// TestSingleBorder2 tests drawing double border with only top and bottom set.
func TestSingleBorder2(t *testing.T) {
	border := newBorder(100, 100, 100, 100)
	border.SetColorBottom(ColorGreen)
	border.SetColorTop(ColorGreen)

	border.SetWidthBottom(3)
	border.SetWidthTop(3)

	border.SetStyleBottom(CellBorderStyleDouble)
	border.SetStyleTop(CellBorderStyleDouble)

	c := New()
	c.Draw(border)

	err := c.WriteToFile(tempFile("border_single2.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// TestSingleBorder3 tests drawing double border with only left and bottom borders.
func TestSingleBorder3(t *testing.T) {
	border := newBorder(100, 100, 100, 100)
	border.SetColorLeft(ColorRed)
	border.SetColorRight(ColorRed)

	border.SetWidthLeft(3)
	border.SetWidthRight(3)

	border.SetStyleLeft(CellBorderStyleDouble)
	border.SetStyleRight(CellBorderStyleDouble)

	c := New()
	c.Draw(border)

	err := c.WriteToFile(tempFile("border_single3.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}

// TestSingleBorder4 test drawing a combination of double and single borders with single border on top and left,
// and double on bottom and right.
func TestSingleBorder4(t *testing.T) {
	border := newBorder(100, 100, 100, 100)
	border.SetColorBottom(ColorGreen)
	border.SetColorTop(ColorGreen)
	border.SetColorLeft(ColorRed)
	border.SetColorRight(ColorRed)

	border.SetWidthBottom(3)
	border.SetWidthTop(3)
	border.SetWidthLeft(3)
	border.SetWidthRight(3)

	border.SetStyleTop(CellBorderStyleSingle)
	border.SetStyleBottom(CellBorderStyleDouble)
	border.SetStyleLeft(CellBorderStyleSingle)
	border.SetStyleRight(CellBorderStyleDouble)

	c := New()
	c.Draw(border)

	err := c.WriteToFile(tempFile("border_single4.pdf"))
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}
