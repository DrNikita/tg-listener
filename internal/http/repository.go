package http

import (
	"log/slog"
	"tg-listener/internal/telegram"

	"github.com/gofiber/fiber/v3"
)

type httpRepository struct {
	chat   telegram.TgChatWorker
	logger *slog.Logger
}

func NewHttpRepository(chat telegram.TgChatWorker, logger *slog.Logger) httpRepository {
	return httpRepository{
		chat:   chat,
		logger: logger,
	}
}

func (hr httpRepository) AddListenableChat(c fiber.Ctx) error {
	return nil
}
