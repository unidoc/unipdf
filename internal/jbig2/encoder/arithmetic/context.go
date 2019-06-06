package arithmetic

import (
	"github.com/unidoc/unipdf/internal/jbig2/bitmap"
	"io"
)

// EncContext is the jbig2 arithmetic encoder context
type EncContext struct {
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

	context []uint8

	intctx [13][512]uint8

	iaidctx uint8
}

// Init initializaes a new context
func (c *EncContext) Init() {

}

// Flush all the data stored in a context
func (c *EncContext) Flush() {

}

// Reset the arithmetic coder back to a init state
func (c *EncContext) Reset() {

}

// Final flush any remaining arithmetic encoder context to the output
func (c *EncContext) Final() {

}

// returns the number of ybtes of output in the given context
func (c *EncContext) datasize() uint {
	return 0
}

func (c *EncContext) toBuffer(w io.Writer) error {
	return nil
}

func (c *EncContext) encodeInteger(proc encClass, value int) {

}

func (c *EncContext) encodeIAID(len, value int) {

}

func (c *EncContext) encodeOOB(proc encClass) {

}

// EncodeImage a bitmap with the arithmetic encoder
// duplicateLineRemoval if true TPGD is used
func EncodeImage(ctx *EncContext, data *bitmap.Bitmap, duplicateLineRemoval bool) {

}

// EncodeBitImage encodes packed data. Designed for the Leptonica's 1bpp packed format image.
// Each row is some number of 32 bit words
func EncodeBitImage(ctx *EncContext, data *bitmap.Bitmap, mx, my int, duplicateLineRemoval bool) {

}

type encClass int

// encoding classes
const (
	IAAI encClass = iota
	IADH
	IADS
	IADT
	IADW
	IAEX
	IAFS
	IAIT
	IARDH
	IARDW
	IARDX
	IARDY
	IARI
)
