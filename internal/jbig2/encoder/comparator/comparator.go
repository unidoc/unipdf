package comparator

import (
	"github.com/unidoc/unipdf/internal/jbig2/bitmap"
)

// AreEquivalent compares two pix and tell if they are equivalent by trying
// to decide if these symbols are equivalent from visual point of view.
func AreEquivalent(firstTemplate, secondTemplate *bitmap.Pix) bool {
	// TODO: jbig2comparator.c:45
	return false
}
