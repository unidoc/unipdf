/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package license helps manage commercial licenses and check if they are valid for the version of unidoc used.
package license

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
)

// Defaults to unlicensed.
var licenseKey = MakeUnlicensedKey()

// SetLicenseKey sets and validates the license key.
func SetLicenseKey(content string, customerName string) error {
	lk, err := licenseKeyDecode(content)
	if err != nil {
		return err
	}

	if strings.ToLower(lk.CustomerName) != strings.ToLower(customerName) {
		return fmt.Errorf("customer name mismatch, expected '%s', but got '%s'", customerName, lk.CustomerName)
	}

	err = lk.Validate()
	if err != nil {
		return err
	}

	licenseKey = &lk

	return nil
}

const licensePathEnvironmentVar = `UNIPDF_LICENSE_PATH`
const licenseCustomerNameEnvironmentVar = `UNIPDF_CUSTOMER_NAME`

func init() {
	lpath := os.Getenv(licensePathEnvironmentVar)
	custName := os.Getenv(licenseCustomerNameEnvironmentVar)

	if len(lpath) == 0 || len(custName) == 0 {
		return
	}

	data, err := ioutil.ReadFile(lpath)
	if err != nil {
		common.Log.Debug("Unable to read license file: %v", err)
		return
	}
	err = SetLicenseKey(string(data), custName)
	if err != nil {
		common.Log.Debug("Unable to load license: %v", err)
		return
	}
}

func GetLicenseKey() *LicenseKey {
	if licenseKey == nil {
		return nil
	}

	// Copy.
	lk2 := *licenseKey
	return &lk2
}
