/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"fmt"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

type PdfActionType string

// // (Section 12.6.4 p. 417).
// See Table 198 - Action types
const (
	ActionTypeGoTo        PdfActionType = "GoTo"        // Go to a destination in the current document
	ActionTypeGoToR       PdfActionType = "GoToR"       // Go to remote, Go to a destination in another document
	ActionTypeGoToE       PdfActionType = "GoToE"       // Go to embedded, PDF 1.6, Got to a destination in an embedded file
	ActionTypeLaunch      PdfActionType = "Launch"      // Launch an application, usually to open a file
	ActionTypeThread      PdfActionType = "Thread"      // Begin reading an article thread
	ActionTypeURI         PdfActionType = "URI"         // Resolves a uniform resource identifier
	ActionTypeSound       PdfActionType = "Sound"       // Play a sound
	ActionTypeMovie       PdfActionType = "Movie"       // Play a movie
	ActionTypeHide        PdfActionType = "Hide"        //Set an annotation's Hidden flag
	ActionTypeNamed       PdfActionType = "Named"       //Execute an action predefined by the conforming reader
	ActionTypeSubmitForm  PdfActionType = "SubmitForm"  // Send data to a uniform resource locator
	ActionTypeResetForm   PdfActionType = "ResetForm"   // Set fields to their default values
	ActionTypeImportData  PdfActionType = "ImportData"  //Import field values from a file
	ActionTypeJavaScript  PdfActionType = "JavaScript"  // Execute a JavaScript script
	ActionTypeSetOCGState PdfActionType = "SetOCGState" // Set the states of optional content groups
	ActionTypeRendition   PdfActionType = "Rendition"   // Controls the playing of multimedia content
	ActionTypeTrans       PdfActionType = "Trans"       // Updates the display of a document, using a transition dictionary
	ActionTypeGoTo3DView  PdfActionType = "GoTo3DView"  // Set the current view of a #D annotation
)

type PdfAction struct {
	context PdfModel

	Type core.PdfObject
	S    core.PdfObject
	Next core.PdfObjectArray

	container *core.PdfIndirectObject
}

// GetContext returns the action context which contains the specific type-dependent context.
// The context represents the subaction.
func (a *PdfAction) GetContext() PdfModel {
	if a == nil {
		return nil
	}
	return a.context
}

// SetContext sets the sub action (context).
func (a *PdfAction) SetContext(ctx PdfModel) {
	a.context = ctx
}


// GetContainingPdfObject implements interface PdfModel.
func (a *PdfAction) GetContainingPdfObject() core.PdfObject {
	return a.container
}

// ToPdfObject implements interface PdfModel.
// Note: Call the sub-annotation's ToPdfObject to set both the generic and non-generic information.
func (a *PdfAction) ToPdfObject() core.PdfObject {
	container := a.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.Clear()

	d.Set("Type", core.MakeName("Action"))
	d.SetIfNotNil("S", a.S)
	//d.SetIfNotNil("Next", a.Next) // TODO

	return container
}

type PdfActionGoTo struct {
	*PdfAction
	D core.PdfObject // name, byte string or array
}

type PdfActionGoToR struct {
	*PdfAction
	F core.PdfObject // TODO: file specification
	D core.PdfObject // name, byte string or array
	NewWindow core.PdfObject
}

type PdfActionGoToE struct {
	*PdfAction
	F core.PdfObject // file specification
	D core.PdfObject // name, byte string or array
	NewWindow core.PdfObject
	T core.PdfObject
}

type PdfActionLaunch struct {
	*PdfAction
	F core.PdfObject
	Win core.PdfObject
	Mac core.PdfObject
	Unix core.PdfObject
	NewWindow core.PdfObject
}

// NewPdfAnnotation returns an initialized generic PDF annotation model.
func NewPdfAction() *PdfAction {
	action := &PdfAction{}
	action.container = core.MakeIndirectObject(core.MakeDict())
	return action
}

// NewPdfActionGoTo returns a new "go to" action
func NewPdfActionGoTo() *PdfActionGoTo {
	action := NewPdfAction()
	goToAction := &PdfActionGoTo{}
	goToAction.PdfAction = action
	action.SetContext(goToAction)
	return goToAction
}

// NewPdfActionGoToR returns a new "go to remote" action
func NewPdfActionGoToR() *PdfActionGoToR {
	action := NewPdfAction()
	goToRAction := &PdfActionGoToR{}
	goToRAction.PdfAction = action
	action.SetContext(goToRAction)
	return goToRAction
}

