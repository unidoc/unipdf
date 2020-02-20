/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package render

import (
	"errors"

	"github.com/adrg/sysfont"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render/context"

	"github.com/unidoc/unipdf/v3/internal/transform"
)

var (
	errType  = errors.New("type check error")
	errRange = errors.New("range check error")
)

type renderer struct {
}

func (r renderer) renderPage(ctx context.Context, page *model.PdfPage) error {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		return err
	}

	// Change coordinate system.
	ctx.Translate(0, float64(ctx.Height()))
	ctx.Scale(1, -1)

	// Create white background.
	ctx.Push()
	ctx.SetRGBA(1, 1, 1, 1)
	ctx.DrawRectangle(0, 0, float64(ctx.Width()), float64(ctx.Height()))
	ctx.Fill()
	ctx.Pop()

	// Set defaults.
	ctx.SetLineWidth(1.0)
	ctx.SetRGBA(0, 0, 0, 1)

	return r.renderContentStream(ctx, contents, page.Resources)
}

func (r renderer) renderContentStream(ctx context.Context, contents string, resources *model.PdfPageResources) error {
	operations, err := contentstream.NewContentStreamParser(contents).Parse()
	if err != nil {
		return err
	}

	textState := ctx.TextState()
	fontCache := map[string]*context.TextFont{}
	fontFinder := sysfont.NewFinder(&sysfont.FinderOpts{
		Extensions: []string{".ttf", ".ttc"},
	})

	processor := contentstream.NewContentStreamProcessor(*operations)
	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
			common.Log.Debug("Processing %s", op.Operand)
			switch op.Operand {
			//
			// Graphics stage operators
			//

			// Push current graphics state to the stack.
			case "q":
				ctx.Push()
			// Pop graphics state from the stack.
			case "Q":
				ctx.Pop()
			// Modify graphics state matrix.
			case "cm":
				if len(op.Params) != 6 {
					return errRange
				}

				fv, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				m := transform.NewMatrix(fv[0], fv[1], fv[2], fv[3], fv[4], fv[5])
				common.Log.Debug("Graphics state matrix: %+v\n", m)
				ctx.SetMatrix(ctx.Matrix().Mult(m))

				// TODO: Take angle into account for line widths (8.4.3.2 Line Width).
				s := (gs.CTM.ScalingFactorX() + gs.CTM.ScalingFactorY()) / 2.0
				ctx.SetLineWidth(s * ctx.LineWidth())
			// Set line width.
			case "w":
				if len(op.Params) != 1 {
					return errRange
				}

				fw, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				// TODO: Take angle into account for line widths (8.4.3.2 Line Width).
				s := (gs.CTM.ScalingFactorX() + gs.CTM.ScalingFactorY()) / 2.0
				ctx.SetLineWidth(s * fw[0])
			// Set line cap style.
			case "J":
				if len(op.Params) != 1 {
					return errRange
				}

				val, ok := core.GetIntVal(op.Params[0])
				if !ok {
					return errType
				}

				switch val {
				// Butt cap.
				case 0:
					ctx.SetLineCap(context.LineCapButt)
				// Round cap.
				case 1:
					ctx.SetLineCap(context.LineCapRound)
				// Projecting square cap.
				case 2:
					ctx.SetLineCap(context.LineCapSquare)
				default:
					common.Log.Debug("Invalid line cap style: %d", val)
					return errRange
				}
			// Set line join style.
			case "j":
				if len(op.Params) != 1 {
					return errRange
				}

				val, ok := core.GetIntVal(op.Params[0])
				if !ok {
					return errType
				}

				switch val {
				// Miter join.
				case 0:
					ctx.SetLineJoin(context.LineJoinBevel)
				// Round join.
				case 1:
					ctx.SetLineJoin(context.LineJoinRound)
				// Bevel join.
				case 2:
					ctx.SetLineJoin(context.LineJoinBevel)
				default:
					common.Log.Debug("Invalid line join style: %d", val)
					return errRange
				}
			// Set miter limit.
			case "M":
				if len(op.Params) != 1 {
					return errRange
				}

				fw, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				// TODO: Add miter support in context.
				// ctx.SetMiterLimit(fw[0])
				_ = fw
				common.Log.Debug("Miter limit not supported")
			// Set line dash pattern.
			case "d":
				if len(op.Params) != 2 {
					return errRange
				}

				dashArray, ok := core.GetArray(op.Params[0])
				if !ok {
					return errType
				}

				phase, ok := core.GetIntVal(op.Params[1])
				if !ok {
					return errType
				}

				dashes, err := core.GetNumbersAsFloat(dashArray.Elements())
				if err != nil {
					return err
				}
				ctx.SetDash(dashes...)

				// TODO: Add support for dash phase in context.
				//ctx.SetDashPhase(phase)
				_ = phase
				common.Log.Debug("Line dash phase not supported")
			// Set color rendering intent.
			case "ri":
				// TODO: Add rendering intent support.
				common.Log.Debug("Rendering intent not supported")
			// Set flatness tolerance.
			case "i":
				// TODO: Add flatness tolerance support.
				common.Log.Debug("Flatness tolerance not supported")
			// Set graphics state from dictionary.
			case "gs":
				if len(op.Params) != 1 {
					return errRange
				}

				rname, ok := core.GetName(op.Params[0])
				if !ok {
					return errType
				}
				if rname == nil {
					return errRange
				}

				extobj, ok := resources.GetExtGState(*rname)
				if !ok {
					common.Log.Debug("ERROR: could not find resource: %s", *rname)
					return errors.New("resource not found")
				}

				extdict, ok := core.GetDict(extobj)
				if !ok {
					common.Log.Debug("ERROR: could get graphics state dict")
					return errType
				}
				common.Log.Debug("GS dict: %s", extdict.String())

			//
			// Path operators
			//

			// Move to.
			case "m":
				if len(op.Params) != 2 {
					common.Log.Debug("WARN: error while processing `m` operator: %s. Output may be incorrect.", errRange)
					return nil
				}

				xy, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				ctx.NewSubPath()
				ctx.MoveTo(xy[0], xy[1])
			// Line to.
			case "l":
				if len(op.Params) != 2 {
					common.Log.Debug("WARN: error while processing `l` operator: %s. Output may be incorrect.", errRange)
					return nil
				}

				xy, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				ctx.LineTo(xy[0], xy[1])
			// Cubic bezier.
			case "c":
				if len(op.Params) != 6 {
					return errRange
				}

				cbp, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				common.Log.Debug("Cubic bezier params: %+v", cbp)
				ctx.CubicTo(cbp[0], cbp[1], cbp[2], cbp[3], cbp[4], cbp[5])
			// Cubic bezier.
			case "v":
				if len(op.Params) != 4 {
					return errRange
				}

				cbp, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				common.Log.Debug("Cubic bezier params: %+v", cbp)
				ctx.CubicTo(0, 0, cbp[0], cbp[1], cbp[2], cbp[3])
			// Close current subpath.
			case "h":
				ctx.ClosePath()
				ctx.NewSubPath()
			// Rectangle.
			case "re":
				if len(op.Params) != 4 {
					return errRange
				}

				xywh, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				ctx.DrawRectangle(xywh[0], xywh[1], xywh[2], xywh[3])
				ctx.NewSubPath()

			//
			// Path painting operators
			//

			// Set path stroke.
			case "S":
				color, err := gs.ColorspaceStroking.ColorToRGB(gs.ColorStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color")
					return err
				}

				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.Stroke()
			// Close and stroke.
			case "s":
				color, err := gs.ColorspaceStroking.ColorToRGB(gs.ColorStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color")
					return err
				}

				ctx.ClosePath()
				ctx.NewSubPath()
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.Stroke()
			// Fill path using non-zero winding number rule.
			case "f", "F":
				color, err := gs.ColorspaceNonStroking.ColorToRGB(gs.ColorNonStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color")
					return err
				}

				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.SetFillRule(context.FillRuleWinding)
				ctx.Fill()
			// Fill path using even-odd rule.
			case "f*":
				color, err := gs.ColorspaceNonStroking.ColorToRGB(gs.ColorNonStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color")
					return err
				}

				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.SetFillRule(context.FillRuleEvenOdd)
				ctx.Fill()
			// Fill then stroke the path using non-zero winding rule.
			case "B":
				// Fill path.
				color, err := gs.ColorspaceNonStroking.ColorToRGB(gs.ColorNonStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor := color.(*model.PdfColorDeviceRGB)
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.SetFillRule(context.FillRuleWinding)
				ctx.FillPreserve()

				// Stroke path.
				color, err = gs.ColorspaceStroking.ColorToRGB(gs.ColorStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor = color.(*model.PdfColorDeviceRGB)
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.Stroke()
			// Fill then stroke the path using even-odd rule.
			case "B*":
				// Fill path.
				color, err := gs.ColorspaceNonStroking.ColorToRGB(gs.ColorNonStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor := color.(*model.PdfColorDeviceRGB)
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.SetFillRule(context.FillRuleEvenOdd)
				ctx.FillPreserve()

				// Stroke path.
				color, err = gs.ColorspaceStroking.ColorToRGB(gs.ColorStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor = color.(*model.PdfColorDeviceRGB)
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.Stroke()
			// Close, fill and stroke the path using non-zero winding rule.
			case "b":
				// Fill path.
				color, err := gs.ColorspaceNonStroking.ColorToRGB(gs.ColorNonStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor := color.(*model.PdfColorDeviceRGB)
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.ClosePath()
				ctx.NewSubPath()
				ctx.SetFillRule(context.FillRuleWinding)
				ctx.FillPreserve()

				// Stroke path.
				color, err = gs.ColorspaceStroking.ColorToRGB(gs.ColorStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor = color.(*model.PdfColorDeviceRGB)
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.Stroke()
			// Close, fill and stroke the path using even-odd rule.
			case "b*":
				// Close current subpath.
				ctx.ClosePath()

				// Fill path.
				color, err := gs.ColorspaceNonStroking.ColorToRGB(gs.ColorNonStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor := color.(*model.PdfColorDeviceRGB)
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.NewSubPath()
				ctx.SetFillRule(context.FillRuleEvenOdd)
				ctx.FillPreserve()

				// Stroke path.
				color, err = gs.ColorspaceStroking.ColorToRGB(gs.ColorStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", err)
					return err
				}

				rgbColor = color.(*model.PdfColorDeviceRGB)
				ctx.SetRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
				ctx.Stroke()
			// End the current path without filling or stroking.
			case "n":
				ctx.ClearPath()

			//
			// Path clipping operators
			//

			// Modify current clipping path using non-zero winding rule.
			case "W":
				// TODO: fix clipping.
				//ctx.StrokePreserve()
				//ctx.Clip()
				ctx.SetFillRule(context.FillRuleWinding)
				ctx.ClipPreserve()
			// Modify current clipping path using even-odd rule.
			case "W*":
				// TODO: fix clipping.
				//ctx.StrokePreserve()
				//ctx.Clip()
				ctx.SetFillRule(context.FillRuleEvenOdd)
				ctx.ClipPreserve()

			//
			// Color operators
			//

			// Set RGB non-stroking color.
			case "rg":
				rgbColor, ok := gs.ColorNonStroking.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color: %v", gs.ColorNonStroking)
					return nil
				}
				ctx.SetFillRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
			// Set RGB stroking color.
			case "RG":
				rgbColor, ok := gs.ColorStroking.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color: %v", gs.ColorStroking)
					return nil
				}
				ctx.SetStrokeRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
			// Set CMYK non-stroking color.
			case "k":
				cmykColor, ok := gs.ColorNonStroking.(*model.PdfColorDeviceCMYK)
				if !ok {
					common.Log.Debug("Error converting color: %v", gs.ColorNonStroking)
					return nil
				}
				color, err := gs.ColorspaceNonStroking.ColorToRGB(cmykColor)
				if err != nil {
					common.Log.Debug("Error converting color: %v", gs.ColorNonStroking)
					return nil
				}
				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color: %v", color)
					return nil
				}
				ctx.SetFillRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
			// Set CMYK stroking color.
			case "K":
				cmykColor, ok := gs.ColorStroking.(*model.PdfColorDeviceCMYK)
				if !ok {
					common.Log.Debug("Error converting color: %v", gs.ColorStroking)
					return nil
				}
				color, err := gs.ColorspaceStroking.ColorToRGB(cmykColor)
				if err != nil {
					common.Log.Debug("Error converting color: %v", gs.ColorStroking)
					return nil
				}
				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color: %v", color)
					return nil
				}
				ctx.SetStrokeRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
			// Set Grayscale non-stroking color.
			case "g":
				grayColor, ok := gs.ColorNonStroking.(*model.PdfColorDeviceGray)
				if !ok {
					common.Log.Debug("Error converting color: %v", gs.ColorNonStroking)
					return nil
				}
				color, err := gs.ColorspaceNonStroking.ColorToRGB(grayColor)
				if err != nil {
					common.Log.Debug("Error converting color: %v", gs.ColorNonStroking)
					return nil
				}
				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color: %v", color)
					return nil
				}
				ctx.SetFillRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
			// Set Grayscale stroking color.
			case "G":
				grayColor, ok := gs.ColorStroking.(*model.PdfColorDeviceGray)
				if !ok {
					common.Log.Debug("Error converting color: %v", gs.ColorStroking)
					return nil
				}
				color, err := gs.ColorspaceStroking.ColorToRGB(grayColor)
				if err != nil {
					common.Log.Debug("Error converting color: %v", gs.ColorStroking)
					return nil
				}
				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color: %v", color)
					return nil
				}
				ctx.SetStrokeRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
			case "cs", "sc", "scn":
				color, err := gs.ColorspaceNonStroking.ColorToRGB(gs.ColorNonStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", gs.ColorNonStroking)
					return nil
				}
				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color: %v", color)
					return nil
				}
				ctx.SetFillRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)
			case "CS", "SC", "SCN":
				color, err := gs.ColorspaceStroking.ColorToRGB(gs.ColorStroking)
				if err != nil {
					common.Log.Debug("Error converting color: %v", gs.ColorStroking)
					return nil
				}
				rgbColor, ok := color.(*model.PdfColorDeviceRGB)
				if !ok {
					common.Log.Debug("Error converting color: %v", color)
					return nil
				}
				ctx.SetStrokeRGBA(rgbColor.R(), rgbColor.G(), rgbColor.B(), 1)

			//
			// Image operators
			//

			// Display xobjects.
			case "Do":
				if len(op.Params) != 1 {
					return errRange
				}

				name, ok := core.GetName(op.Params[0])
				if !ok {
					return errType
				}

				_, xtype := resources.GetXObjectByName(*name)
				switch xtype {
				case model.XObjectTypeImage:
					common.Log.Debug("XObject image: %s", name.String())

					ximg, err := resources.GetXObjectImageByName(*name)
					if err != nil {
						return err
					}

					img, err := ximg.ToImage()
					if err != nil {
						return err
					}

					goImg, err := img.ToGoImage()
					if err != nil {
						return err
					}
					bounds := goImg.Bounds()

					// TODO: Handle soft masks.
					ctx.Push()
					ctx.Scale(1.0/float64(bounds.Dx()), -1.0/float64(bounds.Dy()))
					ctx.DrawImageAnchored(goImg, 0, 0, 0, 1)
					ctx.Pop()
				case model.XObjectTypeForm:
					common.Log.Debug("XObject form: %s", name.String())

					// Go through the XObject Form content stream.
					xform, err := resources.GetXObjectFormByName(*name)
					if err != nil {
						return err
					}

					formContent, err := xform.GetContentStream()
					if err != nil {
						return err
					}

					formResources := xform.Resources
					if formResources == nil {
						formResources = resources
					}

					ctx.Push()
					if xform.Matrix != nil {
						array, ok := core.GetArray(xform.Matrix)
						if !ok {
							return errType
						}

						mf, err := core.GetNumbersAsFloat(array.Elements())
						if err != nil {
							return err
						}
						if len(mf) != 6 {
							return errRange
						}

						m := transform.NewMatrix(mf[0], mf[1], mf[2], mf[3], mf[4], mf[5])
						ctx.SetMatrix(ctx.Matrix().Mult(m))
					}

					if xform.BBox != nil {
						array, ok := core.GetArray(xform.BBox)
						if !ok {
							return errType
						}

						bf, err := core.GetNumbersAsFloat(array.Elements())
						if err != nil {
							return err
						}
						if len(bf) != 4 {
							common.Log.Debug("Len = %d", len(bf))
							return errRange
						}

						// Set clipping region.
						ctx.DrawRectangle(bf[0], bf[1], bf[2]-bf[0], bf[3]-bf[1])
						ctx.SetRGBA(1, 0, 0, 1)
						ctx.Clip()
					} else {
						common.Log.Debug("ERROR: Required BBox missing on XObject Form")
					}

					// Process the content stream in the Form object.
					err = r.renderContentStream(ctx, string(formContent), formResources)
					if err != nil {
						return err
					}
					ctx.Pop()
				}
			// Display inline image.
			case "BI":
				if len(op.Params) != 1 {
					return errRange
				}

				iimg, ok := op.Params[0].(*contentstream.ContentStreamInlineImage)
				if !ok {
					return nil
				}

				img, err := iimg.ToImage(resources)
				if err != nil {
					return err
				}

				goImg, err := img.ToGoImage()
				if err != nil {
					return err
				}
				bounds := goImg.Bounds()

				ctx.Push()
				ctx.Scale(1.0/float64(bounds.Dx()), -1.0/float64(bounds.Dy()))
				ctx.DrawImageAnchored(goImg, 0, 0, 0, 1)
				ctx.Pop()

			//
			// Text operators
			//

			// Begin text.
			case "BT":
				textState.Reset()
			// End text.
			case "ET":
				textState.Reset()
			// Set text leading.
			case "TL":
				if len(op.Params) != 1 {
					return errRange
				}

				tl, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}

				textState.Tl = tl
			// Set character spacing.
			case "Tc":
				if len(op.Params) != 1 {
					return errRange
				}

				tc, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}

				textState.Tc = tc
			// Set word spacing.
			case "Tw":
				if len(op.Params) != 1 {
					return errRange
				}

				tw, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}

				textState.Tw = tw
			// Set horizontal scaling.
			case "Tz":
				if len(op.Params) != 1 {
					return errRange
				}

				th, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}

				textState.Th = th
			// Set text rise.
			case "Ts":
				if len(op.Params) != 1 {
					return errRange
				}

				ts, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}

				textState.Ts = ts
			// Move to the next line with specified offsets.
			case "Td":
				if len(op.Params) != 2 {
					return errRange
				}

				fv, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				textState.ProcTd(fv[0], fv[1])
			// Move to the next line with specified offsets.
			case "TD":
				if len(op.Params) != 2 {
					return errRange
				}

				fv, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				textState.ProcTD(fv[0], fv[1])
			// Move to the start of the next line.
			case "T*":
				textState.ProcTStar()
			// Set text line matrix.
			case "Tm":
				if len(op.Params) != 6 {
					return errRange
				}
				fv, err := core.GetNumbersAsFloat(op.Params)
				if err != nil {
					return err
				}

				common.Log.Debug("Text matrix: %+v", fv)
				textState.ProcTm(fv[0], fv[1], fv[2], fv[3], fv[4], fv[5])
			// Move to the next line and show text string.
			case `'`:
				if len(op.Params) != 1 {
					return errRange
				}

				charcodes, ok := core.GetStringBytes(op.Params[0])
				if !ok {
					return errType
				}
				common.Log.Debug("' string: %s", string(charcodes))

				textState.ProcQ(charcodes, ctx)
			// Move to the next line and show text string.
			case `''`:
				if len(op.Params) != 3 {
					return errRange
				}

				aw, err := core.GetNumberAsFloat(op.Params[0])
				if err != nil {
					return err
				}

				ac, err := core.GetNumberAsFloat(op.Params[1])
				if err != nil {
					return err
				}

				charcodes, ok := core.GetStringBytes(op.Params[2])
				if !ok {
					return errType
				}

				textState.ProcDQ(charcodes, aw, ac, ctx)
			// Show text string.
			case "Tj":
				if len(op.Params) != 1 {
					return errRange
				}

				charcodes, ok := core.GetStringBytes(op.Params[0])
				if !ok {
					return errType
				}
				common.Log.Debug("Tj string: %s", string(charcodes))

				textState.ProcTj(charcodes, ctx)
			// Show array of text strings.
			case "TJ":
				if len(op.Params) != 1 {
					return errRange
				}

				array, ok := core.GetArray(op.Params[0])
				if !ok {
					common.Log.Debug("Type: %T", array)
					return errType
				}
				common.Log.Debug("TJ array: %+v", array)

				for _, obj := range array.Elements() {
					switch t := obj.(type) {
					case *core.PdfObjectString:
						if t != nil {
							textState.ProcTj(t.Bytes(), ctx)
						}
					case *core.PdfObjectFloat, *core.PdfObjectInteger:
						val, err := core.GetNumberAsFloat(t)
						if err == nil {
							textState.Translate(-val*0.001*textState.Tf.Size, 0)
						}
					}
				}
			// Set font and font size.
			case "Tf":
				if len(op.Params) != 2 {
					return errRange
				}
				common.Log.Debug("%#v", op.Params)

				// Get font name.
				fontName, ok := core.GetName(op.Params[0])
				if !ok || fontName == nil {
					common.Log.Debug("invalid font name object: %v", op.Params[0])
					return errType
				}
				common.Log.Debug("Font name: %s", fontName.String())

				// Get font size.
				fontSize, err := core.GetNumberAsFloat(op.Params[1])
				if err != nil {
					common.Log.Debug("invalid font size object: %v", op.Params[1])
					return errType
				}
				common.Log.Debug("Font size: %v", fontSize)

				// Search font in resources.
				fObj, has := resources.GetFontByName(*fontName)
				if !has {
					common.Log.Debug("ERROR: Font %s not found", fontName.String())
					return errors.New("font not found")
				}
				common.Log.Debug("Font: %T", fObj)

				fontDict, ok := core.GetDict(fObj)
				if !ok {
					common.Log.Debug("ERROR: could not get font dict")
					return errType
				}

				pdfFont, err := model.NewPdfFontFromPdfObject(fontDict)
				if err != nil {
					common.Log.Debug("ERROR: could not load font from object")
					return err
				}

				baseFont := pdfFont.BaseFont()
				if baseFont == "" {
					baseFont = fontName.String()
				}

				textFont, ok := fontCache[baseFont]
				if !ok {
					textFont, err = context.NewTextFont(pdfFont, fontSize)
					if err != nil {
						common.Log.Debug("ERROR: %v", err)
					}
				}

				if textFont == nil {
					// Treat cases such as: OPEIOA+ArialMT
					if len(baseFont) > 7 && baseFont[6] == '+' {
						baseFont = baseFont[7:]
					}

					substitutes := []string{baseFont, "Times New Roman", "Arial", "DejaVu Sans"}
					for _, name := range substitutes {
						common.Log.Debug("DEBUG: searching system font `%s`", name)

						// Check if font is cached.
						if textFont, ok = fontCache[name]; ok {
							break
						}

						// Find font or suitable alternative.
						fontInfo := fontFinder.Match(name)
						if fontInfo == nil {
							common.Log.Debug("could not find font file %s", name)
							continue
						}

						// Load matched font.
						textFont, err = context.NewTextFontFromPath(fontInfo.Filename, fontSize)
						if err != nil {
							common.Log.Debug("could not load font file %s", fontInfo.Filename)
							continue
						}

						// Update font cache.
						common.Log.Debug("Substituting font %s with %s (%s)", baseFont, fontInfo.Name, fontInfo.Filename)
						fontCache[name] = textFont
						break
					}
				}

				if textFont == nil {
					common.Log.Debug("ERROR: could not find any suitable font")
					return errors.New("could not find any suitable font")
				}

				// Set font.
				textState.ProcTf(textFont.WithSize(fontSize, pdfFont))

			//
			// Marked content operators
			//

			// Begin a marked-content sequence.
			case "BMC", "BDC":
				textState.Reset()
			// End a marked-content sequence.
			case "EMC":
				textState.Reset()
			default:
				common.Log.Debug("ERROR: unsupported operand: %s", op.Operand)
			}

			return nil
		})

	err = processor.Process(resources)
	if err != nil {
		return err
	}

	return nil
}
