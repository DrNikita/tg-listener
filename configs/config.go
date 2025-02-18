package configs

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
)

type AppConfigs struct {
	MediaDefaultDirectory string `envconfig:"MEDIA_DEFAULT_DIRECTORY"`
}

type TgConfigs struct {
	ApiID       int32  `envconfig:"api_id"`
	ApiHash     string `envconfig:"api_hash"`
	PhoneNumber string `envconfig:"phone_number"`
	AuthCode    string `envconfig:"auth_code"`
	TestChatTag string `envconfig:"chat_tag"`
}

type MongoConfigs struct {
	Uri string `envconfig:"MONGO_URI"`
}

func AppConfig() (*AppConfigs, error) {
	var appConfigs AppConfigs

	err := envconfig.Process("", &appConfigs)
	if err != nil {
		return nil, err
	}

	return &appConfigs, nil
}

func TgConfig() (*TgConfigs, error) {
	var tgConfigs TgConfigs

	err := envconfig.Process("tg", &tgConfigs)
	if err != nil {
		return nil, err
	}

	return &tgConfigs, nil
}

func MongoConfig() (*MongoConfigs, error) {
	var mongoConfigs MongoConfigs

	err := envconfig.Process("db", &mongoConfigs)
	if err != nil {
		return nil, err
	}

	return &mongoConfigs, nil
}
