package bitmap

// BMPointer is the BitmapPointer structure that contain the parameters that points to the bitmap
// pixel
type BMPointer struct {
	X, Y          int
	Width, Height int

	Bits, Count int
	Output      bool

	BM *Bitmap
}

func NewPointer(b *Bitmap) *BMPointer {
	return &BMPointer{
		BM: b,
	}
}

func (b *BMPointer) SetPointer(x, y int) {
	b.X = x
	b.Y = y

	b.Output = true

	if y < 0 || y >= b.Height || x >= b.Width {
		b.Output = false
	}

	b.Count = y * b.Width
}

func (b *BMPointer) NextPixel() int {
	if !b.Output {
		return 0
	} else if b.X < 0 || b.X >= b.Width {
		b.X++
		return 0
	}
	pixel, err := b.BM.Data.Get(uint(b.Count + b.X))
	if err != nil {
		panic(err)
	}
	b.X++

	if pixel {
		return 1
	}
	return 0
}
