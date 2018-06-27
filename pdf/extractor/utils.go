/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package extractor

import (
	"bytes"
	"fmt"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/common/license"
	. "github.com/unidoc/unidoc/pdf/core"
	"github.com/unidoc/unidoc/pdf/model"
)

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

// toFloatList returns `objs` as 2 floats, if that's what it is, or an error if it isn't
func toFloatXY(objs []PdfObject) (x, y float64, err error) {
	if len(objs) != 2 {
		err = fmt.Errorf("Invalid number of params: %d", len(objs))
		common.Log.Debug("toFloatXY: err=%v", err)
		return
	}
	floats, err := model.GetNumbersAsFloat(objs)
	if err != nil {
		return
	}
	x, y = floats[0], floats[1]
	return
}

// truncate returns the first `n` characters in string `s`
func truncate(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}
