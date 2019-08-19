/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package encoder

// Parameters contains encoder required parameters.
type Parameters struct {
	Width, Height int
	FullHeaders   bool

	Thresh       float64
	WeightFactor float64
}
