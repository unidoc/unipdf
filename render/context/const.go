package context

import "image/color"

// FillRule represents the fill style used by a context instance.
type FillRule int

// Fill rules.
const (
	FillRuleWinding FillRule = iota
	FillRuleEvenOdd
)

// LineCap represents the line cap style used by a context instance.
type LineCap int

// Line cap styles.
const (
	LineCapRound LineCap = iota
	LineCapButt
	LineCapSquare
)

// LineJoin represents the line join style used by a context instance.
type LineJoin int

// Line join styles.
const (
	LineJoinRound LineJoin = iota
	LineJoinBevel
)

// Pattern represents a pattern which can be rendered by a context instance.
type Pattern interface {
	ColorAt(x, y int) color.Color
}

// Gradient represents a gradient pattern which can be rendered by a context instance.
type Gradient interface {
	Pattern
	AddColorStop(offset float64, color color.Color)
}
