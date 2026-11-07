package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(text string) []string {
	frequencies := make(map[string]int)
	for _, word := range strings.Fields(text) {
		frequencies[word]++
	}

	words := make([]string, 0, len(frequencies))
	for word := range frequencies {
		words = append(words, word)
	}

	sort.Slice(words, func(i, j int) bool {
		if frequencies[words[i]] == frequencies[words[j]] {
			return words[i] < words[j]
		}
		return frequencies[words[i]] > frequencies[words[j]]
	})

	if len(words) > 10 {
		words = words[:10]
	}
	return words
}
