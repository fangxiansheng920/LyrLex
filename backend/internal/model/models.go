package model

import "time"

type Language string

const (
	LangEN Language = "en"
	LangJA Language = "ja"
)

// Song 本地歌曲库。
type Song struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"index" json:"title"`
	Artist    string    `gorm:"index" json:"artist"`
	Language  string    `json:"language"` // "en" | "ja"
	Source    string    `json:"source"`   // "online" | "manual"
	CreatedAt time.Time `json:"created_at"`
}

// SongLine 歌词行（含翻译与分词缓存）。
type SongLine struct {
	ID            int64  `gorm:"primaryKey" json:"id"`
	SongID        int64  `gorm:"index" json:"song_id"`
	LineIndex     int    `json:"line_index"`
	Text          string `json:"text"`           // 原始行
	TranslationZH string `json:"translation_zh"` // 整句翻译（可空）
	TokensJSON    string `json:"tokens_json"`    // 分词结果缓存
}

// EnglishWord 英文生词（纯收藏）。
type EnglishWord struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	Word       string    `gorm:"uniqueIndex" json:"word"` // 入库词（默认基本形）
	Lemma      string    `json:"lemma"`                   // 词典基本形
	PhoneticUK string    `json:"phonetic_uk"`             // 英式 IPA
	PhoneticUS string    `json:"phonetic_us"`             // 美式 IPA
	AudioPath  string    `json:"audio_path"`              // 本地音频缓存路径（可空）
	CreatedAt  time.Time `json:"created_at"`
}

// JapaneseWord 日文生词（纯收藏）。
type JapaneseWord struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Word      string    `gorm:"uniqueIndex" json:"word"` // 入库词（默认基本形）
	Lemma     string    `json:"lemma"`                   // 基本形
	Kana      string    `json:"kana"`                    // 假名
	Romaji    string    `json:"romaji"`                  // 罗马音
	Accent    string    `json:"accent"`                  // 声调（可空）
	POS       string    `json:"pos"`                     // 词性
	AudioPath string    `json:"audio_path"`              // 本地音频缓存路径（可空）
	CreatedAt time.Time `json:"created_at"`
}

// WordSense 词义（英/日共享）。
type WordSense struct {
	ID       int64  `gorm:"primaryKey" json:"id"`
	WordType string `gorm:"index" json:"word_type"` // "en" | "ja"
	WordID   int64  `gorm:"index" json:"word_id"`
	POS      string `json:"pos"`
	Meaning  string `json:"meaning"` // 中文释义
	Source   string `json:"source"`  // 来源词典/API
}

// Example 语境例句（来自哪首歌哪一句）。
type Example struct {
	ID            int64  `gorm:"primaryKey" json:"id"`
	WordType      string `gorm:"index" json:"word_type"` // "en" | "ja"
	WordID        int64  `gorm:"index" json:"word_id"`
	SenseID       int64  `json:"sense_id"` // 用户勾选的词义（可空）
	SongID        int64  `gorm:"index" json:"song_id"`
	LineIndex     int    `json:"line_index"`
	Text          string `json:"text"`           // 原句
	TranslationZH string `json:"translation_zh"` // 原句翻译
}

// WordCache 词查询结果缓存。
type WordCache struct {
	Lang      string    `gorm:"uniqueIndex:idx_lang_word" json:"lang"`
	Word      string    `gorm:"uniqueIndex:idx_lang_word" json:"word"`
	Payload   string    `json:"payload"` // DictEntry JSON
	UpdatedAt time.Time `json:"updated_at"`
}

// TranslationCache 整句翻译缓存。
type TranslationCache struct {
	Hash      string    `gorm:"uniqueIndex" json:"hash"` // sha256(source+target+text)
	Source    string    `json:"source"`
	Target    string    `json:"target"`
	Text      string    `json:"text"`
	Result    string    `json:"result"`
	UpdatedAt time.Time `json:"updated_at"`
}
