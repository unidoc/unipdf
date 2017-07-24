/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package license

import (
	"fmt"
	"strings"
	"time"

	"github.com/unidoc/unidoc/common"
)

const (
	LicenseTypeCommercial = "commercial"
	LicenseTypeOpensource = "opensource"
)

const opensourceLicenseId = "01aa523c-b4c6-4d57-bbdd-5a88d2bd5300"
const opensourceLicenseUuid = "01aa523c-b4c6-4d57-bbdd-5a88d2bd5301"

func getSupportedFeatures() []string {
	return []string{"unidoc", "unidoc-cli"}
}

// Make sure all time is at least after this for sanity check.
var testTime = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)

type LicenseKey struct {
	LicenseId    string    `json:"license_id"`
	CustomerId   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Type         string    `json:"type"`
	Features     []string  `json:"features"`
	CreatedAt    time.Time `json:"-"`
	CreatedAtInt int64     `json:"created_at"`
	ExpiresAt    time.Time `json:"-"`
	ExpiresAtInt int64     `json:"expires_at"`
	CreatedBy    string    `json:"created_by"`
	CreatorName  string    `json:"creator_name"`
	CreatorEmail string    `json:"creator_email"`
}

func (this *LicenseKey) Validate() error {
	if len(this.LicenseId) < 10 {
		return fmt.Errorf("Invalid license: License Id")
	}

	if len(this.CustomerId) < 10 {
		return fmt.Errorf("Invalid license: Customer Id")
	}

	if len(this.CustomerName) < 1 {
		return fmt.Errorf("Invalid license: Customer Name")
	}

	if this.Features == nil || len(this.Features) < 1 {
		return fmt.Errorf("Invalid license: No features")
	}

	for _, feature := range this.Features {
		found := false

		for _, sf := range getSupportedFeatures() {
			if sf == feature {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Invalid license: Unsupported feature %s", feature)
		}
	}

	if testTime.After(this.CreatedAt) {
		return fmt.Errorf("Invalid license: Created At is invalid")
	}

	if this.CreatedAt.After(this.ExpiresAt) {
		return fmt.Errorf("Invalid license: Created At cannot be Greater than Expires At")
	}

	if common.ReleasedAt.After(this.ExpiresAt) {
		return fmt.Errorf("Expired license, expired at: %s", common.UtcTimeFormat(this.ExpiresAt))
	}

	if len(this.CreatorName) < 1 {
		return fmt.Errorf("Invalid license: Creator name")
	}

	if len(this.CreatorEmail) < 1 {
		return fmt.Errorf("Invalid license: Creator email")
	}

	return nil
}

func (this *LicenseKey) TypeToString() string {
	ret := "AGPLv3 Open Source License"

	if this.Type == LicenseTypeCommercial {
		ret = "Commercial License"
	}

	return ret
}

func (this *LicenseKey) ToString() string {
	str := fmt.Sprintf("License Id: %s\n", this.LicenseId)
	str += fmt.Sprintf("Customer Id: %s\n", this.CustomerId)
	str += fmt.Sprintf("Customer Name: %s\n", this.CustomerName)
	str += fmt.Sprintf("Type: %s\n", this.Type)
	str += fmt.Sprintf("Features: %s\n", strings.Join(this.Features, ", "))
	str += fmt.Sprintf("Created At: %s\n", common.UtcTimeFormat(this.CreatedAt))
	str += fmt.Sprintf("Expires At: %s\n", common.UtcTimeFormat(this.ExpiresAt))
	str += fmt.Sprintf("Creator: %s <%s>\n", this.CreatorName, this.CreatorEmail)
	return str
}

func MakeOpensourceLicenseKey() *LicenseKey {
	lk := LicenseKey{}
	lk.LicenseId = opensourceLicenseId
	lk.CustomerId = opensourceLicenseUuid
	lk.CustomerName = "Open Source Evangelist"
	lk.Type = LicenseTypeOpensource
	lk.Features = getSupportedFeatures()
	lk.CreatedAt = time.Now().UTC()
	lk.CreatedAtInt = lk.CreatedAt.Unix()
	lk.ExpiresAt = lk.CreatedAt.AddDate(10, 0, 0)
	lk.ExpiresAtInt = lk.ExpiresAt.Unix()
	lk.CreatorName = "UniDoc Support"
	lk.CreatorEmail = "support@unidoc.io"
	return &lk
}
