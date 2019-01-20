package bitmap

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/mmr"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// Bitmap is the jbig2 bitmap representation
type Bitmap struct {
	Width, Height, Line int

	// BitmapNumber is the number
	BitmapNumber int

	// Data saves the bits data for the bitmap
	Data *BitSet

	// Decoder is the decoders used in the decode procedure
	Decoder *container.Decoder
}

// New creates new bitmap with the parameters as provided in the arguments
func New(width, height int, decoder *container.Decoder) *Bitmap {
	bm := &Bitmap{
		Width:   width,
		Height:  height,
		Decoder: decoder,
		Line:    (width + 7) >> 3,
		Data:    NewBitSet(width * height),
	}

	return bm
}

// Read reads the bitmap from the provided reader with respect to the
// provided flags 'useMMR', 'useSkip', 'typicalPrediciton' and template
func (b *Bitmap) Read(
	r *reader.Reader,
	useMMR bool,
	template int,
	typicalPrediction, useSkip bool,
	skipBitmap *Bitmap,
	AdaptiveTemplateX []int8, AdaptiveTemplateY []int8,
	mmrDataLength int,
) error {
	common.Log.Debug("[BITMAP][READ] begins")
	defer func() { common.Log.Debug("[BITMAP][READ] finished") }()

	if useMMR {
		b.Decoder.MMR.Reset()

		var (
			referenceLine []int = make([]int, b.Width+2)
			codingLine    []int = make([]int, b.Width+2)
		)
		codingLine[0], codingLine[1] = b.Width, b.Width
		for row := 0; row < b.Height; row++ {
			i := 0
			for ; codingLine[i] < b.Width; i++ {
				referenceLine[i] = codingLine[i]
			}
			referenceLine[i], referenceLine[i+1] = b.Width, b.Width

			var referenceI, codingI, a0 int

			for {
				code1, err := b.Decoder.MMR.Get2DCode(r)
				if err != nil {
					common.Log.Debug("MMR.Get2DCode(reader) failed. %v", err)
					return err
				}

				var code2, code3 int

				switch code1 {
				case mmr.TwoDimensionalPass:
					if referenceLine[referenceI] < b.Width {
						a0 = referenceLine[referenceI+1]
						referenceI += 2
					}

				case mmr.TwoDimensionalHorizontal:
					if (codingI & 1) != 0 {
						code1 = 0

						for {
							code3, err = b.Decoder.MMR.GetBlackCode(r)
							if err != nil {
								common.Log.Debug("MMR.GetBlackCode() - code1 failed. %v", err)
								return err
							}
							code1 += code3
							if code3 >= 64 {
								break
							}
						}
						code2 = 0
						for {
							code3, err = b.Decoder.MMR.GetWhiteCode(r)
							if err != nil {
								common.Log.Debug("MMR.GetWhiteCode() - code2 failed. %v", err)
								return err
							}
							code2 += code3
							if code3 >= 64 {
								break
							}
						}
					} else {
						code1 = 0
						for {
							code3, err = b.Decoder.MMR.GetWhiteCode(r)
							if err != nil {
								common.Log.Debug("MMR.GetWhiteCode() - code1 failed. %v", err)
								return err
							}
							code1 += code3
							if code3 >= 64 {
								break
							}
						}

						code2 = 0
						for {
							code3, err = b.Decoder.MMR.GetBlackCode(r)
							if err != nil {
								common.Log.Debug("MMR.GetBlackCode() - code2 failed. %v", err)
								return err
							}
							code2 += code3
							if code3 >= 64 {
								break
							}
						}
					}
					if code1 > 0 || code2 > 0 {
						var v int = a0 + code1
						codingLine[codingI] = v
						a0 = v
						codingI++

						v = a0 + code2
						codingLine[codingI] = v
						a0 = v
						codingI++

						for referenceLine[referenceI] <= a0 && referenceLine[referenceI] < b.Width {
							referenceI += 2
						}
					}

				case mmr.TwoDimensionalVertical0:
					var v int = referenceLine[referenceI]
					codingLine[codingI] = v
					a0 = v
					codingI++
					if referenceLine[referenceI] < b.Width {
						referenceI++
					}

				case mmr.TwoDimensionalVerticalR1:
					var v int = referenceLine[referenceI] + 1
					codingLine[codingI] = v
					a0 = v
					codingI++

					if referenceLine[referenceI] < b.Width {
						referenceI++
						for referenceLine[referenceI] <= a0 && referenceLine[referenceI] < b.Width {
							referenceI += 2
						}
					}

				case mmr.TwoDimensionalVerticalR2:
					var v int = referenceLine[referenceI] + 2
					codingLine[codingI] = v
					a0 = v
					codingI++

					if referenceLine[referenceI] < b.Width {
						referenceI++
						for referenceLine[referenceI] <= a0 && referenceLine[referenceI] < b.Width {
							referenceI += 2
						}
					}

				case mmr.TwoDimensionalVerticalR3:
					var v int = referenceLine[referenceI] + 3
					codingLine[codingI] = v
					a0 = v
					codingI++

					if referenceLine[referenceI] < b.Width {
						referenceI++
						for referenceLine[referenceI] <= a0 && referenceLine[referenceI] < b.Width {
							referenceI += 2
						}
					}

				case mmr.TwoDimensionalVerticalL1:
					var v int = referenceLine[referenceI] - 1
					codingLine[codingI] = v
					a0 = v
					codingI++

					if referenceI > 0 {
						referenceI--
					} else {
						referenceI++
					}

					for referenceLine[referenceI] <= a0 && referenceLine[referenceI] < b.Width {
						referenceI += 2
					}

				case mmr.TwoDimensionalVerticalL2:
					var v int = referenceLine[referenceI] - 2
					codingLine[codingI] = v
					a0 = v
					codingI++

					if referenceI > 0 {
						referenceI--
					} else {
						referenceI++
					}

					for referenceLine[referenceI] <= a0 && referenceLine[referenceI] < b.Width {
						referenceI += 2
					}

				case mmr.TwoDimensionalVerticalL3:
					var v int = referenceLine[referenceI] - 3
					codingLine[codingI] = v
					a0 = v
					codingI++

					if referenceI > 0 {
						referenceI--
					} else {
						referenceI++
					}

					for referenceLine[referenceI] <= a0 && referenceLine[referenceI] < b.Width {
						referenceI += 2
					}

				default:
					common.Log.Debug("Illegal code in MMR bitmap data")

				}

				// while condition
				if a0 < b.Width {
					break
				}
			}

			codingLine[codingI] = b.Width
			codingI++

			for j := 0; codingLine[j] < b.Width; j += 2 {
				for col := codingLine[j]; col < codingLine[j+1]; col++ {
					b.SetPixel(col, row, 1)
				}
			}
		}

		if mmrDataLength >= 0 {
			err := b.Decoder.MMR.SkipTo(r, mmrDataLength)
			if err != nil {
				return err
			}
		} else {
			b, err := b.Decoder.MMR.Get24Bits(r)
			if err != nil {
				return err
			}
			if b != 0x001001 {
				common.Log.Debug("Missing EOFB in JBIG2 MMR bitmap data")
			}
		}
	} else {
		cxPtr0, cxPtr1 := NewPointer(b), NewPointer(b)
		axPtr0, axPtr1 := NewPointer(b), NewPointer(b)
		axPtr2, axPtr3 := NewPointer(b), NewPointer(b)

		var ltpCX int64
		if typicalPrediction {
			switch template {
			case 0:
				ltpCX = 0x3953
			case 1:
				ltpCX = 0x079a
			case 2:
				ltpCX = 0x0e3
			case 3:
				ltpCX = 0x18a
			}
		}

		var (
			ltp               bool
			cx, cx0, cx1, cx2 int64
		)

		for row := 0; row < b.Height; row++ {
			if typicalPrediction {
				bit, err := b.Decoder.Arithmetic.DecodeBit(r, uint64(ltpCX), b.Decoder.Arithmetic.GenericRegionStats)
				if err != nil {
					return err
				}

				if bit != 0 {
					ltp = !ltp
				}

				if ltp {
					b.duplicateRow(row, row-1)
					continue
				}
			}

			var pixel int

			switch template {
			case 0:
				cxPtr0.SetPointer(0, row-2)
				cx0 = int64(cxPtr0.NextPixel() << 1)
				cx0 |= int64(cxPtr0.NextPixel())

				cxPtr1.SetPointer(0, row-1)
				cx1 = int64(cxPtr1.NextPixel() << 2)
				cx1 |= int64(cxPtr1.NextPixel() << 1)
				cx1 |= int64(cxPtr1.NextPixel())

				cx2 = 0

				axPtr0.SetPointer(int(AdaptiveTemplateX[0]), row+int(AdaptiveTemplateY[0]))
				axPtr1.SetPointer(int(AdaptiveTemplateX[1]), row+int(AdaptiveTemplateY[1]))
				axPtr2.SetPointer(int(AdaptiveTemplateX[2]), row+int(AdaptiveTemplateY[2]))
				axPtr3.SetPointer(int(AdaptiveTemplateX[3]), row+int(AdaptiveTemplateY[3]))

				for col := 0; col < b.Width; col++ {
					cx = ((cx0 << 13) | (cx1 << 8) | (cx2 << 4) | int64(axPtr0.NextPixel()<<3) | int64(axPtr1.NextPixel()<<2) | int64(axPtr2.NextPixel()<<1) | int64(axPtr3.NextPixel()))

					if useSkip && !skipBitmap.GetPixel(col, row) {
						pixel = 0
					} else {
						var err error
						pixel, err = b.Decoder.Arithmetic.DecodeBit(r, uint64(cx), b.Decoder.Arithmetic.GenericRegionStats)
						if err != nil {
							return err
						}
						if pixel != 0 {
							b.Data.Set(uint(row*b.Width+col), true)
						}
					}

					cx0 = ((cx0 << 1) | int64(cxPtr0.NextPixel()&0x07))
					cx1 = ((cx1 << 1) | int64(cxPtr1.NextPixel()&0x1f))
					cx2 = ((cx2 << 1) | int64(pixel)) & 0x0f
				}
			case 1:
				cxPtr0.SetPointer(0, row-2)
				cx0 = int64(cxPtr0.NextPixel() << 2)
				cx0 |= int64(cxPtr0.NextPixel() << 1)
				cx0 |= int64(cxPtr0.NextPixel())

				cxPtr1.SetPointer(0, row-1)
				cx1 = int64(cxPtr1.NextPixel() << 2)
				cx1 |= int64(cxPtr1.NextPixel() << 1)
				cx1 |= int64(cxPtr1.NextPixel())

				cx2 = 0

				axPtr0.SetPointer(int(AdaptiveTemplateX[0]), row+int(AdaptiveTemplateY[0]))

				for col := 0; col < b.Width; col++ {

					cx = ((cx0 << 9) | (cx1 << 4) | (cx2 << 1) | int64(axPtr0.NextPixel()))

					if useSkip && !skipBitmap.GetPixel(col, row) {
						pixel = 0
					} else {
						var err error
						pixel, err = b.Decoder.Arithmetic.DecodeBit(r, uint64(cx), b.Decoder.Arithmetic.GenericRegionStats)
						if err != nil {
							return err
						}
						if pixel != 0 {
							b.Data.Set(uint(row*b.Width+col), true)
						}
					}

					cx0 = ((cx0 << 1) | int64(cxPtr0.NextPixel()&0x0f))
					cx1 = ((cx1 << 1) | int64(cxPtr1.NextPixel()&0x1f))
					cx2 = ((cx2 << 1) | int64(pixel)) & 0x07
				}

			case 2:
				cxPtr0.SetPointer(0, row-2)
				cx0 = int64(cxPtr0.NextPixel() << 1)
				cx0 |= int64(cxPtr0.NextPixel())

				cxPtr1.SetPointer(0, row-1)
				cx1 = int64(cxPtr1.NextPixel() << 1)
				cx1 |= int64(cxPtr1.NextPixel())

				cx2 = 0

				axPtr0.SetPointer(int(AdaptiveTemplateX[0]), row+int(AdaptiveTemplateY[0]))

				for col := 0; col < b.Width; col++ {

					cx = ((cx0 << 7) | (cx1 << 3) | (cx2 << 1) | int64(axPtr0.NextPixel()))

					if useSkip && !skipBitmap.GetPixel(col, row) {
						pixel = 0
					} else {
						var err error
						pixel, err = b.Decoder.Arithmetic.DecodeBit(r, uint64(cx), b.Decoder.Arithmetic.GenericRegionStats)
						if err != nil {
							return err
						}
						if pixel != 0 {
							b.Data.Set(uint(row*b.Width+col), true)
						}
					}

					cx0 = ((cx0 << 1) | int64(cxPtr0.NextPixel()&0x07))
					cx1 = ((cx1 << 1) | int64(cxPtr1.NextPixel()&0x0f))
					cx2 = ((cx2 << 1) | int64(pixel)) & 0x03
				}
			case 3:

				cxPtr1.SetPointer(0, row-1)
				cx1 = int64(cxPtr1.NextPixel() << 1)
				cx1 |= int64(cxPtr1.NextPixel())

				cx2 = 0

				axPtr0.SetPointer(int(AdaptiveTemplateX[0]), row+int(AdaptiveTemplateY[0]))

				for col := 0; col < b.Width; col++ {

					cx = ((cx1 << 5) | (cx2 << 1) | int64(axPtr0.NextPixel()))

					if useSkip && !skipBitmap.GetPixel(col, row) {
						pixel = 0
					} else {
						var err error
						pixel, err = b.Decoder.Arithmetic.DecodeBit(r, uint64(cx), b.Decoder.Arithmetic.GenericRegionStats)
						if err != nil {
							return err
						}
						if pixel != 0 {
							b.Data.Set(uint(row*b.Width+col), true)
						}
					}

					cx1 = ((cx1 << 1) | int64(cxPtr1.NextPixel()&0x1f))
					cx2 = ((cx2 << 1) | int64(pixel)) & 0x0f
				}
			}

		}

	}
	return nil
}

