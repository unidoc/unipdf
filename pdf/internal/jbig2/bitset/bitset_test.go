package bitset

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unidoc/unidoc/common"
	// "github.com/unidoc/unidoc/common"
	"testing"
)

func TestBitsetSet(t *testing.T) {

	// common.Log = common.NewConsoleLogger(common.LogLevelDebug)
	type bitsetSet struct {
		index  uint
		value  bool
		length int

		testFunc func(t *testing.T, bs *BitSet, bss bitsetSet)
	}

	checkIndex := func(t *testing.T, bs *BitSet, bss bitsetSet) {
		t.Helper()

		dIndex := bss.index / 64

		assert.True(t, int(dIndex) <= len(bs.Data))

	}

	testCases := map[string]bitsetSet{
		"FirstIndex": {
			index:  62,
			value:  true,
			length: 1,
			testFunc: func(t *testing.T, bs *BitSet, bss bitsetSet) {

				checkIndex(t, bs, bss)

				err := bs.Set(bss.index, bss.value)
				if assert.NoError(t, err) {
					if assert.Len(t, bs.Data, 1) {
						assert.Equal(t, "0100000000000000000000000000000000000000000000000000000000000000",
							fmt.Sprintf("%064b", bs.Data[0]),
						)
					}

				}
			},
		},
		"NthIndex": {
			index:  1279,
			value:  true,
			length: 1280,
			testFunc: func(t *testing.T, bs *BitSet, bss bitsetSet) {
				checkIndex(t, bs, bss)

				err := bs.Set(bss.index, bss.value)
				if assert.NoError(t, err) {
					if assert.Len(t, bs.Data, (bss.length / 64)) {
						assert.Equal(t,
							"1000000000000000000000000000000000000000000000000000000000000000",
							fmt.Sprintf("%064b", bs.Data[19]),
						)
					}
				}
			},
		},
		"OutOfRange": {
			index:  65,
			value:  true,
			length: 64,
			testFunc: func(t *testing.T, bs *BitSet, bss bitsetSet) {
				assert.Error(t, bs.Set(bss.index, bss.value))
			},
		},

		"SetFalse": {
			index:  63,
			value:  true,
			length: 64,
			testFunc: func(t *testing.T, bs *BitSet, bss bitsetSet) {
				checkIndex(t, bs, bss)

				if assert.NoError(t, bs.Set(bss.index, bss.value)) {
					assert.Equal(t,
						"1000000000000000000000000000000000000000000000000000000000000000",
						fmt.Sprintf("%064b", bs.Data[0]),
					)

					if assert.NoError(t, bs.Set(bss.index, !bss.value)) {
						assert.Equal(t,
							"0000000000000000000000000000000000000000000000000000000000000000",
							fmt.Sprintf("%064b", bs.Data[0]),
						)
					}
				}
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			bs := NewBitSet(testCase.length)

			testCase.testFunc(t, bs, testCase)

		})
	}
}

func TestBitSetSetAll(t *testing.T) {
	bs := NewBitSet(10)

	bs.SetAll(true)

	for _, v := range bs.Data {
		assert.Equal(t,
			"1111111111111111111111111111111111111111111111111111111111111111",
			fmt.Sprintf("%064b", v),
		)
	}

	bs.SetAll(false)

	for _, v := range bs.Data {
		assert.Equal(t,
			"0000000000000000000000000000000000000000000000000000000000000000",
			fmt.Sprintf("%064b", v),
		)
	}

}

func TestBitSetGet(t *testing.T) {
	t.Run("FirstIndex", func(t *testing.T) {
		bs := NewBitSet(129)

		if assert.Len(t, bs.Data, 3) {

			bs.Data[0] = uint64(1) << 63

			value, err := bs.Get(63)
			if assert.NoError(t, err) {
				assert.True(t, value)
			}
		}
	})

	t.Run("NthIndex", func(t *testing.T) {
		bs := NewBitSet(129)

		if assert.Len(t, bs.Data, 3) {

			bs.Data[1] = uint64(1) << 63

			value, err := bs.Get(64 + 63)
			if assert.NoError(t, err) {
				assert.True(t, value, fmt.Sprintf("%064b", bs.Data[1]))
			}
		}
	})
}

