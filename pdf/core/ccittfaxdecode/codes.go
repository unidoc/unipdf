package ccittfaxdecode

var (
	BTerms        map[int]Code
	WTerms        map[int]Code
	BMakeups      map[int]Code
	WMakeups      map[int]Code
	CommonMakeups map[int]Code
	masks         map[int]byte

	EOL = Code{
		Code:        1 << 4,
		BitsWritten: 12,
	}

	EOL1 = Code{
		Code:        3 << 3,
		BitsWritten: 13,
	}

	EOL0 = Code{
		Code:        2 << 3,
		BitsWritten: 13,
	}

	P = Code{
		Code:        1 << 12,
		BitsWritten: 4,
	}

	H = Code{
		Code:        1 << 13,
		BitsWritten: 3,
	}

	V0 = Code{
		Code:        1 << 15,
		BitsWritten: 1,
	}

	V1R = Code{
		Code:        3 << 13,
		BitsWritten: 3,
	}

	V2R = Code{
		Code:        3 << 10,
		BitsWritten: 6,
	}

	V3R = Code{
		Code:        3 << 9,
		BitsWritten: 7,
	}

	V1L = Code{
		Code:        2 << 13,
		BitsWritten: 3,
	}

	V2L = Code{
		Code:        2 << 10,
		BitsWritten: 6,
	}

	V3L = Code{
		Code:        2 << 9,
		BitsWritten: 7,
	}
)

type Code struct {
	Code        uint16
	BitsWritten int
}

