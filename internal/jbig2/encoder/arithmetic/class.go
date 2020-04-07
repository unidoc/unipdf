/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package arithmetic

// Class is the arithmetic encoding context class.
type Class int

// Enumerated definitions for the encoding classes.
const (
	// IAAI used to decode number of symbol instances in an aggregation.
	IAAI Class = iota
	// IADH used to decode difference in height between two height classes.
	IADH
	// IADS used to decode the S coordinate of the second and subsequent
	// symbol instances in a strip.
	IADS
	// IADT used to decode the T coordinate of the second and subsequent
	// symbol instances in a strip.
	IADT
	// IADW used to decode the difference in width between two symbols in
	// a height class
	IADW
	// IAEX used to decode export flags.
	IAEX
	// IAFS used to decode the S coordinate of the first symbol instance
	// in a strip.
	IAFS
	// IAIT used to decode the T coordinate of the symbol instances in a strip.
	IAIT
	// IARDH used to decode the delta height of symbol instance refinements.
	IARDH
	// IARDW used to decode the delta width of symbol instance refinements.
	IARDW
	// IARDX used to decode the delta X position of symbol instance refinements.
	IARDX
	// IARDY used to decode the delta Y position of symbol instance refinements.
	IARDY
	// IARI used to decode the Ri bit of symbol instances.
	IARI
)

// String implements fmt.Stringer interface.
func (c Class) String() string {
	switch c {
	case IAAI:
		return "IAAI"
	case IADH:
		return "IADH"
	case IADS:
		return "IADS"
	case IADT:
		return "IADT"
	case IADW:
		return "IADW"
	case IAEX:
		return "IAEX"
	case IAFS:
		return "IAFS"
	case IAIT:
		return "IAIT"
	case IARDH:
		return "IARDH"
	case IARDW:
		return "IARDW"
	case IARDX:
		return "IARDX"
	case IARDY:
		return "IARDY"
	case IARI:
		return "IARI"
	default:
		return "UNKNOWN"
	}
}
