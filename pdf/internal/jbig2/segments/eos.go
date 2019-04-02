package segments

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

// EndOfStripe flags an end of stripe JBIG2 ISO Standard 7.4.9
type EndOfStripe struct {
	r          reader.StreamReader
	lineNumber int
}

func (e *EndOfStripe) parseHeader(h *Header, r reader.StreamReader) error {
	temp, err := e.r.ReadBits(32)
	if err != nil {
		return err
	}
	e.lineNumber = int(temp)
	return nil
}

// Init implements Segmenter interface
func (e *EndOfStripe) Init(h *Header, r reader.StreamReader) error {
	e.r = r
	return e.parseHeader(h, r)
}

// LineNumber gets the EndOfStripe line number
func (e *EndOfStripe) LineNumber() int {
	return e.lineNumber
}
