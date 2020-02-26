/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// Settings keeps the settings for the classer.
type Settings struct {
	// MaxCompWidth is max component width allowed.
	MaxCompWidth int
	// MaxCompHeight is max component height allowed.
	MaxCompHeight int
	// SizeHaus is the size of square struct elem for hausdorf method.
	SizeHaus int
	// Rank val of hausdorf method match.
	RankHaus float64
	// Thresh is the threshold value for the correlation score.
	Thresh float64
	// Corrects thresh value for heavier components; 0 for no correction.
	WeightFactor float64
	// KeepClassInstances is a flag that defines if the class instances should be stored
	// in the 'ClassInstances' BitmapsArray.
	KeepClassInstances bool
	// Components is the setting the classification.
	Components bitmap.Component
	// Method is the encoding method.
	Method Method
}

// DefaultSettings returns default settings struct.
func DefaultSettings() Settings {
	s := &Settings{}
	s.SetDefault()
	return *s
}

// SetDefault sets the default value for the settings.
func (s *Settings) SetDefault() {
	// if max width is not defined get the value from the constants.
	if s.MaxCompWidth == 0 {
		switch s.Components {
		case bitmap.ComponentConn:
			s.MaxCompWidth = MaxConnCompWidth
		case bitmap.ComponentCharacters:
			s.MaxCompWidth = MaxCharCompWidth
		case bitmap.ComponentWords:
			s.MaxCompWidth = MaxWordCompWidth
		}
	}
	// if max height is not defined take the 'MaxCompHeight' value.
	if s.MaxCompHeight == 0 {
		s.MaxCompHeight = MaxCompHeight
	}

	if s.Thresh == 0.0 {
		s.Thresh = 0.9
	}
	if s.WeightFactor == 0.0 {
		s.WeightFactor = 0.75
	}
	if s.RankHaus == 0.0 {
		s.RankHaus = 0.97
	}
	if s.SizeHaus == 0 {
		s.SizeHaus = 2
	}
}

// Validate validates the settings input.
func (s Settings) Validate() error {
	const processName = "Settings.Validate"
	if s.Thresh < 0.4 || s.Thresh > 0.98 {
		return errors.Error(processName, "jbig2 encoder thresh not in range [0.4 - 0.98]")
	}
	if s.WeightFactor < 0.0 || s.WeightFactor > 1.0 {
		return errors.Error(processName, "jbig2 encoder weight factor not in range [0.0 - 1.0]")
	}
	if s.RankHaus < 0.5 || s.RankHaus > 1.0 {
		return errors.Error(processName, "jbig2 encoder rank haus value not in range [0.5 - 1.0]")
	}
	if s.SizeHaus < 1 || s.SizeHaus > 10 {
		return errors.Error(processName, "jbig2 encoder size haus value not in range [1 - 10]")
	}
	switch s.Components {
	case bitmap.ComponentConn, bitmap.ComponentCharacters, bitmap.ComponentWords:
	default:
		return errors.Error(processName, "invalid classer component")
	}
	return nil
}