// ReadGenericRefinementRegion reads the generic refinement region.
func (b *Bitmap) ReadGenericRefinementRegion(
	r *reader.Reader,
	template int,
	typicalPrediction bool,
	referred *Bitmap,
	referenceDX, referenceDY int,
	AdaptiveTemplateX []int8,
	AdaptiveTemplateY []int8,
) error {

	var cxPtr0, cxPtr1, cxPtr2, cxPtr3, cxPtr4, cxPtr5, cxPtr6,
		typPredicGenRefCxPtr0, typPredicGenRefCxPtr1, typPredicGenRefCxPtr2 *BMPointer

	var ltpCx int64

	if template != 0 {
		ltpCx = 0x008

		cxPtr0 = NewPointer(b)
		cxPtr1 = NewPointer(b)
		cxPtr2 = NewPointer(referred)
		cxPtr3 = NewPointer(referred)
		cxPtr4 = NewPointer(referred)
		cxPtr5 = NewPointer(b)
		cxPtr6 = NewPointer(b)
		typPredicGenRefCxPtr0 = NewPointer(referred)
		typPredicGenRefCxPtr1 = NewPointer(referred)
		typPredicGenRefCxPtr2 = NewPointer(referred)
	} else {
		ltpCx = 0x0010

		cxPtr0 = NewPointer(b)
		cxPtr1 = NewPointer(b)
		cxPtr2 = NewPointer(referred)
		cxPtr3 = NewPointer(referred)
		cxPtr4 = NewPointer(referred)
		cxPtr5 = NewPointer(b)
		cxPtr6 = NewPointer(referred)

		typPredicGenRefCxPtr0 = NewPointer(referred)
		typPredicGenRefCxPtr1 = NewPointer(referred)
		typPredicGenRefCxPtr2 = NewPointer(referred)

	}

	var (
		cx, cx0, cx2, cx3, cx4 int64

		typPredictGenRefCx0, typPredictGenRefCx1, typPredictGenRefCx2 int64

		ltp bool
	)

	for row := 0; row < b.Height; row++ {
		if template != 0 {

			cxPtr0.SetPointer(0, row-1)
			cx0 = int64(cxPtr0.NextPixel())

			cxPtr1.SetPointer(-1, row)

			cxPtr2.SetPointer(-referenceDX, row-1-referenceDY)

			cxPtr3.SetPointer(-1-referenceDX, row-referenceDY)
			cx3 = int64(cxPtr3.NextPixel())
			cx3 = (cx3 << 1) | int64(cxPtr3.NextPixel())

			cxPtr4.SetPointer(-referenceDX, row+1-referenceDY)
			cx4 = int64(cxPtr4.NextPixel())

			typPredictGenRefCx0, typPredictGenRefCx1, typPredictGenRefCx2 = 0, 0, 0

			if typicalPrediction {
				typPredicGenRefCxPtr0.SetPointer(-1-referenceDX, row-1-referenceDY)
				typPredictGenRefCx0 = int64(typPredicGenRefCxPtr0.NextPixel())
				typPredictGenRefCx0 = ((typPredictGenRefCx0 << 1) | int64(typPredicGenRefCxPtr0.NextPixel()))
				typPredictGenRefCx0 = ((typPredictGenRefCx0 << 1) | int64(typPredicGenRefCxPtr0.NextPixel()))

				typPredicGenRefCxPtr1.SetPointer(-1-referenceDX, row-1-referenceDY)
				typPredictGenRefCx1 = int64(typPredicGenRefCxPtr1.NextPixel())
				typPredictGenRefCx1 = ((typPredictGenRefCx1 << 1) | int64(typPredicGenRefCxPtr1.NextPixel()))
				typPredictGenRefCx1 = ((typPredictGenRefCx1 << 1) | int64(typPredicGenRefCxPtr1.NextPixel()))

				typPredicGenRefCxPtr2.SetPointer(-1-referenceDX, row-1-referenceDY)
				typPredictGenRefCx2 = int64(typPredicGenRefCxPtr2.NextPixel())
				typPredictGenRefCx2 = ((typPredictGenRefCx2 << 1) | int64(typPredicGenRefCxPtr2.NextPixel()))
				typPredictGenRefCx2 = ((typPredictGenRefCx2 << 1) | int64(typPredicGenRefCxPtr2.NextPixel()))
			}

			for col := 0; col < b.Width; col++ {

				cx0 = ((cx0 << 1) | int64(cxPtr0.NextPixel())) & 7
				cx3 = ((cx3 << 1) | int64(cxPtr3.NextPixel())) & 7
				cx4 = ((cx4 << 1) | int64(cxPtr4.NextPixel())) & 3

				if typicalPrediction {
					typPredictGenRefCx0 = ((typPredictGenRefCx0 << 1) | int64(typPredicGenRefCxPtr0.NextPixel())) & 7
					typPredictGenRefCx1 = ((typPredictGenRefCx1 << 1) | int64(typPredicGenRefCxPtr1.NextPixel())) & 7
					typPredictGenRefCx2 = ((typPredictGenRefCx2 << 1) | int64(typPredicGenRefCxPtr2.NextPixel())) & 7

					decodeBit, err := b.Decoder.Arithmetic.DecodeBit(r, uint64(ltpCx), b.Decoder.Arithmetic.RefinementRegionStats)
					if err != nil {
						return err
					}

					if decodeBit != 0 {
						ltp = !ltp
					}
					if typPredictGenRefCx0 == 0 && typPredictGenRefCx1 == 0 && typPredictGenRefCx2 == 0 {
						b.SetPixel(col, row, 0)
						continue
					} else if typPredictGenRefCx0 == 7 && typPredictGenRefCx1 == 7 && typPredictGenRefCx2 == 7 {
						b.SetPixel(col, row, 1)
						continue
					}
				}

				cx = (cx0 << 7) | int64(cxPtr1.NextPixel()<<6) |
					int64(cxPtr2.NextPixel()<<5) | int64(cx3<<2) | cx4

				pixel, err := b.Decoder.Arithmetic.DecodeBit(r, uint64(cx), b.Decoder.Arithmetic.RefinementRegionStats)
				if err != nil {
					return err
				}

				if pixel == 1 {
					b.Data.Set(uint(row*b.Width+col), true)
				}
			}

		} else {
			cxPtr0.SetPointer(0, row-1)
			cx0 = int64(cxPtr0.NextPixel())

			cxPtr1.SetPointer(-1, row)

			cxPtr2.SetPointer(-referenceDX, row-1-referenceDY)
			cx2 = int64(cxPtr2.NextPixel())

			cxPtr3.SetPointer(-1-referenceDX, row-referenceDY)
			cx3 = int64(cxPtr3.NextPixel())
			cx3 = (cx3 << 1) | int64(cxPtr3.NextPixel())

			cxPtr4.SetPointer(-1-referenceDX, row-referenceDY)
			cx4 = int64(cxPtr4.NextPixel())
			cx4 = (cx4 << 1) | int64(cxPtr4.NextPixel())

			cxPtr5.SetPointer(int(AdaptiveTemplateX[0]), row+int(AdaptiveTemplateY[0]))

			cxPtr6.SetPointer(int(AdaptiveTemplateX[1])-referenceDX, row+int(AdaptiveTemplateY[1])-referenceDY)

			if typicalPrediction {

				typPredicGenRefCxPtr0.SetPointer(-1-referenceDX, row-1-referenceDY)
				typPredictGenRefCx0 = int64(typPredicGenRefCxPtr0.NextPixel())
				typPredictGenRefCx0 = typPredictGenRefCx0 << 1
				typPredictGenRefCx0 = typPredictGenRefCx0 << 1

				typPredicGenRefCxPtr1.SetPointer(-1-referenceDX, row-1-referenceDY)
				typPredictGenRefCx1 = int64(typPredicGenRefCxPtr1.NextPixel())
				typPredictGenRefCx1 = typPredictGenRefCx1 << 1
				typPredictGenRefCx1 = typPredictGenRefCx1 << 1

				typPredicGenRefCxPtr2.SetPointer(-1-referenceDX, row-1-referenceDY)
				typPredictGenRefCx2 = int64(typPredicGenRefCxPtr2.NextPixel())
				typPredictGenRefCx2 = typPredictGenRefCx2 << 1
				typPredictGenRefCx2 = typPredictGenRefCx2 << 1

			}

			for col := 0; col < b.Width; col++ {
				cx0 = ((cx0 << 1) | int64(cxPtr0.NextPixel())) & 3
				cx2 = ((cx2 << 1) | int64(cxPtr2.NextPixel())) & 3
				cx3 = ((cx3 << 1) | int64(cxPtr3.NextPixel())) & 7
				cx4 = ((cx4 << 1) | int64(cxPtr4.NextPixel())) & 7

				if typicalPrediction {
					typPredictGenRefCx0 = ((typPredictGenRefCx0 << 1) | int64(typPredicGenRefCxPtr0.NextPixel())) & 7
					typPredictGenRefCx1 = ((typPredictGenRefCx1 << 1) | int64(typPredicGenRefCxPtr1.NextPixel())) & 7
					typPredictGenRefCx2 = ((typPredictGenRefCx2 << 1) | int64(typPredicGenRefCxPtr2.NextPixel())) & 7

					decodeBit, err := b.Decoder.Arithmetic.DecodeBit(r, uint64(ltpCx), b.Decoder.Arithmetic.RefinementRegionStats)
					if err != nil {
						return err
					}
					if decodeBit == 1 {
						ltp = !ltp
					}

					if typPredictGenRefCx0 == 0 && typPredictGenRefCx1 == 0 &&
						typPredictGenRefCx2 == 0 {
						b.SetPixel(col, row, 0)
						continue
					} else if typPredictGenRefCx0 == 7 && typPredictGenRefCx1 == 7 && typPredictGenRefCx2 == 7 {
						b.SetPixel(col, row, 1)
						continue
					}
				}

				cx = (cx0 << 11) | (int64(cxPtr1.NextPixel()) << 10) | (cx2 << 8) | (cx3 << 5) | (cx4 << 2) | (int64(cxPtr5.NextPixel()) << 1) | int64(cxPtr6.NextPixel())

				pixel, err := b.Decoder.Arithmetic.DecodeBit(r, uint64(cx), b.Decoder.Arithmetic.RefinementRegionStats)
				if err != nil {
					return err
				}

				if pixel == 1 {
					b.SetPixel(col, row, 1)
				}
			}

		}
	}

	return nil
}

