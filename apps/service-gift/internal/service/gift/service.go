package gift

import (
	"context"
	"encoding/base64"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	giftEvents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	telegrambotv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/telegrambot/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	txMgr                    pg.TxManager
	repo                     gift.GiftRepository
	log                      *logger.Logger
	paymentPrivateClient     paymentv1.PaymentPrivateServiceClient
	telegramBotPrivateClient telegrambotv1.TelegramBotPrivateServiceClient
}

func New(repo gift.GiftRepository, txMgr pg.TxManager, log *logger.Logger, clients *clients.Clients) *Service {
	return &Service{
		repo:                     repo,
		txMgr:                    txMgr,
		log:                      log,
		paymentPrivateClient:     clients.Payment.Private,
		telegramBotPrivateClient: clients.TelegramBot.Private,
	}
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

type ExecuteWithdrawResult struct {
	Gifts             []*gift.Gift
	StarsInvoiceURL   string // только для Stars валюты
	IsStarsCommission bool   // указывает, что это Stars комиссия
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

func (s *Service) ExecuteWithdraw(ctx context.Context, telegramUserID int64, giftIDs []string, commissionCurrency giftv1.ExecuteWithdrawRequest_CommissionCurrency) (*ExecuteWithdrawResult, error) {
	switch commissionCurrency {
	case giftv1.ExecuteWithdrawRequest_COMMISSION_CURRENCY_TON:
		return s.executeWithdrawTON(ctx, telegramUserID, giftIDs)
	case giftv1.ExecuteWithdrawRequest_COMMISSION_CURRENCY_STARS:
		return s.executeWithdrawStars(ctx, telegramUserID, giftIDs)
	default:
		return nil, errors.NewValidationError("commission currency", "invalid commission currency")
	}
}

func (s *Service) executeWithdrawTON(ctx context.Context, telegramUserID int64, giftIDs []string) (*ExecuteWithdrawResult, error) {
	// Начинаем транзакцию, чтобы сохранить пометки и публиковать события
	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	var commitErr error
	defer func() {
		if commitErr != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.log.Error("rollback failed", zap.Error(rbErr))
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
		s.log.Error("publisher init failed", zap.Error(err))
		return nil, err
	}
	fwdPub := forwarder.NewPublisher(sqlPub, forwarder.PublisherConfig{
		ForwarderTopic: giftEvents.SqlOutboxTopic,
	})

	// Валидация: получаем все подарки и проверяем владельца/статус
	gifts, err := repo.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		commitErr = err
		s.log.Error("failed to get gifts by IDs", zap.Error(err))
		return nil, err
	}

	for _, g := range gifts {
		if g.OwnerTelegramID != telegramUserID {
			commitErr = err
			s.log.Error("gift does not belong to user",
				zap.String("giftID", g.ID),
				zap.Int64("ownerID", g.OwnerTelegramID),
			)
			return nil, errors.NewGiftNotOwnedError("one or more gifts do not belong to you")
		}
		if g.Status != gift.StatusOwned {
			commitErr = err
			s.log.Error("gift is not in owned status",
				zap.String("giftID", g.ID),
				zap.String("status", string(g.Status)),
			)
			return nil, errors.NewGiftNotOwnedError("one or more gifts are not available for withdrawal")
		}
	}

	var result []*gift.Gift
	var eventsToPublish []*message.Message

	for _, giftID := range giftIDs {
		gift, err := repo.GetGiftByID(ctx, giftID)
		if err != nil {
			commitErr = err
			s.log.Error("failed to get gift by ID", zap.Error(err))
			return nil, err
		}
		previewResp, err := s.paymentPrivateClient.PreviewWithdraw(ctx, &paymentv1.PrivatePreviewWithdrawRequest{
			TonAmount: &sharedv1.TonAmount{Value: gift.Price},
		})
		if err != nil {
			commitErr = err
			s.log.Error("failed to preview withdraw", zap.Error(err))
			return nil, err
		}
		// списываем комиссию и получаем её величину
		_, err = s.paymentPrivateClient.SpendUserBalance(ctx, &paymentv1.SpendUserBalanceRequest{
			TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
			TonAmount:      &sharedv1.TonAmount{Value: previewResp.TotalTonFee.Value},
			Reason:         paymentv1.TransactionReason_TRANSACTION_REASON_WITHDRAW,
			Metadata: &paymentv1.TransactionMetadata{
				Data: &paymentv1.TransactionMetadata_Gift{
					Gift: &paymentv1.TransactionMetadata_GiftDetails{
						GiftId: gift.ID,
						Title:  gift.Title,
						Slug:   gift.Slug,
					},
				},
			},
		})
		if err != nil {
			commitErr = err
			s.log.Error("failed to spend withdrawal commission", zap.Error(err))
			return nil, err
		}

		// помечаем подарок на вывод
		g, err := repo.MarkGiftForWithdrawal(ctx, giftID)
		if err != nil {
			commitErr = err
			s.log.Error("failed to mark gift for withdrawal", zap.Error(err))
			return nil, err
		}
		result = append(result, g)

		// готовим событие для публикации после коммита
		ev := &giftv1.GiftWithdrawRequestedEvent{
			GiftId:           &sharedv1.GiftId{Value: g.ID},
			OwnerTelegramId:  &sharedv1.TelegramUserId{Value: g.OwnerTelegramID},
			TelegramGiftId:   &sharedv1.GiftTelegramId{Value: g.TelegramGiftID},
			CollectibleId:    int32(g.CollectibleID),
			UpgradeMessageId: g.UpgradeMessageID,
			Price:            &sharedv1.TonAmount{Value: g.Price},
			CommissionAmount: &sharedv1.TonAmount{Value: previewResp.TotalTonFee.Value},
			Title:            g.Title,
			Slug:             g.Slug,
		}
		payload, err := proto.Marshal(ev)
		if err != nil {
			commitErr = err
			s.log.Error("marshal event failed", zap.Error(err))
			return nil, err
		}
		msg := message.NewMessage(watermill.NewUUID(), payload)
		eventsToPublish = append(eventsToPublish, msg)
	}

	// публикуем все события только после успешной обработки всех подарков
	for _, msg := range eventsToPublish {
		if err := fwdPub.Publish(giftEvents.TopicGiftWithdrawRequested.String(), msg); err != nil {
			commitErr = err
			s.log.Error("publish event failed", zap.Error(err))
			return nil, err
		}
	}

	// коммитим всё
	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		s.log.Error("transaction commit failed", zap.Error(commitErr))
		return nil, commitErr
	}

	return &ExecuteWithdrawResult{
		Gifts:             result,
		IsStarsCommission: false,
	}, nil
}

