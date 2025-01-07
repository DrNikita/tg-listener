package db

import (
	"context"
	"log/slog"
	"tg-listener/configs"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type StorageWorker interface {
	InitListeningChats(listteningChats ListeningChats) error
}

type MongoRepository struct {
	client *mongo.Client
	config *configs.MongoConfigs
	logger *slog.Logger
	ctx    context.Context
}

func NewMongoRepository(client *mongo.Client, config *configs.MongoConfigs, logger *slog.Logger, ctx context.Context) *MongoRepository {
	return &MongoRepository{
		client: client,
		config: config,
		logger: logger,
		ctx:    ctx,
	}
}

func (mr *MongoRepository) InitListeningChats(listteningChats ListeningChats) error {
	chats := mr.client.Database("listening").Collection("chats")

	_, err := chats.InsertOne(mr.ctx, listteningChats)
	if err != nil {
		mr.logger.Error(err.Error())
		return err
	}

	return nil
}
