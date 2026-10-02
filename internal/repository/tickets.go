package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type (
	Ticket struct {
		ID           int64          `db:"id"`
		Title        string         `db:"title"`
		Status       string         `db:"status"`
		Assignee     string         `db:"assignee"`
		Due          sql.NullTime   `db:"due"`
		Description  sql.NullString `db:"description"`
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`
		DeletedAt    sql.NullTime   `db:"deleted_at"`
		SubAssignees []string       `db:"-"`
		Stakeholders []string       `db:"-"`
		Tags         []string       `db:"-"`
	}

	CreateTicketParams struct {
		Title        string
		Description  sql.NullString
		Status       string
		Assignee     string
		SubAssignees []string
		Stakeholders []string
		Due          sql.NullTime
		Tags         []string
	}

	GetTicketsParams struct {
		Assignee string
		Status   string
		Sort     string
	}
)

var (
	ErrTicketNotFound   = fmt.Errorf("ticket not found")
	ErrInvalidStatus    = fmt.Errorf("invalid status")
	ErrInvalidSort      = fmt.Errorf("invalid sort option")
	ErrTagContainsComma = fmt.Errorf("tag contains comma")
)

func validateStatus(status string) error {
	switch status {
	case "not_planned", "not_written", "waiting_review", "waiting_sent", "sent", "milestone_scheduled", "completed", "forgotten":
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidStatus, status)
	}
}

func validateTicketSort(sort string) error {
	switch sort {
	case "due_asc", "due_desc", "created_desc":
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSort, sort)
	}
}

func validateTags(tags []string) error {
	for _, tag := range tags {
		if strings.Contains(tag, ",") {
			return fmt.Errorf("%w: %s", ErrTagContainsComma, tag)
		}
	}

	return nil
}

func (r *Repository) GetTickets(ctx context.Context, params GetTicketsParams) ([]*Ticket, error) {
	query := `
		SELECT
			t.id, t.title, t.status, t.assignee, t.due, t.description, t.created_at, t.updated_at, t.deleted_at,
			GROUP_CONCAT(DISTINCT tsa.sub_assignee) AS sub_assignees,
			GROUP_CONCAT(DISTINCT ts.stakeholder) AS stakeholders,
			GROUP_CONCAT(DISTINCT tt.tag) AS tags
		FROM tickets t
		LEFT JOIN ticket_sub_assignees tsa ON t.id = tsa.ticket_id
		LEFT JOIN ticket_stakeholders ts ON t.id = ts.ticket_id
		LEFT JOIN ticket_tags tt ON t.id = tt.ticket_id
		WHERE t.deleted_at IS NULL`

	args := []interface{}{}
	if params.Assignee != "" {
		query += " AND t.assignee = ?"
		args = append(args, params.Assignee)
	}
	if params.Status != "" {
		if err := validateStatus(params.Status); err != nil {
			return nil, err
		}
		query += " AND t.status = ?"
		args = append(args, params.Status)
	}
	query += " GROUP BY t.id"
	if err := validateTicketSort(params.Sort); err == nil {
		switch params.Sort {
		case "due_asc":
			query += " ORDER BY t.due ASC"
		case "due_desc":
			query += " ORDER BY t.due DESC"
		case "created_desc":
			query += " ORDER BY t.created_at DESC"
		}
	} else {
		query += " ORDER BY t.created_at DESC"
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		var t Ticket
		var subAssignees, stakeholders, tags sql.NullString
		err := rows.Scan(
			&t.ID, &t.Title, &t.Status, &t.Assignee, &t.Due, &t.Description, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
			&subAssignees, &stakeholders, &tags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		if subAssignees.Valid && subAssignees.String != "" {
			t.SubAssignees = strings.Split(subAssignees.String, ",")
		} else {
			t.SubAssignees = []string{}
		}
		if stakeholders.Valid && stakeholders.String != "" {
			t.Stakeholders = strings.Split(stakeholders.String, ",")
		} else {
			t.Stakeholders = []string{}
		}
		if tags.Valid && tags.String != "" {
			t.Tags = strings.Split(tags.String, ",")
		} else {
			t.Tags = []string{}
		}
		tickets = append(tickets, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tickets: %w", err)
	}

	return tickets, nil
}

func (r *Repository) CreateTicket(ctx context.Context, params CreateTicketParams) (int64, error) {
	if err := validateStatus(params.Status); err != nil {
		return 0, err
	}
	if err := validateTags(params.Tags); err != nil {
		return 0, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("failed to rollback: %v\n", err)
		}
	}()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO tickets (title, description, status, assignee, due) VALUES (?, ?, ?, ?, ?)
	`, params.Title, params.Description, params.Status, params.Assignee, params.Due)
	if err != nil {
		return 0, fmt.Errorf("failed to insert ticket: %w", err)
	}
	ticketID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	if len(params.SubAssignees) > 0 {
		placeholders := make([]string, 0, len(params.SubAssignees))
		args := make([]interface{}, 0, len(params.SubAssignees)*2)
		seen := make(map[string]struct{})
		for _, subAssignee := range params.SubAssignees {
			if _, ok := seen[subAssignee]; ok {
				continue
			}
			seen[subAssignee] = struct{}{}
			placeholders = append(placeholders, "(?, ?)")
			args = append(args, ticketID, subAssignee)
		}
		query := fmt.Sprintf("INSERT INTO ticket_sub_assignees (ticket_id, sub_assignee) VALUES %s", strings.Join(placeholders, ","))
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, fmt.Errorf("failed to insert sub_assignees: %w", err)
		}
	}

	if len(params.Stakeholders) > 0 {
		placeholders := make([]string, 0, len(params.Stakeholders))
		args := make([]interface{}, 0, len(params.Stakeholders)*2)
		seen := make(map[string]struct{})
		for _, stakeholder := range params.Stakeholders {
			if _, ok := seen[stakeholder]; ok {
				continue
			}
			seen[stakeholder] = struct{}{}
			placeholders = append(placeholders, "(?, ?)")
			args = append(args, ticketID, stakeholder)
		}
		query := fmt.Sprintf("INSERT INTO ticket_stakeholders (ticket_id, stakeholder) VALUES %s", strings.Join(placeholders, ","))
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, fmt.Errorf("failed to insert stakeholders: %w", err)
		}
	}

	if len(params.Tags) > 0 {
		placeholders := make([]string, 0, len(params.Tags))
		args := make([]interface{}, 0, len(params.Tags)*2)
		seen := make(map[string]struct{})
		for _, tag := range params.Tags {
			if _, ok := seen[tag]; ok {
				continue
			}
			seen[tag] = struct{}{}
			placeholders = append(placeholders, "(?, ?)")
			args = append(args, ticketID, tag)
		}
		query := fmt.Sprintf("INSERT INTO ticket_tags (ticket_id, tag) VALUES %s", strings.Join(placeholders, ","))
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, fmt.Errorf("failed to insert tags: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	// botMessage := fmt.Sprintf("## 新しいチケット(ID: %d)が作成されました\nタイトル: %s\n担当者: @%s\n副担当: %v\n関係者: %v\nタグ: %v\n締め切り: %v\n%s", ticketID, params.Title, params.Assignee, params.SubAssignees, params.Stakeholders, params.Tags, params.Due.Time, params.Description.String)
	// if err := r.bot.PostMessage(ctx, os.Getenv("CREATE_TICKET_CHANNEL_ID"), botMessage); err != nil {
	// 	fmt.Printf("failed to send ticket creation notification: %v\n", err)
	// }
	// if err := r.bot.PostDirectMessage(ctx, "2e0c6679-166f-455a-b8b0-35cdfd257256", botMessage); err != nil {
	// 	fmt.Printf("failed to send ticket creation notification: %v\n", err)
	// }

	return ticketID, nil
}

func (r *Repository) GetTicketByID(ctx context.Context, ticketID int64) (*Ticket, error) {
	ticket := new(Ticket)
	if err := r.db.GetContext(ctx, ticket, "SELECT * FROM tickets WHERE id = ? AND deleted_at IS NULL", ticketID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTicketNotFound
		}

		return nil, fmt.Errorf("failed to select ticket: %w", err)
	}
	subAssignees := []string{}
	if err := r.db.SelectContext(ctx, &subAssignees, "SELECT sub_assignee FROM ticket_sub_assignees WHERE ticket_id = ?", ticketID); err != nil {
		return nil, fmt.Errorf("failed to select sub_assignees: %w", err)
	}
	ticket.SubAssignees = subAssignees

	stakeholders := []string{}
	if err := r.db.SelectContext(ctx, &stakeholders, "SELECT stakeholder FROM ticket_stakeholders WHERE ticket_id = ?", ticketID); err != nil {
		return nil, fmt.Errorf("failed to select stakeholders: %w", err)
	}
	ticket.Stakeholders = stakeholders

	tags := []string{}
	if err := r.db.SelectContext(ctx, &tags, "SELECT tag FROM ticket_tags WHERE ticket_id = ?", ticketID); err != nil {
		return nil, fmt.Errorf("failed to select tags: %w", err)
	}
	ticket.Tags = tags

	return ticket, nil
}

