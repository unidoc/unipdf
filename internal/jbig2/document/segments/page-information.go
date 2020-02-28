/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
	"github.com/unidoc/unipdf/v3/internal/jbig2/reader"
	"github.com/unidoc/unipdf/v3/internal/jbig2/writer"
)

// PageInformationSegment represents the segment type Page Information 7.4.8.
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
	IsLossless                         bool

	// Page striping information, two byte 7.4.8.6
	IsStripe      bool
	MaxStripeSize uint16
}

// Encode implements SegmentEncoder interface.
func (p *PageInformationSegment) Encode(w writer.BinaryWriter) (n int, err error) {
	const processName = "PageInformationSegment.Encode"
	tm := make([]byte, 4)

	// encode page bitmap width 7.4.8.1
	binary.BigEndian.PutUint32(tm, uint32(p.PageBMWidth))
	n, err = w.Write(tm)
	if err != nil {
		return n, errors.Wrap(err, processName, "width")
	}

	// encode page bitmap height 7.4.8.2
	binary.BigEndian.PutUint32(tm, uint32(p.PageBMHeight))
	var temp int
	temp, err = w.Write(tm)
	if err != nil {
		return temp + n, errors.Wrap(err, processName, "height")
	}
	n += temp

	// encode page x resolution - 7.4.8.3
	binary.BigEndian.PutUint32(tm, uint32(p.ResolutionX))
	temp, err = w.Write(tm)
	if err != nil {
		return temp + n, errors.Wrap(err, processName, "x resolution")
	}
	n += temp

	// encode page y resolution - 7.4.8.4
	binary.BigEndian.PutUint32(tm, uint32(p.ResolutionY))
	if temp, err = w.Write(tm); err != nil {
		return temp + n, errors.Wrap(err, processName, "y resolution")
	}
	n += temp

	// page flags - 7.4.8.5
	if err = p.encodeFlags(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n++

	// page striping information - 7.4.8.6
	if temp, err = p.encodeStripingInformation(w); err != nil {
		return n, errors.Wrap(err, processName, "")
	}
	n += temp
	return n, nil
}

// Init implements Segmenter interface.
func (p *PageInformationSegment) Init(h *Header, r reader.StreamReader) (err error) {
	p.r = r
	if err = p.parseHeader(); err != nil {
		return errors.Wrap(err, "PageInformationSegment.Init", "")
	}
	return nil
}

// CombinationOperator gets the combination operator used by the page information segment.
func (p *PageInformationSegment) CombinationOperator() bitmap.CombinationOperator {
	return p.combinationOperator
}

// CombinationOperatorOverrideAllowed defines if the Page segment has allowed override.
func (p *PageInformationSegment) CombinationOperatorOverrideAllowed() bool {
	return p.combinaitonOperatorOverrideAllowed
}

// DefaultPixelValue returns page segment default pixel.
func (p *PageInformationSegment) DefaultPixelValue() uint8 {
	return p.defaultPixelValue
}

// Size returns the byte size of the page information segment data.
// The page information segment is composed of:
// - bitmap width 			- 4 bytes
// - bitmap height 			- 4 bytes
// - x resolution 			- 4 bytes
// - y resolution 			- 4 bytes
// - flags 					- 1 byte
// - stripping information 	- 2 bytes
// Total 19 bytes
func (p *PageInformationSegment) Size() int {
	return 19
}

// String implements Stringer interface.
func (p *PageInformationSegment) String() string {
	sb := &strings.Builder{}

	sb.WriteString("\n[PAGE-INFORMATION-SEGMENT]\n")
	sb.WriteString(fmt.Sprintf("\t- BMHeight: %d\n", p.PageBMHeight))
	sb.WriteString(fmt.Sprintf("\t- BMWidth: %d\n", p.PageBMWidth))
	sb.WriteString(fmt.Sprintf("\t- ResolutionX: %d\n", p.ResolutionX))
	sb.WriteString(fmt.Sprintf("\t- ResolutionY: %d\n", p.ResolutionY))
	sb.WriteString(fmt.Sprintf("\t- CombinationOperator: %s\n", p.combinationOperator))
	sb.WriteString(fmt.Sprintf("\t- CombinationOperatorOverride: %v\n", p.combinaitonOperatorOverrideAllowed))
	sb.WriteString(fmt.Sprintf("\t- IsLossless: %v\n", p.IsLossless))
	sb.WriteString(fmt.Sprintf("\t- RequiresAuxiliaryBuffer: %v\n", p.requiresAuxiliaryBuffer))
	sb.WriteString(fmt.Sprintf("\t- MightContainRefinements: %v\n", p.mightContainRefinements))
	sb.WriteString(fmt.Sprintf("\t- IsStriped: %v\n", p.IsStripe))
	sb.WriteString(fmt.Sprintf("\t- MaxStripeSize: %v\n", p.MaxStripeSize))
	return sb.String()
}

func (p *PageInformationSegment) checkInput() error {
	if p.PageBMHeight == math.MaxInt32 {
		if !p.IsStripe {
			common.Log.Debug("PageInformationSegment.IsStripe should be true.")
		}
	}
	return nil
}

func (p *PageInformationSegment) encodeFlags(w writer.BinaryWriter) (err error) {
	const processName = "encodeFlags"
	// reserved bits
	if err = w.SkipBits(1); err != nil {
		return errors.Wrap(err, processName, "reserved bit")
	}
	// combination operator override
	var bit int
	if p.CombinationOperatorOverrideAllowed() {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return errors.Wrap(err, processName, "combination operator overridden")
	}

	// requires auxiliary buffer
	bit = 0
	if p.requiresAuxiliaryBuffer {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return errors.Wrap(err, processName, "requires auxiliary buffer")
	}

	// default page combination operator - 2 bits
	if err = w.WriteBit((int(p.combinationOperator) >> 1) & 0x01); err != nil {
		return errors.Wrap(err, processName, "combination operator first bit")
	}
	if err = w.WriteBit(int(p.combinationOperator) & 0x01); err != nil {
		return errors.Wrap(err, processName, "combination operator second bit")
	}

	// default page pixel value
	bit = int(p.defaultPixelValue)
	if err = w.WriteBit(bit); err != nil {
		return errors.Wrap(err, processName, "default page pixel value")
	}

	// contains refinement
	bit = 0
	if p.mightContainRefinements {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return errors.Wrap(err, processName, "contains refinement")
	}

	// page is eventually lossless
	bit = 0
	if p.IsLossless {
		bit = 1
	}
	if err = w.WriteBit(bit); err != nil {
		return errors.Wrap(err, processName, "page is eventually lossless")
	}
	return nil
}

func (p *PageInformationSegment) encodeStripingInformation(w writer.BinaryWriter) (n int, err error) {
	const processName = "encodeStripingInformation"
	if !p.IsStripe {
		// if there is no page striping write two empty bytes
		if n, err = w.Write([]byte{0x00, 0x00}); err != nil {
			return 0, errors.Wrap(err, processName, "no striping")
		}
		return n, nil
	}
	temp := make([]byte, 2)
	binary.BigEndian.PutUint16(temp, p.MaxStripeSize|1<<15)
	if n, err = w.Write(temp); err != nil {
		return 0, errors.Wrapf(err, processName, "striping: %d", p.MaxStripeSize)
	}
	return n, nil
}

func (p *PageInformationSegment) parseHeader() (err error) {
	common.Log.Trace("[PageInformationSegment] ParsingHeader...")
	defer func() {
		var str = "[PageInformationSegment] ParsingHeader Finished"
		if err != nil {
			str += " with error " + err.Error()
		} else {
			str += " successfully"
		}
		common.Log.Trace(str)
	}()

	if err = p.readWidthAndHeight(); err != nil {
		return err
	}

	if err = p.readResolution(); err != nil {
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
	common.Log.Trace("%s", p)
	return nil
}

func (p *PageInformationSegment) readResolution() error {
	tempResolution, err := p.r.ReadBits(32)
	if err != nil {
		return err
	}
	p.ResolutionX = int(tempResolution & math.MaxInt32)

	tempResolution, err = p.r.ReadBits(32)
	if err != nil {
		return err
	}

	p.ResolutionY = int(tempResolution & math.MaxInt32)
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
	p.defaultPixelValue = uint8(b & 0xf)
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
		p.IsLossless = true
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

	p.MaxStripeSize = uint16(b & math.MaxUint16)
	return nil
}

func (p *PageInformationSegment) readWidthAndHeight() error {
	tempInt, err := p.r.ReadBits(32)
	if err != nil {
		return err
	}
	p.PageBMWidth = int(tempInt & math.MaxInt32)

	tempInt, err = p.r.ReadBits(32)
	if err != nil {
		return err
	}
	p.PageBMHeight = int(tempInt & math.MaxInt32)
	return nil
}
