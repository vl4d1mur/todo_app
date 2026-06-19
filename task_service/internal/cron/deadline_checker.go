package cron

import (
	"context"
	"time"

	"task_service/internal/events"
	"task_service/internal/repository"
	"task_service/pkg/log"
)

type DeadlineChecker struct {
	repo     repository.TaskRepository
	interval time.Duration
	stop     chan struct{}
}

func NewDeadlineChecker(repo repository.TaskRepository, interval time.Duration) *DeadlineChecker {
	return &DeadlineChecker{
		repo:     repo,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (d *DeadlineChecker) Start() {
	go func() {
		log.Logger.Info().Dur("interval", d.interval).Msg("Deadline checker started")
		d.check()

		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				d.check()
			case <-d.stop:
				log.Logger.Info().Msg("Deadline checker stopped")
				return
			}
		}
	}()
}

func (d *DeadlineChecker) Stop() {
	close(d.stop)
}

func (d *DeadlineChecker) check() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tasks, err := d.repo.GetTasksWithUpcomingDeadline(ctx)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get tasks with upcoming deadline")
		return
	}

	if len(tasks) == 0 {
		log.Logger.Debug().Msg("No tasks with upcoming deadline")
		return
	}

	log.Logger.Info().Int("count", len(tasks)).Msg("Found tasks with approaching deadline")

	for _, t := range tasks {
		if t.DeadLine == nil {
			continue
		}

		events.PublishTaskDeadlineApproaching(t.ID, t.UserID, t.Title, t.DeadLine.Format(time.RFC3339))

		if err := d.repo.MarkDeadlineNotified(ctx, t.ID); err != nil {
			log.Logger.Error().Err(err).Str("task_id", t.ID.String()).Msg("Failed to mark deadline notified")
		}
	}
}