package gift

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	giftEvents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	txMgr                pg.TxManager
	repo                 gift.GiftRepository
	log                  *logger.Logger
	paymentPrivateClient paymentv1.PaymentPrivateServiceClient
}

func New(repo gift.GiftRepository, txMgr pg.TxManager, log *logger.Logger, clients *clients.Clients) *Service {
	return &Service{repo: repo, txMgr: txMgr, log: log, paymentPrivateClient: clients.Payment.Private}
}

func (s *Service) GetGiftByID(ctx context.Context, id string) (*gift.Gift, error) {
	gift, err := s.repo.GetGiftByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return gift, nil
}

type GetUserGiftsResult struct {
	Gifts      []*gift.Gift
	Total      int32
	TotalValue float64
}

func (s *Service) GetUserGifts(ctx context.Context, telegramUserID int64, pagination *shared.PageRequest) (*GetUserGiftsResult, error) {
	res, err := s.repo.GetUserGifts(ctx, pagination.PageSize(), pagination.Offset(), telegramUserID)
	if err != nil {
		return nil, err
	}

	totalValue := float64(0)
	for _, g := range res.Gifts {
		totalValue += float64(g.Price)
	}

	return &GetUserGiftsResult{
		Gifts:      res.Gifts,
		Total:      int32(res.Total),
		TotalValue: totalValue,
	}, nil
}

func (s *Service) GetUserActiveGifts(ctx context.Context, telegramUserID int64, pagination *shared.PageRequest) (*GetUserGiftsResult, error) {
	res, err := s.repo.GetUserActiveGifts(ctx, pagination.PageSize(), pagination.Offset(), telegramUserID)
	if err != nil {
		return nil, err
	}

	totalValue := float64(0)
	for _, g := range res.Gifts {
		totalValue += float64(g.Price)
	}

	return &GetUserGiftsResult{
		Gifts:      res.Gifts,
		Total:      int32(res.Total),
		TotalValue: totalValue,
	}, nil
}

func (s *Service) StakeGift(ctx context.Context, giftID string) (*gift.Gift, error) {
	gift, err := s.repo.StakeGiftForGame(ctx, giftID)
	if err != nil {
		return nil, err
	}
	return gift, nil
}

func (s *Service) TransferGiftToUser(ctx context.Context, giftID string, telegramUserID int64) (*gift.Gift, error) {
	// First update the gift owner
	gift, err := s.repo.UpdateGiftOwner(ctx, giftID, telegramUserID)
	if err != nil {
		return nil, err
	}

	// Create transfer record
	_, err = s.repo.CreateGiftEvent(ctx, giftID, gift.OwnerTelegramID, telegramUserID)
	if err != nil {
		return nil, err
	}

	return gift, nil
}

func (s *Service) ExecuteWithdraw(ctx context.Context, telegramUserID int64, giftIDs []string) ([]*gift.Gift, error) {
	log := s.log.With(zap.Strings("giftIDs", giftIDs))

	// 2) Начинаем транзакцию, чтобы сохранить пометки и публиковать события
	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	var commitErr error
	defer func() {
		if commitErr != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error("rollback failed", zap.Error(rbErr))
			}
		}
	}()

	repo := s.repo.WithTx(tx)

	// готовим Watermill-паблишер
	sqlPub, err := sql.NewPublisher(
		s.txMgr.Sql(),
		sql.PublisherConfig{SchemaAdapter: sql.DefaultPostgreSQLSchema{}},
		logger.NewWatermill(s.log),
	)
	if err != nil {
		commitErr = err
		log.Error("publisher init failed", zap.Error(err))
		return nil, err
	}
	fwdPub := forwarder.NewPublisher(sqlPub, forwarder.PublisherConfig{
		ForwarderTopic: giftEvents.SqlOutboxTopic,
	})

	var result []*gift.Gift
	for _, giftID := range giftIDs {
		// 3) для каждого подарка списываем комиссию и получаем её величину
		commResp, err := s.paymentPrivateClient.SpendWithdrawalCommission(ctx, &paymentv1.SpendWithdrawalCommissionRequest{
			TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
			TonAmount:      &sharedv1.TonAmount{Value: 1},
		})
		if err != nil {
			commitErr = err
			log.Error("failed to spend withdrawal commission", zap.Error(err))
			return nil, err
		}
		commissionAmt := commResp.CommissionAmount.GetValue()

		// 4) помечаем подарок на вывод и сохраняем commissionAmount в БД
		g, err := repo.MarkGiftForWithdrawal(ctx, giftID)
		if err != nil {
			commitErr = err
			log.Error("failed to mark gift for withdrawal", zap.Error(err))
			return nil, err
		}
		result = append(result, g)

		// 5) публикуем событие с правильным полем CommissionAmount
		ev := &giftv1.GiftWithdrawRequestedEvent{
			GiftId:           &sharedv1.GiftId{Value: g.ID},
			OwnerTelegramId:  &sharedv1.TelegramUserId{Value: g.OwnerTelegramID},
			TelegramGiftId:   &sharedv1.GiftTelegramId{Value: g.TelegramGiftID},
			CollectibleId:    int32(g.CollectibleID),
			UpgradeMessageId: g.UpgradeMessageID,
			Price:            &sharedv1.TonAmount{Value: g.Price},
			CommissionAmount: &sharedv1.TonAmount{Value: commissionAmt},
		}
		payload, err := proto.Marshal(ev)
		if err != nil {
			commitErr = err
			log.Error("marshal event failed", zap.Error(err))
			return nil, err
		}
		msg := message.NewMessage(watermill.NewUUID(), payload)
		if err = fwdPub.Publish(giftEvents.TopicGiftWithdrawRequested.String(), msg); err != nil {
			commitErr = err
			log.Error("publish event failed", zap.Error(err))
			return nil, err
		}
	}

	// 6) коммитим всё
	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		log.Error("transaction commit failed", zap.Error(commitErr))
		return nil, commitErr
	}

	return result, nil
}

func (s *Service) GetGiftsByIDs(ctx context.Context, giftIDs []string) ([]*gift.Gift, error) {
	gifts, err := s.repo.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		return nil, err
	}
	return gifts, nil
}
