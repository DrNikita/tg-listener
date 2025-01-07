package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"tg-listener/configs"
	"tg-listener/internal/db"
	"tg-listener/internal/telegram"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))

	tgConfigs, mongoConfigs, err := configs.MustConfig()
	if err != nil {
		log.Fatal(err)
	}

	mongoClient, _ := mongo.Connect(options.Client().ApplyURI(mongoConfigs.Uri))
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			logger.Error(err.Error())
			log.Fatal(err)
		}
	}()

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Error(err.Error())
		log.Fatal(err)
	}

	var tgClientAuthorizer telegram.TgClientAuthorizer
	tgClientAuthorizer = telegram.NewClientRepository(tgConfigs, logger)
	tdlibClient, _, err := tgClientAuthorizer.Authorize()
	defer func() {
		meta, err := tdlibClient.Destroy()
		if err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("user was successfully destroed", "@type", meta.Type)
	}()

	var storageWorker db.StorageWorker
	var channelWorker telegram.TgChatWorker

	storageWorker = db.NewMongoRepository(mongoClient, mongoConfigs, logger, ctx)

	channelWorker = telegram.NewTelegramRepository(tdlibClient, &storageWorker, tgConfigs, logger)
	_, err = channelWorker.Subscribe(tgConfigs.TestChatTag)
	if err != nil {
		log.Fatal(err)
	}
}
