/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
)

// CleanContentstream cleans up redundant operands in content streams, including Page and XObject Form
// contents. This process includes:
// 1. Marked content operators are removed.
// 2. Some operands are simplified (shorter form).
// TODO: Add more reduction methods and improving the methods for identifying unnecessary operands.
type CleanContentstream struct {
}

// filterOps cleans up the content stream in `ops`:
// 1. Marked content operators are cleaned.
// 2. Tm with 1 0 0 1 params are converted to Td (slightly shorter for same transformation).
// TODO: Add operations that track the state and remove unnecessary operands, such as duplicates
//  or ones setting default values, or ones not drawing anything.
func filterOps(ops *contentstream.ContentStreamOperations) *contentstream.ContentStreamOperations {
	if ops == nil {
		return nil
	}

	filtered := contentstream.ContentStreamOperations{}
	for _, op := range *ops {
		switch op.Operand {
		case "BDC", "BMC", "EMC":
			continue
		case "Tm":
			if len(op.Params) == 6 {
				if nums, err := core.GetNumbersAsFloat(op.Params); err == nil {
					if nums[0] == 1 && nums[1] == 0 && nums[2] == 0 && nums[3] == 1 {
						op = &contentstream.ContentStreamOperation{
							Params: []core.PdfObject{
								op.Params[4],
								op.Params[5],
							},
							Operand: "Td",
						}
					}
				}
			}
		}
		filtered = append(filtered, op)
	}
	return &filtered
}

// reduceContent performs content stream optimization of contents in `cstream` which can either be
// from Page Contents or XObject Form.
// NOTE: If from a Contents array, the operations may be unbalanced.
func reduceContent(cstream *core.PdfObjectStream) error {
	decoded, err := core.DecodeStream(cstream)
	if err != nil {
		return err
	}

	csp := contentstream.NewContentStreamParser(string(decoded))
	ops, err := csp.Parse()
	if err != nil {
		return err
	}

	ops = filterOps(ops)
	cleaned := ops.Bytes()
	if len(cleaned) >= len(decoded) {
		// No need to replace if no improvement.
		return nil
	}

	newstream, err := core.MakeStream(ops.Bytes(), core.NewFlateEncoder())
	if err != nil {
		return err
	}
	cstream.Stream = newstream.Stream
	cstream.Merge(newstream.PdfObjectDictionary)
	return nil
}

// Optimize optimizes PDF objects to decrease PDF size.
func (c *CleanContentstream) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	// Track which content streams to process.
	queuedMap := map[*core.PdfObjectStream]struct{}{}
	var queued []*core.PdfObjectStream
	appendQueue := func(stream *core.PdfObjectStream) {
		if _, has := queuedMap[stream]; !has {
			queuedMap[stream] = struct{}{}
			queued = append(queued, stream)
		}
	}

	// Collect objects to process: XObject Form and Page Content streams.
	for _, obj := range objects {
		switch t := obj.(type) {
		case *core.PdfIndirectObject:
			switch ti := t.PdfObject.(type) {
			case *core.PdfObjectDictionary:
				if name, ok := core.GetName(ti.Get("Type")); !ok || name.String() != "Page" {
					continue
				}

				if stream, ok := core.GetStream(ti.Get("Contents")); ok {
					appendQueue(stream)
				} else if array, ok := core.GetArray(ti.Get("Contents")); ok {
					for _, el := range array.Elements() {
						if stream, ok := core.GetStream(el); ok {
							appendQueue(stream)
						}
					}
				}
			}
		case *core.PdfObjectStream:
			if name, ok := core.GetName(t.Get("Type")); !ok || name.String() != "XObject" {
				continue
			}
			if name, ok := core.GetName(t.Get("Subtype")); !ok || name.String() != "Form" {
				continue
			}
			appendQueue(t)
		}
	}

	// Process the queued content streams.
	for _, stream := range queued {
		err = reduceContent(stream)
		if err != nil {
			return nil, err
		}
	}

	return objects, nil
}
