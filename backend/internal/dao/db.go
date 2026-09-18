package dao

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/google/wire"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"lyrics-server/internal/config"
	"lyrics-server/internal/model"
)

// NewDB 打开 SQLite 并执行 AutoMigrate。路径相对运行目录；
// 若设置了 LYRICS_DATA_DIR（Electron 打包运行时指向 userData），则使用该目录。
func NewDB(cfg *config.Config) (*gorm.DB, error) {
	path := cfg.DB.Path
	if dir := os.Getenv("LYRICS_DATA_DIR"); dir != "" {
		path = filepath.Join(dir, "lyrics.db")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data dir: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}

	if err := db.AutoMigrate(
		&model.Song{},
		&model.SongLine{},
		&model.EnglishWord{},
		&model.JapaneseWord{},
		&model.WordSense{},
		&model.Example{},
		&model.WordCache{},
		&model.TranslationCache{},
	); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}
	return db, nil
}

// ProviderSet 聚合 db 与各 dao 构造器，供 wire 使用。
var ProviderSet = wire.NewSet(
	NewDB,
	LyricDaoProviderSet,
	WordDaoProviderSet,
)
