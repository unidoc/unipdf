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

	border.StyleTop = CellBorderStyleDoubleTop
	border.StyleBottom = CellBorderStyleDoubleBottom
	border.StyleLeft = CellBorderStyleDoubleLeft
	border.StyleRight = CellBorderStyleDoubleRight

	c := New()
	c.Draw(border)

	err := c.WriteToFile("/tmp/border_single.pdf")
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

	border.StyleTop = CellBorderStyleDoubleTop
	border.StyleBottom = CellBorderStyleDoubleBottom

	c := New()
	c.Draw(border)

	err := c.WriteToFile("/tmp/border_single2.pdf")
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

	border.StyleLeft = CellBorderStyleDoubleLeft
	border.StyleRight = CellBorderStyleDoubleRight

	c := New()
	c.Draw(border)

	err := c.WriteToFile("/tmp/border_single3.pdf")
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

	border.StyleTop = CellBorderStyleTop
	border.StyleBottom = CellBorderStyleDoubleBottom
	border.StyleLeft = CellBorderStyleLeft
	border.StyleRight = CellBorderStyleDoubleRight

	c := New()
	c.Draw(border)

	err := c.WriteToFile("/tmp/border_single4.pdf")
	if err != nil {
		t.Errorf("Fail: %v\n", err)
		return
	}
}
