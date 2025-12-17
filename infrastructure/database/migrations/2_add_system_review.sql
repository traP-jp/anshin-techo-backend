-- +goose Up
ALTER TABLE reviews MODIFY COLUMN type ENUM('approve', 'cr', 'comment', 'system') NOT NULL;