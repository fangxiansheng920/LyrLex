// Package provider 定义所有外部能力（翻译、词典、分词、发音、歌词）的统一接口。
// 具体实现从 M2 起逐步落地，业务层只依赖这些接口。
package provider

import "context"

// Meaning 一条词义。
type Meaning struct {
	POS     string `json:"pos"`
	Meaning string `json:"meaning"`
}

// DictEntry 词典查询结果。
type DictEntry struct {
	Surface   string    `json:"surface"`   // 查询原词
	Lemma     string    `json:"lemma"`     // 基本形
	POS       string    `json:"pos"`       // 词性
	Phonetic  string    `json:"phonetic"`  // 英文 IPA
	Kana      string    `json:"kana"`      // 日文假名
	Romaji    string    `json:"romaji"`    // 日文罗马音
	Accent    string    `json:"accent"`    // 日文声调（可空）
	Meanings  []Meaning `json:"meanings"`  // 词义列表
	AudioURLs []string  `json:"audio_urls"` // 音频源地址（可空）
	Source    string    `json:"source"`    // 数据来源标识
}

// Token 分词结果。
type Token struct {
	Surface string `json:"surface"` // 原形文本
	Lemma   string `json:"lemma"`   // 基本形
	POS     string `json:"pos"`     // 词性
	Reading string `json:"reading"` // 读音（日文假名/英文音标）
	Romaji  string `json:"romaji"`  // 日文罗马音（可空）
	Accent  string `json:"accent"`  // 声调（可空）
	GroupID int    `json:"group_id"` // 合并词组 ID（日文功能词合并；0 = 独立）
	Addable bool   `json:"addable"`  // 是否可加入生词本（助词/标点为 false）
}

// SongMeta 在线搜索到的歌曲元信息。
type SongMeta struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

// LyricLine 一行歌词（可带已有翻译）。
type LyricLine struct {
	Text        string `json:"text"`
	Translation string `json:"translation"`
}

// Lyrics 完整歌词。
type Lyrics struct {
	SongID string      `json:"song_id"`
	Lines  []LyricLine `json:"lines"`
}

// TranslationProvider 整句翻译。
type TranslationProvider interface {
	Name() string
	Translate(ctx context.Context, text, source, target string) (string, error)
}

// DictionaryProvider 词典查询（词形还原 + 音标 + 释义 + 音频地址）。
type DictionaryProvider interface {
	Name() string
	Lookup(ctx context.Context, lang string, word string) (*DictEntry, error)
}

// TokenizerProvider 分词（英文 tokenizer / 日文形态分析）。
type TokenizerProvider interface {
	Name() string
	Tokenize(ctx context.Context, lang string, line string) ([]Token, error)
}

// PronunciationProvider 发音（返回音频二进制）。
type PronunciationProvider interface {
	Name() string
	Fetch(ctx context.Context, lang string, word string) ([]byte, error)
}

// LyricsProvider 歌词搜索。
type LyricsProvider interface {
	Name() string
	Search(ctx context.Context, query string) ([]SongMeta, error)
	GetLyrics(ctx context.Context, songID string) (*Lyrics, error)
}
