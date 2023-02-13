package setting

import (
	"gopkg.in/yaml.v3"
	"os"
)

type GlobalConfig struct {
	ApiKey    string `yaml:"apiKey"`
	SecretKey string `yaml:"secretKey"`
	BotKey    string `yaml:"botKey"`
}

var config *GlobalConfig

func Config() *GlobalConfig {
	return config
}

func Setup() {
	configBytes, err := os.ReadFile("./config.yaml")
	if err != nil {
		panic(err)
	}
	config = &GlobalConfig{}
	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		panic(err)
	}
}
