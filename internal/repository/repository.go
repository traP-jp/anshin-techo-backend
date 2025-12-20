package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/traP-jp/anshin-techo-backend/internal/service/bot"
)

type Repository struct {
	db  *sqlx.DB
	bot bot.Client
}

func New(db *sqlx.DB, botClient bot.Client) *Repository {
	return &Repository{db: db, bot: botClient}
}
