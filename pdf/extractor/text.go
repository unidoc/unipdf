/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/contentstream"
	"github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/internal/cmap"
	"github.com/unidoc/unidoc/pdf/model"
)

// ExtractText processes and extracts all text data in content streams and returns as a string. Takes into
// account character encoding via CMaps in the PDF file.
// The text is processed linearly e.g. in the order in which it appears. A best effort is done to add
// spaces and newlines.
func (e *Extractor) ExtractText() (string, error) {
	var buf bytes.Buffer

	cstreamParser := contentstream.NewContentStreamParser(e.contents)
	operations, err := cstreamParser.Parse()
	if err != nil {
		return buf.String(), err
	}

	processor := contentstream.NewContentStreamProcessor(*operations)

	var codemap *cmap.CMap
	inText := false
	xPos, yPos := float64(-1), float64(-1)

	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
			operand := op.Operand
			switch operand {
			case "BT":
				inText = true
			case "ET":
				inText = false
			case "Tf":
				if !inText {
					common.Log.Debug("Tf operand outside text")
					return nil
				}

				if len(op.Params) != 2 {
					common.Log.Debug("Error Tf should only get 2 input params, got %d", len(op.Params))
					return errors.New("Incorrect parameter count")
				}

				codemap = nil

				fontName, ok := op.Params[0].(*core.PdfObjectName)
				if !ok {
					common.Log.Debug("Error Tf font input not a name")
					return errors.New("Tf range error")
				}

				if resources == nil {
					return nil
				}

				fontObj, found := resources.GetFontByName(*fontName)
				if !found {
					common.Log.Debug("Font not found...")
					return errors.New("Font not in resources")
				}

				fontObj = core.TraceToDirectObject(fontObj)
				if fontDict, isDict := fontObj.(*core.PdfObjectDictionary); isDict {
					toUnicode := fontDict.Get("ToUnicode")
					if toUnicode != nil {
						toUnicode = core.TraceToDirectObject(toUnicode)
						toUnicodeStream, ok := toUnicode.(*core.PdfObjectStream)
						if !ok {
							return errors.New("Invalid ToUnicode entry - not a stream")
						}
						decoded, err := core.DecodeStream(toUnicodeStream)
						if err != nil {
							return err
						}

						codemap, err = cmap.LoadCmapFromData(decoded)
						if err != nil {
							return err
						}
					}
				}
			case "T*":
				if !inText {
					common.Log.Debug("T* operand outside text")
					return nil
				}
				buf.WriteString("\n")
			case "Td", "TD":
				if !inText {
					common.Log.Debug("Td/TD operand outside text")
					return nil
				}

				// Params: [tx ty], corresponeds to Tm=Tlm=[1 0 0;0 1 0;tx ty 1]*Tm
				if len(op.Params) != 2 {
					common.Log.Debug("Td/TD invalid arguments")
					return nil
				}
				tx, err := getNumberAsFloat(op.Params[0])
				if err != nil {
					common.Log.Debug("Td Float parse error")
					return nil
				}
				ty, err := getNumberAsFloat(op.Params[1])
				if err != nil {
					common.Log.Debug("Td Float parse error")
					return nil
				}

				if tx > 0 {
					buf.WriteString(" ")
				}
				if ty < 0 {
					// TODO: More flexible space characters?
					buf.WriteString("\n")
				}
			case "Tm":
				if !inText {
					common.Log.Debug("Tm operand outside text")
					return nil
				}

				// Params: a,b,c,d,e,f as in Tm = [a b 0; c d 0; e f 1].
				// The last two (e,f) represent translation.
				if len(op.Params) != 6 {
					return errors.New("Tm: Invalid number of inputs")
				}
				xfloat, ok := op.Params[4].(*core.PdfObjectFloat)
				if !ok {
					xint, ok := op.Params[4].(*core.PdfObjectInteger)
					if !ok {
						return nil
					}
					xfloat = core.MakeFloat(float64(*xint))
				}
				yfloat, ok := op.Params[5].(*core.PdfObjectFloat)
				if !ok {
					yint, ok := op.Params[5].(*core.PdfObjectInteger)
					if !ok {
						return nil
					}
					yfloat = core.MakeFloat(float64(*yint))
				}
				if yPos == -1 {
					yPos = float64(*yfloat)
				} else if yPos > float64(*yfloat) {
					buf.WriteString("\n")
					xPos = float64(*xfloat)
					yPos = float64(*yfloat)
					return nil
				}
				if xPos == -1 {
					xPos = float64(*xfloat)
				} else if xPos < float64(*xfloat) {
					buf.WriteString("\t")
					xPos = float64(*xfloat)
				}
			case "TJ":
				if !inText {
					common.Log.Debug("TJ operand outside text")
					return nil
				}
				if len(op.Params) < 1 {
					return nil
				}
				paramList, ok := op.Params[0].(*core.PdfObjectArray)
				if !ok {
					return fmt.Errorf("Invalid parameter type, no array (%T)", op.Params[0])
				}
				for _, obj := range paramList.Elements() {
					switch v := obj.(type) {
					case *core.PdfObjectString:
						if codemap != nil {
							buf.WriteString(codemap.CharcodeBytesToUnicode(v.Bytes()))
						} else {
							buf.WriteString(v.Str())
						}
					case *core.PdfObjectFloat:
						if *v < -100 {
							buf.WriteString(" ")
						}
					case *core.PdfObjectInteger:
						if *v < -100 {
							buf.WriteString(" ")
						}
					}
				}
			case "Tj":
				if !inText {
					common.Log.Debug("Tj operand outside text")
					return nil
				}
				if len(op.Params) < 1 {
					return nil
				}
				param, ok := op.Params[0].(*core.PdfObjectString)
				if !ok {
					return fmt.Errorf("Invalid parameter type, not string (%T)", op.Params[0])
				}
				if codemap != nil {
					buf.WriteString(codemap.CharcodeBytesToUnicode(param.Bytes()))
				} else {
					buf.WriteString(param.Str())
				}
			}

			return nil
		})

	err = processor.Process(e.resources)
	if err != nil {
		common.Log.Error("Error processing: %v", err)
		return buf.String(), err
	}

	procBuf(&buf)

	return buf.String(), nil
}
