package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

type Handler struct {
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
	repo *repository.Repository,
) *Handler {
	return &Handler{
		//photo,
		repo: repo,
	}
}

func (h *Handler) NewError(ctx context.Context, err error) *api.ErrorResponseStatusCode {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		message := httpErr.Message
		if message == nil {
			message = http.StatusText(httpErr.Code)
		}

		if httpErr.Code >= http.StatusInternalServerError {
			// Log server-side errors to aid debugging while returning a safe message to clients.
			slog.ErrorContext(ctx, "http error", "error", err)
		}

		return &api.ErrorResponseStatusCode{
			StatusCode: httpErr.Code,
			Response:   api.Error{Message: fmt.Sprint(message)},
		}
	}

	slog.ErrorContext(ctx, "internal server error", "error", err)

	return &api.ErrorResponseStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response:   api.Error{Message: "internal server error"},
	}
}
