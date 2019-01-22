package e2etest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/unidoc/unidoc/common"
)

var (
	ghostscriptBinPath = os.Getenv("UNIDOC_GS_BIN_PATH")
)

// validatePdf a pdf file using Ghostscript, returns an error if unable to execute.
// Also returns the number of output warnings, which can be used as some sort of measure
// of validity, especially when comparing with a transformed version of same file.
func validatePdf(path string, password string) (error, int) {
	if len(ghostscriptBinPath) == 0 {
		return errors.New("UNIDOC_GS_BIN_PATH not set"), 0
	}
	common.Log.Debug("Validating: %s", path)

	var cmd *exec.Cmd
	if len(password) > 0 {
		option := fmt.Sprintf("-sPDFPassword=%s", password)
		cmd = exec.Command(ghostscriptBinPath, "-dBATCH", "-dNODISPLAY", "-dNOPAUSE", option, path)
	} else {
		cmd = exec.Command(ghostscriptBinPath, "-dBATCH", "-dNODISPLAY", "-dNOPAUSE", path)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		common.Log.Debug("%s", out.String())
		common.Log.Debug("%s", errOut.String())
		common.Log.Error("GS failed with error %s", err)
		return fmt.Errorf("GS failed with error (%s)", err), 0
	}

	outputErr := errOut.String()
	warnings := strings.Count(outputErr, "****")
	common.Log.Error(": - %d warnings %s", warnings, outputErr)

	if warnings > 1 {
		if len(outputErr) > 80 {
			outputErr = outputErr[:80] // Trim the output.
		}
		common.Log.Error("Invalid - %d warnings %s", warnings, outputErr)
		//return fmt.Errorf("Invalid - %d warnings (%s)", warnings, outputErr), warnings
		return nil, warnings
	}

	// Valid if no error.
	return nil, 0
}
