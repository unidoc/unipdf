/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"crypto/md5"

	"github.com/unidoc/unipdf/v3/core"
)

// CombineDuplicateStreams combines duplicated streams by its data hash.
// It implements interface model.Optimizer.
type CombineDuplicateStreams struct {
}

// Optimize optimizes PDF objects to decrease PDF size.
func (dup *CombineDuplicateStreams) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	replaceTable := make(map[core.PdfObject]core.PdfObject)
	toDelete := make(map[core.PdfObject]struct{})
	streamsByHash := make(map[string][]*core.PdfObjectStream)
	for _, obj := range objects {
		if stream, isStreamObj := obj.(*core.PdfObjectStream); isStreamObj {
			hasher := md5.New()
			hasher.Write([]byte(stream.Stream))
			hash := string(hasher.Sum(nil))
			streamsByHash[hash] = append(streamsByHash[hash], stream)
		}
	}
	for _, streams := range streamsByHash {
		if len(streams) < 2 {
			continue
		}
		firstStream := streams[0]
		for i := 1; i < len(streams); i++ {
			stream := streams[i]
			replaceTable[stream] = firstStream
			toDelete[stream] = struct{}{}
		}
	}

	optimizedObjects = make([]core.PdfObject, 0, len(objects)-len(toDelete))
	for _, obj := range objects {
		if _, found := toDelete[obj]; found {
			continue
		}
		optimizedObjects = append(optimizedObjects, obj)
	}
	replaceObjectsInPlace(optimizedObjects, replaceTable)
	return optimizedObjects, nil
}
