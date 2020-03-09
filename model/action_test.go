/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/core"
)

// testAction loads an action object from object number 1 loaded from `rawText` PDF content and checks that
// it matches the `actionType`. Then it applies testFunc() on the action which does action-specific checks.
// Lastly the serialized output is checked against the input PDF object.
func testAction(t *testing.T, rawText string, actionType PdfActionType, testFunc func(t *testing.T, action *PdfAction)) {
	// Read raw text
	r := NewReaderForText(rawText)

	err := r.ParseIndObjSeries()
	require.NoError(t, err)

	// Load the field from object number 1 as all actions in these tests are defined in object 1
	obj, err := r.parser.LookupByNumber(1)
	require.NoError(t, err)

	ind, ok := obj.(*core.PdfIndirectObject)
	require.True(t, ok)

	// Parse action
	action, err := r.newPdfActionFromIndirectObject(ind)
	require.NoError(t, err)

	// Check if raw text can be parsed to the expected action objects

	// The object should be of type action + the actionType should match the expected action
	require.Equal(t, "Action", action.Type.String())
	require.Equal(t, string(actionType), action.S.String())

	// Verify some action specific fields
	testFunc(t, action)

	// Check if object can be serialized to the expected text
	outDict, ok := core.GetDict(action.context.ToPdfObject())
	if !ok {
		t.Fatalf("error")
	}

	require.Containsf(
		t,
		strings.Replace(rawText, "\n", "", -1),
		outDict.WriteString(),
		"generated output doesn't match the expected output - %s",
		outDict.WriteString())
}

func TestPdfActionGoTo(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /GoTo
/D (name)
>>
endobj
`
	testAction(t, rawText, ActionTypeGoTo, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionGoTo)
		require.True(t, ok)
		require.Equal(t, "name", contextAction.D.String())
	})
}

func TestPdfActionGoToR(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /GoToR
/F <</Type /Filespec
/F (someFile.pdf)
>>
/D (name)
/NewWindow true
>>
endobj
`

	testAction(t, rawText, ActionTypeGoToR, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionGoToR)
		require.True(t, ok)
		require.Equal(t, "name", contextAction.D.String())
		require.Equal(t, "true", contextAction.NewWindow.String())
		require.IsType(t, &PdfFilespec{}, contextAction.F)
	})
}

func TestPdfActionGoToE(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /GoToE
/F <</Type /Filespec
/F (someFile.pdf)
>>
/D (name)
/NewWindow true
/T <</R /C
/N (Embedded document)>>
>>
endobj
`

	testAction(t, rawText, ActionTypeGoToE, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionGoToE)
		require.True(t, ok)
		require.Equal(t, "name", contextAction.D.String())
		require.Equal(t, "true", contextAction.NewWindow.String())
		require.IsType(t, &PdfFilespec{}, contextAction.F)
	})
}

func TestPdfActionLaunch(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /Launch
/F <</Type /Filespec
/F (someFile.pdf)
>>
/NewWindow true
>>
endobj
`

	testAction(t, rawText, ActionTypeLaunch, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionLaunch)
		require.True(t, ok)
		require.Equal(t, "true", contextAction.NewWindow.String())
		require.IsType(t, &PdfFilespec{}, contextAction.F)
	})
}

func TestPdfActionThread(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /Thread
/D 4
/B 5
>>
endobj
`

	testAction(t, rawText, ActionTypeThread, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionThread)
		require.True(t, ok)
		require.Equal(t, "4", contextAction.D.String())
		require.Equal(t, "5", contextAction.B.String())
	})
}

func TestPdfActionURI(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /URI
/URI (https://unidoc.io/)
/IsMap true
>>
endobj
`

	testAction(t, rawText, ActionTypeURI, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionURI)
		require.True(t, ok)
		require.Equal(t, "https://unidoc.io/", contextAction.URI.String())
		require.Equal(t, "true", contextAction.IsMap.String())
	})
}

