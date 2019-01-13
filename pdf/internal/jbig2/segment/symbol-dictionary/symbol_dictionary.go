package symboldict

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/bitmap"
	"github.com/unidoc/unidoc/pdf/internal/jbig2/segment/model"
)

type SymbolDictionarySegment struct {
	*model.Segment
}

func (s *SymbolDictionarySegment) ExportedSymbolsNumber() int {
	return int
}

func (s *SymbolDictionarySegment) ListBitmaps() []*bitmap.Bitmap {

}
