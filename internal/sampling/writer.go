package sampling

import (
	"github.com/unidoc/unipdf/v3/internal/bitwise"
	"github.com/unidoc/unipdf/v3/internal/imageutil"
)

// SampleWriter allows to write single or multiple samples at once.
type SampleWriter interface {
	WriteSample(sample uint32) error
	WriteSamples(samples []uint32) error
}

// Writer allows to get samples from provided image
type Writer struct {
	img          imageutil.ImageBase
	w            *bitwise.Writer
	x, component int
	hasPadding   bool
}

// NewWriter creates new image based sample reader with specified bitsPerOutputSample.
func NewWriter(img imageutil.ImageBase) *Writer {
	return &Writer{w: bitwise.NewWriterMSB(img.Data), img: img, component: img.ColorComponents, hasPadding: img.BytesPerLine*8 != img.ColorComponents*img.BitsPerComponent*img.Width}
}

// WriteSample writes the sample into given writer.
func (w *Writer) WriteSample(sample uint32) error {
	if _, err := w.w.WriteBits(uint64(sample), w.img.BitsPerComponent); err != nil {
		return err
	}
	w.component--
	if w.component == 0 {
		w.component = w.img.ColorComponents
		w.x++
	}
	if w.x == w.img.Width {
		if w.hasPadding {
			w.w.FinishByte()
		}
		w.x = 0
	}
	return nil
}

// WriteSamples writes multiple samples at once.
func (w *Writer) WriteSamples(samples []uint32) error {
	for i := 0; i < len(samples); i++ {
		if err := w.WriteSample(samples[i]); err != nil {
			return err
		}
	}
	return nil
}
