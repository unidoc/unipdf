/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"bytes"
	"errors"

	"github.com/unidoc/unitype"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/internal/textencoding"
	"github.com/unidoc/unipdf/v3/model"
)

// CleanFonts cleans up embedded fonts, reducing font sizes.
type CleanFonts struct {
	// Subset embedded fonts if encountered (if true).
	// Otherwise attempts to reduce the font program.
	Subset bool
}

func optimizeFontsWithSubsetting(objects []core.PdfObject) (processed map[*core.PdfObjectStream]struct{}, err error) {
	// 1. Identify all fonts.
	// 2. Identify content streams and their Resources dictionaries (both via page, forms and annotations).
	// 3. Process content streams.
	processed = map[*core.PdfObjectStream]struct{}{}

	fontMap := map[*model.PdfFont]struct{}{}

	objstr := getObjectStructure(objects)
	for _, p := range objstr.pages {
		pdict, ok := core.GetDict(p.PdfObject)
		if !ok {
			continue
		}
		resourcesDict, ok := core.GetDict(pdict.Get("Resources"))
		if !ok {
			continue
		}
		contents, _ := getPageContents(pdict.Get("Contents"))
		resources, err := model.NewPdfPageResourcesFromDict(resourcesDict)
		if err != nil {
			return nil, err
		}

		allContents := []content{
			{
				content:   contents,
				resources: resources,
			},
		}

		annotContents := getAnnotationContents(pdict.Get("Annots"))
		if annotContents != nil {
			allContents = append(allContents, annotContents...)
		}

		for _, cont := range allContents {
			e, err := extractor.NewFromContents(cont.content, cont.resources)
			if err != nil {
				return nil, err
			}

			pt, _, _, err := e.ExtractPageText()
			if err != nil {
				return nil, err
			}

			for _, el := range pt.Marks().Elements() {
				if el.Font == nil {
					continue
				}
				if _, has := fontMap[el.Font]; !has {
					fontMap[el.Font] = struct{}{}
				}
			}
		}
	}

	// Map of font program stream to font. Multiple fonts can use the same font program.
	fontFileMap := map[*core.PdfObjectStream][]*model.PdfFont{}
	for font := range fontMap {
		fontDesc := font.FontDescriptor()
		if fontDesc == nil || fontDesc.FontFile2 == nil {
			continue
		}
		stream, ok := core.GetStream(fontDesc.FontFile2)
		if !ok {
			continue
		}
		fontFileMap[stream] = append(fontFileMap[stream], font)
	}

	for stream := range fontFileMap {
		var allRunes []rune
		var allIndices []unitype.GlyphIndex

		for _, font := range fontFileMap[stream] {
			switch t := font.Encoder().(type) {
			case *textencoding.IdentityEncoder:
				// TODO: This terminology is wrong as those are not runes, just charcodes cast as runes.
				//   Identity encoder maps via 2-byte encoding directly from 2byte charcode to glyph index.
				runes := t.RegisteredRunes()
				indices := make([]unitype.GlyphIndex, len(runes))
				for i, r := range runes {
					indices[i] = unitype.GlyphIndex(r)
				}
				allIndices = append(allIndices, indices...)
			case *textencoding.TrueTypeFontEncoder:
				runes := t.RegisteredRunes()
				allRunes = append(allRunes, runes...)
			case textencoding.SimpleEncoder:
				charcodes := t.Charcodes()
				for _, c := range charcodes {
					r, ok := t.CharcodeToRune(c)
					if !ok {
						common.Log.Debug("Charcode<->rune not found: %d", c)
						continue
					}
					allRunes = append(allRunes, r)
				}
			}
		}

		err = subsetFontStream(stream, allRunes, allIndices)
		if err != nil {
			common.Log.Debug("ERROR subsetting font stream: %v", err)
			return nil, err
		}
		processed[stream] = struct{}{}
	}
	return processed, nil
}

// Subsets the font program in `stream` with the subset based on the `runes` and glyph `indices`.
func subsetFontStream(stream *core.PdfObjectStream, runes []rune, indices []unitype.GlyphIndex) error {
	stream, ok := core.GetStream(stream)
	if !ok {
		common.Log.Debug("Embedded font object not found -- ABORT subsetting")
		return errors.New("fontfile2 not found")
	}
	decoded, err := core.DecodeStream(stream)
	if err != nil {
		common.Log.Debug("Decode error: %v", err)
		return err
	}

	fnt, err := unitype.Parse(bytes.NewReader(decoded))
	if err != nil {
		common.Log.Debug("Error parsing %d byte font", len(stream.Stream))
		return err
	}

	allIndices := indices
	if len(runes) > 0 {
		indices := fnt.LookupRunes(runes)
		allIndices = append(allIndices, indices...)
	}

	fnt, err = fnt.SubsetKeepIndices(allIndices)
	if err != nil {
		common.Log.Debug("ERROR subsetting font: %v", err)
		return err
	}

	var buf bytes.Buffer
	err = fnt.Write(&buf)
	if err != nil {
		common.Log.Debug("ERROR Writing font: %v", err)
		return err
	}
	if buf.Len() > len(decoded) {
		common.Log.Debug("Re-written font is larger than original - skip")
		return nil
	}

	newstream, err := core.MakeStream(buf.Bytes(), core.NewFlateEncoder())
	if err != nil {
		common.Log.Debug("ERROR Writing font: %v", err)
		return err
	}
	// Overwrite.
	*stream = *newstream
	stream.Set("Length1", core.MakeInteger(int64(buf.Len())))

	return nil
}

