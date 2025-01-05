package telegram

import (
	"log/slog"
	"tg-listener/configs"

	"github.com/zelenin/go-tdlib/client"
)

// type TgChannelWorker interface {
// 	Subscribe(channelName string) (*client.Response, error)
// 	Monitor(channelName string) error
// }

type ChannelRepository struct {
	Me                   *client.User
	Client               *client.Client
	ChannelSubscriptions map[string]string
	Config               *configs.TgConfig
	Logger               *slog.Logger
}

// func NewTelegramRepository(client *client.Client, config *configs.TgConfig, logger *slog.Logger) channelRepository {
// 	return channelRepository{
// 		client: client,
// 		config: config,
// 		logger: logger,
// 	}
// }

func (tr ChannelRepository) Subscribe(chatTag string) (*client.Ok, error) {
	chatId, err := tr.getChatId(chatTag)
	if err != nil {
		return nil, err
	}

	joinChatReq := client.JoinChatRequest{
		ChatId: chatId,
	}

	ok, err := tr.Client.JoinChat(&joinChatReq)
	if err != nil {
		return nil, err
	}

	tr.Logger.Info("successfully joined to chat", "chatID", chatId)

	chatMember, err := tr.Client.GetChatMember(&client.GetChatMemberRequest{
		ChatId: chatId,
		MemberId: &client.MessageSenderUser{
			UserId: tr.Me.Id,
		},
	})
	if err != nil {
		return nil, err
	}

	tr.Logger.Info("Chat member status: ", chatMember.Status)

	return ok, nil
}

func (tr ChannelRepository) getChatId(chatTag string) (int64, error) {
	searchChatReq := client.SearchPublicChatRequest{
		Username: chatTag,
	}

	chat, err := tr.Client.SearchPublicChat(&searchChatReq)
	if err != nil {
		return 0, err
	}

	return chat.Id, nil
}

// func (tr channelRepository) Monitor(channelUsername string) error {
// 	return nil
// }
