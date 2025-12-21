package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Note struct {
	ID        int64        `db:"id"`
	TicketID  int64        `db:"ticket_id"`
	UserID    string       `db:"author"`
	Content   string       `db:"content"`
	Type      string       `db:"type"`
	Status    string       `db:"status"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (r *Repository) CreateNote(ctx context.Context, ticketID int64, author, content, noteType string) (*Note, error) {
	query := `
		INSERT INTO notes (ticket_id, author, content, type, status)
		VALUES (?, ?, ?, ?, 'draft')`

	result, err := r.db.ExecContext(ctx, query, ticketID, author, content, noteType)
	if err != nil {
		return nil, fmt.Errorf("insert note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	note := &Note{
		ID:        0,
		TicketID:  0,
		UserID:    "",
		Content:   "",
		Type:      "",
		Status:    "",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
		DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
	}
	getQuery := `SELECT * FROM notes WHERE id = ?`
	if err := r.db.GetContext(ctx, note, getQuery, id); err != nil {
		return nil, fmt.Errorf("get created note: %w", err)
	}

	return note, nil
}

func (r *Repository) GetNotes(ctx context.Context, ticketID int64) ([]*Note, error) {
	notes := []*Note{}
	query := `SELECT * FROM notes WHERE ticket_id = ? AND deleted_at IS NULL ORDER BY created_at ASC`

	if err := r.db.SelectContext(ctx, &notes, query, ticketID); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *Repository) GetReviewsByNoteIDs(ctx context.Context, ticketID int64, noteIDs []int64) ([]*Review, error) {
	if len(noteIDs) == 0 {
		return []*Review{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT r.id, r.note_id, r.type, r.status, r.weight, r.author, r.comment, r.created_at, r.updated_at
		FROM reviews r
		JOIN notes n ON r.note_id = n.id
		WHERE r.note_id IN (?) AND n.ticket_id = ? AND r.deleted_at IS NULL AND n.deleted_at IS NULL
		ORDER BY r.created_at ASC
	`, noteIDs, ticketID)
	if err != nil {
		return nil, fmt.Errorf("build review select query: %w", err)
	}

	query = r.db.Rebind(query)

	reviews := []*Review{}
	if err := r.db.SelectContext(ctx, &reviews, query, args...); err != nil {
		return nil, fmt.Errorf("select reviews: %w", err)
	}

	return reviews, nil
}

func (r *Repository) UpdateNote(ctx context.Context, ticketID, noteID int64, content string) error {
	query := `UPDATE notes SET content = ?, updated_at = NOW() WHERE id = ? AND ticket_id = ?`

	result, err := r.db.ExecContext(ctx, query, content, noteID, ticketID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) DeleteNote(ctx context.Context, ticketID, noteID int64) error {
	query := `DELETE FROM notes WHERE id = ? AND ticket_id = ?`

	result, err := r.db.ExecContext(ctx, query, noteID, ticketID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) GetNoteByID(ctx context.Context, ticketID, noteID int64) (*Note, error) {
	note := &Note{}
	query := `SELECT * FROM notes WHERE id = ? AND ticket_id = ?`

	if err := r.db.GetContext(ctx, note, query, noteID, ticketID); err != nil {
		return nil, err
	}
	return note, nil
}
