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

// --- Notes ---

// TicketsTicketIdNotesPost implements POST /tickets/{ticketId}/notes operation.
func (h *Handler) TicketsTicketIdNotesPost(_ context.Context, _ *api.TicketsTicketIdNotesPostReq, _ api.TicketsTicketIdNotesPostParams) (*api.Note, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdPut implements PUT /tickets/{ticketId}/notes/{noteId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdPut(_ context.Context, _ *api.TicketsTicketIdNotesNoteIdPutReq, _ api.TicketsTicketIdNotesNoteIdPutParams) (api.TicketsTicketIdNotesNoteIdPutRes, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdDelete implements DELETE /tickets/{ticketId}/notes/{noteId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdDelete(_ context.Context, _ api.TicketsTicketIdNotesNoteIdDeleteParams) error {
	return fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdReviewsReviewIdDelete implements DELETE /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdReviewsReviewIdDelete(_ context.Context, _ api.TicketsTicketIdNotesNoteIdReviewsReviewIdDeleteParams) error {
	return fmt.Errorf("not implemented")
}

//revive:enable:var-naming
