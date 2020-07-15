/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package sampling

import (
	"io"

	"github.com/unidoc/unipdf/v3/internal/bitwise"
	"github.com/unidoc/unipdf/v3/internal/imageutil"
)

// SampleReader is an interface that allows to read samples.
type SampleReader interface {
	ReadSample() (uint32, error)
	ReadSamples(samples []uint32) error
}

// Reader allows to get samples from provided image
type Reader struct {
	img             imageutil.ImageBase
	r               *bitwise.Reader
	x, y, component int
	hasPadding      bool
}

// NewReader creates new image based sample reader with specified bitsPerOutputSample.
func NewReader(img imageutil.ImageBase) *Reader {
	return &Reader{r: bitwise.NewReader(img.Data), img: img, component: img.ColorComponents, hasPadding: img.BytesPerLine*8 != img.ColorComponents*img.BitsPerComponent*img.Width}
}

// ReadSample gets next sample from the reader. If there is no more samples, the function returns an error.
func (r *Reader) ReadSample() (uint32, error) {
	if r.y == r.img.Height {
		return 0, io.EOF
	}
	bits, err := r.r.ReadBits(byte(r.img.BitsPerComponent))
	if err != nil {
		return 0, err
	}
	r.component--
	if r.component == 0 {
		r.component = r.img.ColorComponents
		r.x++
	}
	if r.x == r.img.Width {
		if r.hasPadding {
			r.r.ConsumeRemainingBits()
		}
		r.x = 0
		r.y++
	}
	return uint32(bits), nil
}

// ReadSamples reads as many samples as possible up to the size of provided slice.
func (r *Reader) ReadSamples(samples []uint32) (err error) {
	for i := 0; i < len(samples); i++ {
		samples[i], err = r.ReadSample()
		if err != nil {
			return err
		}
	}
	return nil
}
