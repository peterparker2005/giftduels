package gift

import "github.com/peterparker2005/giftduels/packages/events"

const SQLOutboxTopic = "gift_sql_outbox"

const (
	TopicGiftWithdrawRequested events.Topic = "gift.withdraw.requested"
	TopicGiftDeposited         events.Topic = "gift.deposited"
)
