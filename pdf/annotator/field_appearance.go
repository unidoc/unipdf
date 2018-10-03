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
	style         *AppearanceStyle
}

// AppearanceStyle defines style parameters for appearance stream generation.
type AppearanceStyle struct {
	// How much of Rect height to fill when autosizing text.
	AutoFontSizeFraction float64
	// Glyph used for check mark in checkboxes (for ZapfDingbats font).
	CheckmarkGlyph string
}

const (
	defaultAutoFontSizeFraction = 0.818
	defaultCheckmarkGlyph       = "a20"
)

// SetStyle applies appearance `style` to `fa`.
func (fa *FieldAppearance) SetStyle(style AppearanceStyle) {
	fa.style = &style
}

// Style returns the appearance style of `fa`. If not specified, returns default style.
func (fa FieldAppearance) Style() AppearanceStyle {
	if fa.style != nil {
		return *fa.style
	}
	return AppearanceStyle{
		AutoFontSizeFraction: defaultAutoFontSizeFraction,
		CheckmarkGlyph:       defaultCheckmarkGlyph,
	}
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
				appDict, err := genFieldTextCombAppearance(wa, ftxt, form.DR, fa.Style())
				if err != nil {
					return nil, err
				}
				return appDict, nil
			}
		}

		appDict, err := genFieldTextAppearance(wa, ftxt, form.DR, fa.Style())
		if err != nil {
			return nil, err
		}

		return appDict, nil
	case *model.PdfFieldButton:
		fbtn := t
		if fbtn.IsCheckbox() {
			appDict, err := genFieldCheckboxAppearance(wa, fbtn, form.DR, fa.Style())
			if err != nil {
				return nil, err
			}
			return appDict, nil
		}
		common.Log.Debug("TODO: UNHANDLED button type: %+v", fbtn.GetType())
	default:
		common.Log.Debug("TODO: UNHANDLED field type: %T", t)
	}

	return nil, nil
}

// genTextAppearance generates the appearance stream for widget annotation `wa` with text field `ftxt`.
// It requires access to the form resources DR entry via `dr`.
func genFieldTextAppearance(wa *model.PdfAnnotationWidget, ftxt *model.PdfFieldText, dr *model.PdfPageResources, style AppearanceStyle) (*core.PdfObjectDictionary, error) {
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
	autosize := true

	fontsizeDef := height * style.AutoFontSizeFraction
	for _, op := range *daOps {
		// When Tf specified with font size is 0, it means we should set on our own based on the Rect (autosize).
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
			} else {
				// Disable autosize when font size (>0) explicitly specified.
				autosize = false
			}
			// Skip over (set fontsize in code below).
			continue
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

	// If no text, no appearance needed.
	if len(text) < 1 {
		return nil, nil
	}

	lines := []string{text}

	var tx float64
	tx = 2.0 // Default left margin. // TODO(gunnsth): Add to style options.

	// Handle multi line fields.
	isMultiline := false
	if ftxt.Flags().Has(model.FieldFlagMultiline) {
		isMultiline = true
		text = strings.Replace(text, "\r\n", "\n", -1)
		text = strings.Replace(text, "\r", "\n", -1)
		lines = strings.Split(text, "\n")
	}

	maxLinewidth := 0.0
	textlines := 0
	if encoder != nil {
		l := len(lines)
		i := 0
		for i < l {
			linewidth := 0.0
			lastbreakindex := -1
			lastwidth := 0.0
			for index, r := range lines[i] {
				glyph, has := encoder.RuneToGlyph(r)
				if !has {
					common.Log.Debug("Encoder w/o rune '%c' (%X) - skip", r, r)
					continue
				}
				if glyph == "space" {
					lastbreakindex = index
					lastwidth = linewidth
				}
				metrics, has := font.GetGlyphCharMetrics(glyph)
				if !has {
					common.Log.Debug("Font does not have glyph metrics for %s - skipping", glyph)
					continue
				}
				linewidth += metrics.Wx

				if isMultiline && !autosize && fontsize*linewidth/1000.0 > width && lastbreakindex > 0 {
					part2 := lines[i][lastbreakindex+1:]

					if i < len(lines)-1 {
						lines[i+1] += part2
					} else {
						lines = append(lines, part2)
						l++
					}
					lines[i] = lines[i][0:lastbreakindex]
					linewidth = lastwidth
					break
				}
			}
			if linewidth > maxLinewidth {
				maxLinewidth = linewidth
			}

			lines[i] = string(encoder.Encode(lines[i]))
			if len(lines[i]) > 0 {
				textlines++
			}
			i++
		}
	}

	// Check if text goes out of bounds, if goes out of bounds, then adjust font size until just within bounds.
	if fontsize == 0 || autosize && maxLinewidth > 0 && maxLinewidth*fontsize/1000.0 > width {
		// TODO(gunnsth): Add to style options.
		fontsize = 0.95 * 1000.0 * width / maxLinewidth
	}

	// Account for horizontal alignment (quadding).
	{
		quadding := 0 // Default (left aligned).
		if val, has := core.GetIntVal(ftxt.Q); has {
			quadding = val
		}
		switch quadding {
		case 0: // Left aligned.
		case 1: // Centered.
			remaining := width - maxLinewidth*fontsize/1000.0
			if remaining > 0 {
				tx = remaining / 2
			}
		case 2: // Right justified.
			remaining := width - maxLinewidth*fontsize/1000.0
			tx = remaining
		default:
			common.Log.Debug("ERROR: Unsupported quadding: %d", quadding)
		}
	}

	lineheight := 1.0 * fontsize

	// Vertical alignment.
	ty := 2.0
	{
		textheight := float64(textlines) * lineheight
		if autosize && ty+textheight > height {
			fontsize = 0.95 * (height - ty) / float64(textlines)
			lineheight = 1.0 * fontsize
			textheight = float64(textlines) * lineheight
		}

		if height > textheight {
			if len(lines) > 1 {
				a := (height - textheight) / 2.0
				b := a + textheight - lineheight
				ty = b
			} else {
				ty = (height - textheight) / 2.0
				ty += 1.50 // TODO(gunnsth): Make configurable/part of style parameter.
			}
		}
	}

	cc.Add_Tf(*fontname, fontsize)
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
	xform.SetContentStream(cc.Bytes(), defStreamEncoder())

	apDict := core.MakeDict()
	apDict.Set("N", xform.ToPdfObject())

	return apDict, nil
}

