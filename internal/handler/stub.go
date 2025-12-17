package handler

import (
	"context"
	"fmt"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
)

// --- Config ---
// ConfigGet implements GET /config operation.
func (h *Handler) ConfigGet(_ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// ConfigPost implements POST /config operation.
func (h *Handler) ConfigPost(_ctx context.Context, _req *api.ConfigPostReq) (api.ConfigPostRes, error) {
	return nil, fmt.Errorf("not implemented")
}


// --- Tickets ---
// TicketsGet implements GET /tickets operation.
func (h *Handler) TicketsGet(_ctx context.Context, _params api.TicketsGetParams) (api.TicketsGetRes, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsPost implements POST /tickets operation.
func (h *Handler) TicketsPost(_ctx context.Context, _req *api.TicketsPostReq) (api.TicketsPostRes, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdDelete implements DELETE /tickets/{ticketId} operation.
func (h *Handler) TicketsTicketIdDelete(_ctx context.Context, _params api.TicketsTicketIdDeleteParams) (api.TicketsTicketIdDeleteRes, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdGet implements GET /tickets/{ticketId} operation.
func (h *Handler) TicketsTicketIdGet(_ctx context.Context, _params api.TicketsTicketIdGetParams) (api.TicketsTicketIdGetRes, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdPatch implements PATCH /tickets/{ticketId} operation.
func (h *Handler) TicketsTicketIdPatch(_ctx context.Context, _req api.OptTicketsTicketIdPatchReq, _params api.TicketsTicketIdPatchParams) (api.TicketsTicketIdPatchRes, error) {
	return nil, fmt.Errorf("not implemented")
}

// --- Notes ---

// TicketsTicketIdNotesPost implements POST /tickets/{ticketId}/notes operation.
func (h *Handler) TicketsTicketIdNotesPost(_ctx context.Context, _req *api.TicketsTicketIdNotesPostReq, _params api.TicketsTicketIdNotesPostParams) (*api.Note, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdPut implements PUT /tickets/{ticketId}/notes/{noteId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdPut(_ctx context.Context, _req *api.TicketsTicketIdNotesNoteIdPutReq, _params api.TicketsTicketIdNotesNoteIdPutParams) (api.TicketsTicketIdNotesNoteIdPutRes, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdDelete implements DELETE /tickets/{ticketId}/notes/{noteId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdDelete(_ctx context.Context, _params api.TicketsTicketIdNotesNoteIdDeleteParams) error {
	return fmt.Errorf("not implemented")
}

// --- Reviews ---

// TicketsTicketIdNotesNoteIdReviewsPost implements POST /tickets/{ticketId}/notes/{noteId}/reviews operation.
func (h *Handler) TicketsTicketIdNotesNoteIdReviewsPost(_ctx context.Context, _req *api.TicketsTicketIdNotesNoteIdReviewsPostReq, _params api.TicketsTicketIdNotesNoteIdReviewsPostParams) (*api.Review, error) {
	return nil, fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdReviewsReviewIdPut implements PUT /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdReviewsReviewIdPut(_ctx context.Context, _req api.OptTicketsTicketIdNotesNoteIdReviewsReviewIdPutReq, _params api.TicketsTicketIdNotesNoteIdReviewsReviewIdPutParams) error {
	return fmt.Errorf("not implemented")
}

// TicketsTicketIdNotesNoteIdReviewsReviewIdDelete implements DELETE /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) TicketsTicketIdNotesNoteIdReviewsReviewIdDelete(_ctx context.Context, _params api.TicketsTicketIdNotesNoteIdReviewsReviewIdDeleteParams) error {
	return fmt.Errorf("not implemented")
}
