package db

import (
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func Connect(dsn string) *sqlx.DB {
	conn, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatal("db connect:", err)
	}
	return conn
}