// NewPdfActionGoToR returns a new "go to remote" action
func NewPdfActionLaunch() *PdfActionLaunch {
	action := NewPdfAction()
	launchAction := &PdfActionLaunch{}
	launchAction.PdfAction = action
	action.SetContext(launchAction)
	return launchAction
}

// ToPdfObject implements interface PdfModel.
func (gotoAct *PdfActionGoTo) ToPdfObject() core.PdfObject {
	gotoAct.PdfAction.ToPdfObject()
	container := gotoAct.container
	d := container.PdfObject.(*core.PdfObjectDictionary)


	d.SetIfNotNil("S", core.MakeName(string(ActionTypeLaunch)))
	d.SetIfNotNil("D", gotoAct.D)
	return container
}

// ToPdfObject implements interface PdfModel.
func (launchAct *PdfActionLaunch) ToPdfObject() core.PdfObject {
	launchAct.PdfAction.ToPdfObject()
	container := launchAct.container
	d := container.PdfObject.(*core.PdfObjectDictionary)


	d.SetIfNotNil("S", core.MakeName(string(ActionTypeLaunch)))
	d.SetIfNotNil("F", launchAct.F)
	d.SetIfNotNil("Win", launchAct.Win)
	d.SetIfNotNil("Mac", launchAct.Mac)
	d.SetIfNotNil("Unix", launchAct.Unix)
	d.SetIfNotNil("NewWindow", launchAct.NewWindow)
	return container
}

// Used for PDF parsing.  Loads a PDF annotation model from a PDF dictionary.
// Loads the common PDF annotation dictionary, and anything needed for the annotation subtype.
func (r *PdfReader) newPdfActionFromIndirectObject(container *core.PdfIndirectObject) (*PdfAction, error) {
	d, isDict := container.PdfObject.(*core.PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("action indirect object not containing a dictionary")
	}

	// Check if cached, return cached model if exists.
	if model := r.modelManager.GetModelFromPrimitive(d); model != nil {
		action, ok := model.(*PdfAction)
		if !ok {
			return nil, fmt.Errorf("cached model not a PDF action")
		}
		return action, nil
	}

	action := &PdfAction{}
	action.container = container
	r.modelManager.Register(d, action)

	if obj := d.Get("Type"); obj != nil {
		str, ok := obj.(*core.PdfObjectName)
		if !ok {
			common.Log.Trace("Incompatibility! Invalid type of Type (%T) - should be Name", obj)
		} else {
			if *str != "Action" {
				// Log a debug message.
				// Not returning an error on this.
				common.Log.Trace("Unsuspected Type != Action (%s)", *str)
			}
		}
	}

	// TODO
	/*if obj := d.Get("Next"); obj != nil {
		action.Next = obj
	}*/

	if obj := d.Get("S"); obj != nil {
		action.S = obj
	}

	actionType, ok := action.S.(*core.PdfObjectName)
	if !ok {
		common.Log.Debug("ERROR: Invalid S object type != name (%T)", action.S)
		return nil, fmt.Errorf("invalid S object type != name (%T)", action.S)
	}
	switch string(*actionType) {
	case string(ActionTypeGoTo):
		ctx, err := r.newPdfActionGotoFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAction = action
		action.context = ctx
		return action, nil
	case string(ActionTypeLaunch):
		ctx, err := r.newPdfActionLaunchFromDict(d)
		if err != nil {
			return nil, err
		}
		ctx.PdfAction = action
		action.context = ctx
		return action, nil

	}

	common.Log.Debug("ERROR: Ignoring unknown action: %s", *actionType)
	return nil, nil
}

func (r *PdfReader) newPdfActionGotoFromDict(d *core.PdfObjectDictionary) (*PdfActionGoTo, error) {
	action := PdfActionGoTo{}

	action.D = d.Get("D")

	return &action, nil
}

func (r *PdfReader) newPdfActionLaunchFromDict(d *core.PdfObjectDictionary) (*PdfActionLaunch, error) {
	action := PdfActionLaunch{}

	action.F = d.Get("F")
	action.Win = d.Get("Win")
	action.Mac = d.Get("Mac")
	action.Unix = d.Get("Unix")
	action.NewWindow = d.Get("NewWindow")

	return &action, nil
}