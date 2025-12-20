-- +goose Up

ALTER TABLE notes
  DROP FOREIGN KEY `2`;

ALTER TABLE note_review_assignees
  DROP FOREIGN KEY `2`;

ALTER TABLE reviews
  DROP FOREIGN KEY `2`;
