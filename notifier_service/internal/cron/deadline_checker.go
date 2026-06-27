package cron

import (
	"context"
	"time"

	"notifier_service/internal/repository"
	"notifier_service/internal/service"
	"notifier_service/pkg/log"
	"notifier_service/pkg/metrics"
)

type DeadlineChecker struct {
	repo     repository.DeadlineRepository
	notifier service.NotifierServiceInterface
	interval time.Duration
	stop     chan struct{}
}

func NewDeadlineChecker(
	repo repository.DeadlineRepository,
	notifier service.NotifierServiceInterface,
	interval time.Duration,
) *DeadlineChecker {
	return &DeadlineChecker{
		repo:     repo,
		notifier: notifier,
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
	metrics.DeadlineChecks.Inc()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deadlines, err := d.repo.GetPending(ctx)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get pending deadlines")
		return
	}

	if len(deadlines) == 0 {
		log.Logger.Debug().Msg("No pending deadlines")
		return
	}

	log.Logger.Info().Int("count", len(deadlines)).Msg("Found pending deadlines")

	for _, deadline := range deadlines {
		err := d.notifier.HandleDeadlineApproaching(
			ctx,
			deadline.UserID,
			deadline.TaskID,
			deadline.Title,
			deadline.Deadline.Format(time.RFC3339),
		)
		if err != nil {
			log.Logger.Error().Err(err).Str("task_id", deadline.TaskID.String()).Msg("Failed to send deadline notification")
			continue
		}

		if err := d.repo.MarkNotified(ctx, deadline.TaskID); err != nil {
			log.Logger.Error().Err(err).Str("task_id", deadline.TaskID.String()).Msg("Failed to mark deadline notified")
		} else {
			metrics.DeadlineNotificationsSent.Inc()
		}
	}
}
