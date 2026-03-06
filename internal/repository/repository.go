package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db  *sqlx.DB
	//bot bot.Client
}

// func New(db *sqlx.DB, botClient bot.Client) *Repository {
// 	return &Repository{db: db, bot: botClient}
// }
func New(db *sqlx.DB, ) *Repository {
 	return &Repository{db: db}
 }
