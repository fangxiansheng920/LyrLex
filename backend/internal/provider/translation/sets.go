package translation

import (
	"github.com/google/wire"

	"lyrics-server/internal/provider"
)

// ProviderSet 翻译 Provider：有道优先（配置密钥时），否则/失败回退 MyMemory。
var ProviderSet = wire.NewSet(
	NewMyMemoryTranslationProvider,
	NewYoudaoTranslationProvider,
	NewCompositeTranslationProvider,
	wire.Bind(new(provider.TranslationProvider), new(*CompositeTranslationProvider)),
)
