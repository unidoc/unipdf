package arithmetic

import (
	"io"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
)

// Encoder is the jbig2 arithmetic encoder context.
type Encoder struct {
	c     uint32
	a     uint16
	ct, b uint8
	bp    int

	// the list of output chunks, not including the current one
	outputChunks [][]uint8

	// current output chunk
	outbuf []uint8

	// number of bytes used in outbuf
	outbufUsed int
	context    []uint8
	intctx     [13][512]uint8
	iaidctx    uint8
}

// Init initializaes a new context
func (e *Encoder) Init() {

}

// Flush all the data stored in a context
func (e *Encoder) Flush() {

}

// Reset the arithmetic coder back to a init state
func (e *Encoder) Reset() {

}

// Final flush any remaining arithmetic encoder context to the output
func (e *Encoder) Final() {

}

// returns the number of ybtes of output in the given context
func (e *Encoder) datasize() uint {
	return 0
}

func (e *Encoder) encodeInteger(proc Class, value int) error {
	// TODO: jbig2arith.cc:381
	return nil
}

func (e *Encoder) encodeIAID(len, value int) error {
	// TODO: jbig2arith.cc:426
	return nil
}

func (e *Encoder) encodeOOB(proc Class) error {
	// TODO: jbig2arith.cc:370
	return nil
}

func (e *Encoder) toBuffer(w io.Writer) error {
	return nil
}

// EncodeImage a bitmap with the arithmetic encoder
// duplicateLineRemoval if true TPGD is used
func EncodeImage(ctx *Encoder, data *bitmap.Bitmap, duplicateLineRemoval bool) {

}

// EncodeBitImage encodes packed data. Designed for the Leptonica's 1bpp packed format image.
// Each row is some number of 32 bit words
func EncodeBitImage(ctx *Encoder, data *bitmap.Bitmap, mx, my int, duplicateLineRemoval bool) {

}

// context is a single state for a single state of the adaptive arithmetic compressor.
type context struct {
	qe       uint16
	mps, lps uint8
}

// emit a byte from the compressor by appending to the current
// output buffer. If the buffer is full, allocate a new one.
func (e *Encoder) emit() {
	// TODO: jbig2arith.cc:185
}

func (e *Encoder) byteout() {
	// TODO: jbig2arith.cc:199
}

func (e *Encoder) encodeBit(contxt uint8, ctxNum uint32, d uint8) error {
	// TODO: jbig2arith.cc:238
	return nil
}

func (e *Encoder) enocdeFinal() {
	// TODO: jbig2arith.cc:306
}
