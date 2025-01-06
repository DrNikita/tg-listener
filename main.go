package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"tg-listener/configs"
	"tg-listener/internal/telegram"

	"github.com/zelenin/go-tdlib/client"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))

	tgConfigs, err := configs.MustConfig()
	if err != nil {
		log.Fatal(err)
	}

	tdlibParameters := &client.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseMessageDatabase:  true,
		UseSecretChats:      false,
		ApiId:               tgConfigs.ApiID,
		ApiHash:             tgConfigs.ApiHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
	}
	// client authorizer
	authorizer := client.ClientAuthorizer(tdlibParameters)
	go client.CliInteractor(authorizer)

	// or bot authorizer
	// botToken := "000000000:gsVCGG5YbikxYHC7bP5vRvmBqJ7Xz6vG6td"
	// authorizer := client.BotAuthorizer(tdlibParameters, botToken)

	_, err = client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 1,
	})
	if err != nil {
		log.Fatalf("SetLogVerbosityLevel error: %s", err)
	}

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}

	versionOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	commitOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "commit_hash",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	log.Printf("TDLib version: %s (commit: %s)", versionOption.(*client.OptionValueString).Value, commitOption.(*client.OptionValueString).Value)

	me, err := tdlibClient.GetMe()
	if err != nil {
		log.Fatalf("GetMe error: %s", err)
	}

	defer func() {
		meta, err := tdlibClient.Destroy()
		if err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("user was successfully destroed", "@type", meta.Type)
	}()

	log.Printf("Me: %s %s", me.FirstName, me.LastName)

	var channelWorker telegram.TgChatWorker

	chatList := make(map[int64]*client.Chat)

	channelWorker = telegram.NewTelegramRepository(tdlibClient, chatList, tgConfigs, logger)

	_, err = channelWorker.Subscribe(tgConfigs.TestChatTag)
	if err != nil {
		log.Fatal(err)
	}
}
