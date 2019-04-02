package segments

import (
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// PageInformationSegment represents the segment type Page Information 7.4.8
type PageInformationSegment struct {
	r reader.StreamReader

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

func newPageInformation(h *Header) *PageInformationSegment {
	p := &PageInformationSegment{}

	return p
}

// Init initialize the PageInformation Segment
func (p *PageInformationSegment) Init(h *Header, r reader.StreamReader) error {
	p.r = r
	p.parseHeader()
	return nil
}

// CombinationOperator gets the combination operator for the page
func (p *PageInformationSegment) CombinationOperator() bitmap.CombinationOperator {
	return p.combinationOperator
}

// CombinationOperatorOverrideAllowed defines if the Page segment has allowed override
func (p *PageInformationSegment) CombinationOperatorOverrideAllowed() bool {
	return p.combinaitonOperatorOverrideAllowed
}

// DefaultPixelValue returns page segment default pixel
func (p *PageInformationSegment) DefaultPixelValue() uint8 {
	return p.defaultPixelValue
}

func (p *PageInformationSegment) parseHeader() (err error) {
	common.Log.Debug("[PageInformationSegment] ParsingHeader...")
	defer func() {
		var str = "[PageInformationSegment] ParsingHeader Finished"
		if err != nil {
			str += " with error"
		} else {
			str += " succesfully"
		}
		common.Log.Debug(str)
	}()
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

func (p *PageInformationSegment) init(header *Header, r *reader.Reader) error {
	p.r = r
	return nil

}
