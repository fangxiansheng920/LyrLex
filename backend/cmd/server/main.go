package main

import (
	"flag"
	"fmt"
	"log"

	"lyrics-server/internal"
	"lyrics-server/internal/config"
)

func main() {
	cfgPath := flag.String("config", "", "配置文件路径（默认 configs/config.yaml）")
	port := flag.Int("port", 0, "监听端口（0 = 使用配置）")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if *port != 0 {
		cfg.Server.Port = *port
	}

	app, err := internal.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("lyrics-server listening on http://%s", addr)
	if err := app.Router.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
