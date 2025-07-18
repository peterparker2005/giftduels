package asynq

import (
	"time"

	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
)

// SchedulerProvider wraps the concrete Scheduler to implement the domain interface
type SchedulerProvider struct {
	scheduler *Scheduler
}

// NewSchedulerProvider creates a new provider that implements the domain Scheduler interface
func NewSchedulerProvider(scheduler *Scheduler) dueldomain.Scheduler {
	return &SchedulerProvider{scheduler: scheduler}
}

// ScheduleAutoRoll implements dueldomain.Scheduler interface
func (p *SchedulerProvider) ScheduleAutoRoll(duelID dueldomain.ID, deadline time.Time) error {
	return p.scheduler.ScheduleAutoRoll(duelID, deadline)
}

// CancelAutoRoll implements dueldomain.Scheduler interface
func (p *SchedulerProvider) CancelAutoRoll(duelID dueldomain.ID) error {
	return p.scheduler.CancelAutoRoll(duelID)
}
