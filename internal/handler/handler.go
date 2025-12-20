package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

type Handler struct {
	//photo *photo.Service
	repo *repository.Repository
}

type ctxKey string

const traqIDCtxKey ctxKey = "traq_id"

func traqIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	id, ok := ctx.Value(traqIDCtxKey).(string)
	if !ok || id == "" {
		return "", false
	}

	return id, true
}

func New(
	//photo *photo.Service,
	repo *repository.Repository,
) *Handler {
	return &Handler{
		//photo,
		repo: repo,
	}
}

func (h *Handler) NewError(ctx context.Context, err error) error {
	if _, ok := err.(*echo.HTTPError); ok {
		return err
	}

	slog.ErrorContext(ctx, "internal server error", "error", err)

	return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
}

func (h *Handler) HandleTraQAuth(ctx context.Context, _ api.OperationName, t api.TraQAuth) (context.Context, error) {
	return context.WithValue(ctx, traqIDCtxKey, t.APIKey), nil
}
