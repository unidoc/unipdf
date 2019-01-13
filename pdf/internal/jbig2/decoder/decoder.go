package decoder

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/arithmetic"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/huffman"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/decoder/mmr"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"io"
)

// JBIG2Decoder decodes the stream of
type JBIG2Decoder struct {
	// internal Decoders
	arithmetic *arithmetic.Decoder
	huffman    *huffman.HuffmanDecoder
	mmr        *mmr.MmrDecoder
}

func (j *JBIG2Decoder) Decode(r io.Reader) ([]byte, error) {
	return nil, nil

}

func (j *JBIG2Decoder) readSegments(r io.Reader) ([]segment.Segmenter, error) {
	return nil, nil
}

func (j *JBIG2Decoder) readSegmentHeader(r io.Reader) (*header.Header, error) {
	// header := &segment.Header{}

	return nil, nil

}
