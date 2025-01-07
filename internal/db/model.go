package db

type ListeningChats struct {
	UserId         int64              `bson:"user_id"`
	ListeningChats []TgListeningChats `bson:"listening_chats"`
}

type TgListeningChats struct {
	Id  int64  `bson:"id"`
	Tag string `bson:"tag"`
}
