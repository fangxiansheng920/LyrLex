// Package translation 整句翻译 Provider 实现。
package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/wire"

	"lyrics-server/internal/provider"
)

// MyMemoryTranslationProvider 免费无 key 翻译源（个人使用）。
// 后续可新增 Youdao/Baidu/DeepL Provider，通过配置切换。
type MyMemoryTranslationProvider struct {
	client *http.Client
}

func NewMyMemoryTranslationProvider() *MyMemoryTranslationProvider {
	return &MyMemoryTranslationProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *MyMemoryTranslationProvider) Name() string { return "mymemory" }

func (p *MyMemoryTranslationProvider) Translate(ctx context.Context, text, source, target string) (string, error) {
	u := "https://api.mymemory.translated.net/get?q=" + url.QueryEscape(text) +
		"&langpair=" + url.QueryEscape(source+"|"+target)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("mymemory request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mymemory http status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("mymemory read: %w", err)
	}

	var parsed struct {
		ResponseStatus interface{} `json:"responseStatus"`
		ResponseData   struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("mymemory parse: %w", err)
	}

	status := 0
	switch v := parsed.ResponseStatus.(type) {
	case float64:
		status = int(v)
	case string:
		_, _ = fmt.Sscanf(v, "%d", &status)
	}
	if status != 200 {
		return "", fmt.Errorf("mymemory status: %v", parsed.ResponseStatus)
	}
	if parsed.ResponseData.TranslatedText == "" {
		return "", fmt.Errorf("mymemory empty translation")
	}
	return parsed.ResponseData.TranslatedText, nil
}

var ProviderSet = wire.NewSet(
	NewMyMemoryTranslationProvider,
	wire.Bind(new(provider.TranslationProvider), new(*MyMemoryTranslationProvider)),
)
