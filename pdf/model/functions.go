/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"math"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model/sampling"
	"github.com/unidoc/unidoc/pdf/ps"
)

type PdfValue interface{}

type PdfFunction interface {
	Evaluate([]float64) ([]float64, error)
	ToPdfObject() PdfObject
}

// In PDF: A function object may be a dictionary or a stream, depending on the type of function.
// - Stream: Type 0, Type 4
// - Dictionary: Type 2, Type 3.

// Loads a PDF Function from a PdfObject (can be either stream or dictionary).
func newPdfFunctionFromPdfObject(obj PdfObject) (PdfFunction, error) {
	if stream, is := obj.(*PdfObjectStream); is {
		dict := stream.PdfObjectDictionary

		ftype, ok := dict.Get("FunctionType").(*PdfObjectInteger)
		if !ok {
			common.Log.Error("FunctionType number missing")
			return nil, errors.New("Invalid parameter or missing")
		}

		if *ftype == 0 {
			return newPdfFunctionType0FromStream(stream)
		} else if *ftype == 4 {
			return newPdfFunctionType4FromStream(stream)
		} else {
			return nil, errors.New("Invalid function type")
		}
	} else if indObj, is := obj.(*PdfIndirectObject); is {
		// Indirect object containing a dictionary.
		// The indirect object is the container (which is tracked).
		dict, ok := indObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			common.Log.Error("Function Indirect object not containing dictionary")
			return nil, errors.New("Invalid parameter or missing")
		}

		ftype, ok := dict.Get("FunctionType").(*PdfObjectInteger)
		if !ok {
			common.Log.Error("FunctionType number missing")
			return nil, errors.New("Invalid parameter or missing")
		}

		if *ftype == 2 {
			return newPdfFunctionType2FromPdfObject(indObj)
		} else if *ftype == 3 {
			return newPdfFunctionType3FromPdfObject(indObj)
		} else {
			return nil, errors.New("Invalid function type")
		}
	} else if dict, is := obj.(*PdfObjectDictionary); is {
		ftype, ok := dict.Get("FunctionType").(*PdfObjectInteger)
		if !ok {
			common.Log.Error("FunctionType number missing")
			return nil, errors.New("Invalid parameter or missing")
		}

		if *ftype == 2 {
			return newPdfFunctionType2FromPdfObject(dict)
		} else if *ftype == 3 {
			return newPdfFunctionType3FromPdfObject(dict)
		} else {
			return nil, errors.New("Invalid function type")
		}
	} else {
		common.Log.Debug("Function Type error: %#v", obj)
		return nil, errors.New("Type error")
	}
}

// Simple linear interpolation from the PDF manual.
func interpolate(x, xmin, xmax, ymin, ymax float64) float64 {
	if math.Abs(xmax-xmin) < 0.000001 {
		return ymin
	}

	y := ymin + (x-xmin)*(ymax-ymin)/(xmax-xmin)
	return y
}

//
// Type 0 functions use a sequence of sample values (contained in a stream) to provide an approximation
// for functions whose domains and ranges are bounded. The samples are organized as an m-dimensional
// table in which each entry has n components
//
type PdfFunctionType0 struct {
	Domain []float64 // required; 2*m length; where m is the number of input values
	Range  []float64 // required (type 0); 2*n length; where n is the number of output values

	NumInputs  int
	NumOutputs int

	Size          []int
	BitsPerSample int
	Order         int // Values 1 or 3 (linear or cubic spline interpolation)
	Encode        []float64
	Decode        []float64

	rawData []byte
	data    []uint32

	container *PdfObjectStream
}

