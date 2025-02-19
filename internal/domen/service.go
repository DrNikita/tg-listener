package domen

import (
	"context"
	"fmt"
	"sync"
	"tg-listener/configs"
	"tg-listener/internal/db"
	"time"
)

func (dr *DomenRepository) BackgroundListening() {
	//TODO:remove
	// defer func() {
	// 	dr.store.DropDB(context.Background())
	// 	os.RemoveAll("./.tdlib")
	// }()

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

			for i, msg := range mongoMessages {
				if path, err := dr.saveFile(msg.FileID); err != nil {
					mongoMessages[i].Path = path
					fmt.Println("_____________________________________________", mongoMessages[i].Path)
				}
			}

			if err := dr.store.InsertMessages(ctx, mongoMessages); err != nil {
				dr.logger.Error("failed to save messages", "err", err)
			}
		}()
	}

	dr.logger.Info("cron_monitoring started")
	wg.Wait()
}

func (dr *DomenRepository) saveFile(fileID int32) (string, error) {
	appConfigs, err := configs.AppConfig()
	if err != nil {
		return "", err
	}

	fmt.Println(appConfigs.MediaDefaultDirectory)

	mediaFile, err := dr.chatWorker.GetFile(fileID)
	if err != nil {
		return "", err
	}

	if mediaFile.Local.IsDownloadingCompleted {
		dr.logger.Info("media file already downloaded", "file path", mediaFile.Local.Path)
		return mediaFile.Local.Path, nil
	}

	downloadedFile, err := dr.chatWorker.DownlaodFile(fileID)
	if err != nil {
		dr.logger.Error("failed to download file", "file_ID", fileID, "err", err)
		return "", err
	}

	return downloadedFile.Local.Path, nil
}
