package security

import "math"

// Permissions is a bitmask of access permissions for a PDF file.
type Permissions uint32

const (
	// PermOwner grants all permissions.
	PermOwner = Permissions(math.MaxUint32)

	PermPrinting        = Permissions(1 << 2) // bit 3
	PermModify          = Permissions(1 << 3) // bit 4
	PermExtractGraphics = Permissions(1 << 4) // bit 5
	PermAnnotate        = Permissions(1 << 5) // bit 6
	// PermFillForms allow form filling, if annotation is disabled?  If annotation enabled, is not looked at.
	PermFillForms         = Permissions(1 << 8) // bit 9
	PermDisabilityExtract = Permissions(1 << 9) // bit 10 // TODO: not clear what this means!
	// PermRotateInsert allows rotating, editing page order.
	PermRotateInsert = Permissions(1 << 10) // bit 11
	// PermFullPrintQuality limits print quality (lowres), assuming Printing bit is set.
	PermFullPrintQuality = Permissions(1 << 11) // bit 12
)

// Allowed checks if a set of permissions can be granted.
func (p Permissions) Allowed(p2 Permissions) bool {
	return p&p2 == p2
}
