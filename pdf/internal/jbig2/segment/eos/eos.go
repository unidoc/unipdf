package eos

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/model"
)

type EndOfStripeSegment struct {
	*model.Segment
}

func New(d *container.Decoder, h *header.Header) *EndOfStripeSegment {
	return nil
}

//
func (e *EndOfStripeSegment) Decode(r *reader.Reader) error {
	return nil
}
