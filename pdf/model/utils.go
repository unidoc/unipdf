/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/core"
)

func getUniDocVersion() string {
	return common.Version
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
