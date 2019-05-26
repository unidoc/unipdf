/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

func getUniDocVersion() string {
	return common.Version
}

// NewReaderForText makes a new PdfReader for an input PDF content string. For use in testing.
func NewReaderForText(txt string) *PdfReader {
	// Create the parser, loads the cross reference table and trailer.
	return &PdfReader{
		traversed:    map[core.PdfObject]struct{}{},
		modelManager: newModelManager(),
		parser:       core.NewParserFromString(txt),
	}
}

// Handy function for debugging in development.
func debugObject(obj core.PdfObject) {
	common.Log.Debug("obj: %T %s", obj, obj.String())

	if stream, is := obj.(*core.PdfObjectStream); is {
		decoded, err := core.DecodeStream(stream)
		if err != nil {
			common.Log.Debug("Error: %v", err)
			return
		}
		common.Log.Debug("Decoded: %s", decoded)
	} else if indObj, is := obj.(*core.PdfIndirectObject); is {
		common.Log.Debug("%T %v", indObj.PdfObject, indObj.PdfObject)
		common.Log.Debug("%s", indObj.PdfObject.String())
	}
}
