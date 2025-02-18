package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"tg-listener/configs"
	"tg-listener/internal/db"

	"github.com/zelenin/go-tdlib/client"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TDLibAPIProvider interface {
	InitInitialSubscriptions(ctx context.Context) error
	Subscribe(ctx context.Context, chatTag string) (*client.Chat, error)
	GetNewMessages(ctx context.Context, chatTag string) (*client.Messages, error)
	GetAuthorizedUserID() int64
}

type NoMessagesError struct {
	ChatId int64
}

func (e NoMessagesError) Error() string {
	return fmt.Sprintf("no new messages found for chat: chat_id: %d", e.ChatId)
}

type chatRepository struct {
	me       *client.User
	client   *client.Client
	chatList map[string]int64
	store    db.StorageWorker
	config   *configs.TgConfigs
	logger   *slog.Logger
}

func New(store db.StorageWorker, config *configs.TgConfigs, logger *slog.Logger) (*chatRepository, func(), error) {
	tdlibParameters := &client.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseMessageDatabase:  true,
		UseSecretChats:      false,
		ApiId:               config.ApiID,
		ApiHash:             config.ApiHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
	}

	authorizer := client.ClientAuthorizer(tdlibParameters)
	go client.CliInteractor(authorizer)

	_, err := client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 1,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, nil, err
	}

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil, err
	}

	versionOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, nil, err
	}

	commitOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "commit_hash",
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, nil, err
	}

	log.Printf("TDLib version: %s (commit: %s)", versionOption.(*client.OptionValueString).Value, commitOption.(*client.OptionValueString).Value)

	me, err := tdlibClient.GetMe()

	destroyClient := func() {
		meta, err := tdlibClient.Destroy()
		if err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("user was successfully destroed", "@type", meta.Type)
	}

	if err != nil {
		logger.Error(err.Error())
		return nil, destroyClient, err
	}

	logger.Info("user was successfully authorized", "Me", me.FirstName+" "+me.LastName)

	return &chatRepository{
		me:       me,
		client:   tdlibClient,
		chatList: make(map[string]int64),
		store:    store,
		config:   config,
		logger:   logger,
	}, destroyClient, nil
}

// initialization of initial subscriptions
func (cr *chatRepository) InitInitialSubscriptions(ctx context.Context) error {
	listeningChats, err := cr.store.GetListeningChats(ctx, cr.me.Id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// list of initial chats for listening
		listeningChatTags := []string{
			"@evelone192gg",
			"@FlattyBy",
		}

		for _, chatTag := range listeningChatTags {
			_, err := cr.Subscribe(ctx, chatTag)
			if err != nil {
				cr.logger.Error("coulnt subscrite to chat", "chat_tag", chatTag, "err", err)
			}
		}

		if err := cr.initListeningChats(ctx); err != nil {
			return err
		}

	} else if err != nil {
		cr.logger.Error("error getting listening chats from db", "err", err)
	}

	if listeningChats != nil && len(listeningChats.ListeningChats) != 0 {
		for _, listeningChat := range listeningChats.ListeningChats {
			cr.chatList[listeningChat.Tag] = listeningChat.Id
		}

		return nil
	}

	return nil
}

// TODO: refactor: remove *client.Chat as returnning param: seems like chat param is only need to check chatTag
func (cr *chatRepository) Subscribe(ctx context.Context, chatTag string) (*client.Chat, error) {
	chatId, err := cr.getChatId(chatTag)
	if err != nil {
		cr.logger.Error("chat id not found", "err", err)
		return nil, err
	}

	chat, err := cr.client.GetChat(&client.GetChatRequest{
		ChatId: chatId,
	})
	if err != nil {
		cr.logger.Error("get chat error", "err", err)
		return nil, err
	}

	cr.chatList[chatTag] = chatId
	cr.logger.Info("chat subscribed", "chat", chatTag)

	return chat, nil
}

// get new message since last messages request
func (cr *chatRepository) GetNewMessages(ctx context.Context, chatTag string) (*client.Messages, error) {
	chat, err := cr.client.SearchPublicChat(&client.SearchPublicChatRequest{
		Username: chatTag,
	})
	if err != nil {
		cr.logger.Error("failed to get chat", "err", err)
		return nil, err
	}

	messages, err := cr.client.GetChatHistory(&client.GetChatHistoryRequest{
		ChatId:    chat.Id,
		Limit:     10,
		OnlyLocal: false,
	})
	if err != nil {
		cr.logger.Error("get chat messages history error", "chat_id", chatTag, "err", err)
		return nil, err
	}

	if messages == nil || messages.TotalCount == 0 {
		cr.logger.Info("no messages were found")

		return nil, NoMessagesError{
			ChatId: chat.Id,
		}
	}

	cr.logger.Info("messages were successfully got", "total messages count", messages.TotalCount)

	// important to save this to get rid of duplicated requests with the same messages ->
	// -> return error if failed to inserte/update
	lastMessage, err := cr.store.GetChatLastMessage(ctx, chat.Id)
	if err != nil {
		err = cr.store.InsertLastMessage(ctx, db.LastMessage{
			ChatId:        chat.Id,
			LastMessageId: messages.Messages[0].Id,
		})
		if err != nil {
			cr.logger.Error("failed to save last message to mongo db", "err", err)
			return nil, err
		}

		return messages, nil

	} else {
		_, err = cr.store.UpdateLastMessage(ctx, db.LastMessage{
			ChatId:        chat.Id,
			LastMessageId: messages.Messages[0].Id,
		})
		if err != nil {
			cr.logger.Error("failed to update last message in mongo db", "err", err)
			return nil, err
		}
	}

	if lastMessage != nil {
		for messageId, message := range messages.Messages {
			if message.Id == lastMessage.LastMessageId {
				messages.Messages = messages.Messages[:messageId]
				messages.TotalCount = int32(len(messages.Messages))
				cr.logger.Info("new messages count", "chat_tag", chatTag, "count", len(messages.Messages))
				return messages, nil
			}
		}
	}

	cr.logger.Info("new messages count", "chat_tag", chatTag, "count", len(messages.Messages))

	return messages, nil
}

// service functions for saving listening chats
func (cr *chatRepository) initListeningChats(ctx context.Context) error {
	var listeningChats []db.TgListeningChat

	for chatTag, chatId := range cr.chatList {
		listeningChats = append(listeningChats, db.TgListeningChat{
			Tag: chatTag,
			Id:  chatId,
		})
	}

	initialChats := db.ListeningChats{
		UserId:         cr.me.Id,
		ListeningChats: listeningChats,
	}

	if err := cr.store.InsertInitialtListeningChats(ctx, initialChats); err != nil {
		cr.logger.Error(err.Error())
		return err
	}

	return nil
}

func (cr *chatRepository) getChatId(chatTag string) (int64, error) {
	chat, err := cr.client.SearchPublicChat(&client.SearchPublicChatRequest{
		Username: chatTag,
	})
	if err != nil {
		cr.logger.Error("Chat not found", "chat_tag", chatTag, "err", err)
		return 0, err
	}

	cr.logger.Info("Chat found", "chatId", chat.Id)

	return chat.Id, nil
}

func (cr *chatRepository) GetAuthorizedUserID() int64 {
	return cr.me.Id
}
