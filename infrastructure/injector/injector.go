package injector

import (
	"github.com/jmoiron/sqlx"
	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/handler"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
	"github.com/traP-jp/anshin-techo-backend/internal/service/bot"
)

type Dependencies struct {
	DB  *sqlx.DB
	Bot bot.BotClient
}

func InjectServer(deps Dependencies) (*api.Server, error) {
	// photo := photo_service.NewPhotoService()
	repo := repository.New(deps.DB, deps.Bot)
	// h := handler.New(photo, repo, deps.Bot)
	h := handler.New(repo)
	s, err := api.NewServer(h, h)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func InjectBotHandlerService(deps Dependencies) *bot.HandlerService {
	return bot.NewHandlerService(deps.Bot)
}
