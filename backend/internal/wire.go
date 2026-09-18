//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/google/wire"

	"lyrics-server/internal/config"
	"lyrics-server/internal/dao"
	"lyrics-server/internal/handler"
	"lyrics-server/internal/provider/dictionary"
	"lyrics-server/internal/provider/lyrics"
	"lyrics-server/internal/provider/pronunciation"
	"lyrics-server/internal/provider/tokenizer"
	"lyrics-server/internal/provider/translation"
	"lyrics-server/internal/router"
	"lyrics-server/internal/service"
)

// InitializeApp 由 wire 生成 wire_gen.go 中的实现。
func InitializeApp(cfg *config.Config) (*App, error) {
	wire.Build(
		dao.ProviderSet,
		translation.ProviderSet,
		dictionary.ProviderSet,
		tokenizer.ProviderSet,
		pronunciation.ProviderSet,
		lyrics.ProviderSet,
		service.ProviderSet,
		handler.ProviderSet,
		router.ProviderSet,
		NewApp,
	)
	return nil, nil
}
