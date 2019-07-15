/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package writer contains the data writer that implements
// io.Writer, io.ByteWriter and allows to write single bits.
// The writer WriteBit method might be used in a dual way.
// By default it writes next bit starting from the LSB - least
// significant bit. By creating the Writer with NewMSB function
// the writer would write all bits starting from the MSB - most
// significant bit.
package writer
