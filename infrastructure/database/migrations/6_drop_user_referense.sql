-- +goose Up
ALTER TABLE notes DROP FOREIGN KEY `2`;
ALTER TABLE note_review_assignees DROP FOREIGN KEY `2`;
ALTER TABLE reviews DROP FOREIGN KEY `2`;

-- +goose Down
ALTER TABLE notes ADD CONSTRAINT `2` FOREIGN KEY (author) REFERENCES users(traq_id);
ALTER TABLE note_review_assignees ADD CONSTRAINT `2` FOREIGN KEY (assignee) REFERENCES users(traq_id);
ALTER TABLE reviews ADD CONSTRAINT `2` FOREIGN KEY (author) REFERENCES users(traq_id);