package handler

//revive:disable:var-naming

import (
	"context"
	"fmt"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
)

// --- Config ---
// ConfigGet implements GET /config operation.
func (h *Handler) ConfigGet(_ context.Context) error {
	return fmt.Errorf("not implemented")
}

// ConfigPost implements POST /config operation.
func (h *Handler) ConfigPost(_ context.Context, _ *api.ConfigPostReq) (api.ConfigPostRes, error) {
	return nil, fmt.Errorf("not implemented")
}

