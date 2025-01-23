package configs

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
)

type HttpConfigs struct {
	AppHost string `envconfig:"host"`
	AppPort string `envconfig:"port"`
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

func MustConfig() (*HttpConfigs, *TgConfigs, *MongoConfigs, error) {
	var httpConfigs HttpConfigs
	var tgConfigs TgConfigs
	var mongoConfigs MongoConfigs

	err := envconfig.Process("app", &httpConfigs)
	if err != nil {
		return nil, nil, nil, err
	}

	err = envconfig.Process("tg", &tgConfigs)
	if err != nil {
		return nil, nil, nil, err
	}

	err = envconfig.Process("db", &mongoConfigs)
	if err != nil {
		return nil, nil, nil, err
	}

	return &httpConfigs, &tgConfigs, &mongoConfigs, nil
}
