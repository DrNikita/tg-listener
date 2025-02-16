package telegram

import (
	"errors"
	"fmt"
	"log/slog"
	"tg-listener/configs"
	"tg-listener/internal/db"

	"github.com/zelenin/go-tdlib/client"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TgChatWorker interface {
	InitInitialSubscriptions() error
	Subscribe(chatTag string) (*client.Chat, error)
	GetNewMessages(chatTag string) (*client.Messages, error)
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

func NewChatRepository(me *client.User, client *client.Client, store db.StorageWorker, config *configs.TgConfigs, logger *slog.Logger) *chatRepository {
	return &chatRepository{
		me:       me,
		client:   client,
		chatList: make(map[string]int64),
		store:    store,
		config:   config,
		logger:   logger,
	}
}

// initialization of initial subscriptions
func (cr *chatRepository) InitInitialSubscriptions() error {
	listeningChats, err := cr.store.GetListeningChats(cr.me.Id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// list of initial chats for listening
		listeningChatTags := []string{
			"@evelone192gg",
			"@FlattyBy",
		}

		for _, chatTag := range listeningChatTags {
			_, err := cr.Subscribe(chatTag)
			if err != nil {
				cr.logger.Error("coulnt subscrite to chat", "chat_tag", chatTag, "err", err)
			}
		}

		if err := cr.initListeningChats(); err != nil {
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
func (cr *chatRepository) Subscribe(chatTag string) (*client.Chat, error) {
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
func (cr *chatRepository) GetNewMessages(chatTag string) (*client.Messages, error) {
	chat, err := cr.client.SearchPublicChat(&client.SearchPublicChatRequest{
		Username: chatTag,
	})
	if err != nil {
		cr.logger.Error("failed to get chat", "err", err)
		return nil, err
	}

	messages, err := cr.client.GetChatHistory(&client.GetChatHistoryRequest{
		ChatId:    chat.Id,
		Limit:     100,
		OnlyLocal: false,
	})
	if err != nil {
		cr.logger.Error("get chat messages history error", "chat_id", chatTag, "err", err)
		return nil, err
	}

	if messages == nil || len(messages.Messages) == 0 {
		cr.logger.Info("no messages were found")

		return nil, NoMessagesError{
			ChatId: chat.Id,
		}
	}

	cr.logger.Info("messages were successfully got", "total messages count", messages.TotalCount)

	// important to save this to get rid of duplicated requests with the same messages ->
	// -> return error if failed to inserte/update
	lastMessage, err := cr.store.GetChatLastMessage(chat.Id)
	if err != nil {
		err = cr.store.InsertLastMessage(db.LastMessage{
			ChatId:        chat.Id,
			LastMessageId: messages.Messages[0].Id,
		})
		if err != nil {
			cr.logger.Error("failed to save last message to mongo db", "err", err)
			return nil, err
		}

		return messages, nil

	} else {
		_, err = cr.store.UpdateLastMessage(db.LastMessage{
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
				cr.logger.Info("new messages count", "chat_tag", chatTag, "count", len(messages.Messages))
				return messages, nil
			}
		}
	}

	cr.logger.Info("new messages count", "chat_tag", chatTag, "count", len(messages.Messages))

	return messages, nil
}

// service functions for saving listening chats
func (cr *chatRepository) initListeningChats() error {
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

	if err := cr.store.InsertInitialtListeningChats(initialChats); err != nil {
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
