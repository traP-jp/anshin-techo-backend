package injector

import (
	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/handler"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
	photo_service "github.com/traP-jp/anshin-techo-backend/internal/service/photo"

	"github.com/jmoiron/sqlx"
)

func InjectServer(db *sqlx.DB) (*api.Server, error) {
	photo := photo_service.NewPhotoService()
	repo := repository.New(db)
	h := handler.New(photo, repo)

	s, err := api.NewServer(h)
	if err != nil {
		return nil, err
	}

	return s, nil
}
