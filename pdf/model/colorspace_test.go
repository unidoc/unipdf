/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"fmt"
	"testing"
)

func TestSeparationCS1(t *testing.T) {
	rawObject := `
% Colour space
5 0 obj
[ /Separation /LogoGreen /DeviceCMYK 12 0 R ]
endobj
% Tint transformation function
12 0 obj
<<
	/FunctionType 4
	/Domain [0.0 1.0]
	/Range [ 0.0 1.0 0.0 1.0 0.0 1.0 0.0 1.0 ]
	/Length 65
>>
stream
{ dup 0.84 mul
exch 0.00 exch dup 0.44 mul exch 0.21 mul
}
endstream endobj
	`

	// Test a few lookups and see if it is accurate.
	// Test rgb conversion for a few specific values also.

	fmt.Println(rawObject)

	//t.Errorf("Test not implemented yet")
}

func TestDeviceNCS1(t *testing.T) {
	// Implement Example 3 on p. 172

	//t.Errorf("Test not implemented yet")
}
