/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package annotator

import (
	"errors"
	"strings"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

// FieldAppearance implements interface model.FieldAppearanceGenerator and generates appearance streams
// for fields taking into account what value is in the field.  A common use case is for generating the
// appearance stream prior to flattening fields.
//
// If `OnlyIfMissing` is true, the field appearance is generated only for fields that do not have an
// appearance stream specified.
type FieldAppearance struct {
	OnlyIfMissing bool
}

// GenerateAppearanceDict generates an appearance dictionary for widget annotation `wa` for the `field` in `form`.
func (fa FieldAppearance) GenerateAppearanceDict(form *model.PdfAcroForm, field *model.PdfField, wa *model.PdfAnnotationWidget) (*core.PdfObjectDictionary, error) {
	common.Log.Trace("GenerateAppearanceDict for %v  V: %+v", field.PartialName(), field.V)
	appDict, has := core.GetDict(wa.AP)
	if has && fa.OnlyIfMissing {
		common.Log.Trace("Already populated - ignoring")
		return appDict, nil
	}

	// Generate the appearance.
	switch t := field.GetContext().(type) {
	case *model.PdfFieldText:
		ftxt := t

		common.Log.Trace("Flags: %s", ftxt.Flags())

		// Handle special cases.
		switch {
		case ftxt.Flags().Has(model.FieldFlagPassword):
			// Should never store password values.
			return nil, nil
		case ftxt.Flags().Has(model.FieldFlagFileSelect):
			// Not supported.
			return nil, nil
		case ftxt.Flags().Has(model.FieldFlagComb):
			// Special handling for comb. Only if max len is set.
			if ftxt.MaxLen != nil {
				appDict, err := genFieldTextCombAppearance(wa, ftxt, form.DR)
				if err != nil {
					return nil, err
				}
				return appDict, nil
			}
		}

		appDict, err := genFieldTextAppearance(wa, ftxt, form.DR)
		if err != nil {
			return nil, err
		}

		return appDict, nil
	case *model.PdfFieldButton:
		fbtn := t
		if fbtn.IsCheckbox() {
			appDict, err := genFieldCheckboxAppearance(wa, fbtn, form.DR)
			if err != nil {
				return nil, err
			}
			return appDict, nil
		}
		common.Log.Debug("UNHANDLED button type: %+v", fbtn.GetType())
	default:
		common.Log.Debug("UNHANDLED field type: %T", t)
	}

	return nil, nil
}

// genTextAppearance generates the appearance stream for widget annotation `wa` with text field `ftxt`.
// It requires access to the form resources DR entry via `dr`.
func genFieldTextAppearance(wa *model.PdfAnnotationWidget, ftxt *model.PdfFieldText, dr *model.PdfPageResources) (*core.PdfObjectDictionary, error) {
	resources := model.NewPdfPageResources()

	// Get bounding Rect.
	array, ok := core.GetArray(wa.Rect)
	if !ok {
		return nil, errors.New("invalid Rect")
	}
	rect, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	if len(rect) != 4 {
		return nil, errors.New("len(Rect) != 4")
	}

	width := rect[2] - rect[0]
	height := rect[3] - rect[1]

	// Get and process the default appearance string (DA) operands.
	da := getDA(ftxt.PdfField)
	csp := contentstream.NewContentStreamParser(da)
	daOps, err := csp.Parse()
	if err != nil {
		return nil, err
	}

	cc := contentstream.NewContentCreator()
	cc.Add_BMC("Tx")
	cc.Add_q()
	// Graphic state changes.
	cc.Add_BT()

	// Add DA operands.
	var fontsize float64
	var fontname *core.PdfObjectName
	var font *model.PdfFont
	fontsizeDef := height * 0.64
	for _, op := range *daOps {
		// TODO: If TF specified and font size is 0, it means we should set on our own based on the Rect.
		if op.Operand == "Tf" && len(op.Params) == 2 {
			if name, ok := core.GetName(op.Params[0]); ok {
				fontname = name
			}
			num, err := core.GetNumberAsFloat(op.Params[1])
			if err == nil {
				fontsize = num
			} else {
				common.Log.Debug("ERROR invalid font size: %v", op.Params[1])
			}
			if fontsize == 0 {
				// Use default if zero.
				fontsize = fontsizeDef
			}
			op.Params[1] = core.MakeFloat(fontsize)
		}
		cc.AddOperand(*op)
	}

	// If fontname not set need to make a new font or use one defined in the resources.
	// e.g. Helv commonly used for Helvetica.
	if fontname == nil {
		// Font not set, revert to Helvetica with name "Helv".
		fontname = core.MakeName("Helv")
		helv, err := model.NewStandard14Font("Helvetica")
		if err != nil {
			return nil, err
		}
		font = helv
		resources.SetFontByName(*fontname, helv.ToPdfObject())
		cc.Add_Tf(*fontname, fontsizeDef)
	} else {
		fontobj, has := dr.GetFontByName(*fontname)
		if !has {
			return nil, errors.New("font not in DR")
		}
		font, err = model.NewPdfFontFromPdfObject(fontobj)
		if err != nil {
			common.Log.Debug("ERROR loading default appearance font: %v", err)
			return nil, err
		}
		resources.SetFontByName(*fontname, fontobj)
	}
	encoder := font.Encoder()

	tx := 2.0
	ty := 0.26 * height

	text := ""
	if str, ok := core.GetString(ftxt.V); ok {
		text = str.Decoded()
	}

	lines := []string{text}

	// Handle multi line fields.
	if ftxt.Flags().Has(model.FieldFlagMultiline) {
		// Start at the top of the bounding box.
		ty = height - ty
		text = strings.Replace(text, "\r\n", "\n", -1)
		text = strings.Replace(text, "\r", "\n", -1)
		lines = strings.Split(text, "\n")
	}

	if encoder != nil {
		for i := range lines {
			lines[i] = encoder.Encode(lines[i])
		}
	}

	lineheight := 1.0 * fontsize
	cc.Add_Td(tx, ty)
	for i, line := range lines {
		cc.Add_Tj(*core.MakeString(line))

		if i < len(lines)-1 {
			cc.Add_Td(0, -lineheight)
		}
	}

	cc.Add_ET()
	cc.Add_Q()
	cc.Add_EMC()

	xform := model.NewXObjectForm()
	xform.Resources = resources
	xform.BBox = core.MakeArrayFromFloats([]float64{0, 0, width, height})
	xform.SetContentStream(cc.Bytes(), core.NewRawEncoder())

	apDict := core.MakeDict()
	apDict.Set("N", xform.ToPdfObject())

	return apDict, nil
}

// genFieldTextCombAppearance generates an appearance dictionary for a comb text field where the width is split
// into equal size boxes.
func genFieldTextCombAppearance(wa *model.PdfAnnotationWidget, ftxt *model.PdfFieldText, dr *model.PdfPageResources) (*core.PdfObjectDictionary, error) {
	resources := model.NewPdfPageResources()

	// Get bounding Rect.
	array, ok := core.GetArray(wa.Rect)
	if !ok {
		return nil, errors.New("invalid Rect")
	}
	rect, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	if len(rect) != 4 {
		return nil, errors.New("len(Rect) != 4")
	}

	width := rect[2] - rect[0]
	height := rect[3] - rect[1]

	maxLen, has := core.GetIntVal(ftxt.MaxLen)
	if !has {
		return nil, errors.New("maxlen not set")
	}
	if maxLen <= 0 {
		return nil, errors.New("maxLen invalid")
	}

	boxwidth := float64(width) / float64(maxLen)

	// Get and process the default appearance string (DA) operands.
	da := getDA(ftxt.PdfField)
	csp := contentstream.NewContentStreamParser(da)
	daOps, err := csp.Parse()
	if err != nil {
		return nil, err
	}

	cc := contentstream.NewContentCreator()
	cc.Add_BMC("Tx")
	cc.Add_q()
	// Graphic state changes.
	cc.Add_BT()

	// Add DA operands.
	var fontsize float64
	var fontname *core.PdfObjectName
	var font *model.PdfFont
	fontsizeDef := height * 0.64
	for _, op := range *daOps {
		// TODO: If TF specified and font size is 0, it means we should set on our own based on the Rect.
		if op.Operand == "Tf" && len(op.Params) == 2 {
			if name, ok := core.GetName(op.Params[0]); ok {
				fontname = name
			}
			num, err := core.GetNumberAsFloat(op.Params[1])
			if err == nil {
				fontsize = num
			} else {
				common.Log.Debug("ERROR invalid font size: %v", op.Params[1])
			}
			if fontsize == 0 {
				// Use default if zero.
				fontsize = fontsizeDef
			}
			op.Params[1] = core.MakeFloat(fontsize)
		}
		cc.AddOperand(*op)
	}

	// If fontname not set need to make a new font or use one defined in the resources.
	// e.g. Helv commonly used for Helvetica.
	if fontname == nil {
		// Font not set, revert to Helvetica with name "Helv".
		fontname = core.MakeName("Helv")
		helv, err := model.NewStandard14Font("Helvetica")
		if err != nil {
			return nil, err
		}
		font = helv
		resources.SetFontByName(*fontname, helv.ToPdfObject())
		cc.Add_Tf(*fontname, fontsizeDef)
	} else {
		fontobj, has := dr.GetFontByName(*fontname)
		if !has {
			return nil, errors.New("font not in DR")
		}
		font, err = model.NewPdfFontFromPdfObject(fontobj)
		if err != nil {
			common.Log.Debug("ERROR loading default appearance font: %v", err)
			return nil, err
		}
		resources.SetFontByName(*fontname, fontobj)
	}
	encoder := font.Encoder()

	text := ""
	if str, ok := core.GetString(ftxt.V); ok {
		text = str.Decoded()
	}

	//tx := 2.0

	ty := 0.26 * height
	cc.Add_Td(0, ty)

	if quadding, has := core.GetIntVal(ftxt.Q); has {
		switch quadding {
		case 2: // Right justified.
			if len(text) < maxLen {
				offset := float64(maxLen-len(text)) * boxwidth
				cc.Add_Td(offset, 0)
			}
		}
	}

	for i, r := range text {
		tx := 2.0
		encoded := string(r)
		if encoder != nil {
			glyph, has := encoder.RuneToGlyph(r)
			if !has {
				common.Log.Debug("ERROR: Rune not found %#v - skipping over", r)
				continue
			}
			metrics, found := font.GetGlyphCharMetrics(glyph)
			if !found {
				common.Log.Debug("ERROR: Glyph not found in font: %v - skipping over", glyph)
				continue
			}

			encoded = encoder.Encode(encoded)

			// Calculate indent such that the glyph is positioned in the center.
			glyphwidth := fontsize * metrics.Wx / 1000.0
			calcIndent := (boxwidth - glyphwidth) / 2
			tx = calcIndent
		}

		cc.Add_Td(tx, 0)
		cc.Add_Tj(*core.MakeString(encoded))

		if i != len(text)-1 {
			cc.Add_Td(boxwidth-tx, 0)
		}
	}

	cc.Add_ET()
	cc.Add_Q()
	cc.Add_EMC()

	xform := model.NewXObjectForm()
	xform.Resources = resources
	xform.BBox = core.MakeArrayFromFloats([]float64{0, 0, width, height})
	xform.SetContentStream(cc.Bytes(), core.NewRawEncoder())

	apDict := core.MakeDict()
	apDict.Set("N", xform.ToPdfObject())

	return apDict, nil
}

// genFieldCheckboxAppearance generates an appearance dictionary for a widget annotation `wa` referenced by
// a button field `fbtn` with form resources `dr` (DR).
func genFieldCheckboxAppearance(wa *model.PdfAnnotationWidget, fbtn *model.PdfFieldButton, dr *model.PdfPageResources) (*core.PdfObjectDictionary, error) {
	dataEncoder := core.NewRawEncoder() // For debugging.

	// Get bounding Rect.
	array, ok := core.GetArray(wa.Rect)
	if !ok {
		return nil, errors.New("invalid Rect")
	}
	rect, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	if len(rect) != 4 {
		return nil, errors.New("len(Rect) != 4")
	}

	width := rect[2] - rect[0]
	height := rect[3] - rect[1]

	xformOn := model.NewXObjectForm()
	{
		cc := contentstream.NewContentCreator()
		cc.Add_g(0.75293).
			Add_re(0, 0, width, height).
			Add_f().
			Add_re(0, 0, width-1, height-1).
			Add_s()
		cc.Add_q().
			Add_re(1, 1, width-2, height-2).
			Add_W().
			Add_n().
			Add_g(0).
			Add_BT().
			Add_Tf("ZaDb", 0.65*height).
			Add_Td(0.15*width, 0.27*height).
			Add_TL(0.66 * height).
			Add_Tj(*core.MakeString("\064")).
			Add_ET().
			Add_Q()

		zapfdb, err := model.NewStandard14Font("ZapfDingbats")
		if err != nil {
			return nil, err
		}

		xformOn.Resources = model.NewPdfPageResources()
		xformOn.Resources.SetFontByName("ZaDb", zapfdb.ToPdfObject())
		xformOn.BBox = core.MakeArrayFromFloats([]float64{0, 0, width, height})
		xformOn.SetContentStream(cc.Bytes(), dataEncoder)
	}

	xformOff := model.NewXObjectForm()
	{
		cc := contentstream.NewContentCreator()
		cc.Add_g(0.75293)
		cc.Add_re(0, 0, width, height)
		cc.Add_f()
		cc.Add_re(0, 0, width-1, height-1)
		xformOff.BBox = core.MakeArrayFromFloats([]float64{0, 0, width, height})
		xformOff.SetContentStream(cc.Bytes(), dataEncoder)
	}

	dchoiceapp := core.MakeDict()
	dchoiceapp.Set("Off", xformOff.ToPdfObject())
	dchoiceapp.Set("Yes", xformOn.ToPdfObject())

	appDict := core.MakeDict()
	appDict.Set("N", dchoiceapp)

	return appDict, nil
}

// getDA returns the default appearance text (DA) for a given field `ftxt`.
// If not set for `ftxt` then checks if set by Parent (inherited), otherwise
// returns "".
func getDA(field *model.PdfField) string {
	if field == nil {
		return ""
	}

	ftxt, ok := field.GetContext().(*model.PdfFieldText)
	if !ok {
		return getDA(field.Parent)
	}

	if ftxt.DA != nil {
		return ftxt.DA.Str()
	}

	return getDA(ftxt.Parent)
}
