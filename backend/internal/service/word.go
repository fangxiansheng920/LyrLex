package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
	"unicode"

	"github.com/google/wire"

	"lyrics-server/internal/config"
	"lyrics-server/internal/dao"
	"lyrics-server/internal/model"
	"lyrics-server/internal/provider"
)

// LookupResult 查词结果。
type LookupResult struct {
	Surface  string            `json:"surface"`
	Lemma    string            `json:"lemma"`
	Phonetic string            `json:"phonetic"`
	Kana     string            `json:"kana"`
	Romaji   string            `json:"romaji"`
	Meanings []provider.Meaning `json:"meanings"`
	Audio    bool              `json:"audio"`
	Source   string            `json:"source"`
}

// WordService 单词查询与发音。
type WordService interface {
	Lookup(ctx context.Context, lang, word, contextText string, force bool) (*LookupResult, error)
	Audio(ctx context.Context, lang, word string) ([]byte, error)
}

type wordService struct {
	dao   dao.WordDao
	dict  provider.DictionaryProvider
	tok   provider.TokenizerProvider
	pron  provider.PronunciationProvider
	trans provider.TranslationProvider
	cfg   *config.Config
}

var _ WordService = (*wordService)(nil)

func NewWordService(
	d dao.WordDao,
	dict provider.DictionaryProvider,
	tok provider.TokenizerProvider,
	pron provider.PronunciationProvider,
	trans provider.TranslationProvider,
	cfg *config.Config,
) *wordService {
	return &wordService{dao: d, dict: dict, tok: tok, pron: pron, trans: trans, cfg: cfg}
}

func (s *wordService) Lookup(ctx context.Context, lang, word, contextText string, force bool) (*LookupResult, error) {
	// 命中本地缓存（force=true 时跳过，强制重新查询词典）
	if !force {
		if wc, err := s.dao.GetWordCache(ctx, lang, word); err == nil {
			var entry provider.DictEntry
			if json.Unmarshal([]byte(wc.Payload), &entry) == nil {
				return toLookupResult(lang, &entry), nil
			}
		}
	}

	entry, err := s.dict.Lookup(ctx, lang, word)
	if err != nil {
		// 日文词典不可达时，用本地分词结果兜底：至少给出假名与罗马音
		if lang == "ja" {
			if minimal := s.jaMinimalEntry(ctx, word); minimal != nil {
				return toLookupResult(lang, minimal), nil
			}
		}
		return nil, err
	}

	// 把非中文释义翻成中文（有道优先，失败回退免费源）
	s.localizeMeanings(ctx, entry)

	if payload, err := json.Marshal(entry); err == nil {
		_ = s.dao.SetWordCache(ctx, &model.WordCache{
			Lang:      lang,
			Word:      word,
			Payload:   string(payload),
			UpdatedAt: time.Now(),
		})
	}
	return toLookupResult(lang, entry), nil
}

// localizeMeanings 把非中文释义翻译成中文（有道优先，失败回退免费源）。
func (s *wordService) localizeMeanings(ctx context.Context, entry *provider.DictEntry) {
	const maxMeanings = 4
	for i := range entry.Meanings {
		if i >= maxMeanings {
			break
		}
		m := &entry.Meanings[i]
		if m.Meaning == "" || isMostlyChinese(m.Meaning) {
			continue
		}
		if zh, err := s.trans.Translate(ctx, m.Meaning, "en", "zh-CN"); err == nil && zh != "" {
			m.Meaning = zh
		}
	}
}

func isMostlyChinese(s string) bool {
	var han, total int
	for _, r := range s {
		total++
		if unicode.Is(unicode.Han, r) {
			han++
		}
	}
	if total == 0 {
		return false
	}
	return float64(han)/float64(total) >= 0.3
}

// jaMinimalEntry 用 kagome 对单词本身分词，构造最小词典条目。
func (s *wordService) jaMinimalEntry(ctx context.Context, word string) *provider.DictEntry {
	tokens, err := s.tok.Tokenize(ctx, "ja", word)
	if err != nil {
		return nil
	}
	for _, t := range tokens {
		if !t.Addable {
			continue
		}
		return &provider.DictEntry{
			Surface: word,
			Lemma:   t.Lemma,
			Kana:    t.Reading,
			Romaji:  t.Romaji,
			Source:  "kagome",
		}
	}
	return nil
}

func toLookupResult(lang string, e *provider.DictEntry) *LookupResult {
	return &LookupResult{
		Surface:  e.Surface,
		Lemma:    e.Lemma,
		Phonetic: e.Phonetic,
		Kana:     e.Kana,
		Romaji:   e.Romaji,
		Meanings: e.Meanings,
		Audio:    lang == "en" || lang == "ja",
		Source:   e.Source,
	}
}

func (s *wordService) Audio(ctx context.Context, lang, word string) ([]byte, error) {
	dir := filepath.Join(s.cfg.DataDir(), "audio", lang)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, audioName(lang, word)+".mp3")

	if b, err := os.ReadFile(path); err == nil {
		return b, nil
	}

	b, err := s.pron.Fetch(ctx, lang, word)
	if err != nil {
		return nil, err
	}
	_ = os.WriteFile(path, b, 0o644) // 缓存失败不阻断发音
	return b, nil
}

func audioName(lang, word string) string {
	sum := sha256.Sum256([]byte(lang + "\x00" + word))
	return hex.EncodeToString(sum[:])
}

var WordProviderSet = wire.NewSet(
	NewWordService,
	wire.Bind(new(WordService), new(*wordService)),
)
