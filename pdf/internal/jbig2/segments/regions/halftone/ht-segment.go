package halftone

import (
	"encoding/binary"
	"errors"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/pageinformation"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/pattern-dictionary"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/regions"
)

// HalftoneSegment is the struct model for the JBIG2 Haltone Region Segment
type HalftoneSegment struct {
	*regions.RegionSegment
	HFlags *HalftoneSegmentFlags

	// Grid positions and size
	GridWidth  int
	GridHeight int
	GridX      int
	GridY      int

	bm *bitmap.Bitmap

	inlineImage bool
}

// New creates new HalftoneSegment
func New(d *container.Decoder, h *header.Header, inline bool) *HalftoneSegment {
	ht := &HalftoneSegment{
		RegionSegment: regions.NewRegionSegment(d, h),
		HFlags:        newFlags(),
		inlineImage:   inline,
	}

	return ht
}

// Decode decodes the Halftone from the provided reader
func (h *HalftoneSegment) Decode(r *reader.Reader) error {
	common.Log.Debug("[HalftonSegment][DECODE] begins.")
	defer func() { common.Log.Debug("[HalftonSegment][DECODE] finished.") }()

	// Decode the Region segment at first
	if err := h.RegionSegment.Decode(r); err != nil {
		common.Log.Debug("RegionSegment Decode failed: %v", err)
		return err
	}

	// Read Halftone Flags
	if err := h.readHalftoneFlags(r); err != nil {
		return err
	}

	// Read Grid Width
	var buf []byte = make([]byte, 4)
	_, err := r.Read(buf)
	if err != nil {
		common.Log.Debug("Reading Halftone Grid Width failed. %v", err)
		return err
	}
	h.GridWidth = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)
	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Reading Halftone Grid Height failed. %v", err)
		return err
	}
	h.GridHeight = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)
	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Reading Halftone Grid X failed. %v", err)
		return err
	}
	h.GridX = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)
	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Reading Halftone Grid Y failed. %v", err)
		return err
	}
	h.GridY = int(binary.BigEndian.Uint32(buf))

	common.Log.Debug("Halftone Grids: %+v", h)

	var (
		stepX, stepY uint16
	)

	buf = make([]byte, 2)
	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Halftone Reading StepX failed. %v", err)
		return err
	}

	stepX = binary.BigEndian.Uint16(buf)

	buf = make([]byte, 2)
	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Halftone Reading StepY failed. %v", err)
		return err
	}

	stepY = binary.BigEndian.Uint16(buf)

	common.Log.Debug("Step Size - X: %v, Y: %v", stepX, stepY)

	if len(h.Header.ReferredToSegments) != 1 {
		err = errors.New("Decoding Halftone Segment failed. This segment should refer only to single segment.")
		common.Log.Debug("Halfton.Header.ReferredToSegments has invalid value. %v", err)
		return err
	}

	// Find the pattern dictionary refered segment
	seg := h.Decoders.FindSegment(h.Header.ReferredToSegments[0])
	if seg == nil {
		err = errors.New("Cannot find referred segment for the Halftone Segment")
		common.Log.Debug("Decoders.FindSegment failed: %v", err)
		return err
	}

	// The kind must be a PatternDictionary
	if seg.Kind() != kind.PatternDictionary {
		err = errors.New("The Halftone Segment refers to a bad symbol ditcionary reference")
		common.Log.Debug("Invalid refered segment kind. %v", err)
		return err
	}

	// cast a pattern dictionary segment
	pd, ok := seg.(*patterndict.PatternDictionarySegment)
	if !ok {
		err = errors.New("Internal Error. Invalid Refered to Segment found.")
		common.Log.Error("%s", err)
		common.Log.Debug("Refered Segment doesn't cast to PatternDictionarySegment.")
		return err
	}

	var (
		bitsPerValue int
		i            int = 1
	)

	for i < pd.Size {
		bitsPerValue++
		i <<= 1
	}

	// PatternDictionary should have a bitmap
	if len(pd.Bitmaps) < 1 {
		err = errors.New("Internal Error. Refered PatternDictionary has no bitmaps")
		return err
	}

	var patternBM *bitmap.Bitmap = pd.Bitmaps[0]

	common.Log.Debug("Pattern Size: %v, %v", patternBM.Width, patternBM.Height)

	var (
		useMMR   bool = h.HFlags.GetValue(H_MMR) != 0
		template int  = h.HFlags.GetValue(H_TEMPLATE)
	)

	if !useMMR {
		common.Log.Debug("Decoding Halftone Segment Without MMR")
		h.Decoders.Arithmetic.ResetGenericStats(template, nil)
		err = h.Decoders.Arithmetic.Start(r)
		if err != nil {
			return err
		}
	}

	var halfoneDefaultPixel int = h.HFlags.GetValue(H_DEF_PIXEL)

	bm := bitmap.New(h.BMWidth, h.BMHeight, h.Decoders)
	bm.Clear(halfoneDefaultPixel != 0)

	var enableSkip bool = h.HFlags.GetValue(H_ENABLE_SKIP) != 0

	var skipBitmap *bitmap.Bitmap
	if enableSkip {
		skipBitmap = bitmap.New(h.GridWidth, h.GridHeight, h.Decoders)
		skipBitmap.Clear(false)

		for y := 0; y < h.GridHeight; y++ {
			for x := 0; x < h.GridWidth; x++ {
				xx := h.GridX + y*int(stepY) + x*int(stepX)
				yy := h.GridY + y*int(stepX) - x*int(stepY)

				if (xx+patternBM.Width)>>8 <= 0 || (xx>>8) >= h.BMWidth ||
					(yy+patternBM.Height>>8) <= 0 || (yy>>8) >= h.BMHeight {
					skipBitmap.SetPixel(y, x, 1)
				}
			}
		}
	}

	var (
		grayScaleImage []int  = make([]int, h.GridWidth*h.GridHeight)
		genBATX        []int8 = make([]int8, 4)
		genBATY        []int8 = make([]int8, 4)
	)

	if template <= 1 {
		genBATX[0] = 3
	} else {
		genBATX[0] = 2
	}

	genBATY[0] = -1

	genBATX[1] = -3
	genBATY[1] = -1

	genBATX[2] = 2
	genBATY[2] = -2

	genBATX[3] = -2
	genBATY[3] = -2

	var grayBM *bitmap.Bitmap

	for j := bitsPerValue - 1; j >= 0; {
		j -= 1
		grayBM = bitmap.New(h.GridWidth, h.GridHeight, h.Decoders)

		err = grayBM.Read(r, useMMR, template, false, enableSkip, skipBitmap, genBATX, genBATY, -1)
		if err != nil {
			common.Log.Debug("GrayBM at index: '%d' have failed: %v", err)
			return err
		}

		i := 0
		for row := 0; row < h.GridHeight; row++ {
			for col := 0; col < h.GridWidth; col++ {
				var bit int
				if grayBM.GetPixel(col, row) {
					bit = 1
				}

				bit = bit ^ grayScaleImage[i]&1

				grayScaleImage[i] = (grayScaleImage[i] << 1) | bit
				i++
			}
		}
	}

	var combinationOperator int = h.HFlags.GetValue(H_COMB_OP)

	i = 0
	for col := 0; col < h.GridHeight; col++ {
		xx := h.GridX + col*int(stepY)
		yy := h.GridY + col*int(stepX)

		for row := 0; row < h.GridWidth; row++ {
			if !enableSkip && skipBitmap.GetPixel(col, row) {
				pattern := pd.Bitmaps[grayScaleImage[i]]
				bm.Combine(pattern, xx>>8, yy>>8, int64(combinationOperator))
			}

			xx += int(stepX)
			yy -= int(stepY)

			i += 1
		}
	}

	// Page 217
	if h.inlineImage {
		ps := h.Decoders.FindPageSegment(h.Header.PageAssociation)
		if ps == nil {
			common.Log.Debug("Can't find association page segment.")
			return errors.New("Association page segment not found.")
		}

		p := ps.(*pageinformation.PageInformationSegment)

		externalCombinationOperator := int64(h.RegionFlags.GetValue(regions.ExternalCombinationOperator))

		common.Log.Debug("External Combination operator: %v", externalCombinationOperator)

		err := p.PageBitmap.Combine(bm, h.BMXLocation, h.BMYLocation, externalCombinationOperator)
		if err != nil {
			common.Log.Debug("PageBitmap Combine failed: %v", err)
			return err
		}
	} else {
		bm.BitmapNumber = h.Header.SegmentNumber
		h.bm = bm
	}

	return nil
}

func (h *HalftoneSegment) GetBitmap() *bitmap.Bitmap {
	return h.bm
}

func (h *HalftoneSegment) readHalftoneFlags(r *reader.Reader) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}

	h.HFlags.SetValue(int(b))

	return nil
}
