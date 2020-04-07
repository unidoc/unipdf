/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// BoundaryCondition is the global enum variable used to define morph operation boundary conditions.
// More information about the definition could be found at: http://www.leptonica.org/binary-morphology.html#boundary-conditions
type BoundaryCondition int

const (
	// AsymmetricMorphBC defines the asymmetric boundary condition for morph functions.
	AsymmetricMorphBC BoundaryCondition = iota
	// SymmetricMorphBC defines the symmetric boundary condition for morph funcftions.
	SymmetricMorphBC
)

// MorphBC defines current morph boundary condition used by the morph functions.
// By default it is set to 'AsymetricMorphBC'.
var MorphBC BoundaryCondition

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

// MorphProcess is the combination of the morph operator with it's values.
type MorphProcess struct {
	Operation MorphOperation
	Arguments []int
}

var intlogBase2 = [5]int{1, 2, 3, 0, 4}

func (p MorphProcess) verify(i int, netRed, border *int) error {
	const processName = "MorphProcess.verify"
	switch p.Operation {
	case MopDilation, MopErosion, MopOpening, MopClosing:
		if len(p.Arguments) != 2 {
			return errors.Error(processName, "Operation: 'd', 'e', 'o', 'c' requires at least 2 arguments")
		}
		w, h := p.getWidthHeight()
		if w <= 0 || h <= 0 {
			return errors.Error(processName, "Operation: 'd', 'e', 'o', 'c'  requires both width and height to be >= 0")
		}
	case MopRankBinaryReduction:
		nred := len(p.Arguments)
		*netRed += nred
		if nred < 1 || nred > 4 {
			return errors.Error(processName, "Operation: 'r' requires at least 1 and at most 4 arguments")
		}
		for i := 0; i < nred; i++ {
			if p.Arguments[i] < 1 || p.Arguments[i] > 4 {
				return errors.Error(processName, "RankBinaryReduction level must be in range (0, 4>")
			}
		}
	case MopReplicativeBinaryExpansion:
		if len(p.Arguments) == 0 {
			return errors.Error(processName, "ReplicativeBinaryExpansion requires one argument")
		}
		fact := p.Arguments[0]
		if fact != 2 && fact != 4 && fact != 8 {
			return errors.Error(processName, "ReplicativeBinaryExpansion must be of factor in range {2,4,8}")
		}
		*netRed -= intlogBase2[fact/4]
	case MopAddBorder:
		if len(p.Arguments) == 0 {
			return errors.Error(processName, "AddBorder requires one argument")
		}
		fact := p.Arguments[0]
		if i > 0 {
			return errors.Error(processName, "AddBorder must be a first morph process")
		}
		if fact < 1 {
			return errors.Error(processName, "AddBorder value lower than 0")
		}
		*border = fact
	}
	return nil
}

func (p MorphProcess) getWidthHeight() (width, height int) {
	return p.Arguments[0], p.Arguments[1]
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
func Centroids(bms []*Bitmap) (*Points, error) {
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
	pts := Points(pta)
	return &pts, nil
}

// DilateBrick dilate with all sel being hit.
func DilateBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	return dilateBrick(d, s, hSize, vSize)
}

