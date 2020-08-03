/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package classer

import (
	"image"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
	"github.com/unidoc/unipdf/v3/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/errors"
)

// Classer holds all the data accumulated during the classification
// process that can be used for a compressed jbig2-type representation
// of a set of images.
type Classer struct {
	// BaseIndex is number of components already processed on fully processed pages.
	BaseIndex int
	// Settings are current classer settings.
	Settings Settings

	// Number of components on each page - 'nacomps'- for each page added to the classer a new entry to the slice
	// is added with the value of components per page.
	ComponentsNumber *basic.IntSlice
	// Width * Height of each template without extra border pixels - 'naarea'.
	TemplateAreas *basic.IntSlice

	// Widths is max width of original src images.
	Widths map[int]int
	// Heights is max height of original src images.
	Heights map[int]int

	// NumberOfClasses is the current number of classes - 'nclass'.
	NumberOfClasses int
	// ClassInstances is the slice of bitmaps for each class. Unbordered - 'pixaa'.
	ClassInstances *bitmap.BitmapsArray
	// UndilatedTemplates for each class. Bordered and not dilated - 'pixat'.
	UndilatedTemplates *bitmap.Bitmaps
	// DilatedTemplates for each class. Bordered and dilated - 'pixatd'.
	DilatedTemplates *bitmap.Bitmaps

	// Hash table to find templates by their size - 'dahash'.
	TemplatesSize basic.IntsMap
	// FgTemplates - foreground areas of undilated templates. Used for rank < 1.0 - 'nafgt'.
	FgTemplates *basic.NumSlice

	// CentroidPoints centroids of all bordered cc.
	CentroidPoints *bitmap.Points
	// CentroidPointsTemplates centroids of all bordered template cc.
	CentroidPointsTemplates *bitmap.Points
	// ClassIDs is the slice of class ids for each component - 'naclass'.
	ClassIDs *basic.IntSlice
	// ComponentPageNumbers is the slice of page numbers for each component - 'napage'.
	// The index is the component id.
	ComponentPageNumbers *basic.IntSlice
	// PtaUL is the slice of UL corners at which the template
	// is to be placed for each component.
	PtaUL *bitmap.Points
	// PtaLL is the slice of LL corners at which the template
	// is to be placed for each component.
	PtaLL *bitmap.Points
}

// Init initializes the classer with the provided settings.
func Init(settings Settings) (*Classer, error) {
	const processName = "classer.Init"
	c := &Classer{
		Settings:                settings,
		Widths:                  map[int]int{},
		Heights:                 map[int]int{},
		TemplatesSize:           basic.IntsMap{},
		TemplateAreas:           &basic.IntSlice{},
		ComponentPageNumbers:    &basic.IntSlice{},
		ClassIDs:                &basic.IntSlice{},
		ComponentsNumber:        &basic.IntSlice{},
		CentroidPoints:          &bitmap.Points{},
		CentroidPointsTemplates: &bitmap.Points{},
		UndilatedTemplates:      &bitmap.Bitmaps{},
		DilatedTemplates:        &bitmap.Bitmaps{},
		ClassInstances:          &bitmap.BitmapsArray{},
		FgTemplates:             &basic.NumSlice{},
	}
	if err := c.Settings.Validate(); err != nil {
		return nil, errors.Wrap(err, processName, "")
	}
	return c, nil
}

// AddPage adds the 'inputPage' to the classer 'c'.
func (c *Classer) AddPage(inputPage *bitmap.Bitmap, pageNumber int, method Method) (err error) {
	const processName = "Classer.AddPage"
	c.Widths[pageNumber] = inputPage.Width
	c.Heights[pageNumber] = inputPage.Height

	if err = c.verifyMethod(method); err != nil {
		return errors.Wrap(err, processName, "")
	}

	comps, boxes, err := inputPage.GetComponents(c.Settings.Components, c.Settings.MaxCompWidth, c.Settings.MaxCompHeight)
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	common.Log.Debug("Components: %v", comps)
	// add the computed components to the page using provided method.
	if err = c.addPageComponents(inputPage, boxes, comps, pageNumber, method); err != nil {
		return errors.Wrap(err, processName, "")
	}
	return nil
}

