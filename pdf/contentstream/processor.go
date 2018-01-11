/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package contentstream

import (
	"errors"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
	. "github.com/unidoc/unidoc/pdf/model"
)

// Basic graphics state implementation.
// Initially only implementing and tracking a portion of the information specified.  Easy to add more.
type GraphicsState struct {
	ColorspaceStroking    PdfColorspace
	ColorspaceNonStroking PdfColorspace
	ColorStroking         PdfColor
	ColorNonStroking      PdfColor
}

type GraphicStateStack []GraphicsState

func (gsStack *GraphicStateStack) Push(gs GraphicsState) {
	*gsStack = append(*gsStack, gs)
}

func (gsStack *GraphicStateStack) Pop() GraphicsState {
	gs := (*gsStack)[len(*gsStack)-1]
	*gsStack = (*gsStack)[:len(*gsStack)-1]
	return gs
}

// ContentStreamProcessor defines a data structure and methods for processing a content stream, keeping track of the
// current graphics state, and allowing external handlers to define their own functions as a part of the processing,
// for example rendering or extracting certain information.
type ContentStreamProcessor struct {
	graphicsStack GraphicStateStack
	operations    []*ContentStreamOperation
	graphicsState GraphicsState

	handlers     []HandlerEntry
	currentIndex int
}

type HandlerFunc func(op *ContentStreamOperation, gs GraphicsState, resources *PdfPageResources) error

type HandlerEntry struct {
	Condition HandlerConditionEnum
	Operand   string
	Handler   HandlerFunc
}

type HandlerConditionEnum int

func (this HandlerConditionEnum) All() bool {
	return this == HandlerConditionEnumAllOperands
}

func (this HandlerConditionEnum) Operand() bool {
	return this == HandlerConditionEnumOperand
}

const (
	HandlerConditionEnumOperand     HandlerConditionEnum = iota
	HandlerConditionEnumAllOperands HandlerConditionEnum = iota
)

func NewContentStreamProcessor(ops []*ContentStreamOperation) *ContentStreamProcessor {
	csp := ContentStreamProcessor{}
	csp.graphicsStack = GraphicStateStack{}

	// Set defaults..
	gs := GraphicsState{}

	csp.graphicsState = gs

	csp.handlers = []HandlerEntry{}
	csp.currentIndex = 0
	csp.operations = ops

	return &csp
}

func (csp *ContentStreamProcessor) AddHandler(condition HandlerConditionEnum, operand string, handler HandlerFunc) {
	entry := HandlerEntry{}
	entry.Condition = condition
	entry.Operand = operand
	entry.Handler = handler
	csp.handlers = append(csp.handlers, entry)
}

func (csp *ContentStreamProcessor) getColorspace(name string, resources *PdfPageResources) (PdfColorspace, error) {
	switch name {
	case "DeviceGray":
		return NewPdfColorspaceDeviceGray(), nil
	case "DeviceRGB":
		return NewPdfColorspaceDeviceRGB(), nil
	case "DeviceCMYK":
		return NewPdfColorspaceDeviceCMYK(), nil
	case "Pattern":
		return NewPdfColorspaceSpecialPattern(), nil
	}

	// Next check the colorspace dictionary.
	cs, has := resources.ColorSpace.Colorspaces[name]
	if has {
		return cs, nil
	}

	// Lastly check other potential colormaps.
	switch name {
	case "CalGray":
		return NewPdfColorspaceCalGray(), nil
	case "CalRGB":
		return NewPdfColorspaceCalRGB(), nil
	case "Lab":
		return NewPdfColorspaceLab(), nil
	}

	// Otherwise unsupported.
	common.Log.Debug("Unknown colorspace requested: %s", name)
	return nil, errors.New("Unsupported colorspace")
}

