package huffman

import (
	"github.com/unidoc/unidoc/pdf/internal/jbig2/reader"
)

const (
	HTLOW = 0xfffffffd
	HTOOB = 0xfffffffe
	EOT   = int(int32(-1))
)

var (
	TableA [][]int = [][]int{{0, 1, 4, 0x000}, {16, 2, 8, 0x002}, {272, 3, 16, 0x006}, {65808, 3, 32, 0x007}, {0, 0, EOT, 0}}
	TableB [][]int = [][]int{{0, 1, 0, 0x000}, {1, 2, 0, 0x002}, {2, 3, 0, 0x006}, {3, 4, 3, 0x00e}, {11, 5, 6, 0x01e}, {75, 6, 32, 0x03e}, {0, 6, HTOOB, 0x03f}, {0, 0, EOT, 0}}
	TableC [][]int = [][]int{{0, 1, 0, 0x000}, {1, 2, 0, 0x002}, {2, 3, 0, 0x006}, {3, 4, 3, 0x00e}, {11, 5, 6, 0x01e}, {0, 6, HTOOB, 0x03e}, {75, 7, 32, 0x0fe}, {-256, 8, 8, 0x0fe}, {-257, 8, HTLOW, 0x0ff}, {0, 0, EOT, 0}}
	TableD [][]int = [][]int{{1, 1, 0, 0x000}, {2, 2, 0, 0x002}, {3, 3, 0, 0x006}, {4, 4, 3, 0x00e}, {12, 5, 6, 0x01e}, {76, 5, 32, 0x01f}, {0, 0, EOT, 0}}
	TableE [][]int = [][]int{{1, 1, 0, 0x000}, {2, 2, 0, 0x002}, {3, 3, 0, 0x006}, {4, 4, 3, 0x00e}, {12, 5, 6, 0x01e}, {76, 6, 32, 0x03e}, {-255, 7, 8, 0x07e}, {-256, 7, HTLOW, 0x07f}, {0, 0, EOT, 0}}
	TableF [][]int = [][]int{{0, 2, 7, 0x000}, {128, 3, 7, 0x002}, {256, 3, 8, 0x003}, {-1024, 4, 9, 0x008}, {-512, 4, 8, 0x009}, {-256, 4, 7, 0x00a}, {-32, 4, 5, 0x00b}, {512, 4, 9, 0x00c}, {1024, 4, 10, 0x00d}, {-2048, 5, 10, 0x01c}, {-128, 5, 6, 0x01d}, {-64, 5, 5, 0x01e}, {-2049, 6, HTLOW, 0x03e}, {2048, 6, 32, 0x03f}, {0, 0, EOT, 0}}
	TableG [][]int = [][]int{{-512, 3, 8, 0x000}, {256, 3, 8, 0x001}, {512, 3, 9, 0x002}, {1024, 3, 10, 0x003}, {-1024, 4, 9, 0x008}, {-256, 4, 7, 0x009}, {-32, 4, 5, 0x00a}, {0, 4, 5, 0x00b}, {128, 4, 7, 0x00c}, {-128, 5, 6, 0x01a}, {-64, 5, 5, 0x01b}, {32, 5, 5, 0x01c}, {64, 5, 6, 0x01d}, {-1025, 5, HTLOW, 0x01e}, {2048, 5, 32, 0x01f}, {0, 0, EOT, 0}}
	TableH [][]int = [][]int{{0, 2, 1, 0x000}, {0, 2, HTOOB, 0x001}, {4, 3, 4, 0x004}, {-1, 4, 0, 0x00a}, {22, 4, 4, 0x00b}, {38, 4, 5, 0x00c}, {2, 5, 0, 0x01a}, {70, 5, 6, 0x01b}, {134, 5, 7, 0x01c}, {3, 6, 0, 0x03a}, {20, 6, 1, 0x03b}, {262, 6, 7, 0x03c}, {646, 6, 10, 0x03d}, {-2, 7, 0, 0x07c}, {390, 7, 8, 0x07d}, {-15, 8, 3, 0x0fc}, {-5, 8, 1, 0x0fd}, {-7, 9, 1, 0x1fc}, {-3, 9, 0, 0x1fd}, {-16, 9, HTLOW, 0x1fe}, {1670, 9, 32, 0x1ff}, {0, 0, EOT, 0}}
	TableI [][]int = [][]int{{0, 2, HTOOB, 0x000}, {-1, 3, 1, 0x002}, {1, 3, 1, 0x003}, {7, 3, 5, 0x004}, {-3, 4, 1, 0x00a}, {43, 4, 5, 0x00b}, {75, 4, 6, 0x00c}, {3, 5, 1, 0x01a}, {139, 5, 7, 0x01b}, {267, 5, 8, 0x01c}, {5, 6, 1, 0x03a}, {39, 6, 2, 0x03b}, {523, 6, 8, 0x03c}, {1291, 6, 11, 0x03d}, {-5, 7, 1, 0x07c}, {779, 7, 9, 0x07d}, {-31, 8, 4, 0x0fc}, {-11, 8, 2, 0x0fd}, {-15, 9, 2, 0x1fc}, {-7, 9, 1, 0x1fd}, {-32, 9, HTLOW, 0x1fe}, {3339, 9, 32, 0x1ff}, {0, 0, EOT, 0}}
	TableJ [][]int = [][]int{{-2, 2, 2, 0x000}, {6, 2, 6, 0x001}, {0, 2, HTOOB, 0x002}, {-3, 5, 0, 0x018}, {2, 5, 0, 0x019}, {70, 5, 5, 0x01a}, {3, 6, 0, 0x036}, {102, 6, 5, 0x037}, {134, 6, 6, 0x038}, {198, 6, 7, 0x039}, {326, 6, 8, 0x03a}, {582, 6, 9, 0x03b}, {1094, 6, 10, 0x03c}, {-21, 7, 4, 0x07a}, {-4, 7, 0, 0x07b}, {4, 7, 0, 0x07c}, {2118, 7, 11, 0x07d}, {-5, 8, 0, 0x0fc}, {5, 8, 0, 0x0fd}, {-22, 8, HTLOW, 0x0fe}, {4166, 8, 32, 0x0ff}, {0, 0, EOT, 0}}
	TableK [][]int = [][]int{{1, 1, 0, 0x000}, {2, 2, 1, 0x002}, {4, 4, 0, 0x00c}, {5, 4, 1, 0x00d}, {7, 5, 1, 0x01c}, {9, 5, 2, 0x01d}, {13, 6, 2, 0x03c}, {17, 7, 2, 0x07a}, {21, 7, 3, 0x07b}, {29, 7, 4, 0x07c}, {45, 7, 5, 0x07d}, {77, 7, 6, 0x07e}, {141, 7, 32, 0x07f}, {0, 0, EOT, 0}}
	TableL [][]int = [][]int{{1, 1, 0, 0x000}, {2, 2, 0, 0x002}, {3, 3, 1, 0x006}, {5, 5, 0, 0x01c}, {6, 5, 1, 0x01d}, {8, 6, 1, 0x03c}, {10, 7, 0, 0x07a}, {11, 7, 1, 0x07b}, {13, 7, 2, 0x07c}, {17, 7, 3, 0x07d}, {25, 7, 4, 0x07e}, {41, 8, 5, 0x0fe}, {73, 8, 32, 0x0ff}, {0, 0, EOT, 0}}
	TableM [][]int = [][]int{{1, 1, 0, 0x000}, {2, 3, 0, 0x004}, {7, 3, 3, 0x005}, {3, 4, 0, 0x00c}, {5, 4, 1, 0x00d}, {4, 5, 0, 0x01c}, {15, 6, 1, 0x03a}, {17, 6, 2, 0x03b}, {21, 6, 3, 0x03c}, {29, 6, 4, 0x03d}, {45, 6, 5, 0x03e}, {77, 7, 6, 0x07e}, {141, 7, 32, 0x07f}, {0, 0, EOT, 0}}
	TableN [][]int = [][]int{{0, 1, 0, 0x000}, {-2, 3, 0, 0x004}, {-1, 3, 0, 0x005}, {1, 3, 0, 0x006}, {2, 3, 0, 0x007}, {0, 0, EOT, 0}}
	TableO [][]int = [][]int{{0, 1, 0, 0x000}, {-1, 3, 0, 0x004}, {1, 3, 0, 0x005}, {-2, 4, 0, 0x00c}, {2, 4, 0, 0x00d}, {-4, 5, 1, 0x01c}, {3, 5, 1, 0x01d}, {-8, 6, 2, 0x03c}, {5, 6, 2, 0x03d}, {-24, 7, 4, 0x07c}, {9, 7, 4, 0x07d}, {-25, 7, HTLOW, 0x07e}, {25, 7, 32, 0x07f}, {0, 0, EOT, 0}}
)

