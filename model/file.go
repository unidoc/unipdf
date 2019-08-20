package model

import (
	"fmt"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// (Section 7.11.3 p. 102).
// See Table 44 - Entries in a file specification dictionary

// PdfFilespec represents a file specification
type PdfFilespec struct {
	Type core.PdfObject
	FS   core.PdfObject
	F    core.PdfObject // A file specification string
	UF   core.PdfObject //A Unicode text string that provides file specification
	DOS  core.PdfObject // A file specification string representing a DOS file name. OBSOLETE
	Mac  core.PdfObject // A file specification string representing a Mac OS file name. OBSOLETE
	Unix core.PdfObject // A file specification string representing a UNIX file name. OBSOLETE
	ID   core.PdfObject // An array of two byte strings constituting a file identifier
	V    core.PdfObject // A flag indicating whether the file referenced by the file specification is volatile (changes frequently with time).
	EF   core.PdfObject // A dictionary containing a subset of the keys F, UF, DOS, Mac, and Unix, corresponding to the entries by those names in the file specification dictionary
	RF   core.PdfObject
	Desc core.PdfObject // Descriptive text associated with the file specification
	CI   core.PdfObject // A collection item dictionary, which shall be used to create the user interface for portable collections

	container *core.PdfIndirectObject
}

// GetContainingPdfObject implements interface PdfModel.
func (f *PdfFilespec) GetContainingPdfObject() core.PdfObject {
	return f.container
}

// ToPdfObject implements interface PdfModel.
func (f *PdfFilespec) ToPdfObject() core.PdfObject {
	container := f.container
	d := container.PdfObject.(*core.PdfObjectDictionary)

	d.Clear()

	d.Set("Type", core.MakeName("Filespec"))
	d.SetIfNotNil("FS", f.FS)
	d.SetIfNotNil("F", f.F)
	d.SetIfNotNil("UF", f.UF)
	d.SetIfNotNil("DOS", f.DOS)
	d.SetIfNotNil("Mac", f.Mac)
	d.SetIfNotNil("Unix", f.Unix)
	d.SetIfNotNil("ID", f.ID)
	d.SetIfNotNil("V", f.V)
	d.SetIfNotNil("EF", f.EF)
	d.SetIfNotNil("RF", f.RF)
	d.SetIfNotNil("Desc", f.Desc)
	d.SetIfNotNil("CI", f.CI)

	return container
}

// Used for PDF parsing.  Loads a PDF filespec model from a PDF dictionary.
func (r *PdfReader) newPdfFilespecFromIndirectObject(container *core.PdfIndirectObject) (*PdfFilespec, error) {
	d, isDict := container.PdfObject.(*core.PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("annotation indirect object not containing a dictionary")
	}

	// Check if cached, return cached model if exists.
	if model := r.modelManager.GetModelFromPrimitive(d); model != nil {
		fs, ok := model.(*PdfFilespec)
		if !ok {
			return nil, fmt.Errorf("cached model not a PDF annotation")
		}
		return fs, nil
	}

	fs := &PdfFilespec{}
	fs.container = container
	r.modelManager.Register(d, fs)

	if obj := d.Get("Type"); obj != nil {
		str, ok := obj.(*core.PdfObjectName)
		if !ok {
			common.Log.Trace("Incompatibility! Invalid type of Type (%T) - should be Name", obj)
		} else {
			if *str != "Filespec" {
				// Log a debug message.
				// Not returning an error on this.
				common.Log.Trace("Unsuspected Type != Filespec (%s)", *str)
			}
		}
	}

	if obj := d.Get("FS"); obj != nil {
		fs.FS = obj
	}

	if obj := d.Get("F"); obj != nil {
		fs.F = obj
	}

	if obj := d.Get("UF"); obj != nil {
		fs.UF = obj
	}

	if obj := d.Get("DOS"); obj != nil {
		fs.DOS = obj
	}

	if obj := d.Get("Mac"); obj != nil {
		fs.Mac = obj
	}

	if obj := d.Get("Unix"); obj != nil {
		fs.Unix = obj
	}

	if obj := d.Get("ID"); obj != nil {
		fs.ID = obj
	}

	if obj := d.Get("V"); obj != nil {
		fs.V = obj
	}

	if obj := d.Get("EF"); obj != nil {
		fs.EF = obj
	}

	if obj := d.Get("RF"); obj != nil {
		fs.RF = obj
	}

	if obj := d.Get("Desc"); obj != nil {
		fs.Desc = obj
	}

	if obj := d.Get("CI"); obj != nil {
		fs.CI = obj
	}

	return fs, nil
}

// NewPdfFilespec returns an initialized generic PDF filespec model.
func NewPdfFilespec() *PdfFilespec {
	action := &PdfFilespec{}
	action.container = core.MakeIndirectObject(core.MakeDict())
	return action
}