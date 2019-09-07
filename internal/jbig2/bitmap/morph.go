/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"errors"
)

// MorphProcess is the combination of the morph operator with it's values.
type MorphProcess struct {
	Operation MorphOperation
	Arguments [4]int
}

func (p MorphProcess) getWidthHeight() (width, height int) {
	return p.Arguments[0], p.Arguments[1]
}

// MorphOperation is an enum that wraps the morph operations.
type MorphOperation int

// Enum morph operations.
const (
	MopDilation MorphOperation = iota
	MopErosion
	MopOpening
	MopClosing
	MopRankBinaryReduction
	MopReplicativeBinaryExpansion
	MopAddBorder
)

// Point is the basic structure that contains x, y float32 values.
// In compare with image.Point the x and y are floats not integers.
type Point struct {
	X, Y float32
}

// Centroid gets the centroid of the provided 'bm' bitmap.
// The parameters 'centTab' and 'sumTab' are optional.
// 'centTab' is a table for finding centroids.
// 'sumTab' is a table for finding pixel sums.
// The centorid point is relative to the UL corner.
func Centroid(bm *Bitmap, centTab, sumTab []int) (Point, error) {
	return bm.centroid(centTab, sumTab)
}

// Centroids gets the centroids relative to the UL corner of each pix for the provided bitmaps.
func Centroids(bms []*Bitmap) ([]Point, error) {
	pta := make([]Point, len(bms))
	centTab := makePixelCentroidTab8()
	sumTab := makePixelSumTab8()

	var err error
	for i, bm := range bms {
		pta[i], err = bm.centroid(centTab, sumTab)
		if err != nil {
			return nil, err
		}
	}
	return pta, nil
}

// DilateBrick dilate with all sel being hit.
func DilateBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	// TODO: leptonica/morph.c:684
	return nil, nil
}

// Dilate the source bitmap 's' using hits in the selection 'sel'.
// The 'd' destination bitmap is optional.
// The following cases are possible:
//	'd' == 's' 	the function writes the result back to 'src'.
// 	'd' == nil 	the function creates new bitmap and writes the result to it.
//	'd' != 's' 	puts the results into existing 'd'.
func Dilate(d *Bitmap, s *Bitmap, sel *Selection) (*Bitmap, error) {
	// NOTE: leptonica/morph.c:209
	var (
		t   *Bitmap
		err error
	)

	d, err = processMorphArgs1(d, s, sel, &t)
	if err != nil {
		return nil, err
	}
	if err = d.clearAll(); err != nil {
		return nil, err
	}
	var selData SelectionValue
	for i := 0; i < sel.Height; i++ {
		for j := 0; j < sel.Width; j++ {
			selData = sel.Data[i][j]
			if selData == 1 {
				if err = d.RasterOperation(j-sel.Cx, i-sel.Cy, s.Width, s.Height, PixSrcOrDst, t, 0, 0); err != nil {
					return nil, err
				}
			}
		}
	}
	return d, nil
}

// MakePixelSumTab8 creates table of integers that gives
// the number of 1 bits in the 8 bit index.
func MakePixelSumTab8() []int {
	return makePixelSumTab8()
}

// MakePixelCentroidTab8 creates table of integers gives
// the centroid weight of the 1 bits in the 8 bit index.
func MakePixelCentroidTab8() []int {
	return makePixelCentroidTab8()
}

// MorphSequence does the morph processes over the 'src' Bitmap with the provided sequence.
func MorphSequence(src *Bitmap, sequence ...MorphProcess) (*Bitmap, error) {
	return morphSequence(src, sequence...)
}

func (bm *Bitmap) centroid(centTab, sumTab []int) (Point, error) {
	pt := Point{}
	bm.setPadBits(0)
	if centTab == nil || len(centTab) == 0 {
		centTab = makePixelCentroidTab8()
	}
	if sumTab == nil || len(sumTab) == 0 {
		sumTab = makePixelSumTab8()
	}
	var xsum, ysum, pixsum, rowsum, i, j int
	var bt byte
	for i = 0; i < bm.Height; i++ {
		line := bm.RowStride * i
		rowsum = 0
		for j = 0; j < bm.RowStride; j++ {
			bt = bm.Data[line+j]
			if bt != 0 {
				rowsum += sumTab[bt]
				xsum += centTab[bt] + j*8*sumTab[bt]
			}
		}
		pixsum += rowsum
		ysum += rowsum * i
	}
	if pixsum != 0 {
		pt.X = float32(xsum) / float32(pixsum)
		pt.Y = float32(ysum) / float32(pixsum)
	}
	return pt, nil
}

func closeSafeBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	// TODO: leptonica/src/morph.c:949
	return nil, nil
}

func erodeBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	// TODO: leptonica/src/morph.c:748
	return nil, nil
}

