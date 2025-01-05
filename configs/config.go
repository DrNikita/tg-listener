package configs

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
)

type TgConfig struct {
	ApiID       int32  `envconfig:"api_id"`
	ApiHash     string `envconfig:"api_hash"`
	PhoneNumber string `envconfig:"phone_number"`
	AuthCode    string `envconfig:"auth_code"`
}

func MustConfig() (*TgConfig, error) {
	var config TgConfig

	err := envconfig.Process("tg", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