// ReadTextRegion reads the provided text region and saves into the given bitmap
func (b *Bitmap) ReadTextRegion(
	r *reader.Reader,
	huffman, symbolRefine bool,
	symbolInstancesNumber, logStrips, symbolsNo uint,
	symbolCodeTable [][]int,
	symbolCodeLength int,
	symbols []*Bitmap,
	defaultPixel int, combinationOperator int64,
	transposed bool,
	referenceCorner, sOffset int,
	huffmanFSTable, huffmanDSTable, huffmanDTTable, huffmanRDWTable,
	huffmanRDHTable, huffmanRDXTable, huffmanRDYTable, huffmanRSizeTable [][]int,
	template int,
	symbolRegionAdaptiveTemplateX, symbolReigonAdaptiveTemplateY []int8,
) error {
	common.Log.Debug("[BITMAP][READ_TEXT_REGION] begins")
	defer func() { common.Log.Debug("[BITMAP][READ_TEXT_REGION] finished") }()

	var symbolBitmap *Bitmap

	var strips int = 1 << logStrips

	b.Clear(defaultPixel == 1)

	var (
		t   int
		err error
	)

	if huffman {
		t, _, err = b.Decoder.Huffman.DecodeInt(r, huffmanDTTable)
		if err != nil {
			common.Log.Debug("init Huffman.DecodeInt failed: %v ", err)
			return err
		}
	} else {
		t, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IadtStats)
		if err != nil {
			common.Log.Debug("init Arithmetic.DecodeInt(iadtStats) failed. %v", err)
			return err
		}
	}

	t *= -strips

	var (
		currentInstance, firstS, dt, tt, ds, s int
	)

	for uint(currentInstance) < symbolInstancesNumber {
		if huffman {
			dt, _, err = b.Decoder.Huffman.DecodeInt(r, huffmanDTTable)
			if err != nil {
				common.Log.Debug("CurrentInstance: %d Huffman.DecodeInt(DTTable) failed. %v", currentInstance, err)
				return err
			}
		} else {
			dt, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IadtStats)
			if err != nil {
				common.Log.Debug("CurrentInstance: %d Arithmetic.DecodeInt(IadtStats) failed: %v", currentInstance, err)
				return err
			}
		}
		t += dt * strips

		if huffman {
			ds, _, err = b.Decoder.Huffman.DecodeInt(r, huffmanFSTable)
			if err != nil {
				common.Log.Debug("CurrentInstance: %d Huffman.DecodeInt(FSTable) failed. %v", currentInstance, err)
				return err
			}
		} else {
			ds, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IafsStats)
			if err != nil {
				common.Log.Debug("CurrentInstance: %d Arithmetic.DecodeInt(IafsStats) failed: %v", currentInstance, err)
				return err
			}
		}
		firstS += ds
		s = firstS

		for {
			if strips == 1 {
				dt = 0
			} else if huffman {
				bits, err := r.ReadBits(byte(logStrips))
				if err != nil {
					common.Log.Debug("ReadBits for logStrips: %d failed: %v", logStrips, err)
					return err
				}
				dt = int(bits)
			} else {
				dt, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IaitStats)
				if err != nil {
					common.Log.Debug("Arithmetic.DecodeInt(IaitStats) failed: %v", err)
					return err
				}
			}

			tt = t + dt

			var symbolID int64

			if huffman {
				if symbolCodeTable != nil {
					symbol, _, err := b.Decoder.Huffman.DecodeInt(r, symbolCodeTable)
					if err != nil {
						common.Log.Debug("Huffman.DecodeInt(symbolCodeTable) failed: %v", err)
						return err
					}
					symbolID = int64(symbol)
				} else {
					symbols, err := r.ReadBits(byte(symbolCodeLength))
					if err != nil {
						common.Log.Debug("ReadBits(symbolCodeLength) failed: %v", err)
						return err
					}
					symbolID = int64(symbols)
				}
			} else {
				symbols, err := b.Decoder.Arithmetic.DecodeIAID(r, uint64(symbolCodeLength), b.Decoder.Arithmetic.IaidStats)
				if err != nil {
					common.Log.Debug("Arithmetic.DecodeIAID(symbolCodeLength) failed: %v", err)
					return err
				}
				symbolID = int64(symbols)
			}

			if symbolID >= int64(symbolsNo) {
				common.Log.Debug("Invalid symbol number: %d in JBIG2 text region.", symbolID)
			} else {
				symbolBitmap = nil

				var ri int
				if symbolRefine {
					if huffman {
						ri, err = r.ReadBit()
						if err != nil {
							common.Log.Debug("symbolRefine ReadBit() failed. %v", err)
							return err
						}
					} else {
						ri, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IariStats)
						if err != nil {
							common.Log.Debug("SymbolRefine Arithmetic->DecodeInt(IariStats) failed : %v ", err)
							return err
						}
					}
				} else {
					ri = 0
				}

				if ri != 0 {
					var refineDWidth, refineDHeight, refineDX, refineDY int

					if huffman {
						refineDWidth, _, err = b.Decoder.Huffman.DecodeInt(r, huffmanRDWTable)
						if err != nil {
							common.Log.Debug("refineDWidth Huffman.DecodeInt(huffmanRDWTable). %v", err)
							return err
						}

						refineDHeight, _, err = b.Decoder.Huffman.DecodeInt(r, huffmanRDHTable)
						if err != nil {
							common.Log.Debug("refineDHeight Huffman.DecodeInt(huffmanRDHTable). %v", err)
							return err
						}

						refineDX, _, err = b.Decoder.Huffman.DecodeInt(r, huffmanRDXTable)
						if err != nil {
							common.Log.Debug("refineDX Huffman.DecodeInt(huffmanRDXTable). %v", err)
							return err
						}

						refineDY, _, err = b.Decoder.Huffman.DecodeInt(r, huffmanRDYTable)
						if err != nil {
							common.Log.Debug("refineDY Huffman.DecodeInt(huffmanRDYTable). %v", err)
							return err
						}

						r.ConsumeRemainingBits()
						if err := b.Decoder.Arithmetic.Start(r); err != nil {
							common.Log.Debug("Arithmetic Decoder Start failed: %v", err)
							return err
						}
					} else {
						refineDWidth, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IardwStats)
						if err != nil {
							common.Log.Debug("refineDWidth Arithmetic.DecodeInt(IardwStats). %v", err)
							return err
						}

						refineDHeight, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IardhStats)
						if err != nil {
							common.Log.Debug("refineDHeight Arithmetic.DecodeInt(IardhStats). %v", err)
							return err
						}

						refineDX, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IardxStats)
						if err != nil {
							common.Log.Debug("refineDX Arithmetic.DecodeInt(IardxStats). %v", err)
							return err
						}

						refineDY, _, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IardyStats)
						if err != nil {
							common.Log.Debug("refineDY Arithmetic.DecodeInt(IardyStats). %v", err)
							return err
						}
					}

					var temp int
					if refineDWidth >= 0 {
						temp = refineDWidth
					} else {
						temp = refineDWidth - 1
					}

					refineDX = temp/2 + refineDX

					if refineDHeight >= 0 {
						temp = refineDHeight
					} else {
						temp = refineDHeight - 1
					}

					refineDY = temp/2 + refineDY

					symbolBitmap = New(refineDWidth+symbols[symbolID].Width, refineDHeight+symbols[symbolID].Height, b.Decoder)

					err = symbolBitmap.ReadGenericRefinementRegion(r, template, false, symbols[symbolID], refineDX, refineDY, symbolRegionAdaptiveTemplateX, symbolReigonAdaptiveTemplateY)
					if err != nil {
						common.Log.Debug("symbolBitmap.ReadGenericRefinementRegion failed: %v", err)
						return err
					}
				} else {
					symbolBitmap = symbols[symbolID]
				}

				var (
					bitmapWidth  int = symbolBitmap.Width - 1
					bitmapHeight int = symbolBitmap.Height - 1
				)

				if transposed {
					switch referenceCorner {
					case 0:
						b.Combine(symbolBitmap, tt, s, combinationOperator)
					case 1:
						b.Combine(symbolBitmap, tt, s, combinationOperator)
					case 2:
						b.Combine(symbolBitmap, (tt - bitmapWidth), s, combinationOperator)
					case 3:
						b.Combine(symbolBitmap, (tt - bitmapWidth), s, combinationOperator)
					}
					s += bitmapHeight
				} else {
					switch referenceCorner {
					case 0:
						b.Combine(symbolBitmap, (tt - bitmapHeight), s, combinationOperator)
					case 1:
						b.Combine(symbolBitmap, tt, s, combinationOperator)
					case 2:
						b.Combine(symbolBitmap, (tt - bitmapHeight), s, combinationOperator)
					case 3:
						b.Combine(symbolBitmap, tt, s, combinationOperator)
					}
					s += bitmapWidth
				}
			}
			currentInstance++

			var (
				intRes  int
				boolRes bool
			)

			if huffman {
				intRes, boolRes, err = b.Decoder.Huffman.DecodeInt(r, huffmanDSTable)
				if err != nil {
					common.Log.Debug("Huffman DecodeInt(huffmanDSTable) failed: %v", err)
					return err
				}
			} else {
				intRes, boolRes, err = b.Decoder.Arithmetic.DecodeInt(r, b.Decoder.Arithmetic.IadsStats)
				if err != nil {
					common.Log.Debug("Arithmetic DecodeInt(IadsStats) failed: %v", err)
					return err
				}
			}

			if !boolRes {
				break
			}

			ds = intRes
			s += sOffset + ds
		}

	}

	return nil
}

