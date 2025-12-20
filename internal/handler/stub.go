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

// --- Reviews ---

// TicketsTicketIdNotesNoteIdReviewsPost implements POST /tickets/{ticketId}/notes/{noteId}/reviews operation.
func (h *Handler) TicketsTicketIdNotesNoteIdReviewsPost(_ context.Context, _ *api.TicketsTicketIdNotesNoteIdReviewsPostReq, _ api.TicketsTicketIdNotesNoteIdReviewsPostParams) (*api.Review, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdReviewsReviewIdPut implements PUT /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdReviewsReviewIdPut(_ context.Context, _ api.OptTicketsTicketIdNotesNoteIdReviewsReviewIdPutReq, _ api.TicketsTicketIdNotesNoteIdReviewsReviewIdPutParams) error {
	return fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdReviewsReviewIdDelete implements DELETE /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdReviewsReviewIdDelete(_ context.Context, _ api.TicketsTicketIdNotesNoteIdReviewsReviewIdDeleteParams) error {
	return fmt.Errorf("not implemented")
}

//revive:enable:var-naming
