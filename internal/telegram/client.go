package telegram

import (
	"log"
	"log/slog"
	"path/filepath"
	"tg-listener/configs"

	"github.com/zelenin/go-tdlib/client"
)

type TgClientAuthorizer interface {
	Authorize() (*client.Client, *client.User, error)
}

type clientRepository struct {
	configs *configs.TgConfigs
	logger  *slog.Logger
}

func NewClientRepository(configs *configs.TgConfigs, logger *slog.Logger) *clientRepository {
	return &clientRepository{
		configs: configs,
		logger:  logger,
	}
}

func (cr *clientRepository) Authorize() (*client.Client, *client.User, error) {
	tdlibParameters := &client.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseMessageDatabase:  true,
		UseSecretChats:      false,
		ApiId:               cr.configs.ApiID,
		ApiHash:             cr.configs.ApiHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
	}

	// client authorizer
	authorizer := client.ClientAuthorizer(tdlibParameters)
	phoneChan := make(chan string, 1)
	passChan := make(chan string, 1)
	go func() {
		authorizer.PhoneNumber = phoneChan
		authorizer.Password = passChan
		client.CliInteractor(authorizer)
	}()

	// or bot authorizer
	// botToken := "000000000:gsVCGG5YbikxYHC7bP5vRvmBqJ7Xz6vG6td"
	// authorizer := client.BotAuthorizer(tdlibParameters, botToken)

	_, err := client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 1,
	})
	if err != nil {
		cr.logger.Error(err.Error())
		return nil, nil, err
	}

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		cr.logger.Error(err.Error())
		return nil, nil, err
	}

	versionOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		cr.logger.Error(err.Error())
		return nil, nil, err
	}

	commitOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "commit_hash",
	})
	if err != nil {
		cr.logger.Error(err.Error())
		return nil, nil, err
	}

	log.Printf("TDLib version: %s (commit: %s)", versionOption.(*client.OptionValueString).Value, commitOption.(*client.OptionValueString).Value)

	me, err := tdlibClient.GetMe()
	if err != nil {
		cr.logger.Error(err.Error())
		return nil, nil, err
	}

	cr.logger.Info("user was successfully authorized", "Me", me.FirstName+" "+me.LastName)
	return tdlibClient, me, nil
}
