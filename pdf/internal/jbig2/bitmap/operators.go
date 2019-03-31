package bitmap

// Combination Operator is the operator used while combining the bitmaps
type CombinationOperator int

const (
	CmbOpOr CombinationOperator = iota
	CmbOpAnd
	CmbOpXor
	CmbOpXNor
	CmbOpReplace
	CmbOpNot
)

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
