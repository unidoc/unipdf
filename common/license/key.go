/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package license

import (
	"fmt"
	"time"

	"github.com/unidoc/unipdf/v3/common"
)

const (
	LicenseTierUnlicensed = "unlicensed"
	LicenseTierCommunity  = "community"
	LicenseTierIndividual = "individual"
	LicenseTierBusiness   = "business"
)

// Make sure all time is at least after this for sanity check.
var testTime = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)

// Old licenses had expiry that were not meant to expire. Only checking expiry
// on licenses issued later than this date.
var startCheckingExpiry = time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC)

type LicenseKey struct {
	LicenseId    string     `json:"license_id"`
	CustomerId   string     `json:"customer_id"`
	CustomerName string     `json:"customer_name"`
	Tier         string     `json:"tier"`
	CreatedAt    time.Time  `json:"-"`
	CreatedAtInt int64      `json:"created_at"`
	ExpiresAt    *time.Time `json:"-"`
	ExpiresAtInt int64      `json:"expires_at"`
	CreatedBy    string     `json:"created_by"`
	CreatorName  string     `json:"creator_name"`
	CreatorEmail string     `json:"creator_email"`
}

func (k *LicenseKey) isExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}

	if k.CreatedAt.Before(startCheckingExpiry) {
		return false
	}

	utcNow := time.Now().UTC()
	return utcNow.After(*k.ExpiresAt)
}

func (k *LicenseKey) Validate() error {
	if len(k.LicenseId) < 10 {
		return fmt.Errorf("invalid license: License Id")
	}

	if len(k.CustomerId) < 10 {
		return fmt.Errorf("invalid license: Customer Id")
	}

	if len(k.CustomerName) < 1 {
		return fmt.Errorf("invalid license: Customer Name")
	}

	if testTime.After(k.CreatedAt) {
		return fmt.Errorf("invalid license: Created At is invalid")
	}

	if k.ExpiresAt != nil {
		if k.CreatedAt.After(*k.ExpiresAt) {
			return fmt.Errorf("invalid license: Created At cannot be Greater than Expires At")
		}
	}

	if k.isExpired() {
		return fmt.Errorf("invalid license: The license has already expired")
	}

	if len(k.CreatorName) < 1 {
		return fmt.Errorf("invalid license: Creator name")
	}

	if len(k.CreatorEmail) < 1 {
		return fmt.Errorf("invalid license: Creator email")
	}

	return nil
}

func (k *LicenseKey) TypeToString() string {
	if k.Tier == LicenseTierUnlicensed {
		return "Unlicensed"
	}

	if k.Tier == LicenseTierCommunity {
		return "AGPLv3 Open Source Community License"
	}

	if k.Tier == LicenseTierIndividual || k.Tier == "indie" {
		return "Commercial License - Individual"
	}

	return "Commercial License - Business"
}

func (k *LicenseKey) ToString() string {
	str := fmt.Sprintf("License Id: %s\n", k.LicenseId)
	str += fmt.Sprintf("Customer Id: %s\n", k.CustomerId)
	str += fmt.Sprintf("Customer Name: %s\n", k.CustomerName)
	str += fmt.Sprintf("Tier: %s\n", k.Tier)
	str += fmt.Sprintf("Created At: %s\n", common.UtcTimeFormat(k.CreatedAt))

	if k.ExpiresAt == nil {
		str += fmt.Sprintf("Expires At: Never\n")
	} else {
		str += fmt.Sprintf("Expires At: %s\n", common.UtcTimeFormat(*k.ExpiresAt))
	}

	str += fmt.Sprintf("Creator: %s <%s>\n", k.CreatorName, k.CreatorEmail)
	return str
}

func (k *LicenseKey) IsLicensed() bool {
	return k.Tier != LicenseTierUnlicensed
}

func MakeUnlicensedKey() *LicenseKey {
	lk := LicenseKey{}
	lk.CustomerName = "Unlicensed"
	lk.Tier = LicenseTierUnlicensed
	lk.CreatedAt = time.Now().UTC()
	lk.CreatedAtInt = lk.CreatedAt.Unix()
	return &lk
}
