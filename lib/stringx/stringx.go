package stringx

func Filter(s string, filter func(r rune) bool) string {
	var n int
	runes := []rune(s)
	for i, r := range runes {
		if n < i {
			runes[n] = r
		}
		if !filter(r) {
			n++
		}
	}

	return string(runes[:n])
}
