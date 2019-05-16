/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package e2etest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
)

// To enable ghostscript validation, the path to the binary needs to be specified.
// Set environment variable:
//		UNIDOC_GS_BIN_PATH to the path of the ghostscript binary (gs).
var (
	ghostscriptBinPath = os.Getenv("UNIDOC_GS_BIN_PATH")
)

// validatePdf a pdf file using Ghostscript, returns an error if unable to execute.
// Also returns the number of output warnings, which can be used as some sort of measure
// of validity, especially when comparing with a transformed version of same file.
func validatePdf(path string, password string) (int, error) {
	if len(ghostscriptBinPath) == 0 {
		return 0, errors.New("UNIDOC_GS_BIN_PATH not set")
	}
	common.Log.Debug("Validating: %s", path)

	params := []string{"-dBATCH", "-dNODISPLAY", "-dNOPAUSE"}
	if len(password) > 0 {
		params = append(params, fmt.Sprintf("-sPDFPassword=%s", password))
	}
	params = append(params, path)

	var (
		out    bytes.Buffer
		errOut bytes.Buffer
	)
	cmd := exec.Command(ghostscriptBinPath, params...)
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		common.Log.Debug("%s", out.String())
		common.Log.Debug("%s", errOut.String())
		common.Log.Error("GS failed with error %s", err)
		return 0, fmt.Errorf("GS failed with error (%s)", err)
	}

	outputErr := errOut.String()
	warnings := strings.Count(outputErr, "****")
	common.Log.Debug("ERROR: - %d warnings %s", warnings, outputErr)

	if warnings > 1 {
		if len(outputErr) > 80 {
			outputErr = outputErr[:80] // Trim the output.
		}
		common.Log.Debug("ERROR: Invalid - %d warnings %s", warnings, outputErr)
		return warnings, nil
	}

	// Valid if no error.
	return 0, nil
}
