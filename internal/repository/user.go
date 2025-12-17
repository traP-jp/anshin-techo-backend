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

func (r *Repository) UpdateUsers(ctx context.Context, users []*User) (err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, "DELETE FROM users")
	if err != nil {
		return fmt.Errorf("delete all users: %w", err)
	}

	if len(users) == 0 {
		err = tx.Commit()

		return err
	}

	query := "INSERT INTO users (traq_id, role) VALUES (:traq_id, :role)"
	_, err = tx.NamedExecContext(ctx, query, users)
	if err != nil {
		return fmt.Errorf("bulk insert users: %w", err)
	}

	err = tx.Commit()

	return err
}
