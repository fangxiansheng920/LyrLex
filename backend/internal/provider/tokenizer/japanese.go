// Package tokenizer 分词 Provider 实现。
package tokenizer

import (
	"context"
	"fmt"

	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"lyrics-server/internal/provider"
)

// JapaneseTokenizer 日语形态分析（kagome + UniDic，纯 Go 离线），
// 并做功能词合并：动词/形容词 + 助词(て/で) + 助动词 + 补助动词 → 一个词组。
type JapaneseTokenizer struct {
	t *tokenizer.Tokenizer
}

func NewJapaneseTokenizer() (*JapaneseTokenizer, error) {
	t, err := tokenizer.New(uni.Dict())
	if err != nil {
		return nil, fmt.Errorf("init kagome unidic: %w", err)
	}
	return &JapaneseTokenizer{t: t}, nil
}

func (j *JapaneseTokenizer) Name() string { return "kagome-unidic" }

func (j *JapaneseTokenizer) Tokenize(ctx context.Context, lang, line string) ([]provider.Token, error) {
	if lang != "ja" {
		return nil, fmt.Errorf("japanese tokenizer: unsupported language %q", lang)
	}

	ktokens := j.t.Tokenize(line)
	out := make([]provider.Token, 0, len(ktokens))
	groupID := 0

	for i := 0; i < len(ktokens); i++ {
		k := ktokens[i]
		pos0 := posOf(k)
		if pos0 == "" {
			// BOS/EOS 边界 token
			continue
		}

		switch pos0 {
		case "空白", "記号", "補助記号":
			out = append(out, provider.Token{Surface: k.Surface, Addable: false})

		case "助詞", "助動詞", "接頭辞", "接尾辞":
			out = append(out, provider.Token{
				Surface: k.Surface,
				Reading: readingOf(k),
				Romaji:  provider.RomajiJA(readingOf(k)),
				POS:     pos0,
				Addable: false,
			})

		case "動詞", "形容詞", "形状詞":
			// 开始一个合并词组
			groupID++
			surface := k.Surface
			readings := readingOf(k)
			lemma, _ := k.BaseForm()

			i++
			for i < len(ktokens) {
				n := ktokens[i]
				np0 := posOf(n)
				merge := np0 == "助動詞" ||
					(np0 == "助詞" && (n.Surface == "て" || n.Surface == "で")) ||
					(np0 == "動詞" && isAuxVerb(n))
				if !merge {
					break
				}
				surface += n.Surface
				readings += readingOf(n)
				i++
			}
			i-- // 外层 for 会 i++

			out = append(out, provider.Token{
				Surface: surface,
				Lemma:   lemma,
				POS:     pos0,
				Reading: readings,
				Romaji:  provider.RomajiJA(readings),
				GroupID: groupID,
				Addable: true,
			})

		default:
			// 名詞 / 代名詞 / 副詞 / 連体詞 / 感動詞 等独立可收藏词
			lemma, _ := k.BaseForm()
			out = append(out, provider.Token{
				Surface: k.Surface,
				Lemma:   lemma,
				POS:     pos0,
				Reading: readingOf(k),
				Romaji:  provider.RomajiJA(readingOf(k)),
				Addable: true,
			})
		}
	}
	return out, nil
}

func posOf(k tokenizer.Token) string {
	pos := k.POS()
	if len(pos) == 0 {
		return ""
	}
	return pos[0]
}

func readingOf(k tokenizer.Token) string {
	if r, ok := k.Pronunciation(); ok && r != "" {
		return r
	}
	r, _ := k.Reading()
	return r
}

// isAuxVerb 补助动词（しまった/いる/いく/くる…）。
func isAuxVerb(k tokenizer.Token) bool {
	base, _ := k.BaseForm()
	switch base {
	case "いる", "いく", "くる", "しまう", "おく", "みる", "くれる", "もらう", "やる", "あげる":
		return true
	}
	return false
}
