package stream

import (
	"context"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	eventv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/event/v1"
)

// Server абстрагирует grpc.ServerStream.
type Server interface {
	Send(*eventv1.StreamResponse) error
	Context() context.Context
}

// Session хранит все подписки для одного стрима и маршрутизирует команды.
type Session struct {
	subscriber message.Subscriber
	mapper     MessageMapper
	logger     *logger.Logger
	closeAll   func()

	mu   sync.Mutex
	subs map[string]context.CancelFunc

	Recv func() (*eventv1.StreamRequest, error)
	Send func(*eventv1.StreamResponse) error
	Ctx  context.Context
}

type MessageMapper func(topic string, msg *message.Message) (*eventv1.StreamResponse, error)
