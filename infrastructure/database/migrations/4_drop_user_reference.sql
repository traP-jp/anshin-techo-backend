-- +goose Up

ALTER TABLE notes
  DROP FOREIGN KEY `fk_notes_author`;

ALTER TABLE note_review_assignees
  DROP FOREIGN KEY `fk_note_review_assignees_assignee`;

ALTER TABLE reviews
  DROP FOREIGN KEY `fk_reviews_author`;

-- +goose Down
ALTER TABLE notes
  ADD CONSTRAINT `fk_notes_author` FOREIGN KEY (author) REFERENCES users(traq_id);

ALTER TABLE note_review_assignees
  ADD CONSTRAINT `fk_note_review_assignees_assignee` FOREIGN KEY (assignee) REFERENCES users(traq_id);

ALTER TABLE reviews
  ADD CONSTRAINT `fk_reviews_author` FOREIGN KEY (author) REFERENCES users(traq_id);