/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

import (
	"errors"
	"io"

	"github.com/unidoc/unipdf/common"
	"github.com/unidoc/unipdf/internal/jbig2/bitmap"
)

// CoderDebugging is the variable used to debug the encoder.
var CoderDebugging bool

// Encoder is the jbig2 arithmetic encoder context.
type Encoder struct {
	c     uint32
	a     uint16
	ct, b uint8
	bp    int

	ec int

	// the list of output chunks, not including the current one
	outputChunks [][]byte

	// current output chunk
	outbuf []byte

	// number of bytes used in outbuf
	outbufUsed int
	context    *codingContext
	intctx     [13]*codingContext
	iaidctx    *codingContext
}

// Init initializaes a new context
func (e *Encoder) Init() {
	e.context = newContext(maxCtx)
	e.a = 0x8000
	e.c = 0
	e.ct = 12
	e.bp = -1
	e.b = 0
	e.outbufUsed = 0
	e.outbuf = make([]byte, outputBufferSize)
	for i := 0; i < len(e.intctx); i++ {
		e.intctx[i] = newContext(512)
	}

	e.iaidctx = nil
}

const (
	tpgdCTX = 0x9b25
)

// EncodeBitmap encodes packed data. Designed for the Leptonica's 1bpp packed format image.
func (e *Encoder) EncodeBitmap(bm *bitmap.Bitmap, mx, my int, duplicateLineRemoval bool) error {
	var ltp, sltp uint8

	for y := 0; y < bm.Height; y++ {
		x := 0
		var (
			c1, c2, c3 uint16
			b1, b2, b3 byte
		)

		if y >= 2 {
			b1 = bm.Data[(y-2)*bm.RowStride]
		}
		if y >= 1 {
			b2 = bm.Data[(y-1)*bm.RowStride]

			if duplicateLineRemoval {
				if bm.Data[y*bm.RowStride] == bm.Data[(y-1)*bm.RowStride] {
					sltp = ltp ^ 1
					ltp = 1
				} else {
					sltp = ltp
					ltp = 0
				}
			}
		}

		if duplicateLineRemoval {
			if err := e.encodeBit(e.context, tpgdCTX, sltp); err != nil {
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
			tval := uint32(c1<<11 | c2<<4 | c3)
			v := (b3 & 0x80) >> 7

			err := e.encodeBit(e.context, tval, v)
			if err != nil {
				return err
			}

			c1 <<= 1
			c2 <<= 1
			c3 <<= 1

			c1 |= uint16((b1 & 0x80) >> 7)
			c2 |= uint16((b2 & 0x80) >> 7)
			c3 = uint16(v)

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

// Flush all the data stored in a context
func (e *Encoder) Flush() {
	e.outbufUsed = 0
	e.outputChunks = nil
	e.bp = -1
}

// Refine encodes the refinement of an exemplar to a bitmap.
// It encodes the differences between the template and the target image.
// The values 'ox, oy' are limited to [-1, 0, 1].
func (e *Encoder) Refine(itempl, itarget *bitmap.Bitmap, ox, oy int) error {
	for y := 0; y < itarget.Height; y++ {
		var x int
		temply := y + oy
		var (
			c1, c2, c3, c4, c5 uint16
			b1, b2, b3, b4, b5 byte
		)
		if temply >= 1 && (temply-1) < itempl.Height {
			b1 = itempl.Data[(temply-1)*itempl.RowStride]
		}

		if temply >= 0 && temply < itempl.Height {
			b2 = itempl.Data[temply*itempl.RowStride]
		}

		if temply >= -1 && temply+1 < itempl.Height {
			b3 = itempl.Data[(temply+1)*itempl.RowStride]
		}

		if y >= 1 {
			b4 = itarget.Data[(y-1)*itarget.RowStride]
		}
		b5 = itarget.Data[y*itarget.RowStride]

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

		for x = 0; x < itarget.Width; x++ {
			tval := (c1 << 10) | (c2 << 7) | (c3 << 4) | (c4 << 1) | c5
			v := b5 >> 7

			err := e.encodeBit(e.context, uint32(tval), v)
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
				if wordNo < itempl.RowStride && temply >= 1 && (temply-1) < itempl.Height {
					b1 = itempl.Data[(temply-1)*itempl.RowStride+wordNo]
				}
				if wordNo < itempl.RowStride && temply >= 0 && temply < itempl.Height {
					b2 = itempl.Data[temply*itempl.RowStride+wordNo]
				}
				if wordNo < itempl.RowStride && temply >= -1 && (temply+1) < itempl.Height {
					b3 = itempl.Data[(temply+1)*itempl.RowStride+wordNo]
				}
			} else {
				b1 <<= 1
				b2 <<= 1
				b3 <<= 1
			}

			if m == 5 && y >= 1 {
				b4 = 0
				if wordNo < itarget.RowStride {
					b4 = itarget.Data[(y-1)*itarget.RowStride+wordNo]
				}
			} else {
				b4 <<= 1
			}

			if m == 7 {
				b5 = 0
				if wordNo < itarget.RowStride {
					b5 = itarget.Data[y*itarget.RowStride+wordNo]
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
	e.iaidctx = nil
	e.context = newContext(maxCtx)
}

// Final flush any remaining arithmetic encoder context to the output
func (e *Encoder) Final() {
	e.flush()
}

func (e *Encoder) byteout() {
	if e.b == 0xff {
		e.rblock()
		return
	}

	if e.c < 0x8000000 {
		e.lblock()
		return
	}
	e.b++

	if e.b != 0xff {
		e.lblock()
		return
	}
	e.c &= 0x7ffffff
	e.rblock()
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

// returns the number of ybtes of output in the given context
func (e *Encoder) datasize() uint {
	return uint(outputBufferSize*len(e.outputChunks) + e.outbufUsed)
}

// emit a byte from the compressor by appending to the current
// output buffer. If the buffer is full, allocate a new one.
func (e *Encoder) emit() {
	if CoderDebugging {
		common.Log.Debug("Emit: %02X", e.b)
	}
	if e.outbufUsed == outputBufferSize {
		e.outputChunks = append(e.outputChunks, e.outbuf)
		e.outbuf = make([]byte, outputBufferSize)
		e.outbufUsed = 0
	}

	e.outbuf[e.outbufUsed] = e.b
	e.outbufUsed++
}

func (e *Encoder) encodeBit(ctx *codingContext, ctxNum uint32, d uint8) error {
	// NOTE: jbig2arith.cc:238
	e.ec++

	if ctxNum >= uint32(len(ctx.context)) {
		return errors.New("arithmetic encoder - invalid ctx number")
	}

	i := ctx.context[ctxNum]
	mps := ctx.mps(ctxNum)
	qe := stateTable[i].qe

	if CoderDebugging {
		common.Log.Debug("EC: %d\t D: %d\t I: %d\t MPS: %d\t QE: %04X\t  A: %04X\t C: %08X\t CT: %d\t B: %02X\t BP: %d", e.ec, d, i, mps, qe, e.a, e.c, e.ct, e.b, e.bp)
	}

	if d == 0 {
		e.code0(ctx, ctxNum, qe, i)
	} else {
		e.code1(ctx, ctxNum, qe, i)
	}
	return nil
}

func (e *Encoder) encodeInteger(proc Class, value int) error {
	// NOTE: jbig2arith.cc:381
	ctx := e.intctx[proc]

	var i int

	if value > 2000000000 || value < -2000000000 {
		return errors.New("arithmetic encoder - invalid integer value")
	}

	prev := uint32(1)

	for {
		if intEncRange[i].bot <= value && intEncRange[i].top >= value {
			break
		}
		i++
	}

	if value < 0 {
		value = -value
	}
	value -= int(intEncRange[i].delta)

	data := uint8(intEncRange[i].data)
	for j := uint8(0); j < intEncRange[i].bits; j++ {
		v := data & 1
		if err := e.encodeBit(ctx, prev, v); err != nil {
			return err
		}

		data >>= 1
		if prev&0x100 > 0 {
			prev = (((prev << 1) | uint32(v)) & 0x1ff) | 0x100
		} else {
			prev = (prev << 1) | uint32(v)
		}
	}
	return nil
}

func (e *Encoder) encodeIAID(symcodelen, value int) error {
	// NOTE: jbig2arith.cc:426
	if e.iaidctx == nil {
		e.iaidctx = newContext(1 << uint(symcodelen))
	}

	mask := uint32(1<<uint32(symcodelen+1)) - 1
	value <<= uint(32 - symcodelen)
	prev := uint32(1)

	for i := 0; i < symcodelen; i++ {
		tval := prev & mask
		v := uint8((value & 0x80000000) >> 31)

		if err := e.encodeBit(e.iaidctx, tval, v); err != nil {
			return err
		}
		prev = (prev << 1) | uint32(v)
		value <<= 1
	}
	return nil
}

func (e *Encoder) encodeOOB(proc Class) error {
	// NOTE: jbig2arith.cc:370
	ctx := e.intctx[proc]

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
	e.byteout()

	e.c <<= e.ct
	e.byteout()
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

func (e *Encoder) lblock() {
	if e.bp >= 0 {
		e.emit()
	}
	e.bp++
	e.b = uint8(e.c >> 19)
	e.c &= 0x7ffff
	e.ct = 8
}

func (e *Encoder) rblock() {
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
			e.byteout()
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

func (e *Encoder) toBuffer(w io.Writer) (int, error) {
	var total int
	for _, chunk := range e.outputChunks {
		n, err := w.Write(chunk)
		if err != nil {
			return n, err
		}
		total += n
	}

	buf := make([]byte, e.outbufUsed)
	copy(buf, e.outbuf)

	n, err := w.Write(buf)
	if err != nil {
		return n, err
	}
	total += n
	return total, nil
}