// ComputeLLCorners computes the position of the LL (lower left) corners.
func (c *Classer) ComputeLLCorners() (err error) {
	const processName = "Classer.ComputeLLCorners"
	if c.PtaUL == nil {
		return errors.Error(processName, "UL Corners not defined")
	}
	n := len(*c.PtaUL)

	c.PtaLL = &bitmap.Points{}

	var (
		x1, y1    float32
		iClass, h int
		bm        *bitmap.Bitmap
	)
	for i := 0; i < n; i++ {
		x1, y1, err = c.PtaUL.GetGeometry(i)
		if err != nil {
			common.Log.Debug("Getting PtaUL failed: %v", err)
			return errors.Wrap(err, processName, "PtaUL Geometry")
		}
		iClass, err = c.ClassIDs.Get(i)
		if err != nil {
			common.Log.Debug("Getting ClassID failed: %v", err)
			return errors.Wrap(err, processName, "ClassID")
		}
		bm, err = c.UndilatedTemplates.GetBitmap(iClass)
		if err != nil {
			common.Log.Debug("Getting UndilatedTemplates failed: %v", err)
			return errors.Wrap(err, processName, "Undilated Templates")
		}
		h = bm.Height
		// Add the global LL corner point.
		c.PtaLL.AddPoint(x1, y1+float32(h)) // )-1-2*float32(JbAddedPixels)
	}
	return nil
}

/**

Private methods and functions

*/

// addPageComponents adds the components to the 'inputPage'.
func (c *Classer) addPageComponents(inputPage *bitmap.Bitmap, boxas *bitmap.Boxes, components *bitmap.Bitmaps, pageNumber int, method Method) error {
	const processName = "Classer.AddPageComponents"
	if inputPage == nil {
		return errors.Error(processName, "nil input page")
	}
	if boxas == nil || components == nil || len(*boxas) == 0 {
		common.Log.Trace("AddPageComponents: %s. No components found", inputPage)
		return nil
	}
	var err error
	switch method {
	case RankHaus:
		err = c.classifyRankHaus(boxas, components, pageNumber)
	case Correlation:
		err = c.classifyCorrelation(boxas, components, pageNumber)
	default:
		common.Log.Debug("Unknown classify method: '%v'", method)
		return errors.Error(processName, "unknown classify method")
	}
	if err != nil {
		return errors.Wrap(err, processName, "")
	}

	if err = c.getULCorners(inputPage, boxas); err != nil {
		return errors.Wrap(err, processName, "")
	}
	n := len(*boxas)
	c.BaseIndex += n
	if err = c.ComponentsNumber.Add(n); err != nil {
		return errors.Wrap(err, processName, "")
	}
	return nil
}

// getULCorners get the ul corners.
func (c *Classer) getULCorners(s *bitmap.Bitmap, boxa *bitmap.Boxes) error {
	const processName = "getULCorners"
	if s == nil {
		return errors.Error(processName, "nil image bitmap")
	}
	if boxa == nil {
		return errors.Error(processName, "nil bounds")
	}
	if c.PtaUL == nil {
		c.PtaUL = &bitmap.Points{}
	}

	n := len(*boxa)
	var (
		index, iClass, idelX, idelY int
		x1, y1, x2, y2              float32
		err                         error
		box                         *image.Rectangle
		t                           *bitmap.Bitmap
		pt                          image.Point
	)

	for i := 0; i < n; i++ {
		index = c.BaseIndex + i
		if x1, y1, err = c.CentroidPoints.GetGeometry(index); err != nil {
			return errors.Wrap(err, processName, "CentroidPoints")
		}
		if iClass, err = c.ClassIDs.Get(index); err != nil {
			return errors.Wrap(err, processName, "ClassIDs.Get")
		}
		if x2, y2, err = c.CentroidPointsTemplates.GetGeometry(iClass); err != nil {
			return errors.Wrap(err, processName, "CentroidPointsTemplates")
		}
		delX := x2 - x1
		delY := y2 - y1
		if delX >= 0 {
			idelX = int(delX + 0.5)
		} else {
			idelX = int(delX - 0.5)
		}
		if delY >= 0 {
			idelY = int(delY + 0.5)
		} else {
			idelY = int(delY - 0.5)
		}
		if box, err = boxa.Get(i); err != nil {
			return errors.Wrap(err, processName, "")
		}
		x, y := box.Min.X, box.Min.Y

		// finalPositionForAligment()
		t, err = c.UndilatedTemplates.GetBitmap(iClass)
		if err != nil {
			return errors.Wrap(err, processName, "UndilatedTemplates.Get(iClass)")
		}

		pt, err = finalAlignmentPositioning(s, x, y, idelX, idelY, t)
		if err != nil {
			return errors.Wrap(err, processName, "")
		}

		c.PtaUL.AddPoint(float32(x-idelX+pt.X), float32(y-idelY+pt.Y))
	}
	return nil
}

func (c *Classer) verifyMethod(method Method) error {
	if method != RankHaus && method != Correlation {
		return errors.Error("verifyMethod", "invalid classer method")
	}
	return nil
}
