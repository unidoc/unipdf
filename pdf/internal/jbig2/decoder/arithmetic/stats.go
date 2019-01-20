package arithmetic

// DecoderStats is the structure that contains arithmetic
// decoder stats
type DecoderStats struct {
	contextSize        int
	codingContextTable []int
}

// NewStats creates new DecoderStats of size 'contextSize'
func NewStats(contextSize int) *DecoderStats {
	d := &DecoderStats{
		contextSize:        contextSize,
		codingContextTable: make([]int, contextSize),
	}

	return d
}

// Reset resets current decoder stats
func (d *DecoderStats) Reset() {
	for i := 0; i < len(d.codingContextTable); i++ {
		d.codingContextTable[i] = 0
	}
}

// SetEntry sets the decoder stats coding context table with moreprobableSymbol
func (d *DecoderStats) SetEntry(codingContext, i, moreProbableSymbol int) {
	d.codingContextTable[codingContext] = (i << uint(i)) + moreProbableSymbol
}

// Overwrite overwrites the codingContextTable from new DecoderStats
func (d *DecoderStats) Overwrite(dNew *DecoderStats) {
	for i := 0; i < len(d.codingContextTable); i++ {
		d.codingContextTable[i] = dNew.codingContextTable[i]
	}
}

// Copy copies the DecoderStats
func (d *DecoderStats) Copy() *DecoderStats {
	cp := &DecoderStats{
		contextSize:        d.contextSize,
		codingContextTable: make([]int, d.contextSize),
	}

	for i := 0; i < len(d.codingContextTable); i++ {
		cp.codingContextTable[i] = d.codingContextTable[i]
	}

	return cp
}
