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
	if !(params.Status == "open" || params.Status == "in_progress" || params.Status == "closed") {
		return fmt.Errorf("invalid status: %s", params.Status)
	}
	users, err := r.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}
	assignee_found := false
	for _, user := range users {
		if user.Name == params.Assignee {
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
			if user.Name == subAssignee {
				sub_assignee_found = true
				break
			}
		}
		if !sub_assignee_found {
			return fmt.Errorf("sub-assignee not found: %s", subAssignee)
		}
	}
	
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO tickets (title, description, status, assignee, due)
		VALUES (?, ?, ?, ?, ?)
	`, params.Title, params.Description, params.Status, params.Assignee, params.Due)
	if err != nil {
		return fmt.Errorf("failed to insert ticket: %w", err)
	}
	
	ticket := new(Ticket)
	if err := r.db.GetContext(ctx, ticket, "SELECT * FROM tickets WHERE id = LAST_INSERT_ID()"); err != nil {
		return fmt.Errorf("failed to retrieve inserted ticket: %w", err)
	}
	id := ticket.ID

	for _, subAssignee := range params.SubAssignees {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO ticket_sub_assignees (ticket_id, sub_assignee)
			VALUES (?, ?)
		`, id, subAssignee)
		if err != nil {
			return fmt.Errorf("failed to insert sub-assignee: %w", err)
		}
	}

	for _, stakeholder := range params.Stakeholders {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO ticket_stakeholders (ticket_id, stakeholder)
			VALUES (?, ?)
		`, id, stakeholder)
		if err != nil {
			return fmt.Errorf("failed to insert stakeholder: %w", err)
		}
	}

	for _, tag := range params.Tags {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO ticket_tags (ticket_id, tag)
			VALUES (?, ?)
		`, id, tag)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}
	}

	return nil
}