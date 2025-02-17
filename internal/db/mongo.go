package db

import (
	"context"
	"log/slog"
	"tg-listener/configs"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	AppCollectionName     string = "tg_listener"
	ChatColleaction       string = "chat"
	LastMessageCollection string = "last_message"
	MessageCollection     string = "message"
)

type StorageWorker interface {
	InsertInitialtListeningChats(listteningChats ListeningChats) error
	GetListeningChats(userId int64) (*ListeningChats, error)
	GetChatLastMessage(chatId int64) (*LastMessage, error)
	InsertLastMessage(lastMessage LastMessage) error
	UpdateLastMessage(lastMessage LastMessage) (*mongo.UpdateResult, error)
	InsertMessages(messages []Message) error
	GetMessages(from time.Time, to time.Time) ([]Message, error)
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

func (mr MongoRepository) InsertInitialtListeningChats(listteningChats ListeningChats) error {
	chatCollection := mr.client.Database(AppCollectionName).Collection(ChatColleaction)

	_, err := chatCollection.InsertOne(mr.ctx, listteningChats)
	if err != nil {
		mr.logger.Error(err.Error())
		return err
	}

	return nil
}

func (mr MongoRepository) GetListeningChats(userId int64) (*ListeningChats, error) {
	chatCollection := mr.client.Database(AppCollectionName).Collection(ChatColleaction)

	var listeningChats ListeningChats

	userIdFilter := bson.D{{"user_id", userId}}
	err := chatCollection.FindOne(mr.ctx, userIdFilter).Decode(&listeningChats)
	if err != nil {
		return nil, err
	}

	mr.logger.Info("listening chats", "amount", len(listeningChats.ListeningChats))

	return &listeningChats, nil
}

func (mr MongoRepository) GetChatLastMessage(chatId int64) (*LastMessage, error) {
	lastMessageCollection := mr.client.Database(AppCollectionName).Collection(LastMessageCollection)

	var lastMessage LastMessage

	chatIdFilter := bson.D{{"chat_id", chatId}}
	err := lastMessageCollection.FindOne(mr.ctx, chatIdFilter).Decode(&lastMessage)
	if err != nil {
		mr.logger.Error("failed to get last message", "err", err)
		return nil, err
	}

	return &lastMessage, nil
}

func (mr MongoRepository) InsertLastMessage(lastMessage LastMessage) error {
	lastMessageCollection := mr.client.Database(AppCollectionName).Collection(LastMessageCollection)

	_, err := lastMessageCollection.InsertOne(mr.ctx, lastMessage)
	if err != nil {
		mr.logger.Error("failed to save last message", "err", err)
		return err
	}

	mr.logger.Info("last message saved successfully", "last_message_id", lastMessage.LastMessageId)

	return nil
}

func (mr MongoRepository) UpdateLastMessage(lastMessage LastMessage) (*mongo.UpdateResult, error) {
	lastMessageCollection := mr.client.Database(AppCollectionName).Collection(LastMessageCollection)

	chatIdFilter := bson.D{{"chat_id", lastMessage.ChatId}}
	update := bson.D{{"$set", lastMessage}}

	updateResult, err := lastMessageCollection.UpdateOne(mr.ctx, chatIdFilter, update)
	if err != nil {
		mr.logger.Error("failed to update last message", "err", err)
		return nil, err
	}

	mr.logger.Info("last message successfully updated", "msg_id", lastMessage.LastMessageId)

	return updateResult, nil
}

func (mr MongoRepository) InsertMessages(messages []Message) error {
	messageCollection := mr.client.Database(AppCollectionName).Collection(MessageCollection)

	_, err := messageCollection.InsertMany(mr.ctx, messages)
	if err != nil {
		return err
	}

	return nil
}

func (mr MongoRepository) GetMessages(from time.Time, to time.Time) ([]Message, error) {
	return nil, nil
}
