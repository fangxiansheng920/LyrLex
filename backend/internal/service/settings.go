package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/wire"

	"lyrics-server/internal/config"
)

// APIConfig 翻译 API 密钥（留空则使用默认免费源）。
type APIConfig struct {
	YoudaoAppKey    string `json:"youdao_app_key"`
	YoudaoAppSecret string `json:"youdao_app_secret"`
	BaiduAppID      string `json:"baidu_app_id"`
	BaiduSecret     string `json:"baidu_secret"`
	DeepLKey        string `json:"deepl_key"`
}

// Settings 应用设置。
type Settings struct {
	APIKeys             APIConfig `json:"api_keys"`
	PronunciationSource string    `json:"pronunciation_source"` // 发音源，目前支持 youdao
}

// SettingsService 设置读写（持久化到数据目录 settings.json）。
type SettingsService interface {
	Get(ctx context.Context) (*Settings, error)
	Save(ctx context.Context, s *Settings) error
	DataDir() string
}

type settingsService struct {
	cfg *config.Config
}

var _ SettingsService = (*settingsService)(nil)

func NewSettingsService(cfg *config.Config) *settingsService {
	return &settingsService{cfg: cfg}
}

func (s *settingsService) path() string {
	return filepath.Join(s.cfg.DataDir(), "settings.json")
}

func (s *settingsService) Get(ctx context.Context) (*Settings, error) {
	settings := &Settings{PronunciationSource: "youdao"}
	data, err := os.ReadFile(s.path())
	if err != nil {
		if os.IsNotExist(err) {
			return settings, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}
	if settings.PronunciationSource == "" {
		settings.PronunciationSource = "youdao"
	}
	return settings, nil
}

func (s *settingsService) Save(ctx context.Context, settings *Settings) error {
	if err := os.MkdirAll(filepath.Dir(s.path()), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), data, 0o644)
}

func (s *settingsService) DataDir() string {
	return s.cfg.DataDir()
}

var SettingsProviderSet = wire.NewSet(
	NewSettingsService,
	wire.Bind(new(SettingsService), new(*settingsService)),
)
