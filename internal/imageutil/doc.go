/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package imageutil provides utility functions, structures and interfaces that allows to operate on images.
// The function NewImage creates a new Image implementation based on provided input parameters.
// The functions ColorXXXAt allows to get the the colors from provided data and parameters without the need of creating
// image abstraction.
// Image converters allows to convert between different color spaces. As Image interface implements also image.Image
// it was possible to use standard golang image.Image as an input for the converters.
package imageutil
