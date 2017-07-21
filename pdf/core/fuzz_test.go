package core

import "testing"

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
