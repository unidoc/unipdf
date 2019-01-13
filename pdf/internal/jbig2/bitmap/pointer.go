package bitmap

// BMPointer is the BitmapPointer structure that contain the parameters that points to the bitmap
// pixel
type BMPointer struct {
	X, Y   int
	Output bool

	Count int
}
