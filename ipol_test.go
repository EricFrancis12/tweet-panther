package main

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncIpol(t *testing.T) {
	type FuncIpolTest struct {
		input     string
		expected  string
		shouldErr bool
	}

	f := newFuncIpol("|*", "*|")
	f.RegisterFn("pathEscape", func(args ...string) (string, error) {
		if len(args) != 1 {
			return "", fmt.Errorf("pathEscape requires 1 argument, but got (%d) arguments instead", len(args))
		}
		return url.PathEscape(args[0]), nil
	})

	tests := []*FuncIpolTest{
		// Empty string
		{
			input:     "",
			expected:  "",
			shouldErr: false,
		},

		// No substitutions
		{
			input:     "some text",
			expected:  "some text",
			shouldErr: false,
		},

		// Proper usasge
		{
			input:     "|*pathEscape(https://example.com/news/offbeat/video-bear-swim-swimming-pool)*|",
			expected:  "https:%2F%2Fexample.com%2Fnews%2Foffbeat%2Fvideo-bear-swim-swimming-pool",
			shouldErr: false,
		},
		{
			input:     "This url is path-escaped: |*pathEscape(https://example.com/news/offbeat/video-bear-swim-swimming-pool)*|",
			expected:  "This url is path-escaped: https:%2F%2Fexample.com%2Fnews%2Foffbeat%2Fvideo-bear-swim-swimming-pool",
			shouldErr: false,
		},
	}

	for _, test := range tests {
		s, err := f.Eval(test.input)
		assert.Equal(t, test.expected, s)
		if test.shouldErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
