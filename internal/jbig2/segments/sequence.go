/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

// According to D.4.2. - File header bit 0
// defines the stream sequence organisation
const (
	ORandom     uint8 = 0
	OSequential uint8 = 1
)
