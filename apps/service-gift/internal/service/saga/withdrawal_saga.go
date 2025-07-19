package saga

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg"
	giftDomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	giftEvents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	telegrambotv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/telegrambot/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type WithdrawalSaga struct {
	txMgr                    pg.TxManager
	repo                     giftDomain.Repository
	log                      *logger.Logger
	paymentPrivateClient     paymentv1.PaymentPrivateServiceClient
	telegramBotPrivateClient telegrambotv1.TelegramBotPrivateServiceClient
}

func NewWithdrawalSaga(
	repo giftDomain.Repository,
	txMgr pg.TxManager,
	log *logger.Logger,
	clients *clients.Clients,
) *WithdrawalSaga {
	return &WithdrawalSaga{
		repo:                     repo,
		txMgr:                    txMgr,
		log:                      log,
		paymentPrivateClient:     clients.Payment.Private,
		telegramBotPrivateClient: clients.TelegramBot.Private,
	}
}

type ExecuteWithdrawResult struct {
	Gifts             []*giftDomain.Gift
	StarsInvoiceURL   string // только для Stars валюты
	IsStarsCommission bool   // указывает, что это Stars комиссия
}

func (s *WithdrawalSaga) ExecuteWithdraw(
	ctx context.Context,
	telegramUserID int64,
	giftIDs []string,
	commissionCurrency giftv1.ExecuteWithdrawRequest_CommissionCurrency,
) (*ExecuteWithdrawResult, error) {
	switch commissionCurrency {
	case giftv1.ExecuteWithdrawRequest_COMMISSION_CURRENCY_TON:
		return s.executeWithdrawTON(ctx, telegramUserID, giftIDs)
	case giftv1.ExecuteWithdrawRequest_COMMISSION_CURRENCY_STARS:
		return s.executeWithdrawStars(ctx, telegramUserID, giftIDs)
	case giftv1.ExecuteWithdrawRequest_COMMISSION_CURRENCY_UNSPECIFIED:
		return nil, giftDomain.ErrInvalidCommissionCurrency
	default:
		return nil, giftDomain.ErrInvalidCommissionCurrency
	}
}

