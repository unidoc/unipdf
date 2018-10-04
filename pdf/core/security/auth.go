package security

// AuthEvent is an event type that triggers authentication.
type AuthEvent string

const (
	EventDocOpen = AuthEvent("DocOpen") // document open
	EventEFOpen  = AuthEvent("EFOpen")  // embedded file open
)
