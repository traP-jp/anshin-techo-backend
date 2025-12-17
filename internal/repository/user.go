package repository

import (
	"context"
	"fmt"
)

type (
	User struct {
		TraqID string `db:"traq_id"`
		Role   string `db:"role"`
	}
)

func (r *Repository) GetUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
	if err := r.db.SelectContext(ctx, &users, "SELECT traq_id, role FROM users"); err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}

	return users, nil
}

func (r *Repository) UpdateUsers(ctx context.Context, users []*User) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM users"); err != nil {
		return fmt.Errorf("delete all users: %w", err)
	}

	if len(users) == 0 {
		return tx.Commit()
	}

	query := "INSERT INTO users (traq_id, role) VALUES (:traq_id, :role)"
	if _, err := tx.NamedExecContext(ctx, query, users); err != nil {
		return fmt.Errorf("bulk insert users: %w", err)
	}

	return tx.Commit()
}
