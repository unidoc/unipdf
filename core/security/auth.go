/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package security

// AuthEvent is an event type that triggers authentication.
type AuthEvent string

const (
	// EventDocOpen is an event triggered when opening the document.
	EventDocOpen = AuthEvent("DocOpen")
	// EventEFOpen is an event triggered when accessing an embedded file.
	EventEFOpen = AuthEvent("EFOpen")
)
