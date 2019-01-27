package decoder

import (
	"encoding/binary"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/pageinformation"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/pattern-dictionary"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/regions/generic"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/regions/halftone"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/regions/refinement"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/regions/text"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/symbol-dictionary"
	"io"
)

// JBIG2Decoder decodes the stream of
type JBIG2Decoder struct {
	r *reader.Reader

	KnownNumberOfPages       bool
	NumberOfPages            int
	RandomAccessOrganisation bool

	Container *container.Decoder
	Bitmaps   []*bitmap.Bitmap

	GlobalData []byte
}

func New() *JBIG2Decoder {
	return &JBIG2Decoder{
		Container: container.New(),
	}
}

func (j *JBIG2Decoder) Reset() {
	j.KnownNumberOfPages = false
	j.RandomAccessOrganisation = false

	j.NumberOfPages = -1

	j.Container = nil
	j.Bitmaps = nil

}

func (j *JBIG2Decoder) Decode(r io.Reader) ([]byte, error) {
	return nil, nil
}

// DecodeBytes decodes the encoded bytes into the
func (j *JBIG2Decoder) DecodeBytes(encoded []byte) ([]byte, error) {
	// Set Reader
	j.r = reader.New(encoded)

	// CheckHeader
	common.Log.Debug("Checking encoded header")
	validFile, err := j.checkHeader()
	if err != nil {
		common.Log.Debug("Checking header failed: %v", err)
		return nil, err
	}

	if !validFile {
		common.Log.Debug("The encoded stream is not a valid file")
		// This states that this must be some kind of stream from a PDF
		j.KnownNumberOfPages = true
		j.RandomAccessOrganisation = false
		j.NumberOfPages = 1

		if j.GlobalData != nil {
			// Get the segments from the global data at first
			r := reader.New(j.GlobalData)

			if err = j.decodeSegments(r); err != nil {
				return nil, err
			}
		} else {
			_, err := j.r.Seek(-8, io.SeekCurrent)
			if err != nil {
				common.Log.Debug("Seek -8 failed: '%v'", err)
				return nil, err
			}
		}
	} else {
		common.Log.Debug("Encoded bytes is a valid file, setting File Header flags")
		// should be rathr out of the scope of the unidoc
		// this should be the valid stand-alone file
		if err = j.setFileHeaderFlags(j.r); err != nil {
			common.Log.Debug("Setting File Header Flags failed: %v", err)
			return nil, err
		}

		if j.KnownNumberOfPages {
			common.Log.Debug("The Page Number is known.")
			if err = j.numberOfPages(j.r); err != nil {
				common.Log.Debug("Getting Page Number failed: %v", err)
				return nil, err
			}
			common.Log.Debug("The decoded number of pages: %v", j.NumberOfPages)
		} else {
			common.Log.Debug("The page number is unknown.")
		}
	}

	err = j.decodeSegments(j.r)
	if err != nil {
		return nil, err
	}

	return nil, nil

}

func (j *JBIG2Decoder) checkHeader() (bool, error) {
	var (
		controlHeader []byte = []byte{151, 74, 66, 50, 13, 10, 26, 10}
		actualHeader  []byte = make([]byte, 8)
	)

	_, err := j.r.Read(actualHeader)
	if err != nil {
		return false, err
	}

	for i := 0; i < len(actualHeader); i++ {
		if actualHeader[i] != controlHeader[i] {
			return false, nil
		}
	}
	return true, nil

}

func (j *JBIG2Decoder) numberOfPages(r *reader.Reader) error {
	var numberOfPages []byte = make([]byte, 4)
	_, err := r.Read(numberOfPages)
	if err != nil {
		return err
	}

	j.NumberOfPages = int(binary.BigEndian.Uint32(numberOfPages))

	return nil
}

func (j *JBIG2Decoder) decodeSegments(r *reader.Reader) error {
	common.Log.Debug("[decodeSegments] begins")
	defer func() { common.Log.Debug("[decodeSegments] finished") }()

	var (
		finished bool
		err      error
	)

	for !finished {
		h := &header.Header{}

		_, err = h.Decode(r)
		if err != nil {
			return err
		}

		var s segment.Segmenter
		k := kind.SegmentKind(h.SegmentType)

		var inline bool
		common.Log.Debug("Reading Kind: %s", k)
		// Switch over the header type
		switch k {
		case kind.SymbolDictionary:
			s = symboldict.New(j.Container, h)
		case kind.IntermediateTextRegion:
			s = text.New(j.Container, h, false)
		case kind.ImmediateTextRegion:
			inline = true
			s = text.New(j.Container, h, inline)
		case kind.ImmediateLosslessTextRegion:
			inline = true
			s = text.New(j.Container, h, inline)
		case kind.PatternDictionary:
			s = patterndict.New(j.Container, h)
		case kind.IntermediateHalftoneRegion:
			s = halftone.New(j.Container, h, inline)
		case kind.ImmediateHalftoneRegion:
			inline = true
			s = halftone.New(j.Container, h, inline)
		case kind.ImmediateLosslessHalftoneRegion:
			inline = true
			s = halftone.New(j.Container, h, inline)
		case kind.IntermediateGenericRegion:
			s = generic.NewGenericRegionSegment(j.Container, h, inline)
		case kind.ImmediateGenericRegion:
			inline = true
			s = generic.NewGenericRegionSegment(j.Container, h, inline)
		case kind.ImmediateLosslessGenericRegion:
			inline = true
			s = generic.NewGenericRegionSegment(j.Container, h, inline)
		case kind.IntermediateGenericRefinementRegion:
			s = refinement.New(j.Container, h, inline)
		case kind.ImmediateGenericRefinementRegion:
			inline = true
			s = refinement.New(j.Container, h, inline)
		case kind.ImmediateLosslessGenericRefinementRegion:
			inline = true
			s = refinement.New(j.Container, h, inline)
		case kind.PageInformation:
			s = pageinformation.New(j.Container, h)
		case kind.EndOfPage:
			continue
		case kind.EndOfStrip:

		case kind.EndOfFile:
			finished = true
		case kind.Profiles:
			common.Log.Debug("Unimplemented")
		case kind.Tables:
			common.Log.Debug("Unimplemented")
		case kind.Extension:

		default:
			common.Log.Error("Unknown Segment type in JBIG2 stream")
		}

		if !j.RandomAccessOrganisation {
			if err = s.Decode(r); err != nil {
				return err
			}

			bmg, ok := s.(bitmap.BitmapGetter)
			if !inline && ok {
				j.Bitmaps = append(j.Bitmaps, bmg.GetBitmap())
			}
		}

		j.Container.Segments = append(j.Container.Segments)
	}

	if j.RandomAccessOrganisation {
		for _, s := range j.Container.Segments {
			if err = s.Decode(r); err != nil {
				return err
			}

			bmg, ok := s.(bitmap.BitmapGetter)
			if ok {
				bm := bmg.GetBitmap()
				if bm != nil {
					j.Bitmaps = append(j.Bitmaps, bm)
				}
			}
		}
	}

	return nil
}

func (j *JBIG2Decoder) setFileHeaderFlags(r *reader.Reader) error {
	headerFlags, err := r.ReadByte()
	if err != nil {
		return err
	}

	if (headerFlags & 0xfc) != 0 {
		common.Log.Warning("Warning, reserved bits (2-7) of file header flags are not zero: '%b'", headerFlags)
	}

	// Read Random Access Oragnisation flag
	j.RandomAccessOrganisation = ((headerFlags & 1) == 0)

	// Read Known Number of Pages flag
	j.KnownNumberOfPages = ((headerFlags & 2) == 0)

	return nil
}
