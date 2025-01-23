package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"tg-listener/configs"
	"tg-listener/internal/cron"
	"tg-listener/internal/db"
	"tg-listener/internal/http"
	"tg-listener/internal/telegram"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/zelenin/go-tdlib/client"
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

	httpConfigs, tgConfigs, mongoConfigs, err := configs.MustConfig()
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

	tdlibClientChan := make(chan *client.Client)
	meChan := make(chan *client.User)
	defer func() {
		close(tdlibClientChan)
		close(meChan)
	}()

	///// must be block for all api if user unauthorized

	storageWorker := db.NewMongoRepository(mongoClient, mongoConfigs, logger, ctx)

	channelWorker := telegram.NewTelegramRepository(me, tdlibClient, storageWorker, tgConfigs, logger)

	if err := channelWorker.InitInitialSubscriptions(); err != nil {
		log.Fatal(err)
	}

	cronRepository := cron.NewCronRepository(channelWorker, storageWorker, logger, ctx)
	cronRepository.Start(me.Id)

	app := fiber.New()

	httpRepository := http.NewHttpRepository(tgClientAuthorizer, channelWorker, nil)
	httpRepository.SetupRouts(app)

	// Uuups...
	// Global variables??))))))

	log.Fatal(app.Listen(fmt.Sprintf("%s:%s", httpConfigs.AppHost, httpConfigs.AppPort)))
}
