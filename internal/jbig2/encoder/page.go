package encoder

// Page is the encoded JBIG2 page structure helper.
type Page struct {
	Components       []int
	SingleUseSymbols []uint
	XRes, YRes       int
	Width, Height    int
	// BaseIndex is the number of the first symbol of each page
	// Used only on Refinement.
	BaseIndex int
}
