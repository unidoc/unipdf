package mmr

import (
	"fmt"
)

type code struct {
	bitLength      int
	codeWord       int
	runLength      int
	subTable       []*code
	nonNilSubTable bool
}

func newCode(codeData [3]int) *code {
	return &code{
		bitLength: codeData[0],
		codeWord:  codeData[1],
		runLength: codeData[2],
	}
}

// String returns code string
func (c *code) String() string {
	return fmt.Sprintf("%d/%d/%d", c.bitLength, c.codeWord, c.runLength)
}

func (c *code) equals(obj interface{}) bool {
	second, ok := obj.(*code)
	if !ok {
		return false
	}

	return c.bitLength == second.bitLength &&
		c.codeWord == second.codeWord &&
		c.runLength == second.runLength
}
