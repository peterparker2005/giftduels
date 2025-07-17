package stream

import (
	"context"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

// NewSession создаёт новую «сессию» стриминга.
// closeFn можно использовать для общей очистки, если нужно.
func NewSession(
	sub message.Subscriber,
	srv Server,
	mapper MessageMapper,
	logger *logger.Logger,
	closeFn func(),
) *Session {
	return &Session{
		subscriber: sub,
		mapper:     mapper,
		logger:     logger,
		closeAll:   closeFn,
		mu:         sync.Mutex{},
		subs:       make(map[string]context.CancelFunc),
		Send:       srv.Send,
		Ctx:        srv.Context(),
	}
}

// Run запускает приём команд и доставку событий.
// Возвращает, когда клиент закрыл стрим или произошла фатальная ошибка.
func (s *Session) Run() error {
	// subscribe to your "all duels" topic
	duelsCh, err := s.subscriber.Subscribe(s.Ctx, "duels")
	if err != nil {
		return err
	}
	defer s.Cleanup(s.Subs()) // or however you clean up

	s.forward("duels", duelsCh)

	return nil
}

func (s *Session) forward(
	topic string,
	ch <-chan *message.Message,
) {
	for msg := range ch {
		resp, err := s.mapper(topic, msg)
		if err != nil {
			s.logger.Error("mapper failed", zap.Error(err))
			msg.Nack()
			continue
		}
		// если mapper вернул (nil, nil) — пропускаем
		if resp == nil {
			msg.Ack()
			continue
		}
		if err = s.Send(resp); err != nil {
			s.logger.Error("send failed", zap.Error(err))
			msg.Nack()
			return
		}
		msg.Ack()
	}
}

func (s *Session) Subs() map[string]context.CancelFunc {
	s.mu.Lock()
	defer s.mu.Unlock()
	subCopy := make(map[string]context.CancelFunc, len(s.subs))
	for k, v := range s.subs {
		subCopy[k] = v
	}
	return subCopy
}

func (s *Session) Cleanup(cancels map[string]context.CancelFunc) {
	for _, cancel := range cancels {
		cancel()
	}
	if s.closeAll != nil {
		s.closeAll()
	}
}

func (s *Session) SubscribeDuels(duelIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range duelIDs {
		if _, ok := s.subs[id]; ok {
			continue // уже подписаны
		}
		ctx, cancel := context.WithCancel(s.Ctx)
		ch, err := s.subscriber.Subscribe(ctx, "duel:"+id)
		if err != nil {
			cancel()
			s.logger.Error("subscribe failed", zap.String("duel_id", id), zap.Error(err))
			return err
		}
		s.subs[id] = cancel

		go s.forward("duel:"+id, ch)
	}
	return nil
}

// UnsubscribeDuels отменяет подписку
func (s *Session) UnsubscribeDuels(duelIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range duelIDs {
		if cancel, ok := s.subs[id]; ok {
			cancel()
			delete(s.subs, id)
		}
	}
}
