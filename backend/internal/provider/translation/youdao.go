package translation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"lyrics-server/internal/provider"
)

const youdaoAPI = "https://openapi.youdao.com/api"

// APIError 有道接口返回的错误码（如 108/110/202/401/411）。
type APIError struct {
	Code string
}

func (e *APIError) Error() string { return "youdao errorCode=" + e.Code }

// YoudaoTranslationProvider 有道智云文本翻译（NMT）。
// 密钥来自设置中心，填写后即时生效；未配置时由 Composite 回退到免费源。
type YoudaoTranslationProvider struct {
	client *http.Client
	keys   provider.CredentialsProvider
}

func NewYoudaoTranslationProvider(keys provider.CredentialsProvider) *YoudaoTranslationProvider {
	return &YoudaoTranslationProvider{
		client: &http.Client{Timeout: 10 * time.Second},
		keys:   keys,
	}
}

func (p *YoudaoTranslationProvider) Name() string { return "youdao" }

// HasKey 是否已配置有道密钥。
func (p *YoudaoTranslationProvider) HasKey() bool {
	appKey, appSecret := p.keys.YoudaoKeys()
	return appKey != "" && appSecret != ""
}

func (p *YoudaoTranslationProvider) Translate(ctx context.Context, text, source, target string) (string, error) {
	appKey, appSecret := p.keys.YoudaoKeys()
	return doTranslate(ctx, p.client, appKey, appSecret, text, source, target)
}

var defaultHTTPClient = &http.Client{Timeout: 10 * time.Second}

// TranslateWithKeys 用指定密钥直接翻译（供设置页「测试」使用，不经过 wire 依赖图）。
func TranslateWithKeys(ctx context.Context, appKey, appSecret, text, source, target string) (string, error) {
	return doTranslate(ctx, defaultHTTPClient, appKey, appSecret, text, source, target)
}

func doTranslate(ctx context.Context, client *http.Client, appKey, appSecret, text, source, target string) (string, error) {
	if appKey == "" || appSecret == "" {
		return "", fmt.Errorf("youdao: 未配置应用 ID/密钥")
	}

	salt := randomSalt()
	curtime := strconv.FormatInt(time.Now().Unix(), 10)
	sign := sha256Hex(appKey + truncateQ(text) + salt + curtime + appSecret)

	form := url.Values{}
	form.Set("q", text)
	form.Set("from", mapLangFrom(source))
	form.Set("to", mapLangTo(target))
	form.Set("appKey", appKey)
	form.Set("salt", salt)
	form.Set("sign", sign)
	form.Set("signType", "v3")
	form.Set("curtime", curtime)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, youdaoAPI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("youdao request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("youdao http status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	var parsed struct {
		ErrorCode   string   `json:"errorCode"`
		Translation []string `json:"translation"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("youdao parse: %w", err)
	}
	if parsed.ErrorCode != "0" {
		return "", &APIError{Code: parsed.ErrorCode}
	}
	if len(parsed.Translation) == 0 {
		return "", fmt.Errorf("youdao: 译文为空")
	}
	return parsed.Translation[0], nil
}

// truncateQ 有道签名规则：长度 > 20 时取「前10 + 长度 + 后10」。
func truncateQ(q string) string {
	runes := []rune(q)
	if len(runes) <= 20 {
		return q
	}
	return string(runes[:10]) + strconv.Itoa(len(runes)) + string(runes[len(runes)-10:])
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func randomSalt() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func mapLangFrom(source string) string {
	switch strings.ToLower(source) {
	case "en":
		return "en"
	case "ja":
		return "ja"
	default:
		return "auto"
	}
}

func mapLangTo(target string) string {
	// 有道简体中文语言码为 zh-CHS
	return "zh-CHS"
}
