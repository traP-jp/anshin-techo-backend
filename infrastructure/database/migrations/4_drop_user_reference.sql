-- +goose Up

ALTER TABLE notes
  DROP FOREIGN KEY notes_ibfk_2;

ALTER TABLE note_review_assignees
  DROP FOREIGN KEY note_review_assignees_ibfk_2;

ALTER TABLE reviews
  DROP FOREIGN KEY reviews_ibfk_2;

-- +goose Down

ALTER TABLE notes
  ADD CONSTRAINT fk_notes_author FOREIGN KEY (author) REFERENCES users(traq_id);

ALTER TABLE note_review_assignees
  ADD CONSTRAINT fk_note_review_assignees_assignee FOREIGN KEY (assignee) REFERENCES users(traq_id);

ALTER TABLE reviews
  ADD CONSTRAINT fk_reviews_author FOREIGN KEY (author) REFERENCES users(traq_id);
