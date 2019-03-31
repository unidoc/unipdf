package jbig2

// SegmentType defines the segment
type SegmentType int

const (
	TSymbolDictionary                         SegmentType = 0
	TIntermediateTextRegion                   SegmentType = 4
	TImmediateTextRegion                      SegmentType = 6
	TImmediateLosslessTextRegion              SegmentType = 7
	TPatternDictionary                        SegmentType = 16
	TIntermediateHalftoneRegion               SegmentType = 20
	TImmediateHalftoneRegion                  SegmentType = 22
	TImmediateLosslessHalftoneRegion          SegmentType = 23
	TIntermediateGenericRegion                SegmentType = 36
	TImmediateGenericRegion                   SegmentType = 38
	TImmediateLosslessGenericRegion           SegmentType = 39
	TIntermediateGenericRefinementRegion      SegmentType = 40
	TImmediateGenericRefinementRegion         SegmentType = 42
	TImmediateLosslessGenericRefinementRegion SegmentType = 43
	TPageInformation                          SegmentType = 48
	TEndOfPage                                SegmentType = 49
	TEndOfStrip                               SegmentType = 50
	TEndOfFile                                SegmentType = 51
	TProfiles                                 SegmentType = 52
	TTables                                   SegmentType = 53
	TExtension                                SegmentType = 62
	TBitmap                                   SegmentType = 70
)

// String SegmentType implements Stringer interface
func (k SegmentType) String() string {
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

var emptySegment Segmenter

var kindMap map[SegmentType]func() Segmenter = map[SegmentType]func() Segmenter{
	TSymbolDictionary: func() Segmenter {
		return &SymbolDictionarySegment{}
	},
	TIntermediateTextRegion:                   func() Segmenter { return &TextRegion{} },
	TImmediateTextRegion:                      func() Segmenter { return &TextRegion{} },
	TImmediateLosslessTextRegion:              func() Segmenter { return &TextRegion{} },
	TPatternDictionary:                        func() Segmenter { return &PatternDictionary{} },
	TIntermediateHalftoneRegion:               func() Segmenter { return &HalftoneRegion{} },
	TImmediateHalftoneRegion:                  func() Segmenter { return &HalftoneRegion{} },
	TImmediateLosslessHalftoneRegion:          func() Segmenter { return &HalftoneRegion{} },
	TIntermediateGenericRegion:                func() Segmenter { return &GenericRegionSegment{} },
	TImmediateGenericRegion:                   func() Segmenter { return &GenericRegionSegment{} },
	TImmediateLosslessGenericRegion:           func() Segmenter { return &GenericRegionSegment{} },
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
