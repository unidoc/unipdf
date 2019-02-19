/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package annotator

import (
	"bytes"
	"errors"
	"math"
	"unicode"

	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

// TextFieldOptions defines optional parameter for a text field in a form.
type TextFieldOptions struct {
	MaxLen int    // Ignored if <= 0.
	Value  string // Ignored if empty ("").
}

// NewTextField generates a new text field with partial name `name` at location
// specified by `rect` on given `page` and with field specific options `opt`.
func NewTextField(page *model.PdfPage, name string, rect []float64, opt TextFieldOptions) (*model.PdfFieldText, error) {
	if page == nil {
		return nil, errors.New("page not specified")
	}
	if len(name) <= 0 {
		return nil, errors.New("required attribute not specified")
	}
	if len(rect) != 4 {
		return nil, errors.New("invalid range")
	}

	field := model.NewPdfField()
	textfield := &model.PdfFieldText{}
	field.SetContext(textfield)
	textfield.PdfField = field

	textfield.T = core.MakeString(name)

	if opt.MaxLen > 0 {
		textfield.MaxLen = core.MakeInteger(int64(opt.MaxLen))
	}
	if len(opt.Value) > 0 {
		textfield.V = core.MakeString(opt.Value)
	}

	widget := model.NewPdfAnnotationWidget()
	widget.Rect = core.MakeArrayFromFloats(rect) //[]float64{144.0, 595.89, 294.0, 617.9})
	widget.P = page.ToPdfObject()
	widget.F = core.MakeInteger(4) // 4 (100 -> Print/show annotations).
	widget.Parent = textfield.ToPdfObject()

	textfield.Annotations = append(textfield.Annotations, widget)

	return textfield, nil
}

// CheckboxFieldOptions defines optional parameters for a checkbox field a form.
type CheckboxFieldOptions struct {
	Checked bool
}

// NewCheckboxField generates a new checkbox field with partial name `name` at location `rect`
// on specified `page` and with field specific options `opt`.
func NewCheckboxField(page *model.PdfPage, name string, rect []float64, opt CheckboxFieldOptions) (*model.PdfFieldButton, error) {
	if page == nil {
		return nil, errors.New("page not specified")
	}
	if len(name) <= 0 {
		return nil, errors.New("required attribute not specified")
	}
	if len(rect) != 4 {
		return nil, errors.New("invalid range")
	}

	zapfdb := fonts.NewFontZapfDingbats()

	field := model.NewPdfField()
	buttonfield := &model.PdfFieldButton{}
	field.SetContext(buttonfield)
	buttonfield.PdfField = field

	buttonfield.T = core.MakeString(name)
	buttonfield.SetType(model.ButtonTypeCheckbox)

	state := "Off"
	if opt.Checked {
		state = "Yes"
	}

	buttonfield.V = core.MakeName(state)

	widget := model.NewPdfAnnotationWidget()
	widget.Rect = core.MakeArrayFromFloats(rect)
	widget.P = page.ToPdfObject()
	widget.F = core.MakeInteger(4)
	widget.Parent = buttonfield.ToPdfObject()

	w := rect[2] - rect[0]
	h := rect[3] - rect[1]

	// Off state.
	var cs bytes.Buffer
	cs.WriteString("q\n")
	cs.WriteString("0 0 1 rg\n")
	cs.WriteString("BT\n")
	cs.WriteString("/ZaDb 12 Tf\n")
	cs.WriteString("ET\n")
	cs.WriteString("Q\n")

	cc := contentstream.NewContentCreator()
	cc.Add_q()
	cc.Add_rg(0, 0, 1)
	cc.Add_BT()
	cc.Add_Tf(*core.MakeName("ZaDb"), 12)
	cc.Add_Td(0, 0)
	cc.Add_ET()
	cc.Add_Q()

	xformOff := model.NewXObjectForm()
	xformOff.SetContentStream(cc.Bytes(), core.NewRawEncoder())
	xformOff.BBox = core.MakeArrayFromFloats([]float64{0, 0, w, h})
	xformOff.Resources = model.NewPdfPageResources()
	xformOff.Resources.SetFontByName("ZaDb", zapfdb.ToPdfObject())

	// On state (Yes).
	cc = contentstream.NewContentCreator()
	cc.Add_q()
	cc.Add_re(0, 0, w, h)
	cc.Add_W().Add_n()
	cc.Add_rg(0, 0, 1)
	cc.Translate(0, 3.0)
	cc.Add_BT()
	cc.Add_Tf(*core.MakeName("ZaDb"), 12)
	cc.Add_Td(0, 0)
	cc.Add_Tj(*core.MakeString("\064"))
	cc.Add_ET()
	cc.Add_Q()

	xformOn := model.NewXObjectForm()
	xformOn.SetContentStream(cc.Bytes(), core.NewRawEncoder())
	xformOn.BBox = core.MakeArrayFromFloats([]float64{0, 0, w, h})
	xformOn.Resources = model.NewPdfPageResources()
	xformOn.Resources.SetFontByName("ZaDb", zapfdb.ToPdfObject())

	dchoiceapp := core.MakeDict()
	dchoiceapp.Set("Off", xformOff.ToPdfObject())
	dchoiceapp.Set("Yes", xformOn.ToPdfObject())

	appearance := core.MakeDict()
	appearance.Set("N", dchoiceapp)

	widget.AP = appearance
	widget.AS = core.MakeName(state)

	buttonfield.Annotations = append(buttonfield.Annotations, widget)

	return buttonfield, nil
}

// ComboboxFieldOptions defines optional parameters for a combobox form field.
type ComboboxFieldOptions struct {
	// Choices is the list of string values that can be selected.
	Choices []string
}

// NewComboboxField generates a new combobox form field with partial name `name` at location `rect`
// on specified `page` and with field specific options `opt`.
func NewComboboxField(page *model.PdfPage, name string, rect []float64, opt ComboboxFieldOptions) (*model.PdfFieldChoice, error) {
	if page == nil {
		return nil, errors.New("page not specified")
	}
	if len(name) <= 0 {
		return nil, errors.New("required attribute not specified")
	}
	if len(rect) != 4 {
		return nil, errors.New("invalid range")
	}

	field := model.NewPdfField()
	chfield := &model.PdfFieldChoice{}
	field.SetContext(chfield)
	chfield.PdfField = field

	chfield.T = core.MakeString(name)
	chfield.Opt = core.MakeArray()
	for _, choicestr := range opt.Choices {
		chfield.Opt.Append(core.MakeString(choicestr))
	}
	chfield.SetFlag(model.FieldFlagCombo)

	widget := model.NewPdfAnnotationWidget()
	widget.Rect = core.MakeArrayFromFloats(rect)
	widget.P = page.ToPdfObject()
	widget.F = core.MakeInteger(4) // TODO: Make flags for these values and a way to set.
	widget.Parent = chfield.ToPdfObject()

	chfield.Annotations = append(chfield.Annotations, widget)

	return chfield, nil
}

type SignatureLine struct {
	Desc string
	Text string
}

func NewSignatureLine(desc, text string) *SignatureLine {
	return &SignatureLine{
		Desc: desc,
		Text: text,
	}
}

type SignatureFieldOpts struct {
	Rect     []float64
	AutoSize bool

	Font       *model.PdfFont
	FontSize   float64
	LineHeight float64
	TextColor  model.PdfColor

	FillColor   model.PdfColor
	BorderSize  float64
	BorderColor model.PdfColor
}

func NewSignatureFieldOpts() *SignatureFieldOpts {
	return &SignatureFieldOpts{
		Font:        model.DefaultFont(),
		FontSize:    10,
		LineHeight:  1,
		AutoSize:    true,
		TextColor:   model.NewPdfColorDeviceGray(0),
		BorderColor: model.NewPdfColorDeviceGray(0),
		FillColor:   model.NewPdfColorDeviceGray(1),
	}
}

func NewSignatureField(signature *model.PdfSignature, fields []*SignatureLine, opts *SignatureFieldOpts) (*model.PdfFieldSignature, error) {
	if signature == nil {
		return nil, errors.New("signature cannot be nil")
	}

	field := model.NewPdfFieldSignature(signature)
	field.T = core.MakeString("")

	apDict, err := genFieldSignatureAppearance(fields, opts)
	if err != nil {
		return nil, err
	}

	widget := model.NewPdfAnnotationWidget()
	widget.Rect = core.MakeArrayFromFloats(opts.Rect)
	widget.F = core.MakeInteger(4)
	widget.Parent = field.ToPdfObject()
	widget.AP = apDict

	field.Annotations = append(field.Annotations, widget)
	return field, nil
}

func genFieldSignatureAppearance(fields []*SignatureLine, opts *SignatureFieldOpts) (*core.PdfObjectDictionary, error) {
	if opts == nil {
		opts = NewSignatureFieldOpts()
	}

	// Get font.
	var err error
	var fontName *core.PdfObjectName
	font := opts.Font

	if font != nil {
		descriptor, _ := font.GetFontDescriptor()
		if descriptor != nil {
			if f, ok := descriptor.FontName.(*core.PdfObjectName); ok {
				fontName = f
			}
		}
		if fontName == nil {
			fontName = core.MakeName("Font1")
		}
	} else {
		if font, err = model.NewStandard14Font("Helvetica"); err != nil {
			return nil, err
		}
		fontName = core.MakeName("Helv")
	}

	// Get font size and line height.
	fontSize := opts.FontSize
	if fontSize <= 0 {
		fontSize = 10
	}

	if opts.LineHeight <= 0 {
		opts.LineHeight = 1
	}
	lineHeight := opts.LineHeight * fontSize

	// Get space character width.
	spaceMetrics, found := font.GetRuneMetrics(' ')
	if !found {
		return nil, errors.New("the font does not have a space glyph")
	}
	spaceWidth := spaceMetrics.Wx

	// Generate lines.
	var maxLineWidth float64
	var lines []string

	for _, field := range fields {
		if field.Text == "" {
			continue
		}

		line := field.Text
		if field.Desc != "" {
			line = field.Desc + ": " + line
		}
		lines = append(lines, line)

		var lineWidth float64
		for _, r := range line {
			metrics, has := font.GetRuneMetrics(r)
			if !has {
				continue
			}

			lineWidth += metrics.Wx
		}

		if lineWidth > maxLineWidth {
			maxLineWidth = lineWidth
		}
	}

	maxLineWidth = maxLineWidth * fontSize / 1000.0
	height := float64(len(lines)) * lineHeight

	// Calculate annotation rectangle.
	rect := opts.Rect
	if rect == nil {
		rect = []float64{0, 0, maxLineWidth, height}
		opts.Rect = rect
	}
	rectWidth := rect[2] - rect[0]
	rectHeight := rect[3] - rect[1]

	// Fit contents
	var offsetY float64
	if opts.AutoSize {
		if maxLineWidth > rectWidth || height > rectHeight {
			scale := math.Min(rectWidth/maxLineWidth, rectHeight/height)
			fontSize *= scale
		}

		lineHeight = opts.LineHeight * fontSize
		offsetY += (rectHeight - float64(len(lines))*lineHeight) / 2
	}

	// Draw signature.
	cc := contentstream.NewContentCreator()

	if opts.BorderSize > 0 {
		cc.Add_q().
			Add_re(rect[0], rect[1], rectWidth, rectHeight).
			Add_w(opts.BorderSize).
			SetStrokingColor(opts.BorderColor).
			SetNonStrokingColor(opts.FillColor).
			Add_B().
			Add_Q()
	}

	cc.Add_q()
	cc.Translate(rect[0], rect[3]-lineHeight-offsetY)
	cc.Add_BT()

	encoder := font.Encoder()
	for _, line := range lines {
		var encStr []byte
		for _, r := range line {
			if unicode.IsSpace(r) {
				if len(encStr) > 0 {
					cc.SetNonStrokingColor(opts.TextColor).
						Add_Tf(*fontName, fontSize).
						Add_TL(lineHeight).
						Add_TJ([]core.PdfObject{core.MakeStringFromBytes(encStr)}...)
					encStr = nil
				}

				cc.Add_Tf(*fontName, fontSize).
					Add_TL(lineHeight).
					Add_TJ([]core.PdfObject{core.MakeFloat(-spaceWidth)}...)
			} else {
				encStr = append(encStr, encoder.Encode(string(r))...)
			}
		}

		if len(encStr) > 0 {
			cc.SetNonStrokingColor(opts.TextColor).
				Add_Tf(*fontName, fontSize).
				Add_TL(lineHeight).
				Add_TJ([]core.PdfObject{core.MakeStringFromBytes(encStr)}...)
		}

		cc.Add_Td(0, -lineHeight)
	}

	cc.Add_ET()
	cc.Add_Q()

	// Create appearance dictionary.
	resources := model.NewPdfPageResources()
	resources.SetFontByName(*fontName, font.ToPdfObject())

	xform := model.NewXObjectForm()
	xform.Resources = resources
	xform.BBox = core.MakeArrayFromFloats(rect)
	xform.SetContentStream(cc.Bytes(), defStreamEncoder())

	apDict := core.MakeDict()
	apDict.Set("N", xform.ToPdfObject())
	return apDict, nil
}