// Get initial color for a given colorspace.
func (csp *ContentStreamProcessor) getInitialColor(cs PdfColorspace) (PdfColor, error) {
	switch cs := cs.(type) {
	case *PdfColorspaceDeviceGray:
		return NewPdfColorDeviceGray(0.0), nil
	case *PdfColorspaceDeviceRGB:
		return NewPdfColorDeviceRGB(0.0, 0.0, 0.0), nil
	case *PdfColorspaceDeviceCMYK:
		return NewPdfColorDeviceCMYK(0.0, 0.0, 0.0, 1.0), nil
	case *PdfColorspaceCalGray:
		return NewPdfColorCalGray(0.0), nil
	case *PdfColorspaceCalRGB:
		return NewPdfColorCalRGB(0.0, 0.0, 0.0), nil
	case *PdfColorspaceLab:
		l := 0.0
		a := 0.0
		b := 0.0
		if cs.Range[0] > 0 {
			l = cs.Range[0]
		}
		if cs.Range[2] > 0 {
			a = cs.Range[2]
		}
		return NewPdfColorLab(l, a, b), nil
	case *PdfColorspaceICCBased:
		if cs.Alternate == nil {
			// Alternate not defined.
			// Try to fall back to DeviceGray, DeviceRGB or DeviceCMYK.
			common.Log.Trace("ICC Based not defined - attempting fall back (N = %d)", cs.N)
			if cs.N == 1 {
				common.Log.Trace("Falling back to DeviceGray")
				return csp.getInitialColor(NewPdfColorspaceDeviceGray())
			} else if cs.N == 3 {
				common.Log.Trace("Falling back to DeviceRGB")
				return csp.getInitialColor(NewPdfColorspaceDeviceRGB())
			} else if cs.N == 4 {
				common.Log.Trace("Falling back to DeviceCMYK")
				return csp.getInitialColor(NewPdfColorspaceDeviceCMYK())
			} else {
				return nil, errors.New("Alternate space not defined for ICC")
			}
		}
		return csp.getInitialColor(cs.Alternate)
	case *PdfColorspaceSpecialIndexed:
		if cs.Base == nil {
			return nil, errors.New("Indexed base not specified")
		}
		return csp.getInitialColor(cs.Base)
	case *PdfColorspaceSpecialSeparation:
		if cs.AlternateSpace == nil {
			return nil, errors.New("Alternate space not specified")
		}
		return csp.getInitialColor(cs.AlternateSpace)
	case *PdfColorspaceDeviceN:
		if cs.AlternateSpace == nil {
			return nil, errors.New("Alternate space not specified")
		}
		return csp.getInitialColor(cs.AlternateSpace)
	case *PdfColorspaceSpecialPattern:
		// FIXME/check: A pattern does not have an initial color...
		return nil, nil
	}

	common.Log.Debug("Unable to determine initial color for unknown colorspace: %T", cs)
	return nil, errors.New("Unsupported colorspace")
}

// Process the entire operations.
func (this *ContentStreamProcessor) Process(resources *PdfPageResources) error {
	// Initialize graphics state
	this.graphicsState.ColorspaceStroking = NewPdfColorspaceDeviceGray()
	this.graphicsState.ColorspaceNonStroking = NewPdfColorspaceDeviceGray()
	this.graphicsState.ColorStroking = NewPdfColorDeviceGray(0)
	this.graphicsState.ColorNonStroking = NewPdfColorDeviceGray(0)

	for _, op := range this.operations {
		var err error

		// Internal handling.
		switch op.Operand {
		case "q":
			this.graphicsStack.Push(this.graphicsState)
		case "Q":
			this.graphicsState = this.graphicsStack.Pop()

		// Color operations (Table 74 p. 179)
		case "CS":
			err = this.handleCommand_CS(op, resources)
		case "cs":
			err = this.handleCommand_cs(op, resources)
		case "SC":
			err = this.handleCommand_SC(op, resources)
		case "SCN":
			err = this.handleCommand_SCN(op, resources)
		case "sc":
			err = this.handleCommand_sc(op, resources)
		case "scn":
			err = this.handleCommand_scn(op, resources)
		case "G":
			err = this.handleCommand_G(op, resources)
		case "g":
			err = this.handleCommand_g(op, resources)
		case "RG":
			err = this.handleCommand_RG(op, resources)
		case "rg":
			err = this.handleCommand_rg(op, resources)
		case "K":
			err = this.handleCommand_K(op, resources)
		case "k":
			err = this.handleCommand_k(op, resources)
		}
		if err != nil {
			common.Log.Debug("Processor handling error (%s): %v", op.Operand, err)
			common.Log.Debug("Operand: %#v", op.Operand)
			return err
		}

		// Check if have external handler also, and process if so.
		for _, entry := range this.handlers {
			var err error
			if entry.Condition.All() {
				err = entry.Handler(op, this.graphicsState, resources)
			} else if entry.Condition.Operand() && op.Operand == entry.Operand {
				err = entry.Handler(op, this.graphicsState, resources)
			}
			if err != nil {
				common.Log.Debug("Processor handler error: %v", err)
				return err
			}
		}
	}

	return nil
}

