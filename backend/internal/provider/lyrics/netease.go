// Package lyrics 歌词搜索 Provider 实现。
package lyrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"lyrics-server/internal/provider"
)

var lrcTimeRe = regexp.MustCompile(`^\[\d{1,2}:\d{2}(?:\.\d+)?\]\s*`)

// NeteaseCloudMusicApiProvider 调用 NeteaseCloudMusicApi 服务（默认 http://localhost:3000）。
// 启动方式见 README（Docker 一行命令）；地址可用 NETEASE_API_BASE 覆盖。
type NeteaseCloudMusicApiProvider struct {
	client *http.Client
	base   string
}

func NewNeteaseCloudMusicApiProvider() *NeteaseCloudMusicApiProvider {
	base := strings.TrimRight(os.Getenv("NETEASE_API_BASE"), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	return &NeteaseCloudMusicApiProvider{
		client: &http.Client{Timeout: 10 * time.Second},
		base:   base,
	}
}

func (p *NeteaseCloudMusicApiProvider) Name() string { return "netease-cloud-music-api" }

func (p *NeteaseCloudMusicApiProvider) Search(ctx context.Context, query string) ([]provider.SongMeta, error) {
	u := p.base + "/search?keywords=" + url.QueryEscape(query) + "&limit=10"
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Result struct {
			Songs []struct {
				ID      int    `json:"id"`
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
			} `json:"songs"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("netease search parse: %w", err)
	}

	metas := make([]provider.SongMeta, 0, len(parsed.Result.Songs))
	for _, s := range parsed.Result.Songs {
		artist := ""
		if len(s.Artists) > 0 {
			names := make([]string, 0, len(s.Artists))
			for _, a := range s.Artists {
				names = append(names, a.Name)
			}
			artist = strings.Join(names, "/")
		}
		metas = append(metas, provider.SongMeta{
			ID:     strconv.Itoa(s.ID),
			Title:  s.Name,
			Artist: artist,
		})
	}
	return metas, nil
}

func (p *NeteaseCloudMusicApiProvider) GetLyrics(ctx context.Context, songID string) (*provider.Lyrics, error) {
	u := p.base + "/lyric?id=" + url.QueryEscape(songID)
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Lrc struct {
			Lyric string `json:"lyric"`
		} `json:"lrc"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("netease lyric parse: %w", err)
	}
	lines := parseLRC(parsed.Lrc.Lyric)
	if len(lines) == 0 {
		return nil, fmt.Errorf("netease: 未获取到歌词")
	}
	return &provider.Lyrics{SongID: songID, Lines: lines}, nil
}

func (p *NeteaseCloudMusicApiProvider) get(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("netease api request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("netease api http status: %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 2<<20))
}

// parseLRC 去掉 [mm:ss.xx] 时间戳与元数据行，返回纯歌词行。
func parseLRC(lrc string) []provider.LyricLine {
	var lines []provider.LyricLine
	for _, raw := range strings.Split(lrc, "\n") {
		line := strings.TrimSpace(lrcTimeRe.ReplaceAllString(raw, ""))
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "[ti:") || strings.HasPrefix(lower, "[ar:") ||
			strings.HasPrefix(lower, "[al:") || strings.HasPrefix(lower, "[by:") ||
			strings.HasPrefix(lower, "[offset:") {
			continue
		}
		lines = append(lines, provider.LyricLine{Text: line})
	}
	return lines
}
