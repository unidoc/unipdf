package contentstream

import . "github.com/unidoc/unidoc/pdf/core"

func makeParamsFromFloats(vals []float64) []PdfObject {
	params := []PdfObject{}
	for _, val := range vals {
		params = append(params, MakeFloat(val))
	}
	return params
}

func makeParamsFromNames(vals []PdfObjectName) []PdfObject {
	params := []PdfObject{}
	for _, val := range vals {
		params = append(params, MakeName(string(val)))
	}
	return params
}

func makeParamsFromInts(vals []int64) []PdfObject {
	params := []PdfObject{}
	for _, val := range vals {
		params = append(params, MakeInteger(val))
	}
	return params
}
