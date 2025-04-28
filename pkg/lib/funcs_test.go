package lib

import (
	"testing"
)

func TestMaskedText(t *testing.T) {
	testCases := []struct {
		input    string
		n        int
		expected string
	}{
		{"1234567890", 3, "123...890"},
		{"hello world", 5, "hel...rld"},
		{"test", 10, "t...t"},
	}

	for _, tc := range testCases {
		result := MaskedText(tc.input, tc.n)
		if result != tc.expected {
			t.Errorf("MaskedText(%s, %d) = %s; want %s", tc.input, tc.n, result, tc.expected)
		}
	}
}
