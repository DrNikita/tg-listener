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
	chatList map[string]int64
	config   *configs.TgConfig
	logger   *slog.Logger
}

func NewTelegramRepository(client *client.Client, config *configs.TgConfig, logger *slog.Logger) chatRepository {
	return chatRepository{
		client:   client,
		chatList: make(map[string]int64),
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

	tr.chatList[chatTag] = chatId
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
