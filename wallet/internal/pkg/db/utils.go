package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

var postgresConenction *pgxpool.Pool

func connectPostgres(ctx context.Context, databaseURI string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	db, err := pgxpool.Connect(ctx, databaseURI)
	if err != nil {
		log.Panic(err)
	}

	if err := db.Ping(ctx); err != nil {
		panic(err)
	}

	postgresConenction = db

	log.Println("Connection to DB success")
}

func GetPostgresConnection(ctx context.Context, databaseURI string) *pgxpool.Pool {
	if postgresConenction == nil {
		connectPostgres(ctx, databaseURI)
	}

	return postgresConenction
}
