// Package dictionary 词典查询 Provider 实现（英文优先）。
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

// FreeDictionaryProvider 使用 dictionaryapi.dev（免费无 key），
// 词形还原用 snowball 词干兜底；ECDICT 本地词典后续接入作为离线优先层。
type FreeDictionaryProvider struct {
	client *http.Client
}

func NewFreeDictionaryProvider() *FreeDictionaryProvider {
	return &FreeDictionaryProvider{
		client: &http.Client{Timeout: 6 * time.Second},
	}
}

func (p *FreeDictionaryProvider) Name() string { return "freedict" }

func (p *FreeDictionaryProvider) Lookup(ctx context.Context, lang, word string) (*provider.DictEntry, error) {
	if lang != "en" {
		return nil, fmt.Errorf("freedict: unsupported language %q", lang)
	}

	entry := p.fetch(ctx, word)
	if entry == nil {
		lemma := provider.LemmaEN(word)
		if lemma != "" && lemma != strings.ToLower(word) {
			entry = p.fetch(ctx, lemma)
		}
	}
	if entry == nil {
		// dictionaryapi.dev 不可达时，用百度翻译 sug 兜底（国内可达、免费无 key）
		entry = p.baiduSug(ctx, word)
	}
	if entry == nil {
		return nil, fmt.Errorf("freedict: word not found: %s", word)
	}
	entry.Surface = word
	return entry, nil
}

func (p *FreeDictionaryProvider) baiduSug(ctx context.Context, word string) *provider.DictEntry {
	form := url.Values{}
	form.Set("kw", word)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://fanyi.baidu.com/sug", strings.NewReader(form.Encode()))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := p.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil
	}

	var parsed struct {
		Errno int `json:"errno"`
		Data  []struct {
			K string `json:"k"`
			V string `json:"v"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Errno != 0 || len(parsed.Data) == 0 {
		return nil
	}
	meaning := parsed.Data[0].V
	if meaning == "" {
		return nil
	}
	return &provider.DictEntry{
		Surface:  word,
		Lemma:    provider.LemmaEN(word),
		Meanings: []provider.Meaning{{Meaning: meaning}},
		Source:   "baidu-sug",
	}
}

func (p *FreeDictionaryProvider) fetch(ctx context.Context, word string) *provider.DictEntry {
	u := "https://api.dictionaryapi.dev/api/v2/entries/en/" + url.PathEscape(word)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	resp, err := p.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil
	}

	var entries []struct {
		Word      string `json:"word"`
		Phonetic  string `json:"phonetic"`
		Phonetics []struct {
			Text  string `json:"text"`
			Audio string `json:"audio"`
		} `json:"phonetics"`
		Meanings []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions  []struct {
				Definition string `json:"definition"`
			} `json:"definitions"`
		} `json:"meanings"`
	}
	if err := json.Unmarshal(body, &entries); err != nil || len(entries) == 0 {
		return nil
	}
	e := entries[0]

	entry := &provider.DictEntry{
		Surface:  word,
		Lemma:    e.Word,
		Phonetic: e.Phonetic,
		Source:   "freedict",
	}
	for _, ph := range e.Phonetics {
		if entry.Phonetic == "" && ph.Text != "" {
			entry.Phonetic = ph.Text
		}
		if ph.Audio != "" {
			entry.AudioURLs = append(entry.AudioURLs, ph.Audio)
		}
	}
	for _, m := range e.Meanings {
		for _, d := range m.Definitions {
			entry.Meanings = append(entry.Meanings, provider.Meaning{
				POS:     m.PartOfSpeech,
				Meaning: d.Definition,
			})
		}
	}
	if len(entry.Meanings) == 0 {
		return nil
	}
	return entry
}
