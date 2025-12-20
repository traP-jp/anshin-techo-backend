package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/traP-jp/anshin-techo-backend/internal/service/bot"
)

type Repository struct {
	db  *sqlx.DB
	bot bot.BotClient
}

func New(db *sqlx.DB, botClient bot.BotClient) *Repository {
	return &Repository{db: db, bot: botClient}
}
