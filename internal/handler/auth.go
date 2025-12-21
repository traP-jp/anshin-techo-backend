package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/anshin-techo-backend/internal/api"
)

type contextKey string

const userKey contextKey = "user"

func (h *Handler) HandleTraQAuth(ctx context.Context, _ string, t api.TraQAuth) (context.Context, error) {
	if t.APIKey == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing user header")
	}

	return context.WithValue(ctx, userKey, t.APIKey), nil
}

// getUserID : ユーザーIDをコンテキストから取得
func getUserID(ctx context.Context) string {
	if v, ok := ctx.Value(userKey).(string); ok {
		return v
	}

	return ""
}
