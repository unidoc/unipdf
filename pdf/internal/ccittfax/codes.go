/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package ccittfax

// map pixel run lengths to the encoded bit sequences
// all the bit sequences for all groups could be found here https://www.itu.int/rec/T-REC-T.6-198811-I/en
var (
	bTerms        map[int]code
	wTerms        map[int]code
	bMakeups      map[int]code
	wMakeups      map[int]code
	commonMakeups map[int]code
	masks         map[int]byte

	eol = code{
		Code:        1 << 4,
		BitsWritten: 12,
	}

	eol1 = code{
		Code:        3 << 3,
		BitsWritten: 13,
	}

	eol0 = code{
		Code:        2 << 3,
		BitsWritten: 13,
	}

	p = code{
		Code:        1 << 12,
		BitsWritten: 4,
	}

	h = code{
		Code:        1 << 13,
		BitsWritten: 3,
	}

	v0 = code{
		Code:        1 << 15,
		BitsWritten: 1,
	}

	v1r = code{
		Code:        3 << 13,
		BitsWritten: 3,
	}

	v2r = code{
		Code:        3 << 10,
		BitsWritten: 6,
	}

	v3r = code{
		Code:        3 << 9,
		BitsWritten: 7,
	}

	v1l = code{
		Code:        2 << 13,
		BitsWritten: 3,
	}

	v2l = code{
		Code:        2 << 10,
		BitsWritten: 6,
	}

	v3l = code{
		Code:        2 << 9,
		BitsWritten: 7,
	}
)

// code describes the encoded bit sequence and the bits actually
// used to represent it within the uint16.
type code struct {
	Code        uint16
	BitsWritten int
}

