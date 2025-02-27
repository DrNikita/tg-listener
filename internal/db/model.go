package db

import (
	"time"
)

type Message struct {
	Content   `bson:",inline"`
	SenderIDs []int64   `bson:"sender_id"`
	CreatedAt time.Time `bson:"created_at"`
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
