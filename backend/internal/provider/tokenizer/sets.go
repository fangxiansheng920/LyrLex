package tokenizer

import (
	"context"
	"fmt"

	"github.com/google/wire"

	"lyrics-server/internal/provider"
)

// CompositeTokenizer 按语言路由到具体分词器。
type CompositeTokenizer struct {
	en *EnglishTokenizer
	ja *JapaneseTokenizer
}

func NewCompositeTokenizer(en *EnglishTokenizer, ja *JapaneseTokenizer) *CompositeTokenizer {
	return &CompositeTokenizer{en: en, ja: ja}
}

func (c *CompositeTokenizer) Name() string { return "composite" }

func (c *CompositeTokenizer) Tokenize(ctx context.Context, lang, line string) ([]provider.Token, error) {
	switch lang {
	case "en":
		return c.en.Tokenize(ctx, lang, line)
	case "ja":
		return c.ja.Tokenize(ctx, lang, line)
	default:
		return nil, fmt.Errorf("tokenizer: unsupported language %q", lang)
	}
}

// ProviderSet 聚合各分词器，并绑定接口。
var ProviderSet = wire.NewSet(
	NewEnglishTokenizer,
	NewJapaneseTokenizer,
	NewCompositeTokenizer,
	wire.Bind(new(provider.TokenizerProvider), new(*CompositeTokenizer)),
)