func TestBitSetBytes(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
	}

	t.Run("LargerThanSingleUint64", func(t *testing.T) {
		bs := NewBitSet(129)
		err := bs.Set(129, true)
		require.NoError(t, err)

		bytes := bs.Bytes()

		assert.Len(t, bytes, (129/8 + 1))
		for i := 0; i < len(bytes)-1; i++ {
			assert.Equal(t, byte(0), bytes[i])
		}
		assert.Equal(t, byte(2), bytes[len(bytes)-1])

		b, err := bs.GetByte(124)
		if assert.NoError(t, err) {
			assert.Equal(t, byte(1<<5), b)
		}
	})

	t.Run("SmallerThan8bytes", func(t *testing.T) {
		bs := NewBitSet(15)

		err := bs.Set(15, true)
		require.NoError(t, err)

		bytes := bs.Bytes()
		assert.Len(t, bytes, 2)

		assert.Equal(t, byte(0), bytes[0])
		assert.Equal(t, byte(128), bytes[1])

		b, err := bs.GetByte(8)
		if assert.NoError(t, err) {
			assert.Equal(t, byte(1<<7), b)
		}
	})

}

func benchmarkInitBitSet(size int, b *testing.B) {
	NewBitSet(size)
}

func BenchmarkInitBitSet1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkInitBitSet(1000, b)
	}
}

func BenchmarkInitBitSet1M(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkInitBitSet(1000000, b)
	}
}

func BenchmarkInitBitSet1G(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkInitBitSet(1000000000, b)
	}
}

func benchmarkInitBoolArray(size int, b *testing.B) {
	_ = make([]bool, size)

}
func BenchmarkInitBoolArray1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkInitBoolArray(1000, b)
	}
}

func BenchmarkInitBoolArray1M(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkInitBoolArray(1000000, b)
	}
}

func BenchmarkInitBoolArray1G(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkInitBoolArray(1000000000, b)
	}
}

func benchmarkBitsetSetWithInit(size int, b *testing.B) {
	bs := NewBitSet(size)

	bs.Set(0, true)

	bs.Set(uint(size/2), true)

	if size > 0 {
		bs.Set(uint(size-1), true)
	}
}

func benchmarkBitsetSetWithoutInit(size int, b *testing.B) {
	b.StopTimer()
	bs := NewBitSet(size)
	b.StartTimer()

	bs.Set(0, true)

	bs.Set(uint(size/2), true)

	if size > 0 {
		bs.Set(uint(size-1), true)
	}

}

func benchmarkBoolArraySetWithoutInit(b *testing.B, bs []bool, size int) {

	// firsts
	bs[0] = true
	// mids
	bs[size/2] = true
	// lasts
	if size > 0 {
		bs[size-1] = true
	}
}

func benchmarkBoolArraySetWithInit(size int, b *testing.B) {

	bs := make([]bool, size)

	// firsts
	bs[0] = true
	// mids
	bs[size/2] = true
	// lasts
	if size > 0 {
		bs[size-1] = true
	}

}

func BenchmarkBitsetSetWithInit1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkBitsetSetWithInit(1000, b)
	}
}

func BenchmarkBitsetSetWithInit1000000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkBitsetSetWithInit(1000000, b)
	}
}

func BenchmarkBitsetSetWithInit1000000000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkBitsetSetWithInit(1000000000, b)
	}
}

// func BenchmarkBitsetSetWithoutInit1000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		b.
// 		benchmarkBitsetSetWithoutInit(b)
// 	}
// }

// func BenchmarkBitsetSetWithoutInit1000000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		benchmarkBitsetSetWithoutInit(1000000, b)
// 	}
// }

// func BenchmarkBitsetSetWithoutInit1000000000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		benchmarkBitsetSetWithoutInit(1000000000, b)
// 	}
// }

// func BenchmarkBoolArraySetWithoutInit1000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		benchmarkBoolArraySetWithoutInit(1000, b)
// 	}
// }

// func BenchmarkBoolArraySetWithoutInit1000000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		benchmarkBoolArraySetWithoutInit(1000000, b)
// 	}
// }

// func BenchmarkBoolArraySetWithoutInit1000000000(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		benchmarkBoolArraySetWithoutInit(1000000000, b)
// 	}
// }

func BenchmarkBoolArraySetWithInit1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkBoolArraySetWithInit(1000, b)
	}
}

func BenchmarkBoolArraySetWithInit1000000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkBoolArraySetWithInit(1000000, b)
	}
}

func BenchmarkBoolArraySetWithInit1000000000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkBoolArraySetWithInit(1000000000, b)
	}
}

// func BenchmarkBitSetSetAll(b *testing.B) {
// 	for i := 1; i < b.N; i++ {
// 		bs := NewBitSet(i)
// 		bs.SetAll(true)
// 	}
// }

// func BenchmarkBoolArraySetAll(b *testing.B) {
// 	for i := 1; i < b.N; i++ {
// 		bs := make([]bool, i)

// 		for j := 0; j < len(bs); j++ {
// 			bs[j] = true
// 		}
// 	}
// }
