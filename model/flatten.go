/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// ContentStreamWrapper wraps the Page's contentstream into q ... Q blocks.
type ContentStreamWrapper interface {
	WrapContentStream(page *PdfPage) error
}

// FieldAppearanceGenerator generates appearance stream for a given field.
type FieldAppearanceGenerator interface {
	ContentStreamWrapper
	GenerateAppearanceDict(form *PdfAcroForm, field *PdfField, wa *PdfAnnotationWidget) (*core.PdfObjectDictionary, error)
}

// FlattenFields flattens the form fields and annotations for the PDF loaded in `pdf` and makes
// non-editable.
// Looks up all widget annotations corresponding to form fields and flattens them by drawing the content
// through the content stream rather than annotations.
// References to flattened annotations will be removed from Page Annots array. For fields the AcroForm entry
// will be emptied.
// When `allannots` is true, all annotations will be flattened. Keep false if want to keep non-form related
// annotations intact.
// When `appgen` is not nil, it will be used to generate appearance streams for the field annotations.
func (r *PdfReader) FlattenFields(allannots bool, appgen FieldAppearanceGenerator) error {
	// Load all target widget annotations to be flattened into a map.
	// The bool value indicates whether the annotation has value content.
	ftargets := map[*PdfAnnotation]bool{}
	{
		var fields []*PdfField
		acroForm := r.AcroForm
		if acroForm != nil {
			fields = acroForm.AllFields()
		}

		for _, field := range fields {
			for _, wa := range field.Annotations {
				// TODO(gunnsth): Check if wa.Flags() has Print flag then include, otherwise exclude.

				// NOTE(gunnsth): May be better to check field.V only if no appearance stream available.
				ftargets[wa.PdfAnnotation] = field.V != nil

				if appgen != nil {
					// appgen generates the appearance based on the form/field/annotation and other settings
					// based on the implementation (for example may only generate appearance if none set).
					apDict, err := appgen.GenerateAppearanceDict(acroForm, field, wa)
					if err != nil {
						return err
					}
					wa.AP = apDict
				}
			}
		}
	}

	// If all annotations are to be flattened, add to targets.
	if allannots {
		for _, page := range r.PageList {
			annotations, err := page.GetAnnotations()
			if err != nil {
				return err
			}

			for _, annot := range annotations {
				ftargets[annot] = true
			}
		}
	}

	// Go through all pages and flatten specified annotations.
	for _, page := range r.PageList {
		var annots []*PdfAnnotation

		// Wrap the content streams.
		if appgen != nil {
			if err := appgen.WrapContentStream(page); err != nil {
				return err
			}
		}

		annotations, err := page.GetAnnotations()
		if err != nil {
			return err
		}

		for _, annot := range annotations {
			hasV, toflatten := ftargets[annot]
			if !toflatten {
				// Not to be flattened.
				annots = append(annots, annot)
				continue
			}

			// Flatten annotation.
			// Annotations not requiring an appearance dictionary.
			switch annot.GetContext().(type) {
			case *PdfAnnotationPopup:
				continue
			case *PdfAnnotationLink:
				continue
			case *PdfAnnotationProjection:
				continue
			}

			xform, rect, err := getAnnotationActiveAppearance(annot)
			if err != nil {
				if !hasV {
					common.Log.Trace("Field without V -> annotation without appearance stream - skipping over")
					continue
				}
				common.Log.Debug("ERROR Annotation without appearance stream, err : %v - skipping over", err)
				continue
			}
			if xform == nil {
				// No appearance.
				continue
			}

			// Add the XForm to Page resources and draw it in the contentstream.
			name := page.Resources.GenerateXObjectName()
			page.Resources.SetXObjectFormByName(name, xform)

			// TODO(gunnsth): Take Matrix and potential scaling of annotation Rect and appearance
			// BBox into account. Have yet to find a case where that actually is required.

			// Placement for XForm.
			xRect := math.Min(rect.Llx, rect.Urx)
			yRect := math.Min(rect.Lly, rect.Ury) // Needed for rect in: govdocs 019693.pdf.

			// Generate the content stream to display the XForm.
			// TODO(gunnsth): Creating the contentstream directly here as cannot import contentstream package into
			// model (as contentstream depends on model). Consider if we can change the dependency pattern.
			var ops []string
			ops = append(ops, "q")
			ops = append(ops, fmt.Sprintf("%.6f %.6f %.6f %.6f %.6f %.6f cm", 1.0, 0.0, 0.0, 1.0, xRect, yRect))
			ops = append(ops, fmt.Sprintf("/%s Do", name.String()))
			ops = append(ops, "Q")
			contentstr := strings.Join(ops, "\n")

			err = page.AppendContentStream(contentstr)
			if err != nil {
				return err
			}

			// TODO: Add clever function to merge Resources, renaming and modifying contentstream if conflicts.
			// Could be based on similar functionality already available in creator, perhaps refactored to an
			// internal utility package, so can be accessed widely.
			if xform.Resources != nil {
				xfontDict, has := core.GetDict(xform.Resources.Font)
				if has {
					for _, fname := range xfontDict.Keys() {
						// Only set if no matching font in page resources.
						if !page.Resources.HasFontByName(fname) {
							page.Resources.SetFontByName(fname, xfontDict.Get(fname))
						}
					}
				}
			}
		}

		// Remove reference to flattened annotations.
		if len(annots) > 0 {
			page.annotations = annots
		} else {
			page.annotations = []*PdfAnnotation{}
		}
	}

	r.AcroForm = nil

	return nil
}

