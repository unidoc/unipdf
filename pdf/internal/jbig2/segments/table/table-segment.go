package table

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/header"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/model"
)

// TableSegment is the model for the JBIG2 Table segment
type TableSegment struct {
	*model.Segment
}

// New creates new TableSegment
func New(h *header.Header) *TableSegment {
	return nil
}

// Decode decodes the table segment from the reader
func (t *TableSegment) Decode(r *reader.Reader) error {
	return nil
}
