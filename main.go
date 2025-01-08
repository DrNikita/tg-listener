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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))

	tgConfigs, mongoConfigs, err := configs.MustConfig()
	if err != nil {
		log.Fatal(err)
	}

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoConfigs.Uri))
	if err != nil {
		logger.Error("error connecting to mongo DB", "err", err)
		log.Fatal(err)
	}

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
	tdlibClient, me, err := tgClientAuthorizer.Authorize()
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

	channelWorker = telegram.NewTelegramRepository(me, tdlibClient, &storageWorker, tgConfigs, logger)

	if err := channelWorker.InitialSubscribe(); err != nil {
		log.Fatal(err)
	}
}
