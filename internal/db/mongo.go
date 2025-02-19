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
	InsertInitialtListeningChats(ctx context.Context, listteningChats ListeningChats) error
	GetListeningChats(ctx context.Context, userId int64) (*ListeningChats, error)
	GetChatLastMessage(ctx context.Context, chatId int64) (*LastMessage, error)
	InsertLastMessage(ctx context.Context, lastMessage LastMessage) error
	UpdateLastMessage(ctx context.Context, lastMessage LastMessage) (*mongo.UpdateResult, error)
	InsertMessages(ctx context.Context, messages []Message) error
	GetMessages(ctx context.Context, from time.Time, to time.Time) ([]Message, error)
	//TODO:remove
	DropDB(ctx context.Context) error
}

type MongoRepository struct {
	client *mongo.Client
	config *configs.MongoConfigs
	logger *slog.Logger
}

func NewMongoRepository(client *mongo.Client, config *configs.MongoConfigs, logger *slog.Logger) *MongoRepository {
	return &MongoRepository{
		client: client,
		config: config,
		logger: logger,
	}
}

func (mr MongoRepository) InsertInitialtListeningChats(ctx context.Context, listteningChats ListeningChats) error {
	chatCollection := mr.client.Database(AppCollectionName).Collection(ChatColleaction)

	_, err := chatCollection.InsertOne(ctx, listteningChats)
	if err != nil {
		mr.logger.Error(err.Error())
		return err
	}

	return nil
}

func (mr MongoRepository) GetListeningChats(ctx context.Context, userId int64) (*ListeningChats, error) {
	chatCollection := mr.client.Database(AppCollectionName).Collection(ChatColleaction)

	var listeningChats ListeningChats

	userIdFilter := bson.D{{"user_id", userId}}
	err := chatCollection.FindOne(ctx, userIdFilter).Decode(&listeningChats)
	if err != nil {
		return nil, err
	}

	mr.logger.Info("listening chats", "amount", len(listeningChats.ListeningChats))

	return &listeningChats, nil
}

func (mr MongoRepository) GetChatLastMessage(ctx context.Context, chatId int64) (*LastMessage, error) {
	lastMessageCollection := mr.client.Database(AppCollectionName).Collection(LastMessageCollection)

	var lastMessage LastMessage

	chatIdFilter := bson.D{{"chat_id", chatId}}
	err := lastMessageCollection.FindOne(ctx, chatIdFilter).Decode(&lastMessage)
	if err != nil {
		mr.logger.Error("failed to get last message", "err", err)
		return nil, err
	}

	return &lastMessage, nil
}

func (mr MongoRepository) InsertLastMessage(ctx context.Context, lastMessage LastMessage) error {
	lastMessageCollection := mr.client.Database(AppCollectionName).Collection(LastMessageCollection)

	_, err := lastMessageCollection.InsertOne(ctx, lastMessage)
	if err != nil {
		mr.logger.Error("failed to save last message", "err", err)
		return err
	}

	mr.logger.Info("last message saved successfully", "last_message_id", lastMessage.LastMessageId)

	return nil
}

func (mr MongoRepository) UpdateLastMessage(ctx context.Context, lastMessage LastMessage) (*mongo.UpdateResult, error) {
	lastMessageCollection := mr.client.Database(AppCollectionName).Collection(LastMessageCollection)

	chatIdFilter := bson.D{{"chat_id", lastMessage.ChatId}}
	update := bson.D{{"$set", lastMessage}}

	updateResult, err := lastMessageCollection.UpdateOne(ctx, chatIdFilter, update)
	if err != nil {
		mr.logger.Error("failed to update last message", "err", err)
		return nil, err
	}

	mr.logger.Info("last message successfully updated", "msg_id", lastMessage.LastMessageId)

	return updateResult, nil
}

func (mr MongoRepository) InsertMessages(ctx context.Context, messages []Message) error {
	messageCollection := mr.client.Database(AppCollectionName).Collection(MessageCollection)

	_, err := messageCollection.InsertMany(ctx, messages)
	if err != nil {
		return err
	}

	return nil
}

func (mr MongoRepository) GetMessages(ctx context.Context, from time.Time, to time.Time) ([]Message, error) {
	return nil, nil
}

// TODO:remove
func (mr MongoRepository) DropDB(ctx context.Context) error {
	if err := mr.client.Database(AppCollectionName).Drop(ctx); err != nil {
		return err
	}

	return nil
}
