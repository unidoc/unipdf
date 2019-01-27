package extension

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/model"
	"io"
)

type ExtensionSegment struct {
	*model.Segment
}

func New(h *header.Header) *ExtensionSegment {
	return nil
}

// Decode decodes the extension segment. Implements the Segmenter interface method
func (e *ExtensionSegment) Decode(r *reader.Reader) error {
	_, err := r.Seek(int64(e.Header.DataLength), io.SeekCurrent)
	return err
}
