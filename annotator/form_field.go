/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package annotator

import (
	"bytes"
	"errors"

	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
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

	zapfdb, err := model.NewStandard14Font(model.ZapfDingbatsName)
	if err != nil {
		return nil, err
	}

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

// SignatureLine represents a line of information in the signature field appearance.
type SignatureLine struct {
	Desc string
	Text string
}

// NewSignatureLine returns a new signature line displayed as a part of the
// signature field appearance.
func NewSignatureLine(desc, text string) *SignatureLine {
	return &SignatureLine{
		Desc: desc,
		Text: text,
	}
}

// SignatureFieldOpts represents a set of options used to configure
// an appearance widget dictionary.
type SignatureFieldOpts struct {
	// Rect represents the area the signature annotation is displayed on.
	Rect []float64

	// AutoSize specifies if the content of the appearance should be
	// scaled to fit in the annotation rectangle.
	AutoSize bool

	// Font specifies the font of the text content.
	Font *model.PdfFont

	// FontSize specifies the size of the text content.
	FontSize float64

	// LineHeight specifies the height of a line of text in the appearance annotation.
	LineHeight float64

	// TextColor represents the color of the text content displayed.
	TextColor model.PdfColor

	// FillColor represents the background color of the appearance annotation area.
	FillColor model.PdfColor

	// BorderSize represents border size of the appearance annotation area.
	BorderSize float64

	// BorderColor represents the border color of the appearance annotation area.
	BorderColor model.PdfColor
}

// NewSignatureFieldOpts returns a new initialized instance of options
// used to generate a signature appearance.
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

// NewSignatureField returns a new signature field with a visible appearance
// containing the specified signature lines and styled according to the
// specified options.
func NewSignatureField(signature *model.PdfSignature, lines []*SignatureLine, opts *SignatureFieldOpts) (*model.PdfFieldSignature, error) {
	if signature == nil {
		return nil, errors.New("signature cannot be nil")
	}

	apDict, err := genFieldSignatureAppearance(lines, opts)
	if err != nil {
		return nil, err
	}

	field := model.NewPdfFieldSignature(signature)
	field.Rect = core.MakeArrayFromFloats(opts.Rect)
	field.AP = apDict
	return field, nil
}
