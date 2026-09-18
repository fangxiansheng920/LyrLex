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
	FindSongByTitleArtist(ctx context.Context, title, artist string) (*model.Song, error)
	ReplaceSongLines(ctx context.Context, songID int64, lines []model.SongLine) error
	DeleteSong(ctx context.Context, id int64) error
	DeleteSongs(ctx context.Context, ids []int64) error
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

// FindSongByTitleArtist 按歌名+歌手找已存在的歌曲；不存在返回 (nil, nil)。
func (d *lyricDaoGorm) FindSongByTitleArtist(ctx context.Context, title, artist string) (*model.Song, error) {
	var song model.Song
	err := d.db.WithContext(ctx).
		Where("title = ? AND artist = ?", title, artist).
		First(&song).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &song, nil
}

// ReplaceSongLines 用新的歌词行替换某歌曲的全部旧行。
func (d *lyricDaoGorm) ReplaceSongLines(ctx context.Context, songID int64, lines []model.SongLine) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("song_id = ?", songID).Delete(&model.SongLine{}).Error; err != nil {
			return err
		}
		for i := range lines {
			lines[i].SongID = songID
		}
		if len(lines) == 0 {
			return nil
		}
		return tx.Create(&lines).Error
	})
}

// DeleteSong 删除歌曲及其歌词行。
func (d *lyricDaoGorm) DeleteSong(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("song_id = ?", id).Delete(&model.SongLine{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Song{}, id).Error
	})
}

// DeleteSongs 批量删除歌曲及其歌词行。
func (d *lyricDaoGorm) DeleteSongs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("song_id IN ?", ids).Delete(&model.SongLine{}).Error; err != nil {
			return err
		}
		return tx.Where("id IN ?", ids).Delete(&model.Song{}).Error
	})
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
