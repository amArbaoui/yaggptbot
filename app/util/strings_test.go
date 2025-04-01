package util

import (
	"slices"
	"testing"
)

func TestSliceString(t *testing.T) {
	tests := []struct {
		testData string
		maxLen   int
		expected []string
	}{
		{"test string", 2, []string{"te", "st", " s", "tr", "in", "g"}},
		{"test string", 0, []string{"test string"}},
		{"test string", 100, []string{"test string"}},
		{"", 2, []string{}},
	}
	for _, tt := range tests {
		slicedString := SliceString(tt.testData, tt.maxLen)
		if slices.Compare(slicedString, tt.expected) != 0 {
			t.Errorf("case %s, with len %d, expected %s, got %s", tt.testData, tt.maxLen, tt.expected, slicedString)
		}
	}
}