// Dilate the source bitmap 's' using hits in the selection 'sel'.
// The 'd' destination bitmap is optional.
// The following cases are possible:
//	'd' == 's' 	the function writes the result back to 'src'.
// 	'd' == nil 	the function creates new bitmap and writes the result to it.
//	'd' != 's' 	puts the results into existing 'd'.
func Dilate(d *Bitmap, s *Bitmap, sel *Selection) (*Bitmap, error) {
	return dilate(d, s, sel)
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

func (b *Bitmap) centroid(centTab, sumTab []int) (Point, error) {
	pt := Point{}
	b.setPadBits(0)
	if len(centTab) == 0 {
		centTab = makePixelCentroidTab8()
	}
	if len(sumTab) == 0 {
		sumTab = makePixelSumTab8()
	}
	var xsum, ysum, pixsum, rowsum, i, j int
	var bt byte
	for i = 0; i < b.Height; i++ {
		line := b.RowStride * i
		rowsum = 0
		for j = 0; j < b.RowStride; j++ {
			bt = b.Data[line+j]
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

var (
	tabExpand2x = makeExpandTab2x()
	tabExpand4x = makeExpandTab4x()
	tabExpand8x = makeExpandTab8x()
)

// closeBitmap does two morph functions in a row - a dilate and erode.
func closeBitmap(d, s *Bitmap, sel *Selection) (*Bitmap, error) {
	const processName = "closeBitmap"
	var err error
	if d, err = processMorphArgs2(d, s, sel); err != nil {
		return nil, err
	}
	// dilate first
	t, err := dilate(nil, s, sel)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	// then erode
	if _, err = erode(d, t, sel); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return d, nil
}

func closeBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	const processName = "closeBrick"
	if s == nil {
		return nil, errors.Error(processName, "source not defined")
	}
	if hSize < 1 || vSize < 1 {
		return nil, errors.Error(processName, "hSize and vSize not >= 1")
	}
	if hSize == 1 && vSize == 1 {
		return s.Copy(), nil
	}
	if hSize == 1 || vSize == 1 {
		sel := SelCreateBrick(vSize, hSize, vSize/2, hSize/2, SelHit)
		var err error
		d, err = closeBitmap(d, s, sel)
		if err != nil {
			return nil, errors.Wrap(err, processName, "hSize == 1 || vSize == 1")
		}
		return d, nil
	}
	selH := SelCreateBrick(1, hSize, 0, hSize/2, SelHit)
	selV := SelCreateBrick(vSize, 1, vSize/2, 0, SelHit)
	t, err := dilate(nil, s, selH)
	if err != nil {
		return nil, errors.Wrap(err, processName, "1st dilate")
	}
	if d, err = dilate(d, t, selV); err != nil {
		return nil, errors.Wrap(err, processName, "2nd dilate")
	}
	// erode
	if _, err = erode(t, d, selH); err != nil {
		return nil, errors.Wrap(err, processName, "1st erode")
	}
	if _, err = erode(d, t, selV); err != nil {
		return nil, errors.Wrap(err, processName, "2nd erode")
	}
	return d, nil
}

func closeSafeBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	const processName = "closeSafeBrick"
	if s == nil {
		return nil, errors.Error(processName, "source is nil")
	}
	if hSize < 1 || vSize < 1 {
		return nil, errors.Error(processName, "hsize and vsize not >= 1")
	}
	if hSize == 1 && vSize == 1 {
		return copyBitmap(d, s)
	}
	if MorphBC == SymmetricMorphBC {
		// Symmetric handles the pixels correctly
		bm, err := closeBrick(d, s, hSize, vSize)
		if err != nil {
			return nil, errors.Wrap(err, processName, "SymmetricMorphBC")
		}
		return bm, nil
	}

	maxTrans := max(hSize/2, vSize/2)
	bordSize := 8 * ((maxTrans + 7) / 8)

	bsb, err := s.AddBorder(bordSize, 0)
	if err != nil {
		return nil, errors.Wrapf(err, processName, "BorderSize: %d", bordSize)
	}

	var bdb, t *Bitmap
	if hSize == 1 || vSize == 1 {
		sel := SelCreateBrick(vSize, hSize, vSize/2, hSize/2, SelHit)
		bdb, err = closeBitmap(nil, bsb, sel)
		if err != nil {
			return nil, errors.Wrap(err, processName, "hSize == 1 || vSize == 1")
		}
	} else {
		selH := SelCreateBrick(1, hSize, 0, hSize/2, SelHit)
		t, err := dilate(nil, bsb, selH)
		if err != nil {
			return nil, errors.Wrap(err, processName, "regular - first dilate")
		}

		selV := SelCreateBrick(vSize, 1, vSize/2, 0, SelHit)
		bdb, err = dilate(nil, t, selV)
		if err != nil {
			return nil, errors.Wrap(err, processName, "regular - second dilate")
		}

		if _, err = erode(t, bdb, selH); err != nil {
			return nil, errors.Wrap(err, processName, "regular - first erode")
		}
		if _, err = erode(bdb, t, selV); err != nil {
			return nil, errors.Wrap(err, processName, "regular - second erode")
		}
	}

	if t, err = bdb.RemoveBorder(bordSize); err != nil {
		return nil, errors.Wrap(err, processName, "regular")
	}
	if d == nil {
		return t, nil
	}

	if _, err = copyBitmap(d, t); err != nil {
		return nil, err
	}
	return d, nil
}

func dilate(d *Bitmap, s *Bitmap, sel *Selection) (*Bitmap, error) {
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
			if selData == SelHit {
				if err = d.RasterOperation(j-sel.Cx, i-sel.Cy, s.Width, s.Height, PixSrcOrDst, t, 0, 0); err != nil {
					return nil, err
				}
			}
		}
	}
	return d, nil
}

