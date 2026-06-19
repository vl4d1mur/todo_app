package consumer

import (
	"context"
	"encoding/json"
	"time"

	"notifier_service/internal/config"
	"notifier_service/internal/models"
	"notifier_service/internal/service"
	"notifier_service/pkg/log"

	"github.com/nats-io/nats.go"
)

type Consumer struct {
	conn     *nats.Conn
	notifier service.NotifierServiceInterface
	sub      *nats.Subscription
}

func NewConsumer(notifier service.NotifierServiceInterface) (*Consumer, error) {
	conn, err := nats.Connect(config.NatsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second))
	if err != nil {
		return nil, err
	}

	log.Logger.Info().Str("url", config.NatsURL).Msg("Connected to NATS")

	return &Consumer{
		conn:     conn,
		notifier: notifier,
	}, nil
}

func (c *Consumer) handleMessage(msg *nats.Msg) {
	var event models.TaskEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to unmarshal event")
		return
	}

	log.Logger.Info().Str("event_type", event.EventType).
		Str("task_id", event.TaskID.String()).
		Msg("Event received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch event.EventType {
	case "task.status_changed":
		c.handleStatusChanged(ctx, event)
	case "task.deadline_approaching":
		c.handleDeadlineApproaching(ctx, event)
	default:
		log.Logger.Debug().Str("event_type", event.EventType).Msg("Event ignored")
	}

}

func (c *Consumer) handleStatusChanged(ctx context.Context, event models.TaskEvent) {
	newStatus, _ := event.Payload["new_status"].(string)
	taskTitle, _ := event.Payload["title"].(string)

	if newStatus == "" || taskTitle == "" {
		log.Logger.Warn().Str("task_id", event.TaskID.String()).
			Msg("Missing fields in status_changed event")
		return
	}
	if err := c.notifier.HandleStatusChanged(ctx, event.UserID, event.TaskID, newStatus, taskTitle); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to handle status_changed")
	}
}

func (c *Consumer) handleDeadlineApproaching(ctx context.Context, event models.TaskEvent) {
	taskTitle, _ := event.Payload["title"].(string)
	deadline, _ := event.Payload["deadline"].(string)

	if taskTitle == "" || deadline == "" {
		log.Logger.Warn().
			Str("task_id", event.TaskID.String()).
			Msg("Missing fields in deadline_approaching event")
		return
	}

	if err := c.notifier.HandleDeadlineApproaching(ctx, event.UserID, event.TaskID, taskTitle, deadline); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to handle deadline_approaching")
	}
}

func (c *Consumer) Conn() *nats.Conn {
	return c.conn
}

func (c *Consumer) Start() error {
	sub, err := c.conn.Subscribe("task-events", c.handleMessage)
	if err != nil {
		return err
	}

	c.sub = sub
	log.Logger.Info().Str("subject", "task-events").Msg("Subscribed to NATS")
	return nil
}

func (c *Consumer) Close() {
	if c.sub != nil {
		_ = c.sub.Unsubscribe()
	}
	if c.conn != nil {
		_ = c.conn.Drain()
		log.Logger.Info().Msg("NATS connection closed")
	}
}
