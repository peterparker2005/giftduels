package telegram

import "github.com/peterparker2005/giftduels/packages/events"

const (
	TopicTelegramGiftReceived             events.Topic = "gift.received"
	TopicTelegramGiftWithdrawn            events.Topic = "gift.withdrawn"
	TopicTelegramGiftWithdrawFailed       events.Topic = "gift.withdraw.failed"
	TopicTelegramGiftWithdrawUserNotFound events.Topic = "gift.withdraw.user-not-found"
)
