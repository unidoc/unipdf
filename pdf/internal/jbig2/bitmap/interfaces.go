package bitmap

// Getter gets the bitmap from the segment
type Getter interface {
	GetBitmap() *Bitmap
}

// BitmapsLister is the interface that returns bitmaps for given struct
type BitmapsLister interface {
	ListBitmaps() []*Bitmap
}
