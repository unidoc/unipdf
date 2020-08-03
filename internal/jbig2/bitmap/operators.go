/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

// CombinationOperator is the operator used for combining the bitmaps.
type CombinationOperator int

const (
	// CmbOpOr is the 'OR' CombinationOperator.
	CmbOpOr CombinationOperator = iota
	// CmbOpAnd is the 'AND' CombinationOperator.
	CmbOpAnd
	// CmbOpXor is the 'XOR' CombinationOperator.
	CmbOpXor
	// CmbOpXNor is the 'XNOR' CombinationOperator.
	CmbOpXNor
	// CmbOpReplace is the 'REPLACE' CombinationOperator.
	CmbOpReplace
	// CmbOpNot is the 'NOT' CombinationOperator.
	CmbOpNot
)

// String implements Stringer interface.
func (c CombinationOperator) String() string {
	var result string
	switch c {
	case CmbOpOr:
		result = "OR"
	case CmbOpAnd:
		result = "AND"
	case CmbOpXor:
		result = "XOR"
	case CmbOpXNor:
		result = "XNOR"
	case CmbOpReplace:
		result = "REPLACE"
	case CmbOpNot:
		result = "NOT"
	}
	return result
}
