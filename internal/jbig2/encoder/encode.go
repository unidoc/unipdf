/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package encoder

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/document"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// Encoder encodes the data into jbig2 encoding.
type Encoder struct {
	// some paramete
	Document    *document.Document
	FullHeaders bool
}

// AddPage adds the page to the current encoder context.
func (e *Encoder) AddPage(bitmap *bitmap.Bitmap, settings Parameters) error {
	if err := settings.Validate(); err != nil {
		return errors.Wrap(err, "NewEncoder", "")
	}
	return nil
}

// EncodeBytes encodes input 'data' and encoding 'parameters' into jbig2 encoding.
func (e *Encoder) EncodeBytes(data []byte, settings Parameters) ([]byte, error) {
	const processName = "Encoder.EncodeBytes"
	if err := settings.Validate(); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	// try to create a bitmap from the provided byte stream.
	bm, err := bitmap.NewWithData(settings.Width, settings.Height, data)
	if err != nil {
		return nil, errors.Wrap(err, processName, "can't create a bitmap from the input data")
	}

	if err := e.addPage(bm, settings); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}

	return nil, nil
}

func (e *Encoder) addPage(b *bitmap.Bitmap, settings Parameters) (err error) {
	const processName = "addPage"

	if err = settings.Validate(); err != nil {
		return errors.Wrap(err, "addPage", "")
	}
	switch settings.Method {
	case Generic:
		if err = e.Document.AddGenericPage(b, settings.DuplicateLineRemoval); err != nil {
			return errors.Wrap(err, processName, "")
		}
	case ClassifyCorrelation:
		return errors.Errorf(processName, "add correlation classified page not implemented yet")
	case ClassifyRankHaus:
		return errors.Errorf(processName, "add ranked hausdorff page not implemented yet")
	}
	return nil
}

func (e *Encoder) encode() ([]byte, error) {
	return nil, nil
}

// Method is the jbig2 encode method.
type Method int

const (
	// Generic is the encoding method using Generic Region Segment.
	Generic Method = iota
	// ClassifyCorrelation is the encoding method using classifier based on the image correlation.
	ClassifyCorrelation
	// ClassifyRankHaus is the encoding mehod using classifier based on the image rank hausdorff.
	ClassifyRankHaus
)

// Parameters is the struct that contains jbig2 encoder parameters.
type Parameters struct {
	Method                   Method
	ResolutionX, ResolutionY int
	Threshold                float64
	Rank                     float64
	Width, Height            int
	IsBitmap                 bool
	DuplicateLineRemoval     bool
}

// Validate the input parameters 'p'.
func (p Parameters) Validate() error {
	const processName = "Parameters.Validate"
	switch p.Method {
	case Generic:
	case ClassifyCorrelation:
		// check if the threshold is valid
		return errors.Error(processName, "ClassifyCorrelation method is not implemented yet")
	case ClassifyRankHaus:
		// check if the rank is valid
		return errors.Error(processName, "ClassifyRankHaus method is not implemented yet")
	default:
	}
	if p.ResolutionX < 0 || p.ResolutionY < 0 {
		return errors.Error(processName, "resolution must be a positive value")
	}
	if p.Width < 0 || p.Height < 0 {
		return errors.Error(processName, "width and height must be a positive value")
	}
	// if the input data is provided as a byte slice not a bitmap then check if the width and height are set.
	if !p.IsBitmap && p.Width == 0 || p.Height == 0 {
		return errors.Error(processName, "width and height non defined")
	}
	return nil
}
