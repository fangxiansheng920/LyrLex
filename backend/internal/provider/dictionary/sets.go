package dictionary

import (
	"context"
	"fmt"

	"github.com/google/wire"

	"lyrics-server/internal/provider"
)

// CompositeDictionaryProvider 按语言路由到具体词典。
type CompositeDictionaryProvider struct {
	en *FreeDictionaryProvider
	ja *JishoDictionaryProvider
}

func NewCompositeDictionaryProvider(en *FreeDictionaryProvider, ja *JishoDictionaryProvider) *CompositeDictionaryProvider {
	return &CompositeDictionaryProvider{en: en, ja: ja}
}

func (c *CompositeDictionaryProvider) Name() string { return "composite" }

func (c *CompositeDictionaryProvider) Lookup(ctx context.Context, lang, word string) (*provider.DictEntry, error) {
	switch lang {
	case "en":
		return c.en.Lookup(ctx, lang, word)
	case "ja":
		return c.ja.Lookup(ctx, lang, word)
	default:
		return nil, fmt.Errorf("dictionary: unsupported language %q", lang)
	}
}

// ProviderSet 聚合各词典实现，并绑定接口。
var ProviderSet = wire.NewSet(
	NewFreeDictionaryProvider,
	NewJishoDictionaryProvider,
	NewCompositeDictionaryProvider,
	wire.Bind(new(provider.DictionaryProvider), new(*CompositeDictionaryProvider)),
)
