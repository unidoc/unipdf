package patterndict

import (
	"encoding/binary"
	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/model"
)

// PatternDicitonraySegment is the model for the JBIG2 Pattern Dictionary Segment
type PatternDictionarySegment struct {
	*model.Segment
	PDFlags *PatternDictionaryFlags
	Width   int
	Height  int
	GrayMax int

	Bitmaps []*bitmap.Bitmap
	Size    int
}

// New creates the PatternDictionarySegment for the provided header and decoders
func New(d *container.Decoder, h *header.Header) *PatternDictionarySegment {
	p := &PatternDictionarySegment{
		Segment: model.New(d, h),
		PDFlags: NewPatternDictionaryFlags(),
	}

	return p
}

// Decode decodes the Pattern Dictionary Segment from the provided reader
func (p *PatternDictionarySegment) Decode(r *reader.Reader) error {
	common.Log.Debug("[PatternDictionarySegment] Decode begin")
	defer func() { common.Log.Debug("[PatternDictionarySegment] Decode finished.") }()

	// Read the flags
	err := p.readFlags(r)
	if err != nil {
		return err
	}

	var b byte

	// Get Width
	b, err = r.ReadByte()
	if err != nil {
		common.Log.Debug("Read Width byte failed. %v", err)
		return err
	}
	p.Width = int(b)

	// Get Height
	b, err = r.ReadByte()
	if err != nil {
		common.Log.Debug("Read Height byte failed. %v", err)
		return err
	}
	p.Height = int(b)

	common.Log.Debug("PatternDictionary Size - Width: %v, Height: %v", p.Width, p.Height)

	// buf is the buffer for the reads
	var buf []byte = make([]byte, 4)

	// Read the GrayMax int from 4 bytes length
	_, err = r.Read(buf)
	if err != nil {
		common.Log.Debug("Read GrayMax failed: %v", err)
		return err
	}
	p.GrayMax = int(binary.BigEndian.Uint32(buf))

	common.Log.Debug("Gray max set to: %v", p.GrayMax)

	var (
		useMMR   bool = p.PDFlags.GetValue(HD_MMR) == 1
		template int  = p.PDFlags.GetValue(HD_TEMPLATE)
	)

	if !useMMR {
		p.Decoders.Arithmetic.ResetGenericStats(template, nil)
		if err = p.Decoders.Arithmetic.Start(r); err != nil {
			common.Log.Debug("Arithmetic Decoder Start failed: %v", err)
			return err
		}
	}

	var (
		genericBAdaptiveTemplateX []int8 = make([]int8, 4)
		genericBAdaptiveTemplateY []int8 = make([]int8, 4)
	)

	genericBAdaptiveTemplateX[0] = int8(-p.Width)
	genericBAdaptiveTemplateY[0] = 0
	genericBAdaptiveTemplateX[1] = -3
	genericBAdaptiveTemplateY[1] = -1
	genericBAdaptiveTemplateX[2] = 2
	genericBAdaptiveTemplateY[2] = -2
	genericBAdaptiveTemplateX[3] = -2
	genericBAdaptiveTemplateY[3] = -2

	p.Size = p.GrayMax + 1

	// Read the bitmap with all patterns
	bm := bitmap.New(p.Size*p.Width, p.Height, p.Decoders)
	bm.Clear(false)
	err = bm.Read(r, useMMR, template, false, false, nil,
		genericBAdaptiveTemplateX, genericBAdaptiveTemplateY, p.Header.DataLength-7)
	if err != nil {
		common.Log.Debug("Read Bitmap failed: %v", err)
		return err
	}

	bitmaps := make([]*bitmap.Bitmap, p.Size)

	var x int
	for i := 0; i < p.Size; i++ {
		bitmaps[i], err = bm.GetSlice(x, 0, p.Width, p.Height)
		if err != nil {
			common.Log.Debug("GettingSlice of '%d' bitmap failed. %v", i, err)
			return err
		}
		x += p.Width
	}

	// Set the bitmap slices as the PatternDictionary Bitmaps
	p.Bitmaps = bitmaps

	return nil
}

// readFlags reads the flags for the PatternDictionarySegment
func (p *PatternDictionarySegment) readFlags(r *reader.Reader) error {
	// Read flags byte from reader
	flag, err := r.ReadByte()
	if err != nil {
		return err
	}

	// SetValue for flags
	p.PDFlags.SetValue(int(flag))

	common.Log.Debug("Pattern Dictionary Flags: %v ", p.PDFlags)
	return nil
}
