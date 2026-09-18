package lyrics

import (
	"github.com/google/wire"

	"lyrics-server/internal/provider"
)

// ProviderSet 供 wire 装配。
var ProviderSet = wire.NewSet(
	NewNeteaseCloudMusicApiProvider,
	wire.Bind(new(provider.LyricsProvider), new(*NeteaseCloudMusicApiProvider)),
)
