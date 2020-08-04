/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unidoc/unipdf/v3/core"
)

func TestBlockWithoutOpacity(t *testing.T) {
	block := Block{}
	gsName, err := block.setOpacity(1.0, 1.0)
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.Equal(t, "", gsName)
}

func TestBlockWithOpacity(t *testing.T) {
	block := NewBlock(100, 100)
	gsName, err := block.setOpacity(0.1, 0.2)
	if err != nil {
		t.Errorf("Fail: %v", err)
		return
	}
	assert.Equal(t, "GS0", gsName)

	extGState, ok := core.TraceToDirectObject(block.resources.ExtGState).(*core.PdfObjectDictionary)
	if !ok {
		t.Errorf("Failed to convert ExtGState to dictionary")
		return
	}

	gsDictionary, ok := core.TraceToDirectObject(extGState.Get(core.PdfObjectName(gsName))).(*core.PdfObjectDictionary)
	if !ok {
		t.Errorf("Failed to convert ExtGState to dictionary")
		return
	}
	assert.Equal(t, "0.1", gsDictionary.Get("ca").WriteString())
	assert.Equal(t, "0.2", gsDictionary.Get("CA").WriteString())
}
