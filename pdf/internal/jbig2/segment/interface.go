package segment

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/kind"
)

// SegmentReader is the interface that reads the input read stream
// and saves the data into the SegmentReader structure
// Returns error on fail.
type SegmentReader interface {
	Decode(r *reader.Reader) error
}

type Segmenter interface {
	SegmentReader
	Kind() kind.SegmentKind
	PageAssociation() int
	Number() int
}

// SymbolDictionarySegmenter is the interface that is used by SymbolDictionarySegment to return
// exported symbols number
type SymbolDictionarySegmenter interface {
	AmmountOfExportedSymbols() int
	Segmenter
}
