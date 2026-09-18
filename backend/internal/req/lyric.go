package req

// ParseReq 解析歌词请求。
type ParseReq struct {
	Text string `json:"text" binding:"required"`
}

// TranslateReq 批量翻译请求。
type TranslateReq struct {
	Source string   `json:"source" binding:"required"`
	Target string   `json:"target" binding:"required"`
	Lines  []string `json:"lines" binding:"required,min=1"`
}

// TokenizeReq 分词请求。
type TokenizeReq struct {
	Language string `json:"language" binding:"required,oneof=en ja"`
	Line     string `json:"line" binding:"required"`
}

// LookupReq 查词请求。
type LookupReq struct {
	Language string `json:"language" binding:"required,oneof=en ja"`
	Word     string `json:"word" binding:"required"`
	Context  string `json:"context"`
}

// MeaningIn 勾选的词义。
type MeaningIn struct {
	POS     string `json:"pos"`
	Meaning string `json:"meaning"`
}

// ContextIn 加入生词本时的语境信息。
type ContextIn struct {
	Text          string `json:"text"`
	TranslationZH string `json:"translation_zh"`
	SongID        int64  `json:"song_id"`
	LineIndex     int    `json:"line_index"`
}

// AddVocabularyReq 加入生词本请求。
type AddVocabularyReq struct {
	Language string      `json:"language" binding:"required,oneof=en ja"`
	Word     string      `json:"word" binding:"required"`
	Lemma    string      `json:"lemma"`
	Kana     string      `json:"kana"`
	Romaji   string      `json:"romaji"`
	Meanings []MeaningIn `json:"meanings"`
	Context  *ContextIn  `json:"context"`
}

// SongLineIn 保存歌曲时的行数据。
type SongLineIn struct {
	Text          string `json:"text" binding:"required"`
	TranslationZH string `json:"translation_zh"`
	Tokens        []struct {
		Surface string `json:"surface"`
		Lemma   string `json:"lemma"`
		POS     string `json:"pos"`
		Reading string `json:"reading"`
		Romaji  string `json:"romaji"`
		Accent  string `json:"accent"`
		GroupID int    `json:"group_id"`
		Addable bool   `json:"addable"`
	} `json:"tokens"`
}

// SaveSongReq 保存歌曲请求。
type SaveSongReq struct {
	Title    string       `json:"title"`
	Artist   string       `json:"artist"`
	Language string       `json:"language" binding:"required,oneof=en ja"`
	Lines    []SongLineIn `json:"lines" binding:"required,min=1"`
}
