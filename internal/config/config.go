package config

import (
	"os"

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

func LoadConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
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