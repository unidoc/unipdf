package model

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/container"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
)

// Segment defines jbig2 segment base structure
type Segment struct {
	Header *header.Header

	Decoders *container.Decoder
}

// New creates new Segment with the reference for the decoders container
func New(decoders *container.Decoder, h *header.Header) *Segment {
	return &Segment{
		Decoders: decoders,
		Header:   h,
	}
}

// PageAssociation returns page association
// Implements SegmentPageAssociator
func (s *Segment) PageAssociation() int {
	return s.Header.PageAssociation
}

func (s *Segment) Kind() kind.SegmentKind {
	return kind.SegmentKind(s.Header.SegmentType)
}

// Number returns segment number
func (s *Segment) Number() int {
	return s.Header.SegmentNumber
}