// genFieldTextCombAppearance generates an appearance dictionary for a comb text field where the width is split
// into equal size boxes.
func genFieldTextCombAppearance(wa *model.PdfAnnotationWidget, ftxt *model.PdfFieldText, dr *model.PdfPageResources, style AppearanceStyle) (*core.PdfObjectDictionary, error) {
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
	fontsizeDef := height * style.AutoFontSizeFraction
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

	// TODO(gunnsth): Add to style options.
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

			encoded = string(encoder.Encode(encoded))

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
	xform.SetContentStream(cc.Bytes(), defStreamEncoder())

	apDict := core.MakeDict()
	apDict.Set("N", xform.ToPdfObject())

	return apDict, nil
}

// genFieldCheckboxAppearance generates an appearance dictionary for a widget annotation `wa` referenced by
// a button field `fbtn` with form resources `dr` (DR).
func genFieldCheckboxAppearance(wa *model.PdfAnnotationWidget, fbtn *model.PdfFieldButton, dr *model.PdfPageResources, style AppearanceStyle) (*core.PdfObjectDictionary, error) {
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

	common.Log.Debug("Checkbox, wa BS: %v", wa.BS)

	width := rect[2] - rect[0]
	height := rect[3] - rect[1]

	xformOn := model.NewXObjectForm()
	{
		cc := contentstream.NewContentCreator()
		zapfdb, err := model.NewStandard14Font("ZapfDingbats")
		if err != nil {
			return nil, err
		}

		fontsize := style.AutoFontSizeFraction * height

		checkmetrics, ok := zapfdb.GetGlyphCharMetrics(style.CheckmarkGlyph)
		if !ok {
			return nil, errors.New("glyph not found")
		}
		checkcode, ok := zapfdb.Encoder().GlyphToCharcode(style.CheckmarkGlyph)
		if !ok {
			return nil, errors.New("checkmark glyph - charcode mapping not found")
		}
		checkwidth := checkmetrics.Wx * fontsize / 1000.0
		checkheight := checkwidth
		if checkmetrics.Wy > 0 {
			checkheight = checkmetrics.Wy * fontsize / 1000.0
		}

		tx := 2.0
		ty := 1.0
		if checkwidth < width {
			tx = (width - checkwidth) / 2.0
		}
		if checkheight < height {
			ty = (height - checkheight) / 2.0
		}
		ty += 1.0

		cc.Add_q().
			Add_g(0).
			Add_BT().
			Add_Tf("ZaDb", fontsize).
			Add_Td(tx, ty).
			Add_Tj(*core.MakeString(string(checkcode))).
			Add_ET().
			Add_Q()

		xformOn.Resources = model.NewPdfPageResources()
		xformOn.Resources.SetFontByName("ZaDb", zapfdb.ToPdfObject())
		xformOn.BBox = core.MakeArrayFromFloats([]float64{0, 0, width, height})
		xformOn.SetContentStream(cc.Bytes(), defStreamEncoder())
	}

	xformOff := model.NewXObjectForm()
	{
		cc := contentstream.NewContentCreator()
		cc.Add_re(0, 0, width, height)
		cc.Add_re(0, 0, width-1, height-1)
		xformOff.BBox = core.MakeArrayFromFloats([]float64{0, 0, width, height})
		xformOff.SetContentStream(cc.Bytes(), defStreamEncoder())
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

// defStreamEncoder returns the default stream encoder. Typically FlateEncoder, although RawEncoder
// can be useful for debugging.
func defStreamEncoder() core.StreamEncoder {
	return core.NewFlateEncoder()
}
