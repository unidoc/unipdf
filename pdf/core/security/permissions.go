/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package security

import "math"

// Permissions is a bitmask of access permissions for a PDF file.
type Permissions uint32

const (
	// PermOwner grants all permissions.
	PermOwner = Permissions(math.MaxUint32)

	// PermPrinting allows printing the document with a low quality.
	PermPrinting = Permissions(1 << 2)
	// PermModify allows to modify the document.
	PermModify = Permissions(1 << 3)
	// PermExtractGraphics allows to extract graphics from the document.
	PermExtractGraphics = Permissions(1 << 4)
	// PermAnnotate allows annotating the document.
	PermAnnotate = Permissions(1 << 5)
	// PermFillForms allow form filling, if annotation is disabled?  If annotation enabled, is not looked at.
	PermFillForms = Permissions(1 << 8)
	// PermDisabilityExtract allows to extract graphics in accessibility mode.
	PermDisabilityExtract = Permissions(1 << 9)
	// PermRotateInsert allows rotating, editing page order.
	PermRotateInsert = Permissions(1 << 10)
	// PermFullPrintQuality limits print quality (lowres), assuming Printing bit is set.
	PermFullPrintQuality = Permissions(1 << 11)
)

// Allowed checks if a set of permissions can be granted.
func (p Permissions) Allowed(p2 Permissions) bool {
	return p&p2 == p2
}
