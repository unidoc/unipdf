/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package license helps manage commercial licenses and check if they are valid for the version of unidoc used.
package license

import (
	"fmt"
	"strings"
)

// Defaults to the open source license.
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

func GetLicenseKey() *LicenseKey {
	if licenseKey == nil {
		return nil
	}

	// Copy.
	lk2 := *licenseKey
	return &lk2
}
