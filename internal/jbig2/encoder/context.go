package encoder

import (
	"github.com/unidoc/unipdf/internal/jbig2/bitmap"
	"github.com/unidoc/unipdf/v3/internal/jbig2/encoder/classer"
)

// Document is the context for multi-page JBIG2 document.
type Document struct {
	// Classer is the document classifier.
	Classer classer.Classer
	// XRes and YRes are the PPI for the x and y direction.
	XRes, YRes int
	// FullHeaders is a flag that defines if the encoder should produce full JBIG2 files.
	FullHeaders bool
	// PDFPageNumbering is a flag that defines if all text pages are in PDF mode - single page no #1.
	PDFPageNumbering bool

	Pages         map[int]*Page
	GlobalSymbols map[int]int
}

// New creates new JBIG2 encoding context Document.
func New(thresh, weightFactor float32, xRes, yRes int, fullHeaders bool, refineLevel int) *Document {
	// TODO: jbig2enc.cc:122
	return &Document{}
}

// UniteTemplates ...
func (d *Document) UniteTemplates(newRepresentant int, templatesToUnited []int) error {
	// TODO: jbig2enc.cc:167
	return nil
}

func (d *Document) addPage(input *bitmap.Pix) {
	// TODO: jbig2enc.cc:491
}

func (d *Document) autoThreshold() {
	// TODO: jbig2enc.cc:349
}

func (d *Document) autoThresholdUsingHash() {
	// TODO: jbig2enc.cc:420
}

func (d *Document) encodeGeneric() {
	// TODO: jbig2enc.cc:890
}

func (d *Document) producePage(pageNo, xRes, yRes int, length int) {
	e // TODO: jbig2enc.cc:717

}

func (d *Document) removeTemplates(templatesToRemove []int) error {
	// TODO: jbig2enc.cc:216
	return nil
}

func (d *Document) uniteTemplatesWithIndexes(firstTemplateIndex, secondTemplateIndex int) error {
	// TODO: jbig2enc.cc:288
	return nil
}

// CountHash ...
func CountHash(p *bitmap.Pix, m map[uint][]int, templateIndex int) error {
	// TODO: jbig2enc.cc:388
	return nil
}
