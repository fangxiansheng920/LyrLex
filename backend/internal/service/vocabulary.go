package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/wire"

	"lyrics-server/internal/dao"
	"lyrics-server/internal/model"
)

// MeaningIn 加入生词本时勾选的词义。
type MeaningIn struct {
	POS     string `json:"pos"`
	Meaning string `json:"meaning"`
}

// ContextIn 语境信息。
type ContextIn struct {
	Text          string `json:"text"`
	TranslationZH string `json:"translation_zh"`
	SongID        int64  `json:"song_id"`
	LineIndex     int    `json:"line_index"`
}

// AddVocabularyReq 加入生词本参数。
type AddVocabularyReq struct {
	Language string      `json:"language"`
	Word     string      `json:"word"`
	Lemma    string      `json:"lemma"`
	Kana     string      `json:"kana"`
	Romaji   string      `json:"romaji"`
	Meanings []MeaningIn `json:"meanings"`
	Context  *ContextIn  `json:"context"`
}

// VocabularyItem 生词列表项。
type VocabularyItem struct {
	ID        int64     `json:"id"`
	Word      string    `json:"word"`
	Lemma     string    `json:"lemma"`
	Phonetic  string    `json:"phonetic"`
	Kana      string    `json:"kana"`
	Romaji    string    `json:"romaji"`
	CreatedAt time.Time `json:"created_at"`
}

// VocabularyDetail 生词详情。
type VocabularyDetail struct {
	ID        int64             `json:"id"`
	Word      string            `json:"word"`
	Lemma     string            `json:"lemma"`
	Phonetic  string            `json:"phonetic"`
	Kana      string            `json:"kana"`
	Romaji    string            `json:"romaji"`
	Meanings  []model.WordSense `json:"meanings"`
	Examples  []model.Example   `json:"examples"`
	CreatedAt time.Time         `json:"created_at"`
}

// VocabularyService 生词本业务（纯收藏）。
type VocabularyService interface {
	Add(ctx context.Context, r *AddVocabularyReq) (*VocabularyItem, error)
	List(ctx context.Context, lang, query string, songID int64) ([]VocabularyItem, error)
	Get(ctx context.Context, lang string, id int64) (*VocabularyDetail, error)
	Delete(ctx context.Context, lang string, id int64) error
}

type vocabularyService struct {
	dao dao.WordDao
}

var _ VocabularyService = (*vocabularyService)(nil)

func NewVocabularyService(d dao.WordDao) *vocabularyService {
	return &vocabularyService{dao: d}
}

func (s *vocabularyService) Add(ctx context.Context, r *AddVocabularyReq) (*VocabularyItem, error) {
	lemma := r.Lemma
	if lemma == "" {
		lemma = r.Word
	}

	var item *VocabularyItem
	var wordType string

	switch r.Language {
	case "en":
		w := &model.EnglishWord{Word: r.Word, Lemma: lemma, CreatedAt: time.Now()}
		if err := s.dao.UpsertEnglishWord(ctx, w); err != nil {
			return nil, err
		}
		item = &VocabularyItem{ID: w.ID, Word: w.Word, Lemma: w.Lemma, CreatedAt: w.CreatedAt}
		wordType = "en"
	case "ja":
		w := &model.JapaneseWord{
			Word:      r.Word,
			Lemma:     lemma,
			Kana:      r.Kana,
			Romaji:    r.Romaji,
			CreatedAt: time.Now(),
		}
		if err := s.dao.UpsertJapaneseWord(ctx, w); err != nil {
			return nil, err
		}
		item = &VocabularyItem{
			ID:        w.ID,
			Word:      w.Word,
			Lemma:     w.Lemma,
			Kana:      w.Kana,
			Romaji:    w.Romaji,
			CreatedAt: w.CreatedAt,
		}
		wordType = "ja"
	default:
		return nil, fmt.Errorf("不支持的语言 %q", r.Language)
	}

	if len(r.Meanings) > 0 {
		senses := make([]model.WordSense, 0, len(r.Meanings))
		for _, m := range r.Meanings {
			senses = append(senses, model.WordSense{POS: m.POS, Meaning: m.Meaning, Source: "lookup"})
		}
		if err := s.dao.ReplaceSenses(ctx, wordType, item.ID, senses); err != nil {
			return nil, err
		}
	}

	if r.Context != nil && r.Context.Text != "" {
		_ = s.dao.AddExample(ctx, &model.Example{
			WordType:      wordType,
			WordID:        item.ID,
			SongID:        r.Context.SongID,
			LineIndex:     r.Context.LineIndex,
			Text:          r.Context.Text,
			TranslationZH: r.Context.TranslationZH,
		})
	}
	return item, nil
}

