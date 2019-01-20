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

	GenericRegionFlags *GenericRegionFlags

	inlineImage, unknownLength bool
	UseMMR                     bool

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
		GenericRegionFlags: newFlags(),
	}

	return g
}

// Decode reads the segment for the provided reader.
// Implements Segmenter interface
func (g *GenericRegionSegment) Decode(r *reader.Reader) error {
	common.Log.Debug("[READ] Generic Region Segment started.")
	defer common.Log.Debug("[READ] Generic Region Segment finished.")

	common.Log.Debug("Decode RegionSegment basics")
	// Read the segment basics
	if err := g.RegionSegment.Decode(r); err != nil {
		common.Log.Debug("RegionSegment.Decode failed.")
		return err
	}

	common.Log.Debug("ReadGenericSegmentFlags")
	// Read generic region segment flags
	if err := g.readGenericRegionFlags(r); err != nil {
		common.Log.Debug("readGenericRegionFlags failed. %v", err)
		return err
	}

	g.UseMMR = g.GenericRegionFlags.GetValue(MMR) != 0
	var (
		template = g.GenericRegionFlags.GetValue(GBTemplate)

		AdaptiveTemplateX, AdaptiveTemplateY []int8 = make([]int8, 4), make([]int8, 4)
	)

	common.Log.Debug("UseMMR: %v ", g.UseMMR)
	if !g.UseMMR {

		var err error
		var buf []byte = make([]byte, 2)
		if template == 0 {

			for i := 0; i < 4; i++ {
				if _, err := r.Read(buf); err != nil {
					common.Log.Debug("Reading AdaptiveTemplate at %d failed", i)
					return err
				}

				AdaptiveTemplateX[i] = int8(buf[0])
				AdaptiveTemplateY[i] = int8(buf[1])

			}

		} else {
			if _, err := r.Read(buf); err != nil {
				common.Log.Debug("Reading first AdaptiveTemplate failed")
				return err
			}
			AdaptiveTemplateX[0] = int8(buf[0])
			AdaptiveTemplateY[0] = int8(buf[1])
		}
		common.Log.Debug("AdaptiveTemplateX %v", AdaptiveTemplateX)
		common.Log.Debug("AdaptiveTemplateY %v", AdaptiveTemplateY)

		g.Decoders.Arithmetic.ResetGenericStats(template, nil)

		if err = g.Decoders.Arithmetic.Start(r); err != nil {
			common.Log.Debug("ArithmeticDecoder Start failed. %v", err)
			return err
		}
	}

	var (
		typicalPrediction bool = g.GenericRegionFlags.GetValue(TPGDOn) != 0
		length            int  = g.Segment.Header.DataLength
	)

	if length != -1 {
		// If lenght of data is unknown it needs to be determined through examination of the data
		// JBIG2 7.2.7

		common.Log.Debug("Length set to: %v", length)

		g.unknownLength = true

		var match1, match2 byte

		if g.UseMMR {
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

	common.Log.Debug("Create new bitmap - w: %d, h: %d", g.BMHeight, g.BMWidth)
	bm := bitmap.New(g.BMWidth, g.BMHeight, g.Decoders)
	bm.Clear(false)

	var mmrDataLength int
	if !g.UseMMR {
		mmrDataLength = length - 18
	}

	err := bm.Read(
		r,
		g.UseMMR,
		template,
		typicalPrediction,
		false,
		nil,
		AdaptiveTemplateX,
		AdaptiveTemplateY,
		mmrDataLength,
	)
	if err != nil {
		return err
	}

	if g.inlineImage {
		// If the segment is an inline image get the page info segment
		// and combine it's bitmap with this 'bitmap'
		pageSegmenter := g.Decoders.FindPageSegment(g.PageAssociation())
		if pageSegmenter != nil {
			// the page segmenter must be pageinformation.PISegment

			pageSegment := pageSegmenter.(*pageinformation.PageInformationSegment)

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
		if _, err := r.Seek(4, io.SeekCurrent); err != nil {
			common.Log.Debug("Seek 4 from current failed: %v", err)
			return err
		}

	}

	return nil
}

func (g *GenericRegionSegment) readGenericRegionFlags(r io.ByteReader) error {
	GenericRegionFlags, err := r.ReadByte()
	if err != nil {
		return err
	}

	g.GenericRegionFlags.SetValue(int(GenericRegionFlags))

	common.Log.Debug("Generic Region Segment Flags: %d", GenericRegionFlags)
	return nil
}
