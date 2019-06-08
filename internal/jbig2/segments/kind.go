/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package segments

// Type defines the jbig2 segment type - see 7.3.
type Type int

// Enumerate segment type definitions.
const (
	TSymbolDictionary                         Type = 0
	TIntermediateTextRegion                   Type = 4
	TImmediateTextRegion                      Type = 6
	TImmediateLosslessTextRegion              Type = 7
	TPatternDictionary                        Type = 16
	TIntermediateHalftoneRegion               Type = 20
	TImmediateHalftoneRegion                  Type = 22
	TImmediateLosslessHalftoneRegion          Type = 23
	TIntermediateGenericRegion                Type = 36
	TImmediateGenericRegion                   Type = 38
	TImmediateLosslessGenericRegion           Type = 39
	TIntermediateGenericRefinementRegion      Type = 40
	TImmediateGenericRefinementRegion         Type = 42
	TImmediateLosslessGenericRefinementRegion Type = 43
	TPageInformation                          Type = 48
	TEndOfPage                                Type = 49
	TEndOfStrip                               Type = 50
	TEndOfFile                                Type = 51
	TProfiles                                 Type = 52
	TTables                                   Type = 53
	TExtension                                Type = 62
	TBitmap                                   Type = 70
)

// String implements Stringer interface.
func (k Type) String() string {
	switch k {
	case TSymbolDictionary:
		return "Symbol Dictionary"
	case TIntermediateTextRegion:
		return "Intermediate Text Region"
	case TImmediateTextRegion:
		return "Immediate Text Region"
	case TImmediateLosslessTextRegion:
		return "Immediate Lossless Text Region"
	case TPatternDictionary:
		return "Pattern Dictionary"
	case TIntermediateHalftoneRegion:
		return "Intermediate Halftone Region"
	case TImmediateHalftoneRegion:
		return "Immediate Halftone Region"
	case TImmediateLosslessHalftoneRegion:
		return "Immediate Lossless Halftone Region"
	case TIntermediateGenericRegion:
		return "Intermediate Generic Region"
	case TImmediateGenericRegion:
		return "Immediate Generic Region"
	case TImmediateLosslessGenericRegion:
		return "Immediate Lossless Generic Region"
	case TIntermediateGenericRefinementRegion:
		return "Intermediate Generic Refinement Region"
	case TImmediateGenericRefinementRegion:
		return "Immediate Generic Refinement Region"
	case TImmediateLosslessGenericRefinementRegion:
		return "Immediate Lossless Generic Refinement Region"
	case TPageInformation:
		return "Page Information"
	case TEndOfPage:
		return "End Of Page"
	case TEndOfStrip:
		return "End Of Strip"
	case TEndOfFile:
		return "End Of File"
	case TProfiles:
		return "Profiles"
	case TTables:
		return "Tables"
	case TExtension:
		return "Extension"
	case TBitmap:
		return "Bitmap"
	}
	return "Invalid Segment Kind"
}

// set the segment type number to it's creator function.
var (
	emptySegment Segmenter
	kindMap      = map[Type]func() Segmenter{
		TSymbolDictionary:                         func() Segmenter { return &SymbolDictionary{} },
		TIntermediateTextRegion:                   func() Segmenter { return &TextRegion{} },
		TImmediateTextRegion:                      func() Segmenter { return &TextRegion{} },
		TImmediateLosslessTextRegion:              func() Segmenter { return &TextRegion{} },
		TPatternDictionary:                        func() Segmenter { return &PatternDictionary{} },
		TIntermediateHalftoneRegion:               func() Segmenter { return &HalftoneRegion{} },
		TImmediateHalftoneRegion:                  func() Segmenter { return &HalftoneRegion{} },
		TImmediateLosslessHalftoneRegion:          func() Segmenter { return &HalftoneRegion{} },
		TIntermediateGenericRegion:                func() Segmenter { return &GenericRegion{} },
		TImmediateGenericRegion:                   func() Segmenter { return &GenericRegion{} },
		TImmediateLosslessGenericRegion:           func() Segmenter { return &GenericRegion{} },
		TIntermediateGenericRefinementRegion:      func() Segmenter { return &GenericRefinementRegion{} },
		TImmediateGenericRefinementRegion:         func() Segmenter { return &GenericRefinementRegion{} },
		TImmediateLosslessGenericRefinementRegion: func() Segmenter { return &GenericRefinementRegion{} },
		TPageInformation:                          func() Segmenter { return &PageInformationSegment{} },
		TEndOfPage:                                func() Segmenter { return emptySegment },
		TEndOfStrip:                               func() Segmenter { return &EndOfStripe{} },
		TEndOfFile:                                func() Segmenter { return emptySegment },
		TProfiles:                                 func() Segmenter { return emptySegment },
		TTables:                                   func() Segmenter { return &TableSegment{} },
		TExtension:                                func() Segmenter { return emptySegment },
		TBitmap:                                   func() Segmenter { return emptySegment },
	}
)
