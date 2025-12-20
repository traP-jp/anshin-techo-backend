package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

// POST /tickets
// 本職・補佐のみ
func (h *Handler) CreateTicket(ctx context.Context, req *api.CreateTicketReq) (api.CreateTicketRes, error) {
	creator := getUserID(ctx)
	role, err := h.repo.GetUserRoleByTraqID(ctx, creator)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &api.CreateTicketForbidden{}, nil
		}

		return nil, fmt.Errorf("get user role from repository: %w", err)
	}
	if role != "manager" && role != "assistant" {
		return &api.CreateTicketForbidden{}, nil
	}

	description := sql.NullString{String: "", Valid: false}
	if req.Description.Set {
		description = sql.NullString{String: req.Description.Value, Valid: true}
	}

	due := sql.NullTime{Time: time.Time{}, Valid: false}
	if req.Due.Set {
		due = sql.NullTime{Time: req.Due.Value, Valid: true}
	}

	repoTicket := repository.CreateTicketParams{
		Title:        req.Title,
		Description:  description,
		Status:       string(req.Status),
		Assignee:     req.Assignee,
		SubAssignees: req.SubAssignees,
		Stakeholders: req.Stakeholders,
		Due:          due,
		Tags:         req.Tags,
	}

	ticketID, err := h.repo.CreateTicket(ctx, repoTicket)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidStatus) {
			return &api.CreateTicketBadRequest{}, nil
		}
		if errors.Is(err, repository.ErrTagContainsComma) {
			return &api.CreateTicketBadRequest{}, nil
		}

		return nil, fmt.Errorf("create ticket in repository: %w", err)
	}
	ticket, err := h.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("get created ticket from repository: %w", err)
	}
	res := &api.Ticket{
		ID:           ticket.ID,
		Title:        ticket.Title,
		Description:  api.OptString{Value: ticket.Description.String, Set: ticket.Description.Valid},
		Due:          api.OptNilDate{Value: ticket.Due.Time, Set: ticket.Due.Valid, Null: !ticket.Due.Valid},
		Status:       api.TicketStatus(ticket.Status),
		Assignee:     ticket.Assignee,
		SubAssignees: ticket.SubAssignees,
		Stakeholders: ticket.Stakeholders,
		Tags:         ticket.Tags,
		CreatedAt:    ticket.CreatedAt,
		UpdatedAt:    api.OptDateTime{Value: ticket.UpdatedAt, Set: true},
	}

	return res, nil
}

// GET /tickets
// 誰でも
func (h *Handler) GetTickets(ctx context.Context, params api.GetTicketsParams) (api.GetTicketsRes, error) {
	repoParams := repository.GetTicketsParams{
		Assignee: "",
		Status:   "",
		Sort:     "",
	}
	if params.Assignee.Set {
		repoParams.Assignee = params.Assignee.Value
	}
	if params.Status.Set {
		repoParams.Status = string(params.Status.Value)
	}
	if params.Sort.Set {
		repoParams.Sort = string(params.Sort.Value)
	}
	tickets, err := h.repo.GetTickets(ctx, repoParams)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidStatus) {
			return &api.GetTicketsBadRequest{}, nil
		}
		if errors.Is(err, repository.ErrInvalidSort) {
			return &api.GetTicketsBadRequest{}, nil
		}

		return nil, fmt.Errorf("get tickets from repository: %w", err)
	}

	res := make([]api.Ticket, 0, len(tickets))
	for _, ticket := range tickets {

		res = append(res, api.Ticket{
			ID:    ticket.ID,
			Title: ticket.Title,
			Description: api.OptString{
				Value: ticket.Description.String,
				Set:   ticket.Description.Valid,
			},
			Due: api.OptNilDate{
				Value: ticket.Due.Time,
				Set:   ticket.Due.Valid,
				Null:  !ticket.Due.Valid,
			},
			Status:       api.TicketStatus(ticket.Status),
			Assignee:     ticket.Assignee,
			SubAssignees: ticket.SubAssignees,
			Stakeholders: ticket.Stakeholders,
			Tags:         ticket.Tags,
			CreatedAt:    ticket.CreatedAt,
			UpdatedAt: api.OptDateTime{
				Value: ticket.UpdatedAt,
				Set:   true,
			},
		})
	}
	result := api.GetTicketsOKApplicationJSON(res)
	
	return &result, nil
}

// DELETE /tickets/{ticketId}
// 本職のみ
func (h *Handler) DeleteTicketByID(ctx context.Context, params api.DeleteTicketByIDParams) (api.DeleteTicketByIDRes, error) {
	deleter := getUserID(ctx)
	role, err := h.repo.GetUserRoleByTraqID(ctx, deleter)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &api.DeleteTicketByIDForbidden{}, nil
		}

		return nil, fmt.Errorf("get user role from repository: %w", err)
	}
	if role != "manager" {
		return &api.DeleteTicketByIDForbidden{}, nil
	}
	
	id := params.TicketId
	if err := h.repo.DeleteTicket(ctx, id); err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return &api.DeleteTicketByIDNotFound{}, nil
		}

		return nil, fmt.Errorf("delete ticket in repository: %w", err)
	}
	result := api.DeleteTicketByIDNoContent{}

	return &result, nil
}

