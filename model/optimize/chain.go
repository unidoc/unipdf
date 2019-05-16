/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package optimize

import (
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

// Chain allows to use sequence of optimizers.
// It implements interface model.Optimizer.
type Chain struct {
	optimizers []model.Optimizer
}

// Append appends optimizers to the chain.
func (c *Chain) Append(optimizers ...model.Optimizer) {
	c.optimizers = append(c.optimizers, optimizers...)
}

// Optimize optimizes PDF objects to decrease PDF size.
func (c *Chain) Optimize(objects []core.PdfObject) (optimizedObjects []core.PdfObject, err error) {
	optimizedObjects = objects
	for _, optimizer := range c.optimizers {
		optimizedObjects, err = optimizer.Optimize(optimizedObjects)
		if err != nil {
			return optimizedObjects, err
		}
	}
	return optimizedObjects, nil
}
