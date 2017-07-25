/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package common contains common properties used by the subpackages.
package common

import (
	"time"
)

const releaseYear = 2017
const releaseMonth = 7
const releaseDay = 17
const releaseHour = 20
const releaseMin = 45

// Holds version information, when bumping this make sure to bump the released at stamp also.
const Version = "2.0.0-rc.1"

var ReleasedAt = time.Date(releaseYear, releaseMonth, releaseDay, releaseHour, releaseMin, 0, 0, time.UTC)
