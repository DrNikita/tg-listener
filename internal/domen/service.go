package domen

import (
	"context"
	"sync"
	"tg-listener/internal/db"
	"time"
)

func (dr *DomenRepository) BackgroundListening() {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	avaliableChats, err := dr.store.GetListeningChats(ctx, dr.chatWorker.GetAuthorizedUserID())
	if err != nil {
		dr.logger.Error("error getting listening chats", "err", err)
		return
	}
	if len(avaliableChats.ListeningChats) == 0 {
		dr.logger.Error("error getting listening chats")
		return
	}

	// errPull := make(chan error, 10)
	wg := sync.WaitGroup{}

	for _, chat := range avaliableChats.ListeningChats {

		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
			defer cancel()

			dr.logger.Info("chat_id", chat.Id, "chat_tag", chat.Tag)
			messages, err := dr.chatWorker.GetNewMessages(ctx, chat.Tag)
			if err != nil {
				dr.logger.Error("error getting messages", "err", err)
				return
			}

			if len(messages.Messages) == 0 {
				dr.logger.Info("no new messeges found")
				return
			}

			mongoMessages := db.NewMessages(messages)
			if err := dr.store.InsertMessages(ctx, mongoMessages); err != nil {
				dr.logger.Error("failed to save messages", "err", err)
			}
		}()
	}

	dr.logger.Info("cron_monitoring started")
	wg.Wait()
}
