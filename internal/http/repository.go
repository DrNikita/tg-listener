package http

import (
	"log/slog"
	"tg-listener/internal/telegram"

	"github.com/gofiber/fiber/v3"
)

type httpRepository struct {
	tgClientAuthorizer telegram.TgClientAuthorizer
	chat               telegram.TgChatWorker
	logger             *slog.Logger
}

func NewHttpRepository(tgClientAuthorizer telegram.TgClientAuthorizer, chat telegram.TgChatWorker, logger *slog.Logger) httpRepository {
	return httpRepository{
		tgClientAuthorizer: tgClientAuthorizer,
		chat:               chat,
		logger:             logger,
	}
}

func (hr httpRepository) SetupRouts(app *fiber.App) {
	mainGroup := app.Group("/api/v1")
	mainGroup.Post("/authorize", hr.AuthorizeTgUser)
}

func (hr httpRepository) AuthorizeTgUser(c fiber.Ctx) error {
	//get phone/pass

	return nil
}
