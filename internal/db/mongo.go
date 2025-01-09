package db

import (
	"context"
	"log/slog"
	"tg-listener/configs"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type StorageWorker interface {
	InsertInitialtListeningChats(listteningChats ListeningChats) error
	GetListeningChats(userId int64) (*ListeningChats, error)
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

func (mr *MongoRepository) InsertInitialtListeningChats(listteningChats ListeningChats) error {
	chats := mr.client.Database("listening").Collection("chats")

	_, err := chats.InsertOne(mr.ctx, listteningChats)
	if err != nil {
		mr.logger.Error(err.Error())
		return err
	}

	return nil
}

func (mr *MongoRepository) GetListeningChats(userId int64) (*ListeningChats, error) {
	chats := mr.client.Database("listening").Collection("chats")

	var listeningChats ListeningChats

	filter := bson.D{{"user_id", userId}}
	err := chats.FindOne(mr.ctx, filter).Decode(&listeningChats)
	if err != nil {
		return nil, err
	}
	// if errors.Is(err, mongo.ErrNoDocuments) {
	// 	// Do something when no record was found
	// } else if err != nil {
	// 	log.Fatal(err)
	// }

	mr.logger.Info("listening chats", "chats____________", len(listeningChats.ListeningChats))

	return &listeningChats, nil
}
