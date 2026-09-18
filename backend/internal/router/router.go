// Package router 路由注册。
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"lyrics-server/internal/handler"
	"lyrics-server/internal/middleware"
)

func NewRouter(
	h *handler.HealthHandler,
	lyric *handler.LyricHandler,
	word *handler.WordHandler,
	vocab *handler.VocabularyHandler,
	song *handler.SongHandler,
	settings *handler.SettingsHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Recovery(), middleware.CORS())

	r.GET("/healthz", h.Check)

	api := r.Group("/api")
	{
		api.GET("/healthz", h.Check)

		api.POST("/lyrics/parse", lyric.Parse)
		api.POST("/lyrics/translate", lyric.Translate)
		api.POST("/lyrics/tokenize", lyric.Tokenize)

		api.POST("/words/lookup", word.Lookup)
		api.GET("/words/audio/:language/:word", word.Audio)

		api.POST("/vocabulary", vocab.Add)
		api.GET("/vocabulary", vocab.List)
		api.GET("/vocabulary/:language/:id", vocab.Get)
		api.DELETE("/vocabulary/:language/:id", vocab.Delete)

		api.POST("/songs", song.Save)
		api.GET("/songs", song.List)
		api.POST("/songs/search", song.Search)
		api.POST("/songs/import", song.Import)
		api.GET("/songs/:id/lyrics", song.GetLyrics)

		api.GET("/settings", settings.Get)
		api.PUT("/settings", settings.Save)
		api.GET("/settings/datadir", settings.DataDir)
	}
	return r
}

var ProviderSet = wire.NewSet(NewRouter)
