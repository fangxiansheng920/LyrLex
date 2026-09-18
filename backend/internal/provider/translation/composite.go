package translation

import "context"

// CompositeTranslationProvider：有道优先（配置了密钥时），失败或未配置回退免费源 MyMemory。
type CompositeTranslationProvider struct {
	youdao   *YoudaoTranslationProvider
	fallback *MyMemoryTranslationProvider
}

func NewCompositeTranslationProvider(
	youdao *YoudaoTranslationProvider,
	fallback *MyMemoryTranslationProvider,
) *CompositeTranslationProvider {
	return &CompositeTranslationProvider{youdao: youdao, fallback: fallback}
}

func (c *CompositeTranslationProvider) Name() string { return "composite" }

func (c *CompositeTranslationProvider) Translate(ctx context.Context, text, source, target string) (string, error) {
	if c.youdao.HasKey() {
		if result, err := c.youdao.Translate(ctx, text, source, target); err == nil && result != "" {
			return result, nil
		}
	}
	return c.fallback.Translate(ctx, text, source, target)
}
