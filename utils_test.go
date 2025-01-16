package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemSuffixIfExists(t *testing.T) {
	type RemSuffixIfExistsTest struct {
		s        string
		suffix   string
		expected string
	}

	tests := []RemSuffixIfExistsTest{
		// Empty strings
		{
			s:        "",
			suffix:   "",
			expected: "",
		},
		{
			s:        "",
			suffix:   "a",
			expected: "",
		},
		{
			s:        "",
			suffix:   "abc",
			expected: "",
		},
		{
			s:        "a",
			suffix:   "",
			expected: "a",
		},
		{
			s:        "abc",
			suffix:   "",
			expected: "abc",
		},

		// Proper usage
		{
			s:        "12345678",
			suffix:   "8",
			expected: "1234567",
		},
		{
			s:        "12345678",
			suffix:   "678",
			expected: "12345",
		},
		{
			s:        "12345678",
			suffix:   "a",
			expected: "12345678",
		},
		{
			s:        "12345678",
			suffix:   "abc",
			expected: "12345678",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, remSuffixIfExists(test.s, test.suffix))
	}
}
