package refinement

import (
	"errors"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/pageinformation"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/regions"
)

// RefinementRegionSegment is the model for the JBIG2 Refinement Region Segment
type RefinementRegionSegment struct {
	*regions.RegionSegment

	// RefinementFlags defines the refinement region flags
	RefinementFlags *RefinementRegionFlags

	inlineImage bool
	bm          *bitmap.Bitmap
}

// New creates new Refinement Region Segment
func New(
	d *container.Decoder,
	h *header.Header,
	inline bool,
) *RefinementRegionSegment {
	r := &RefinementRegionSegment{
		RegionSegment:   regions.NewRegionSegment(d, h),
		RefinementFlags: newFlags(),
		inlineImage:     inline,
	}
	return r
}

func (s *RefinementRegionSegment) Decode(r *reader.Reader) (err error) {
	common.Log.Debug("[REFINEMENT-REGION-SEGMENT][DECODE] begins")
	defer func() { common.Log.Debug("[REFINEMENT-REGION-SEGMENT][DECODE] finished") }()

	// Decode the RegionSegment
	if err = s.RegionSegment.Decode(r); err != nil {
		common.Log.Debug("Decode RegionSegment failed: %v", err)
		return err
	}

	common.Log.Debug("RegionSegment Decoded")

	// Read the Flags
	if err = s.readFlags(r); err != nil {
		common.Log.Debug("readFlags failed. %v", err)
		return err
	}

	common.Log.Debug("RefinementRegionFlags decoded")

	// Define the GenericAdaptiveTemplates
	var (
		genericRegionATX []int8 = make([]int8, 2)
		genericRegionATY []int8 = make([]int8, 2)
		b                byte
	)

	template := s.RefinementFlags.GetValue(GR_TEMPLATE)
	if template == 0 {
		for i := 0; i < 2; i++ {
			b, err = r.ReadByte()
			if err != nil {
				common.Log.Debug("Reading byte for %d GenericRegionATX failed. %v", i, err)
				return err
			}
			genericRegionATX[i] = int8(b)

			b, err = r.ReadByte()
			if err != nil {
				common.Log.Debug("Reading byte for %d GenericRegionATY failed. %v", i, err)
				return err
			}
			genericRegionATY[i] = int8(b)
		}
	}

	if s.Header.ReferredToSegmentCount == 0 || s.inlineImage {
		ps := s.Decoders.FindPageSegment(s.Header.PageAssociation)
		if ps == nil {
			err = errors.New("Decode RefinementRegionSegment failed. Page Segment not found")
			common.Log.Debug("%s", err)
			return
		}

		p := ps.(*pageinformation.PageInformationSegment)

		if p.PageBMHeight == -1 && s.BMYLocation+s.BMHeight > p.PageBitmap.Height {
			p.PageBitmap.Expand(s.BMYLocation+s.BMHeight, p.PageInfoFlags.GetValue(pageinformation.DefaultPixelValue))
		}
	}

	if s.Header.ReferredToSegmentCount > 1 {
		err = errors.New("Bad reference in JBIG2 geneirc refinement Segment")
		common.Log.Debug("Err: %v. ReferredToSegmentCount > 1.", err)
		return err
	}

	var referedBitmap *bitmap.Bitmap
	if s.Header.ReferredToSegmentCount == 1 {
		referedSeg := s.Decoders.FindSegment(s.Header.ReferredToSegments[0])
		referedBitmap = referedSeg.(bitmap.BitmapGetter).GetBitmap()
	} else {
		ps := s.Decoders.FindPageSegment(s.Header.PageAssociation)
		if ps == nil {
			err = errors.New("Decode RefinementRegionSegment failed. Page Segment not found")
			common.Log.Debug("%s", err)
			return
		}

		p := ps.(*pageinformation.PageInformationSegment)
		referedBitmap, err = p.PageBitmap.GetSlice(s.BMXLocation, s.BMYLocation, s.BMWidth, s.
			BMHeight)
		if err != nil {
			common.Log.Debug("PageInformationSegment.Bitmap - GetSlice failed. %v", err)
			return
		}
	}

	s.Decoders.Arithmetic.ResetRefinementStats(template, nil)
	if err = s.Decoders.Arithmetic.Start(r); err != nil {
		common.Log.Debug("Arithmetic Decoder Start failed: %v", err)
		return
	}

	var typicalPrediction bool = s.RefinementFlags.GetValue(TPGDON) != 0

	bm := bitmap.New(s.BMWidth, s.BMHeight, s.Decoders)

	err = bm.ReadGenericRefinementRegion(r, template, typicalPrediction, referedBitmap, 0, 0, genericRegionATX, genericRegionATY)
	if err != nil {
		common.Log.Debug("ReadGeneric Refinement Region failed. %v", err)
		return
	}

	if s.inlineImage {
		common.Log.Debug("Finding Page Segment: %v", err)
		ps := s.Decoders.FindPageSegment(s.Header.PageAssociation)
		if ps == nil {
			err = errors.New("Decode RefinementRegionSegment failed. Page Segment not found")
			common.Log.Debug("%s", err)
			return
		}

		p := ps.(*pageinformation.PageInformationSegment)

		extComboOp := s.RegionFlags.GetValue(regions.ExternalCombinationOperator)

		common.Log.Debug("Combine PageBitmap with RefinementRegionBitmap")
		err = p.PageBitmap.Combine(bm, s.BMXLocation, s.BMYLocation, int64(extComboOp))
		if err != nil {
			common.Log.Debug("Combine page bitmap with this.bm failed: %v", err)
			return
		}
	} else {
		bm.BitmapNumber = s.Header.SegmentNumber
	}

	return nil
}

func (s *RefinementRegionSegment) GetBitmap() *bitmap.Bitmap {
	return s.bm
}

func (s *RefinementRegionSegment) readFlags(r *reader.Reader) error {
	// Read the flags in the first byte
	b, err := r.ReadByte()
	if err != nil {
		return err
	}

	s.RefinementFlags.SetValue(int(b))

	return nil

}
