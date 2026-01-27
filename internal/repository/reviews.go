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
	ErrReviewNotFound      = fmt.Errorf("review not found")
	ErrReviewForbidden     = fmt.Errorf("review forbidden")
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
	Type    string
	Weight  int
	Comment sql.NullString
}

type UpdateReviewParams struct {
	Type       string
	TypeSet    bool
	Weight     int
	WeightSet  bool
	Comment    sql.NullString
	CommentSet bool
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
			role = "member"
		} else {
			return nil, fmt.Errorf("select reviewer role: %w", err)
		}
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

	if params.Weight < 0 || params.Weight > maxWeight {
		return 0, ErrInvalidReviewWeight
	}

	return params.Weight, nil
}

func normalizeUpdateWeight(newType string, weightSet bool, weight int, currentWeight int, role string) (int, error) {
	if newType != "approve" {
		return 0, nil
	}

	finalWeight := currentWeight
	if weightSet {
		finalWeight = weight
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

	if finalWeight <= 0 || finalWeight > maxWeight {
		return 0, ErrInvalidReviewWeight
	}

	return finalWeight, nil
}

func (r *Repository) UpdateReview(ctx context.Context, ticketID, noteID, reviewID int64, reviewer string, params UpdateReviewParams) (*Review, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("failed to rollback: %v\n", err)
		}
	}()

	current := new(Review)
	var noteStatus string
	if err := tx.QueryRowxContext(ctx, `
		SELECT r.id, r.note_id, r.type, r.status, r.weight, r.author, r.comment, r.created_at, r.updated_at, n.status AS note_status
		FROM reviews r
		JOIN notes n ON r.note_id = n.id
		WHERE r.id = ? AND r.note_id = ? AND n.ticket_id = ? AND r.deleted_at IS NULL AND n.deleted_at IS NULL
		FOR UPDATE
	`, reviewID, noteID, ticketID).Scan(&current.ID, &current.NoteID, &current.Type, &current.Status, &current.Weight, &current.Author, &current.Comment, &current.CreatedAt, &current.UpdatedAt, &noteStatus); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReviewNotFound
		}

		return nil, fmt.Errorf("select review: %w", err)
	}

	if current.Author != reviewer {
		return nil, ErrReviewForbidden
	}

	newType := current.Type
	if params.TypeSet {
		if !isValidReviewType(params.Type) {
			return nil, ErrInvalidReviewType
		}
		newType = params.Type
	}

	var role string
	if err := tx.QueryRowContext(ctx, `SELECT role FROM users WHERE traq_id = ?`, reviewer).Scan(&role); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReviewerNotFound
		}

		return nil, fmt.Errorf("select reviewer role: %w", err)
	}

	newWeight := current.Weight
	var weightErr error
	if newType == "approve" || params.TypeSet {
		newWeight, weightErr = normalizeUpdateWeight(newType, params.WeightSet, params.Weight, current.Weight, role)
		if weightErr != nil {
			return nil, weightErr
		}
	} else if params.WeightSet {
		newWeight = params.Weight
	}

	newComment := current.Comment
	if params.CommentSet {
		newComment = params.Comment
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE reviews SET type = ?, weight = ?, comment = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, newType, newWeight, newComment, reviewID); err != nil {
		return nil, fmt.Errorf("update review: %w", err)
	}

	if err := maybeUpdateNoteStatus(ctx, tx, noteID, noteStatus); err != nil {
		return nil, err
	}

	updated := new(Review)
	if err := tx.GetContext(ctx, updated, `
		SELECT id, note_id, type, status, weight, author, comment, created_at, updated_at
		FROM reviews
		WHERE id = ?
	`, reviewID); err != nil {
		return nil, fmt.Errorf("select updated review: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return updated, nil
}

func (r *Repository) DeleteReview(ctx context.Context, ticketID, noteID, reviewID int64, reviewer string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("failed to rollback: %v\n", err)
		}
	}()

	var author string
	if err := tx.QueryRowContext(ctx, `
		SELECT r.author
		FROM reviews r
		JOIN notes n ON r.note_id = n.id
		WHERE r.id = ? AND r.note_id = ? AND n.ticket_id = ? AND r.deleted_at IS NULL AND n.deleted_at IS NULL
		FOR UPDATE
	`, reviewID, noteID, ticketID).Scan(&author); err != nil {
		if err == sql.ErrNoRows {
			return ErrReviewNotFound
		}

		return fmt.Errorf("select review: %w", err)
	}

	if author != reviewer {
		return ErrReviewForbidden
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE reviews SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?
	`, reviewID); err != nil {
		return fmt.Errorf("delete review: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
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
