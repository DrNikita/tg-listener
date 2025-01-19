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
	GetChatLastMessage(chatId int64) (*LastMessage, error)
	InsertLastMessage(lastMessage LastMessage) error
	UpdateLastMessage(lastMessage LastMessage) (*mongo.UpdateResult, error)
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

	userIdFilter := bson.D{{"user_id", userId}}
	err := chats.FindOne(mr.ctx, userIdFilter).Decode(&listeningChats)
	if err != nil {
		return nil, err
	}

	mr.logger.Info("listening chats", "chats____________", len(listeningChats.ListeningChats))

	return &listeningChats, nil
}

func (mr *MongoRepository) GetChatLastMessage(chatId int64) (*LastMessage, error) {
	message := mr.client.Database("listening").Collection("last_message")

	var lastMessage LastMessage

	chatIdFilter := bson.D{{"chat_id", chatId}}
	err := message.FindOne(mr.ctx, chatIdFilter).Decode(&lastMessage)
	if err.Error() == mongo.ErrNoDocuments.Error() {
		mr.logger.Info("error", "not documents found", "err", err)
		return nil, nil
	}
	if err != nil {
		mr.logger.Error("failed to get last message", "err", err)
		return nil, err
	}

	return &lastMessage, nil
}

func (mr *MongoRepository) InsertLastMessage(lastMessage LastMessage) error {
	chats := mr.client.Database("listening").Collection("last_message")

	_, err := chats.InsertOne(mr.ctx, lastMessage)
	if err != nil {
		mr.logger.Error(err.Error())
		return err
	}

	return nil
}

func (mr *MongoRepository) UpdateLastMessage(lastMessage LastMessage) (*mongo.UpdateResult, error) {
	message := mr.client.Database("listening").Collection("last_message")

	chatIdFilter := bson.D{{"chat_id", lastMessage.ChatId}}

	update := bson.D{
		{"$set", bson.M{
			"chat_id":         lastMessage.ChatId,
			"last_message_id": lastMessage.LastMessageId,
			"offset":          lastMessage.Offset,
		},
		},
	}

	updateResult, err := message.UpdateOne(mr.ctx, chatIdFilter, update)
	if err != nil {
		mr.logger.Error(err.Error())
		return nil, err
	}

	return updateResult, nil
}
