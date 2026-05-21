package handlers

import (
	"finquest/config"

	"github.com/jmoiron/sqlx"
)

type Handler struct {
	DB  *sqlx.DB
	Cfg *config.Config
}

func New(db *sqlx.DB, cfg *config.Config) *Handler {
	return &Handler{DB: db, Cfg: cfg}
}
