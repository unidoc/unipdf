package imageutil

import (
	"image"
	"image/draw"
	"testing"
)

func BenchmarkConvertDrawImage(b *testing.B) {
	type benchmarkSuite struct {
		name string
		src  image.Image
		dst  func() draw.Image
	}

	rect := image.Rect(0, 0, 1024, 1024)

	suites := []benchmarkSuite{
		{
			name: "GrayToNRGB",
			src:  image.NewGray(rect),
			dst:  func() draw.Image { return image.NewNRGBA(rect) },
		},
		{
			name: "GrayToCMYK",
			src:  image.NewGray(rect),
			dst:  func() draw.Image { return image.NewCMYK(rect) },
		},
		{
			name: "CMYKToNRGB",
			src:  image.NewCMYK(rect),
			dst:  func() draw.Image { return image.NewNRGBA(rect) },
		},
		{
			name: "CMYKToGray",
			src:  image.NewCMYK(rect),
			dst:  func() draw.Image { return image.NewGray(rect) },
		},
		{
			name: "NRGBToCMYK",
			src:  image.NewNRGBA(rect),
			dst:  func() draw.Image { return image.NewCMYK(rect) },
		},
		{
			name: "NRGBToGray",
			src:  image.NewNRGBA(rect),
			dst:  func() draw.Image { return image.NewGray(rect) },
		},
		{
			name: "RGBToNRGBA",
			src:  image.NewRGBA(rect),
			dst:  func() draw.Image { return image.NewNRGBA(rect) },
		},
		{
			name: "RGBToGray",
			src:  image.NewRGBA(rect),
			dst:  func() draw.Image { return image.NewGray(rect) },
		},
		{
			name: "RGBToCMYK",
			src:  image.NewNRGBA(rect),
			dst:  func() draw.Image { return image.NewCMYK(rect) },
		},
	}

	for _, suite := range suites {
		b.Run(suite.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				dst := suite.dst()
				src := suite.src
				for x := 0; x < rect.Max.X; x++ {
					for y := 0; y < rect.Max.Y; y++ {
						dst.Set(x, y, src.At(x, y))
					}
				}
			}
		})
	}
}

func BenchmarkConverterFunc(b *testing.B) {
	type benchmarkSuite struct {
		name      string
		src       image.Image
		converter ColorConverter
	}
	rect := image.Rect(0, 0, 1024, 1024)
	suites := []benchmarkSuite{
		{
			name:      "GrayToNRGB",
			src:       image.NewGray(rect),
			converter: NRGBAConverter,
		},
		{
			name:      "GrayToCMYK",
			src:       image.NewGray(rect),
			converter: CMYKConverter,
		},
		{
			name:      "CMYKToNRGB",
			src:       image.NewCMYK(rect),
			converter: NRGBAConverter,
		},
		{
			name:      "CMYKToGray",
			src:       image.NewCMYK(rect),
			converter: GrayConverter,
		},
		{
			name:      "NRGBToCMYK",
			src:       image.NewNRGBA(rect),
			converter: CMYKConverter,
		},
		{
			name:      "NRGBToGray",
			src:       image.NewNRGBA(rect),
			converter: GrayConverter,
		},
		{
			name:      "RGBToNRGBA",
			src:       image.NewRGBA(rect),
			converter: NRGBAConverter,
		},
		{
			name:      "RGBToGray",
			src:       image.NewRGBA(rect),
			converter: GrayConverter,
		},
		{
			name:      "RGBToCMYK",
			src:       image.NewNRGBA(rect),
			converter: CMYKConverter,
		},
	}
	for _, suite := range suites {
		b.Run(suite.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				src := suite.src
				_, _ = suite.converter.Convert(src)
			}
		})
	}
}
