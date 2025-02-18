package db

import (
	"errors"
	"time"

	"github.com/zelenin/go-tdlib/client"
)

type Message struct {
	Content   `bson:",inline"`
	CreatedAt time.Time `bson:"created_at"`
}

func NewMessage(msg *client.Message) (Message, error) {
	switch content := msg.Content.(type) {
	case *client.MessageText:
		return Message{
			Content: Content{
				Type: ContentText,
				Text: content.Text.Text,
			},
			CreatedAt: time.Now(),
		}, nil
	case *client.MessagePhoto:
		return Message{
			Content: Content{
				Type:    ContentPhoto,
				Text:    content.Caption.Text,
				MediaID: content.Photo.Sizes[len(content.Photo.Sizes)-1].Photo.Id,
			},
			CreatedAt: time.Now(),
		}, nil
	case *client.MessageVideo:
		return Message{
			Content: Content{
				Type:    ContentVideo,
				Text:    content.Caption.Text,
				MediaID: content.Video.Video.Id,
			},
			CreatedAt: time.Now(),
		}, nil
	case *client.MessageVoiceNote:
		return Message{
			Content: Content{
				Type:    ContentVoice,
				Text:    content.Caption.Text,
				MediaID: content.VoiceNote.Voice.Id,
			},
			CreatedAt: time.Now(),
		}, nil
	case *client.MessageDocument:
		return Message{
			Content: Content{
				Type:    ContentDocument,
				Text:    content.Caption.Text,
				MediaID: content.Document.Document.Id,
			},
			CreatedAt: time.Now(),
		}, nil
	default:
		return Message{}, errors.New("incorrect message type")
	}
}

func NewMessages(tgMsg *client.Messages) []Message {
	mongoMessages := make([]Message, tgMsg.TotalCount)

	for _, msg := range tgMsg.Messages {
		if mongoMsg, err := NewMessage(msg); err == nil {
			mongoMessages = append(mongoMessages, mongoMsg)
		}
	}

	return mongoMessages
}

type ListeningChats struct {
	UserId         int64             `bson:"user_id"`
	ListeningChats []TgListeningChat `bson:"listening_chats"`
}

type TgListeningChat struct {
	Id  int64  `bson:"id"`
	Tag string `bson:"tag"`
}

type LastMessage struct {
	ChatId        int64 `bson:"chat_id"`
	LastMessageId int64 `bson:"last_message_id"`
	Offset        int32 `bson:"offset"`
}
