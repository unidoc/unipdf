/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

type intEncRangeS struct {
	bot, top int // the range of numbers for which this is valid.
	// the bits of data to write first, and the number which are valid
	// These bits are taken from the bottom of the uint8, in reverse order.
	data, bits uint8
	// the amount of subtract from the value before encoding it.
	delta uint16
	// intBits number of bits to use to encode to the integer.
	intBits uint8
}

var intEncRange = []intEncRangeS{
	{0, 3, 0, 2, 0, 2},
	{-1, -1, 9, 4, 0, 0},
	{-3, -2, 5, 3, 2, 1},
	{4, 19, 2, 3, 4, 4},
	{-19, -4, 3, 3, 4, 4},
	{20, 83, 6, 4, 20, 6},
	{-83, -20, 7, 4, 20, 6},
	{84, 339, 14, 5, 84, 8},
	{-339, -84, 15, 5, 84, 8},
	{340, 4435, 30, 6, 340, 12},
	{-4435, -340, 31, 6, 340, 12},
	{4436, 2000000000, 62, 6, 4436, 32},
	{-2000000000, -4436, 63, 6, 4436, 32},
}
