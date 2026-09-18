package resp

import "lyrics-server/internal/provider"

// ParseResp 歌词解析响应。
type ParseResp struct {
	Language string   `json:"language"`
	Lines    []string `json:"lines"`
}

// TranslateResp 批量翻译响应。
type TranslateResp struct {
	Translations []string `json:"translations"`
}

// TokenizeResp 分词响应。
type TokenizeResp struct {
	Tokens []provider.Token `json:"tokens"`
}
