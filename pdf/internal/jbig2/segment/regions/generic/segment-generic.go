package generic

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/pageinformation"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/regions"
	"io"
)

var log = common.Log

type GenericRegionSegment struct {
	*regions.RegionSegment

	genericRegionFlags *GenericRegionFlags

	inlineImage, unknownLength bool

	Bitmap *bitmap.Bitmap
}

// NewGenericRegionSegment creates new GenericRegionSegment
func NewGenericRegionSegment(
	decoders *container.Decoder,
	h *header.Header,
	inlineImage bool,
) *GenericRegionSegment {

	g := &GenericRegionSegment{
		RegionSegment:      regions.NewRegionSegment(decoders, h),
		inlineImage:        inlineImage,
		genericRegionFlags: newFlags(),
	}

	return g
}

// Decode reads the segment for the provided reader.
// Implements Segmenter interface
func (g *GenericRegionSegment) Decode(r *reader.Reader) error {
	common.Log.Debug("[READ] Generic Region Segment started.")
	defer common.Log.Debug("[READ] Generic Region Segment finished.")

	// Read the segment basics
	if err := g.RegionSegment.Decode(r); err != nil {
		common.Log.Debug("RegionSegment.Decode failed.")
		return err
	}

	// Read generic region segment flags
	if err := g.readGenericRegionFlags(r); err != nil {
		common.Log.Debug("readGenericRegionFlags failed. %v", err)
		return err
	}

	var (
		useMMR   bool = g.genericRegionFlags.GetValue(mmr) != 0
		template      = g.genericRegionFlags.GetValue(gbTemplate)

		genericBAdaptiveTemplateX []byte = make([]byte, 4)
		genericBAdaptiveTemplateY []byte = make([]byte, 4)
	)

	if !useMMR {
		var err error
		if template == 0 {
			for i := 0; i < 4; i++ {
				genericBAdaptiveTemplateX[i], err = r.ReadByte()
				if err != nil {
					common.Log.Debug("GenericBAdaptiveTemplate X at %d failed", i)
					return err
				}
				genericBAdaptiveTemplateY[i], err = r.ReadByte()
				if err != nil {
					common.Log.Debug("GenericBAdaptiveTemplate Y at %d failed", i)
					return err
				}
			}

		} else {
			genericBAdaptiveTemplateX[0], err = r.ReadByte()
			if err != nil {
				common.Log.Debug("GenericBAdaptiveTemplate X at %d failed")
				return err
			}
			genericBAdaptiveTemplateY[0], err = r.ReadByte()
			if err != nil {
				common.Log.Debug("GenericBAdaptiveTemplate Y at %d failed")
				return err
			}
		}

		g.Decoders.Arithmetic.ResetGenericStats(template, nil)

		if err = g.Decoders.Arithmetic.Start(r); err != nil {
			common.Log.Debug("ArithmeticDecoder Start failed. %v", err)
			return err
		}
	}

	var (
		typicalPrediction bool = g.genericRegionFlags.GetValue(tpgdon) != 0
		length            int  = g.Segment.Header.DataLength
	)

	if length != -1 {
		// If lenght of data is unknown it needs to be determined through examination of the data
		// JBIG2 7.2.7

		g.unknownLength = true

		var match1, match2 byte

		if useMMR {
			match1 = 0
			match2 = 0
		} else {
			match1 = 255
			match2 = 172
		}

		var bytesRead int64
		for {
			b1, err := r.ReadByte()
			if err != nil {
				common.Log.Debug("GenericSegment->Decode->ReadByte failed. %v", err)
				return err
			}

			bytesRead += 1

			if b1 == match1 {
				b2, err := r.ReadByte()
				if err != nil {
					common.Log.Debug("GenericSegment->Decode->ReadByte Byte2 failed. %v", err)
					return err
				}

				if b2 == match2 {
					length = int(bytesRead) - 2
					break
				}
			}
		}

		_, err := r.Seek(-bytesRead, io.SeekCurrent)
		if err != nil {
			log.Debug("GenericSegment->Decode->Seek() failed: %v", err)
			return err
		}
	}

	bm := bitmap.New(g.BMWidth, g.BMHeight, g.Decoders)

	bm.Clear(false)

	var mmrDataLength int
	if !useMMR {
		mmrDataLength = length - 18
	}

	bm.Read(template, useMMR, typicalPrediction, false, nil, genericBAdaptiveTemplateX, genericBAdaptiveTemplateY, mmrDataLength)

	if g.inlineImage {
		// If the segment is an inline image get the page info segment
		// and combine it's bitmap with this 'bitmap'
		pageSegmenter := g.Decoders.FindPageSegment(g.PageAssociation())
		if pageSegmenter != nil {
			// the page segmenter must be pageinformation.PISegment
			pageSegment := pageSegmenter.(*pageinformation.PISegment)

			// get page bitmap
			pageBM := pageSegment.PageBitmap

			extComboOp := g.RegionFlags.GetValue(regions.ExternalCombinationOperator)

			if pageSegment.PageBMHeight == -1 && g.BMYLocation+g.BMHeight > pageBM.Height {
				pageBM.Expand(g.BMYLocation+g.BMHeight, pageSegment.PageInfoFlags.GetValue(pageinformation.DefaultPixelValue))

			}

			pageBM.Combine(bm, g.BMXLocation, g.BMYLocation, int64(extComboOp))

		} else {
			log.Debug("Page segment is nil.")
		}
	} else {
		// if not an inline image set the bitmap to the segment variable
		bm.BitmapNumber = g.Header.SegmentNumber
		g.Bitmap = bm
	}

	if g.unknownLength {
		r.Seek(4, io.SeekCurrent)
	}

	return nil
}

func (g *GenericRegionSegment) readGenericRegionFlags(r io.ByteReader) error {
	genericRegionFlags, err := r.ReadByte()
	if err != nil {
		return err
	}

	g.genericRegionFlags.SetValue(int(genericRegionFlags))

	common.Log.Debug("Generic Region Segment Flags: %d", genericRegionFlags)
	return nil
}