// Optimize optimizes PDF objects to decrease PDF size.
func (c *CleanFonts) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	var processed map[*core.PdfObjectStream]struct{}
	if c.Subset {
		var err error
		processed, err = optimizeFontsWithSubsetting(objects)
		if err != nil {
			return nil, err
		}
	}

	// Clean font streams by loading and rewriting with minimal needed tables.
	for _, obj := range objects {
		stream, isStreamObj := core.GetStream(obj)
		if !isStreamObj {
			continue
		}
		if _, has := processed[stream]; has {
			// Skip - has been processed.
			continue
		}

		encoder, err := core.NewEncoderFromStream(stream)
		if err != nil {
			common.Log.Debug("ERROR getting encoder: %v - ignoring", err)
			continue
		}

		decoded, err := encoder.DecodeStream(stream)
		if err != nil {
			common.Log.Debug("Decoding error : %v - ignoring", err)
			continue
		}
		if len(decoded) < 4 {
			continue
		}

		version := string(decoded[:4])
		if version == "OTTO" {
			// Fonts based on PostScript outlines not supported yet.
			// See https://docs.microsoft.com/en-us/typography/opentype/spec/otff
			continue
		}
		if version != "\x00\x01\x00\x00" && version != "true" {
			continue
		}

		fnt, err := unitype.Parse(bytes.NewReader(decoded))
		if err != nil {
			common.Log.Debug("ERROR Parsing font: %v - ignoring", err)
			continue
		}
		err = fnt.Optimize()
		if err != nil {
			continue
		}

		var buf bytes.Buffer
		err = fnt.Write(&buf)
		if err != nil {
			common.Log.Debug("ERROR Writing font: %v - ignoring", err)
			continue
		}
		if buf.Len() > len(decoded) {
			common.Log.Debug("Re-written font is larger than original - skip")
			continue
		}

		newstream, err := core.MakeStream(buf.Bytes(), core.NewFlateEncoder())
		if err != nil {
			continue
		}
		// Overwrite.
		*stream = *newstream
		stream.Set("Length1", core.MakeInteger(int64(buf.Len())))
	}
	return objects, nil
}

// content describes page or font contents which is a content stream along with resources.
type content struct {
	content   string
	resources *model.PdfPageResources
}

// Best effort to get annotation contents.
func getAnnotationContents(annotsObj core.PdfObject) []content {
	if annotsObj == nil {
		return nil
	}
	annotsArr, ok := core.GetArray(annotsObj)
	if !ok {
		common.Log.Debug("Annots not an array")
		return nil
	}

	var annotContents []content
	for _, obj := range annotsArr.Elements() {
		annotDict, ok := core.GetDict(obj)
		if !ok {
			// Ignore any non dict elements.
			common.Log.Debug("Ignoring non-dict element in Annots")
			continue
		}

		// Appearance.
		appDict, ok := core.GetDict(annotDict.Get("AP"))
		if !ok {
			common.Log.Debug("No AP entry - skipping")
			continue
		}

		normal := core.TraceToDirectObject(appDict.Get("N"))
		if normal == nil {
			common.Log.Debug("No N entry - skipping")
			continue
		}

		var stream *core.PdfObjectStream
		switch t := normal.(type) {
		case *core.PdfObjectDictionary:
			appState, ok := core.GetName(annotDict.Get("AS"))
			if !ok {
				common.Log.Debug("No AS entry - skipping")
				continue
			}
			stream, ok = core.GetStream(t.Get(*appState))
			if !ok {
				common.Log.Debug("Form not found - skipping")
				continue
			}
		case *core.PdfObjectStream:
			stream = t
		}
		if stream == nil {
			common.Log.Debug("Form not found (nil) - skipping")
			continue
		}

		xform, err := model.NewXObjectFormFromStream(stream)
		if err != nil {
			common.Log.Debug("Error loading form: %v - ignoring", err)
			continue
		}

		contents, err := xform.GetContentStream()
		if err != nil {
			common.Log.Debug("Error decoding contents: %v", err)
			continue
		}

		annotContents = append(annotContents, content{
			content:   string(contents),
			resources: xform.Resources,
		})
	}

	return annotContents
}
