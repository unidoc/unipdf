/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

import (
	"bytes"
	"io"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// Encoder is the jbig2 arithmetic encoder context.
type Encoder struct {
	c     uint32
	a     uint16
	ct, b uint8
	bp    int
	ec    int

	// the list of output chunks, not including the current one
	outputChunks [][]byte
	// current output chunk
	outbuf []byte
	// number of bytes used in outbuf
	outbufUsed int
	context    *codingContext
	intCtx     [13]*codingContext
	iaidCtx    *codingContext
}

// New creates new initialized arithmetic encoder.
func New() *Encoder {
	e := &Encoder{}
	e.Init()
	return e
}

// Init initializes a new context.
func (e *Encoder) Init() {
	e.context = newContext(maxCtx)
	e.a = 0x8000
	e.c = 0
	e.ct = 12
	e.bp = -1
	e.b = 0
	e.outbufUsed = 0
	e.outbuf = make([]byte, outputBufferSize)
	for i := 0; i < len(e.intCtx); i++ {
		e.intCtx[i] = newContext(512)
	}
	e.iaidCtx = nil
}

// DataSize returns the size of encoded data
func (e *Encoder) DataSize() int {
	return e.dataSize()
}

const tpgdCTX = 0x9b25

// EncodeBitmap encodes packed data used for the 1bpp packed format image.
func (e *Encoder) EncodeBitmap(bm *bitmap.Bitmap, duplicateLineRemoval bool) error {
	common.Log.Trace("Encode Bitmap [%dx%d], %s", bm.Width, bm.Height, bm)
	var (
		ltp, sLtp                       uint8
		c1, c2, c3                      uint16
		b1, b2, b3                      byte
		x, thisLineIndex, lastLineIndex int
		thisLine, lastLine              []byte
	)

	for y := 0; y < bm.Height; y++ {
		// reset the bits context values for each row.
		// the temple used by the coder is fixed as '0' with the floating bits in the default locations.
		// reset the byte values which are the values of the upper rows.
		b1, b2 = 0, 0
		if y >= 2 {
			b1 = bm.Data[(y-2)*bm.RowStride]
		}
		if y >= 1 {
			b2 = bm.Data[(y-1)*bm.RowStride]
			// check if the last row was not the same as this one.
			if duplicateLineRemoval {
				// get this line slice data
				thisLineIndex = y * bm.RowStride
				thisLine = bm.Data[thisLineIndex : thisLineIndex+bm.RowStride]
				// get last line slice data
				lastLineIndex = (y - 1) * bm.RowStride
				lastLine = bm.Data[lastLineIndex : lastLineIndex+bm.RowStride]
				if bytes.Equal(thisLine, lastLine) {
					sLtp = ltp ^ 1
					ltp = 1
				} else {
					sLtp = ltp
					ltp = 0
				}
			}
		}

		if duplicateLineRemoval {
			if err := e.encodeBit(e.context, tpgdCTX, sLtp); err != nil {
				return err
			}
			if ltp != 0 {
				continue
			}
		}
		b3 = bm.Data[y*bm.RowStride]

		// top three bits are the start of the context
		c1 = uint16(b1 >> 5)
		c2 = uint16(b2 >> 4)

		// unused bits needs to be removed.
		b1 <<= 3
		b2 <<= 4
		c3 = 0

		for x = 0; x < bm.Width; x++ {
			tVal := uint32(c1<<11 | c2<<4 | c3)
			v := (b3 & 0x80) >> 7

			err := e.encodeBit(e.context, tVal, v)
			if err != nil {
				return err
			}

			c1 <<= 1
			c2 <<= 1
			c3 <<= 1

			c1 |= uint16((b1 & 0x80) >> 7)
			c2 |= uint16((b2 & 0x80) >> 7)
			c3 |= uint16(v)

			m := x % 8
			byteNo := x/8 + 1
			if m == 4 && y >= 2 {
				// roll in another byte from two lines up
				b1 = 0
				if byteNo < bm.RowStride {
					b1 = bm.Data[(y-2)*bm.RowStride+byteNo]
				}
			} else {
				b1 <<= 1
			}

			if m == 3 && y >= 1 {
				b2 = 0
				if byteNo < bm.RowStride {
					b2 = bm.Data[(y-1)*bm.RowStride+byteNo]
				}
			} else {
				b2 <<= 1
			}

			if m == 7 {
				b3 = 0
				if byteNo < bm.RowStride {
					b3 = bm.Data[y*bm.RowStride+byteNo]
				}
			} else {
				b3 <<= 1
			}

			c1 &= 31
			c2 &= 127
			c3 &= 15
		}
	}
	return nil
}

// EncodeIAID encodes the integer ID 'value'. The symbol code length is the binary length of the value.
func (e *Encoder) EncodeIAID(symbolCodeLength, value int) (err error) {
	common.Log.Trace("Encode IAID. SymbolCodeLength: '%d', Value: '%d'", symbolCodeLength, value)
	if err = e.encodeIAID(symbolCodeLength, value); err != nil {
		return errors.Wrap(err, "EncodeIAID", "")
	}
	return nil
}

// EncodeInteger encodes the integer 'value' for the given class 'proc'.
func (e *Encoder) EncodeInteger(proc Class, value int) (err error) {
	common.Log.Trace("Encode Integer:'%d' with Class: '%s'", value, proc)
	if err = e.encodeInteger(proc, value); err != nil {
		return errors.Wrap(err, "EncodeInteger", "")
	}
	return nil
}

// EncodeOOB encodes out of band value for given class process.
func (e *Encoder) EncodeOOB(proc Class) (err error) {
	common.Log.Trace("Encode OOB with Class: '%s'", proc)
	if err = e.encodeOOB(proc); err != nil {
		return errors.Wrap(err, "EncodeOOB", "")
	}
	return nil
}

// Final flush any remaining arithmetic encoder context to the output
func (e *Encoder) Final() {
	e.flush()
}

// Flush all the data stored in a context
func (e *Encoder) Flush() {
	e.outbufUsed = 0
	e.outputChunks = nil
	e.bp = -1
}

// Refine encodes the refinement of an exemplar to a bitmap.
// It encodes the differences between the template and the target image.
// The values 'ox, oy' are limited to [-1, 0, 1].
func (e *Encoder) Refine(iTemp, iTarget *bitmap.Bitmap, ox, oy int) error {
	for y := 0; y < iTarget.Height; y++ {
		var x int
		templateY := y + oy
		var (
			c1, c2, c3, c4, c5 uint16
			b1, b2, b3, b4, b5 byte
		)
		if templateY >= 1 && (templateY-1) < iTemp.Height {
			b1 = iTemp.Data[(templateY-1)*iTemp.RowStride]
		}

		if templateY >= 0 && templateY < iTemp.Height {
			b2 = iTemp.Data[templateY*iTemp.RowStride]
		}

		if templateY >= -1 && templateY+1 < iTemp.Height {
			b3 = iTemp.Data[(templateY+1)*iTemp.RowStride]
		}

		if y >= 1 {
			b4 = iTarget.Data[(y-1)*iTarget.RowStride]
		}
		b5 = iTarget.Data[y*iTarget.RowStride]

		shiftOffset := uint(6 + ox)
		c1 = uint16(b1 >> shiftOffset)
		c2 = uint16(b2 >> shiftOffset)
		c3 = uint16(b3 >> shiftOffset)

		c4 = uint16(b4 >> 6)

		bitsToTrim := uint(2 - ox)
		b1 <<= bitsToTrim
		b2 <<= bitsToTrim
		b3 <<= bitsToTrim

		b4 <<= 2

		for x = 0; x < iTarget.Width; x++ {
			tVal := (c1 << 10) | (c2 << 7) | (c3 << 4) | (c4 << 1) | c5
			v := b5 >> 7

			err := e.encodeBit(e.context, uint32(tVal), v)
			if err != nil {
				return err
			}

			c1 <<= 1
			c2 <<= 1
			c3 <<= 1
			c4 <<= 1

			c1 |= uint16(b1 >> 7)
			c2 |= uint16(b2 >> 7)
			c3 |= uint16(b3 >> 7)
			c4 |= uint16(b4 >> 7)
			c5 = uint16(v)

			m := x % 8
			wordNo := x/8 + 1
			if m == 5+ox {
				b1, b2, b3 = 0, 0, 0
				if wordNo < iTemp.RowStride && templateY >= 1 && (templateY-1) < iTemp.Height {
					b1 = iTemp.Data[(templateY-1)*iTemp.RowStride+wordNo]
				}
				if wordNo < iTemp.RowStride && templateY >= 0 && templateY < iTemp.Height {
					b2 = iTemp.Data[templateY*iTemp.RowStride+wordNo]
				}
				if wordNo < iTemp.RowStride && templateY >= -1 && (templateY+1) < iTemp.Height {
					b3 = iTemp.Data[(templateY+1)*iTemp.RowStride+wordNo]
				}
			} else {
				b1 <<= 1
				b2 <<= 1
				b3 <<= 1
			}

			if m == 5 && y >= 1 {
				b4 = 0
				if wordNo < iTarget.RowStride {
					b4 = iTarget.Data[(y-1)*iTarget.RowStride+wordNo]
				}
			} else {
				b4 <<= 1
			}

			if m == 7 {
				b5 = 0
				if wordNo < iTarget.RowStride {
					b5 = iTarget.Data[y*iTarget.RowStride+wordNo]
				}
			} else {
				b5 <<= 1
			}

			c1 &= 7
			c2 &= 7
			c3 &= 7
			c4 &= 7
		}
	}
	return nil
}

// Reset the arithmetic coder back to a init state
func (e *Encoder) Reset() {
	e.a = 0x8000
	e.c = 0
	e.ct = 12
	e.bp = -1
	e.b = 0
	e.iaidCtx = nil
	e.context = newContext(maxCtx)
}

// compile time check for the io.WriterTo interface.
var _ io.WriterTo = &Encoder{}

// WriteTo implements io.WriterTo interface.
func (e *Encoder) WriteTo(w io.Writer) (int64, error) {
	const processName = "Encoder.WriteTo"
	var total int64
	for i, chunk := range e.outputChunks {
		n, err := w.Write(chunk)
		if err != nil {
			return 0, errors.Wrapf(err, processName, "failed at i'th: '%d' chunk", i)
		}
		total += int64(n)
	}

	// remove unused bytes.
	e.outbuf = e.outbuf[:e.outbufUsed]
	n, err := w.Write(e.outbuf)
	if err != nil {
		return 0, errors.Wrap(err, processName, "buffered chunks")
	}
	total += int64(n)
	return total, nil
}

func (e *Encoder) byteOut() {
	if e.b == 0xff {
		e.rBlock()
		return
	}

	if e.c < 0x8000000 {
		e.lBlock()
		return
	}
	e.b++

	if e.b != 0xff {
		e.lBlock()
		return
	}
	e.c &= 0x7ffffff
	e.rBlock()
}

// code0 as defined in Figure E.5.
func (e *Encoder) code0(ctx *codingContext, ctxNum uint32, qe uint16, i byte) {
	if ctx.mps(ctxNum) == 0 {
		e.codeMPS(ctx, ctxNum, qe, i)
	} else {
		e.codeLPS(ctx, ctxNum, qe, i)
	}
}

// code1 as defined in Figure E.4.
func (e *Encoder) code1(ctx *codingContext, ctxNum uint32, qe uint16, i byte) {
	if ctx.mps(ctxNum) == 1 {
		e.codeMPS(ctx, ctxNum, qe, i)
	} else {
		e.codeLPS(ctx, ctxNum, qe, i)
	}
}

// codeMPS as defined in Figure E.7.
func (e *Encoder) codeMPS(ctx *codingContext, ctxNum uint32, qe uint16, i byte) {
	e.a -= qe

	if e.a&0x8000 != 0 {
		e.c += uint32(qe)
		return
	}

	if e.a < qe {
		e.a = qe
	} else {
		e.c += uint32(qe)
	}
	ctx.context[ctxNum] = stateTable[i].mps
	e.renormalize()
}

// codeLPS as defined in Figure E.6.
func (e *Encoder) codeLPS(ctx *codingContext, ctxNum uint32, qe uint16, i byte) {
	e.a -= qe
	if e.a < qe {
		e.c += uint32(qe)
	} else {
		e.a = qe
	}
	if stateTable[i].swtc == 1 {
		ctx.flipMps(ctxNum)
	}
	ctx.context[ctxNum] = stateTable[i].lps
	e.renormalize()
}

// returns the number of bytes of output in the given context
func (e *Encoder) dataSize() int {
	return outputBufferSize*len(e.outputChunks) + e.outbufUsed
}

// emit a byte from the compressor by appending to the current
// output buffer. If the buffer is full, allocate a new one.
func (e *Encoder) emit() {
	if e.outbufUsed == outputBufferSize {
		e.outputChunks = append(e.outputChunks, e.outbuf)
		e.outbuf = make([]byte, outputBufferSize)
		e.outbufUsed = 0
	}

	e.outbuf[e.outbufUsed] = e.b
	e.outbufUsed++
}

func (e *Encoder) encodeBit(ctx *codingContext, ctxNum uint32, d uint8) error {
	const processName = "Encoder.encodeBit"
	e.ec++

	if ctxNum >= uint32(len(ctx.context)) {
		return errors.Errorf(processName, "arithmetic encoder - invalid ctx number: '%d'", ctxNum)
	}

	i := ctx.context[ctxNum]
	mps := ctx.mps(ctxNum)
	qe := stateTable[i].qe
	common.Log.Trace("EC: %d\t D: %d\t I: %d\t MPS: %d\t QE: %04X\t  A: %04X\t C: %08X\t CT: %d\t B: %02X\t BP: %d", e.ec, d, i, mps, qe, e.a, e.c, e.ct, e.b, e.bp)

	if d == 0 {
		e.code0(ctx, ctxNum, qe, i)
	} else {
		e.code1(ctx, ctxNum, qe, i)
	}
	return nil
}

func (e *Encoder) encodeInteger(proc Class, value int) error {
	const processName = "Encoder.encodeInteger"
	if value > 2000000000 || value < -2000000000 {
		return errors.Errorf(processName, "arithmetic encoder - invalid integer value: '%d'", value)
	}

	ctx := e.intCtx[proc]
	prev := uint32(1)
	var i int

	for ; ; i++ {
		if intEncRange[i].bot <= value && intEncRange[i].top >= value {
			break
		}
	}
	if value < 0 {
		value = -value
	}
	value -= int(intEncRange[i].delta)

	data := intEncRange[i].data
	for j := uint8(0); j < intEncRange[i].bits; j++ {
		v := data & 1
		if err := e.encodeBit(ctx, prev, v); err != nil {
			return errors.Wrap(err, processName, "")
		}

		data >>= 1
		if prev&0x100 > 0 {
			prev = (((prev << 1) | uint32(v)) & 0x1ff) | 0x100
		} else {
			prev = (prev << 1) | uint32(v)
		}
	}

	// move the data value
	value <<= 32 - intEncRange[i].intBits
	for j := uint8(0); j < intEncRange[i].intBits; j++ {
		v := uint8((uint32(value) & 0x80000000) >> 31)
		if err := e.encodeBit(ctx, prev, v); err != nil {
			return errors.Wrap(err, processName, "move data to the top of word")
		}
		value <<= 1
		if prev&0x100 != 0 {
			prev = (((prev << 1) | uint32(v)) & 0x1ff) | 0x100
		} else {
			prev = (prev << 1) | uint32(v)
		}
	}
	return nil
}

func (e *Encoder) encodeIAID(symCodeLen, value int) error {
	if e.iaidCtx == nil {
		e.iaidCtx = newContext(1 << uint(symCodeLen))
	}

	mask := uint32(1<<uint32(symCodeLen+1)) - 1
	value <<= uint(32 - symCodeLen)
	prev := uint32(1)

	for i := 0; i < symCodeLen; i++ {
		tVal := prev & mask
		v := uint8((uint32(value) & 0x80000000) >> 31)

		if err := e.encodeBit(e.iaidCtx, tVal, v); err != nil {
			return err
		}
		prev = (prev << 1) | uint32(v)
		value <<= 1
	}
	return nil
}

func (e *Encoder) encodeOOB(proc Class) error {
	ctx := e.intCtx[proc]

	err := e.encodeBit(ctx, 1, 1)
	if err != nil {
		return err
	}

	err = e.encodeBit(ctx, 3, 0)
	if err != nil {
		return err
	}

	err = e.encodeBit(ctx, 6, 0)
	if err != nil {
		return err
	}

	err = e.encodeBit(ctx, 12, 0)
	if err != nil {
		return err
	}

	return nil
}

func (e *Encoder) flush() {
	e.setBits()

	e.c <<= e.ct
	e.byteOut()

	e.c <<= e.ct
	e.byteOut()
	e.emit()

	if e.b != 0xff {
		e.bp++
		e.b = 0xff
		e.emit()
	}

	e.bp++
	e.b = 0xac
	e.bp++
	e.emit()
}

func (e *Encoder) lBlock() {
	if e.bp >= 0 {
		e.emit()
	}
	e.bp++
	e.b = uint8(e.c >> 19)
	e.c &= 0x7ffff
	e.ct = 8
}

func (e *Encoder) rBlock() {
	if e.bp >= 0 {
		e.emit()
	}
	e.bp++
	e.b = uint8(e.c >> 20)
	e.c &= 0xfffff
	e.ct = 7
}

func (e *Encoder) renormalize() {
	for {
		e.a <<= 1
		e.c <<= 1
		e.ct--
		if e.ct == 0 {
			e.byteOut()
		}
		if (e.a & 0x8000) != 0 {
			break
		}
	}
}

func (e *Encoder) setBits() {
	tempC := e.c + uint32(e.a)
	e.c |= 0xffff
	if e.c >= tempC {
		e.c -= 0x8000
	}
}
