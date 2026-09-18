package handler

import "github.com/google/wire"

// ProviderSet 聚合全部 handler 构造器，供 wire 使用。
var ProviderSet = wire.NewSet(
	NewHealthHandler,
	NewLyricHandler,
	NewWordHandler,
	NewVocabularyHandler,
	NewSongHandler,
	NewSettingsHandler,
)
