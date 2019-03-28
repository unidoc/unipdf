/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package common contains common properties used by the subpackages.
package common

import (
	"time"
)

const releaseYear = 2019
const releaseMonth = 3
const releaseDay = 28
const releaseHour = 10
const releaseMin = 11

// Version holds version information, when bumping this make sure to bump the released at stamp also.
const Version = "3.0.0-alpha.3"

var ReleasedAt = time.Date(releaseYear, releaseMonth, releaseDay, releaseHour, releaseMin, 0, 0, time.UTC)
