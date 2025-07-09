package gift

import "github.com/peterparker2005/giftduels/packages/events"

const SqlOutboxTopic = "gift_sql_outbox"

const (
	TopicGiftWithdrawRequested events.Topic = "gift.withdraw.requested"
	TopicGiftWithdrawFailed    events.Topic = "gift.withdraw.failed"
)
