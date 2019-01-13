package container

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/huffman"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/mmr"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
)

// Decoders is the container for the common jbig2 decoders instances
type Decoder struct {
	Arithmetic *arithmetic.Decoder
	Huffman    *huffman.HuffmanDecoder
	MMR        *mmr.MmrDecoder

	Segments []segment.Segmenter
}

func New() *Decoder {
	return &Decoder{
		Arithmetic: arithmetic.New(),
		Huffman:    huffman.New(),
		MMR:        mmr.New(),
	}
}

// FindPageSegment searches for the page information segment int the segments slice
func (d *Decoder) FindPageSegment(page int) segment.Segmenter {
	for _, s := range d.Segments {
		if s.Kind() == kind.PageInformation && s.PageAssociation() == page {
			return s
		}
	}
	return nil
}

// FindSegment finds the segment by the number provided as the argument
func (d *Decoder) FindSegment(number int) segment.Segmenter {
	for _, s := range d.Segments {
		if s.Number() == number {
			return s
		}
	}
	return nil
}