func init() {
	BTerms = make(map[int]Code)

	BTerms[0] = Code{
		Code:        13<<8 | 3<<6,
		BitsWritten: 10,
	}

	BTerms[1] = Code{
		Code:        2 << (5 + 8),
		BitsWritten: 3,
	}

	BTerms[2] = Code{
		Code:        3 << (6 + 8),
		BitsWritten: 2,
	}

	BTerms[3] = Code{
		Code:        2 << (6 + 8),
		BitsWritten: 2,
	}

	BTerms[4] = Code{
		Code:        3 << (5 + 8),
		BitsWritten: 3,
	}

	BTerms[5] = Code{
		Code:        3 << (4 + 8),
		BitsWritten: 4,
	}

	BTerms[6] = Code{
		Code:        2 << (4 + 8),
		BitsWritten: 4,
	}

	BTerms[7] = Code{
		Code:        3 << (3 + 8),
		BitsWritten: 5,
	}

	BTerms[8] = Code{
		Code:        5 << (2 + 8),
		BitsWritten: 6,
	}

	BTerms[9] = Code{
		Code:        4 << (2 + 8),
		BitsWritten: 6,
	}

	BTerms[10] = Code{
		Code:        4 << (1 + 8),
		BitsWritten: 7,
	}

	BTerms[11] = Code{
		Code:        5 << (1 + 8),
		BitsWritten: 7,
	}

	BTerms[12] = Code{
		Code:        7 << (1 + 8),
		BitsWritten: 7,
	}

	BTerms[13] = Code{
		Code:        4 << 8,
		BitsWritten: 8,
	}

	BTerms[14] = Code{
		Code:        7 << 8,
		BitsWritten: 8,
	}

	BTerms[15] = Code{
		Code:        12 << 8,
		BitsWritten: 9,
	}

	BTerms[16] = Code{
		Code:        5<<8 | 3<<6,
		BitsWritten: 10,
	}

	BTerms[17] = Code{
		Code:        6 << 8,
		BitsWritten: 10,
	}

	BTerms[18] = Code{
		Code:        2 << 8,
		BitsWritten: 10,
	}

	BTerms[19] = Code{
		Code:        12<<8 | 7<<5,
		BitsWritten: 11,
	}

	BTerms[20] = Code{
		Code:        13 << 8,
		BitsWritten: 11,
	}

	BTerms[21] = Code{
		Code:        13<<8 | 4<<5,
		BitsWritten: 11,
	}

	BTerms[22] = Code{
		Code:        6<<8 | 7<<5,
		BitsWritten: 11,
	}

	BTerms[23] = Code{
		Code:        5 << 8,
		BitsWritten: 11,
	}

	BTerms[24] = Code{
		Code:        2<<8 | 7<<5,
		BitsWritten: 11,
	}

	BTerms[25] = Code{
		Code:        3 << 8,
		BitsWritten: 11,
	}

	BTerms[26] = Code{
		Code:        12<<8 | 10<<4,
		BitsWritten: 12,
	}

	BTerms[27] = Code{
		Code:        12<<8 | 11<<4,
		BitsWritten: 12,
	}

	BTerms[28] = Code{
		Code:        12<<8 | 12<<4,
		BitsWritten: 12,
	}

	BTerms[29] = Code{
		Code:        12<<8 | 13<<4,
		BitsWritten: 12,
	}

	BTerms[30] = Code{
		Code:        6<<8 | 8<<4,
		BitsWritten: 12,
	}

	BTerms[31] = Code{
		Code:        6<<8 | 9<<4,
		BitsWritten: 12,
	}

	BTerms[32] = Code{
		Code:        6<<8 | 10<<4,
		BitsWritten: 12,
	}

	BTerms[33] = Code{
		Code:        6<<8 | 11<<4,
		BitsWritten: 12,
	}

	BTerms[34] = Code{
		Code:        13<<8 | 2<<4,
		BitsWritten: 12,
	}

	BTerms[35] = Code{
		Code:        13<<8 | 3<<4,
		BitsWritten: 12,
	}

	BTerms[36] = Code{
		Code:        13<<8 | 4<<4,
		BitsWritten: 12,
	}

	BTerms[37] = Code{
		Code:        13<<8 | 5<<4,
		BitsWritten: 12,
	}

	BTerms[38] = Code{
		Code:        13<<8 | 6<<4,
		BitsWritten: 12,
	}

	BTerms[39] = Code{
		Code:        13<<8 | 7<<4,
		BitsWritten: 12,
	}

	BTerms[40] = Code{
		Code:        6<<8 | 12<<4,
		BitsWritten: 12,
	}

	BTerms[41] = Code{
		Code:        6<<8 | 13<<4,
		BitsWritten: 12,
	}

	BTerms[42] = Code{
		Code:        13<<8 | 10<<4,
		BitsWritten: 12,
	}

	BTerms[43] = Code{
		Code:        13<<8 | 11<<4,
		BitsWritten: 12,
	}

	BTerms[44] = Code{
		Code:        5<<8 | 4<<4,
		BitsWritten: 12,
	}

	BTerms[45] = Code{
		Code:        5<<8 | 5<<4,
		BitsWritten: 12,
	}

	BTerms[46] = Code{
		Code:        5<<8 | 6<<4,
		BitsWritten: 12,
	}

	BTerms[47] = Code{
		Code:        5<<8 | 7<<4,
		BitsWritten: 12,
	}

	BTerms[48] = Code{
		Code:        6<<8 | 4<<4,
		BitsWritten: 12,
	}

	BTerms[49] = Code{
		Code:        6<<8 | 5<<4,
		BitsWritten: 12,
	}

	BTerms[50] = Code{
		Code:        5<<8 | 2<<4,
		BitsWritten: 12,
	}

	BTerms[51] = Code{
		Code:        5<<8 | 3<<4,
		BitsWritten: 12,
	}

	BTerms[52] = Code{
		Code:        2<<8 | 4<<4,
		BitsWritten: 12,
	}

	BTerms[53] = Code{
		Code:        3<<8 | 7<<4,
		BitsWritten: 12,
	}

	BTerms[54] = Code{
		Code:        3<<8 | 8<<4,
		BitsWritten: 12,
	}

	BTerms[55] = Code{
		Code:        2<<8 | 7<<4,
		BitsWritten: 12,
	}

	BTerms[56] = Code{
		Code:        2<<8 | 8<<4,
		BitsWritten: 12,
	}

	BTerms[57] = Code{
		Code:        5<<8 | 8<<4,
		BitsWritten: 12,
	}

	BTerms[58] = Code{
		Code:        5<<8 | 9<<4,
		BitsWritten: 12,
	}

	BTerms[59] = Code{
		Code:        2<<8 | 11<<4,
		BitsWritten: 12,
	}

	BTerms[60] = Code{
		Code:        2<<8 | 12<<4,
		BitsWritten: 12,
	}

	BTerms[61] = Code{
		Code:        5<<8 | 10<<4,
		BitsWritten: 12,
	}

	BTerms[62] = Code{
		Code:        6<<8 | 6<<4,
		BitsWritten: 12,
	}

	BTerms[63] = Code{
		Code:        6<<8 | 7<<4,
		BitsWritten: 12,
	}

	WTerms = make(map[int]Code)

	WTerms[0] = Code{
		Code:        53 << 8,
		BitsWritten: 8,
	}

	WTerms[1] = Code{
		Code:        7 << (2 + 8),
		BitsWritten: 6,
	}

	WTerms[2] = Code{
		Code:        7 << (4 + 8),
		BitsWritten: 4,
	}

	WTerms[3] = Code{
		Code:        8 << (4 + 8),
		BitsWritten: 4,
	}

	WTerms[4] = Code{
		Code:        11 << (4 + 8),
		BitsWritten: 4,
	}

	WTerms[5] = Code{
		Code:        12 << (4 + 8),
		BitsWritten: 4,
	}

	WTerms[6] = Code{
		Code:        14 << (4 + 8),
		BitsWritten: 4,
	}

	WTerms[7] = Code{
		Code:        15 << (4 + 8),
		BitsWritten: 4,
	}

	WTerms[8] = Code{
		Code:        19 << (3 + 8),
		BitsWritten: 5,
	}

	WTerms[9] = Code{
		Code:        20 << (3 + 8),
		BitsWritten: 5,
	}

	WTerms[10] = Code{
		Code:        7 << (3 + 8),
		BitsWritten: 5,
	}

	WTerms[11] = Code{
		Code:        8 << (3 + 8),
		BitsWritten: 5,
	}

	WTerms[12] = Code{
		Code:        8 << (2 + 8),
		BitsWritten: 6,
	}

	WTerms[13] = Code{
		Code:        3 << (2 + 8),
		BitsWritten: 6,
	}

	WTerms[14] = Code{
		Code:        52 << (2 + 8),
		BitsWritten: 6,
	}

	WTerms[15] = Code{
		Code:        53 << (2 + 8),
		BitsWritten: 6,
	}

	WTerms[16] = Code{
		Code:        42 << (2 + 8),
		BitsWritten: 6,
	}

	WTerms[17] = Code{
		Code:        43 << (2 + 8),
		BitsWritten: 6,
	}

	WTerms[18] = Code{
		Code:        39 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[19] = Code{
		Code:        12 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[20] = Code{
		Code:        8 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[21] = Code{
		Code:        23 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[22] = Code{
		Code:        3 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[23] = Code{
		Code:        4 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[24] = Code{
		Code:        40 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[25] = Code{
		Code:        43 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[26] = Code{
		Code:        19 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[27] = Code{
		Code:        36 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[28] = Code{
		Code:        24 << (1 + 8),
		BitsWritten: 7,
	}

	WTerms[29] = Code{
		Code:        2 << 8,
		BitsWritten: 8,
	}

	WTerms[30] = Code{
		Code:        3 << 8,
		BitsWritten: 8,
	}

	WTerms[31] = Code{
		Code:        26 << 8,
		BitsWritten: 8,
	}

	WTerms[32] = Code{
		Code:        27 << 8,
		BitsWritten: 8,
	}

	WTerms[33] = Code{
		Code:        18 << 8,
		BitsWritten: 8,
	}

	WTerms[34] = Code{
		Code:        19 << 8,
		BitsWritten: 8,
	}

	WTerms[35] = Code{
		Code:        20 << 8,
		BitsWritten: 8,
	}

	WTerms[36] = Code{
		Code:        21 << 8,
		BitsWritten: 8,
	}

	WTerms[37] = Code{
		Code:        22 << 8,
		BitsWritten: 8,
	}

	WTerms[38] = Code{
		Code:        23 << 8,
		BitsWritten: 8,
	}

	WTerms[39] = Code{
		Code:        40 << 8,
		BitsWritten: 8,
	}

	WTerms[40] = Code{
		Code:        41 << 8,
		BitsWritten: 8,
	}

	WTerms[41] = Code{
		Code:        42 << 8,
		BitsWritten: 8,
	}

	WTerms[42] = Code{
		Code:        43 << 8,
		BitsWritten: 8,
	}

	WTerms[43] = Code{
		Code:        44 << 8,
		BitsWritten: 8,
	}

	WTerms[44] = Code{
		Code:        45 << 8,
		BitsWritten: 8,
	}

	WTerms[45] = Code{
		Code:        4 << 8,
		BitsWritten: 8,
	}

	WTerms[46] = Code{
		Code:        5 << 8,
		BitsWritten: 8,
	}

	WTerms[47] = Code{
		Code:        10 << 8,
		BitsWritten: 8,
	}

	WTerms[48] = Code{
		Code:        11 << 8,
		BitsWritten: 8,
	}

	WTerms[49] = Code{
		Code:        82 << 8,
		BitsWritten: 8,
	}

	WTerms[50] = Code{
		Code:        83 << 8,
		BitsWritten: 8,
	}

	WTerms[51] = Code{
		Code:        84 << 8,
		BitsWritten: 8,
	}

	WTerms[52] = Code{
		Code:        85 << 8,
		BitsWritten: 8,
	}

	WTerms[53] = Code{
		Code:        36 << 8,
		BitsWritten: 8,
	}

	WTerms[54] = Code{
		Code:        37 << 8,
		BitsWritten: 8,
	}

	WTerms[55] = Code{
		Code:        88 << 8,
		BitsWritten: 8,
	}

	WTerms[56] = Code{
		Code:        89 << 8,
		BitsWritten: 8,
	}

	WTerms[57] = Code{
		Code:        90 << 8,
		BitsWritten: 8,
	}

	WTerms[58] = Code{
		Code:        91 << 8,
		BitsWritten: 8,
	}

	WTerms[59] = Code{
		Code:        74 << 8,
		BitsWritten: 8,
	}

	WTerms[60] = Code{
		Code:        75 << 8,
		BitsWritten: 8,
	}

	WTerms[61] = Code{
		Code:        50 << 8,
		BitsWritten: 8,
	}

	WTerms[62] = Code{
		Code:        51 << 8,
		BitsWritten: 8,
	}

	WTerms[63] = Code{
		Code:        52 << 8,
		BitsWritten: 8,
	}

	BMakeups = make(map[int]Code)

	BMakeups[64] = Code{
		Code:        3<<8 | 3<<6,
		BitsWritten: 10,
	}

	BMakeups[128] = Code{
		Code:        12<<8 | 8<<4,
		BitsWritten: 12,
	}

	BMakeups[192] = Code{
		Code:        12<<8 | 9<<4,
		BitsWritten: 12,
	}

	BMakeups[256] = Code{
		Code:        5<<8 | 11<<4,
		BitsWritten: 12,
	}

	BMakeups[320] = Code{
		Code:        3<<8 | 3<<4,
		BitsWritten: 12,
	}

	BMakeups[384] = Code{
		Code:        3<<8 | 4<<4,
		BitsWritten: 12,
	}

	BMakeups[448] = Code{
		Code:        3<<8 | 5<<4,
		BitsWritten: 12,
	}

	BMakeups[512] = Code{
		Code:        3<<8 | 12<<3,
		BitsWritten: 13,
	}

	BMakeups[576] = Code{
		Code:        3<<8 | 13<<3,
		BitsWritten: 13,
	}

	BMakeups[640] = Code{
		Code:        2<<8 | 10<<3,
		BitsWritten: 13,
	}

	BMakeups[704] = Code{
		Code:        2<<8 | 11<<3,
		BitsWritten: 13,
	}

	BMakeups[768] = Code{
		Code:        2<<8 | 12<<3,
		BitsWritten: 13,
	}

	BMakeups[832] = Code{
		Code:        2<<8 | 13<<3,
		BitsWritten: 13,
	}

	BMakeups[896] = Code{
		Code:        3<<8 | 18<<3,
		BitsWritten: 13,
	}

	BMakeups[960] = Code{
		Code:        3<<8 | 19<<3,
		BitsWritten: 13,
	}

	BMakeups[1024] = Code{
		Code:        3<<8 | 20<<3,
		BitsWritten: 13,
	}

	BMakeups[1088] = Code{
		Code:        3<<8 | 21<<3,
		BitsWritten: 13,
	}

	BMakeups[1152] = Code{
		Code:        3<<8 | 22<<3,
		BitsWritten: 13,
	}

	BMakeups[1216] = Code{
		Code:        119 << 3,
		BitsWritten: 13,
	}

	BMakeups[1280] = Code{
		Code:        2<<8 | 18<<3,
		BitsWritten: 13,
	}

	BMakeups[1344] = Code{
		Code:        2<<8 | 19<<3,
		BitsWritten: 13,
	}

	BMakeups[1408] = Code{
		Code:        2<<8 | 20<<3,
		BitsWritten: 13,
	}

	BMakeups[1472] = Code{
		Code:        2<<8 | 21<<3,
		BitsWritten: 13,
	}

	BMakeups[1536] = Code{
		Code:        2<<8 | 26<<3,
		BitsWritten: 13,
	}

	BMakeups[1600] = Code{
		Code:        2<<8 | 27<<3,
		BitsWritten: 13,
	}

	BMakeups[1664] = Code{
		Code:        3<<8 | 4<<3,
		BitsWritten: 13,
	}

	BMakeups[1728] = Code{
		Code:        3<<8 | 5<<3,
		BitsWritten: 13,
	}

	WMakeups = make(map[int]Code)

	WMakeups[64] = Code{
		Code:        27 << (3 + 8),
		BitsWritten: 5,
	}

	WMakeups[128] = Code{
		Code:        18 << (3 + 8),
		BitsWritten: 5,
	}

	WMakeups[192] = Code{
		Code:        23 << (2 + 8),
		BitsWritten: 6,
	}

	WMakeups[256] = Code{
		Code:        55 << (1 + 8),
		BitsWritten: 7,
	}

	WMakeups[320] = Code{
		Code:        54 << 8,
		BitsWritten: 8,
	}

	WMakeups[384] = Code{
		Code:        55 << 8,
		BitsWritten: 8,
	}

	WMakeups[448] = Code{
		Code:        100 << 8,
		BitsWritten: 8,
	}

	WMakeups[512] = Code{
		Code:        101 << 8,
		BitsWritten: 8,
	}

	WMakeups[576] = Code{
		Code:        104 << 8,
		BitsWritten: 8,
	}

	WMakeups[640] = Code{
		Code:        103 << 8,
		BitsWritten: 8,
	}

	WMakeups[704] = Code{
		Code:        102 << 8,
		BitsWritten: 9,
	}

	WMakeups[768] = Code{
		Code:        102<<8 | 1<<7,
		BitsWritten: 9,
	}

	WMakeups[832] = Code{
		Code:        105 << 8,
		BitsWritten: 9,
	}

	WMakeups[896] = Code{
		Code:        105<<8 | 1<<7,
		BitsWritten: 9,
	}

	WMakeups[960] = Code{
		Code:        106 << 8,
		BitsWritten: 9,
	}

	WMakeups[1024] = Code{
		Code:        106<<8 | 1<<7,
		BitsWritten: 9,
	}

	WMakeups[1088] = Code{
		Code:        107 << 8,
		BitsWritten: 9,
	}

	WMakeups[1152] = Code{
		Code:        107<<8 | 1<<7,
		BitsWritten: 9,
	}

	WMakeups[1216] = Code{
		Code:        108 << 8,
		BitsWritten: 9,
	}

	WMakeups[1280] = Code{
		Code:        108<<8 | 1<<7,
		BitsWritten: 9,
	}

	WMakeups[1344] = Code{
		Code:        109 << 8,
		BitsWritten: 9,
	}

	WMakeups[1408] = Code{
		Code:        109<<8 | 1<<7,
		BitsWritten: 9,
	}

	WMakeups[1472] = Code{
		Code:        76 << 8,
		BitsWritten: 9,
	}

	WMakeups[1536] = Code{
		Code:        76<<8 | 1<<7,
		BitsWritten: 9,
	}

	WMakeups[1600] = Code{
		Code:        77 << 8,
		BitsWritten: 9,
	}

	WMakeups[1664] = Code{
		Code:        24 << (2 + 8),
		BitsWritten: 6,
	}

	WMakeups[1728] = Code{
		Code:        77<<8 | 1<<7,
		BitsWritten: 9,
	}

	CommonMakeups = make(map[int]Code)

	CommonMakeups[1792] = Code{
		Code:        1 << 8,
		BitsWritten: 11,
	}

	CommonMakeups[1856] = Code{
		Code:        1<<8 | 4<<5,
		BitsWritten: 11,
	}

	CommonMakeups[1920] = Code{
		Code:        1<<8 | 5<<5,
		BitsWritten: 11,
	}

	CommonMakeups[1984] = Code{
		Code:        1<<8 | 2<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2048] = Code{
		Code:        1<<8 | 3<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2112] = Code{
		Code:        1<<8 | 4<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2176] = Code{
		Code:        1<<8 | 5<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2240] = Code{
		Code:        1<<8 | 6<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2304] = Code{
		Code:        1<<8 | 7<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2368] = Code{
		Code:        1<<8 | 12<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2432] = Code{
		Code:        1<<8 | 13<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2496] = Code{
		Code:        1<<8 | 14<<4,
		BitsWritten: 12,
	}

	CommonMakeups[2560] = Code{
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