package health

import (
	"context"
	"net/http"
	"time"

	"auth_service/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Checker struct {
	postgres *pgxpool.Pool
	redis    *redis.Client
}

func NewChecker(
	postgres *pgxpool.Pool,
	redis *redis.Client,
) *Checker {
	return &Checker{
		postgres: postgres,
		redis:    redis,
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

	if !healthy {
		middleware.RespondWithJSON(w, http.StatusServiceUnavailable, status)
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, status)
}
