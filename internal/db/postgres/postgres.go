package postgres

import (
	"context"
	"log"
	"time"

	"todo/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectPostgres() {
	//uri := "postgres://postgres:root@localhost:5432/todo_app"
	config, err := pgxpool.ParseConfig(config.PostgresDSN)
	if err != nil {
		log.Fatal("Config parse error:", err)
	}
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("Connection error:", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Ping error:", err)
	}

	DB = pool
	log.Println("Connected to PG")
}

func ClosePostgres() {
	if DB != nil {
		DB.Close()
		log.Println("PG connection Closed")
	}
}

func GetDB() *pgxpool.Pool {
	return DB
}

func ExecRowAffected(ctx context.Context, query string, args ...interface{}) (int64, error) {
    ct, err := DB.Exec(ctx, query, args...)
    if err != nil {
        return 0, err
    }
    return ct.RowsAffected(), nil
}
