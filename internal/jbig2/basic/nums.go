/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package basic

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// IntSlice is the integer slice that contains panic safe methods.
type IntSlice []int

// NewIntSlice creates new integer slice.
func NewIntSlice(i int) *IntSlice {
	sl := IntSlice(make([]int, i))
	return &sl
}

// Add adds the integer 'v' to the slice
func (i *IntSlice) Add(v int) error {
	if i == nil {
		return errors.Error("IntSlice.Add", "slice not defined")
	}
	*i = append(*i, v)
	return nil
}

// Copy creates a copy of given int slice.
func (i *IntSlice) Copy() *IntSlice {
	cp := IntSlice(make([]int, len(*i)))
	copy(cp, *i)
	return &cp
}

// Get gets the integer at 'index'.
// Returns error if the index is out of range or given integer doesn't exists.
func (i IntSlice) Get(index int) (int, error) {
	if index > len(i)-1 {
		return 0, errors.Errorf("IntSlice.Get", "index: %d out of range", index)
	}
	return i[index], nil
}

// Size returns the size of the int slice.
func (i IntSlice) Size() int {
	return len(i)
}

// NewNumSlice creates a new NumSlice pointer.
func NewNumSlice(i int) *NumSlice {
	arr := NumSlice(make([]float32, i))
	return &arr
}

// NumSlice is the slice of the numbers that has a panic safe API.
type NumSlice []float32

// Add adds the float32 'v' value.
func (n *NumSlice) Add(v float32) {
	*n = append(*n, v)
}

// AddInt adds the 'v' integer value to the num slice.
func (n *NumSlice) AddInt(v int) {
	*n = append(*n, float32(v))
}

// Get the float32 value at 'i' index. Returns error if the index 'i' is out of range.
func (n NumSlice) Get(i int) (float32, error) {
	if i < 0 || i > len(n)-1 {
		return 0, errors.Errorf("NumSlice.Get", "index: '%d' out of range", i)
	}
	return n[i], nil
}

// GetInt gets the integer value at the 'i' position.
// The functions return errors if the index 'i' is out of range.
// Returns '0' on error.
func (n NumSlice) GetInt(i int) (int, error) {
	const processName = "GetInt"
	if i < 0 || i > len(n)-1 {
		return 0, errors.Errorf(processName, "index: '%d' out of range", i)
	}
	v := n[i]
	return int(v + Sign(v)*0.5), nil
}

// GetIntSlice gets the slice of integers from the provided 'NumSlice' values.
func (n NumSlice) GetIntSlice() []int {
	sl := make([]int, len(n))
	for i, v := range n {
		sl[i] = int(v)
	}
	return sl
}
