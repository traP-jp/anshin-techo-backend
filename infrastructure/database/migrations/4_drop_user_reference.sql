-- +goose Up

ALTER TABLE notes
  DROP FOREIGN KEY `fk_notes_author`;

ALTER TABLE note_review_assignees
  DROP FOREIGN KEY `fk_note_review_assignees_assignee`;

ALTER TABLE reviews
  DROP FOREIGN KEY `fk_reviews_author`;