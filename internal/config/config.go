package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MQTT MQTTConfig `yaml:"mqtt"`
}

type MQTTConfig struct {
	Broker   string   `yaml:"broker"`
	ClientID string   `yaml:"clientId"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Topics   Topics   `yaml:"topics"`
}

type Topics struct {
	Upgrade string `yaml:"upgrade"`
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	// 首先检查当前目录
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}

	// 检查可执行文件所在目录
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		configPath := filepath.Join(execDir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// 对于 macOS .app 包，检查 Resources 目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		if filepath.Base(filepath.Dir(execDir)) == "Contents" {
			// 在 .app 包内，检查 Resources 目录
			configPath := filepath.Join(filepath.Dir(execDir), "Resources", "config.yaml")
			if _, err := os.Stat(configPath); err == nil {
				return configPath
			}
		}
	}

	// 返回默认路径
	return "config/config.yaml"
}

func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(buf, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}