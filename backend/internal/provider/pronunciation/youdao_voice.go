// Package pronunciation 发音 Provider 实现。
package pronunciation

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/wire"

	"lyrics-server/internal/provider"
)

// YoudaoDictVoiceProvider 有道词典发音接口（免费无 key，英文可靠）。
// 后续可加 Edge TTS 兜底。
type YoudaoDictVoiceProvider struct {
	client *http.Client
}

func NewYoudaoDictVoiceProvider() *YoudaoDictVoiceProvider {
	return &YoudaoDictVoiceProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *YoudaoDictVoiceProvider) Name() string { return "youdao-dictvoice" }

func (p *YoudaoDictVoiceProvider) Fetch(ctx context.Context, lang, word string) ([]byte, error) {
	// type=1 英音 / type=2 美音；日文用 le=jap
	var u string
	switch lang {
	case "en":
		u = "https://dict.youdao.com/dictvoice?audio=" + url.QueryEscape(word) + "&type=2"
	case "ja":
		u = "https://dict.youdao.com/dictvoice?audio=" + url.QueryEscape(word) + "&le=jap"
	default:
		return nil, fmt.Errorf("youdao dictvoice: unsupported language %q", lang)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youdao dictvoice request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youdao dictvoice http status: %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 5<<20))
}

var ProviderSet = wire.NewSet(
	NewYoudaoDictVoiceProvider,
	wire.Bind(new(provider.PronunciationProvider), new(*YoudaoDictVoiceProvider)),
)
