package hw02unpackstring

import (
	"errors"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	var result strings.Builder
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		current := runes[i]
		if isDigit(current) {
			return "", ErrInvalidString
		}

		if current == '\\' {
			i++
			if i == len(runes) || (!isDigit(runes[i]) && runes[i] != '\\') {
				return "", ErrInvalidString
			}
			current = runes[i]
		}

		count := 1
		if i+1 < len(runes) && isDigit(runes[i+1]) {
			i++
			count = int(runes[i] - '0')
		}
		for j := 0; j < count; j++ {
			result.WriteRune(current)
		}
	}

	return result.String(), nil
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
