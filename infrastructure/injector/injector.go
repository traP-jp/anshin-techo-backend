package injector

import (
	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/handler"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"

	"github.com/jmoiron/sqlx"
)

func InjectServer(db *sqlx.DB) (*api.Server, error) {
	repo := repository.New(db)
	h := handler.New(repo)
	s, err := api.NewServer(h, h)
	if err != nil {
		return nil, err
	}

	return s, nil
}
