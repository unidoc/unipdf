package draw

import (
	pdfcontent "github.com/unidoc/unipdf/v3/contentstream"
)

// DrawPathWithCreator makes the path with the content creator.
// Adds the PDF commands to draw the path to the creator instance.
func DrawPathWithCreator(path Path, creator *pdfcontent.ContentCreator) {
	for idx, p := range path.Points {
		if idx == 0 {
			creator.Add_m(p.X, p.Y)
		} else {
			creator.Add_l(p.X, p.Y)
		}
	}
}

// DrawBezierPathWithCreator makes the bezier path with the content creator.
// Adds the PDF commands to draw the path to the creator instance.
func DrawBezierPathWithCreator(bpath CubicBezierPath, creator *pdfcontent.ContentCreator) {
	for idx, c := range bpath.Curves {
		if idx == 0 {
			creator.Add_m(c.P0.X, c.P0.Y)
		}
		creator.Add_c(c.P1.X, c.P1.Y, c.P2.X, c.P2.Y, c.P3.X, c.P3.Y)
	}
}
