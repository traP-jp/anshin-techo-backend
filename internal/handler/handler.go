package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
	"github.com/traP-jp/anshin-techo-backend/internal/service/photo"
)

type Handler struct {
	photo *photo.Service
	repo  *repository.Repository
}

func New(
	photo *photo.Service,
	repo *repository.Repository,
) *Handler {
	return &Handler{
		photo,
		repo,
	}
}

func (h *Handler) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	if apiErr, ok := err.(*api.ErrorStatusCode); ok {
		return apiErr
	}

	slog.ErrorContext(ctx, "internal server error", "error", err)

	return &api.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: api.Error{
			Message: "internal server error",
		},
	}
}
