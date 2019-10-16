package context

import "image/color"

// FillRule.
type FillRule int

const (
	FillRuleWinding FillRule = iota
	FillRuleEvenOdd
)

// LineCap.
type LineCap int

const (
	LineCapRound LineCap = iota
	LineCapButt
	LineCapSquare
)

// LineJoin.
type LineJoin int

const (
	LineJoinRound LineJoin = iota
	LineJoinBevel
)

// Pattern.
type Pattern interface {
	ColorAt(x, y int) color.Color
}