func TestPdfActionSound(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /Sound
/Sound 2 0 R
/Volume 0.5
/Synchronous true
/Repeat true
/Mix true
>>
endobj

2 0 obj
<<
/B 16
/C 2
/E /Signed
/Filter /FlateDecode
/Length 12
/R 44100
/Type /Sound
>>
stream
abcdefghijkl
endstream
endobj
`

	testAction(t, rawText, ActionTypeSound, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionSound)
		require.True(t, ok)
		require.Equal(t, "Object stream 2: Dict(\"B\": 16, \"C\": 2, \"E\": Signed, \"Filter\": FlateDecode, \"Length\": 12, \"R\": 44100, \"Type\": Sound, )", contextAction.Sound.String())
		require.Equal(t, "0.500000", contextAction.Volume.String())
		require.Equal(t, "true", contextAction.Synchronous.String())
		require.Equal(t, "true", contextAction.Repeat.String())
		require.Equal(t, "true", contextAction.Mix.String())
	})
}

func TestPdfActionMovie(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /Movie
/Annotation <</Foo (bar)>>
/T (Title of the movie)
/Operation /Stop
>>
endobj
`

	testAction(t, rawText, ActionTypeMovie, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionMovie)
		require.True(t, ok)
		require.Equal(t, "Dict(\"Foo\": bar, )", contextAction.Annotation.String())
		require.Equal(t, "Title of the movie", contextAction.T.String())
		require.Equal(t, "Stop", contextAction.Operation.String())
	})
}

func TestPdfActionHide(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /Hide
/T (Field)
/H false
>>
endobj
`

	testAction(t, rawText, ActionTypeHide, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionHide)
		require.True(t, ok)
		require.Equal(t, "Field", contextAction.T.String())
		require.Equal(t, "false", contextAction.H.String())
	})
}

func TestPdfActionNamed(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /Named
/N /NextPage
>>
endobj
`

	testAction(t, rawText, ActionTypeNamed, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionNamed)
		require.True(t, ok)
		require.Equal(t, "NextPage", contextAction.N.String())
	})
}

func TestPdfActionSubmitForm(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /SubmitForm
/F <</Type /Filespec
/F (someFile.pdf)
>>
/Fields [(Address) (By) (Date) (Email) (TelNum) (Title)]
/Flags 2
>>
endobj
`

	testAction(t, rawText, ActionTypeSubmitForm, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionSubmitForm)
		require.True(t, ok)
		require.Equal(t, "[Address, By, Date, Email, TelNum, Title]", contextAction.Fields.String())
		require.Equal(t, "2", contextAction.Flags.String())
	})
}

func TestPdfActionResetForm(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /ResetForm
/Fields [(Address) (By) (Date) (Email) (TelNum) (Title)]
/Flags 2
>>
endobj
`

	testAction(t, rawText, ActionTypeResetForm, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionResetForm)
		require.True(t, ok)
		require.Equal(t, "[Address, By, Date, Email, TelNum, Title]", contextAction.Fields.String())
		require.Equal(t, "2", contextAction.Flags.String())
	})
}

func TestPdfActionImportData(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /ImportData
/F <</Type /Filespec
/F (someFile.pdf)
>>
>>
endobj
`

	testAction(t, rawText, ActionTypeImportData, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionImportData)
		require.True(t, ok)
		require.IsType(t, &PdfFilespec{}, contextAction.F)
	})
}

func TestPdfActionSetOCGState(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /SetOCGState
/State [/Off <</OffFoo (Bar)>> /Toggle <</ToggleFoo (Bar)>> /ON <</OnFoo (Bar)>>]
/PreserveRB false
>>
endobj
`

	testAction(t, rawText, ActionTypeSetOCGState, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionSetOCGState)
		require.True(t, ok)
		require.Equal(t, "[Off, Dict(\"OffFoo\": Bar, ), Toggle, Dict(\"ToggleFoo\": Bar, ), ON, Dict(\"OnFoo\": Bar, )]", contextAction.State.String())
		require.Equal(t, "false", contextAction.PreserveRB.String())
	})
}

