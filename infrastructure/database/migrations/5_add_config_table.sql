-- +goose Up

CREATE TABLE IF NOT EXISTS configs (
    id INT UNSIGNED NOT NULL PRIMARY KEY,
    revise_prompt TEXT NOT NULL,
    notesent_hour INT NOT NULL,
    overdue_day JSON NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO configs (id, revise_prompt, notesent_hour, overdue_day)
VALUES (1, '', 0, JSON_ARRAY())
ON DUPLICATE KEY UPDATE id = VALUES(id);
