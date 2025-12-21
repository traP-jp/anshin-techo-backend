package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

func (h *Handler) ConfigGet(ctx context.Context) (api.ConfigGetRes, error) {
	userID := getUserID(ctx)
	role, err := h.repo.GetUserRoleByTraqID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &api.ConfigGetForbidden{}, nil
		}

		return nil, fmt.Errorf("get user role from repository: %w", err)
	}
	if role != "manager" {
		return &api.ConfigGetForbidden{}, nil
	}

	cfg, err := h.repo.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return nil, fmt.Errorf("config not found")
		}

		return nil, fmt.Errorf("get config from repository: %w", err)
	}

	return toAPIConfig(cfg), nil
}

func (h *Handler) ConfigPost(ctx context.Context, req *api.Config) (api.ConfigPostRes, error) {
	userID := getUserID(ctx)
	role, err := h.repo.GetUserRoleByTraqID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &api.ConfigPostForbidden{}, nil
		}

		return nil, fmt.Errorf("get user role from repository: %w", err)
	}
	if role != "manager" {
		return &api.ConfigPostForbidden{}, nil
	}

	repoCfg := toRepositoryConfig(req)
	if err := h.repo.UpsertConfig(ctx, repoCfg); err != nil {
		return nil, fmt.Errorf("upsert config in repository: %w", err)
	}

	updatedCfg, err := h.repo.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config from repository: %w", err)
	}

	return toAPIConfig(updatedCfg), nil
}

func toAPIConfig(cfg *repository.Config) *api.Config {
	overdueDay := cfg.ReminderInterval.OverdueDay
	if overdueDay == nil {
		overdueDay = []int{}
	}

	return &api.Config{
		ReminderInterval: api.ConfigReminderInterval{
			OverdueDay:   overdueDay,
			NotesentHour: cfg.ReminderInterval.NotesentHour,
		},
		RevisePrompt: cfg.RevisePrompt,
	}
}

func toRepositoryConfig(cfg *api.Config) repository.Config {
	overdueDay := cfg.ReminderInterval.OverdueDay
	if overdueDay == nil {
		overdueDay = []int{}
	}

	return repository.Config{
		ReminderInterval: repository.ConfigReminderInterval{
			OverdueDay:   overdueDay,
			NotesentHour: cfg.ReminderInterval.NotesentHour,
		},
		RevisePrompt: cfg.RevisePrompt,
	}
}