// Construct the PDF function object from a stream object (typically loaded from a PDF file).
func newPdfFunctionType0FromStream(stream *PdfObjectStream) (*PdfFunctionType0, error) {
	fun := &PdfFunctionType0{}

	fun.container = stream

	dict := stream.PdfObjectDictionary

	// Domain
	array, has := TraceToDirectObject(dict.Get("Domain")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Domain not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	if len(*array) < 0 || len(*array)%2 != 0 {
		common.Log.Error("Domain invalid")
		return nil, errors.New("Invalid domain range")
	}
	fun.NumInputs = len(*array) / 2
	domain, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Domain = domain

	// Range
	array, has = TraceToDirectObject(dict.Get("Range")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Range not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	if len(*array) < 0 || len(*array)%2 != 0 {
		return nil, errors.New("Invalid range")
	}
	fun.NumOutputs = len(*array) / 2
	rang, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Range = rang

	// Number of samples in each input dimension
	array, has = TraceToDirectObject(dict.Get("Size")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Size not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	tablesize, err := array.ToIntegerArray()
	if err != nil {
		return nil, err
	}
	if len(tablesize) != fun.NumInputs {
		common.Log.Error("Table size not matching number of inputs")
		return nil, errors.New("Range check")
	}
	fun.Size = tablesize

	// BitsPerSample
	bps, has := TraceToDirectObject(dict.Get("BitsPerSample")).(*PdfObjectInteger)
	if !has {
		common.Log.Error("BitsPerSample not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	if *bps != 1 && *bps != 2 && *bps != 4 && *bps != 8 && *bps != 12 && *bps != 16 && *bps != 24 && *bps != 32 {
		common.Log.Error("Bits per sample outside range (%d)", *bps)
		return nil, errors.New("Range check")
	}
	fun.BitsPerSample = int(*bps)

	fun.Order = 1
	order, has := TraceToDirectObject(dict.Get("Order")).(*PdfObjectInteger)
	if has {
		if *order != 1 && *order != 3 {
			common.Log.Error("Invalid order (%d)", *order)
			return nil, errors.New("Range check")
		}
		fun.Order = int(*order)
	}

	// Encode: is a 2*m array specifying the linear mapping of input values into the domain of the function's
	// sample table.
	array, has = TraceToDirectObject(dict.Get("Encode")).(*PdfObjectArray)
	if has {
		encode, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Encode = encode
	}

	// Decode
	array, has = TraceToDirectObject(dict.Get("Decode")).(*PdfObjectArray)
	if has {
		decode, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Decode = decode
	}

	data, err := DecodeStream(stream)
	if err != nil {
		return nil, err
	}
	fun.rawData = data

	return fun, nil
}

func (this *PdfFunctionType0) ToPdfObject() PdfObject {
	container := this.container
	if container != nil {
		this.container = &PdfObjectStream{}
	}

	dict := MakeDict()
	dict.Set("FunctionType", MakeInteger(0))

	// Domain (required).
	domainArray := &PdfObjectArray{}
	for _, val := range this.Domain {
		domainArray.Append(MakeFloat(val))
	}
	dict.Set("Domain", domainArray)

	// Range (required).
	rangeArray := &PdfObjectArray{}
	for _, val := range this.Range {
		rangeArray.Append(MakeFloat(val))
	}
	dict.Set("Range", rangeArray)

	// Size (required).
	sizeArray := &PdfObjectArray{}
	for _, val := range this.Size {
		sizeArray.Append(MakeInteger(int64(val)))
	}
	dict.Set("Size", sizeArray)

	dict.Set("BitsPerSample", MakeInteger(int64(this.BitsPerSample)))

	if this.Order != 1 {
		dict.Set("Order", MakeInteger(int64(this.Order)))
	}

	// TODO: Encode.
	// Either here, or automatically later on when writing out.
	dict.Set("Length", MakeInteger(int64(len(this.rawData))))
	container.Stream = this.rawData

	container.PdfObjectDictionary = dict
	return container
}

func (this *PdfFunctionType0) Evaluate(x []float64) ([]float64, error) {
	if len(x) != this.NumInputs {
		common.Log.Error("Number of inputs not matching what is needed")
		return nil, errors.New("Range check error")
	}

	if this.data == nil {
		// Process the samples if not already done.
		err := this.processSamples()
		if err != nil {
			return nil, err
		}
	}

	// Fall back to default Encode/Decode params if not set.
	encode := this.Encode
	if encode == nil {
		encode = []float64{}
		for i := 0; i < len(this.Size); i++ {
			encode = append(encode, 0)
			encode = append(encode, float64(this.Size[i]-1))
		}
	}
	decode := this.Decode
	if decode == nil {
		decode = this.Range
	}

	indices := []int{}
	// Start with nearest neighbour interpolation.
	for i := 0; i < len(x); i++ {
		xi := x[i]

		xip := math.Min(math.Max(xi, this.Domain[2*i]), this.Domain[2*i+1])

		ei := interpolate(xip, this.Domain[2*i], this.Domain[2*i+1], encode[2*i], encode[2*i+1])
		eip := math.Min(math.Max(ei, 0), float64(this.Size[i]))
		// eip represents coordinate into the data table.
		// At this point it is real values.

		// Interpolation shall be used to to determine output values
		// from the nearest surrounding values in the sample table.

		// Initial implementation is simply nearest neighbour.
		// Then will add the linear and possibly bicubic/spline.
		index := int(math.Floor(eip + 0.5))
		if index < 0 {
			index = 0
		} else if index > this.Size[i] {
			index = this.Size[i] - 1
		}
		indices = append(indices, index)

	}

	// Calculate the index
	m := indices[0]
	for i := 1; i < this.NumInputs; i++ {
		add := indices[i]
		for j := 0; j < i; j++ {
			add *= this.Size[j]
		}
		m += add
	}
	m *= this.NumOutputs

	// Output values.
	outputs := []float64{}
	for j := 0; j < this.NumOutputs; j++ {
		rj := this.data[m+j]
		rjp := interpolate(float64(rj), 0, math.Pow(2, float64(this.BitsPerSample)), decode[2*j], decode[2*j+1])
		yj := math.Min(math.Max(rjp, this.Range[2*j]), this.Range[2*j+1])
		outputs = append(outputs, yj)
	}

	return outputs, nil
}

// Convert raw data to data table.  The maximum supported BitsPerSample is 32, so we store the resulting data
// in a uint32 array.  This is somewhat wasteful in the case of a small BitsPerSample, but these tables are
// presumably not huge at any rate.
func (this *PdfFunctionType0) processSamples() error {
	data := sampling.ResampleBytes(this.rawData, this.BitsPerSample)
	this.data = data

	return nil
}

//
// Type 2 functions define an exponential interpolation of one input value and n
// output values:
//      f(x) = y_0, ..., y_(n-1)
// y_j = C0_j + x^N * (C1_j - C0_j); for 0 <= j < n
// When N=1 ; linear interpolation between C0 and C1.
//
type PdfFunctionType2 struct {
	Domain []float64
	Range  []float64

	C0 []float64
	C1 []float64
	N  float64

	container *PdfIndirectObject
}

// Can be either indirect object or dictionary.  If indirect, then must be holding a dictionary,
// i.e. acting as a container. When converting back to pdf object, will use the container provided.

func newPdfFunctionType2FromPdfObject(obj PdfObject) (*PdfFunctionType2, error) {
	fun := &PdfFunctionType2{}

	var dict *PdfObjectDictionary
	if indObj, is := obj.(*PdfIndirectObject); is {
		d, ok := indObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			return nil, errors.New("Type check error")
		}
		fun.container = indObj
		dict = d
	} else if d, is := obj.(*PdfObjectDictionary); is {
		dict = d
	} else {
		return nil, errors.New("Type check error")
	}

	common.Log.Trace("FUNC2: %s", dict.String())

	// Domain
	array, has := TraceToDirectObject(dict.Get("Domain")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Domain not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	if len(*array) < 0 || len(*array)%2 != 0 {
		common.Log.Error("Domain range invalid")
		return nil, errors.New("Invalid domain range")
	}
	domain, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Domain = domain

	// Range
	array, has = TraceToDirectObject(dict.Get("Range")).(*PdfObjectArray)
	if has {
		if len(*array) < 0 || len(*array)%2 != 0 {
			return nil, errors.New("Invalid range")
		}

		rang, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Range = rang
	}

	// C0.
	array, has = TraceToDirectObject(dict.Get("C0")).(*PdfObjectArray)
	if has {
		c0, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.C0 = c0
	}

	// C1.
	array, has = TraceToDirectObject(dict.Get("C1")).(*PdfObjectArray)
	if has {
		c1, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.C1 = c1
	}

	if len(fun.C0) != len(fun.C1) {
		common.Log.Error("C0 and C1 not matching")
		return nil, errors.New("Range check")
	}

	// Exponent.
	N, err := getNumberAsFloat(TraceToDirectObject(dict.Get("N")))
	if err != nil {
		common.Log.Error("N missing or invalid, dict: %s", dict.String())
		return nil, err
	}
	fun.N = N

	return fun, nil
}

func (this *PdfFunctionType2) ToPdfObject() PdfObject {
	dict := MakeDict()

	dict.Set("FunctionType", MakeInteger(2))

	// Domain (required).
	domainArray := &PdfObjectArray{}
	for _, val := range this.Domain {
		domainArray.Append(MakeFloat(val))
	}
	dict.Set("Domain", domainArray)

	// Range (required).
	if this.Range != nil {
		rangeArray := &PdfObjectArray{}
		for _, val := range this.Range {
			rangeArray.Append(MakeFloat(val))
		}
		dict.Set("Range", rangeArray)
	}

	// C0.
	if this.C0 != nil {
		c0Array := &PdfObjectArray{}
		for _, val := range this.C0 {
			c0Array.Append(MakeFloat(val))
		}
		dict.Set("C0", c0Array)
	}

	// C1.
	if this.C1 != nil {
		c1Array := &PdfObjectArray{}
		for _, val := range this.C1 {
			c1Array.Append(MakeFloat(val))
		}
		dict.Set("C1", c1Array)
	}

	// exponent
	dict.Set("N", MakeFloat(this.N))

	// Wrap in a container if we have one already specified.
	if this.container != nil {
		this.container.PdfObject = dict
		return this.container
	} else {
		return dict
	}

}

func (this *PdfFunctionType2) Evaluate(x []float64) ([]float64, error) {
	if len(x) != 1 {
		common.Log.Error("Only one input allowed")
		return nil, errors.New("Range check")
	}

	// Prepare.
	c0 := []float64{0.0}
	if this.C0 != nil {
		c0 = this.C0
	}
	c1 := []float64{1.0}
	if this.C1 != nil {
		c1 = this.C1
	}

	y := []float64{}
	for i := 0; i < len(c0); i++ {
		yi := c0[i] + math.Pow(x[0], this.N)*(c1[i]-c0[i])
		y = append(y, yi)
	}

	return y, nil
}

//
// Type 3 functions define stitching of the subdomains of serveral 1-input functions to produce
// a single new 1-input function.
//
type PdfFunctionType3 struct {
	Domain []float64
	Range  []float64

	Functions []PdfFunction // k-1 input functions
	Bounds    []float64     // k-1 numbers; defines the intervals where each function applies
	Encode    []float64     // Array of 2k numbers..

	container *PdfIndirectObject
}

func (this *PdfFunctionType3) Evaluate(x []float64) ([]float64, error) {
	if len(x) != 1 {
		common.Log.Error("Only one input allowed")
		return nil, errors.New("Range check")
	}

	// Determine which function to use

	// Encode

	return nil, errors.New("Not implemented yet")
}

func newPdfFunctionType3FromPdfObject(obj PdfObject) (*PdfFunctionType3, error) {
	fun := &PdfFunctionType3{}

	var dict *PdfObjectDictionary
	if indObj, is := obj.(*PdfIndirectObject); is {
		d, ok := indObj.PdfObject.(*PdfObjectDictionary)
		if !ok {
			return nil, errors.New("Type check error")
		}
		fun.container = indObj
		dict = d
	} else if d, is := obj.(*PdfObjectDictionary); is {
		dict = d
	} else {
		return nil, errors.New("Type check error")
	}

	// Domain
	array, has := TraceToDirectObject(dict.Get("Domain")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Domain not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	if len(*array) != 2 {
		common.Log.Error("Domain invalid")
		return nil, errors.New("Invalid domain range")
	}
	domain, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Domain = domain

	// Range
	array, has = TraceToDirectObject(dict.Get("Range")).(*PdfObjectArray)
	if has {
		if len(*array) < 0 || len(*array)%2 != 0 {
			return nil, errors.New("Invalid range")
		}
		rang, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Range = rang
	}

	// Functions.
	array, has = TraceToDirectObject(dict.Get("Functions")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Functions not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	fun.Functions = []PdfFunction{}
	for _, obj := range *array {
		subf, err := newPdfFunctionFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		fun.Functions = append(fun.Functions, subf)
	}

	// Bounds
	array, has = TraceToDirectObject(dict.Get("Bounds")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Bounds not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	bounds, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Bounds = bounds
	if len(fun.Bounds) != len(fun.Functions)-1 {
		common.Log.Error("Bounds (%d) and num functions (%d) not matching", len(fun.Bounds), len(fun.Functions))
		return nil, errors.New("Range check")
	}

	// Encode.
	array, has = TraceToDirectObject(dict.Get("Encode")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Encode not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	encode, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Encode = encode
	if len(fun.Encode) != 2*len(fun.Functions) {
		common.Log.Error("Len encode (%d) and num functions (%d) not matching up", len(fun.Encode), len(fun.Functions))
		return nil, errors.New("Range check")
	}

	return fun, nil
}

func (this *PdfFunctionType3) ToPdfObject() PdfObject {
	dict := MakeDict()

	dict.Set("FunctionType", MakeInteger(3))

	// Domain (required).
	domainArray := &PdfObjectArray{}
	for _, val := range this.Domain {
		domainArray.Append(MakeFloat(val))
	}
	dict.Set("Domain", domainArray)

	// Range (required).
	if this.Range != nil {
		rangeArray := &PdfObjectArray{}
		for _, val := range this.Range {
			rangeArray.Append(MakeFloat(val))
		}
		dict.Set("Range", rangeArray)
	}

	// Functions
	if this.Functions != nil {
		fArray := &PdfObjectArray{}
		for _, fun := range this.Functions {
			fArray.Append(fun.ToPdfObject())
		}
		dict.Set("Functions", fArray)
	}

	// Bounds.
	if this.Bounds != nil {
		bArray := &PdfObjectArray{}
		for _, val := range this.Bounds {
			bArray.Append(MakeFloat(val))
		}
		dict.Set("Bounds", bArray)
	}

	// Encode.
	if this.Encode != nil {
		eArray := &PdfObjectArray{}
		for _, val := range this.Encode {
			eArray.Append(MakeFloat(val))
		}
		dict.Set("Encode", eArray)
	}

	// Wrap in a container if we have one already specified.
	if this.container != nil {
		this.container.PdfObject = dict
		return this.container
	} else {
		return dict
	}
}

//
// Type 4.  Postscript calculator functions.
//
type PdfFunctionType4 struct {
	Domain  []float64
	Range   []float64
	Program *ps.PSProgram

	executor    *ps.PSExecutor
	decodedData []byte

	container *PdfObjectStream
}

// Input [x1 x2 x3]
func (this *PdfFunctionType4) Evaluate(xVec []float64) ([]float64, error) {
	if this.executor == nil {
		this.executor = ps.NewPSExecutor(this.Program)
	}

	inputs := []ps.PSObject{}
	for _, val := range xVec {
		inputs = append(inputs, ps.MakeReal(val))
	}

	outputs, err := this.executor.Execute(inputs)
	if err != nil {
		return nil, err
	}

	// After execution the outputs are on the stack [y1 ... yM]
	// Convert to floats.
	yVec, err := ps.PSObjectArrayToFloat64Array(outputs)
	if err != nil {
		return nil, err
	}

	return yVec, nil
}

// Load a type 4 function from a PDF stream object.
func newPdfFunctionType4FromStream(stream *PdfObjectStream) (*PdfFunctionType4, error) {
	fun := &PdfFunctionType4{}

	fun.container = stream

	dict := stream.PdfObjectDictionary

	// Domain
	array, has := TraceToDirectObject(dict.Get("Domain")).(*PdfObjectArray)
	if !has {
		common.Log.Error("Domain not specified")
		return nil, errors.New("Required attribute missing or invalid")
	}
	if len(*array)%2 != 0 {
		common.Log.Error("Domain invalid")
		return nil, errors.New("Invalid domain range")
	}
	domain, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Domain = domain

	// Range
	array, has = TraceToDirectObject(dict.Get("Range")).(*PdfObjectArray)
	if has {
		if len(*array) < 0 || len(*array)%2 != 0 {
			return nil, errors.New("Invalid range")
		}
		rang, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Range = rang
	}

	// Program.  Decode the program and parse the PS code.
	decoded, err := DecodeStream(stream)
	if err != nil {
		return nil, err
	}
	fun.decodedData = decoded

	psParser := ps.NewPSParser([]byte(decoded))
	prog, err := psParser.Parse()
	if err != nil {
		return nil, err
	}
	fun.Program = prog

	return fun, nil
}

func (this *PdfFunctionType4) ToPdfObject() PdfObject {
	container := this.container
	if container == nil {
		this.container = &PdfObjectStream{}
		container = this.container
	}

	dict := MakeDict()
	dict.Set("FunctionType", MakeInteger(4))

	// Domain (required).
	domainArray := &PdfObjectArray{}
	for _, val := range this.Domain {
		domainArray.Append(MakeFloat(val))
	}
	dict.Set("Domain", domainArray)

	// Range (required).
	rangeArray := &PdfObjectArray{}
	for _, val := range this.Range {
		rangeArray.Append(MakeFloat(val))
	}
	dict.Set("Range", rangeArray)

	if this.decodedData == nil && this.Program != nil {
		// Update data.  This is used for created functions (not parsed ones).
		this.decodedData = []byte(this.Program.String())
	}

	// TODO: Encode.
	// Either here, or automatically later on when writing out.
	dict.Set("Length", MakeInteger(int64(len(this.decodedData))))

	container.Stream = this.decodedData
	container.PdfObjectDictionary = dict

	return container
}
