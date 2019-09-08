/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// (Section 7.11.3 p. 102).
// See Table 44 - Entries in a file specification dictionary

/*
 * A PDF file can refer to the contents of another file by using a file specification (PDF 1.1), which shall take either
 * of two forms:
 * • A simple file specification shall give just the name of the target file in a standard format, independent of the
 * naming conventions of any particular file system. It shall take the form of either a string or a dictionary
 * • A full file specification shall include information related to one or more specific file systems. It shall only be
 * represented as a dictionary.
 *
 * A file specification shall refer to a file external to the PDF file or to a file embedded within the referring PDF file,
 * allowing its contents to be stored or transmitted along with the PDF file. The file shall be considered to be
 * external to the PDF file in either case.
 * A file specification could describe a URL-based file system and will follow the rules of Internet RFC 1808, Relative Uniform Resource Locators
 */

// PdfFilespec represents a file specification which can either refer to an external or embedded file.
type PdfFilespec struct {
	Type core.PdfObject
	FS   core.PdfObject
	F    core.PdfObject // A file specification string
	UF   core.PdfObject // A Unicode text string that provides file specification
	DOS  core.PdfObject // A file specification string representing a DOS file name. OBSOLETE
	Mac  core.PdfObject // A file specification string representing a Mac OS file name. OBSOLETE
	Unix core.PdfObject // A file specification string representing a UNIX file name. OBSOLETE
	ID   core.PdfObject // An array of two byte strings constituting a file identifier
	V    core.PdfObject // A flag indicating whether the file referenced by the file specification is volatile (changes frequently with time).
	EF   core.PdfObject // A dictionary containing a subset of the keys F, UF, DOS, Mac, and Unix, corresponding to the entries by those names in the file specification dictionary
	RF   core.PdfObject
	Desc core.PdfObject // Descriptive text associated with the file specification
	CI   core.PdfObject // A collection item dictionary, which shall be used to create the user interface for portable collections

	container core.PdfObject
}

// GetContainingPdfObject implements interface PdfModel.
func (f *PdfFilespec) GetContainingPdfObject() core.PdfObject {
	return f.container
}

func (f *PdfFilespec) getDict() *core.PdfObjectDictionary {
	if indObj, is := f.container.(*core.PdfIndirectObject); is {
		dict, ok := indObj.PdfObject.(*core.PdfObjectDictionary)
		if !ok {
			return nil
		}
		return dict
	} else if dictObj, isDict := f.container.(*core.PdfObjectDictionary); isDict {
		return dictObj
	} else {
		common.Log.Debug("Trying to access Filespec dictionary of invalid object type (%T)", f.container)
		return nil
	}
}

// ToPdfObject implements interface PdfModel.
func (f *PdfFilespec) ToPdfObject() core.PdfObject {
	d := f.getDict()

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

	return f.container
}

// NewPdfFilespecFromObj creates and returns a new PdfFilespec object.
func NewPdfFilespecFromObj(obj core.PdfObject) (*PdfFilespec, error) {
	fs := &PdfFilespec{}

	var dict *core.PdfObjectDictionary

	if indObj, isInd := core.GetIndirect(obj); isInd {
		fs.container = indObj

		d, ok := core.GetDict(indObj.PdfObject)
		if !ok {
			common.Log.Debug("Object not a dictionary type")
			return nil, core.ErrTypeError
		}
		dict = d
	} else if d, isDict := core.GetDict(obj); isDict {
		fs.container = d
		dict = d
	} else {
		common.Log.Debug("Object type unexpected (%T)", obj)
		return nil, core.ErrTypeError
	}

	if dict == nil {
		common.Log.Debug("Dictionary missing")
		return nil, errors.New("dict missing")
	}

	if obj := dict.Get("Type"); obj != nil {
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
	if obj := dict.Get("FS"); obj != nil {
		fs.FS = obj
	}
	if obj := dict.Get("F"); obj != nil {
		fs.F = obj
	}
	if obj := dict.Get("UF"); obj != nil {
		fs.UF = obj
	}
	if obj := dict.Get("DOS"); obj != nil {
		fs.DOS = obj
	}
	if obj := dict.Get("Mac"); obj != nil {
		fs.Mac = obj
	}
	if obj := dict.Get("Unix"); obj != nil {
		fs.Unix = obj
	}
	if obj := dict.Get("ID"); obj != nil {
		fs.ID = obj
	}
	if obj := dict.Get("V"); obj != nil {
		fs.V = obj
	}
	if obj := dict.Get("EF"); obj != nil {
		fs.EF = obj
	}
	if obj := dict.Get("RF"); obj != nil {
		fs.RF = obj
	}
	if obj := dict.Get("Desc"); obj != nil {
		fs.Desc = obj
	}
	if obj := dict.Get("CI"); obj != nil {
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
