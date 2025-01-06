package telegram

import (
	"log/slog"
	"tg-listener/configs"

	"github.com/zelenin/go-tdlib/client"
)

type TgChatWorker interface {
	Subscribe(chatTag string) (*client.Chat, error)
}

type chatRepository struct {
	client   *client.Client
	chatList map[int64]*client.Chat
	config   *configs.TgConfig
	logger   *slog.Logger
}

func NewTelegramRepository(client *client.Client, chatList map[int64]*client.Chat, config *configs.TgConfig, logger *slog.Logger) chatRepository {
	return chatRepository{
		client:   client,
		chatList: chatList,
		config:   config,
		logger:   logger,
	}
}

func (tr chatRepository) Subscribe(chatTag string) (*client.Chat, error) {
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

	tr.chatList[chatId] = chat
	tr.logger.Info("Chat subscribed", "chat", chatTag)

	return chat, nil
}

func (tr chatRepository) getChatId(chatTag string) (int64, error) {
	chat, err := tr.client.SearchPublicChat(&client.SearchPublicChatRequest{
		Username: chatTag,
	})
	if err != nil {
		tr.logger.Error("Chat not found", "err", err)
		return 0, err
	}

	tr.logger.Info("Chat found", "chatId", chat.Id)

	return chat.Id, nil
}

// TODO: remove if not used
func (tr chatRepository) joinChat(chatId int64) (*client.Ok, error) {
	ok, err := tr.client.JoinChat(&client.JoinChatRequest{
		ChatId: chatId,
	})

	if err != nil {
		tr.logger.Error("Join chat error", "err", err)
		return nil, err
	}

	tr.logger.Info("Chat joined", "chatId", chatId)

	return ok, nil
}

// func (tr channelRepository) Monitor(channelUsername string) error {
// 	return nil
// }
