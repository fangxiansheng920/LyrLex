package req

// UpdateSettingsReq 更新设置请求（字段可缺省，仅更新传入项）。
type UpdateSettingsReq struct {
	APIKeys             *APIConfigIn `json:"api_keys"`
	PronunciationSource string       `json:"pronunciation_source"`
}

// APIConfigIn API 密钥配置。
type APIConfigIn struct {
	YoudaoAppKey    string `json:"youdao_app_key"`
	YoudaoAppSecret string `json:"youdao_app_secret"`
	BaiduAppID      string `json:"baidu_app_id"`
	BaiduSecret     string `json:"baidu_secret"`
	DeepLKey        string `json:"deepl_key"`
}
