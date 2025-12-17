package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type (
	Ticket struct {
		ID          int64        `db:"id"`
		Title       string       `db:"title"`
		Status      string       `db:"status"`
		Assignee    string       `db:"assignee"`
		Due         sql.NullTime `db:"due"`
		Description string       `db:"description"`
		CreatedAt   time.Time    `db:"created_at"`
		UpdatedAt   time.Time    `db:"updated_at"`
		DeletedAt   sql.NullTime `db:"deleted_at"`
	}

	CreateTicketParams struct {
		Title        string
		Description  string
		Status       string
		Assignee     string
		SubAssignees []string
		Stakeholders []string
		Due          sql.NullTime
		Tags         []string
	}
)

func (r *Repository) GetTickets(ctx context.Context) ([]*Ticket, error) {
	tickets := []*Ticket{}
	if err := r.db.SelectContext(ctx, &tickets, "SELECT * FROM tickets"); err != nil {
		return nil, fmt.Errorf("failed to select tickets: %w", err)
	}

	return tickets, nil
}

func (r *Repository) CreateTicket(ctx context.Context, params CreateTicketParams) error {
	if !(params.Status == "not_planned" || params.Status == "not_written" || params.Status == "waiting_review" || params.Status == "waiting_sent" || params.Status == "sent" || params.Status == "milestone_scheduled" || params.Status == "completed" || params.Status == "forgotten") {
		return fmt.Errorf("invalid status: %s", params.Status)
	}

	unique_user_ids := make(map[string]struct{})
	unique_user_ids[params.Assignee] = struct{}{}
	for _, subAssignee := range params.SubAssignees {
		unique_user_ids[subAssignee] = struct{}{}
	}
	for _, stakeholder := range params.Stakeholders {
		unique_user_ids[stakeholder] = struct{}{}
	}

	users, err := r.GetUsers(ctx) // TODO: 全ユーザーを取得するのは効率が悪いので改善する
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}
	assignee_found := false
	for _, user := range users {
		if user.TraqID == params.Assignee {
			assignee_found = true
			break
		}
	}
	if !assignee_found {
		return fmt.Errorf("assignee not found: %s", params.Assignee)
	}
	for _, subAssignee := range params.SubAssignees {
		sub_assignee_found := false
		for _, user := range users {
			if user.TraqID == subAssignee {
				sub_assignee_found = true
				break
			}
		}
		if !sub_assignee_found {
			return fmt.Errorf("sub-assignee not found: %s", subAssignee)
		}
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO tickets (title, description, status, assignee, due) VALUES (?, ?, ?, ?, ?)
	`, params.Title, params.Description, params.Status, params.Assignee, params.Due)
	if err != nil {
		return fmt.Errorf("failed to insert ticket: %w", err)
	}
	ticketID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	for _, subAssignee := range params.SubAssignees {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO ticket_sub_assignees (ticket_id, sub_assignee) VALUES (?, ?)
		`, ticketID, subAssignee)
		if err != nil {
			return fmt.Errorf("failed to insert sub-assignee: %w", err)
		}
	}

	for _, stakeholder := range params.Stakeholders {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO ticket_stakeholders (ticket_id, stakeholder) VALUES (?, ?)
		`, ticketID, stakeholder)
		if err != nil {
			return fmt.Errorf("failed to insert stakeholder: %w", err)
		}
	}

	for _, tag := range params.Tags {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO ticket_tags (ticket_id, tag) VALUES (?, ?)
		`, ticketID, tag)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) GetTicketByID(ctx context.Context, ticketID int64) (*Ticket, error) {
	ticket := new(Ticket)
	if err := r.db.GetContext(ctx, ticket, "SELECT * FROM tickets WHERE id = ?", ticketID); err != nil {
		return nil, fmt.Errorf("failed to select ticket: %w", err)
	}
	return ticket, nil
}

func (r *Repository) UpdateTicket(ctx context.Context, ticketID int64, params CreateTicketParams) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE tickets SET title = ?, description = ?, status = ?, assignee = ?, due = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, params.Title, params.Description, params.Status, params.Assignee, params.Due, ticketID)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ticket_sub_assignees WHERE ticket_id = ?`, ticketID); err != nil {
		return fmt.Errorf("failed to delete sub_assignees: %w", err)
	}
	for _, subAssignee := range params.SubAssignees {
		_, err = tx.ExecContext(ctx, `INSERT INTO ticket_sub_assignees (ticket_id, sub_assignee) VALUES (?, ?)`, ticketID, subAssignee)
		if err != nil {
			return fmt.Errorf("failed to insert sub_assignee: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ticket_stakeholders WHERE ticket_id = ?`, ticketID); err != nil {
		return fmt.Errorf("failed to delete stakeholders: %w", err)
	}
	for _, stakeholder := range params.Stakeholders {
		_, err = tx.ExecContext(ctx, `INSERT INTO ticket_stakeholders (ticket_id, stakeholder) VALUES (?, ?)`, ticketID, stakeholder)
		if err != nil {
			return fmt.Errorf("failed to insert stakeholder: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM ticket_tags WHERE ticket_id = ?`, ticketID); err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}
	for _, tag := range params.Tags {
		_, err = tx.ExecContext(ctx, `INSERT INTO ticket_tags (ticket_id, tag) VALUES (?, ?)`, ticketID, tag)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) DeleteTicket(ctx context.Context, ticketID int64) error {
	if err := r.db.QueryRowContext(ctx, `
		UPDATE tickets SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?
	`, ticketID).Err(); err != nil {
		return fmt.Errorf("failed to delete ticket: %w", err)
	}
	return nil
}