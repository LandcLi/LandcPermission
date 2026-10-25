package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	landcpermission "github.com/LandcLi/LandcPermission"
)

func main() {
	// 支持自定义配置路径：LANDC_CONFIG=./prod.yaml
	configPath := os.Getenv("LANDC_CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
	}

	server, err := landcpermission.InitDefault(configPath)
	if err != nil {
		log.Fatalf("LandcPermission 初始化失败: %v", err)
	}

	// 优雅停机：收到 SIGINT/SIGTERM 后等待在途请求完成
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.RunWithContext(ctx); err != nil {
		log.Fatalf("LandcPermission 运行失败: %v", err)
	}
}
