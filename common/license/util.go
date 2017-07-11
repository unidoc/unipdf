/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// The license package helps manage commercial licenses and check if they are valid for the version of unidoc used.
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
