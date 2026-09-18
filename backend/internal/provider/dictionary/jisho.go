package dictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lyrics-server/internal/provider"
)

// JishoDictionaryProvider 使用 jisho.org 免费 API 获取日文读音与英文释义；
// 中文翻译统一由上层翻译源处理（有道优先，回退免费源）。
type JishoDictionaryProvider struct {
	client *http.Client
}

func NewJishoDictionaryProvider() *JishoDictionaryProvider {
	return &JishoDictionaryProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *JishoDictionaryProvider) Name() string { return "jisho" }

func (p *JishoDictionaryProvider) Lookup(ctx context.Context, lang, word string) (*provider.DictEntry, error) {
	if lang != "ja" {
		return nil, fmt.Errorf("jisho: unsupported language %q", lang)
	}

	u := "https://jisho.org/api/v1/search/words?keyword=" + url.QueryEscape(word)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jisho request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jisho http status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Data []struct {
			Slug     string `json:"slug"`
			Japanese []struct {
				Word    string `json:"word"`
				Reading string `json:"reading"`
			} `json:"japanese"`
			Senses []struct {
				PartsOfSpeech      []string `json:"parts_of_speech"`
				EnglishDefinitions []string `json:"english_definitions"`
			} `json:"senses"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || len(parsed.Data) == 0 {
		return nil, fmt.Errorf("jisho: word not found: %s", word)
	}

	d := parsed.Data[0]
	kana := ""
	for _, j := range d.Japanese {
		if j.Reading != "" {
			kana = j.Reading
			break
		}
	}

	entry := &provider.DictEntry{
		Surface: word,
		Lemma:   d.Slug,
		Kana:    kana,
		Romaji:  provider.RomajiJA(kana),
		Source:  "jisho",
	}

	for i, s := range d.Senses {
		if i >= 3 {
			break
		}
		def := strings.Join(s.EnglishDefinitions, "; ")
		if def == "" {
			continue
		}
		entry.Meanings = append(entry.Meanings, provider.Meaning{
			POS:     strings.Join(s.PartsOfSpeech, "/"),
			Meaning: def,
		})
	}
	return entry, nil
}
