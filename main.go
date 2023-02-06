package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"gopkg.in/yaml.v3"
	"os"
)

type GlobalConfig struct {
	ApiKey    string `yaml:"apiKey"`
	SecretKey string `yaml:"secretKey"`
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

func main() {
	Setup()

	client := binance.NewClient(Config().ApiKey, Config().SecretKey)
	res, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", res.Balances)

}
