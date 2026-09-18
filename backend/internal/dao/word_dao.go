package dao

import (
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"

	"lyrics-server/internal/model"
)

// WordDao 单词数据访问接口（词缓存 + 生词本）。
type WordDao interface {
	GetWordCache(ctx context.Context, lang, word string) (*model.WordCache, error)
	SetWordCache(ctx context.Context, wc *model.WordCache) error

	UpsertEnglishWord(ctx context.Context, w *model.EnglishWord) error
	GetEnglishWord(ctx context.Context, id int64) (*model.EnglishWord, error)
	GetEnglishWordByWord(ctx context.Context, word string) (*model.EnglishWord, error)
	ListEnglishWords(ctx context.Context, query string) ([]model.EnglishWord, error)
	DeleteEnglishWord(ctx context.Context, id int64) error

	UpsertJapaneseWord(ctx context.Context, w *model.JapaneseWord) error
	GetJapaneseWord(ctx context.Context, id int64) (*model.JapaneseWord, error)
	GetJapaneseWordByWord(ctx context.Context, word string) (*model.JapaneseWord, error)
	ListJapaneseWords(ctx context.Context, query string) ([]model.JapaneseWord, error)
	DeleteJapaneseWord(ctx context.Context, id int64) error

	ReplaceSenses(ctx context.Context, wordType string, wordID int64, senses []model.WordSense) error
	ListSenses(ctx context.Context, wordType string, wordID int64) ([]model.WordSense, error)
	AddExample(ctx context.Context, e *model.Example) error
	ListExamples(ctx context.Context, wordType string, wordID int64) ([]model.Example, error)
	ListWordIDsBySong(ctx context.Context, wordType string, songID int64) ([]int64, error)
}

type wordDaoGorm struct {
	db *gorm.DB
}

func NewWordDao(db *gorm.DB) *wordDaoGorm {
	return &wordDaoGorm{db: db}
}

func (d *wordDaoGorm) GetWordCache(ctx context.Context, lang, word string) (*model.WordCache, error) {
	var wc model.WordCache
	if err := d.db.WithContext(ctx).
		Where("lang = ? AND word = ?", lang, word).
		First(&wc).Error; err != nil {
		return nil, err
	}
	return &wc, nil
}

func (d *wordDaoGorm) SetWordCache(ctx context.Context, wc *model.WordCache) error {
	return d.db.WithContext(ctx).
		Where("lang = ? AND word = ?", wc.Lang, wc.Word).
		Assign(model.WordCache{Payload: wc.Payload, UpdatedAt: wc.UpdatedAt}).
		FirstOrCreate(wc).Error
}

func (d *wordDaoGorm) UpsertEnglishWord(ctx context.Context, w *model.EnglishWord) error {
	return d.db.WithContext(ctx).
		Where("word = ?", w.Word).
		Assign(model.EnglishWord{
			Lemma:      w.Lemma,
			PhoneticUK: w.PhoneticUK,
			PhoneticUS: w.PhoneticUS,
			AudioPath:  w.AudioPath,
		}).
		FirstOrCreate(w).Error
}

