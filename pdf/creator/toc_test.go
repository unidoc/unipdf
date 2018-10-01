/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
package creator

import (
	"testing"

	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func TestTOCAdvanced(t *testing.T) {
	fontHelvetica := fonts.NewFontHelvetica()
	fontHelveticaBold := fonts.NewFontHelveticaBold()

	c := New()
	c.NewPage()

	toc := NewTOC("Table of Contents")

	// Set separator and margins for all the lines.
	toc.SetLineSeparator(".")
	toc.SetLineMargins(0, 0, 2, 2)
	toc.SetLineLevelOffset(12)

	// Set style for all line numbers.
	grey := ColorRGBFrom8bit(100, 100, 100)
	red := ColorRGBFrom8bit(255, 0, 0)

	style := NewTextStyle()
	style.Font = fontHelveticaBold
	style.Color = grey
	toc.SetLineNumberStyle(style)

	// Set style for all line pages.
	style.Font = fontHelveticaBold
	style.Color = ColorRGBFrom8bit(0, 0, 0)
	toc.SetLinePageStyle(style)

	// Set style for all line separators.
	style.Font = fontHelvetica
	style.FontSize = 9
	toc.SetLineSeparatorStyle(style)

	// Add TOC lines.
	tl := toc.Add("", "Abstract", "i", 1)
	tl.Title.Style.Font = fontHelveticaBold
	tl.SetMargins(0, 0, 5, 5)

	toc.Add("", "Aknowledgements", "ii", 1)

	// Customize line style.
	tl = toc.Add("", "Table of Contents", "iii", 1)
	tl.Title.Style.Font = fontHelveticaBold
	tl.Title.Style.Color = grey

	// Customize line style.
	tl = toc.Add("Chapter 1:", "Introduction", "1", 1)
	tl.Title.Style.Font = fontHelveticaBold
	tl.Title.Style.Color = red
	tl.Number.Style.Color = red
	tl.Page.Style.Color = red
	tl.Separator.Style.Color = red

	toc.Add("1.1", "Second Harmonic Generation (SHG)", "1", 2)
	toc.Add("1.1.1", "Nonlinear induced polarization", "1", 3)
	toc.Add("1.1.2", "Phase matching of the fundamental and emission waves", "2", 3)
	toc.Add("1.1.3", "Collagen as an intrinsic biomarker for SHG generation", "3", 3)
	toc.Add("1.1.4", "Second harmonic imaging microscopy", "6", 3)
	toc.Add("1.2", "Light propagation in tissues", "8", 2)
	toc.Add("1.2.1", "Radiative transfer equation for modeling light propagation in tissue", "8", 3)
	toc.Add("1.2.2", "Monte Carlo method as a convenient and flexible solution to the RTE for modeling light transport in multi layered tissues", "10", 3)
	toc.Add("1.2.3", "Measurement of optical properties", "15", 3)
	toc.Add("1.2.4", "Analytical solution of light scattering: The Born aproximation", "19", 3)
	toc.Add("1.2.5", "Refractive index corellation functions to describe light scattering in tissue", "21", 3)
	toc.Add("1.3", "SHG creation and emission directionality", "24", 2)
	toc.Add("1.4", "Combining SGH creation and emission directionality", "26", 2)
	toc.Add("1.5", "Utilizing light to characterize tissue structure", "26", 2)

	// Customize line style.
	tl = toc.Add("", "References", "28", 2)
	tl.Title.Style.Font = fontHelveticaBold
	tl.Separator.Style.Font = fontHelveticaBold
	tl.SetMargins(0, 0, 5, 0)

	err := c.Draw(toc)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/toc_advanced.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
