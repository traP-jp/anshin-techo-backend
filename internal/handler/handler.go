package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

type Handler struct {
	//photo *photo.Service
	repo *repository.Repository
}

func New(
	//photo *photo.Service,
	repo *repository.Repository,
) *Handler {
	return &Handler{
		//photo,
		repo,
	}
}

func (h *Handler) NewError(ctx context.Context, err error) error {
	if _, ok := err.(*echo.HTTPError); ok {
		return err
	}

	slog.ErrorContext(ctx, "internal server error", "error", err)

	return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
}
