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
