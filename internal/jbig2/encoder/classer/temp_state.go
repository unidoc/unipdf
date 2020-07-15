/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
)

// similarTemplatesFinder stores the state of a state machine which fetches similar sized templates.
type similarTemplatesFinder struct {
	Classer *Classer
	// Desired width
	Width int
	// Desired height
	Height int
	// Index into two_by_two step array
	Index int
	// Current number array
	CurrentNumbers []int
	// Current element of the array
	N int
}

// initSimilarTemplatesFinder initializes the templatesState context.
func initSimilarTemplatesFinder(c *Classer, bms *bitmap.Bitmap) *similarTemplatesFinder {
	return &similarTemplatesFinder{
		Width:   bms.Width,
		Height:  bms.Height,
		Classer: c,
	}
}

// Next finds next template state.
func (f *similarTemplatesFinder) Next() int {
	var (
		desireDH, desireDW, size, templ int
		ok                              bool
		bmT                             *bitmap.Bitmap
		err                             error
	)

	for {
		if f.Index >= 25 {
			return -1
		}
		desireDW = f.Width + TwoByTwoWalk[2*f.Index]
		desireDH = f.Height + TwoByTwoWalk[2*f.Index+1]
		if desireDH < 1 || desireDW < 1 {
			f.Index++
			continue
		}

		if len(f.CurrentNumbers) == 0 {
			f.CurrentNumbers, ok = f.Classer.TemplatesSize.GetSlice(uint64(desireDW) * uint64(desireDH))
			if !ok {
				f.Index++
				continue
			}
			f.N = 0
		}
		size = len(f.CurrentNumbers)
		for ; f.N < size; f.N++ {
			templ = f.CurrentNumbers[f.N]
			bmT, err = f.Classer.DilatedTemplates.GetBitmap(templ)
			if err != nil {
				common.Log.Debug("FindNextTemplate: template not found: ")
				return 0
			}
			if bmT.Width-2*JbAddedPixels == desireDW && bmT.Height-2*JbAddedPixels == desireDH {
				return templ
			}
		}
		f.Index++
		f.CurrentNumbers = nil
	}
}
