package handler

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

// POST /tickets
func (h *Handler) TicketsPost(ctx context.Context, req *api.TicketsPostReq) (api.TicketsPostRes, error) {
	description := sql.NullString{Valid: false}
	if req.Description.Set {
		description = sql.NullString{String: *&req.Description.Value, Valid: true}
	}

	due := sql.NullTime{Valid: false}
	if req.Due.Set {
		due = sql.NullTime{Time: *&req.Due.Value, Valid: true}
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
		Due:          api.OptNilDate{Value: ticket.Due.Time, Set: ticket.Due.Valid},
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
	// TODO: 絞り込みはまだ実装していない
	tickets, err := h.repo.GetTickets(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tickets from repository: %w", err)
	}

	res := make([]api.Ticket, 0, len(tickets))
	for _, ticket := range tickets {
		if ticket.DeletedAt.Valid {
			continue
		}

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