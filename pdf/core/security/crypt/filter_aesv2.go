package crypt

import "fmt"

func init() {
	registerFilter("AESV2", newFilterAESV2)
}

// NewFilterAESV2 creates an AES-based filter with a 128 bit key (AESV2).
func NewFilterAESV2() Filter {
	f, err := newFilterAESV2(FilterDict{})
	if err != nil {
		panic(err)
	}
	return f
}

func newFilterAESV2(d FilterDict) (Filter, error) {
	if d.Length != 0 && d.Length != 16 {
		return nil, fmt.Errorf("Invalid AESV2 crypt filter length (%d)", d.Length)
	}
	return filterAESV2{}, nil
}

// filterAESV2 is an AES-based filter (128 bit key, PDF 1.6)
type filterAESV2 struct {
	filterAES
}

func (filterAESV2) PDFVersion() [2]int {
	return [2]int{1, 5}
}

func (filterAESV2) HandlerVersion() (V, R int) {
	V, R = 4, 4
	return
}

func (filterAESV2) Name() string {
	return "AESV2"
}

func (filterAESV2) KeyLength() int {
	return 128 / 8
}

func (filterAESV2) MakeKey(objNum, genNum uint32, ekey []byte) ([]byte, error) {
	return makeKeyV2(objNum, genNum, ekey, true)
}
