/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/common"
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
			// !(dx == 0 && sx == 0) && (sx % 7 != dx % 7)
			t.Run("SrcNotAligned", func(t *testing.T) {
				// sx%7 != 0 and dx%7 == 0 -> sHang has some value and dHang has some value
				t.Run("GreaterSrcHang", func(t *testing.T) {
					for _, op := range operators {
						common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
						t.Run(op.Name, func(t *testing.T) {
							rd := rand.New(rand.NewSource(123))
							dest := New(18, 12)

							// generate random bytes for dest
							i, err := rd.Read(dest.Data)
							require.NoError(t, err)
							assert.Equal(t, len(dest.Data), i)

							// clear unnecessary bits - the bitmap data have 0th bits at the end of the row.
							for i := 0; i < len(dest.Data); i++ {
								if i%3 == 2 {
									dest.Data[i] &= lmaskByte[2]
								}
							}

							// create a copy of the dest data for testing purpose
							destDataCopy := make([]byte, len(dest.Data))
							copy(destDataCopy, dest.Data)

							src := New(20, 16)
							// generate random bytes for source
							i, err = rd.Read(src.Data)
							require.NoError(t, err)
							assert.Equal(t, len(src.Data), i)

							for i := 0; i < len(src.Data); i++ {
								if i%3 == 2 {
									src.Data[i] &= lmaskByte[4]
								}
							}

							// do the raster operations
							err = RasterOperation(dest, 0, 0, dest.Width, dest.Height, op.Operator, src, 2, 0)
							require.NoError(t, err)

							for i := 0; i < len(dest.Data); i++ {
								var shouldBe, srcByte byte
								mask := byte(0xff)
								// example src.Data
								// 11011000 01101100 11010000 01101001 11010111 11010000
								// in example with PixSrc should be
								// 01100001 10110011 01000000 10100111 01011111 01000000
								switch i % 3 {
								case 0, 1:
									// example srcData
									// src.Data[i]: 11011000, src.Data[i+1] = 01101100
									// the result would be a 	11011000 << 2 = 01100000
									// in union with 			01101100 >> 6 = 00000001
									//											01100001
									// this means byte values are shifted by 2
									srcByte = src.Data[i]<<2 | src.Data[i+1]>>6
								case 2:
									// example srcData
									// src.Data[i]: 	11010000 << 2 = 01100000
									srcByte = src.Data[i] << 2
									mask = lmaskByte[2]
								}

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
								assert.Equal(t, shouldBe, dest.Data[i], "i: %d shouldBe: %08b is: %08b", i, shouldBe, dest.Data[i])
							}
						})
					}
				})

				t.Run("GreaterDstHang", func(t *testing.T) {
					for _, op := range operators {
						common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
						t.Run(op.Name, func(t *testing.T) {
							rd := rand.New(rand.NewSource(123))
							dest := New(18, 12)

							// generate random bytes for dest
							i, err := rd.Read(dest.Data)
							require.NoError(t, err)
							assert.Equal(t, len(dest.Data), i)

							// clear unnecessary bits - the bitmap data have 0th bits at the end of the row.
							for i := 0; i < len(dest.Data); i++ {
								if i%3 == 2 {
									dest.Data[i] &= lmaskByte[2]
								}
							}

							// create a copy of the dest data for testing purpose
							destDataCopy := make([]byte, len(dest.Data))
							copy(destDataCopy, dest.Data)

							src := New(20, 16)
							// generate random bytes for source
							i, err = rd.Read(src.Data)
							require.NoError(t, err)
							assert.Equal(t, len(src.Data), i)

							for i := 0; i < len(src.Data); i++ {
								if i%3 == 2 {
									src.Data[i] &= lmaskByte[4]
								}
							}

							// do the raster operations
							err = RasterOperation(dest, 2, 0, dest.Width-2, dest.Height, op.Operator, src, 0, 0)
							require.NoError(t, err)

							for i := 0; i < len(dest.Data); i++ {
								var shouldBe, srcByte byte
								mask := byte(0x00)
								// example src.Data
								// 11011000 01101100 11010000 01101001 11010111 11010000
								// example dest.Data
								// 10100101 10001011 01000000 01011010 11110101 01000000
								// in example with PixSrc should be where dest is shifted right by 2
								// 10110110 00011011 00000000 01011010 01110101 11000000
								switch i % 3 {
								case 0:
									// example srcData
									// src.Data[0]>>2 		=	11011000 >> 2 		= 	00110110
									// dst.Data[0]&11000000	=	10100101 & 11000000 = 	10000000
									// the result would be a union						10110110
									// this means byte values are shifted by 2
									srcByte = src.Data[i] >> 2
									mask = rmaskByte[6]
								case 1:
									// example srcData
									// src.Data[i]: 	11010000 << 2 = 01100000
									srcByte = src.Data[i-1]<<6 | src.Data[i]>>2
									mask = 0xff
								case 2:
									// example srcData
									// src.Data[i]: 	11010000 << 2 = 01100000
									srcByte = src.Data[i-1]<<6 | src.Data[i]>>2
									mask = lmaskByte[2]
								}

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
								assert.Equal(t, shouldBe, dest.Data[i], "i: %d shouldBe: %08b is: %08b", i, shouldBe, dest.Data[i])
							}
						})
					}
				})
			})

		})
	})

	t.Run("Simple", func(t *testing.T) {
		// having two bitmaps:
		// Dest - width: 18, height: 8 with data:
		//
		// 11111111	11111111 11000000
		// 10000000 00000000 01000000
		// 10000100 01001111 01000000
		// 10000010 10001001 01000000
		// 10000001 00001001 01000000
		// 10000001 00001111 01000000
		// 10000000 00000000 01000000
		// 11111111 11111111 11000000
		dData := []byte{
			0xFF, 0xFF, 0xC0,
			0x80, 0x00, 0x40,
			0x84, 0x4F, 0x40,
			0x82, 0x89, 0x40,
			0x81, 0x09, 0x40,
			0x81, 0x0F, 0x40,
			0x80, 0x00, 0x40,
			0xFF, 0xFF, 0xC0,
		}

		// Source: width 8, height: 4 with data:
		//
		// 01111110
		// 01011001
		// 01010110
		// 01100010
		sData := []byte{0x7E, 0x59, 0x56, 0x62}

		dst, err := NewWithData(18, 8, dData)
		require.NoError(t, err)

		src, err := NewWithData(8, 4, sData)
		require.NoError(t, err)

		t.Run("Dst", func(t *testing.T) {
			t.Run("Clr", func(t *testing.T) {
				dst := dst.Copy()

				err := dst.RasterOperation(7, 2, 9, 4, PixClr, nil, 0, 0)
				require.NoError(t, err)

				// 11111111	11111111 11000000
				// 10000000 00000000 01000000
				// 10000100 00000000 01000000
				// 10000010 00000000 01000000
				// 10000000 00000000 01000000
				// 10000000 00000000 01000000
				// 10000000 00000000 01000000
				// 11111111 11111111 11000000
				assert.Equal(t, dst.Data, []byte{
					0xFF, 0xFF, 0xC0,
					0x80, 0x00, 0x40,
					0x84, 0x00, 0x40,
					0x82, 0x00, 0x40,
					0x80, 0x00, 0x40,
					0x80, 0x00, 0x40,
					0x80, 0x00, 0x40,
					0xFF, 0xFF, 0xC0,
				})
			})

			t.Run("Set", func(t *testing.T) {
				dst := dst.Copy()

				err := dst.RasterOperation(2, 2, 14, 4, PixSet, nil, 0, 0)
				require.NoError(t, err)

				// should be a:
				// 11111111	11111111 11000000
				// 10000000 00000000 01000000
				// 10111111 11111111 01000000
				// 10111111 11111111 01000000
				// 10111111 11111111 01000000
				// 10111111 11111111 01000000
				// 10000000 00000000 01000000
				// 11111111 11111111 11000000
				assert.Equal(t, dst.Data, []byte{
					0xFF, 0xFF, 0xC0,
					0x80, 0x00, 0x40,
					0xBF, 0xFF, 0x40,
					0xBF, 0xFF, 0x40,
					0xBF, 0xFF, 0x40,
					0xBF, 0xFF, 0x40,
					0x80, 0x00, 0x40,
					0xFF, 0xFF, 0xC0,
				})
			})

			t.Run("NotDst", func(t *testing.T) {
				dst := dst.Copy()

				err := dst.RasterOperation(5, 3, 11, 4, PixNotDst, nil, 0, 0)
				require.NoError(t, err)

				// 11111111	11111111 11000000
				// 10000000 00000000 01000000
				// 10000100 01001111 01000000
				// 10000101 01110110 01000000
				// 10000110 11110110 01000000
				// 10000110 11110000 01000000
				// 10000111 11111111 01000000
				// 11111111 11111111 11000000
				assert.Equal(t, dst.Data, []byte{
					0xFF, 0xFF, 0xC0,
					0x80, 0x00, 0x40,
					0x84, 0x4F, 0x40,
					0x85, 0x76, 0x40,
					0x86, 0xF6, 0x40,
					0x86, 0xF0, 0x40,
					0x87, 0xFF, 0x40,
					0xFF, 0xFF, 0xC0,
				})
			})
		})

		t.Run("WithSrc", func(t *testing.T) {
			t.Run("Src", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixSrc, src, 0, 0)
					assert.NoError(t, err)

					// 01111110
					// 01011001
					// 01010110
					// 01100010

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 01111110 01000000
					// 10000010 01011001 01000000
					// 10000001 01010110 01000000
					// 10000001 01100010 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x7E, 0x40,
						0x82, 0x59, 0x40,
						0x81, 0x56, 0x40,
						0x81, 0x62, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})
				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixSrc, src, 0, 0)
					assert.NoError(t, err)

					// Dst:		00010011 10100010 01000010 01000011
					// Src:		01111110 01011001 01010110 01100010

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000101 11111011 01000000
					// 10000001 01100101 01000000
					// 10000001 01011001 01000000
					// 10000001 10001011 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x85, 0xFB, 0x40,
						0x81, 0x65, 0x40,
						0x81, 0x59, 0x40,
						0x81, 0x8B, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

			})

			t.Run("NotSrc", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixNotSrc, src, 0, 0)
					assert.NoError(t, err)

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 10000001 01000000
					// 10000010 10100110 01000000
					// 10000001 10101001 01000000
					// 10000001 10011101 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x81, 0x40,
						0x82, 0xA6, 0x40,
						0x81, 0xA9, 0x40,
						0x81, 0x9D, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixNotSrc, src, 0, 0)
					assert.NoError(t, err)

					// Dst:		00010011 10100010 01000010 01000011
					// NotSrc:	10000001 10100110 10101001 10011101

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000110 00000111 01000000
					// 10000010 10011001 01000000
					// 10000010 10100101 01000000
					// 10000010 01110111 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x86, 0x07, 0x40,
						0x82, 0x99, 0x40,
						0x82, 0xA5, 0x40,
						0x82, 0x77, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})
			})

			t.Run("SrcAndDst", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixSrcAndDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 		01001111 10001001 00001001 00001111
					// Src: 		01111110 01011001 01010110 01100010
					// SrcAndDst: 	01001110 00001001 00000000 00000010
					//					0x4E     0x09     0x00     0x02

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 01001110 01000000
					// 10000010 00001001 01000000
					// 10000001 00000000 01000000
					// 10000001 00000010 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x4E, 0x40,
						0x82, 0x09, 0x40,
						0x81, 0x00, 0x40,
						0x81, 0x02, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})
				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixSrcAndDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 		00010011 10100010 01000010 01000011
					// Src:			01111110 01011001 01010110 01100010
					// SrcAndDst:	00010010 00000000 01000010 01000010

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 01001011 01000000
					// 10000000 00000001 01000000
					// 10000001 00001001 01000000
					// 10000001 00001011 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x4B, 0x40,
						0x80, 0x01, 0x40,
						0x81, 0x09, 0x40,
						0x81, 0x0B, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("SrcOrDst", func(t *testing.T) {
				t.Run("Algined", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixSrcOrDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 		01001111 10001001 00001001 00001111
					// Src: 		01111110 01011001 01010110 01100010
					// SrcOrDst: 	01111111 11011001 01011111 01101111
					//					0xEF     0xD9     0x5F     0x6F

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 01111111 01000000
					// 10000010 11011001 01000000
					// 10000001 01011111 01000000
					// 10000001 01101110 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x7F, 0x40,
						0x82, 0xD9, 0x40,
						0x81, 0x5F, 0x40,
						0x81, 0x6F, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})
				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixSrcOrDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 		00010011 10100010 01000010 01000011
					// Src: 		01111110 01011001 01010110 01100010
					// SrcOrDst: 	01111111 11111011 01010110 01100011

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000101 11111111 01000000
					// 10000011 11101101 01000000
					// 10000001 01011001 01000000
					// 10000001 10001111 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x85, 0xFF, 0x40,
						0x83, 0xED, 0x40,
						0x81, 0x59, 0x40,
						0x81, 0x8F, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("SrcXorDst", func(t *testing.T) {
				t.Run("Algined", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixSrcXorDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 		01001111 10001001 00001001 00001111
					// Src: 		01111110 01011001 01010110 01100010
					// SrcXorDst: 	00110001 11010000 01011111 01101101
					//					0x31     0xD0     0x5F     0x6D

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 00110001 01000000
					// 10000010 11010000 01000000
					// 10000001 01011111 01000000
					// 10000001 01101101 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x31, 0x40,
						0x82, 0xD0, 0x40,
						0x81, 0x5F, 0x40,
						0x81, 0x6D, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixSrcXorDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 		00010011 10100010 01000010 01000011
					// Src: 		01111110 01011001 01010110 01100010
					// SrcXorDst:	01101101 11111011 00010100 00100001

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000101 10110111 01000000
					// 10000011 11101101 01000000
					// 10000000 01010001 01000000
					// 10000000 10000111 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x85, 0xB7, 0x40,
						0x83, 0xED, 0x40,
						0x80, 0x51, 0x40,
						0x80, 0x87, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("SrcAndNotDst", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixSrcAndNotDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 			01001111 10001001 00001001 00001111
					// ^Dst:			10110000 01110110 11110110 11110000
					// Src: 			01111110 01011001 01010110 01100010
					// SrcAndNotDst: 	00110000 01010000 01010110 01100000
					//						0x30     0x50     0x56     0x60

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 00110000 01000000
					// 10000010 01010000 01000000
					// 10000001 01010110 01000000
					// 10000001 01100000 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x30, 0x40,
						0x82, 0x50, 0x40,
						0x81, 0x56, 0x40,
						0x81, 0x60, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})
				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixSrcAndNotDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 			00010011 10100010 01000010 01000011
					// ^Dst: 			11101100 01011101 10111101 10111100
					// Src: 			01111110 01011001 01010110 01100010
					// SrcAndNotDst:	01101100 01011001 00010100 00100000

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000101 10110011 01000000
					// 10000001 01100101 01000000
					// 10000000 01010001 01000000
					// 10000000 10000011 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x85, 0xB3, 0x40,
						0x81, 0x65, 0x40,
						0x80, 0x51, 0x40,
						0x80, 0x83, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("SrcOrNotDst", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixSrcOrNotDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 			01001111 10001001 00001001 00001111
					// ^Dst:			10110000 01110110 11110110 11110000
					// Src: 			01111110 01011001 01010110 01100010
					// SrcOrNotDst: 	11111110 01111111 11110110 11110010
					//						0xFE     0x7F     0xF6     0xF2

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 00110000 01000000
					// 10000010 01010000 01000000
					// 10000001 01010110 01000000
					// 10000001 01100000 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0xFE, 0x40,
						0x82, 0x7F, 0x40,
						0x81, 0xF6, 0x40,
						0x81, 0xF2, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixSrcOrNotDst, src, 0, 0)
					assert.NoError(t, err)

					// Dst: 			00010011 10100010 01000010 01000011
					// ^Dst: 			11101100 01011101 10111101 10111100
					// Src: 			01111110 01011001 01010110 01100010
					// SrcOrNotDst:		11111110 01011101 11111111 11111110

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000111 11111011 01000000
					// 10000001 01110101 01000000
					// 10000011 11111101 01000000
					// 10000011 11111011 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x87, 0xFB, 0x40,
						0x81, 0x75, 0x40,
						0x83, 0xFD, 0x40,
						0x83, 0xFB, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("NotSrcAndDst", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixNotSrcAndDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// ^Src:			10000001 10100110 10101001 10011101
					// Dst: 			01001111 10001001 00001001 00001111
					// NotSrcAndDst: 	00000001 10000000 00001001 00001101
					//						0x01     0x80     0x09     0x0D

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 00000001 01000000
					// 10000010 10000000 01000000
					// 10000001 00001001 01000000
					// 10000001 00001101 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x01, 0x40,
						0x82, 0x80, 0x40,
						0x81, 0x09, 0x40,
						0x81, 0x0D, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixNotSrcAndDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// ^Src:			10000001 10100110 10101001 10011101
					// Dst: 			00010011 10100010 01000010 01000011
					// NotSrcAndDst:	00000001 10100010 00000000 00000001

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 00000111 01000000
					// 10000010 10001001 01000000
					// 10000000 00000001 01000000
					// 10000000 00000111 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x07, 0x40,
						0x82, 0x89, 0x40,
						0x80, 0x01, 0x40,
						0x80, 0x07, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("NotSrcOrDst", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixNotSrcOrDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// ^Src:			10000001 10100110 10101001 10011101
					// Dst: 			01001111 10001001 00001001 00001111
					// NotSrcOrDst: 	11001111 10101111 10101001 10011111
					//						0xCF     0xAF     0xA9     0x9F

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 11001111 01000000
					// 10000010 10101111 01000000
					// 10000001 10101001 01000000
					// 10000001 10011111 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0xCF, 0x40,
						0x82, 0xAF, 0x40,
						0x81, 0xA9, 0x40,
						0x81, 0x9F, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixNotSrcOrDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// ^Src:			10000001 10100110 10101001 10011101
					// Dst: 			00010011 10100010 01000010 01000011
					// NotSrcOrDst:		10010011 10100110 11101011 11011111

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000110 01001111 01000000
					// 10000010 10011001 01000000
					// 10000011 10101101 01000000
					// 10000011 01111111 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x86, 0x4F, 0x40,
						0x82, 0x99, 0x40,
						0x83, 0xAD, 0x40,
						0x83, 0x7F, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("Not(SrcOrDst)", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixNotPixSrcOrDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// Dst: 			01001111 10001001 00001001 00001111

					// SrcOrDst: 		01111111 11011001 01011111 01101111
					// Not(SrcOrDst): 	10000000 00100110 10100000 10010000
					//						0x80     0x26     0xA0     0x90

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 10000000 01000000
					// 10000010 00100110 01000000
					// 10000001 10100000 01000000
					// 10000001 10010000 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0x80, 0x40,
						0x82, 0x26, 0x40,
						0x81, 0xA0, 0x40,
						0x81, 0x90, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixNotPixSrcOrDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// Dst: 			00010011 10100010 01000010 01000011
					// SrcOrDst:		01111111 11111011 01010110 01100011
					// Not(SrcOrDst):	10000000 00000100 10101001 10011100

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000110 00000011 01000000
					// 10000000 00010001 01000000
					// 10000010 10100101 01000000
					// 10000010 01110011 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x86, 0x03, 0x40,
						0x80, 0x11, 0x40,
						0x82, 0xA5, 0x40,
						0x82, 0x73, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("Not(SrcAndDst)", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixNotPixSrcAndDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// ^Src:			10000001 10100110 10101001 10011101

					// SrcAndDst: 		01001110 00001001 00000000 00000010
					// Not(SrcAndDst): 	10110001 11110110 11111111 11111101
					//						0xB1     0xF6     0xFF     0xFD

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 10110001 01000000
					// 10000010 11110110 01000000
					// 10000001 11111111 01000000
					// 10000001 11111101 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0xB1, 0x40,
						0x82, 0xF6, 0x40,
						0x81, 0xFF, 0x40,
						0x81, 0xFD, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixNotPixSrcAndDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// Dst: 			00010011 10100010 01000010 01000011
					// SrcAndDst:		00010010 00000000 01000010 01000010
					// Not(SrcOrDst):	11101101 11111111 10111101 10111101

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000111 10110111 01000000
					// 10000011 11111101 01000000
					// 10000010 11110101 01000000
					// 10000010 11110111 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x87, 0xB7, 0x40,
						0x83, 0xFD, 0x40,
						0x82, 0xF5, 0x40,
						0x82, 0xF7, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})

			t.Run("Not(SrcXorDst)", func(t *testing.T) {
				t.Run("Aligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(8, 2, 8, 4, PixNotPixSrcXorDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// ^Src:			10000001 10100110 10101001 10011101

					// SrcXorDst: 		00110001 11010000 01011111 01101101
					// Not(SrcAndDst): 	11001110 00101111 10100000 10010010
					//						0xCE     0x2F     0xA0     0x92

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000100 11001110 01000000
					// 10000010 00101111 01000000
					// 10000001 10100000 01000000
					// 10000001 10010010 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, dst.Data, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x84, 0xCE, 0x40,
						0x82, 0x2F, 0x40,
						0x81, 0xA0, 0x40,
						0x81, 0x92, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					})
				})

				t.Run("NotAligned", func(t *testing.T) {
					dst := dst.Copy()

					err = dst.RasterOperation(6, 2, 8, 4, PixNotPixSrcXorDst, src, 0, 0)
					assert.NoError(t, err)

					// Src: 			01111110 01011001 01010110 01100010
					// Dst: 			00010011 10100010 01000010 01000011
					// SrcXorDst:		01101101 11111011 00010100 00100001
					// Not(SrcOrDst):	10010010 00000100 11101011 11011110

					// 11111111	11111111 11000000
					// 10000000 00000000 01000000
					// 10000110 01001011 01000000
					// 10000000 00010001 01000000
					// 10000011 10101101 01000000
					// 10000011 01111011 01000000
					// 10000000 00000000 01000000
					// 11111111 11111111 11000000
					assert.Equal(t, []byte{
						0xFF, 0xFF, 0xC0,
						0x80, 0x00, 0x40,
						0x86, 0x4B, 0x40,
						0x80, 0x11, 0x40,
						0x83, 0xAD, 0x40,
						0x83, 0x7B, 0x40,
						0x80, 0x00, 0x40,
						0xFF, 0xFF, 0xC0,
					}, dst.Data)
				})
			})
		})
	})
}
