package main

import (
	"device-upgrade/internal/config"
	"device-upgrade/internal/gui"
	"log"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建并显示GUI窗口
	window := gui.NewUpgradeWindow(cfg)
	window.Show()
}
