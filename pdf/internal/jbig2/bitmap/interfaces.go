package bitmap

// BitmapGetter gets the bitmap from the segment
type BitmapGetter interface {
	GetBitmap() *Bitmap
}

// BitmapsLister is the interface that returns bitmaps for given struct
type BitmapsLister interface {
	ListBitmaps() []*Bitmap
}
