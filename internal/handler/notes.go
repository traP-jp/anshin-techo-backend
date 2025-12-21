package handler

import (
	"context"
	"fmt"

	"github.com/traP-jp/anshin-techo-backend/internal/api"
)

// POST /tickets/{ticketId}/notes
//
//nolint:revive
func (h *Handler) TicketsTicketIdNotesPost(ctx context.Context, req *api.TicketsTicketIdNotesPostReq, params api.TicketsTicketIdNotesPostParams) (api.TicketsTicketIdNotesPostRes, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("user not found in context (unauthorized)")
	}

	note, err := h.repo.CreateNote(ctx, params.TicketId, userID, req.Content, string(req.Type))
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}

	return &api.Note{
		ID:       note.ID,
		TicketID: note.TicketID,
		Author:   note.UserID,
		Content:  note.Content,
		Type:     api.NoteType(note.Type),
		Status:   api.OptNoteStatus{Value: "", Set: false},
		Reviews:  []api.Review{},

		CreatedAt: api.OptDateTime{
			Value: note.CreatedAt,
			Set:   true,
		},
		UpdatedAt: api.OptDateTime{
			Value: note.UpdatedAt,
			Set:   true,
		},
	}, nil
}

// PUT /tickets/{ticketId}/notes/{noteId}
//
//nolint:revive
func (h *Handler) TicketsTicketIdNotesNoteIdPut(ctx context.Context, req *api.TicketsTicketIdNotesNoteIdPutReq, params api.TicketsTicketIdNotesNoteIdPutParams) (api.TicketsTicketIdNotesNoteIdPutRes, error) {
	if !req.Content.Set {
		return nil, fmt.Errorf("content is required for update")
	}

	if err := h.repo.UpdateNote(ctx, params.TicketId, params.NoteId, req.Content.Value); err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}

	return &api.TicketsTicketIdNotesNoteIdPutOK{}, nil
}

// DELETE /tickets/{ticketId}/notes/{noteId}
//
//nolint:revive
func (h *Handler) TicketsTicketIdNotesNoteIdDelete(ctx context.Context, params api.TicketsTicketIdNotesNoteIdDeleteParams) (api.TicketsTicketIdNotesNoteIdDeleteRes, error) {
	if err := h.repo.DeleteNote(ctx, params.TicketId, params.NoteId); err != nil {
		return nil, fmt.Errorf("delete note: %w", err)
	}

	return &api.TicketsTicketIdNotesNoteIdDeleteNoContent{}, nil
}