// CS: Set the current color space for stroking operations.
func (csp *ContentStreamProcessor) handleCommand_CS(op *ContentStreamOperation, resources *PdfPageResources) error {
	if len(op.Params) < 1 {
		common.Log.Debug("Invalid cs command, skipping over")
		return errors.New("Too few parameters")
	}
	if len(op.Params) > 1 {
		common.Log.Debug("cs command with too many parameters - continuing")
		return errors.New("Too many parameters")
	}
	name, ok := op.Params[0].(*PdfObjectName)
	if !ok {
		common.Log.Debug("ERROR: cs command with invalid parameter, skipping over")
		return errors.New("Type check error")
	}
	// Set the current color space to use for stroking operations.
	// Either device based or referring to resource dict.
	cs, err := csp.getColorspace(string(*name), resources)
	if err != nil {
		return err
	}
	csp.graphicsState.ColorspaceStroking = cs

	// Set initial color.
	color, err := csp.getInitialColor(cs)
	if err != nil {
		return err
	}
	csp.graphicsState.ColorStroking = color

	return nil
}

// cs: Set the current color space for non-stroking operations.
func (csp *ContentStreamProcessor) handleCommand_cs(op *ContentStreamOperation, resources *PdfPageResources) error {
	if len(op.Params) < 1 {
		common.Log.Debug("Invalid CS command, skipping over")
		return errors.New("Too few parameters")
	}
	if len(op.Params) > 1 {
		common.Log.Debug("CS command with too many parameters - continuing")
		return errors.New("Too many parameters")
	}
	name, ok := op.Params[0].(*PdfObjectName)
	if !ok {
		common.Log.Debug("ERROR: CS command with invalid parameter, skipping over")
		return errors.New("Type check error")
	}
	// Set the current color space to use for non-stroking operations.
	// Either device based or referring to resource dict.
	cs, err := csp.getColorspace(string(*name), resources)
	if err != nil {
		return err
	}
	csp.graphicsState.ColorspaceNonStroking = cs

	// Set initial color.
	color, err := csp.getInitialColor(cs)
	if err != nil {
		return err
	}
	csp.graphicsState.ColorNonStroking = color

	return nil
}

