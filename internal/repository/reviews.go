package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const reviewStatusActive = "active"

var (
	ErrNoteNotFound        = fmt.Errorf("note not found")
	ErrReviewerNotFound    = fmt.Errorf("reviewer not found")
	ErrReviewAlreadyExists = fmt.Errorf("review already exists")
	ErrInvalidReviewType   = fmt.Errorf("invalid review type")
	ErrInvalidReviewWeight = fmt.Errorf("invalid review weight")
)

type Review struct {
	ID        int64          `db:"id"`
	NoteID    int64          `db:"note_id"`
	Type      string         `db:"type"`
	Status    string         `db:"status"`
	Weight    int            `db:"weight"`
	Author    string         `db:"author"`
	Comment   sql.NullString `db:"comment"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type CreateReviewParams struct {
	Type      string
	Weight    int
	WeightSet bool
	Comment   sql.NullString
}

func (r *Repository) CreateReview(ctx context.Context, ticketID, noteID int64, reviewer string, params CreateReviewParams) (*Review, error) {
	if !isValidReviewType(params.Type) {
		return nil, ErrInvalidReviewType
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("failed to rollback: %v\n", err)
		}
	}()

	var noteStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT status FROM notes WHERE id = ? AND ticket_id = ? AND deleted_at IS NULL FOR UPDATE
	`, noteID, ticketID).Scan(&noteStatus); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoteNotFound
		}

		return nil, fmt.Errorf("select note: %w", err)
	}

	var role string
	if err := tx.QueryRowContext(ctx, `SELECT role FROM users WHERE traq_id = ?`, reviewer).Scan(&role); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReviewerNotFound
		}

		return nil, fmt.Errorf("select reviewer role: %w", err)
	}

	weight, err := normalizeReviewWeight(params, role)
	if err != nil {
		return nil, err
	}

	if err := ensureReviewerNotDuplicated(ctx, tx, noteID, reviewer); err != nil {
		return nil, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO reviews (note_id, type, status, weight, author, comment)
		VALUES (?, ?, ?, ?, ?, ?)
	`, noteID, params.Type, reviewStatusActive, weight, reviewer, params.Comment)
	if err != nil {
		return nil, fmt.Errorf("insert review: %w", err)
	}

	reviewID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	if err := maybeUpdateNoteStatus(ctx, tx, noteID, noteStatus); err != nil {
		return nil, err
	}

	review := new(Review)
	if err := tx.GetContext(ctx, review, `
		SELECT id, note_id, type, status, weight, author, comment, created_at, updated_at
		FROM reviews
		WHERE id = ?
	`, reviewID); err != nil {
		return nil, fmt.Errorf("select review: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return review, nil
}

func isValidReviewType(t string) bool {
	switch t {
	case "approve", "cr", "comment", "system":
		return true
	default:
		return false
	}
}

func normalizeReviewWeight(params CreateReviewParams, role string) (int, error) {
	if params.Type != "approve" {
		return 0, nil
	}

	maxWeight := 0
	switch role {
	case "manager":
		maxWeight = 5
	case "assistant":
		maxWeight = 4
	default:
		maxWeight = 0
	}

	if !params.WeightSet || params.Weight <= 0 || params.Weight > maxWeight {
		return 0, ErrInvalidReviewWeight
	}

	return params.Weight, nil
}

func ensureReviewerNotDuplicated(ctx context.Context, tx *sqlx.Tx, noteID int64, reviewer string) error {
	var exists int
	if err := tx.QueryRowContext(ctx, `
		SELECT 1
		FROM reviews
		WHERE note_id = ? AND author = ? AND status = 'active' AND deleted_at IS NULL
		LIMIT 1
	`, noteID, reviewer).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}

		return fmt.Errorf("check existing review: %w", err)
	}

	return ErrReviewAlreadyExists
}

func maybeUpdateNoteStatus(ctx context.Context, tx *sqlx.Tx, noteID int64, currentStatus string) error {
	var totalWeight int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(weight), 0)
		FROM reviews
		WHERE note_id = ? AND status = 'active' AND deleted_at IS NULL AND type = 'approve'
	`, noteID).Scan(&totalWeight); err != nil {
		return fmt.Errorf("sum review weights: %w", err)
	}

	if totalWeight < 5 || currentStatus == "waiting_sent" {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE notes SET status = 'waiting_sent', updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, noteID); err != nil {
		return fmt.Errorf("update note status: %w", err)
	}

	return nil
}
