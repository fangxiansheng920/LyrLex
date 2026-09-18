package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 汇总全部配置；非敏感项来自 configs/config.yaml，敏感项来自 .env。
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DBConfig struct {
	Path string `mapstructure:"path"`
}

// DataDir 返回本地数据目录：优先 LYRICS_DATA_DIR（Electron 打包运行时指向 userData），
// 否则取数据库文件所在目录。
func (c *Config) DataDir() string {
	if dir := os.Getenv("LYRICS_DATA_DIR"); dir != "" {
		return dir
	}
	return filepath.Dir(c.DB.Path)
}

// Load 读取配置文件与 .env。cfgPath 为空时使用默认路径 configs/config.yaml；
// 配置文件缺失时回退到默认值（打包环境下可能没有 configs 目录）。
func Load(cfgPath string) (*Config, error) {
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}

	v := viper.New()
	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")

	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", 18080)
	v.SetDefault("db.path", "../data/lyrics.db")

	// 配置文件缺失不致命，使用默认值。
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("[config] 未读取到配置文件 %s，使用默认值: %v\n", cfgPath, err)
	}

	// 敏感数据放 .env（本地环境变量），缺失同样不致命。
	_ = godotenv.Load(".env")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
