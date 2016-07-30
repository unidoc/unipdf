/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package license contains customer opensource and commercial license handling.
package license

// Defaults to the open source license.
var licenseKey *LicenseKey = MakeOpensourceLicenseKey()

// Sets and validates the license key.
func SetLicenseKey(content string) error {
	lk, err := licenseKeyDecode(content)
	if err != nil {
		return err
	}

	err = lk.Validate()
	if err != nil {
		return err
	}

	licenseKey = &lk

	return nil
}

func GetLicenseKey() *LicenseKey {
	return licenseKey
}
