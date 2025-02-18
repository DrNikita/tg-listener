package domen

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

type DomenRepository struct {
	chatWorker telegram.TDLibAPIProvider
	store      db.StorageWorker
	logger     *slog.Logger
}

func New() (*DomenRepository, func(), func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))

	tgConfigs, err := configs.TgConfig()
	if err != nil {
		return nil, nil, nil, err
	}

	mongoConfigs, err := configs.MongoConfig()
	if err != nil {
		return nil, nil, nil, err
	}

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoConfigs.Uri))

	disconnect := func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error("mongodb disconnect error", "err", err)
			log.Fatal(err)
		}
	}

	if err != nil {
		logger.Error("error connecting to mongo DB", "err", err)
		return nil, disconnect, nil, err
	}

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Error(err.Error())
		return nil, disconnect, nil, err
	}

	mongoRepository := db.NewMongoRepository(mongoClient, mongoConfigs, logger)

	chatRepository, destroyClient, err := telegram.New(mongoRepository, tgConfigs, logger)
	if err != nil {
		return nil, disconnect, nil, err
	}

	cctx, ccancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer ccancel()

	if err := chatRepository.InitInitialSubscriptions(cctx); err != nil {
		logger.Error("failed to subscribe for initial cahts", "err", err)
		return nil, disconnect, destroyClient, err
	}

	return &DomenRepository{
		chatWorker: chatRepository,
		store:      mongoRepository,
		logger:     logger,
	}, disconnect, destroyClient, nil
}
