package injector

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/handler"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
	"github.com/traP-jp/anshin-techo-backend/internal/service/bot"
)

type Dependencies struct {
	DB  *sqlx.DB
	Bot bot.Client
}

func InjectServer(deps Dependencies) (*api.Server, error) {
	repo := repository.New(deps.DB) //, deps.Bot)

	botService, ok := deps.Bot.(*bot.Service)
	if !ok {
		return nil, fmt.Errorf("failed to cast bot client to service")
	}

	h := handler.New(repo, botService)
	s, err := api.NewServer(h, h)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func InjectBotHandlerService(deps Dependencies) *bot.HandlerService {
	return bot.NewHandlerService(deps.Bot)
}
