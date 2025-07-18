package duel

import "time"

// Scheduler defines the interface for scheduling auto-roll tasks
type Scheduler interface {
	ScheduleAutoRoll(duelID ID, deadline time.Time) error
	CancelAutoRoll(duelID ID) error
}
