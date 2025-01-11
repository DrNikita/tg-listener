package cron

import (
	"context"
	"log/slog"
	"tg-listener/internal/db"
	"tg-listener/internal/telegram"
)

type CronJober interface{}

type CronRepository struct {
	chatWorker telegram.TgChatWorker
	store      db.StorageWorker
	// configs    configs.TgConfigs
	logger *slog.Logger
	ctx    context.Context
}

func NewCronRepository(chatWorker telegram.TgChatWorker, store db.StorageWorker, logger *slog.Logger, ctx context.Context) *CronRepository {
	return &CronRepository{
		chatWorker: chatWorker,
		store:      store,
		logger:     logger,
		ctx:        ctx,
	}
}

func (cr *CronRepository) Start(clientId int64) {
	avaliableChats, err := cr.store.GetListeningChats(clientId)
	if err != nil {
		cr.logger.Error("error getting listening chats", "err", err)
		return
	}

	// errPull := make(chan error, 10)

	for _, chat := range avaliableChats.ListeningChats {

		go func() {
			messages, err := cr.chatWorker.GetNewMessages(chat.Id)
			if err != nil {
				cr.logger.Error("error getting messages", "err", err)

			}

			if len(messages.Messages) == 0 {
				cr.logger.Info("no new messeges found")
				return
			}

			//TODO: kafka mb??
		}()
	}

	// for err := range errPull {
	// }

	cr.logger.Info("cron_monitoring started")
}
