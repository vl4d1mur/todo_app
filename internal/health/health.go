package health

import (
	"context"
	"net/http"
	"time"

	"todo/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Checker struct {
	postgres *pgxpool.Pool
	redis    *redis.Client
	mongo    *mongo.Database
	nats     *nats.Conn
}

func NewChecker(
	postgres *pgxpool.Pool,
	redis *redis.Client,
	mongo *mongo.Database,
	nats *nats.Conn,
) *Checker {
	return &Checker{
		postgres: postgres,
		redis:    redis,
		mongo:    mongo,
		nats:     nats,
	}
}

func (c *Checker) Liveness(w http.ResponseWriter, r *http.Request) {
	middleware.RespondWithJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (c *Checker) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	status := map[string]string{}
	healthy := true

	if err := c.postgres.Ping(ctx); err != nil {
		status["postgres"] = "unavailable"
		healthy = false
	} else {
		status["postgres"] = "ok"
	}

	if err := c.redis.Ping(ctx).Err(); err != nil {
		status["redis"] = "unavailable"
		healthy = false
	} else {
		status["redis"] = "ok"
	}

	if err := c.mongo.Client().Ping(ctx, nil); err != nil {
		status["mongo"] = "unavailable"
		healthy = false
	} else {
		status["mongo"] = "ok"
	}

	if !c.nats.IsConnected() {
		status["nats"] = "unavailable"
		healthy = false
	} else {
		status["nats"] = "ok"
	}

	if !healthy {
		middleware.RespondWithJSON(w, http.StatusServiceUnavailable, status)
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, status)
}