func init() {
	bTerms = make(map[int]code)

	bTerms[0] = code{
		Code:        13<<8 | 3<<6,
		BitsWritten: 10,
	}

	bTerms[1] = code{
		Code:        2 << (5 + 8),
		BitsWritten: 3,
	}

	bTerms[2] = code{
		Code:        3 << (6 + 8),
		BitsWritten: 2,
	}

	bTerms[3] = code{
		Code:        2 << (6 + 8),
		BitsWritten: 2,
	}

	bTerms[4] = code{
		Code:        3 << (5 + 8),
		BitsWritten: 3,
	}

	bTerms[5] = code{
		Code:        3 << (4 + 8),
		BitsWritten: 4,
	}

	bTerms[6] = code{
		Code:        2 << (4 + 8),
		BitsWritten: 4,
	}

	bTerms[7] = code{
		Code:        3 << (3 + 8),
		BitsWritten: 5,
	}

	bTerms[8] = code{
		Code:        5 << (2 + 8),
		BitsWritten: 6,
	}

	bTerms[9] = code{
		Code:        4 << (2 + 8),
		BitsWritten: 6,
	}

	bTerms[10] = code{
		Code:        4 << (1 + 8),
		BitsWritten: 7,
	}

	bTerms[11] = code{
		Code:        5 << (1 + 8),
		BitsWritten: 7,
	}

	bTerms[12] = code{
		Code:        7 << (1 + 8),
		BitsWritten: 7,
	}

	bTerms[13] = code{
		Code:        4 << 8,
		BitsWritten: 8,
	}

	bTerms[14] = code{
		Code:        7 << 8,
		BitsWritten: 8,
	}

	bTerms[15] = code{
		Code:        12 << 8,
		BitsWritten: 9,
	}

	bTerms[16] = code{
		Code:        5<<8 | 3<<6,
		BitsWritten: 10,
	}

	bTerms[17] = code{
		Code:        6 << 8,
		BitsWritten: 10,
	}

	bTerms[18] = code{
		Code:        2 << 8,
		BitsWritten: 10,
	}

	bTerms[19] = code{
		Code:        12<<8 | 7<<5,
		BitsWritten: 11,
	}

	bTerms[20] = code{
		Code:        13 << 8,
		BitsWritten: 11,
	}

	bTerms[21] = code{
		Code:        13<<8 | 4<<5,
		BitsWritten: 11,
	}

	bTerms[22] = code{
		Code:        6<<8 | 7<<5,
		BitsWritten: 11,
	}

	bTerms[23] = code{
		Code:        5 << 8,
		BitsWritten: 11,
	}

	bTerms[24] = code{
		Code:        2<<8 | 7<<5,
		BitsWritten: 11,
	}

	bTerms[25] = code{
		Code:        3 << 8,
		BitsWritten: 11,
	}

	bTerms[26] = code{
		Code:        12<<8 | 10<<4,
		BitsWritten: 12,
	}

	bTerms[27] = code{
		Code:        12<<8 | 11<<4,
		BitsWritten: 12,
	}

	bTerms[28] = code{
		Code:        12<<8 | 12<<4,
		BitsWritten: 12,
	}

	bTerms[29] = code{
		Code:        12<<8 | 13<<4,
		BitsWritten: 12,
	}

	bTerms[30] = code{
		Code:        6<<8 | 8<<4,
		BitsWritten: 12,
	}

	bTerms[31] = code{
		Code:        6<<8 | 9<<4,
		BitsWritten: 12,
	}

	bTerms[32] = code{
		Code:        6<<8 | 10<<4,
		BitsWritten: 12,
	}

	bTerms[33] = code{
		Code:        6<<8 | 11<<4,
		BitsWritten: 12,
	}

	bTerms[34] = code{
		Code:        13<<8 | 2<<4,
		BitsWritten: 12,
	}

	bTerms[35] = code{
		Code:        13<<8 | 3<<4,
		BitsWritten: 12,
	}

	bTerms[36] = code{
		Code:        13<<8 | 4<<4,
		BitsWritten: 12,
	}

	bTerms[37] = code{
		Code:        13<<8 | 5<<4,
		BitsWritten: 12,
	}

	bTerms[38] = code{
		Code:        13<<8 | 6<<4,
		BitsWritten: 12,
	}

	bTerms[39] = code{
		Code:        13<<8 | 7<<4,
		BitsWritten: 12,
	}

	bTerms[40] = code{
		Code:        6<<8 | 12<<4,
		BitsWritten: 12,
	}

	bTerms[41] = code{
		Code:        6<<8 | 13<<4,
		BitsWritten: 12,
	}

	bTerms[42] = code{
		Code:        13<<8 | 10<<4,
		BitsWritten: 12,
	}

	bTerms[43] = code{
		Code:        13<<8 | 11<<4,
		BitsWritten: 12,
	}

	bTerms[44] = code{
		Code:        5<<8 | 4<<4,
		BitsWritten: 12,
	}

	bTerms[45] = code{
		Code:        5<<8 | 5<<4,
		BitsWritten: 12,
	}

	bTerms[46] = code{
		Code:        5<<8 | 6<<4,
		BitsWritten: 12,
	}

	bTerms[47] = code{
		Code:        5<<8 | 7<<4,
		BitsWritten: 12,
	}

	bTerms[48] = code{
		Code:        6<<8 | 4<<4,
		BitsWritten: 12,
	}

	bTerms[49] = code{
		Code:        6<<8 | 5<<4,
		BitsWritten: 12,
	}

	bTerms[50] = code{
		Code:        5<<8 | 2<<4,
		BitsWritten: 12,
	}

	bTerms[51] = code{
		Code:        5<<8 | 3<<4,
		BitsWritten: 12,
	}

	bTerms[52] = code{
		Code:        2<<8 | 4<<4,
		BitsWritten: 12,
	}

	bTerms[53] = code{
		Code:        3<<8 | 7<<4,
		BitsWritten: 12,
	}

	bTerms[54] = code{
		Code:        3<<8 | 8<<4,
		BitsWritten: 12,
	}

	bTerms[55] = code{
		Code:        2<<8 | 7<<4,
		BitsWritten: 12,
	}

	bTerms[56] = code{
		Code:        2<<8 | 8<<4,
		BitsWritten: 12,
	}

	bTerms[57] = code{
		Code:        5<<8 | 8<<4,
		BitsWritten: 12,
	}

	bTerms[58] = code{
		Code:        5<<8 | 9<<4,
		BitsWritten: 12,
	}

	bTerms[59] = code{
		Code:        2<<8 | 11<<4,
		BitsWritten: 12,
	}

	bTerms[60] = code{
		Code:        2<<8 | 12<<4,
		BitsWritten: 12,
	}

	bTerms[61] = code{
		Code:        5<<8 | 10<<4,
		BitsWritten: 12,
	}

	bTerms[62] = code{
		Code:        6<<8 | 6<<4,
		BitsWritten: 12,
	}

	bTerms[63] = code{
		Code:        6<<8 | 7<<4,
		BitsWritten: 12,
	}

	wTerms = make(map[int]code)

	wTerms[0] = code{
		Code:        53 << 8,
		BitsWritten: 8,
	}

	wTerms[1] = code{
		Code:        7 << (2 + 8),
		BitsWritten: 6,
	}

	wTerms[2] = code{
		Code:        7 << (4 + 8),
		BitsWritten: 4,
	}

	wTerms[3] = code{
		Code:        8 << (4 + 8),
		BitsWritten: 4,
	}

	wTerms[4] = code{
		Code:        11 << (4 + 8),
		BitsWritten: 4,
	}

	wTerms[5] = code{
		Code:        12 << (4 + 8),
		BitsWritten: 4,
	}

	wTerms[6] = code{
		Code:        14 << (4 + 8),
		BitsWritten: 4,
	}

	wTerms[7] = code{
		Code:        15 << (4 + 8),
		BitsWritten: 4,
	}

	wTerms[8] = code{
		Code:        19 << (3 + 8),
		BitsWritten: 5,
	}

	wTerms[9] = code{
		Code:        20 << (3 + 8),
		BitsWritten: 5,
	}

	wTerms[10] = code{
		Code:        7 << (3 + 8),
		BitsWritten: 5,
	}

	wTerms[11] = code{
		Code:        8 << (3 + 8),
		BitsWritten: 5,
	}

	wTerms[12] = code{
		Code:        8 << (2 + 8),
		BitsWritten: 6,
	}

	wTerms[13] = code{
		Code:        3 << (2 + 8),
		BitsWritten: 6,
	}

	wTerms[14] = code{
		Code:        52 << (2 + 8),
		BitsWritten: 6,
	}

	wTerms[15] = code{
		Code:        53 << (2 + 8),
		BitsWritten: 6,
	}

	wTerms[16] = code{
		Code:        42 << (2 + 8),
		BitsWritten: 6,
	}

	wTerms[17] = code{
		Code:        43 << (2 + 8),
		BitsWritten: 6,
	}

	wTerms[18] = code{
		Code:        39 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[19] = code{
		Code:        12 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[20] = code{
		Code:        8 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[21] = code{
		Code:        23 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[22] = code{
		Code:        3 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[23] = code{
		Code:        4 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[24] = code{
		Code:        40 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[25] = code{
		Code:        43 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[26] = code{
		Code:        19 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[27] = code{
		Code:        36 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[28] = code{
		Code:        24 << (1 + 8),
		BitsWritten: 7,
	}

	wTerms[29] = code{
		Code:        2 << 8,
		BitsWritten: 8,
	}

	wTerms[30] = code{
		Code:        3 << 8,
		BitsWritten: 8,
	}

	wTerms[31] = code{
		Code:        26 << 8,
		BitsWritten: 8,
	}

	wTerms[32] = code{
		Code:        27 << 8,
		BitsWritten: 8,
	}

	wTerms[33] = code{
		Code:        18 << 8,
		BitsWritten: 8,
	}

	wTerms[34] = code{
		Code:        19 << 8,
		BitsWritten: 8,
	}

	wTerms[35] = code{
		Code:        20 << 8,
		BitsWritten: 8,
	}

	wTerms[36] = code{
		Code:        21 << 8,
		BitsWritten: 8,
	}

	wTerms[37] = code{
		Code:        22 << 8,
		BitsWritten: 8,
	}

	wTerms[38] = code{
		Code:        23 << 8,
		BitsWritten: 8,
	}

	wTerms[39] = code{
		Code:        40 << 8,
		BitsWritten: 8,
	}

	wTerms[40] = code{
		Code:        41 << 8,
		BitsWritten: 8,
	}

	wTerms[41] = code{
		Code:        42 << 8,
		BitsWritten: 8,
	}

	wTerms[42] = code{
		Code:        43 << 8,
		BitsWritten: 8,
	}

	wTerms[43] = code{
		Code:        44 << 8,
		BitsWritten: 8,
	}

	wTerms[44] = code{
		Code:        45 << 8,
		BitsWritten: 8,
	}

	wTerms[45] = code{
		Code:        4 << 8,
		BitsWritten: 8,
	}

	wTerms[46] = code{
		Code:        5 << 8,
		BitsWritten: 8,
	}

	wTerms[47] = code{
		Code:        10 << 8,
		BitsWritten: 8,
	}

	wTerms[48] = code{
		Code:        11 << 8,
		BitsWritten: 8,
	}

	wTerms[49] = code{
		Code:        82 << 8,
		BitsWritten: 8,
	}

	wTerms[50] = code{
		Code:        83 << 8,
		BitsWritten: 8,
	}

	wTerms[51] = code{
		Code:        84 << 8,
		BitsWritten: 8,
	}

	wTerms[52] = code{
		Code:        85 << 8,
		BitsWritten: 8,
	}

	wTerms[53] = code{
		Code:        36 << 8,
		BitsWritten: 8,
	}

	wTerms[54] = code{
		Code:        37 << 8,
		BitsWritten: 8,
	}

	wTerms[55] = code{
		Code:        88 << 8,
		BitsWritten: 8,
	}

	wTerms[56] = code{
		Code:        89 << 8,
		BitsWritten: 8,
	}

	wTerms[57] = code{
		Code:        90 << 8,
		BitsWritten: 8,
	}

	wTerms[58] = code{
		Code:        91 << 8,
		BitsWritten: 8,
	}

	wTerms[59] = code{
		Code:        74 << 8,
		BitsWritten: 8,
	}

	wTerms[60] = code{
		Code:        75 << 8,
		BitsWritten: 8,
	}

	wTerms[61] = code{
		Code:        50 << 8,
		BitsWritten: 8,
	}

	wTerms[62] = code{
		Code:        51 << 8,
		BitsWritten: 8,
	}

	wTerms[63] = code{
		Code:        52 << 8,
		BitsWritten: 8,
	}

	bMakeups = make(map[int]code)

	bMakeups[64] = code{
		Code:        3<<8 | 3<<6,
		BitsWritten: 10,
	}

	bMakeups[128] = code{
		Code:        12<<8 | 8<<4,
		BitsWritten: 12,
	}

	bMakeups[192] = code{
		Code:        12<<8 | 9<<4,
		BitsWritten: 12,
	}

	bMakeups[256] = code{
		Code:        5<<8 | 11<<4,
		BitsWritten: 12,
	}

	bMakeups[320] = code{
		Code:        3<<8 | 3<<4,
		BitsWritten: 12,
	}

	bMakeups[384] = code{
		Code:        3<<8 | 4<<4,
		BitsWritten: 12,
	}

	bMakeups[448] = code{
		Code:        3<<8 | 5<<4,
		BitsWritten: 12,
	}

	bMakeups[512] = code{
		Code:        3<<8 | 12<<3,
		BitsWritten: 13,
	}

	bMakeups[576] = code{
		Code:        3<<8 | 13<<3,
		BitsWritten: 13,
	}

	bMakeups[640] = code{
		Code:        2<<8 | 10<<3,
		BitsWritten: 13,
	}

	bMakeups[704] = code{
		Code:        2<<8 | 11<<3,
		BitsWritten: 13,
	}

	bMakeups[768] = code{
		Code:        2<<8 | 12<<3,
		BitsWritten: 13,
	}

	bMakeups[832] = code{
		Code:        2<<8 | 13<<3,
		BitsWritten: 13,
	}

	bMakeups[896] = code{
		Code:        3<<8 | 18<<3,
		BitsWritten: 13,
	}

	bMakeups[960] = code{
		Code:        3<<8 | 19<<3,
		BitsWritten: 13,
	}

	bMakeups[1024] = code{
		Code:        3<<8 | 20<<3,
		BitsWritten: 13,
	}

	bMakeups[1088] = code{
		Code:        3<<8 | 21<<3,
		BitsWritten: 13,
	}

	bMakeups[1152] = code{
		Code:        3<<8 | 22<<3,
		BitsWritten: 13,
	}

	bMakeups[1216] = code{
		Code:        119 << 3,
		BitsWritten: 13,
	}

	bMakeups[1280] = code{
		Code:        2<<8 | 18<<3,
		BitsWritten: 13,
	}

	bMakeups[1344] = code{
		Code:        2<<8 | 19<<3,
		BitsWritten: 13,
	}

	bMakeups[1408] = code{
		Code:        2<<8 | 20<<3,
		BitsWritten: 13,
	}

	bMakeups[1472] = code{
		Code:        2<<8 | 21<<3,
		BitsWritten: 13,
	}

	bMakeups[1536] = code{
		Code:        2<<8 | 26<<3,
		BitsWritten: 13,
	}

	bMakeups[1600] = code{
		Code:        2<<8 | 27<<3,
		BitsWritten: 13,
	}

	bMakeups[1664] = code{
		Code:        3<<8 | 4<<3,
		BitsWritten: 13,
	}

	bMakeups[1728] = code{
		Code:        3<<8 | 5<<3,
		BitsWritten: 13,
	}

	wMakeups = make(map[int]code)

	wMakeups[64] = code{
		Code:        27 << (3 + 8),
		BitsWritten: 5,
	}

	wMakeups[128] = code{
		Code:        18 << (3 + 8),
		BitsWritten: 5,
	}

	wMakeups[192] = code{
		Code:        23 << (2 + 8),
		BitsWritten: 6,
	}

	wMakeups[256] = code{
		Code:        55 << (1 + 8),
		BitsWritten: 7,
	}

	wMakeups[320] = code{
		Code:        54 << 8,
		BitsWritten: 8,
	}

	wMakeups[384] = code{
		Code:        55 << 8,
		BitsWritten: 8,
	}

	wMakeups[448] = code{
		Code:        100 << 8,
		BitsWritten: 8,
	}

	wMakeups[512] = code{
		Code:        101 << 8,
		BitsWritten: 8,
	}

	wMakeups[576] = code{
		Code:        104 << 8,
		BitsWritten: 8,
	}

	wMakeups[640] = code{
		Code:        103 << 8,
		BitsWritten: 8,
	}

	wMakeups[704] = code{
		Code:        102 << 8,
		BitsWritten: 9,
	}

	wMakeups[768] = code{
		Code:        102<<8 | 1<<7,
		BitsWritten: 9,
	}

	wMakeups[832] = code{
		Code:        105 << 8,
		BitsWritten: 9,
	}

	wMakeups[896] = code{
		Code:        105<<8 | 1<<7,
		BitsWritten: 9,
	}

	wMakeups[960] = code{
		Code:        106 << 8,
		BitsWritten: 9,
	}

	wMakeups[1024] = code{
		Code:        106<<8 | 1<<7,
		BitsWritten: 9,
	}

	wMakeups[1088] = code{
		Code:        107 << 8,
		BitsWritten: 9,
	}

	wMakeups[1152] = code{
		Code:        107<<8 | 1<<7,
		BitsWritten: 9,
	}

	wMakeups[1216] = code{
		Code:        108 << 8,
		BitsWritten: 9,
	}

	wMakeups[1280] = code{
		Code:        108<<8 | 1<<7,
		BitsWritten: 9,
	}

	wMakeups[1344] = code{
		Code:        109 << 8,
		BitsWritten: 9,
	}

	wMakeups[1408] = code{
		Code:        109<<8 | 1<<7,
		BitsWritten: 9,
	}

	wMakeups[1472] = code{
		Code:        76 << 8,
		BitsWritten: 9,
	}

	wMakeups[1536] = code{
		Code:        76<<8 | 1<<7,
		BitsWritten: 9,
	}

	wMakeups[1600] = code{
		Code:        77 << 8,
		BitsWritten: 9,
	}

	wMakeups[1664] = code{
		Code:        24 << (2 + 8),
		BitsWritten: 6,
	}

	wMakeups[1728] = code{
		Code:        77<<8 | 1<<7,
		BitsWritten: 9,
	}

	commonMakeups = make(map[int]code)

	commonMakeups[1792] = code{
		Code:        1 << 8,
		BitsWritten: 11,
	}

	commonMakeups[1856] = code{
		Code:        1<<8 | 4<<5,
		BitsWritten: 11,
	}

	commonMakeups[1920] = code{
		Code:        1<<8 | 5<<5,
		BitsWritten: 11,
	}

	commonMakeups[1984] = code{
		Code:        1<<8 | 2<<4,
		BitsWritten: 12,
	}

	commonMakeups[2048] = code{
		Code:        1<<8 | 3<<4,
		BitsWritten: 12,
	}

	commonMakeups[2112] = code{
		Code:        1<<8 | 4<<4,
		BitsWritten: 12,
	}

	commonMakeups[2176] = code{
		Code:        1<<8 | 5<<4,
		BitsWritten: 12,
	}

	commonMakeups[2240] = code{
		Code:        1<<8 | 6<<4,
		BitsWritten: 12,
	}

	commonMakeups[2304] = code{
		Code:        1<<8 | 7<<4,
		BitsWritten: 12,
	}

	commonMakeups[2368] = code{
		Code:        1<<8 | 12<<4,
		BitsWritten: 12,
	}

	commonMakeups[2432] = code{
		Code:        1<<8 | 13<<4,
		BitsWritten: 12,
	}

	commonMakeups[2496] = code{
		Code:        1<<8 | 14<<4,
		BitsWritten: 12,
	}

	commonMakeups[2560] = code{
		Code:        1<<8 | 15<<4,
		BitsWritten: 12,
	}

	masks = make(map[int]byte)

	masks[0] = 0xFF
	masks[1] = 0xFE
	masks[2] = 0xFC
	masks[3] = 0xF8
	masks[4] = 0xF0
	masks[5] = 0xE0
	masks[6] = 0xC0
	masks[7] = 0x80
	masks[8] = 0x00
}
