package bitmap

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// Bitmap is the jbig2 bitmap representation
type Bitmap struct {
	Width, Height, Line int
	BitmapNumber        int

	// As the data is represented by the slice of bits
	Data []bool

	Decoder *container.Decoder
}

// New creates new bitmap with the parameters as provided in the arguments
func New(width, height int, decoder *container.Decoder) *Bitmap {
	bm := &Bitmap{
		Width:   width,
		Height:  height,
		Decoder: decoder,
		Line:    (width + 7) >> 3,
		Data:    make([]bool, width*height),
	}

	return bm
}

func (b *Bitmap) Read(
	template int,
	useMMR, typicalPrediction, useSkip bool,
	skipBitmap *Bitmap,
	ADTX []byte, ADTY []byte,
	mmrDataLength int,
) error {
	return nil
}

func (b *Bitmap) ReadGenericRefinementRegion(
	template int,
	typicalPrediction bool,
	referred *Bitmap,
	referenceDX, referenceDY int,
	ADTX []byte, ADTY []byte,
) error {
	return nil
}

func (b *Bitmap) ReadTextRegion(
	r *reader.Reader,
	huffman, symbolRefine bool,
	symbolInstances, logStrips, symbolNo int,
	symbolCodeTable [][]int,
	symbolCodeLength int,
	symbols []*Bitmap,
	defaultPixel, combinationOperator int,
	transposed bool,
	referenceCorner, sOffset int,
	huffmanFSTable, huffmanDSTable, huffmanDTTable, huffmanRDWTable,
	huffmanRDHTable, huffmanRDYTable, huffmanRSizeTable [][]int,
	template int,
	symbolRegionAdaptiveTemplateX, symbolReigonAdaptiveTemplateY []byte,

) error {
	return nil
}

// Clear clears the bitmap according to the defPixel
func (b *Bitmap) Clear(defPixel bool) {
	for i := range b.Data {
		b.Data[i] = defPixel
	}
}

func (b *Bitmap) Combine(bitmap *Bitmap, x, y int, combOp int64) {
	// var (
	// 	srcWidth  bitmap.Width
	// 	srcHeight bitmap.Height
	// )
	return
}

func (b *Bitmap) GetSlice(x, y, width, height int) *Bitmap {
	return nil
}

func (b *Bitmap) GetPixel(col, row int) bool {
	return b.Data[row*b.Width+col]
}

func (b *Bitmap) Expand(newHeight int, defaultPixel int) {

}

func (b *Bitmap) SetPixel(col, row int, value int) {
	b.setPixel(col, row, b.Data, value)
}

func (b *Bitmap) duplicateRow(yDest, ySrc int) {

}

func (b *Bitmap) setPixel(col, row int, data []bool, value int) {
	index := row*b.Width + col
	data[index] = value == 1
}