// GET /tickets/{ticketId}
func (h *Handler) GetTicketByID(ctx context.Context, params api.GetTicketByIDParams) (api.GetTicketByIDRes, error) {
	_, ok := traqIDFromContext(ctx)
	if !ok {
		return &api.GetTicketByIDUnauthorized{}, nil
	}

	id := params.TicketId
	ticket, err := h.repo.GetTicketByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return &api.GetTicketByIDNotFound{}, nil
		}

		return nil, fmt.Errorf("get ticket from repository: %w", err)
	}
	res := &api.GetTicketByIDOK{
		ID:           ticket.ID,
		Title:        ticket.Title,
		Description:  api.OptString{Value: ticket.Description.String, Set: ticket.Description.Valid},
		Due:          api.OptNilDate{Value: ticket.Due.Time, Set: ticket.Due.Valid, Null: !ticket.Due.Valid},
		Status:       api.TicketStatus(ticket.Status),
		Assignee:     ticket.Assignee,
		SubAssignees: ticket.SubAssignees,
		Stakeholders: ticket.Stakeholders,
		Tags:         ticket.Tags,
		CreatedAt:    ticket.CreatedAt,
		UpdatedAt:    api.OptDateTime{Value: ticket.UpdatedAt, Set: true},
		Notes:        []api.Note{}, // TODO: ノート機能実装時に追加
	}

	return res, nil
}

// PATCH /tickets/{ticketId}
// 本職・補佐・関係者のみ
func (h *Handler) UpdateTicketByID(ctx context.Context, req api.OptUpdateTicketByIDReq, params api.UpdateTicketByIDParams) (api.UpdateTicketByIDRes, error) {
	id := params.TicketId
	if !req.Set {
		return nil, fmt.Errorf("request body is required")
	}

	ticket, err := h.repo.GetTicketByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return &api.UpdateTicketByIDNotFound{}, nil
		}

		return nil, fmt.Errorf("get ticket from repository: %w", err)
	}

	updater := getUserID(ctx)
	role, err := h.repo.GetUserRoleByTraqID(ctx, updater)
	if err != nil {
		return nil, fmt.Errorf("get user role from repository: %w", err)
	}
	ok := false
	if role == "manager" || role == "assistant" {
		ok = true
	} else {
		for _, authorized := range append(append(ticket.Stakeholders, ticket.SubAssignees...), ticket.Assignee) {
			if authorized == updater {
				ok = true
				break
			}
		}
	}
	if !ok {
		return &api.UpdateTicketByIDForbidden{}, nil
	}

	title := ticket.Title
	if req.Value.Title.Set {
		title = req.Value.Title.Value
	}
	description := ticket.Description
	if req.Value.Description.Set {
		if req.Value.Description.Value == "" {
			description = sql.NullString{String: "", Valid: false}
		} else {
			description = sql.NullString{
				String: req.Value.Description.Value,
				Valid:  true,
			}
		}
	}
	status := ticket.Status
	if req.Value.Status.Set {
		status = string(req.Value.Status.Value)
	}
	assignee := ticket.Assignee
	if req.Value.Assignee.Set {
		assignee = req.Value.Assignee.Value
	}
	subAssignees := ticket.SubAssignees
	if req.Value.SubAssignees != nil {
		subAssignees = req.Value.SubAssignees
	}
	stakeholders := ticket.Stakeholders
	if req.Value.Stakeholders != nil {
		stakeholders = req.Value.Stakeholders
	}
	due := ticket.Due
	if req.Value.Due.Set {
		if req.Value.Due.Value.IsZero() {
			due = sql.NullTime{Time: time.Time{}, Valid: false}
		} else {
			due = sql.NullTime{
				Time:  req.Value.Due.Value,
				Valid: true,
			}
		}
	}
	tags := ticket.Tags
	if req.Value.Tags != nil {
		tags = req.Value.Tags
	}
	updateParams := repository.CreateTicketParams{
		Title:        title,
		Description:  description,
		Status:       status,
		Assignee:     assignee,
		SubAssignees: subAssignees,
		Stakeholders: stakeholders,
		Due:          due,
		Tags:         tags,
	}
	if err := h.repo.UpdateTicket(ctx, id, updateParams); err != nil {
		if errors.Is(err, repository.ErrInvalidStatus) {
			return &api.UpdateTicketByIDBadRequest{}, nil
		}
		if errors.Is(err, repository.ErrTagContainsComma) {
			return &api.UpdateTicketByIDBadRequest{}, nil
		}
		
		return nil, fmt.Errorf("update ticket in repository: %w", err)
	}
	
	return &api.UpdateTicketByIDOK{}, nil
}