package postgres

import (
	"context"
	//"log"
	"time"

	"auth_service/internal/config"
	"auth_service/pkg/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectPostgres() {
	config, err := pgxpool.ParseConfig(config.PostgresDSN)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Config parse error:")
	}
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Connection error:")
	}

	if err := pool.Ping(ctx); err != nil {
		log.Logger.Fatal().Err(err).Msg("Postgres ping error:")
	}

	DB = pool
	log.Logger.Info().Msg("Connected to PG")
}

func ClosePostgres() {
	if DB != nil {
		DB.Close()
		log.Logger.Info().Msg("PG connection Closed")
	}
}

func GetDB() *pgxpool.Pool {
	return DB
}
