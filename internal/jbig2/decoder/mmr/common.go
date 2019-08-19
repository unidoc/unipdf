/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package mmr

func maxInt(f, s int32) int32 {
	if f < s {
		return s
	}
	return f
}

func minInt(f, s int32) int32 {
	if f > s {
		return s
	}
	return f
}
