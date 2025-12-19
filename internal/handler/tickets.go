package handler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

// POST /tickets
func (h *Handler) TicketsPost(ctx context.Context, req *api.TicketsPostReq) (api.TicketsPostRes, error) {
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
func (h *Handler) TicketsGet(ctx context.Context, params api.TicketsGetParams) (api.TicketsGetRes, error) {
	repoParams := repository.GetTicketsParams{}
	if params.Assignee.Set {
		repoParams.Assignee = params.Assignee.Value
	} else {
		repoParams.Assignee = ""
	}
	if params.Status.Set {
		repoParams.Status = string(params.Status.Value)
	} else {
		repoParams.Status = ""
	}
	if params.Sort.Set {
		repoParams.Sort = string(params.Sort.Value)
	} else {
		repoParams.Sort = ""
	}
	tickets, err := h.repo.GetTickets(ctx, repoParams)
	if err != nil {
		return nil, fmt.Errorf("get tickets from repository: %w", err)
	}

	res := make([]api.Ticket, 0, len(tickets))
	for _, ticket := range tickets {

		res = append(res, api.Ticket{
			ID: ticket.ID,
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
	result := api.TicketsGetOKApplicationJSON(res)
	
	return &result, nil
}

// DELETE /tickets/{ticketId}
//nolint:revive
func (h *Handler) TicketsTicketIdDelete(ctx context.Context, params api.TicketsTicketIdDeleteParams) (api.TicketsTicketIdDeleteRes, error) {
	id := params.TicketId
	if err := h.repo.DeleteTicket(ctx, id); err != nil {
		if err == repository.ErrTicketNotFound {
			return &api.TicketsTicketIdDeleteNotFound{}, nil
		}
		
		return nil, fmt.Errorf("delete ticket in repository: %w", err)
	}
	result := api.TicketsTicketIdDeleteNoContent{}

	return &result, nil
}

// GET /tickets/{ticketId}
//nolint:revive
func (h *Handler) TicketsTicketIdGet(ctx context.Context, params api.TicketsTicketIdGetParams) (api.TicketsTicketIdGetRes, error) {
	id := params.TicketId
	ticket, err := h.repo.GetTicketByID(ctx, id)
	if err != nil {
		if err == repository.ErrTicketNotFound {
			return &api.TicketsTicketIdGetNotFound{}, nil
		}

		return nil, fmt.Errorf("get ticket from repository: %w", err)
	}
	res := &api.TicketsTicketIdGetOK{
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
//nolint:revive
func (h *Handler) TicketsTicketIdPatch(ctx context.Context, req api.OptTicketsTicketIdPatchReq, params api.TicketsTicketIdPatchParams) (api.TicketsTicketIdPatchRes, error) {
	id := params.TicketId
	if !req.Set {
		return nil, fmt.Errorf("request body is required")
	}

	ticket, err := h.repo.GetTicketByID(ctx, id)
	if err != nil {
		if err == repository.ErrTicketNotFound {
			return &api.TicketsTicketIdPatchNotFound{}, nil
		}
		
		return nil, fmt.Errorf("get ticket from repository: %w", err)
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
		return nil, fmt.Errorf("update ticket in repository: %w", err)
	}
	
	return &api.TicketsTicketIdPatchOK{}, nil
}
