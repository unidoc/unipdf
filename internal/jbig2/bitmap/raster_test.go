/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"github.com/unidoc/unipdf/common"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRasterOperation tests the RasterOperation function.
func TestRasterOperation(t *testing.T) {
	if testing.Verbose() {
		common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
	}

	type operatorName struct {
		Operator RasterOperator
		Name     string
	}
	t.Run("NilDest", func(t *testing.T) {
		err := RasterOperation(nil, 1, 2, 0, 0, PixSrcXorDst, New(3, 3), 0, 0)
		require.Error(t, err)
	})

	t.Run("PixDest", func(t *testing.T) {
		dest := New(5, 5)
		err := RasterOperation(dest, 0, 0, 0, 0, PixDst, nil, 0, 0)
		require.NoError(t, err)
	})

	t.Run("UniLow", func(t *testing.T) {
		t.Run("Dw0", func(t *testing.T) {
			dest := New(5, 5)
			err := RasterOperation(dest, 0, 0, 0, 0, PixSet, nil, 0, 0)
			require.NoError(t, err)
		})

		operators := []operatorName{{PixClr, "Clear"}, {PixSet, "Set"}, {PixNotDst, "NotDst"}}
		t.Run("Aligned", func(t *testing.T) {
			for _, op := range operators {
				t.Run(op.Name, func(t *testing.T) {
					data := []byte{0xff, 0xfd, 0xfa, 0xf3, 0xac, 0xbc, 0xdc, 0xaf}
					tocheck := make([]byte, 8)
					copy(tocheck, data)

					dest, err := NewWithData(8, 8, data)
					require.NoError(t, err)

					err = dest.RasterOperation(0, 0, dest.Width, dest.Height, op.Operator, nil, 0, 0)
					require.NoError(t, err)

					switch op.Operator {
					case PixClr:
						for _, bt := range dest.Data {
							assert.Equal(t, byte(0x00), bt)
						}
					case PixSet:
						for _, bt := range dest.Data {
							assert.Equal(t, byte(0xff), bt)
						}
					case PixNotDst:
						for i, bt := range dest.Data {
							assert.Equal(t, (^tocheck[i]), bt)
						}
					}
				})
			}

			t.Run("WithLwBits", func(t *testing.T) {
				common.SetLogger(common.NewConsoleLogger(common.LogLevelTrace))
				for _, op := range operators {
					t.Run(op.Name, func(t *testing.T) {
						data := []byte{0xff, 0xfd, 0xfa, 0xf3, 0xac, 0xbc, 0xdc, 0xaf}
						tocheck := make([]byte, 8)
						copy(tocheck, data)

						dest, err := NewWithData(7, 8, data)
						require.NoError(t, err)

						err = dest.RasterOperation(0, 0, dest.Width, dest.Height, op.Operator, nil, 0, 0)
						require.NoError(t, err)

						// when the width is 7 the last bit is non used.
						// Then the result masked with 11111110 -> 0xfe
						mask := byte(0xfe)
						switch op.Operator {
						case PixClr:
							for i, bt := range dest.Data {
								assert.Equal(t, combinePartial(tocheck[i], 0x00, mask), bt)
							}
						case PixSet:
							for i, bt := range dest.Data {
								assert.Equal(t, combinePartial(tocheck[i], 0xff, mask), bt)
							}
						case PixNotDst:
							for i, bt := range dest.Data {
								assert.Equal(t, combinePartial(tocheck[i], ^tocheck[i], mask), bt)
							}
						}
					})
				}
			})

			t.Run("WidthLessEqualTo8", func(t *testing.T) {
				for _, op := range operators {
					// check all width sizes from 1 - 7.
					for width := 1; width <= 8; width++ {
						t.Run(op.Name+strconv.Itoa(width), func(t *testing.T) {
							data := []byte{0xff, 0xfd, 0xfa, 0xf3}
							// clear the data with the width bits in mind
							// i.e. data with width 1 should be masked by '10000000'.
							// as the bitmap requires that the row must be at least 1 byte size.
							// with the padding with non used bits.
							for i := 0; i < len(data); i++ {
								data[i] &= lmaskByte[width]
							}

							tocheck := make([]byte, 4)
							copy(tocheck, data)

							dest, err := NewWithData(width, 4, data)
							require.NoError(t, err)

							require.NotPanics(t, func() { err = dest.RasterOperation(0, 0, dest.Width, dest.Height, op.Operator, nil, 0, 0) })
							require.NoError(t, err)

							switch op.Operator {
							case PixClr:
								for _, bt := range dest.Data {
									assert.Equal(t, byte(0x00), bt)
								}
							case PixSet:
								for _, bt := range dest.Data {
									assert.Equal(t, byte(0xff)&lmaskByte[width], bt)
								}
							case PixNotDst:
								for i, bt := range dest.Data {
									assert.Equal(t, ^tocheck[i]&lmaskByte[width], bt)
								}
							}
						})
					}
				}
			})
		})

		t.Run("General", func(t *testing.T) {
			t.Run("DoublyPartial", func(t *testing.T) {
				for _, op := range operators {
					t.Run(op.Name, func(t *testing.T) {
						data := []byte{0xff, 0xfd, 0xfa, 0xf3, 0xac, 0xbc, 0xdc, 0xaf}
						tocheck := make([]byte, 8)
						copy(tocheck, data)

						dest, err := NewWithData(8, 8, data)
						require.NoError(t, err)

						err = dest.RasterOperation(2, 0, dest.Width-4, dest.Height, op.Operator, nil, 0, 0)
						require.NoError(t, err)

						switch op.Operator {
						case PixClr:
							for i, bt := range dest.Data {
								assert.Equal(t, combinePartial(data[i], 0x0, 0x3c), bt)
							}
						case PixSet:
							for i, bt := range dest.Data {
								assert.Equal(t, combinePartial(data[i], 0xff, 0x3c), bt)
							}
						case PixNotDst:
							for i, bt := range dest.Data {
								assert.Equal(t, combinePartial(data[i], ^tocheck[i], 0x3c), bt)
							}
						}
					})
				}
			})
			t.Run("Shifted1", func(t *testing.T) {
				for _, op := range operators {
					t.Run(op.Name, func(t *testing.T) {
						data := []byte{0xff, 0xfd, 0xfa, 0xf3, 0xac, 0xbc, 0xdc, 0xaf, 0xff, 0xfd, 0xfa, 0xf3, 0xac, 0xbc, 0xdc, 0xaf}
						tocheck := make([]byte, 16)
						copy(tocheck, data)

						dest, err := NewWithData(9, 8, data)
						require.NoError(t, err)

						err = dest.RasterOperation(1, 0, dest.Width, dest.Height, op.Operator, nil, 0, 0)
						require.NoError(t, err)

						switch op.Operator {
						case PixClr:
							for i, bt := range dest.Data {
								if i%2 == 0 {
									assert.Equal(t, combinePartial(data[i], 0x0, rmaskByte[7]), bt)
								} else {
									assert.Equal(t, combinePartial(data[i], 0x0, lmaskByte[1]), bt)
								}
							}
						case PixSet:
							for i, bt := range dest.Data {
								if i%2 == 0 {
									assert.Equal(t, combinePartial(data[i], 0xff, rmaskByte[7]), bt)
								} else {
									assert.Equal(t, combinePartial(data[i], 0xff, lmaskByte[1]), bt)
								}
							}
						case PixNotDst:
							for i, bt := range dest.Data {
								if i%2 == 0 {
									assert.Equal(t, combinePartial(data[i], ^tocheck[i], rmaskByte[7]), bt)
								} else {
									assert.Equal(t, combinePartial(data[i], ^tocheck[i], lmaskByte[1]), bt)
								}
							}
						}
					})
				}
			})
			t.Run("DoublyPartial", func(t *testing.T) {
				for _, op := range operators {
					t.Run(op.Name, func(t *testing.T) {
						data := []byte{0xff, 0xfd, 0xfa, 0xf3, 0xac, 0xbc, 0xdc, 0xaf, 0xff, 0xfd, 0xfa, 0xf3, 0xac, 0xbc, 0xdc, 0xaf, 0xff, 0xfd, 0xfa, 0xf3, 0xac, 0xbc, 0xdc, 0xaf}
						tocheck := make([]byte, len(data))
						copy(tocheck, data)

						dest, err := NewWithData(18, 8, data)
						require.NoError(t, err)

						err = dest.RasterOperation(1, 0, dest.Width-1, dest.Height, op.Operator, nil, 0, 0)
						require.NoError(t, err)

						switch op.Operator {
						case PixClr:
							for i, bt := range dest.Data {
								ix := i % 3
								switch ix {
								case 0:
									assert.Equal(t, combinePartial(data[i], 0x0, rmaskByte[7]), bt)
								case 1:
									assert.Equal(t, data[i], byte(0x0))
								case 2:
									assert.Equal(t, combinePartial(data[i], 0x0, lmaskByte[2]), bt)
								}
							}
						case PixSet:
							for i, bt := range dest.Data {
								ix := i % 3
								switch ix {
								case 0:
									assert.Equal(t, combinePartial(data[i], 0xff, rmaskByte[7]), bt)
								case 1:
									assert.Equal(t, data[i], byte(0xff))
								case 2:
									assert.Equal(t, combinePartial(data[i], 0xff, lmaskByte[2]), bt)
								}
							}
						case PixNotDst:
							for i, bt := range dest.Data {
								ix := i % 3
								switch ix {
								case 0:
									assert.Equal(t, combinePartial(data[i], ^tocheck[i], rmaskByte[7]), bt)
								case 1:
									assert.Equal(t, data[i], ^tocheck[i])
								case 2:
									assert.Equal(t, combinePartial(data[i], ^tocheck[i], lmaskByte[2]), bt)
								}
							}
						}
					})
				}
			})
		})
	})

	t.Run("Low", func(t *testing.T) {
		t.Run("NilSrc", func(t *testing.T) {
			dest := New(5, 5)
			err := RasterOperation(dest, 0, 0, 0, 0, PixSrc, nil, 0, 0)
			require.Error(t, err)
		})
	})
}
