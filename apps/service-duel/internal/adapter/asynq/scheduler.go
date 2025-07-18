package asynq

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	TypeAutoRoll   = "duel:auto-roll"
	queueName      = "auto_roll" // auto roll queue
	uniqueTTLExtra = time.Minute // extra time to TTL uniqueness
)

type Scheduler struct {
	client *asynq.Client
	insp   *asynq.Inspector
	log    *logger.Logger
}

// NewScheduler takes asynq.Client and already created Redis client from go-redis.
// Inspector is created from this Redis client.
func NewScheduler(client *asynq.Client, redisClient *redis.Client, log *logger.Logger) *Scheduler {
	insp := asynq.NewInspectorFromRedisClient(redisClient) // :contentReference[oaicite:0]{index=0}
	return &Scheduler{client: client, insp: insp, log: log}
}

type PayloadAutoRoll struct {
	DuelID string `json:"duel_id"`
}

// ScheduleAutoRoll schedules auto rolls exactly at the deadline.
// We use ProcessAt and Unique(ttl) for deduplication of the task until execution.
// Unique(ttl) in v0.25.1 takes only one argument — the lifetime of the uniqueness key :contentReference[oaicite:1]{index=1}.
func (s *Scheduler) ScheduleAutoRoll(duelID dueldomain.ID, deadline time.Time) error {
	payload := PayloadAutoRoll{DuelID: duelID.String()}
	b, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(
			"marshal PayloadAutoRoll failed",
			zap.Error(err),
			zap.String("duel_id", duelID.String()),
		)
		return err
	}

	task := asynq.NewTask(TypeAutoRoll, b)
	opts := []asynq.Option{
		asynq.ProcessAt(deadline),
		asynq.Queue(queueName),
		asynq.Unique(time.Until(deadline) + uniqueTTLExtra),
	}

	info, err := s.client.Enqueue(task, opts...)
	if err != nil {
		s.log.Error("enqueue auto-roll failed",
			zap.String("duel_id", duelID.String()),
			zap.Time("deadline", deadline),
			zap.Error(err),
		)
		return err
	}

	s.log.Info("scheduled auto-roll",
		zap.String("duel_id", duelID.String()),
		zap.Time("deadline", deadline),
		zap.String("task_id", info.ID),
	)
	return nil
}

// CancelAutoRoll finds the auto roll task in the scheduled tasks whose payload.DuelID matches,
// and deletes it using the DeleteTask method.
func (s *Scheduler) CancelAutoRoll(duelID dueldomain.ID) error {
	page := 1
	for {
		//nolint:mnd // 100 tasks per page is reasonable
		entries, err := s.insp.ListScheduledTasks(queueName, asynq.PageSize(100), asynq.Page(page))
		if err != nil {
			s.log.Error("ListScheduledTasks failed", zap.Error(err))
			return err
		}
		if len(entries) == 0 {
			break // pages are over
		}
		for _, e := range entries {
			if e.Type != TypeAutoRoll {
				continue
			}
			var p PayloadAutoRoll
			if err = json.Unmarshal(e.Payload, &p); err != nil {
				continue
			}
			if p.DuelID == duelID.String() {
				if err = s.insp.DeleteTask(queueName, e.ID); err != nil {
					s.log.Error("DeleteTask failed", zap.String("task_id", e.ID), zap.Error(err))
					return err
				}
				s.log.Info("cancelled auto-roll task",
					zap.String("duel_id", duelID.String()),
					zap.String("task_id", e.ID),
				)
				return nil
			}
		}
		page++
	}
	s.log.Warn("no scheduled auto-roll found to cancel", zap.String("duel_id", duelID.String()))
	return nil
}
