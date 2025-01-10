package telegram

import (
	"errors"
	"log/slog"
	"tg-listener/configs"
	"tg-listener/internal/db"

	"github.com/zelenin/go-tdlib/client"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TgChatWorker interface {
	InitInitialSubscriptions() error
	Subscribe(chatTag string) (*client.Chat, error)
}

type chatRepository struct {
	me       *client.User
	client   *client.Client
	chatList map[string]int64
	store    db.StorageWorker
	config   *configs.TgConfigs
	logger   *slog.Logger
}

func NewTelegramRepository(me *client.User, client *client.Client, store db.StorageWorker, config *configs.TgConfigs, logger *slog.Logger) *chatRepository {
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
func (tr *chatRepository) InitInitialSubscriptions() error {
	listeningChats, err := tr.store.GetListeningChats(tr.me.Id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// list of initial chats for listening
		listeningChatTags := []string{
			"@evelone192gg",
			"@FlattyBy",
			"@-----incorrect_chat_tag-----)",
		}

		for _, chatTag := range listeningChatTags {
			_, err := tr.Subscribe(chatTag)
			if err != nil {
				tr.logger.Error("coulnt subscrite to chat", "chat_tag", chatTag, "err", err)
			}
		}

		if err := tr.initListeningChats(); err != nil {
			return err
		}

	} else if err != nil {
		tr.logger.Error("error getting listening chats from db", "err", err)
	}

	if listeningChats != nil && len(listeningChats.ListeningChats) != 0 {
		for _, listeningChat := range listeningChats.ListeningChats {
			tr.chatList[listeningChat.Tag] = listeningChat.Id
		}

		return nil
	}

	return nil
}

func (tr *chatRepository) Subscribe(chatTag string) (*client.Chat, error) {
	chatId, err := tr.getChatId(chatTag)
	if err != nil {
		tr.logger.Error("Chat id not found", "err", err)
		return nil, err
	}

	chat, err := tr.client.GetChat(&client.GetChatRequest{
		ChatId: chatId,
	})
	if err != nil {
		tr.logger.Error("Get chat error", "err", err)
		return nil, err
	}

	tr.chatList[chatTag] = chatId
	tr.logger.Info("Chat subscribed", "chat", chatTag)

	return chat, nil
}

// service functions for saving listening chats
func (tr *chatRepository) initListeningChats() error {
	var listeningChats []db.TgListeningChat

	for chatTag, chatId := range tr.chatList {
		listeningChats = append(listeningChats, db.TgListeningChat{
			Tag: chatTag,
			Id:  chatId,
		})
	}

	initialChats := db.ListeningChats{
		UserId:         tr.me.Id,
		ListeningChats: listeningChats,
	}

	if err := tr.store.InsertInitialtListeningChats(initialChats); err != nil {
		tr.logger.Error(err.Error())
		return err
	}

	return nil
}

func (tr *chatRepository) getChatId(chatTag string) (int64, error) {
	chat, err := tr.client.SearchPublicChat(&client.SearchPublicChatRequest{
		Username: chatTag,
	})
	if err != nil {
		tr.logger.Error("Chat not found", "chat_tag", chatTag, "err", err)
		return 0, err
	}

	tr.logger.Info("Chat found", "chatId", chat.Id)

	return chat.Id, nil
}
