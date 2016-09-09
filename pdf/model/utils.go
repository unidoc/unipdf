/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

func getUniDocVersion() string {
	return common.Version
}

func getNumberAsFloat(obj PdfObject) (float64, error) {
	if fObj, ok := obj.(*PdfObjectFloat); ok {
		return float64(*fObj), nil
	}

	if iObj, ok := obj.(*PdfObjectInteger); ok {
		return float64(*iObj), nil
	}

	return 0, errors.New("Not a number")
}

// Cases where expecting an integer, but some implementations actually
// store the number in a floating point format.
func getNumberAsInt64(obj PdfObject) (int64, error) {
	if iObj, ok := obj.(*PdfObjectInteger); ok {
		return int64(*iObj), nil
	}

	if fObj, ok := obj.(*PdfObjectFloat); ok {
		common.Log.Debug("Number expected as integer was stored as float (type casting used)")
		return int64(*fObj), nil
	}

	return 0, errors.New("Not a number")
}

func getNumberAsFloatOrNull(obj PdfObject) (*float64, error) {
	if fObj, ok := obj.(*PdfObjectFloat); ok {
		num := float64(*fObj)
		return &num, nil
	}

	if iObj, ok := obj.(*PdfObjectInteger); ok {
		num := float64(*iObj)
		return &num, nil
	}
	if _, ok := obj.(*PdfObjectNull); ok {
		return nil, nil
	}

	return nil, errors.New("Not a number")
}
