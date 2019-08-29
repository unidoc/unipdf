/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"github.com/unidoc/unipdf/common"
	"math/rand"
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

		operators := []operatorName{{PixSrc, "Source"}, {PixNotSrc, "NotSource"}, {PixSrcOrDst, "SourceOrDest"}, {PixSrcAndDst, "SourceAndDest"}, {PixSrcXorDst, "SourceXorDest"}, {PixNotSrcOrDst, "NotSourceOrDest"}, {PixNotSrcAndDst, "NotSourceAndDest"}, {PixSrcOrNotDst, "SourceAndNotDest"}, {PixSrcAndNotDst, "SourceAndNotDest"}, {PixNotPixSrcOrDst, "Not(SourceOrDest)"}, {PixNotPixSrcAndDst, "Not(SourceAndDest)"}, {PixNotPixSrcXorDst, "Not(SourceXorDest)"}}

		t.Run("ByteAligned", func(t *testing.T) {
			// ByteAligned raster operation occurs when both the resultant 'dx' and 'sx'
			// are divisible by 7 (full byte).
			for _, op := range operators {
				t.Run(op.Name, func(t *testing.T) {
					dest := New(12, 12)
					// generate random bytes for dest
					i, err := rand.Read(dest.Data)
					require.NoError(t, err)
					assert.Equal(t, len(dest.Data), i)

					// clear unnecessary bits - the bitmap data have 0th bits at the end of the row.
					for i := 0; i < len(dest.Data); i++ {
						if i%2 != 0 {
							dest.Data[i] &= lmaskByte[4]
						}
					}

					// create a copy of the dest data for testing purpose
					destDataCopy := make([]byte, len(dest.Data))
					copy(destDataCopy, dest.Data)

					src := New(14, 16)
					// generate random bytes for source
					i, err = rand.Read(src.Data)
					require.NoError(t, err)
					assert.Equal(t, len(src.Data), i)

					for i := 0; i < len(src.Data); i++ {
						if i%2 != 0 {
							src.Data[i] &= lmaskByte[4]
						}
					}

					// do the raster operations
					err = RasterOperation(dest, 0, 0, dest.Width, dest.Height, op.Operator, src, 0, 0)
					require.NoError(t, err)

					for i, bt := range dest.Data {
						var shouldBe byte
						if i%2 != 0 {
							// take the mask of the dest.Width % 8
							mask := lmaskByte[4]
							switch op.Operator {
							case PixSrc:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i], mask)
							case PixNotSrc:
								shouldBe = combinePartial(destDataCopy[i], ^src.Data[i], mask)
							case PixSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]|destDataCopy[i], mask)
							case PixSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]&destDataCopy[i], mask)
							case PixSrcXorDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]^destDataCopy[i], mask)
							case PixNotSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], (^src.Data[i])|destDataCopy[i], mask)
							case PixNotSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], (^src.Data[i])&destDataCopy[i], mask)
							case PixSrcOrNotDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]|^destDataCopy[i], mask)
							case PixSrcAndNotDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]&^destDataCopy[i], mask)
							case PixNotPixSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] | destDataCopy[i]), mask)
							case PixNotPixSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] & destDataCopy[i]), mask)
							case PixNotPixSrcXorDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] ^ destDataCopy[i]), mask)
							}
						} else {
							switch op.Operator {
							case PixSrc:
								shouldBe = src.Data[i]
							case PixNotSrc:
								shouldBe = (^src.Data[i])
							case PixSrcOrDst:
								shouldBe = src.Data[i] | destDataCopy[i]
							case PixSrcAndDst:
								shouldBe = src.Data[i] & destDataCopy[i]
							case PixSrcXorDst:
								shouldBe = src.Data[i] ^ destDataCopy[i]
							case PixNotSrcOrDst:
								shouldBe = ^src.Data[i] | destDataCopy[i]
							case PixNotSrcAndDst:
								shouldBe = ^src.Data[i] & destDataCopy[i]
							case PixSrcOrNotDst:
								shouldBe = src.Data[i] | ^destDataCopy[i]
							case PixSrcAndNotDst:
								shouldBe = src.Data[i] & ^destDataCopy[i]
							case PixNotPixSrcOrDst:
								shouldBe = ^(src.Data[i] | destDataCopy[i])
							case PixNotPixSrcAndDst:
								shouldBe = ^(src.Data[i] & destDataCopy[i])
							case PixNotPixSrcXorDst:
								shouldBe = ^(src.Data[i] ^ destDataCopy[i])
							}
						}
						assert.Equal(t, shouldBe, bt, "At i: %d, should be: %08b, is: %08b", i, shouldBe, bt)
					}
				})
			}
		})

		t.Run("VAligned", func(t *testing.T) {
			// VAligned raster operations occurs when the 'dx' and 'sx' are not byte divisible
			// !(dx&7==0 && sx&7==0) && (dx&7 == sx&7)
			t.Run("WithFullByte", func(t *testing.T) {
				for _, op := range operators {
					t.Run(op.Name, func(t *testing.T) {
						dest := New(18, 12)
						// generate random bytes for dest
						i, err := rand.Read(dest.Data)
						require.NoError(t, err)
						assert.Equal(t, len(dest.Data), i)

						// clear unnecessary bits - the bitmap data have 0th bits at the end of the row.
						for i := 0; i < len(dest.Data); i++ {
							if i%2 != 0 {
								dest.Data[i] &= lmaskByte[4]
							}
						}

						// create a copy of the dest data for testing purpose
						destDataCopy := make([]byte, len(dest.Data))
						copy(destDataCopy, dest.Data)

						src := New(20, 16)
						// generate random bytes for source
						i, err = rand.Read(src.Data)
						require.NoError(t, err)
						assert.Equal(t, len(src.Data), i)

						for i := 0; i < len(src.Data); i++ {
							if i%2 != 0 {
								src.Data[i] &= lmaskByte[4]
							}
						}

						// do the raster operations
						err = RasterOperation(dest, 2, 0, dest.Width-2, dest.Height, op.Operator, src, 2, 0)
						require.NoError(t, err)

						for i, bt := range dest.Data {
							var shouldBe, mask byte
							switch i % 3 {
							case 0:
								mask = rmaskByte[6]
							case 1:
								mask = 0xff
							default:
								mask = lmaskByte[2]
							}

							switch op.Operator {
							case PixSrc:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i], mask)
							case PixNotSrc:
								shouldBe = combinePartial(destDataCopy[i], ^src.Data[i], mask)
							case PixSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]|destDataCopy[i], mask)
							case PixSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]&destDataCopy[i], mask)
							case PixSrcXorDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]^destDataCopy[i], mask)
							case PixNotSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], (^src.Data[i])|destDataCopy[i], mask)
							case PixNotSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], (^src.Data[i])&destDataCopy[i], mask)
							case PixSrcOrNotDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]|^destDataCopy[i], mask)
							case PixSrcAndNotDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]&^destDataCopy[i], mask)
							case PixNotPixSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] | destDataCopy[i]), mask)
							case PixNotPixSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] & destDataCopy[i]), mask)
							case PixNotPixSrcXorDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] ^ destDataCopy[i]), mask)
							}
							assert.Equal(t, shouldBe, bt, "At i: %d, should be: %08b, is: %08b", i, shouldBe, bt)
						}
					})
				}
			})

			t.Run("DoublyPartial", func(t *testing.T) {
				for _, op := range operators {
					t.Run(op.Name, func(t *testing.T) {
						dest := New(8, 12)
						// generate random bytes for dest
						i, err := rand.Read(dest.Data)
						require.NoError(t, err)
						assert.Equal(t, len(dest.Data), i)

						// create a copy of the dest data for testing purpose
						destDataCopy := make([]byte, len(dest.Data))
						copy(destDataCopy, dest.Data)

						src := New(8, 16)
						// generate random bytes for source
						i, err = rand.Read(src.Data)
						require.NoError(t, err)
						assert.Equal(t, len(src.Data), i)

						// do the raster operations
						err = RasterOperation(dest, 2, 0, dest.Width-4, dest.Height, op.Operator, src, 2, 0)
						require.NoError(t, err)

						mask := rmaskByte[len(rmaskByte)-3] & lmaskByte[len(lmaskByte)-3]
						for i, bt := range dest.Data {
							var shouldBe byte

							switch op.Operator {
							case PixSrc:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i], mask)
							case PixNotSrc:
								shouldBe = combinePartial(destDataCopy[i], ^src.Data[i], mask)
							case PixSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]|destDataCopy[i], mask)
							case PixSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]&destDataCopy[i], mask)
							case PixSrcXorDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]^destDataCopy[i], mask)
							case PixNotSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], (^src.Data[i])|destDataCopy[i], mask)
							case PixNotSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], (^src.Data[i])&destDataCopy[i], mask)
							case PixSrcOrNotDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]|^destDataCopy[i], mask)
							case PixSrcAndNotDst:
								shouldBe = combinePartial(destDataCopy[i], src.Data[i]&^destDataCopy[i], mask)
							case PixNotPixSrcOrDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] | destDataCopy[i]), mask)
							case PixNotPixSrcAndDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] & destDataCopy[i]), mask)
							case PixNotPixSrcXorDst:
								shouldBe = combinePartial(destDataCopy[i], ^(src.Data[i] ^ destDataCopy[i]), mask)
							}
							assert.Equal(t, shouldBe, bt, "At i: %d, should be: %08b, is: %08b", i, shouldBe, bt)
						}
					})
				}
			})
		})

		t.Run("General", func(t *testing.T) {
			// General raster operations are all other raster operations.
			// dx != 0 && sx != 0 && sx != dx
			t.Run("BothNotAligned", func(t *testing.T) {
				// sx%7 != 0 and dx%7 != 0 -> sHang has some value and dHang has some value
				t.Run("GreaterSrcHang", func(t *testing.T) {
					for _, op := range operators {
						t.Run(op.Name, func(t *testing.T) {
							dest := New(18, 12)
							// generate random bytes for dest
							i, err := rand.Read(dest.Data)
							require.NoError(t, err)
							assert.Equal(t, len(dest.Data), i)

							// clear unnecessary bits - the bitmap data have 0th bits at the end of the row.
							for i := 0; i < len(dest.Data); i++ {
								if i%3 != 0 {
									dest.Data[i] &= lmaskByte[6]
								}
							}

							// create a copy of the dest data for testing purpose
							destDataCopy := make([]byte, len(dest.Data))
							copy(destDataCopy, dest.Data)

							src := New(20, 16)
							// generate random bytes for source
							i, err = rand.Read(src.Data)
							require.NoError(t, err)
							assert.Equal(t, len(src.Data), i)

							for i := 0; i < len(src.Data); i++ {
								if i%3 != 0 {
									src.Data[i] &= lmaskByte[4]
								}
							}

							// do the raster operations
							err = RasterOperation(dest, 0, 0, dest.Width, dest.Height, op.Operator, src, 2, 0)
							require.NoError(t, err)

							for i, bt := range dest.Data {
								var shouldBe, srcByte byte
								mask := byte(0xff)
								switch i % 3 {
								case 0:
									srcByte = combinePartial(src.Data[0]<<2 | src.Data[1]>>6, 
								case 1:
									srcByte = combinePartial(src.Data[1]<<2 | src.Data[2]>>6, 
								case 2:
									srcByte = (src.Data[2] << 2) & lmaskByte[2]
									mask = lmaskByte[2]
								}
								t.Logf("i: %d, SourceByte: %08b, 0: %08b, 1: %08b, 2: %08b ", i, srcByte, src.Data[0], src.Data[1], src.Data[2])

								switch op.Operator {
								case PixSrc:
									shouldBe = combinePartial(destDataCopy[i], srcByte, mask)
								case PixNotSrc:
									shouldBe = combinePartial(destDataCopy[i], ^srcByte, mask)
								case PixSrcOrDst:
									shouldBe = combinePartial(destDataCopy[i], srcByte|destDataCopy[i], mask)
								case PixSrcAndDst:
									shouldBe = combinePartial(destDataCopy[i], srcByte&destDataCopy[i], mask)
								case PixSrcXorDst:
									shouldBe = combinePartial(destDataCopy[i], srcByte^destDataCopy[i], mask)
								case PixNotSrcOrDst:
									shouldBe = combinePartial(destDataCopy[i], (^srcByte)|destDataCopy[i], mask)
								case PixNotSrcAndDst:
									shouldBe = combinePartial(destDataCopy[i], (^srcByte)&destDataCopy[i], mask)
								case PixSrcOrNotDst:
									shouldBe = combinePartial(destDataCopy[i], srcByte|^destDataCopy[i], mask)
								case PixSrcAndNotDst:
									shouldBe = combinePartial(destDataCopy[i], srcByte&^destDataCopy[i], mask)
								case PixNotPixSrcOrDst:
									shouldBe = combinePartial(destDataCopy[i], ^(srcByte | destDataCopy[i]), mask)
								case PixNotPixSrcAndDst:
									shouldBe = combinePartial(destDataCopy[i], ^(srcByte & destDataCopy[i]), mask)
								case PixNotPixSrcXorDst:
									shouldBe = combinePartial(destDataCopy[i], ^(srcByte ^ destDataCopy[i]), mask)
								}
								assert.Equal(t, shouldBe, bt, "i: %d shouldBe: %08b is: %08b", i, shouldBe, bt)
							}
						})
					}
				})

				t.Run("GreaterDstHang", func(t *testing.T) {
					for _, op := range operators {
						t.Run(op.Name, func(t *testing.T) {

						})
					}
				})
			})

		})
	})
}