// getAnnotationActiveAppearance retrieves the active XObject Form for an appearance dictionary.
// Default gets the N entry, and if it is a dictionary, picks the entry referred to by AS.
// If returned XObject Form is nil (and no errors) it indicates that the annotation has no appearance.
func getAnnotationActiveAppearance(annot *PdfAnnotation) (*XObjectForm, *PdfRectangle, error) {
	// For debugging:
	//common.Log.Trace("----")
	//common.Log.Trace("annot: %#v", annot)
	//common.Log.Trace("context: %#v", annot.GetContext())
	//common.Log.Trace("obj: %v", annot.GetContainingPdfObject())

	// Appearance dictionary entries (Table 168 p. 397).
	apDict, has := core.GetDict(annot.AP)
	if !has {
		return nil, nil, errors.New("field missing AP dictionary")
	}
	if apDict == nil {
		return nil, nil, nil
	}

	// Get the Rect specifying the display rectangle.
	rectArr, has := core.GetArray(annot.Rect)
	if !has || rectArr.Len() != 4 {
		return nil, nil, errors.New("rect invalid")
	}
	rect, err := NewPdfRectangle(*rectArr)
	if err != nil {
		return nil, nil, err
	}

	nobj := core.TraceToDirectObject(apDict.Get("N"))
	switch t := nobj.(type) {
	case *core.PdfObjectStream:
		stream := t
		xform, err := NewXObjectFormFromStream(stream)
		return xform, rect, err
	case *core.PdfObjectDictionary:
		// An annotation representing multiple fields may have many appearances.
		// As an example checkbox may have two appearance states On and Off.
		// Its appearance dictionary would contain /N << /On Ref /Off Ref >>, the choice is
		// determines by the AS entry in the annotation dictionary.
		nDict := t

		state, has := core.GetName(annot.AS)
		if !has {
			// No appearance (nil).
			return nil, nil, nil
		}

		if nDict.Get(*state) == nil {
			common.Log.Debug("ERROR: AS state not specified in AP dict - ignoring")
			return nil, nil, nil
		}

		stream, has := core.GetStream(nDict.Get(*state))
		if !has {
			common.Log.Debug("ERROR: Unable to access appearance stream for %v", state)
			return nil, nil, errors.New("stream missing")
		}
		xform, err := NewXObjectFormFromStream(stream)
		return xform, rect, err
	}

	common.Log.Debug("Invalid type for N: %T", nobj)
	return nil, nil, errors.New("type check error")
}
