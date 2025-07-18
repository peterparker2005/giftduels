package duel

import (
	"github.com/peterparker2005/giftduels/packages/events"
)

const (
	SQLOutboxTopic = "duel_sql_outbox"
)

const (
	TopicDuelCreated   events.Topic = "duel.created"
	TopicDuelCancelled events.Topic = "duel.cancelled"
	TopicDuelCompleted events.Topic = "duel.completed"
	TopicDuelJoined    events.Topic = "duel.joined"

	TopicDuelCreateFailed events.Topic = "duel.create.failed"
)