func (s *Service) executeWithdrawStars(
	ctx context.Context,
	telegramUserID int64,
	giftIDs []string,
) (*ExecuteWithdrawResult, error) {
	// 1) Начинаем транзакцию для блокировки подарков
	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	var commitErr error
	defer func() {
		if commitErr != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	repo := s.repo.WithTx(tx)

	// 2) Получаем все подарки одним запросом
	gifts, err := repo.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		commitErr = err
		s.log.Error("failed to get gifts by IDs", zap.Error(err))
		return nil, err
	}

	// 3) Проверяем, что все подарки принадлежат пользователю и в статусе Owned
	for _, g := range gifts {
		if g.OwnerTelegramID != telegramUserID {
			commitErr = errors.NewGiftNotOwnedError("gift does not belong to user: " + g.ID)
			s.log.Error("ownership validation failed", zap.String("giftID", g.ID))
			return nil, commitErr
		}
		if g.Status != gift.StatusOwned {
			commitErr = errors.NewGiftNotOwnedError("gift is not available for withdrawal: " + g.ID)
			s.log.Error("status validation failed", zap.String("giftID", g.ID), zap.String("status", string(g.Status)))
			return nil, commitErr
		}
	}

	// 4) Блокируем подарки (помечаем withdraw_pending)
	// blocked := make([]*gift.Gift, 0, len(gifts))
	// for _, g := range gifts {
	// 	b, err := repo.MarkGiftForWithdrawal(ctx, g.ID)
	// 	if err != nil {
	// 		commitErr = err
	// 		s.log.Error("failed to mark gift for withdrawal", zap.Error(err), zap.String("giftID", g.ID))
	// 		return nil, err
	// 	}
	// 	blocked = append(blocked, b)
	// }

	// 5) Считаем комиссию в звёздах для каждого подарка
	var totalStars uint32
	commissions := make([]*telegrambotv1.GiftCommission, len(gifts))
	for i, g := range gifts {
		preview, err := s.paymentPrivateClient.PreviewWithdraw(ctx, &paymentv1.PrivatePreviewWithdrawRequest{
			TonAmount: &sharedv1.TonAmount{Value: g.Price},
		})
		if err != nil {
			commitErr = err
			s.log.Error("failed to preview withdraw", zap.Error(err), zap.String("giftID", g.ID))
			return nil, err
		}
		stars := preview.TotalStarsFee.Value
		totalStars += stars

		commissions[i] = &telegrambotv1.GiftCommission{
			GiftId: &sharedv1.GiftId{Value: g.ID},
			Stars:  &sharedv1.StarsAmount{Value: stars},
		}
	}

	// 6) Собираем payload для инвойса
	payloadMsg := &telegrambotv1.StarInvoicePayload{
		Purpose: &telegrambotv1.StarInvoicePayload_GiftWithdrawCommission{
			GiftWithdrawCommission: &telegrambotv1.GiftWithdrawCommission{
				GiftCommissions: commissions,
			},
		},
	}
	payloadBytes, err := proto.Marshal(payloadMsg)
	if err != nil {
		commitErr = err
		s.log.Error("failed to marshal StarInvoicePayload", zap.Error(err))
		return nil, err
	}
	payloadB64 := base64.StdEncoding.EncodeToString(payloadBytes)

	// 7) Создаём инвойс в TelegramBot
	invoiceResp, err := s.telegramBotPrivateClient.CreateStarInvoice(ctx, &telegrambotv1.CreateStarInvoiceRequest{
		TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
		StarsAmount:    &sharedv1.StarsAmount{Value: totalStars}, // <— здесь totalStars
		Title:          "Gift Withdrawal Commission",
		Description:    "Commission for withdrawing your gifts",
		Payload:        payloadB64,
	})
	if err != nil {
		commitErr = err
		s.log.Error("failed to create star invoice", zap.Error(err))
		return nil, err
	}

	// 8) Коммитим транзакцию — подарки заблокированы до завершения оплаты
	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		s.log.Error("transaction commit failed", zap.Error(commitErr))
		return nil, commitErr
	}

	return &ExecuteWithdrawResult{
		Gifts:             gifts,
		StarsInvoiceURL:   invoiceResp.InvoiceUrl,
		IsStarsCommission: true,
	}, nil
}

