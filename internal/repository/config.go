package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type ConfigReminderInterval struct {
	OverdueDay   []int `db:"-"`
	NotesentHour int   `db:"notesent_hour"`
}

type Config struct {
	ReminderInterval ConfigReminderInterval
	RevisePrompt     string `db:"revise_prompt"`
}

var ErrConfigNotFound = fmt.Errorf("config not found")

func (r *Repository) GetConfig(ctx context.Context) (*Config, error) {
	var row struct {
		RevisePrompt string `db:"revise_prompt"`
		NotesentHour int    `db:"notesent_hour"`
		OverdueDay   []byte `db:"overdue_day"`
	}

	if err := r.db.GetContext(ctx, &row, `SELECT revise_prompt, notesent_hour, overdue_day FROM configs WHERE id = 1`); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConfigNotFound
		}

		return nil, fmt.Errorf("select config: %w", err)
	}

	overdueDay := []int{}
	if len(row.OverdueDay) > 0 {
		if err := json.Unmarshal(row.OverdueDay, &overdueDay); err != nil {
			return nil, fmt.Errorf("unmarshal overdue_day: %w", err)
		}
	}

	return &Config{
		ReminderInterval: ConfigReminderInterval{
			OverdueDay:   overdueDay,
			NotesentHour: row.NotesentHour,
		},
		RevisePrompt: row.RevisePrompt,
	}, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, cfg Config) error {
	if cfg.ReminderInterval.OverdueDay == nil {
		cfg.ReminderInterval.OverdueDay = []int{}
	}

	overdueJSON, err := json.Marshal(cfg.ReminderInterval.OverdueDay)
	if err != nil {
		return fmt.Errorf("marshal overdue_day: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `
        INSERT INTO configs (id, revise_prompt, notesent_hour, overdue_day)
        VALUES (1, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
            revise_prompt = VALUES(revise_prompt),
            notesent_hour = VALUES(notesent_hour),
            overdue_day = VALUES(overdue_day)
    `, cfg.RevisePrompt, cfg.ReminderInterval.NotesentHour, overdueJSON); err != nil {
		return fmt.Errorf("upsert config: %w", err)
	}

	return nil
}
