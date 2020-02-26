/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

// LocationFilter is predefined enum wrapper used for selection of boxes and bitmaps
type LocationFilter int

const (
	_ LocationFilter = iota
	// LocSelectWidth is the location filter where the width must satisfy constraint.
	LocSelectWidth
	// LocSelectHeight is the location filter where the height must satisfy constarint.
	LocSelectHeight
	// LocSelectXVal is the location filter where the 'x' value must satisfy constraint.
	LocSelectXVal
	// LocSelectYVal is the location filter where the 'y' value must satisfy constraint.
	LocSelectYVal
	// LocSelectIfEither is the location filter where either width or height can satisfy constraint.
	LocSelectIfEither
	// LocSelectIfBoth is the location filter where both width and height must satisfy constraint.
	LocSelectIfBoth
)

// SizeSelection is the predefined enum wrapper used for size selection for boxes and bitmaps.
type SizeSelection int

const (
	_ SizeSelection = iota
	// SizeSelectByWidth is the size select enum used for selecting by width.
	SizeSelectByWidth
	// SizeSelectByHeight is the size select enum used for selecting by height.
	SizeSelectByHeight
	// SizeSelectByMaxDimension is the size select enum used for selecting by max of width and height.
	SizeSelectByMaxDimension
	// SizeSelectByArea is the size select enum used for selecting by area.
	SizeSelectByArea
	// SizeSelectByPerimeter is the size select enum used for selecting by perimeter.
	SizeSelectByPerimeter
)

// SizeComparison is the predefined enum wrapper used for size comparison.
type SizeComparison int

const (
	_ SizeComparison = iota
	// SizeSelectIfLT is the size comparison used to save the value if it's less than threshold.
	SizeSelectIfLT
	// SizeSelectIfGT is the size comparison used to save the value if it's more than threshold.
	SizeSelectIfGT
	// SizeSelectIfLTE is the size comparison used to save the value if it's less or equal to threshold.
	SizeSelectIfLTE
	// SizeSelectIfGTE is the size comparison used to save the value if it's less more or equal to threshold.
	SizeSelectIfGTE
	// SizeSelectIfEQ is the size comparison used to save the values if it's equal to threshold.
	SizeSelectIfEQ
)