// SC: Set the color to use for stroking operations in a device, CIE-based or Indexed colorspace. (not ICC based)
func (this *ContentStreamProcessor) handleCommand_SC(op *ContentStreamOperation, resources *PdfPageResources) error {
	// For DeviceGray, CalGray, Indexed: one operand is required
	// For DeviceRGB, CalRGB, Lab: 3 operands required

	cs := this.graphicsState.ColorspaceStroking
	if len(op.Params) != cs.GetNumComponents() {
		common.Log.Debug("Invalid number of parameters for SC")
		common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
		return errors.New("Invalid number of parameters")
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorStroking = color
	return nil
}

func isPatternCS(cs PdfColorspace) bool {
	_, isPattern := cs.(*PdfColorspaceSpecialPattern)
	return isPattern
}

// SCN: Same as SC but also supports Pattern, Separation, DeviceN and ICCBased color spaces.
func (this *ContentStreamProcessor) handleCommand_SCN(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := this.graphicsState.ColorspaceStroking

	if !isPatternCS(cs) {
		if len(op.Params) != cs.GetNumComponents() {
			common.Log.Debug("Invalid number of parameters for SC")
			common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
			return errors.New("Invalid number of parameters")
		}
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorStroking = color

	return nil
}

// sc: Same as SC except used for non-stroking operations.
func (this *ContentStreamProcessor) handleCommand_sc(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := this.graphicsState.ColorspaceNonStroking

	if !isPatternCS(cs) {
		if len(op.Params) != cs.GetNumComponents() {
			common.Log.Debug("Invalid number of parameters for SC")
			common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
			return errors.New("Invalid number of parameters")
		}
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorNonStroking = color

	return nil
}

// scn: Same as SCN except used for non-stroking operations.
func (this *ContentStreamProcessor) handleCommand_scn(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := this.graphicsState.ColorspaceNonStroking

	if !isPatternCS(cs) {
		if len(op.Params) != cs.GetNumComponents() {
			common.Log.Debug("Invalid number of parameters for SC")
			common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
			return errors.New("Invalid number of parameters")
		}
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		common.Log.Debug("ERROR: Fail to get color from params: %+v (CS is %+v)", op.Params, cs)
		return err
	}

	this.graphicsState.ColorNonStroking = color

	return nil
}

// G: Set the stroking colorspace to DeviceGray, and the color to the specified graylevel (range [0-1]).
// gray G
func (this *ContentStreamProcessor) handleCommand_G(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := NewPdfColorspaceDeviceGray()
	if len(op.Params) != cs.GetNumComponents() {
		common.Log.Debug("Invalid number of parameters for SC")
		common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
		return errors.New("Invalid number of parameters")
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorspaceStroking = cs
	this.graphicsState.ColorStroking = color

	return nil
}

// g: Same as G, but for non-stroking colorspace and color (range [0-1]).
// gray g
func (this *ContentStreamProcessor) handleCommand_g(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := NewPdfColorspaceDeviceGray()
	if len(op.Params) != cs.GetNumComponents() {
		common.Log.Debug("Invalid number of parameters for SC")
		common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
		return errors.New("Invalid number of parameters")
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorspaceNonStroking = cs
	this.graphicsState.ColorNonStroking = color

	return nil
}

// RG: Sets the stroking colorspace to DeviceRGB and the stroking color to r,g,b. [0-1] ranges.
// r g b RG
func (this *ContentStreamProcessor) handleCommand_RG(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := NewPdfColorspaceDeviceRGB()
	if len(op.Params) != cs.GetNumComponents() {
		common.Log.Debug("Invalid number of parameters for SC")
		common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
		return errors.New("Invalid number of parameters")
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorspaceStroking = cs
	this.graphicsState.ColorStroking = color

	return nil
}

// rg: Same as RG but for non-stroking colorspace, color.
func (this *ContentStreamProcessor) handleCommand_rg(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := NewPdfColorspaceDeviceRGB()
	if len(op.Params) != cs.GetNumComponents() {
		common.Log.Debug("Invalid number of parameters for SC")
		common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
		return errors.New("Invalid number of parameters")
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorspaceNonStroking = cs
	this.graphicsState.ColorNonStroking = color

	return nil
}

// K: Sets the stroking colorspace to DeviceCMYK and the stroking color to c,m,y,k. [0-1] ranges.
// c m y k K
func (this *ContentStreamProcessor) handleCommand_K(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := NewPdfColorspaceDeviceCMYK()
	if len(op.Params) != cs.GetNumComponents() {
		common.Log.Debug("Invalid number of parameters for SC")
		common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
		return errors.New("Invalid number of parameters")
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorspaceStroking = cs
	this.graphicsState.ColorStroking = color

	return nil
}

// k: Same as K but for non-stroking colorspace, color.
func (this *ContentStreamProcessor) handleCommand_k(op *ContentStreamOperation, resources *PdfPageResources) error {
	cs := NewPdfColorspaceDeviceCMYK()
	if len(op.Params) != cs.GetNumComponents() {
		common.Log.Debug("Invalid number of parameters for SC")
		common.Log.Debug("Number %d not matching colorspace %T", len(op.Params), cs)
		return errors.New("Invalid number of parameters")
	}

	color, err := cs.ColorFromPdfObjects(op.Params)
	if err != nil {
		return err
	}

	this.graphicsState.ColorspaceNonStroking = cs
	this.graphicsState.ColorNonStroking = color

	return nil
}
