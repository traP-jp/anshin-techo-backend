-- +goose Up

CREATE TABLE IF NOT EXISTS users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    traq_id VARCHAR(64) NOT NULL UNIQUE,
    role ENUM('member', 'assistant', 'manager') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tickets (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    assignee VARCHAR(64) NOT NULL,
    due DATE,
    status ENUM(
        'not_planned',
        'not_written',
        'waiting_review',
        'waiting_sent',
        'sent',
        'milestone_scheduled',
        'completed',
        'forgotten'
    ) NOT NULL,
    title TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ticket_sub_assignees (
    ticket_id INT UNSIGNED,
    sub_assignee VARCHAR(64) NOT NULL,
    PRIMARY KEY(ticket_id, sub_assignee),
    CONSTRAINT `1` FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ticket_stakeholders (
    ticket_id INT UNSIGNED,
    stakeholder VARCHAR(64) NOT NULL,
    PRIMARY KEY(ticket_id, stakeholder),
    CONSTRAINT `1` FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ticket_tags (
    ticket_id INT UNSIGNED,
    tag VARCHAR(64) NOT NULL,
    PRIMARY KEY(ticket_id, tag),
    CONSTRAINT `1` FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notes (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    ticket_id INT UNSIGNED NOT NULL,
    type ENUM('outgoing', 'incoming', 'other') NOT NULL,
    status ENUM('draft', 'waiting_review', 'waiting_sent', 'sent', 'canceled') NOT NULL,
    author VARCHAR(64) NOT NULL,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT `1` FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE,
    CONSTRAINT `2` FOREIGN KEY (author) REFERENCES users(traq_id)
);

CREATE TABLE IF NOT EXISTS note_review_assignees (
    note_id INT UNSIGNED,
    assignee VARCHAR(64) NOT NULL,
    PRIMARY KEY(note_id, assignee),
    CONSTRAINT `1` FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
    CONSTRAINT `2` FOREIGN KEY (assignee) REFERENCES users(traq_id)
);

CREATE TABLE IF NOT EXISTS reviews (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    note_id INT UNSIGNED NOT NULL,
    type ENUM('approve', 'cr', 'comment') NOT NULL,
    status ENUM('active', 'stale') NOT NULL,
    weight INT NOT NULL DEFAULT 0,
    author VARCHAR(64) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT `1` FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
    CONSTRAINT `2` FOREIGN KEY (author) REFERENCES users(traq_id)
);