// hSize is the width of the brick sel
// vSize is the height of the brick sel
func dilateBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	const processName = "dilateBrick"
	if s == nil {
		common.Log.Debug("dilateBrick source not defined")
		return nil, errors.Error(processName, "source bitmap not defined")
	}
	if hSize < 1 || vSize < 1 {
		return nil, errors.Error(processName, "hSzie and vSize are no greater equal to 1")
	}
	if hSize == 1 && vSize == 1 {
		bm, err := copyBitmap(d, s)
		if err != nil {
			return nil, errors.Wrap(err, processName, "hSize == 1 && vSize == 1")
		}
		return bm, nil
	}
	if hSize == 1 || vSize == 1 {
		sel := SelCreateBrick(vSize, hSize, vSize/2, hSize/2, SelHit)
		d, err := dilate(d, s, sel)
		if err != nil {
			return nil, errors.Wrap(err, processName, "hsize == 1 || vSize == 1")
		}
		return d, nil
	}

	selH := SelCreateBrick(1, hSize, 0, hSize/2, SelHit)
	selV := SelCreateBrick(vSize, 1, vSize/2, 0, SelHit)
	bmT, err := dilate(nil, s, selH)
	if err != nil {
		return nil, errors.Wrap(err, processName, "1st dilate")
	}
	d, err = dilate(d, bmT, selV)
	if err != nil {
		return nil, errors.Wrap(err, processName, "2nd dilate")
	}
	return d, nil
}

func erode(d, s *Bitmap, sel *Selection) (*Bitmap, error) {
	const processName = "erode"
	var (
		err error
		t   *Bitmap
	)
	d, err = processMorphArgs1(d, s, sel, &t)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	if err = d.setAll(); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	var selData SelectionValue
	for i := 0; i < sel.Height; i++ {
		for j := 0; j < sel.Width; j++ {
			selData = sel.Data[i][j]
			if selData == SelHit {
				err = rasterOperation(d, sel.Cx-j, sel.Cy-i, s.Width, s.Height, PixSrcAndDst, t, 0, 0)
				if err != nil {
					return nil, errors.Wrap(err, processName, "")
				}
			}
		}
	}

	if MorphBC == SymmetricMorphBC {
		return d, nil
	}
	// If the MorphBoundaryCondition is set to the AsymmetricMorphBC
	// then the erode function should set the pixels surrounding the image to 'OFF'.
	xp, yp, xn, yn := sel.findMaxTranslations()
	if xp > 0 {
		if err = d.RasterOperation(0, 0, xp, s.Height, PixClr, nil, 0, 0); err != nil {
			return nil, errors.Wrap(err, processName, "xp > 0")
		}
	}
	if xn > 0 {
		if err = d.RasterOperation(s.Width-xn, 0, xn, s.Height, PixClr, nil, 0, 0); err != nil {
			return nil, errors.Wrap(err, processName, "xn > 0")
		}
	}
	if yp > 0 {
		if err = d.RasterOperation(0, 0, s.Width, yp, PixClr, nil, 0, 0); err != nil {
			return nil, errors.Wrap(err, processName, "yp > 0")
		}
	}
	if yn > 0 {
		if err = d.RasterOperation(0, s.Height-yn, s.Width, yn, PixClr, nil, 0, 0); err != nil {
			return nil, errors.Wrap(err, processName, "yn > 0")
		}
	}
	return d, nil
}

func erodeBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	const processName = "erodeBrick"
	if s == nil {
		return nil, errors.Error(processName, "source not defined")
	}
	if hSize < 1 || vSize < 1 {
		return nil, errors.Error(processName, "hsize and vsize are not greater than or equal to 1")
	}

	if hSize == 1 && vSize == 1 {
		bm, err := copyBitmap(d, s)
		if err != nil {
			return nil, errors.Wrap(err, processName, "hSize == 1 && vSize == 1")
		}
		return bm, nil
	}
	if hSize == 1 || vSize == 1 {
		sel := SelCreateBrick(vSize, hSize, vSize/2, hSize/2, SelHit)
		d, err := erode(d, s, sel)
		if err != nil {
			return nil, errors.Wrap(err, processName, "hSize == 1 || vSize == 1")
		}
		return d, nil
	}

	selH := SelCreateBrick(1, hSize, 0, hSize/2, SelHit)
	selV := SelCreateBrick(vSize, 1, vSize/2, 0, SelHit)
	bmT, err := erode(nil, s, selH)
	if err != nil {
		return nil, errors.Wrap(err, processName, "1st erode")
	}
	d, err = erode(d, bmT, selV)
	if err != nil {
		return nil, errors.Wrap(err, processName, "2nd erode")
	}
	return d, nil
}

