// The below implementation of paeth algorithm is from the Go standard library image/png package, which
// is distributed with a BSD-style license (see ACKNOWLEDGEMENTS.md).

package core

// intSize is either 32 or 64.
const intSize = 32 << (^uint(0) >> 63)

func abs(x int) int {
	// m := -1 if x < 0. m := 0 otherwise.
	m := x >> (intSize - 1)

	// In two's complement representation, the negative number
	// of any number (except the smallest one) can be computed
	// by flipping all the bits and add 1. This is faster than
	// code with a branch.
	// See Hacker's Delight, section 2-4.
	return (x ^ m) - m
}

// paeth implements the Paeth filter function, as per the PNG specification.
func paeth(a, b, c uint8) uint8 {
	// This is an optimized version of the sample code in the PNG spec.
	// For example, the sample code starts with:
	//	p := int(a) + int(b) - int(c)
	//	pa := abs(p - int(a))
	// but the optimized form uses fewer arithmetic operations:
	//	pa := int(b) - int(c)
	//	pa = abs(pa)
	pc := int(c)
	pa := int(b) - pc
	pb := int(a) - pc
	pc = abs(pa + pb)
	pa = abs(pa)
	pb = abs(pb)
	if pa <= pb && pa <= pc {
		return a
	} else if pb <= pc {
		return b
	}
	return c
}
