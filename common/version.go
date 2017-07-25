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
const releaseDay = 25
const releaseHour = 14
const releaseMin = 37

// Holds version information, when bumping this make sure to bump the released at stamp also.
const Version = "2.0.0"

var ReleasedAt = time.Date(releaseYear, releaseMonth, releaseDay, releaseHour, releaseMin, 0, 0, time.UTC)
