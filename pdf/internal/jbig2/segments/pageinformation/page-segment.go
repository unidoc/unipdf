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

// PageInformation represents the segment type Page Information 7.4.8
type PageInformation struct {
	r *reader.Reader

	// Page bitmap height, four byte, 7.4.8.1
	PageBMHeight int

	// Page bitmap width, four byte, 7.4.8.1
	PageBMWidth int

	// Page X resolution, four byte 7.4.8.3
	ResolutionX int

	// Page Y resolution, four byte 7.4.8.4
	ResolutionY int

	// Page segment flags, one byte 7.4.8.5
	combinaitonOperatorOverrideAllowed bool
	combinationOperator                bitmap.CombinationOperator
	requiresAuxiliaryBuffer            bool
	defaultPixelValue                  uint8
	mightContainRefinements            bool
	isLossless                         bool

	// Page striping information, two byte 7.4.8.6
}

func New(d *container.Decoder, h *header.Header) *PageInformationSegment {
	p := &PageInformationSegment{
		Segment:       model.New(d, h),
		PageInfoFlags: newFlags(),
	}
	return p
}

func (p *PageInformation) Init() error {
	return nil
}

func (p *PageInformation) parseHeader() error {

}

func (p *PageInformation) readResoultion() error {
	return nil
}

func (p *PageInformation) checkInput() error {
	return nil
}

func (p *PageInformation) readCombinationOperatorOverrideAllowed() error {
	return nil
}

func (p *PageInformation) readRequiresAuxiliaryBuffer() error {
	return nil
}

func (p *PageInformation) readCombinationOperator() error {
	return nil
}

func (p *PageInformation) readDefaultPixelValue() error {
	return nil
}

func (p *PageInformation) readContainsRefinement() error {
	return nil
}

func (p *PageInformation) readIsLossless() error {
	return nil
}

func (p *PageInformation) readIsStriped() error {
	return nil
}

func (p *PageInformation) readMaxStripeSize() error {
	return nil
}

func (p *PageInformation) readWidthAndHeight() error {
	return nil
}

// // Read reads the segment from the input reader
// func (p *PageInformationSegment) Decode(r *reader.Reader) error {
// 	common.Log.Debug("[PAGE-SEGMENT][DECODE] Begins ")
// 	defer func() { common.Log.Debug("[PAGE-SEGMENT][DECODE] Finished") }()

// 	var buf []byte = make([]byte, 4)

// 	_, err := r.Read(buf)
// 	if err != nil {
// 		common.Log.Debug("Read Width block failed. %v", err)
// 		return err
// 	}

// 	p.PageBMWidth = int(binary.BigEndian.Uint32(buf))

// 	buf = make([]byte, 4)

// 	_, err = r.Read(buf)
// 	if err != nil {
// 		common.Log.Debug("Read Height block failed. %v", err)
// 		return err
// 	}

// 	p.PageBMHeight = int(binary.BigEndian.Uint32(buf))

// 	buf = make([]byte, 4)

// 	_, err = r.Read(buf)
// 	if err != nil {
// 		common.Log.Debug("Read Height block failed. %v", err)
// 		return err
// 	}

// 	p.XResolution = int(binary.BigEndian.Uint32(buf))

// 	buf = make([]byte, 4)

// 	_, err = r.Read(buf)
// 	if err != nil {
// 		common.Log.Debug("Read Height block failed. %v", err)
// 		return err
// 	}

// 	p.YResolution = int(binary.BigEndian.Uint32(buf))

// 	common.Log.Debug("Page Bitmap size: Height: %v, Width: %v", p.PageBMHeight, p.PageBMWidth)

// 	flags, err := r.ReadByte()
// 	if err != nil {
// 		common.Log.Debug("Read Flags block failed. %v", err)
// 		return err
// 	}

// 	p.PageInfoFlags.SetValue(int(flags))
// 	common.Log.Debug("Flags: %d", flags)

// 	buf = make([]byte, 2)

// 	_, err = r.Read(buf)
// 	if err != nil {
// 		common.Log.Debug("Read Page Stripping block. %v", err)
// 		return err
// 	}

// 	p.pageStripping = int(binary.BigEndian.Uint16(buf))
// 	common.Log.Debug("Page Stripping: %d", p.pageStripping)

// 	defPix := p.PageInfoFlags.GetValue(DefaultPixelValue)

// 	var height int

// 	if p.PageBMHeight == -1 {
// 		height = p.pageStripping & 0x7fff
// 	} else {
// 		height = p.PageBMHeight
// 	}

// 	p.PageBitmap = bitmap.New(p.PageBMWidth, height, p.Decoders)
// 	p.PageBitmap.Clear(defPix != 0)

// 	return nil
// }
