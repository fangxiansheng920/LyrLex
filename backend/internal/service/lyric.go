package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/wire"

	"lyrics-server/internal/config"
	"lyrics-server/internal/dao"
	"lyrics-server/internal/model"
	"lyrics-server/internal/provider"
)

// ParseResult 歌词解析结果。
type ParseResult struct {
	Language model.Language `json:"language"`
	Lines    []string       `json:"lines"`
}

// SongLineIn 保存歌曲时的行数据。
type SongLineIn struct {
	Text          string           `json:"text"`
	TranslationZH string           `json:"translation_zh"`
	Tokens        []provider.Token `json:"tokens"`
}

// LyricService 歌词学习相关业务。
type LyricService interface {
	Parse(ctx context.Context, text string) (*ParseResult, error)
	Translate(ctx context.Context, source, target string, lines []string) ([]string, error)
	Tokenize(ctx context.Context, lang model.Language, line string) ([]provider.Token, error)
	SaveSong(ctx context.Context, title, artist string, lang model.Language, lines []SongLineIn) (*model.Song, error)
	GetSong(ctx context.Context, id int64) (*model.Song, error)
	GetSongLines(ctx context.Context, songID int64) ([]model.SongLine, error)
	ListSongs(ctx context.Context, query string) ([]model.Song, error)
	SearchSongs(ctx context.Context, query string) ([]provider.SongMeta, error)
	ImportSong(ctx context.Context, meta provider.SongMeta) (*model.Song, error)
}

type lyricService struct {
	dao    dao.LyricDao
	trans  provider.TranslationProvider
	tok    provider.TokenizerProvider
	lyrics provider.LyricsProvider
	cfg    *config.Config
}

var _ LyricService = (*lyricService)(nil)

func NewLyricService(
	d dao.LyricDao,
	trans provider.TranslationProvider,
	tok provider.TokenizerProvider,
	lyrics provider.LyricsProvider,
	cfg *config.Config,
) *lyricService {
	return &lyricService{
		dao:    d,
		trans:  trans,
		tok:    tok,
		lyrics: lyrics,
		cfg:    cfg,
	}
}

func (s *lyricService) Parse(ctx context.Context, text string) (*ParseResult, error) {
	lines := splitLines(text)
	if len(lines) == 0 {
		return nil, fmt.Errorf("歌词为空")
	}
	return &ParseResult{Language: detectLanguage(lines), Lines: lines}, nil
}

func (s *lyricService) Translate(ctx context.Context, source, target string, lines []string) ([]string, error) {
	results := make([]string, len(lines))
	for i, line := range lines {
		if !needsTranslation(line) {
			results[i] = ""
			continue
		}

		hash := translationHash(source, target, line)
		if tc, err := s.dao.GetTranslation(ctx, hash); err == nil {
			results[i] = tc.Result
			continue
		}

		translated, err := s.trans.Translate(ctx, line, source, target)
		if err != nil {
			// 单行失败降级为空，不阻断整首歌
			results[i] = ""
			continue
		}
		results[i] = translated
		_ = s.dao.SetTranslation(ctx, &model.TranslationCache{
			Hash:      hash,
			Source:    source,
			Target:    target,
			Text:      line,
			Result:    translated,
			UpdatedAt: time.Now(),
		})
	}
	return results, nil
}

func (s *lyricService) Tokenize(ctx context.Context, lang model.Language, line string) ([]provider.Token, error) {
	return s.tok.Tokenize(ctx, string(lang), line)
}

func (s *lyricService) SaveSong(ctx context.Context, title, artist string, lang model.Language, lines []SongLineIn) (*model.Song, error) {
	return s.saveSong(ctx, title, artist, lang, "manual", lines)
}

// SearchSongs 在线搜索歌曲（LyricsProvider）。
func (s *lyricService) SearchSongs(ctx context.Context, query string) ([]provider.SongMeta, error) {
	return s.lyrics.Search(ctx, query)
}

// ImportSong 从在线源拉取歌词并保存为歌曲。
func (s *lyricService) ImportSong(ctx context.Context, meta provider.SongMeta) (*model.Song, error) {
	lyrics, err := s.lyrics.GetLyrics(ctx, meta.ID)
	if err != nil {
		return nil, err
	}
	texts := make([]string, 0, len(lyrics.Lines))
	lines := make([]SongLineIn, 0, len(lyrics.Lines))
	for _, l := range lyrics.Lines {
		texts = append(texts, l.Text)
		lines = append(lines, SongLineIn{Text: l.Text, TranslationZH: l.Translation})
	}
	return s.saveSong(ctx, meta.Title, meta.Artist, detectLanguage(texts), "online", lines)
}

func (s *lyricService) saveSong(ctx context.Context, title, artist string, lang model.Language, source string, lines []SongLineIn) (*model.Song, error) {
	song := &model.Song{
		Title:    title,
		Artist:   artist,
		Language: string(lang),
		Source:   source,
	}
	modelLines := make([]model.SongLine, 0, len(lines))
	for i, l := range lines {
		tokensJSON, _ := json.Marshal(l.Tokens)
		modelLines = append(modelLines, model.SongLine{
			LineIndex:     i,
			Text:          l.Text,
			TranslationZH: l.TranslationZH,
			TokensJSON:    string(tokensJSON),
		})
	}
	if err := s.dao.SaveSong(ctx, song, modelLines); err != nil {
		return nil, err
	}
	return song, nil
}

func (s *lyricService) GetSong(ctx context.Context, id int64) (*model.Song, error) {
	return s.dao.GetSong(ctx, id)
}

func (s *lyricService) GetSongLines(ctx context.Context, songID int64) ([]model.SongLine, error) {
	return s.dao.GetSongLines(ctx, songID)
}

func (s *lyricService) ListSongs(ctx context.Context, query string) ([]model.Song, error) {
	return s.dao.ListSongs(ctx, query)
}

func splitLines(text string) []string {
	raw := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(raw))
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func detectLanguage(lines []string) model.Language {
	var ja, en int
	for _, l := range lines {
		for _, r := range l {
			switch {
			case r >= 0x3040 && r <= 0x30FF:
				ja++
			case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
				en++
			}
		}
	}
	if ja > en {
		return model.LangJA
	}
	return model.LangEN
}

// needsTranslation 若一行大部分是汉字，视为已带中文翻译，跳过 API 翻译。
func needsTranslation(line string) bool {
	var cjk, total int
	for _, r := range line {
		total++
		if unicode.Is(unicode.Han, r) {
			cjk++
		}
	}
	if total == 0 {
		return false
	}
	return float64(cjk)/float64(total) < 0.5
}

func translationHash(source, target, text string) string {
	sum := sha256.Sum256([]byte(source + "\x00" + target + "\x00" + text))
	return hex.EncodeToString(sum[:])
}

var LyricProviderSet = wire.NewSet(
	NewLyricService,
	wire.Bind(new(LyricService), new(*lyricService)),
)
