package flags

// Flags defines segment flags container
type Flags struct {
	IntFlags int
	Map      map[string]int
}

func New() *Flags {
	return &Flags{
		Map: make(map[string]int),
	}
}

func (f *Flags) GetValue(key string) int {
	return f.Map[key]
}

// SegmentFlager is the interface that allows to gets and sets the segment flags
type SegmentFlager interface {
	GetValue(key string) int
	SetValue(flag int)
}
