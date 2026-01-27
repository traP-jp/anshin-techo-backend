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

	role, err := h.repo.GetUserRoleByTraqID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user role: %w", err)
	}

	note, err := h.repo.CreateNote(ctx, params.TicketId, userID, req.Content, string(req.Type))
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}

	safeContent := ApplyCensorIfNeed(role, note.Content)

	return &api.Note{
		ID:       note.ID,
		TicketID: note.TicketID,
		Author:   note.UserID,
		Content:  safeContent,
		Type:     api.NoteType(note.Type),
		Status:   api.NoteStatus(note.Status),
		Reviews:  []api.Review{},

		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}, nil
}

// PUT /tickets/{ticketId}/notes/{noteId}
//
//nolint:revive
func (h *Handler) TicketsTicketIdNotesNoteIdPut(ctx context.Context, req *api.TicketsTicketIdNotesNoteIdPutReq, params api.TicketsTicketIdNotesNoteIdPutParams) (api.TicketsTicketIdNotesNoteIdPutRes, error) {
	if err := h.repo.UpdateNote(ctx, params.TicketId, params.NoteId, req.Content); err != nil {
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
