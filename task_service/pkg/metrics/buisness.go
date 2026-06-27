package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TasksCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_created_total",
			Help: "Total tasks created",
		},
	)

	TasksUpdated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_updated_total",
			Help: "Total tasks updated",
		},
	)

	TasksDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_deleted_total",
			Help: "Total tasks deleted",
		},
	)

	TaskStatusChanged = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tasks_status_changed_total",
			Help: "Total task status changes",
		},
		[]string{"new_status"},
	)

	NotesCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_notes_created_total",
			Help: "Total notes created",
		},
	)

	NotesDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_notes_deleted_total",
			Help: "Total notes deleted",
		},
	)

	EventsPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tasks_events_published_total",
			Help: "Total events published to NATS",
		},
		[]string{"event_type"},
	)

	CacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_cache_hits_total",
			Help: "Total Redis cache hits",
		},
	)

	CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_cache_misses_total",
			Help: "Total Redis cache misses",
		},
	)
)