func TestPdfActionRendition(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /Rendition
/R <</R 1>>
/AN <</AN 2>>
/OP 4
/JS (javascript)
>>
endobj
`

	testAction(t, rawText, ActionTypeRendition, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionRendition)
		require.True(t, ok)
		require.Equal(t, "Dict(\"R\": 1, )", contextAction.R.String())
		require.Equal(t, "Dict(\"AN\": 2, )", contextAction.AN.String())
		require.Equal(t, "4", contextAction.OP.String())
		require.Equal(t, "javascript", contextAction.JS.String())
	})
}

func TestPdfActionTrans(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /Trans
/Trans <</X 123/Y 456>>
>>
endobj
`

	testAction(t, rawText, ActionTypeTrans, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionTrans)
		require.True(t, ok)
		require.Equal(t, "Dict(\"X\": 123, \"Y\": 456, )", contextAction.Trans.String())
	})
}

func TestPdfActionGoto3DView(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /GoTo3DView
/TA <<
/X 123
/Y 456
>>
/V <</Name (fake)>>
>>
endobj
`

	testAction(t, rawText, ActionTypeGoTo3DView, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionGoTo3DView)
		require.True(t, ok)
		require.Equal(t, "Dict(\"X\": 123, \"Y\": 456, )", contextAction.TA.String())
		require.Equal(t, "Dict(\"Name\": fake, )", contextAction.V.String())
	})
}

func TestPdfActionJavaScript(t *testing.T) {
	rawText := `
1 0 obj
<</Type /Action
/S /JavaScript
/JS (alert\("test"\))
>>
endobj
`

	testAction(t, rawText, ActionTypeJavaScript, func(t *testing.T, action *PdfAction) {
		contextAction, ok := action.context.(*PdfActionJavaScript)
		require.True(t, ok)
		require.Equal(t, "alert(\"test\")", contextAction.JS.String())
	})
}

func TestNewPdfAction(t *testing.T) {
	action := NewPdfAction()
	require.IsType(t, &PdfAction{}, action)
	require.IsType(t, &core.PdfIndirectObject{}, action.container)
}

func TestNewPdfActionGoTo(t *testing.T) {
	action := NewPdfActionGoTo()
	require.IsType(t, &PdfActionGoTo{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionGoToR(t *testing.T) {
	action := NewPdfActionGoToR()
	require.IsType(t, &PdfActionGoToR{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionGoToE(t *testing.T) {
	action := NewPdfActionGoToE()
	require.IsType(t, &PdfActionGoToE{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionLaunch(t *testing.T) {
	action := NewPdfActionLaunch()
	require.IsType(t, &PdfActionLaunch{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionThread(t *testing.T) {
	action := NewPdfActionThread()
	require.IsType(t, &PdfActionThread{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionURI(t *testing.T) {
	action := NewPdfActionURI()
	require.IsType(t, &PdfActionURI{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionSound(t *testing.T) {
	action := NewPdfActionSound()
	require.IsType(t, &PdfActionSound{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionMovie(t *testing.T) {
	action := NewPdfActionMovie()
	require.IsType(t, &PdfActionMovie{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionHide(t *testing.T) {
	action := NewPdfActionHide()
	require.IsType(t, &PdfActionHide{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionNamed(t *testing.T) {
	action := NewPdfActionNamed()
	require.IsType(t, &PdfActionNamed{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionSubmitForm(t *testing.T) {
	action := NewPdfActionSubmitForm()
	require.IsType(t, &PdfActionSubmitForm{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionResetForm(t *testing.T) {
	action := NewPdfActionResetForm()
	require.IsType(t, &PdfActionResetForm{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionImportData(t *testing.T) {
	action := NewPdfActionImportData()
	require.IsType(t, &PdfActionImportData{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionSetOCGState(t *testing.T) {
	action := NewPdfActionSetOCGState()
	require.IsType(t, &PdfActionSetOCGState{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionRendition(t *testing.T) {
	action := NewPdfActionRendition()
	require.IsType(t, &PdfActionRendition{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionTrans(t *testing.T) {
	action := NewPdfActionTrans()
	require.IsType(t, &PdfActionTrans{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionGoTo3DView(t *testing.T) {
	action := NewPdfActionGoTo3DView()
	require.IsType(t, &PdfActionGoTo3DView{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}

func TestNewPdfActionJavaScript(t *testing.T) {
	action := NewPdfActionJavaScript()
	require.IsType(t, &PdfActionJavaScript{}, action)
	require.IsType(t, &PdfAction{}, action.PdfAction)
	require.Equal(t, action, action.PdfAction.context)
}
