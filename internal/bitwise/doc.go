/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package bitwise provides the utils related to reading and writing bit, byte data in a buffer like manner.
// It defines the StreamReader interface that allows to read bit, bits, byte, bytes, integers change and get the stream
// position, align the bits.
// It also defines the data writer that implements io.Writer, io.ByteWriter and allows to write single bits.
// The writer WriteBit method might be used in a dual way. By default it writes next bit starting from the LSB - least
// significant bit. By creating the Writer with NewWriterMSB function the writer would write all bits starting from
// the MSB - most significant bit.
package bitwise
