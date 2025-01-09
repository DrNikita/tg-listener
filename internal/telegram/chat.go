package telegram

import (
	"log/slog"
	"tg-listener/configs"
	"tg-listener/internal/db"

	"github.com/zelenin/go-tdlib/client"
)

type TgChatWorker interface {
	Subscribe(chatTag string) (*client.Chat, error)
	InitialSubscribe() error
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

func (tr *chatRepository) InitialSubscribe() error {
	listeningChats, err := tr.store.GetListeningChats(tr.me.Id)
	if err != nil {
		tr.logger.Error("error getting listening chats from db", "err", err)
	}

	if listeningChats != nil && len(listeningChats.ListeningChats) != 0 {
		for _, listeningChat := range listeningChats.ListeningChats {
			tr.chatList[listeningChat.Tag] = listeningChat.Id
		}

		return nil
	}

	listeningChatTags := []string{
		"@evelone192gg",
		"@FlattyBy",
		"@хуйня_рандомная)",
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
