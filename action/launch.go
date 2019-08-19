/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package action

import (
	pdfcore "github.com/unidoc/unipdf/v3/core"
	pdf "github.com/unidoc/unipdf/v3/model"
)
type LaunchActionDef struct {
	File      float64
	Win       *LaunchActionWinDef
	Mac       *LaunchActionMacDef
	Unix      *LaunchActionUnixDef
	NewWindow bool
}

type LaunchActionWinDef struct {
	File             string
	DefaultDirectory string
	Operation        string
	Parameters       string
}

type LaunchActionMacDef struct {
}

type LaunchActionUnixDef struct {
}

// CreateCircleAnnotation creates a circle/ellipse annotation object with appearance stream that can be added to
// page PDF annotations.
func CreateLaunchAction(launchDef LaunchActionDef) (*pdf.PdfAction, error) {
	launchAction := pdf.NewPdfActionLaunch()

	if circDef.BorderEnabled {
		r, g, b := circDef.BorderColor.R(), circDef.BorderColor.G(), circDef.BorderColor.B()
		circAnnotation.C = pdfcore.MakeArrayFromFloats([]float64{r, g, b})
		bs := pdf.NewBorderStyle()
		bs.SetBorderWidth(circDef.BorderWidth)
		circAnnotation.BS = bs.ToPdfObject()
	}

	if circDef.FillEnabled {
		r, g, b := circDef.FillColor.R(), circDef.FillColor.G(), circDef.FillColor.B()
		circAnnotation.IC = pdfcore.MakeArrayFromFloats([]float64{r, g, b})
	} else {
		circAnnotation.IC = pdfcore.MakeArrayFromIntegers([]int{}) // No fill.
	}

	if circDef.Opacity < 1.0 {
		circAnnotation.CA = pdfcore.MakeFloat(circDef.Opacity)
	}

	// Make the appearance stream (for uniform appearance).
	apDict, bbox, err := makeCircleAnnotationAppearanceStream(circDef)
	if err != nil {
		return nil, err
	}

	circAnnotation.AP = apDict
	circAnnotation.Rect = pdfcore.MakeArrayFromFloats([]float64{bbox.Llx, bbox.Lly, bbox.Urx, bbox.Ury})

	return circAnnotation.PdfAnnotation, nil

}
