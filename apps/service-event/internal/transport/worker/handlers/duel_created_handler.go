package handlers

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type DuelCreatedHandler struct {
	publisher message.Publisher `name:"redis_publisher"` // теперь RedisStreams Publisher
	logger    *logger.Logger
}

func NewDuelCreatedHandler(p message.Publisher, l *logger.Logger) *DuelCreatedHandler {
	return &DuelCreatedHandler{publisher: p, logger: l}
}

func (h *DuelCreatedHandler) Handle(msg *message.Message) error {
	var evt duelv1.DuelCreatedEvent
	if err := proto.Unmarshal(msg.Payload, &evt); err != nil {
		h.logger.Error("failed to unmarshal duel created event", zap.Error(err))
		msg.Nack()
		return err
	}
	redisTopic := "duel.created"
	pubMsg := message.NewMessage(uuid.New().String(), msg.Payload)
	if err := h.publisher.Publish(redisTopic, pubMsg); err != nil {
		h.logger.Error("failed to publish duel created event", zap.Error(err))
		msg.Nack()
		return err
	}
	msg.Ack()
	return nil
}