func expandReplicate(s *Bitmap, factor int) (*Bitmap, error) {
	// TODO: leptonica/src/scale2.c:867
	return nil, nil
}

func makePixelCentroidTab8() []int {
	tab := make([]int, 256)
	tab[0] = 0
	tab[1] = 7
	var i int
	for i = 2; i < 4; i++ {
		tab[i] = tab[i-2] + 6
	}
	for i = 4; i < 8; i++ {
		tab[i] = tab[i-4] + 5
	}
	for i = 8; i < 16; i++ {
		tab[i] = tab[i-8] + 4
	}
	for i = 16; i < 32; i++ {
		tab[i] = tab[i-16] + 3
	}
	for i = 32; i < 64; i++ {
		tab[i] = tab[i-32] + 2
	}
	for i = 64; i < 128; i++ {
		tab[i] = tab[i-64] + 1
	}
	for i = 128; i < 256; i++ {
		tab[i] = tab[i-128]
	}
	return tab
}

func makePixelSumTab8() []int {
	tab := make([]int, 256)
	for i := 0; i <= 0xff; i++ {
		i := byte(i)
		tab[i] = int(i&0x1) + (int(i>>1) & 0x1) + (int(i>>2) & 0x1) + (int(i>>3) & 0x1) +
			(int(i>>4) & 0x1) + (int(i>>5) & 0x1) + (int(i>>6) & 0x1) + (int(i>>7) & 0x1)
	}
	return tab
}

func morphSequence(s *Bitmap, sequence ...MorphProcess) (*Bitmap, error) {
	var (
		bm1, bm2 *Bitmap
		err      error
	)

	if s == nil {
		return nil, errors.New("morphSequence source bitmap not defined")
	}
	if sequence == nil || len(sequence) == 0 {
		return nil, errors.New("morphSequence, sequence not defined")
	}

	bm1 = s.Copy()
	var w, h, border int
	for _, process := range sequence {
		switch process.Operation {
		case MopDilation:
			// 'd' character
			w, h = process.getWidthHeight()
			bm2, err = DilateBrick(nil, bm1, w, h)
			if err != nil {
				return nil, err
			}
		case MopErosion:
			// 'e' character
			w, h = process.getWidthHeight()
			bm2, err = erodeBrick(nil, bm1, w, h)
			if err != nil {
				return nil, err
			}
		case MopOpening:
			// 'o' character
			// TODO: o4.4
			w, h = process.getWidthHeight()
			bm2, err = openBrick(nil, bm1, w, h)
			if err != nil {
				return nil, err
			}
		case MopClosing:
			// 'c' character
			w, h = process.getWidthHeight()
			bm2, err = closeSafeBrick(nil, bm1, w, h)
			if err != nil {
				return nil, err
			}
		case MopRankBinaryReduction:
			// 'r' character
			bm2, err = reduceRankBinaryCascade(bm1, process.Arguments)
			if err != nil {
				return nil, err
			}
		case MopReplicativeBinaryExpansion:
			// 'x' character
			bm2, err = expandReplicate(bm1, process.Arguments[0])
			if err != nil {
				return nil, err
			}
		case MopAddBorder:
			// 'b' character
			border = process.Arguments[0]
			bm2, err = bm1.AddBorder(border, 0)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("invalid morphOperation provided to the sequence")
		}
		bm1 = bm2
	}

	if border > 0 {
		bm2, err = bm1.RemoveBorder(border)
		if err != nil {
			return nil, err
		}
	}
	bm1 = bm2

	return bm1, nil
}

func openBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	// TODO: leptonica/src/morph.c:812
	return nil, nil
}

// processMorphArgs1 used for generic erosion, dilation and HMT.
func processMorphArgs1(d *Bitmap, s *Bitmap, sel *Selection, t **Bitmap) (*Bitmap, error) {
	if s == nil {
		return nil, errors.New("MorphArgs1 's' not defined")
	}
	if sel == nil {
		return nil, errors.New("MorhpArgs1 'sel' not defined")
	}

	sx, sy := sel.Height, sel.Width
	if sx == 0 || sy == 0 {
		return nil, errors.New("selection of size 0")
	}

	if d == nil {
		d = New(d.Width, d.Height)
		*t = s
		return d, nil
	}
	d.Width = s.Width
	d.Height = s.Height
	d.RowStride = s.RowStride
	d.Color = s.Color
	d.Data = make([]byte, s.RowStride*s.Height)
	if d == s {
		*t = s.Copy()
	} else {
		*t = s
	}
	return d, nil
}

func reduceRankBinaryCascade(s *Bitmap, levels [4]int) (*Bitmap, error) {
	// TODO: leptonica/src/binreduce.c:148
	return nil, nil
}
