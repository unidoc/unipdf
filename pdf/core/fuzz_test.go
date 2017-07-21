package core

import (
	"testing"

	"github.com/unidoc/unidoc/common"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
}

// Fuzz tests based on findings with go-fuzz.

// Test for a crash in
// func (this *PdfParser) Trace(obj PdfObject) (PdfObject, error)
// when passing a reference to a non-existing object.
func TestFuzzParserTrace1(t *testing.T) {
	parser := PdfParser{}
	parser.rs, parser.reader = makeReaderForText(" /Name")

	ref := &PdfObjectReference{ObjectNumber: -1}
	obj, err := parser.Trace(ref)

	// Should return non-err, and a nil object.
	if err != nil {
		t.Errorf("Fail, err != nil (%v)", err)
	}

	if _, isNil := obj.(*PdfObjectNull); !isNil {
		t.Errorf("Fail, obj != PdfObjectNull (%T)", obj)
	}
}

// Test for an endless loop when stream length referring to itself.
/*
Found from fuzzing creating an object like:
	13 0 obj
	<< /Length 13 0 R >>
	stream
	xxx
	endstream

*/
func TestFuzzSelfReference1(t *testing.T) {
	rawText := `13 0 obj
<< /Length 13 0 R >>
stream
xxx
endstream
`

	parser := PdfParser{}
	parser.xrefs = make(XrefTable)
	parser.objstms = make(ObjectStreams)
	parser.rs, parser.reader = makeReaderForText(rawText)

	// Point to the start of the stream (where obj 13 starts).
	parser.xrefs[13] = XrefObject{
		XREF_TABLE_ENTRY,
		13,
		0,
		0,
		0,
		0,
	}

	_, err := parser.ParseIndirectObject()
	if err == nil {
		t.Errorf("Should fail with an error")
	}
}
