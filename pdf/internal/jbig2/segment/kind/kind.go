package kind

type SegmentKind int

const (
	SymbolDictionary                         SegmentKind = 0
	IntermediateTextRegion                   SegmentKind = 4
	ImmediateTextRegion                      SegmentKind = 6
	ImmediateLosslessTextRegion              SegmentKind = 7
	PatternDictionary                        SegmentKind = 16
	IntermediateHalftoneRegion               SegmentKind = 20
	ImmediateHalftoneRegion                  SegmentKind = 22
	ImmediateLosslessHalftoneRegion          SegmentKind = 23
	IntermediateGenericRegion                SegmentKind = 36
	ImmediateGenericRegion                   SegmentKind = 38
	ImmediateLosslessGenericRegion           SegmentKind = 39
	IntermediateGenericRefinementRegion      SegmentKind = 40
	ImmediateGenericRefinementRegion         SegmentKind = 42
	ImmediateLosslessGenericRefinementRegion SegmentKind = 43
	PageInformation                          SegmentKind = 48
	EndOfPage                                SegmentKind = 49
	EndOfStrip                               SegmentKind = 50
	EndOfFile                                SegmentKind = 51
	Profiles                                 SegmentKind = 52
	Tables                                   SegmentKind = 53
	Extension                                SegmentKind = 62
	Bitmap                                   SegmentKind = 70
)

func (k SegmentKind) String() string {
	switch k {
	case SymbolDictionary:
		return "Symbol Dictionary"
	case IntermediateTextRegion:
		return "Intermediate Text Region"
	case ImmediateTextRegion:
		return "Immediate Text Region"
	case ImmediateLosslessTextRegion:
		return "Immediate Lossless Text Region"
	case PatternDictionary:
		return "Pattern Dictionary"
	case IntermediateHalftoneRegion:
		return "Intermediate Halftone Region"
	case ImmediateHalftoneRegion:
		return "Immediate Halftone Region"
	case ImmediateLosslessHalftoneRegion:
		return "Immediate Lossless Halftone Region"
	case IntermediateGenericRegion:
		return "Intermediate Generic Region"
	case ImmediateGenericRegion:
		return "Immediate Generic Region"
	case ImmediateLosslessGenericRegion:
		return "Immediate Lossless Generic Region"
	case IntermediateGenericRefinementRegion:
		return "Intermediate Generic Refinement Region"
	case ImmediateGenericRefinementRegion:
		return "Immediate Generic Refinement Region"
	case ImmediateLosslessGenericRefinementRegion:
		return "Immediate Lossless Generic Refinement Region"
	case PageInformation:
		return "Page Information"
	case EndOfPage:
		return "End Of Page"
	case EndOfStrip:
		return "End Of Strip"
	case EndOfFile:
		return "End Of File"
	case Profiles:
		return "Profiles"
	case Tables:
		return "Tables"
	case Extension:
		return "Extension"
	case Bitmap:
		return "Bitmap"
	}
	return "Invalid Segment Kind"
}
