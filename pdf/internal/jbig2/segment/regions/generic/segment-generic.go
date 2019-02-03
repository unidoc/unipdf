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

// GenericRegionSegment is the model that represents JBIG2 Generic Region Segment
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

	common.Log.Debug("Decode RegionSegment")
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
	common.Log.Debug("UseMMR: %v, Template: %v", g.UseMMR, template)

	if !g.UseMMR {

		// set Adaptive Templates
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

		common.Log.Debug("Read Adaptive Template X, and Adaptive Template Y finished.")
		common.Log.Debug("AdaptiveTemplateX: %v", AdaptiveTemplateX)
		common.Log.Debug("AdaptiveTemplateY: %v", AdaptiveTemplateY)
	}

	var (
		typicalPrediction bool = g.GenericRegionFlags.GetValue(TPGDOn) != 0
		length            int  = g.Header.DataLength
	)

	common.Log.Debug("Typical Prediction: %v", typicalPrediction)

	if length == -1 {
		// If lenght of data is unknown it needs to be determined through examination of the data
		// JBIG2 7.2.7

		common.Log.Debug("Length set to: %v", length)

		g.unknownLength = true

		var match1, match2 byte

		// Match1 And Match2 are the bytes to look for
		if g.UseMMR {
			match1 = 0x00
			match2 = 0x00
		} else {
			match1 = 0xFF
			match2 = 0xAC
		}

		var bytesRead int64

		for {
			//
			b1, err := r.ReadByte()
			if err != nil {
				common.Log.Debug("GenericSegment->Decode->ReadByte failed. %v", err)
				return err
			}

			bytesRead += 1

			// check if b1 matches match1
			// continue to read the
			if b1 == match1 {
				b2, err := r.ReadByte()
				if err != nil {
					common.Log.Debug("GenericSegment->Decode->ReadByte Byte2 failed. %v", err)
					return err
				}
				bytesRead += 1

				// if b2 matches match2
				if b2 == match2 {
					length = int(bytesRead) - 2
					break
				}
			}
		}

		common.Log.Debug("Bytes read: %d. Rewind", bytesRead)
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

	common.Log.Debug("Reading Generic Bitmap with: MMR-%v,template-%v,typicalPredition:%v,ATX:%v, ATY:%v,mmrDL:%v", g.UseMMR, template, typicalPrediction, AdaptiveTemplateX, AdaptiveTemplateY, mmrDataLength)
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

	common.Log.Debug("Generic Bitmap: \n%s", bm)

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

			common.Log.Debug("Combine Operation: %d", extComboOp)
			err := pageBM.Combine(bm, g.BMXLocation, g.BMYLocation, int64(extComboOp))
			if err != nil {
				common.Log.Debug("PageBitmap Combine failed: %v", err)
				return err
			}

		} else {
			common.Log.Debug("Page segment is nil.")
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

// GetBitmap implements bitmap.BitmapGetter interface
func (g *GenericRegionSegment) GetBitmap() *bitmap.Bitmap {
	return g.Bitmap
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
