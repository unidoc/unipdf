/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"bytes"
	"fmt"

	"github.com/unidoc/unitype"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/core"
)

// CleanFonts cleans up embedded fonts, reducing font sizes.
// TODO: Add subsetting option to reduce the number of glyphs in the fonts.
type CleanFonts struct {
	// Subset embedded fonts if encountered (if true).
	Subset bool
}

// Optimize optimizes PDF objects to decrease PDF size.
func (c *CleanFonts) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	// 1. Identify all fonts.
	// 2. Identify content streams and their Resources dictionaries (both via page, forms and annotations).
	// 3. Process content streams.
	/*
		if c.Subset {
			objstr := getObjectStructure(objects)
			fontUse, err := getFontUsage(objstr.pages)
			if err != nil {
				return nil, err
			}
		}
	*/
	/*
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
			fmt.Printf("%s\n", pdict.String())
			resources, err := model.NewPdfPageResourcesFromDict(resourcesDict)
			if err != nil {
				return nil, err
			}

			// TODO: should not use extractor, need more control here, and extractor has overhead, as focus is on getting.
			//  text.
			fmt.Printf("%s\n", contents)
			e, err := extractor.NewFromContents(contents, resources)
			if err != nil {
				return nil, err
			}

			// TODO: This process does not touch the annotations unfortunately.  Looks like we need our own processor
			//  here, also for this we don't need to look at the ToUnicode maps.
			//  It should be an option to keep ToUnicode maps in order.
			fontMap := map[*model.PdfFont]struct{}{}
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
			for font := range fontMap {
				fmt.Printf("Subsetting font\n")
				err = font.SubsetRegistered()
				if err != nil {
					return nil, err
				}
			}
		}
	*/

	for _, obj := range objects {
		stream, isStreamObj := core.GetStream(obj)
		if !isStreamObj {
			continue
		}

		encoder, err := core.NewEncoderFromStream(stream)
		if err != nil {
			fmt.Printf("ERROR getitng encoder: %v - ignoring\n", err)
			continue
		}

		decoded, err := encoder.DecodeStream(stream)
		if err != nil {
			fmt.Printf("Deocding error : %v - ignoring\n", err)
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
			fmt.Printf("ERROR Parsing font: %v - ignoring\n", err)
			continue
		}
		// TODO: Why is pruning cmap sometimes OK?  Would expect that viewer maps straight.
		//  Case different in Adobe.
		err = fnt.PruneTables("name", "post")
		//err = fnt.PruneTables("cmap")
		if err != nil {
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

type fontUsage struct {
	font core.PdfObject
	gids []int64 // or runes?
}

// getFontUsage probes the entire document content streams and identifies what fonts are used a
/*
func getFontUsage(pages []*core.PdfObjectDictionary) ([]fontUsage, error) {
	for _, p := range pages {

	}
}

*/
