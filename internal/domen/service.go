package domen

import (
	"context"
	"fmt"
	"os"
	"sync"
	"tg-listener/configs"
	"tg-listener/internal/db"
	"time"

	"github.com/google/uuid"
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

			for i, msg := range mongoMessages {
				if path, err := dr.saveMedia(msg.MediaID); err != nil {
					mongoMessages[i].Path = path
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

func (dr *DomenRepository) saveMedia(fileID int32) (string, error) {
	appConfigs, err := configs.AppConfig()
	if err != nil {
		return "", err
	}

	fmt.Println(appConfigs.MediaDefaultDirectory)

	mediaFile, err := dr.chatWorker.GetMediaFile(fileID)
	if err != nil {
		return "", err
	}

	if mediaFile.Local.IsDownloadingCompleted {
		dr.logger.Info("media file already downloaded", "file path", mediaFile.Local.Path)
		return mediaFile.Local.Path, nil
	}

	downloadedFile, err := dr.chatWorker.DownlaodFile(fileID)

	saveFile(downloadedFile.Local.Path)

	return downloadedFile.Local.Path, nil
}

func saveFile(filePath string) {
	newPath := "./downloads/" + uuid.New().String() + ".jpg"

	input, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	err = os.WriteFile(newPath, input, 0644)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	fmt.Println("File saved successfully:", newPath)
}