func (d *wordDaoGorm) GetEnglishWord(ctx context.Context, id int64) (*model.EnglishWord, error) {
	var w model.EnglishWord
	if err := d.db.WithContext(ctx).First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (d *wordDaoGorm) GetEnglishWordByWord(ctx context.Context, word string) (*model.EnglishWord, error) {
	var w model.EnglishWord
	if err := d.db.WithContext(ctx).Where("word = ?", word).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (d *wordDaoGorm) ListEnglishWords(ctx context.Context, query string) ([]model.EnglishWord, error) {
	var words []model.EnglishWord
	q := d.db.WithContext(ctx).Order("created_at DESC")
	if query != "" {
		q = q.Where("word LIKE ? OR lemma LIKE ?", "%"+query+"%", "%"+query+"%")
	}
	if err := q.Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
}

func (d *wordDaoGorm) DeleteEnglishWord(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("word_type = ? AND word_id = ?", "en", id).Delete(&model.WordSense{}).Error; err != nil {
			return err
		}
		if err := tx.Where("word_type = ? AND word_id = ?", "en", id).Delete(&model.Example{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.EnglishWord{}, id).Error
	})
}

func (d *wordDaoGorm) UpsertJapaneseWord(ctx context.Context, w *model.JapaneseWord) error {
	return d.db.WithContext(ctx).
		Where("word = ?", w.Word).
		Assign(model.JapaneseWord{
			Lemma:     w.Lemma,
			Kana:      w.Kana,
			Romaji:    w.Romaji,
			Accent:    w.Accent,
			POS:       w.POS,
			AudioPath: w.AudioPath,
		}).
		FirstOrCreate(w).Error
}

func (d *wordDaoGorm) GetJapaneseWord(ctx context.Context, id int64) (*model.JapaneseWord, error) {
	var w model.JapaneseWord
	if err := d.db.WithContext(ctx).First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (d *wordDaoGorm) GetJapaneseWordByWord(ctx context.Context, word string) (*model.JapaneseWord, error) {
	var w model.JapaneseWord
	if err := d.db.WithContext(ctx).Where("word = ?", word).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (d *wordDaoGorm) ListJapaneseWords(ctx context.Context, query string) ([]model.JapaneseWord, error) {
	var words []model.JapaneseWord
	q := d.db.WithContext(ctx).Order("created_at DESC")
	if query != "" {
		q = q.Where("word LIKE ? OR lemma LIKE ? OR kana LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}
	if err := q.Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
}

func (d *wordDaoGorm) DeleteJapaneseWord(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("word_type = ? AND word_id = ?", "ja", id).Delete(&model.WordSense{}).Error; err != nil {
			return err
		}
		if err := tx.Where("word_type = ? AND word_id = ?", "ja", id).Delete(&model.Example{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.JapaneseWord{}, id).Error
	})
}

func (d *wordDaoGorm) ReplaceSenses(ctx context.Context, wordType string, wordID int64, senses []model.WordSense) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("word_type = ? AND word_id = ?", wordType, wordID).Delete(&model.WordSense{}).Error; err != nil {
			return err
		}
		if len(senses) == 0 {
			return nil
		}
		for i := range senses {
			senses[i].WordType = wordType
			senses[i].WordID = wordID
		}
		return tx.Create(&senses).Error
	})
}

func (d *wordDaoGorm) ListSenses(ctx context.Context, wordType string, wordID int64) ([]model.WordSense, error) {
	var senses []model.WordSense
	if err := d.db.WithContext(ctx).
		Where("word_type = ? AND word_id = ?", wordType, wordID).
		Order("id ASC").
		Find(&senses).Error; err != nil {
		return nil, err
	}
	return senses, nil
}

func (d *wordDaoGorm) AddExample(ctx context.Context, e *model.Example) error {
	return d.db.WithContext(ctx).Create(e).Error
}

func (d *wordDaoGorm) ListExamples(ctx context.Context, wordType string, wordID int64) ([]model.Example, error) {
	var examples []model.Example
	if err := d.db.WithContext(ctx).
		Where("word_type = ? AND word_id = ?", wordType, wordID).
		Order("id DESC").
		Find(&examples).Error; err != nil {
		return nil, err
	}
	return examples, nil
}

// ListWordIDsBySong 返回某首歌曲下收藏过的所有生词 ID（去重）。
func (d *wordDaoGorm) ListWordIDsBySong(ctx context.Context, wordType string, songID int64) ([]int64, error) {
	var ids []int64
	if err := d.db.WithContext(ctx).
		Model(&model.Example{}).
		Where("word_type = ? AND song_id = ?", wordType, songID).
		Distinct("word_id").
		Pluck("word_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

var _ WordDao = (*wordDaoGorm)(nil)

var WordDaoProviderSet = wire.NewSet(
	NewWordDao,
	wire.Bind(new(WordDao), new(*wordDaoGorm)),
)
