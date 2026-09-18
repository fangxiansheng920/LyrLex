// Package tokenizer 分词 Provider 实现。
package tokenizer

import (
	"context"
	"fmt"
	"regexp"

	"lyrics-server/internal/provider"
)

// EnglishTokenizer 英文 tokenizer：正则切词 + 缩写处理 + 词形还原。
type EnglishTokenizer struct {
	wordRe *regexp.Regexp
}

func NewEnglishTokenizer() *EnglishTokenizer {
	return &EnglishTokenizer{
		wordRe: regexp.MustCompile(`[A-Za-z]+(?:['’][A-Za-z]+)?`),
	}
}

func (t *EnglishTokenizer) Name() string { return "english-simple" }

func (t *EnglishTokenizer) Tokenize(ctx context.Context, lang, line string) ([]provider.Token, error) {
	if lang != "en" {
		return nil, fmt.Errorf("english tokenizer: unsupported language %q", lang)
	}

	var tokens []provider.Token
	pos := 0
	for _, m := range t.wordRe.FindAllStringIndex(line, -1) {
		if m[0] > pos {
			tokens = append(tokens, gapToken(line[pos:m[0]]))
		}
		surface := line[m[0]:m[1]]
		tokens = append(tokens, provider.Token{
			Surface: surface,
			Lemma:   provider.LemmaEN(surface),
			Addable: true,
		})
		pos = m[1]
	}
	if pos < len(line) {
		tokens = append(tokens, gapToken(line[pos:]))
	}
	return tokens, nil
}

func gapToken(s string) provider.Token {
	return provider.Token{Surface: s, Addable: false}
}
