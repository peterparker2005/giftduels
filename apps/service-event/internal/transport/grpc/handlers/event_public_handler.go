package grpchandlers

import (
	"context"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/transport/stream"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	eventv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/event/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type eventPublicHandler struct {
	eventv1.UnimplementedEventPublicServiceServer

	subscriber message.Subscriber
	mapper     stream.MessageMapper
	logger     *logger.Logger

	sessions sync.Map // map[int64]*stream.Session
}

func NewEventPublicHandler(
	subscriber message.Subscriber,
	mapper stream.MessageMapper,
	logger *logger.Logger,
) eventv1.EventPublicServiceServer {
	return &eventPublicHandler{
		subscriber: subscriber,
		mapper:     mapper,
		logger:     logger,
	}
}

func (h *eventPublicHandler) Stream(
	_ *eventv1.StreamRequest, srv eventv1.EventPublicService_StreamServer,
) error {
	userID, err := authctx.TelegramUserID(srv.Context())
	if err != nil {
		h.logger.Warn("user not authenticated", zap.Error(err))
		return err
	}

	// Remove any existing session for this user
	if existingSess, ok := h.sessions.LoadAndDelete(userID); ok {
		if sess, ok := existingSess.(*stream.Session); ok {
			sess.Cleanup(sess.Subs())
		}
	}

	sess := stream.NewSession(h.subscriber, srv, h.mapper, h.logger, nil)
	h.sessions.Store(userID, sess)

	h.logger.Info("stream session created", zap.Int64("user_id", userID))

	defer func() {
		h.sessions.Delete(userID)
		sess.Cleanup(sess.Subs())
		h.logger.Info("stream session cleaned up", zap.Int64("user_id", userID))
	}()

	// Start the session in a goroutine
	sessionErr := make(chan error, 1)
	go func() {
		if err := sess.Run(); err != nil {
			h.logger.Error("session.Run failed", zap.Error(err), zap.Int64("user_id", userID))
			sessionErr <- err
		}
	}()

	//nolint:mnd // 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-srv.Context().Done():
			h.logger.Info("stream context done", zap.Int64("user_id", userID))
			return nil
		case err := <-sessionErr:
			h.logger.Error("session error", zap.Error(err), zap.Int64("user_id", userID))
			return err
		case <-ticker.C:
			if err := srv.Send(&eventv1.StreamResponse{}); err != nil {
				h.logger.Error("failed to send ping", zap.Error(err), zap.Int64("user_id", userID))
				return err
			}
		}
	}
}

func (h *eventPublicHandler) SubscribeDuels(
	ctx context.Context, req *eventv1.SubscribeDuelsRequest,
) (*eventv1.SubscribeDuelsResponse, error) {
	userID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		h.logger.Warn("user not authenticated", zap.Error(err))
		return nil, err
	}
	sessIface, ok := h.sessions.Load(userID)
	if !ok {
		h.logger.Warn("user session not found", zap.Int64("user_id", userID))
		return nil, status.Error(
			codes.FailedPrecondition,
			"stream session not established. Call Stream() first",
		)
	}
	sess, ok := sessIface.(*stream.Session)
	if !ok {
		h.logger.Error("session is not a stream.Session", zap.Int64("user_id", userID))
		return nil, status.Error(codes.Internal, "session is not a stream.Session")
	}

	ids := make([]string, len(req.GetDuelIds()))
	for i, v := range req.GetDuelIds() {
		ids[i] = v.GetValue()
	}

	h.logger.Info(
		"subscribing to duels",
		zap.Int64("user_id", userID),
		zap.Strings("duel_ids", ids),
	)

	if err = sess.SubscribeDuels(ids); err != nil {
		h.logger.Error(
			"subscribe error",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.Strings("duel_ids", ids),
		)
		return nil, status.Errorf(codes.Internal, "subscribe error: %v", err)
	}

	h.logger.Info(
		"successfully subscribed to duels",
		zap.Int64("user_id", userID),
		zap.Strings("duel_ids", ids),
	)
	return &eventv1.SubscribeDuelsResponse{}, nil
}

func (h *eventPublicHandler) UnsubscribeDuels(
	ctx context.Context, req *eventv1.UnsubscribeDuelsRequest,
) (*eventv1.UnsubscribeDuelsResponse, error) {
	userID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		h.logger.Warn("user not authenticated")
		return nil, err
	}
	sessIface, ok := h.sessions.Load(userID)
	if !ok {
		h.logger.Warn("user session not found", zap.Int64("user_id", userID))
		return nil, status.Error(
			codes.FailedPrecondition,
			"stream session not established. Call Stream() first",
		)
	}
	sess, ok := sessIface.(*stream.Session)
	if !ok {
		h.logger.Error("session is not a stream.Session", zap.Int64("user_id", userID))
		return nil, status.Error(codes.Internal, "session is not a stream.Session")
	}

	ids := make([]string, len(req.GetDuelIds()))
	for i, v := range req.GetDuelIds() {
		ids[i] = v.GetValue()
	}

	h.logger.Info(
		"unsubscribing from duels",
		zap.Int64("user_id", userID),
		zap.Strings("duel_ids", ids),
	)

	sess.UnsubscribeDuels(ids)

	h.logger.Info(
		"successfully unsubscribed from duels",
		zap.Int64("user_id", userID),
		zap.Strings("duel_ids", ids),
	)
	return &eventv1.UnsubscribeDuelsResponse{}, nil
}