func (r *Repository) UpdateTicket(ctx context.Context, ticketID int64, params CreateTicketParams) error {
	if err := validateStatus(params.Status); err != nil {
		return err
	}

	if err := validateTags(params.Tags); err != nil {
		return err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("failed to rollback: %v\n", err)
		}
	}()

	res, err := tx.ExecContext(ctx, `
		UPDATE tickets SET title = ?, description = ?, status = ?, assignee = ?, due = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, params.Title, params.Description, params.Status, params.Assignee, params.Due, ticketID)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTicketNotFound
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ticket_sub_assignees WHERE ticket_id = ?`, ticketID); err != nil {
		return fmt.Errorf("failed to delete sub_assignees: %w", err)
	}
	if len(params.SubAssignees) > 0 {
		placeholders := make([]string, 0, len(params.SubAssignees))
		args := make([]interface{}, 0, len(params.SubAssignees)*2)
		seen := make(map[string]struct{})
		for _, subAssignee := range params.SubAssignees {
			if _, ok := seen[subAssignee]; ok {
				continue
			}
			seen[subAssignee] = struct{}{}
			placeholders = append(placeholders, "(?, ?)")
			args = append(args, ticketID, subAssignee)
		}
		query := fmt.Sprintf("INSERT INTO ticket_sub_assignees (ticket_id, sub_assignee) VALUES %s", strings.Join(placeholders, ","))
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to insert sub_assignees: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ticket_stakeholders WHERE ticket_id = ?`, ticketID); err != nil {
		return fmt.Errorf("failed to delete stakeholders: %w", err)
	}
	if len(params.Stakeholders) > 0 {
		placeholders := make([]string, 0, len(params.Stakeholders))
		args := make([]interface{}, 0, len(params.Stakeholders)*2)
		seen := make(map[string]struct{})
		for _, stakeholder := range params.Stakeholders {
			if _, ok := seen[stakeholder]; ok {
				continue
			}
			seen[stakeholder] = struct{}{}
			placeholders = append(placeholders, "(?, ?)")
			args = append(args, ticketID, stakeholder)
		}
		query := fmt.Sprintf("INSERT INTO ticket_stakeholders (ticket_id, stakeholder) VALUES %s", strings.Join(placeholders, ","))
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to insert stakeholders: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ticket_tags WHERE ticket_id = ?`, ticketID); err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}
	if len(params.Tags) > 0 {
		placeholders := make([]string, 0, len(params.Tags))
		args := make([]interface{}, 0, len(params.Tags)*2)
		seen := make(map[string]struct{})
		for _, tag := range params.Tags {
			if _, ok := seen[tag]; ok {
				continue
			}
			seen[tag] = struct{}{}
			placeholders = append(placeholders, "(?, ?)")
			args = append(args, ticketID, tag)
		}
		query := fmt.Sprintf("INSERT INTO ticket_tags (ticket_id, tag) VALUES %s", strings.Join(placeholders, ","))
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to insert tags: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) DeleteTicket(ctx context.Context, ticketID int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tickets SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?
	`, ticketID)
	if err != nil {
		return fmt.Errorf("failed to delete ticket: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTicketNotFound
	}

	return nil
}

func (r *Repository) GetIncompleteTickets(ctx context.Context) ([]Ticket, error) {
	var tickets []Ticket
	query := `SELECT * FROM tickets WHERE status NOT IN ('completed', 'forgotten')`
	if err := r.db.SelectContext(ctx, &tickets, query); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *Repository) GetTicketsByStatus(ctx context.Context, status string) ([]Ticket, error) {
	var tickets []Ticket
	query := `SELECT * FROM tickets WHERE status = ?`
	if err := r.db.SelectContext(ctx, &tickets, query, status); err != nil {
		return nil, err
	}

	return tickets, nil
}
