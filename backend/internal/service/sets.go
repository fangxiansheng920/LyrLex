package service

import "github.com/google/wire"

// ProviderSet 聚合全部 service 构造器，供 wire 使用。
var ProviderSet = wire.NewSet(
	LyricProviderSet,
	WordProviderSet,
	VocabularyProviderSet,
	SettingsProviderSet,
)
