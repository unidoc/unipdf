/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

// CombinationOperator is the operator used while combining the bitmaps
type CombinationOperator int

const (
	// CmbOpOr the 'OR' operator
	CmbOpOr CombinationOperator = iota

	// CmbOpAnd the 'AND' operator
	CmbOpAnd

	// CmbOpXor the 'XOR' operator
	CmbOpXor

	// CmbOpXNor the 'XNOR' operator
	CmbOpXNor

	// CmbOpReplace the 'REPLACE' operator
	CmbOpReplace

	// CmbOpNot the 'NOT' operator
	CmbOpNot
)

// String implements Stringer interface
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