func (s *vocabularyService) List(ctx context.Context, lang, query string, songID int64) ([]VocabularyItem, error) {
	var songFilter map[int64]bool
	if songID > 0 {
		ids, err := s.dao.ListWordIDsBySong(ctx, lang, songID)
		if err != nil {
			return nil, err
		}
		songFilter = make(map[int64]bool, len(ids))
		for _, id := range ids {
			songFilter[id] = true
		}
	}

	var items []VocabularyItem
	switch lang {
	case "en":
		words, err := s.dao.ListEnglishWords(ctx, query)
		if err != nil {
			return nil, err
		}
		items = make([]VocabularyItem, 0, len(words))
		for _, w := range words {
			if songFilter != nil && !songFilter[w.ID] {
				continue
			}
			phonetic := w.PhoneticUS
			if phonetic == "" {
				phonetic = w.PhoneticUK
			}
			items = append(items, VocabularyItem{
				ID:        w.ID,
				Word:      w.Word,
				Lemma:     w.Lemma,
				Phonetic:  phonetic,
				CreatedAt: w.CreatedAt,
			})
		}
	case "ja":
		words, err := s.dao.ListJapaneseWords(ctx, query)
		if err != nil {
			return nil, err
		}
		items = make([]VocabularyItem, 0, len(words))
		for _, w := range words {
			if songFilter != nil && !songFilter[w.ID] {
				continue
			}
			items = append(items, VocabularyItem{
				ID:        w.ID,
				Word:      w.Word,
				Lemma:     w.Lemma,
				Kana:      w.Kana,
				Romaji:    w.Romaji,
				CreatedAt: w.CreatedAt,
			})
		}
	default:
		return nil, fmt.Errorf("不支持的语言 %q", lang)
	}
	return items, nil
}

func (s *vocabularyService) Get(ctx context.Context, lang string, id int64) (*VocabularyDetail, error) {
	var detail *VocabularyDetail
	switch lang {
	case "en":
		w, err := s.dao.GetEnglishWord(ctx, id)
		if err != nil {
			return nil, err
		}
		phonetic := w.PhoneticUS
		if phonetic == "" {
			phonetic = w.PhoneticUK
		}
		detail = &VocabularyDetail{
			ID:        w.ID,
			Word:      w.Word,
			Lemma:     w.Lemma,
			Phonetic:  phonetic,
			CreatedAt: w.CreatedAt,
		}
	case "ja":
		w, err := s.dao.GetJapaneseWord(ctx, id)
		if err != nil {
			return nil, err
		}
		detail = &VocabularyDetail{
			ID:        w.ID,
			Word:      w.Word,
			Lemma:     w.Lemma,
			Kana:      w.Kana,
			Romaji:    w.Romaji,
			CreatedAt: w.CreatedAt,
		}
	default:
		return nil, fmt.Errorf("不支持的语言 %q", lang)
	}

	senses, _ := s.dao.ListSenses(ctx, lang, id)
	examples, _ := s.dao.ListExamples(ctx, lang, id)
	detail.Meanings = senses
	detail.Examples = examples
	return detail, nil
}

func (s *vocabularyService) Delete(ctx context.Context, lang string, id int64) error {
	switch lang {
	case "en":
		return s.dao.DeleteEnglishWord(ctx, id)
	case "ja":
		return s.dao.DeleteJapaneseWord(ctx, id)
	default:
		return fmt.Errorf("不支持的语言 %q", lang)
	}
}

var VocabularyProviderSet = wire.NewSet(
	NewVocabularyService,
	wire.Bind(new(VocabularyService), new(*vocabularyService)),
)
