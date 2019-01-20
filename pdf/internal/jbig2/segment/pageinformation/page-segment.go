package pageinformation

import (
	"encoding/binary"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/model"
)

var (
	log = common.Log
)

type PageInformationSegment struct {
	*model.Segment

	PageBMHeight, PageBMWidth int
	XResolution, YResolution  int

	PageInfoFlags *Flags
	pageStripping int

	PageBitmap *bitmap.Bitmap
}

// Read reads the segment from the input reader
func (p *PageInformationSegment) Decode(r *reader.Reader) error {
	log.Debug("[READ] Page Information Segment")

	var buf []byte = make([]byte, 4)

	_, err := r.Read(buf)
	if err != nil {
		log.Debug("Read Width block failed. %v", err)
		return err
	}

	p.PageBMWidth = int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, 4)

	_, err = r.Read(buf)
	if err != nil {
		log.Debug("Read Height block failed. %v", err)
		return err
	}

	p.PageBMHeight = int(binary.BigEndian.Uint32(buf))

	log.Debug("Bitmap size. Height: %v, Width: %v", p.PageBMHeight, p.PageBMWidth)

	flags, err := r.ReadByte()
	if err != nil {
		log.Debug("Read Flags block failed. %v", err)
		return err
	}

	p.PageInfoFlags.SetValue(int(flags))
	log.Debug("Flags: %d", flags)

	buf = make([]byte, 2)

	_, err = r.Read(buf)
	if err != nil {
		log.Debug("Read Page Stripping block. %v", err)
		return err
	}

	p.pageStripping = int(binary.BigEndian.Uint16(buf))
	log.Debug("Page Stripping: %d", p.pageStripping)

	defPix := p.PageInfoFlags.GetValue(DefaultPixelValue)

	var height int

	if p.PageBMHeight == -1 {
		height = p.pageStripping & 0x7fff
	} else {
		height = p.PageBMHeight
	}

	p.PageBitmap = bitmap.New(p.PageBMWidth, height, p.Decoders)
	p.PageBitmap.Clear(defPix == 1)
	return nil
}
