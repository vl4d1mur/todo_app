package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	NotificationsSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifier_notifications_sent_total",
			Help: "Total notifications sent",
		},
		[]string{"channel", "event_type"},
	)

	NotificationsFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifier_notifications_failed_total",
			Help: "Total notifications failed",
		},
		[]string{"channel", "event_type"},
	)

	EventsReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifier_events_received_total",
			Help: "Total events received from NATS",
		},
		[]string{"event_type"},
	)

	DeadlineChecks = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "notifier_deadline_checks_total",
			Help: "Total deadline cron checks",
		},
	)

	DeadlineNotificationsSent = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "notifier_deadline_notifications_sent_total",
			Help: "Total deadline notifications sent via cron",
		},
	)
)
