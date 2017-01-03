/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package common contains common properties used by the subpackages.
package common

import (
	"time"
)

const releaseYear = 2016
const releaseMonth = 7
const releaseDay = 9
const releaseHour = 12
const releaseMin = 00

// Holds version information, when bumping this make sure to bump the released
// at stamp also. This is for license information to make sure license
const Version = "1.1.0"

var ReleasedAt = time.Date(releaseYear, releaseMonth, releaseDay, releaseHour, releaseMin, 0, 0, time.UTC)
