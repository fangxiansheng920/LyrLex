package dao

import (
	"context"
	"errors"

	"github.com/google/wire"
	"gorm.io/gorm"

	"lyrics-server/internal/model"
)

// LyricDao 歌词数据访问接口。
type LyricDao interface {
	SaveSong(ctx context.Context, song *model.Song, lines []model.SongLine) error
	GetSong(ctx context.Context, id int64) (*model.Song, error)
	ListSongs(ctx context.Context, query string) ([]model.Song, error)
	GetSongLines(ctx context.Context, songID int64) ([]model.SongLine, error)
	GetTranslation(ctx context.Context, hash string) (*model.TranslationCache, error)
	SetTranslation(ctx context.Context, tc *model.TranslationCache) error
}

type lyricDaoGorm struct {
	db *gorm.DB
}

func NewLyricDao(db *gorm.DB) *lyricDaoGorm {
	return &lyricDaoGorm{db: db}
}

func (d *lyricDaoGorm) SaveSong(ctx context.Context, song *model.Song, lines []model.SongLine) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(song).Error; err != nil {
			return err
		}
		for i := range lines {
			lines[i].SongID = song.ID
		}
		if len(lines) > 0 {
			return tx.Create(&lines).Error
		}
		return nil
	})
}

func (d *lyricDaoGorm) GetSong(ctx context.Context, id int64) (*model.Song, error) {
	var song model.Song
	if err := d.db.WithContext(ctx).First(&song, id).Error; err != nil {
		return nil, err
	}
	return &song, nil
}

func (d *lyricDaoGorm) ListSongs(ctx context.Context, query string) ([]model.Song, error) {
	var songs []model.Song
	q := d.db.WithContext(ctx).Order("created_at DESC")
	if query != "" {
		q = q.Where("title LIKE ? OR artist LIKE ?", "%"+query+"%", "%"+query+"%")
	}
	if err := q.Find(&songs).Error; err != nil {
		return nil, err
	}
	return songs, nil
}

func (d *lyricDaoGorm) GetSongLines(ctx context.Context, songID int64) ([]model.SongLine, error) {
	var lines []model.SongLine
	if err := d.db.WithContext(ctx).
		Where("song_id = ?", songID).
		Order("line_index ASC").
		Find(&lines).Error; err != nil {
		return nil, err
	}
	return lines, nil
}

func (d *lyricDaoGorm) GetTranslation(ctx context.Context, hash string) (*model.TranslationCache, error) {
	var tc model.TranslationCache
	if err := d.db.WithContext(ctx).Where("hash = ?", hash).First(&tc).Error; err != nil {
		return nil, err
	}
	return &tc, nil
}

func (d *lyricDaoGorm) SetTranslation(ctx context.Context, tc *model.TranslationCache) error {
	var existing model.TranslationCache
	err := d.db.WithContext(ctx).Where("hash = ?", tc.Hash).First(&existing).Error
	if err == nil {
		return d.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"result":     tc.Result,
			"updated_at": tc.UpdatedAt,
		}).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return d.db.WithContext(ctx).Create(tc).Error
	}
	return err
}

var _ LyricDao = (*lyricDaoGorm)(nil)

var LyricDaoProviderSet = wire.NewSet(
	NewLyricDao,
	wire.Bind(new(LyricDao), new(*lyricDaoGorm)),
)
