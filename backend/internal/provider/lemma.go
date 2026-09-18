package provider

import (
	"strings"

	"github.com/kljensen/snowball"
)

// LemmaEN 英文词形还原（词典优先，此处为兜底词干还原）。
func LemmaEN(word string) string {
	lower := strings.ToLower(word)
	stem, err := snowball.Stem(lower, "english", false)
	if err != nil || stem == "" {
		return lower
	}
	return stem
}
