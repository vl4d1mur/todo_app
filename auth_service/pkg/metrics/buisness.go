package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	UsersRegistered = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_users_registered_total",
			Help: "Total registered users",
		},
	)

	UsersLoggedIn = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_users_login_total",
			Help: "Total successful logins",
		},
	)

	LoginFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_failures_total",
			Help: "Total login failures by reason",
		},
		[]string{"reason"},
	)

	JWTRefresh = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_jwt_refresh_total",
			Help: "Total refresh token operations",
		},
	)

	Logouts = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_logouts_total",
			Help: "Total logout operations",
		},
	)

	TelegramCodesGenerated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_telegram_codes_generated_total",
			Help: "Total telegram linking codes generated",
		},
	)

	TelegramActivations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_telegram_activations_total",
			Help: "Total telegram bot activations",
		},
	)
)