func (s *Service) CompleteStarsWithdrawal(ctx context.Context, telegramUserID int64, giftIDs []string, starsCommission uint32) ([]*gift.Gift, error) {
	// Начинаем транзакцию, чтобы сохранить пометки и публиковать события
	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	var commitErr error
	defer func() {
		if commitErr != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.log.Error("rollback failed", zap.Error(rbErr))
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
		s.log.Error("publisher init failed", zap.Error(err))
		return nil, err
	}
	fwdPub := forwarder.NewPublisher(sqlPub, forwarder.PublisherConfig{
		ForwarderTopic: giftEvents.SqlOutboxTopic,
	})

	var result []*gift.Gift
	for _, giftID := range giftIDs {
		g, err := repo.MarkGiftForWithdrawal(ctx, giftID)
		if err != nil {
			commitErr = err
			s.log.Error("failed to mark gift for withdrawal", zap.Error(err))
			return nil, err
		}

		// // Проверяем что подарок действительно в нужном статусе
		// if g.Status != gift.StatusWithdrawPending {
		// 	commitErr = err
		// 	s.log.Error("gift is not in withdraw_pending status",
		// 		zap.String("giftID", g.ID),
		// 		zap.String("status", string(g.Status)),
		// 	)
		// 	return nil, errors.NewGiftNotOwnedError("gift is not ready for withdrawal completion")
		// }

		result = append(result, g)

		// Получаем правильную TON комиссию через PreviewWithdraw
		previewResp, err := s.paymentPrivateClient.PreviewWithdraw(ctx, &paymentv1.PrivatePreviewWithdrawRequest{
			TonAmount: &sharedv1.TonAmount{Value: g.Price},
		})
		if err != nil {
			commitErr = err
			s.log.Error("failed to preview withdraw for commission calculation", zap.Error(err))
			return nil, err
		}

		// публикуем событие с правильной TON комиссией
		ev := &giftv1.GiftWithdrawRequestedEvent{
			GiftId:           &sharedv1.GiftId{Value: g.ID},
			OwnerTelegramId:  &sharedv1.TelegramUserId{Value: g.OwnerTelegramID},
			TelegramGiftId:   &sharedv1.GiftTelegramId{Value: g.TelegramGiftID},
			CollectibleId:    int32(g.CollectibleID),
			UpgradeMessageId: g.UpgradeMessageID,
			Price:            &sharedv1.TonAmount{Value: g.Price},
			CommissionAmount: &sharedv1.TonAmount{Value: previewResp.TotalTonFee.Value}, // используем правильную конвертацию
			Title:            g.Title,
			Slug:             g.Slug,
		}
		payload, err := proto.Marshal(ev)
		if err != nil {
			commitErr = err
			s.log.Error("marshal event failed", zap.Error(err))
			return nil, err
		}
		msg := message.NewMessage(watermill.NewUUID(), payload)
		if err = fwdPub.Publish(giftEvents.TopicGiftWithdrawRequested.String(), msg); err != nil {
			commitErr = err
			s.log.Error("publish event failed", zap.Error(err))
			return nil, err
		}
	}

	// коммитим всё
	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		s.log.Error("transaction commit failed", zap.Error(commitErr))
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
