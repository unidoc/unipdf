package extractor

import (
	"errors"

	"github.com/unidoc/unidoc/pdf/core"
)

// getNumberAsFloat can retrieve numeric values from PdfObject (both integer/float).
func getNumberAsFloat(obj core.PdfObject) (float64, error) {
	if fObj, ok := obj.(*core.PdfObjectFloat); ok {
		return float64(*fObj), nil
	}

	if iObj, ok := obj.(*core.PdfObjectInteger); ok {
		return float64(*iObj), nil
	}

	return 0, errors.New("Not a number")
}
