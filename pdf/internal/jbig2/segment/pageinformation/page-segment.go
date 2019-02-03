package pageinformation

import (
	"encoding/binary"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/model"
)

type PageInformationSegment struct {
	*model.Segment

	PageBMHeight, PageBMWidth int
	XResolution, YResolution  int

	PageInfoFlags *Flags
	pageStripping int

	PageBitmap *bitmap.Bitmap
}

func New(d *container.Decoder, h *header.Header) *PageInformationSegment {
	p := &PageInformationSegment{
		Segment:       model.New(d, h),
		PageInfoFlags: newFlags(),
	}
	return p
}

// Read reads the segment from the input reader
func (p *PageInformationSegment) Decode(r *reader.Reader) error {
	common.Log.Debug("[PAGE-SEGMENT][DECODE] Begins ")
	defer func() { common.Log.Debug("[PAGE-SEGMENT][DECODE] Finished") }()

	var buf []byte = make([]byte, 4)

	_, err := r.Read(buf)
	if err != nil {
		common.Log.Debug("Read Width block failed. %v", err)
		return err
	}

	p.PageBMWidth = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)

	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Read Height block failed. %v", err)
		return err
	}

	p.PageBMHeight = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)

	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Read Height block failed. %v", err)
		return err
	}

	p.XResolution = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)

	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Read Height block failed. %v", err)
		return err
	}

	p.YResolution = int(binary.BigEndian.Uint32(buf))

	common.Log.Debug("Page Bitmap size: Height: %v, Width: %v", p.PageBMHeight, p.PageBMWidth)

	flags, err := r.ReadByte()
	if err != nil {
		common.Log.Debug("Read Flags block failed. %v", err)
		return err
	}

	p.PageInfoFlags.SetValue(int(flags))
	common.Log.Debug("Flags: %d", flags)

	buf = make([]byte, 2)

	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Read Page Stripping block. %v", err)
		return err
	}

	p.pageStripping = int(binary.BigEndian.Uint16(buf))
	common.Log.Debug("Page Stripping: %d", p.pageStripping)

	defPix := p.PageInfoFlags.GetValue(DefaultPixelValue)

	var height int

	if p.PageBMHeight == -1 {
		height = p.pageStripping & 0x7fff
	} else {
		height = p.PageBMHeight
	}

	p.PageBitmap = bitmap.New(p.PageBMWidth, height, p.Decoders)
	p.PageBitmap.Clear(defPix != 0)

	return nil
}
