/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Package common contains common properties used by the subpackages.
package common

import (
	"time"
)

const releaseYear = 2018
const releaseMonth = 8
const releaseDay = 14
const releaseHour = 19
const releaseMin = 40

// Version holds version information, when bumping this make sure to bump the released at stamp also.
const Version = "2.1.1"

var ReleasedAt = time.Date(releaseYear, releaseMonth, releaseDay, releaseHour, releaseMin, 0, 0, time.UTC)
