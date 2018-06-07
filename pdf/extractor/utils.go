/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common/license"
	"github.com/unidoc/unidoc/pdf/core"
)

// getNumberAsFloat can retrieve numeric values from PdfObject (both integer/float).
func getNumberAsFloat(obj core.PdfObject) (float64, error) {
	if fObj, ok := obj.(*core.PdfObjectFloat); ok {
		return float64(*fObj), nil
	}

	if iObj, ok := obj.(*core.PdfObjectInteger); ok {
		return float64(*iObj), nil
	}

	return 0, errors.New("Not a number")
}

func procBuf(buf *bytes.Buffer) {
	if isTesting {
		return
	}

	lk := license.GetLicenseKey()
	if lk != nil && lk.IsLicensed() {
		return
	}
	fmt.Printf("Unlicensed copy of unidoc\n")
	fmt.Printf("To get rid of the watermark and keep entire text - Please get a license on https://unidoc.io\n")

	s := "- [Unlicensed UniDoc - Get a license on https://unidoc.io]"
	if buf.Len() > 100 {
		s = "... [Truncated - Unlicensed UniDoc - Get a license on https://unidoc.io]"
		buf.Truncate(buf.Len() - 100)
	}
	buf.WriteString(s)
}
