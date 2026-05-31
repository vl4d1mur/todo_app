package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"todo/pkg/jwt"
	"todo/pkg/log"

	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
	contextKeyEmail  contextKey = "email"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			RespondWithError(w, http.StatusUnauthorized, "JWT token is not provided")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			RespondWithError(w, http.StatusUnauthorized, "Invalid token format")
			return
		}

		tokenString := parts[1]

		claims, err := jwt.ParseAccess(tokenString)
		if err != nil {
			switch err {
			case jwt.ErrExpiredToken:
				RespondWithError(w, http.StatusUnauthorized, "Token expired")
			default:
				RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			}
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		log.Logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Dur("latency", time.Since(start)).
			Msg("Request")
	})
}

func GetUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(contextKeyUserID).(uuid.UUID)
	return userID, ok
}

func GetEmailFromContext(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(contextKeyEmail).(string)
	return email, ok
}
