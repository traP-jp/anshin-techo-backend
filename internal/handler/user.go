package handler

import (
	"context"
	"fmt"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

// GET /users
func (h *Handler) UsersGet(ctx context.Context) ([]api.User, error) {
	users, err := h.repo.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get users from repository: %w", err)
	}

	res := make([]api.User, 0, len(users))
	for _, user := range users {
		res = append(res, api.User{
			TraqID: user.TraqID,
			Role:   api.UserRole(user.Role),
		})
	}

	return res, nil
}

// PUT /users
func (h *Handler) UsersPut(ctx context.Context, req []api.User) (api.UsersPutRes, error) {
	// Empty request means sync to empty set; proceed.
	repoUsers := make([]*repository.User, 0, len(req))
	for _, u := range req {
		repoUsers = append(repoUsers, &repository.User{
			TraqID: u.TraqID,
			Role:   string(u.Role),
		})
	}

	if err := h.repo.UpdateUsers(ctx, repoUsers); err != nil {
		return nil, fmt.Errorf("sync users to repository: %w", err)
	}

	return &api.UsersPutOK{}, nil
}
