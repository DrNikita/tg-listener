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

	for _, chat := range avaliableChats.ListeningChats {
		messages, err := cr.chatWorker.GetNewMessages(chat.Id)
		if err != nil {
			cr.logger.Error("error getting messages", "err", err)
			continue
		}
		//TODO: kafka mb??
	}

	cr.logger.Info("Cron started")
}
