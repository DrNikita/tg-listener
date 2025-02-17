package cron

import (
	"context"
	"log/slog"
	"sync"
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
	if len(avaliableChats.ListeningChats) == 0 {
		cr.logger.Error("error getting listening chats")
		return
	}

	// errPull := make(chan error, 10)
	wg := sync.WaitGroup{}

	for _, chat := range avaliableChats.ListeningChats {

		wg.Add(1)
		go func() {
			defer wg.Done()
			cr.logger.Info("chat_id", chat.Id, "chat_tag", chat.Tag)
			messages, err := cr.chatWorker.GetNewMessages(chat.Tag)
			if err != nil {
				cr.logger.Error("error getting messages", "err", err)
				return
			}

			if len(messages.Messages) == 0 {
				cr.logger.Info("no new messeges found")
				return
			}

			mongoMessages := db.NewMessages(messages)
			if err := cr.store.InsertMessages(mongoMessages); err != nil {
				cr.logger.Error("failed to save messages", "err", err)
			}
		}()
	}

	// for err := range errPull {
	// }

	cr.logger.Info("cron_monitoring started")
	wg.Wait()
}