// Clear clears the bitmap according to the defPixel
func (b *Bitmap) Clear(defPixel bool) {
	b.Data.SetAll(defPixel)
}

// Combine combines the bitmap data with the provided bitmap with respect to the
// coordinates x and y and the provided comination operator 'combOp'
func (b *Bitmap) Combine(bitmap *Bitmap, x, y int, combOp int64) error {
	var (
		srcWidth  int = b.Width
		srcHeight int = b.Height
	)

	var minWidth int = srcWidth

	if (x + srcWidth) > b.Width {
		minWidth = b.Width - x
	}

	if (y + srcHeight) > b.Height {
		srcHeight = b.Height - y
	}

	var srcIndx int

	var indx int = y*b.Width + x

	if combOp == 0 {
		if x == 0 && y == 0 && srcHeight == b.Height && srcWidth == b.Width {
			for i := 0; i < len(b.Data.data); i++ {
				b.Data.data[i] |= bitmap.Data.data[i]
			}
		}
		for row := y; row < y+srcHeight; row++ {
			indx = row*b.Width + x
			b.Data.Or(uint(indx), bitmap.Data, uint(srcIndx), minWidth)
			srcIndx += srcWidth
		}
	} else if combOp == 1 {
		if x == 0 && y == 0 && srcHeight == b.Height && srcWidth == b.Width {
			for i := 0; i < len(b.Data.data); i++ {
				b.Data.data[i] &= bitmap.Data.data[i]
			}
		}
		for row := y; row < y+srcHeight; row++ {
			indx = row*b.Width + x
			for col := 0; col < minWidth; col++ {
				bl, err := bitmap.Data.Get(uint(indx + col))
				if err != nil {
					return err
				}
				ib, err := b.Data.Get(uint(indx))
				if err != nil {
					return err
				}
				b.Data.Set(uint(indx), bl && ib)

				indx++
			}

			srcIndx += srcWidth
		}
	} else if combOp == 2 {
		if x == 0 && y == 0 && srcHeight == b.Height && srcWidth == b.Width {
			for i := 0; i < len(b.Data.data); i++ {
				b.Data.data[i] ^= bitmap.Data.data[i]
			}
		}
		for row := y; row < y+srcHeight; row++ {
			indx = row*b.Width + x
			for col := 0; col < minWidth; col++ {
				bl, err := bitmap.Data.Get(uint(indx + col))
				if err != nil {
					return err
				}
				ib, err := b.Data.Get(uint(indx))
				if err != nil {
					return err
				}

				b.Data.Set(uint(indx), bl != ib)

				indx++
			}

			srcIndx += srcWidth
		}

	} else if combOp == 3 {
		for row := y; row < y+srcHeight; row++ {
			indx = row*b.Width + x
			for col := 0; col < minWidth; col++ {
				source, err := bitmap.Data.Get(uint(indx + col))
				if err != nil {
					return err
				}
				pixel, err := b.Data.Get(uint(indx))
				if err != nil {
					return err
				}

				b.Data.Set(uint(indx), source == pixel)

				indx++
			}

			srcIndx += srcWidth
		}
	} else if combOp == 4 {
		if x == 0 && y == 0 && srcHeight == b.Height && srcWidth == b.Width {
			for i := 0; i < len(b.Data.data); i++ {
				b.Data.data[i] = bitmap.Data.data[i]
			}
		}
	} else {

		for row := y; row < y+srcHeight; row++ {
			indx = row*b.Width + x

			for col := 0; col < minWidth; col++ {
				bl, err := bitmap.Data.Get(uint(srcIndx + col))
				if err != nil {
					return err
				}

				if err := b.Data.Set(uint(indx), bl); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// GetSlice gets the slice of the bitmap with the reference of the x,y starting point
// and the provided width and the height
func (b *Bitmap) GetSlice(x, y, width, height int) (*Bitmap, error) {
	slice := New(width, height, b.Decoder)

	var sliceIndx int

	for row := y; row < height; row++ {
		var indx int = row*b.Width + x

		for col := x; col < x+width; col++ {
			v, err := b.Data.Get(uint(indx))
			if err != nil {
				common.Log.Debug("GetSlice b.Data.Get(%d)", indx)
				return nil, err
			}
			if v {
				if err := slice.Data.Set(uint(sliceIndx), true); err != nil {
					common.Log.Debug("GetSlice slice.Data.Set(%d)", sliceIndx)
					return nil, err
				}
			}
			sliceIndx++
			indx++
		}
	}

	return slice, nil
}

func (b *Bitmap) GetPixel(col, row int) bool {
	p, err := b.Data.Get(uint(row*b.Width + col))
	if err != nil {
		panic(err)
	}
	return p
}

func (b *Bitmap) Expand(newHeight int, defaultPixel int) {
	newData := NewBitSet(newHeight * b.Width)
	for row := 0; row < b.Height; row++ {
		for col := 0; col < b.Width; col++ {
			b.setPixel(col, row, newData, b.GetPixel(col, row))
		}
	}

	b.Height = newHeight
	b.Data = newData
}

func (b *Bitmap) SetPixel(col, row int, value int) {
	b.setPixel(col, row, b.Data, value == 1)
}

func (b *Bitmap) duplicateRow(yDest, ySrc int) {
	for i := 0; i < b.Width; i++ {
		b.setPixel(i, yDest, b.Data, b.GetPixel(i, ySrc))
	}
}

func (b *Bitmap) setPixel(col, row int, data *BitSet, value bool) {
	index := uint(row*b.Width + col)
	data.Set(index, value)
}

func (b *Bitmap) GetData() []byte {
	var bytes []byte

	for i := 0; i < len(b.Data.data); i++ {

	}
	// binary.BigEndian.PutUint64(b, v)
	return bytes
}
