package mmr

func maxInt(f, s int) int {
	if f < s {
		return s
	}
	return f
}

func minInt(f, s int) int {
	if f > s {
		return s
	}
	return f
}
