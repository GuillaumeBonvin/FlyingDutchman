package internal

// returns the minimum value between int a and b
func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
