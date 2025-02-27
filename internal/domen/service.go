package domen

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"tg-listener/configs"
	"tg-listener/internal/db"
	"time"

	"github.com/zelenin/go-tdlib/client"
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

			mongoMessages := make([]db.Message, 0)
			for _, msg := range messages.Messages {

				if mongoMessage, err := newMongoMessage(msg); err == nil {

					commentstorIDs := make([]int64, 0)

					if comments, err := dr.chatWorker.GetComments(msg.Id, msg.ChatId); err == nil {
						for _, comment := range comments {
							commentstorIDs = append(commentstorIDs, getSenderId(comment))
						}

						mongoMessage.SenderIDs = commentstorIDs
					}

					mongoMessages = append(mongoMessages, mongoMessage)
				}
			}

			for i, msg := range mongoMessages {
				if path, err := dr.saveFile(msg.FileID); err != nil {
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

func (dr *DomenRepository) Spam() {
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

func newMongoMessage(msg *client.Message) (db.Message, error) {
	switch content := msg.Content.(type) {
	case *client.MessageText:
		return db.Message{
			Content: db.Content{
				Type: db.ContentText,
				Text: content.Text.Text,
			},
			CreatedAt: time.Now(),
		}, nil
	case *client.MessagePhoto:
		return db.Message{
			Content: db.Content{
				Type: db.ContentPhoto,
				//TODO: mb there is a better way to get Photo text (existing method)
				Text:   content.Caption.Text,
				FileID: content.Photo.Sizes[len(content.Photo.Sizes)-1].Photo.Id,
			},
			CreatedAt: time.Now(),
		}, nil
	case *client.MessageVideo:
		return db.Message{
			Content: db.Content{
				Type:   db.ContentVideo,
				Text:   content.Caption.Text,
				FileID: content.Video.Video.Id,
			},
			CreatedAt: time.Now(),
		}, nil
	case *client.MessageVoiceNote:
		return db.Message{
			Content: db.Content{
				Type:   db.ContentVoice,
				Text:   content.Caption.Text,
				FileID: content.VoiceNote.Voice.Id,
			},
			CreatedAt: time.Now(),
		}, nil
	case *client.MessageDocument:
		return db.Message{
			Content: db.Content{
				Type:   db.ContentDocument,
				Text:   content.Caption.Text,
				FileID: content.Document.Document.Id,
			},
			CreatedAt: time.Now(),
		}, nil
	default:
		return db.Message{}, errors.New("incorrect message type")
	}
}

func getSenderId(msg *client.Message) int64 {
	switch sender := msg.SenderId.(type) {
	case *client.MessageSenderUser:
		return sender.UserId
	case *client.MessageSenderChat:
		return sender.ChatId
	default:
		return 0
	}
}
