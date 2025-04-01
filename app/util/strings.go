package util

import "unicode/utf8"

func SliceString(source string, maxLen int) []string {
	sourceLen := utf8.RuneCountInString(source)
	if source == "" {
		return []string{}
	}
	if maxLen == 0 || maxLen > sourceLen {
		return []string{source}
	}
	batches := sourceLen / maxLen
	if sourceLen%maxLen != 0 {
		batches++
	}
	res := make([]string, 0, batches)
	for i := 0; i < batches; i++ {
		start := i * maxLen
		end := (i + 1) * maxLen
		if end > sourceLen {
			end = sourceLen
		}
		res = append(res, source[start:end])
	}
	return res
}