func expandReplicate(s *Bitmap, factor int) (*Bitmap, error) {
	const processName = "expandReplicate"
	if s == nil {
		return nil, errors.Error(processName, "source not defined")
	}
	if factor <= 0 {
		return nil, errors.Error(processName, "invalid factor - <= 0")
	}
	if factor == 1 {
		bm, err := copyBitmap(nil, s)
		if err != nil {
			return nil, errors.Wrap(err, processName, "factor = 1")
		}
		return bm, nil
	}
	bm, err := expandBinaryReplicate(s, factor, factor)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return bm, nil
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

func morphSequence(s *Bitmap, sequence ...MorphProcess) (d *Bitmap, err error) {
	const processName = "morphSequence"
	if s == nil {
		return nil, errors.Error(processName, "morphSequence source bitmap not defined")
	}
	if len(sequence) == 0 {
		return nil, errors.Error(processName, "morphSequence, sequence not defined")
	}

	if err = verifyMorphProcesses(sequence...); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	var w, h, border int
	d = s.Copy()
	for _, process := range sequence {
		switch process.Operation {
		case MopDilation:
			// 'd' character
			w, h = process.getWidthHeight()
			d, err = DilateBrick(nil, d, w, h)
			if err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		case MopErosion:
			// 'e' character
			w, h = process.getWidthHeight()
			d, err = erodeBrick(nil, d, w, h)
			if err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		case MopOpening:
			// 'o' character
			w, h = process.getWidthHeight()
			d, err = openBrick(nil, d, w, h)
			if err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		case MopClosing:
			// 'c' character
			w, h = process.getWidthHeight()
			d, err = closeSafeBrick(nil, d, w, h)
			if err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		case MopRankBinaryReduction:
			// 'r' character
			d, err = reduceRankBinaryCascade(d, process.Arguments...)
			if err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		case MopReplicativeBinaryExpansion:
			// 'x' character
			d, err = expandReplicate(d, process.Arguments[0])
			if err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		case MopAddBorder:
			// 'b' character
			border = process.Arguments[0]
			d, err = d.AddBorder(border, 0)
			if err != nil {
				return nil, errors.Wrap(err, processName, "")
			}
		default:
			return nil, errors.Error(processName, "invalid morphOperation provided to the sequence")
		}
	}

	if border > 0 {
		d, err = d.RemoveBorder(border)
		if err != nil {
			return nil, errors.Wrap(err, processName, "border > 0")
		}
	}
	return d, nil
}

// open is the function based on the leptonica 'pixOpen' - morph.c:422
// this function does a sequence of erode and dilate morph operations.
func open(d, s *Bitmap, sel *Selection) (*Bitmap, error) {
	const processName = "open"
	var err error

	d, err = processMorphArgs2(d, s, sel)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	t, err := erode(nil, s, sel)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	_, err = dilate(d, t, sel)
	if err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return d, nil
}

func openBrick(d, s *Bitmap, hSize, vSize int) (*Bitmap, error) {
	const processName = "openBrick"

	if s == nil {
		return nil, errors.Error(processName, "source bitmap not defined")
	}
	if hSize < 1 && vSize < 1 {
		return nil, errors.Error(processName, "hSize < 1 && vSize < 1")
	}

	if hSize == 1 && vSize == 1 {
		return s.Copy(), nil
	}

	if hSize == 1 || vSize == 1 {
		var err error
		sel := SelCreateBrick(vSize, hSize, vSize/2, hSize/2, SelHit)
		d, err = open(d, s, sel)
		if err != nil {
			return nil, errors.Wrap(err, processName, "hSize == 1 || vSize == 1")
		}
		return d, nil
	}

	selH := SelCreateBrick(1, hSize, 0, hSize/2, SelHit)
	selV := SelCreateBrick(vSize, 1, vSize/2, 0, SelHit)
	t, err := erode(nil, s, selH)
	if err != nil {
		return nil, errors.Wrap(err, processName, "1st erode")
	}
	d, err = erode(d, t, selV)
	if err != nil {
		return nil, errors.Wrap(err, processName, "2nd erode")
	}

	_, err = dilate(t, d, selH)
	if err != nil {
		return nil, errors.Wrap(err, processName, "1st dilate")
	}
	_, err = dilate(d, t, selV)
	if err != nil {
		return nil, errors.Wrap(err, processName, "2nd dilate")
	}
	return d, nil
}

// processMorphArgs1 used for generic erosion, dilation and HMT.
func processMorphArgs1(d *Bitmap, s *Bitmap, sel *Selection, t **Bitmap) (*Bitmap, error) {
	const processName = "processMorphArgs1"
	if s == nil {
		return nil, errors.Error(processName, "MorphArgs1 's' not defined")
	}
	if sel == nil {
		return nil, errors.Error(processName, "MorhpArgs1 'sel' not defined")
	}

	sx, sy := sel.Height, sel.Width
	if sx == 0 || sy == 0 {
		return nil, errors.Error(processName, "selection of size 0")
	}

	if d == nil {
		d = s.createTemplate()
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

func processMorphArgs2(d, s *Bitmap, sel *Selection) (*Bitmap, error) {
	const processName = "processMorphArgs2"
	var sx, sy int
	if s == nil {
		return nil, errors.Error(processName, "source bitmap is nil")
	}
	if sel == nil {
		return nil, errors.Error(processName, "sel not defined")
	}
	sx = sel.Width
	sy = sel.Height
	if sx == 0 || sy == 0 {
		return nil, errors.Error(processName, "sel of size 0")
	}
	if d == nil {
		return s.createTemplate(), nil
	}
	if err := d.resizeImageData(s); err != nil {
		return nil, err
	}
	return d, nil
}

func verifyMorphProcesses(processes ...MorphProcess) (err error) {
	const processName = "verifyMorphProcesses"
	var netRed, border int
	for i, process := range processes {
		if err = process.verify(i, &netRed, &border); err != nil {
			return errors.Wrap(err, processName, "")
		}
	}
	if border != 0 && netRed != 0 {
		return errors.Error(processName, "Morph sequence - border added but net reduction not 0")
	}
	return nil
}