func (s *WithdrawalSaga) executeWithdrawTON(
	ctx context.Context,
	telegramUserID int64,
	giftIDs []string,
) (*ExecuteWithdrawResult, error) {
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
	fwdPub, err := s.preparePublisher()
	if err != nil {
		commitErr = err
		return nil, err
	}

	// Валидация: получаем все подарки и проверяем владельца/статус
	gifts, err := repo.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		commitErr = err
		s.log.Error("failed to get gifts by IDs", zap.Error(err))
		return nil, err
	}

	if err = s.validateGiftsOwnership(gifts, telegramUserID); err != nil {
		commitErr = err
		return nil, err
	}

	result, eventsToPublish, err := s.processGiftsForWithdrawal(ctx, repo, gifts, telegramUserID)
	if err != nil {
		commitErr = err
		return nil, err
	}

	// публикуем все события только после успешной обработки всех подарков
	if err = s.publishEvents(fwdPub, eventsToPublish); err != nil {
		commitErr = err
		return nil, err
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

func (s *WithdrawalSaga) preparePublisher() (*forwarder.Publisher, error) {
	sqlPub, err := sql.NewPublisher(
		s.txMgr.SQL(),
		sql.PublisherConfig{SchemaAdapter: sql.DefaultPostgreSQLSchema{}},
		logger.NewWatermill(s.log),
	)
	if err != nil {
		s.log.Error("publisher init failed", zap.Error(err))
		return nil, err
	}
	return forwarder.NewPublisher(sqlPub, forwarder.PublisherConfig{
		ForwarderTopic: giftEvents.SQLOutboxTopic,
	}), nil
}

func (s *WithdrawalSaga) validateGiftsOwnership(
	gifts []*giftDomain.Gift,
	telegramUserID int64,
) error {
	for _, g := range gifts {
		if !g.CanBeWithdrawnBy(telegramUserID) {
			s.log.Error("gift cannot be withdrawn by user",
				zap.String("giftID", g.ID),
				zap.Int64("userID", telegramUserID),
				zap.Int64("ownerID", g.OwnerTelegramID),
				zap.String("status", string(g.Status)),
			)
			return giftDomain.ErrGiftNotOwned
		}
	}
	return nil
}

func (s *WithdrawalSaga) processGiftsForWithdrawal(
	ctx context.Context,
	repo giftDomain.Repository,
	gifts []*giftDomain.Gift,
	telegramUserID int64,
) ([]*giftDomain.Gift, []*message.Message, error) {
	var result []*giftDomain.Gift
	var eventsToPublish []*message.Message

	for _, gift := range gifts {
		// Используем domain метод для валидации
		if err := gift.MarkForWithdrawal(); err != nil {
			s.log.Error("failed to mark gift for withdrawal in domain", zap.Error(err))
			return nil, nil, err
		}

		previewResp, err := s.paymentPrivateClient.PreviewWithdraw(
			ctx,
			&paymentv1.PreviewWithdrawRequest{
				Gifts: []*paymentv1.GiftWithdrawRequest{
					{
						GiftId: &sharedv1.GiftId{Value: gift.ID},
						Price:  &sharedv1.TonAmount{Value: gift.Price.String()},
					},
				},
			},
		)
		if err != nil {
			s.log.Error("failed to preview withdraw", zap.Error(err))
			return nil, nil, err
		}

		// списываем комиссию и получаем её величину
		_, err = s.paymentPrivateClient.SpendUserBalance(ctx, &paymentv1.SpendUserBalanceRequest{
			TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
			TonAmount:      &sharedv1.TonAmount{Value: previewResp.GetTotalTonFee().GetValue()},
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
			s.log.Error("failed to spend withdrawal commission", zap.Error(err))
			return nil, nil, err
		}

		// помечаем подарок на вывод в репозитории
		markedGift, err := repo.MarkGiftForWithdrawal(ctx, gift.ID)
		if err != nil {
			s.log.Error("failed to mark gift for withdrawal", zap.Error(err))
			return nil, nil, err
		}
		result = append(result, markedGift)

		// готовим событие для публикации после коммита
		ev := &giftv1.GiftWithdrawRequestedEvent{
			GiftId:           &sharedv1.GiftId{Value: markedGift.ID},
			OwnerTelegramId:  &sharedv1.TelegramUserId{Value: markedGift.OwnerTelegramID},
			TelegramGiftId:   &sharedv1.GiftTelegramId{Value: markedGift.TelegramGiftID},
			CollectibleId:    markedGift.CollectibleID,
			UpgradeMessageId: markedGift.UpgradeMessageID,
			Price:            &sharedv1.TonAmount{Value: markedGift.Price.String()},
			CommissionAmount: &sharedv1.TonAmount{Value: previewResp.GetTotalTonFee().GetValue()},
			Title:            markedGift.Title,
			Slug:             markedGift.Slug,
		}
		payload, err := proto.Marshal(ev)
		if err != nil {
			s.log.Error("marshal event failed", zap.Error(err))
			return nil, nil, err
		}
		msg := message.NewMessage(watermill.NewUUID(), payload)
		eventsToPublish = append(eventsToPublish, msg)
	}

	return result, eventsToPublish, nil
}

func (s *WithdrawalSaga) publishEvents(
	fwdPub *forwarder.Publisher,
	eventsToPublish []*message.Message,
) error {
	for _, msg := range eventsToPublish {
		if err := fwdPub.Publish(giftEvents.TopicGiftWithdrawRequested.String(), msg); err != nil {
			s.log.Error("publish event failed", zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *WithdrawalSaga) calculateStarsCommissions(
	ctx context.Context,
	gifts []*giftDomain.Gift,
) ([]*telegrambotv1.GiftCommission, uint32, error) {
	var totalStars uint32
	commissions := make([]*telegrambotv1.GiftCommission, len(gifts))

	for i, g := range gifts {
		preview, err := s.paymentPrivateClient.PreviewWithdraw(
			ctx,
			&paymentv1.PreviewWithdrawRequest{
				Gifts: []*paymentv1.GiftWithdrawRequest{
					{
						GiftId: &sharedv1.GiftId{Value: g.ID},
						Price:  &sharedv1.TonAmount{Value: g.Price.String()},
					},
				},
			},
		)
		if err != nil {
			s.log.Error("failed to preview withdraw",
				zap.Error(err),
				zap.String("giftID", g.ID),
			)
			return nil, 0, err
		}
		stars := preview.GetTotalStarsFee().GetValue()
		totalStars += stars

		commissions[i] = &telegrambotv1.GiftCommission{
			GiftId: &sharedv1.GiftId{Value: g.ID},
			Stars:  &sharedv1.StarsAmount{Value: stars},
		}
	}

	return commissions, totalStars, nil
}

func (s *WithdrawalSaga) createStarsInvoice(
	ctx context.Context,
	telegramUserID int64,
	totalStars uint32,
	commissions []*telegrambotv1.GiftCommission,
) (string, error) {
	// Собираем payload для инвойса
	payloadMsg := &telegrambotv1.StarInvoicePayload{
		Purpose: &telegrambotv1.StarInvoicePayload_GiftWithdrawCommission{
			GiftWithdrawCommission: &telegrambotv1.GiftWithdrawCommission{
				GiftCommissions: commissions,
			},
		},
	}
	payloadBytes, err := proto.Marshal(payloadMsg)
	if err != nil {
		s.log.Error("failed to marshal StarInvoicePayload", zap.Error(err))
		return "", err
	}
	payloadB64 := base64.StdEncoding.EncodeToString(payloadBytes)

	// Создаём инвойс в TelegramBot
	invoiceResp, err := s.telegramBotPrivateClient.CreateStarInvoice(
		ctx,
		&telegrambotv1.CreateStarInvoiceRequest{
			TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
			StarsAmount:    &sharedv1.StarsAmount{Value: totalStars},
			Title:          "Gift Withdrawal Commission",
			Description:    "Commission for withdrawing your gifts",
			Payload:        payloadB64,
		},
	)
	if err != nil {
		s.log.Error("failed to create star invoice", zap.Error(err))
		return "", err
	}

	return invoiceResp.GetInvoiceUrl(), nil
}

func (s *WithdrawalSaga) executeWithdrawStars(
	ctx context.Context,
	telegramUserID int64,
	giftIDs []string,
) (*ExecuteWithdrawResult, error) {
	// Начинаем транзакцию для блокировки подарков
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

	// Получаем и валидируем подарки
	gifts, err := s.getAndValidateGiftsForStarsWithdrawal(ctx, repo, giftIDs, telegramUserID)
	if err != nil {
		commitErr = err
		return nil, err
	}

	// Рассчитываем комиссии и создаем инвойс
	commissions, totalStars, err := s.calculateStarsCommissions(ctx, gifts)
	if err != nil {
		commitErr = err
		return nil, err
	}

	// Создаем инвойс
	invoiceURL, err := s.createStarsInvoice(ctx, telegramUserID, totalStars, commissions)
	if err != nil {
		commitErr = err
		return nil, err
	}

	// Коммитим транзакцию — подарки заблокированы до завершения оплаты
	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		s.log.Error("transaction commit failed", zap.Error(commitErr))
		return nil, commitErr
	}

	return &ExecuteWithdrawResult{
		Gifts:             gifts,
		StarsInvoiceURL:   invoiceURL,
		IsStarsCommission: true,
	}, nil
}

func (s *WithdrawalSaga) getAndValidateGiftsForStarsWithdrawal(
	ctx context.Context,
	repo giftDomain.Repository,
	giftIDs []string,
	telegramUserID int64,
) ([]*giftDomain.Gift, error) {
	gifts, err := repo.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		s.log.Error("failed to get gifts by IDs", zap.Error(err))
		return nil, err
	}

	// Используем domain методы для валидации
	for _, g := range gifts {
		if !g.CanBeWithdrawnBy(telegramUserID) {
			s.log.Error("gift cannot be withdrawn by user",
				zap.String("giftID", g.ID),
				zap.Int64("userID", telegramUserID),
				zap.Int64("ownerID", g.OwnerTelegramID),
				zap.String("status", string(g.Status)),
			)
			return nil, giftDomain.ErrGiftNotOwned
		}
	}

	return gifts, nil
}

func (s *WithdrawalSaga) CompleteStarsWithdrawal(
	ctx context.Context,
	giftIDs []string,
) ([]*giftDomain.Gift, error) {
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

	fwdPub, err := s.preparePublisher()
	if err != nil {
		commitErr = err
		return nil, err
	}

	result, eventsToPublish, err := s.processGiftsForStarsWithdrawal(ctx, repo, giftIDs)
	if err != nil {
		commitErr = err
		return nil, err
	}

	// публикуем все события
	if err = s.publishEvents(fwdPub, eventsToPublish); err != nil {
		commitErr = err
		return nil, err
	}

	// коммитим всё
	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		s.log.Error("transaction commit failed", zap.Error(commitErr))
		return nil, commitErr
	}

	return result, nil
}

func (s *WithdrawalSaga) processGiftsForStarsWithdrawal(
	ctx context.Context,
	repo giftDomain.Repository,
	giftIDs []string,
) ([]*giftDomain.Gift, []*message.Message, error) {
	var result []*giftDomain.Gift
	var eventsToPublish []*message.Message

	for _, giftID := range giftIDs {
		gift, event, err := s.processSingleGiftForStarsWithdrawal(ctx, repo, giftID)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, gift)
		eventsToPublish = append(eventsToPublish, event)
	}

	return result, eventsToPublish, nil
}

func (s *WithdrawalSaga) processSingleGiftForStarsWithdrawal(
	ctx context.Context,
	repo giftDomain.Repository,
	giftID string,
) (*giftDomain.Gift, *message.Message, error) {
	g, markErr := repo.MarkGiftForWithdrawal(ctx, giftID)
	if markErr != nil {
		s.log.Error("failed to mark gift for withdrawal", zap.Error(markErr))
		return nil, nil, markErr
	}

	// Используем domain метод для валидации
	if err := g.CompleteWithdrawal(time.Now()); err != nil {
		s.log.Error("failed to complete withdrawal in domain",
			zap.String("giftID", g.ID),
			zap.Error(err),
		)
		return nil, nil, err
	}

	// Получаем правильную TON комиссию через PreviewWithdraw
	previewResp, previewErr := s.paymentPrivateClient.PreviewWithdraw(
		ctx,
		&paymentv1.PreviewWithdrawRequest{
			Gifts: []*paymentv1.GiftWithdrawRequest{
				{
					GiftId: &sharedv1.GiftId{Value: g.ID},
					Price:  &sharedv1.TonAmount{Value: g.Price.String()},
				},
			},
		},
	)
	if previewErr != nil {
		s.log.Error(
			"failed to preview withdraw for commission calculation",
			zap.Error(previewErr),
		)
		return nil, nil, previewErr
	}

	// готовим событие для публикации
	event, err := s.createWithdrawEvent(g, previewResp.GetTotalTonFee().GetValue())
	if err != nil {
		return nil, nil, err
	}

	return g, event, nil
}

func (s *WithdrawalSaga) createWithdrawEvent(
	gift *giftDomain.Gift,
	commissionAmount string,
) (*message.Message, error) {
	ev := &giftv1.GiftWithdrawRequestedEvent{
		GiftId:           &sharedv1.GiftId{Value: gift.ID},
		OwnerTelegramId:  &sharedv1.TelegramUserId{Value: gift.OwnerTelegramID},
		TelegramGiftId:   &sharedv1.GiftTelegramId{Value: gift.TelegramGiftID},
		CollectibleId:    gift.CollectibleID,
		UpgradeMessageId: gift.UpgradeMessageID,
		Price:            &sharedv1.TonAmount{Value: gift.Price.String()},
		CommissionAmount: &sharedv1.TonAmount{
			Value: commissionAmount,
		},
		Title: gift.Title,
		Slug:  gift.Slug,
	}
	payload, err := proto.Marshal(ev)
	if err != nil {
		s.log.Error("marshal event failed", zap.Error(err))
		return nil, err
	}
	return message.NewMessage(watermill.NewUUID(), payload), nil
}

func (s *WithdrawalSaga) CancelGiftWithdrawal(
	ctx context.Context,
	giftID string,
) (*giftDomain.Gift, error) {
	return s.repo.CancelGiftWithdrawal(ctx, giftID)
}
