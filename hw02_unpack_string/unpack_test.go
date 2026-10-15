package hw02unpackstring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "🙃0", expected: ""},
		{input: "aaф0b", expected: "aab"},
		{input: "abcd", expected: "abcd"},
		{input: "a", expected: "a"},
		{input: "a1b9", expected: "abbbbbbbbb"},
		{input: "a0b0c", expected: "c"},
		{input: "d\n5abc", expected: "d\n\n\n\n\nabc"},
		{input: "Привет", expected: "Привет"},
		{input: "я3世2🙃3", expected: "яяя世世🙃🙃🙃"},
		{input: "🙃0界2ф0", expected: "界界"},
		{input: "\x002a", expected: "\x00\x00a"},
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},
		{input: `\0`, expected: "0"},
		{input: `\00`, expected: ""},
		{input: `\90x`, expected: "x"},
		{input: `\\0a`, expected: "a"},
		{input: `a2\3`, expected: "aa3"},
		{input: `\\`, expected: `\`},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{
		"3abc", "45", "aaa10b", "0", "9", "a00", "a01", "🙃12", "a2b34",
		`\`, `abc\`, `qw\ne`, `\a`, `\🙃`, `\123`, `\\12`,
	}
	for _, tc := range invalidStrings {
		t.Run(tc, func(t *testing.T) {
			result, err := Unpack(tc)
			require.ErrorIs(t, err, ErrInvalidString)
			require.Empty(t, result)
		})
	}
}
