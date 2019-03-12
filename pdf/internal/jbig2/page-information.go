package jbig2

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// PageInformationSegment represents the segment type Page Information 7.4.8
type PageInformationSegment struct {
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
	IsStripe      bool
	MaxStripeSize uint16
}

func New(h *SegmentHeader) *PageInformationSegment {
	p := &PageInformationSegment{}
	return p
}

func (p *PageInformationSegment) Init() error {
	return nil
}

func (p *PageInformationSegment) parseHeader() (err error) {
	if err = p.readWidthAndHeight(); err != nil {
		return err
	}

	if err = p.readResoultion(); err != nil {
		return err
	}

	// Bit 7 - dirty read
	_, err = p.r.ReadBit()
	if err != nil {
		return err
	}

	// Bit 6
	if err = p.readCombinationOperatorOverrideAllowed(); err != nil {
		return err
	}

	// Bit 5
	if err = p.readRequiresAuxiliaryBuffer(); err != nil {
		return err
	}

	// Bit 3-4
	if err = p.readCombinationOperator(); err != nil {
		return err
	}

	// Bit 2
	if err = p.readDefaultPixelValue(); err != nil {
		return err
	}

	// Bit 1
	if err = p.readContainsRefinement(); err != nil {
		return err
	}

	// Bit 0
	if err = p.readIsLossless(); err != nil {
		return err
	}

	// Bit 15
	if err = p.readIsStriped(); err != nil {
		return err
	}

	// Bit 0 - 14
	if err = p.readMaxStripeSize(); err != nil {
		return err
	}

	if err = p.checkInput(); err != nil {
		return err
	}

	return nil
}

func (p *PageInformationSegment) readResoultion() error {
	tempResolution, err := p.r.ReadBits(32)
	if err != nil {
		return err
	}
	p.ResolutionX = int(tempResolution)

	tempResolution, err = p.r.ReadBits(32)
	if err != nil {
		return err
	}

	p.ResolutionY = int(tempResolution)

	return nil
}

func (p *PageInformationSegment) checkInput() error {
	if p.PageBMHeight == 0xFFFFFFFFFF {
		if !p.IsStripe {
			common.Log.Debug("isStriped should contaion the value true")
		}
	}
	return nil
}

// readCombinationOperatorOverrideAllowed -  Bit 6
func (p *PageInformationSegment) readCombinationOperatorOverrideAllowed() error {

	b, err := p.r.ReadBit()
	if err != nil {
		return err
	}

	if b == 1 {
		p.combinaitonOperatorOverrideAllowed = true
	}
	return nil
}

// readRequiresAuxiliaryBuffer - Bit 5
func (p *PageInformationSegment) readRequiresAuxiliaryBuffer() error {
	b, err := p.r.ReadBit()
	if err != nil {
		return err
	}

	if b == 1 {
		p.requiresAuxiliaryBuffer = true
	}
	return nil
}

// readCombinationOperator - Bit 3 -4
func (p *PageInformationSegment) readCombinationOperator() error {
	b, err := p.r.ReadBits(2)
	if err != nil {
		return err
	}
	p.combinationOperator = bitmap.CombinationOperator(int(b))
	return nil
}

// readDefaultPixelValue - Bit 2
func (p *PageInformationSegment) readDefaultPixelValue() error {
	b, err := p.r.ReadBit()
	if err != nil {
		return err
	}
	p.defaultPixelValue = uint8(b)

	return nil
}

// readContainsRefinement - Bit 1
func (p *PageInformationSegment) readContainsRefinement() error {
	b, err := p.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		p.mightContainRefinements = true
	}
	return nil
}

// readIsLossless - Bit 0
func (p *PageInformationSegment) readIsLossless() error {
	b, err := p.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		p.isLossless = true
	}

	return nil
}

// readIsStriped - Bit 15
func (p *PageInformationSegment) readIsStriped() error {
	b, err := p.r.ReadBit()
	if err != nil {
		return err
	}
	if b == 1 {
		p.IsStripe = true
	}

	return nil
}

// readMaxStripeSize - Bit 0-14
func (p *PageInformationSegment) readMaxStripeSize() error {
	b, err := p.r.ReadBits(15)
	if err != nil {
		return err
	}

	p.MaxStripeSize = uint16(b & 0xFFFF)

	return nil
}

func (p *PageInformationSegment) readWidthAndHeight() error {
	tempInt, err := p.r.ReadBits(32)
	if err != nil {
		return err
	}
	p.PageBMWidth = int(tempInt)

	tempInt, err = p.r.ReadBits(32)
	if err != nil {
		return err
	}
	p.PageBMHeight = int(tempInt)
	return nil
}

func (p *PageInformationSegment) init(header *SegmentHeader, r *reader.Reader) error {
	p.r = r
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