type HuffmanDecoder struct {
}

// DecodeInt
func (h *HuffmanDecoder) DecodeInt(r *reader.Reader, table [][]int) (int, bool, error) {

	var prefix, length int

	for i := 0; table[i][2] != EOT; i++ {
		// common.Log.Debug("value i: '%v'", i)
		// common.Log.Debug("table[i][2] == %b", table[i][2])
		for ; length < table[i][1]; length++ {
			bit, err := r.ReadBit()
			if err != nil {
				return 0, false, err
			}
			// common.Log.Debug("Bit: %b", bit)
			prefix = (prefix << 1) | int(bit)
			// common.Log.Debug("Prefix: %b", prefix)
		}

		if prefix == table[i][3] {
			if table[i][2] == HTOOB {
				return -1, false, nil
			}

			var decoded int
			if table[i][2] == HTLOW {
				readBits, err := r.ReadBits(32)
				if err != nil {
					return -1, false, nil
				}

				decoded = table[i][0] - int(readBits)
			} else if table[i][2] > 0 {
				// common.Log.Debug("table[i][2] > 0")
				readBits, err := r.ReadBits(byte(table[i][2]))
				if err != nil {
					return -1, false, nil
				}

				// common.Log.Debug("Value read: %b%03b", prefix, readBits)
				// common.Log.Debug("readBits of length: '%d' and value: %b", table[i][2], readBits)
				decoded = table[i][0] + int(readBits)
			} else {
				// common.Log.Debug("else")
				decoded = table[i][0]
			}
			// common.Log.Debug("table[i][0] = %d", table[i][0])

			return decoded, true, nil
		}
	}
	return 0, false, nil
}

func (h *HuffmanDecoder) BuildTable(table [][]int, length int) ([][]int, error) {
	var i, j, k, prefix int

	var tab []int

	for i = 0; i < length; i++ {
		for j = i; j < length && table[j][1] == 0; j++ {
		}
		if j == length {
			break
		}

		for k = j + 1; k < length; k++ {
			if table[k][1] > 0 && table[k][1] < table[j][1] {
				j = k
			}
		}

		if j != i {
			tab = table[j]

			for k = j; k > i; k-- {
				table[k] = table[k-1]
			}

			table[i] = tab
		}
	}

	table[i] = table[length]

	table[0][3] = 0
	prefix = 1

	for i = 1; table[i][2] != EOT; i++ {
		prefix <<= uint(table[i][1] - table[i-1][1])
		table[i][3] = int(prefix)
		prefix += 1
	}

	return table, nil

}

func New() *HuffmanDecoder {
	return &HuffmanDecoder{}
}
