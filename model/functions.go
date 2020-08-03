/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"math"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/internal/sampling"
	"github.com/unidoc/unipdf/v3/ps"
)

// PdfFunction interface represents the common methods of a function in PDF.
type PdfFunction interface {
	Evaluate([]float64) ([]float64, error)
	ToPdfObject() core.PdfObject
}

// In PDF: A function object may be a dictionary or a stream, depending on the type of function.
// - Stream: Type 0, Type 4
// - Dictionary: Type 2, Type 3.

// Loads a PDF Function from a PdfObject (can be either stream or dictionary).
func newPdfFunctionFromPdfObject(obj core.PdfObject) (PdfFunction, error) {
	obj = core.ResolveReference(obj)
	if stream, is := obj.(*core.PdfObjectStream); is {
		dict := stream.PdfObjectDictionary

		ftype, ok := dict.Get("FunctionType").(*core.PdfObjectInteger)
		if !ok {
			common.Log.Error("FunctionType number missing")
			return nil, errors.New("invalid parameter or missing")
		}

		if *ftype == 0 {
			return newPdfFunctionType0FromStream(stream)
		} else if *ftype == 4 {
			return newPdfFunctionType4FromStream(stream)
		} else {
			return nil, errors.New("invalid function type")
		}
	} else if indObj, is := obj.(*core.PdfIndirectObject); is {
		// Indirect object containing a dictionary.
		// The indirect object is the container (which is tracked).
		dict, ok := indObj.PdfObject.(*core.PdfObjectDictionary)
		if !ok {
			common.Log.Error("Function Indirect object not containing dictionary")
			return nil, errors.New("invalid parameter or missing")
		}

		ftype, ok := dict.Get("FunctionType").(*core.PdfObjectInteger)
		if !ok {
			common.Log.Error("FunctionType number missing")
			return nil, errors.New("invalid parameter or missing")
		}

		if *ftype == 2 {
			return newPdfFunctionType2FromPdfObject(indObj)
		} else if *ftype == 3 {
			return newPdfFunctionType3FromPdfObject(indObj)
		} else {
			return nil, errors.New("invalid function type")
		}
	} else if dict, is := obj.(*core.PdfObjectDictionary); is {
		ftype, ok := dict.Get("FunctionType").(*core.PdfObjectInteger)
		if !ok {
			common.Log.Error("FunctionType number missing")
			return nil, errors.New("invalid parameter or missing")
		}

		if *ftype == 2 {
			return newPdfFunctionType2FromPdfObject(dict)
		} else if *ftype == 3 {
			return newPdfFunctionType3FromPdfObject(dict)
		} else {
			return nil, errors.New("invalid function type")
		}
	} else {
		common.Log.Debug("Function Type error: %#v", obj)
		return nil, errors.New("type error")
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

// PdfFunctionType0 uses a sequence of sample values (contained in a stream) to provide an approximation
// for functions whose domains and ranges are bounded. The samples are organized as an m-dimensional
// table in which each entry has n components
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

	container *core.PdfObjectStream
}

// Construct the PDF function object from a stream object (typically loaded from a PDF file).
func newPdfFunctionType0FromStream(stream *core.PdfObjectStream) (*PdfFunctionType0, error) {
	fun := &PdfFunctionType0{}

	fun.container = stream

	dict := stream.PdfObjectDictionary

	// Domain
	array, has := core.TraceToDirectObject(dict.Get("Domain")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Domain not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	if array.Len() < 0 || array.Len()%2 != 0 {
		common.Log.Error("Domain invalid")
		return nil, errors.New("invalid domain range")
	}
	fun.NumInputs = array.Len() / 2
	domain, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Domain = domain

	// Range
	array, has = core.TraceToDirectObject(dict.Get("Range")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Range not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	if array.Len() < 0 || array.Len()%2 != 0 {
		return nil, errors.New("invalid range")
	}
	fun.NumOutputs = array.Len() / 2
	rang, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Range = rang

	// Number of samples in each input dimension
	array, has = core.TraceToDirectObject(dict.Get("Size")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Size not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	tablesize, err := array.ToIntegerArray()
	if err != nil {
		return nil, err
	}
	if len(tablesize) != fun.NumInputs {
		common.Log.Error("Table size not matching number of inputs")
		return nil, errors.New("range check")
	}
	fun.Size = tablesize

	// BitsPerSample
	bps, has := core.TraceToDirectObject(dict.Get("BitsPerSample")).(*core.PdfObjectInteger)
	if !has {
		common.Log.Error("BitsPerSample not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	if *bps != 1 && *bps != 2 && *bps != 4 && *bps != 8 && *bps != 12 && *bps != 16 && *bps != 24 && *bps != 32 {
		common.Log.Error("Bits per sample outside range (%d)", *bps)
		return nil, errors.New("range check")
	}
	fun.BitsPerSample = int(*bps)

	fun.Order = 1
	order, has := core.TraceToDirectObject(dict.Get("Order")).(*core.PdfObjectInteger)
	if has {
		if *order != 1 && *order != 3 {
			common.Log.Error("Invalid order (%d)", *order)
			return nil, errors.New("range check")
		}
		fun.Order = int(*order)
	}

	// Encode: is a 2*m array specifying the linear mapping of input values into the domain of the function's
	// sample table.
	array, has = core.TraceToDirectObject(dict.Get("Encode")).(*core.PdfObjectArray)
	if has {
		encode, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Encode = encode
	}

	// Decode
	array, has = core.TraceToDirectObject(dict.Get("Decode")).(*core.PdfObjectArray)
	if has {
		decode, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Decode = decode
	}

	data, err := core.DecodeStream(stream)
	if err != nil {
		return nil, err
	}
	fun.rawData = data

	return fun, nil
}

// ToPdfObject returns the PDF representation of the function.
func (f *PdfFunctionType0) ToPdfObject() core.PdfObject {
	if f.container == nil {
		f.container = &core.PdfObjectStream{}
	}

	dict := core.MakeDict()
	dict.Set("FunctionType", core.MakeInteger(0))

	// Domain (required).
	domainArray := &core.PdfObjectArray{}
	for _, val := range f.Domain {
		domainArray.Append(core.MakeFloat(val))
	}
	dict.Set("Domain", domainArray)

	// Range (required).
	rangeArray := &core.PdfObjectArray{}
	for _, val := range f.Range {
		rangeArray.Append(core.MakeFloat(val))
	}
	dict.Set("Range", rangeArray)

	// Size (required).
	sizeArray := &core.PdfObjectArray{}
	for _, val := range f.Size {
		sizeArray.Append(core.MakeInteger(int64(val)))
	}
	dict.Set("Size", sizeArray)

	dict.Set("BitsPerSample", core.MakeInteger(int64(f.BitsPerSample)))

	if f.Order != 1 {
		dict.Set("Order", core.MakeInteger(int64(f.Order)))
	}

	// TODO: Encode.
	// Either here, or automatically later on when writing out.
	dict.Set("Length", core.MakeInteger(int64(len(f.rawData))))
	f.container.Stream = f.rawData

	f.container.PdfObjectDictionary = dict
	return f.container
}

// Evaluate runs the function on the passed in slice and returns the results.
func (f *PdfFunctionType0) Evaluate(x []float64) ([]float64, error) {
	if len(x) != f.NumInputs {
		common.Log.Error("Number of inputs not matching what is needed")
		return nil, errors.New("range check error")
	}

	if f.data == nil {
		// Process the samples if not already done.
		err := f.processSamples()
		if err != nil {
			return nil, err
		}
	}

	// Fall back to default Encode/Decode params if not set.
	encode := f.Encode
	if encode == nil {
		encode = []float64{}
		for i := 0; i < len(f.Size); i++ {
			encode = append(encode, 0)
			encode = append(encode, float64(f.Size[i]-1))
		}
	}
	decode := f.Decode
	if decode == nil {
		decode = f.Range
	}

	var indices []int
	// Start with nearest neighbour interpolation.
	for i := 0; i < len(x); i++ {
		xi := x[i]

		// See section 7.10.2 Type 0 (Sampled) Functions (pp. 93-94 PDF32000_2008).
		xip := math.Min(math.Max(xi, f.Domain[2*i]), f.Domain[2*i+1])
		ei := interpolate(xip, f.Domain[2*i], f.Domain[2*i+1], encode[2*i], encode[2*i+1])
		eip := math.Min(math.Max(ei, 0), float64(f.Size[i]-1))
		// eip represents coordinate into the data table.
		// At this point it is real values.

		// Interpolation shall be used to to determine output values
		// from the nearest surrounding values in the sample table.

		// Initial implementation is simply nearest neighbour.
		// Then will add the linear and possibly bicubic/spline.
		index := int(math.Floor(eip + 0.5))
		if index < 0 {
			index = 0
		} else if index > f.Size[i] {
			index = f.Size[i] - 1
		}
		indices = append(indices, index)
	}

	// Calculate the index
	m := indices[0]
	for i := 1; i < f.NumInputs; i++ {
		add := indices[i]
		for j := 0; j < i; j++ {
			add *= f.Size[j]
		}
		m += add
	}
	m *= f.NumOutputs

	// Output values.
	var outputs []float64
	for j := 0; j < f.NumOutputs; j++ {
		rjIdx := m + j
		if rjIdx >= len(f.data) {
			common.Log.Debug("WARN: not enough input samples to determine output values. Output may be incorrect.")
			continue
		}

		rj := f.data[rjIdx]
		rjp := interpolate(float64(rj), 0, math.Pow(2, float64(f.BitsPerSample)), decode[2*j], decode[2*j+1])
		yj := math.Min(math.Max(rjp, f.Range[2*j]), f.Range[2*j+1])
		outputs = append(outputs, yj)
	}

	return outputs, nil
}

// Convert raw data to data table.  The maximum supported BitsPerSample is 32, so we store the resulting data
// in a uint32 array.  This is somewhat wasteful in the case of a small BitsPerSample, but these tables are
// presumably not huge at any rate.
func (f *PdfFunctionType0) processSamples() error {
	data := sampling.ResampleBytes(f.rawData, f.BitsPerSample)
	f.data = data

	return nil
}

// PdfFunctionType2 defines an exponential interpolation of one input value and n
// output values:
//      f(x) = y_0, ..., y_(n-1)
// y_j = C0_j + x^N * (C1_j - C0_j); for 0 <= j < n
// When N=1 ; linear interpolation between C0 and C1.
type PdfFunctionType2 struct {
	Domain []float64
	Range  []float64

	C0 []float64
	C1 []float64
	N  float64

	container *core.PdfIndirectObject
}

// Can be either indirect object or dictionary.  If indirect, then must be holding a dictionary,
// i.e. acting as a container. When converting back to pdf object, will use the container provided.
func newPdfFunctionType2FromPdfObject(obj core.PdfObject) (*PdfFunctionType2, error) {
	fun := &PdfFunctionType2{}

	var dict *core.PdfObjectDictionary
	if indObj, is := obj.(*core.PdfIndirectObject); is {
		d, ok := indObj.PdfObject.(*core.PdfObjectDictionary)
		if !ok {
			return nil, errors.New("type check error")
		}
		fun.container = indObj
		dict = d
	} else if d, is := obj.(*core.PdfObjectDictionary); is {
		dict = d
	} else {
		return nil, errors.New("type check error")
	}

	common.Log.Trace("FUNC2: %s", dict.String())

	// Domain
	array, has := core.TraceToDirectObject(dict.Get("Domain")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Domain not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	if array.Len() < 0 || array.Len()%2 != 0 {
		common.Log.Error("Domain range invalid")
		return nil, errors.New("invalid domain range")
	}
	domain, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Domain = domain

	// Range
	array, has = core.TraceToDirectObject(dict.Get("Range")).(*core.PdfObjectArray)
	if has {
		if array.Len() < 0 || array.Len()%2 != 0 {
			return nil, errors.New("invalid range")
		}

		rang, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Range = rang
	}

	// C0.
	array, has = core.TraceToDirectObject(dict.Get("C0")).(*core.PdfObjectArray)
	if has {
		c0, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.C0 = c0
	}

	// C1.
	array, has = core.TraceToDirectObject(dict.Get("C1")).(*core.PdfObjectArray)
	if has {
		c1, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.C1 = c1
	}

	if len(fun.C0) != len(fun.C1) {
		common.Log.Error("C0 and C1 not matching")
		return nil, core.ErrRangeError
	}

	// Exponent.
	N, err := core.GetNumberAsFloat(core.TraceToDirectObject(dict.Get("N")))
	if err != nil {
		common.Log.Error("N missing or invalid, dict: %s", dict.String())
		return nil, err
	}
	fun.N = N

	return fun, nil
}

// ToPdfObject returns the PDF representation of the function.
func (f *PdfFunctionType2) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()

	dict.Set("FunctionType", core.MakeInteger(2))

	// Domain (required).
	domainArray := &core.PdfObjectArray{}
	for _, val := range f.Domain {
		domainArray.Append(core.MakeFloat(val))
	}
	dict.Set("Domain", domainArray)

	// Range (required).
	if f.Range != nil {
		rangeArray := &core.PdfObjectArray{}
		for _, val := range f.Range {
			rangeArray.Append(core.MakeFloat(val))
		}
		dict.Set("Range", rangeArray)
	}

	// C0.
	if f.C0 != nil {
		c0Array := &core.PdfObjectArray{}
		for _, val := range f.C0 {
			c0Array.Append(core.MakeFloat(val))
		}
		dict.Set("C0", c0Array)
	}

	// C1.
	if f.C1 != nil {
		c1Array := &core.PdfObjectArray{}
		for _, val := range f.C1 {
			c1Array.Append(core.MakeFloat(val))
		}
		dict.Set("C1", c1Array)
	}

	// exponent
	dict.Set("N", core.MakeFloat(f.N))

	// Wrap in a container if we have one already specified.
	if f.container != nil {
		f.container.PdfObject = dict
		return f.container
	}

	return dict
}

// Evaluate runs the function on the passed in slice and returns the results.
func (f *PdfFunctionType2) Evaluate(x []float64) ([]float64, error) {
	if len(x) != 1 {
		common.Log.Error("Only one input allowed")
		return nil, errors.New("range check")
	}

	// Prepare.
	c0 := []float64{0.0}
	if f.C0 != nil {
		c0 = f.C0
	}
	c1 := []float64{1.0}
	if f.C1 != nil {
		c1 = f.C1
	}

	var y []float64
	for i := 0; i < len(c0); i++ {
		yi := c0[i] + math.Pow(x[0], f.N)*(c1[i]-c0[i])
		y = append(y, yi)
	}

	return y, nil
}

// PdfFunctionType3 defines stitching of the subdomains of several 1-input functions to produce
// a single new 1-input function.
type PdfFunctionType3 struct {
	Domain []float64
	Range  []float64

	Functions []PdfFunction // k-1 input functions
	Bounds    []float64     // k-1 numbers; defines the intervals where each function applies
	Encode    []float64     // Array of 2k numbers..

	container *core.PdfIndirectObject
}

// Evaluate runs the function on the passed in slice and returns the results.
func (f *PdfFunctionType3) Evaluate(x []float64) ([]float64, error) {
	if len(x) != 1 {
		common.Log.Error("Only one input allowed")
		return nil, errors.New("range check")
	}

	// Determine which function to use

	// Encode

	return nil, errors.New("not implemented yet")
}

func newPdfFunctionType3FromPdfObject(obj core.PdfObject) (*PdfFunctionType3, error) {
	fun := &PdfFunctionType3{}

	var dict *core.PdfObjectDictionary
	if indObj, is := obj.(*core.PdfIndirectObject); is {
		d, ok := indObj.PdfObject.(*core.PdfObjectDictionary)
		if !ok {
			return nil, errors.New("type check error")
		}
		fun.container = indObj
		dict = d
	} else if d, is := obj.(*core.PdfObjectDictionary); is {
		dict = d
	} else {
		return nil, errors.New("type check error")
	}

	// Domain
	array, has := core.TraceToDirectObject(dict.Get("Domain")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Domain not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	if array.Len() != 2 {
		common.Log.Error("Domain invalid")
		return nil, errors.New("invalid domain range")
	}
	domain, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Domain = domain

	// Range
	array, has = core.TraceToDirectObject(dict.Get("Range")).(*core.PdfObjectArray)
	if has {
		if array.Len() < 0 || array.Len()%2 != 0 {
			return nil, errors.New("invalid range")
		}
		rang, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Range = rang
	}

	// Functions.
	array, has = core.TraceToDirectObject(dict.Get("Functions")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Functions not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	fun.Functions = []PdfFunction{}
	for _, obj := range array.Elements() {
		subf, err := newPdfFunctionFromPdfObject(obj)
		if err != nil {
			return nil, err
		}
		fun.Functions = append(fun.Functions, subf)
	}

	// Bounds
	array, has = core.TraceToDirectObject(dict.Get("Bounds")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Bounds not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	bounds, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Bounds = bounds
	if len(fun.Bounds) != len(fun.Functions)-1 {
		common.Log.Error("Bounds (%d) and num functions (%d) not matching", len(fun.Bounds), len(fun.Functions))
		return nil, errors.New("range check")
	}

	// Encode.
	array, has = core.TraceToDirectObject(dict.Get("Encode")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Encode not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	encode, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Encode = encode
	if len(fun.Encode) != 2*len(fun.Functions) {
		common.Log.Error("Len encode (%d) and num functions (%d) not matching up", len(fun.Encode), len(fun.Functions))
		return nil, errors.New("range check")
	}

	return fun, nil
}

// ToPdfObject returns the PDF representation of the function.
func (f *PdfFunctionType3) ToPdfObject() core.PdfObject {
	dict := core.MakeDict()

	dict.Set("FunctionType", core.MakeInteger(3))

	// Domain (required).
	domainArray := &core.PdfObjectArray{}
	for _, val := range f.Domain {
		domainArray.Append(core.MakeFloat(val))
	}
	dict.Set("Domain", domainArray)

	// Range (required).
	if f.Range != nil {
		rangeArray := &core.PdfObjectArray{}
		for _, val := range f.Range {
			rangeArray.Append(core.MakeFloat(val))
		}
		dict.Set("Range", rangeArray)
	}

	// Functions
	if f.Functions != nil {
		fArray := &core.PdfObjectArray{}
		for _, fun := range f.Functions {
			fArray.Append(fun.ToPdfObject())
		}
		dict.Set("Functions", fArray)
	}

	// Bounds.
	if f.Bounds != nil {
		bArray := &core.PdfObjectArray{}
		for _, val := range f.Bounds {
			bArray.Append(core.MakeFloat(val))
		}
		dict.Set("Bounds", bArray)
	}

	// Encode.
	if f.Encode != nil {
		eArray := &core.PdfObjectArray{}
		for _, val := range f.Encode {
			eArray.Append(core.MakeFloat(val))
		}
		dict.Set("Encode", eArray)
	}

	// Wrap in a container if we have one already specified.
	if f.container != nil {
		f.container.PdfObject = dict
		return f.container
	}

	return dict
}

// PdfFunctionType4 is a Postscript calculator functions.
type PdfFunctionType4 struct {
	Domain  []float64
	Range   []float64
	Program *ps.PSProgram

	executor    *ps.PSExecutor
	decodedData []byte

	container *core.PdfObjectStream
}

// Evaluate runs the function. Input is [x1 x2 x3].
func (f *PdfFunctionType4) Evaluate(xVec []float64) ([]float64, error) {
	if f.executor == nil {
		f.executor = ps.NewPSExecutor(f.Program)
	}

	var inputs []ps.PSObject
	for _, val := range xVec {
		inputs = append(inputs, ps.MakeReal(val))
	}

	outputs, err := f.executor.Execute(inputs)
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
func newPdfFunctionType4FromStream(stream *core.PdfObjectStream) (*PdfFunctionType4, error) {
	fun := &PdfFunctionType4{}

	fun.container = stream

	dict := stream.PdfObjectDictionary

	// Domain
	array, has := core.TraceToDirectObject(dict.Get("Domain")).(*core.PdfObjectArray)
	if !has {
		common.Log.Error("Domain not specified")
		return nil, errors.New("required attribute missing or invalid")
	}
	if array.Len()%2 != 0 {
		common.Log.Error("Domain invalid")
		return nil, errors.New("invalid domain range")
	}
	domain, err := array.ToFloat64Array()
	if err != nil {
		return nil, err
	}
	fun.Domain = domain

	// Range
	array, has = core.TraceToDirectObject(dict.Get("Range")).(*core.PdfObjectArray)
	if has {
		if array.Len() < 0 || array.Len()%2 != 0 {
			return nil, errors.New("invalid range")
		}
		rang, err := array.ToFloat64Array()
		if err != nil {
			return nil, err
		}
		fun.Range = rang
	}

	// Program. Decode the program and parse the PS code.
	decoded, err := core.DecodeStream(stream)
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

// ToPdfObject returns the PDF representation of the function.
func (f *PdfFunctionType4) ToPdfObject() core.PdfObject {
	container := f.container
	if container == nil {
		f.container = &core.PdfObjectStream{}
		container = f.container
	}

	dict := core.MakeDict()
	dict.Set("FunctionType", core.MakeInteger(4))

	// Domain (required).
	domainArray := &core.PdfObjectArray{}
	for _, val := range f.Domain {
		domainArray.Append(core.MakeFloat(val))
	}
	dict.Set("Domain", domainArray)

	// Range (required).
	rangeArray := &core.PdfObjectArray{}
	for _, val := range f.Range {
		rangeArray.Append(core.MakeFloat(val))
	}
	dict.Set("Range", rangeArray)

	if f.decodedData == nil && f.Program != nil {
		// Update data.  This is used for created functions (not parsed ones).
		f.decodedData = []byte(f.Program.String())
	}

	// TODO: Encode.
	// Either here, or automatically later on when writing out.
	dict.Set("Length", core.MakeInteger(int64(len(f.decodedData))))

	container.Stream = f.decodedData
	container.PdfObjectDictionary = dict

	return container
}
