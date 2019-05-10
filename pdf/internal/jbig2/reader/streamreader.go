package reader

// StreamReader is the interface that allows to read bit, bits, byte, bytes change and get the
// stream position, align the bits
// Implements io.Reader, io.Seeker interfaces
type StreamReader interface {
	ReadBit() (int, error)
	ReadBool() (bool, error)
	ReadByte() (byte, error)
	ReadBits(n byte) (uint64, error)
	Read(b []byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
	StreamPosition() int64
	Length() uint64
	Align() byte
	BitPosition() int
}

var (
	_ StreamReader = &Reader{}
	_ StreamReader = &SubstreamReader{}
)
