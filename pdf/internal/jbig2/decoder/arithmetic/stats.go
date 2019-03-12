package arithmetic

// DecoderStats is the structure that contains arithmetic
// decoder stats
type DecoderStats struct {
	index              int
	contextSize        int
	codingContextTable []byte
	mps                []byte
}

// NewStats creates new DecoderStats of size 'contextSize'
func NewStats(contextSize int, index int) *DecoderStats {
	d := &DecoderStats{
		index:              index,
		contextSize:        contextSize,
		codingContextTable: make([]byte, contextSize),
		mps:                make([]byte, contextSize),
	}

	return d
}

// Reset resets current decoder stats
func (d *DecoderStats) Reset() {
	for i := 0; i < len(d.codingContextTable); i++ {
		d.codingContextTable[i] = 0
		d.mps[i] = 0
	}
}

// SetEntry sets the decoder stats coding context table with moreprobableSymbol
func (d *DecoderStats) SetEntry(value int) {
	d.codingContextTable[d.index] = byte(value)
}

func (d *DecoderStats) SetIndex(index int) {
	d.index = index
}

// Overwrite overwrites the codingContextTable from new DecoderStats
func (d *DecoderStats) Overwrite(dNew *DecoderStats) {
	for i := 0; i < len(d.codingContextTable); i++ {
		d.codingContextTable[i] = dNew.codingContextTable[i]
		d.mps[i] = dNew.mps[i]
	}
}

func (d *DecoderStats) toggleMps() {
	d.mps[d.index] ^= 1
}

func (d *DecoderStats) getMps() byte {
	return d.mps[d.index]
}

// Copy copies the DecoderStats
func (d *DecoderStats) Copy() *DecoderStats {
	cp := &DecoderStats{
		contextSize:        d.contextSize,
		codingContextTable: make([]byte, d.contextSize),
	}

	for i := 0; i < len(d.codingContextTable); i++ {
		cp.codingContextTable[i] = d.codingContextTable[i]
	}

	return cp
}

func (d *DecoderStats) cx() byte {
	return d.codingContextTable[d.index]
}
