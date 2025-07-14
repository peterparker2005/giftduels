package gift

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ccoveille/go-safecast"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg"
	giftDomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	giftEvents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	telegrambotv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/telegrambot/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	txMgr                    pg.TxManager
	repo                     giftDomain.Repository
	log                      *logger.Logger
	paymentPrivateClient     paymentv1.PaymentPrivateServiceClient
	telegramBotPrivateClient telegrambotv1.TelegramBotPrivateServiceClient
}

func New(
	repo giftDomain.Repository,
	txMgr pg.TxManager,
	log *logger.Logger,
	clients *clients.Clients,
) *Service {
	return &Service{
		repo:                     repo,
		txMgr:                    txMgr,
		log:                      log,
		paymentPrivateClient:     clients.Payment.Private,
		telegramBotPrivateClient: clients.TelegramBot.Private,
	}
}

func (s *Service) GetGiftByID(ctx context.Context, id string) (*giftDomain.Gift, error) {
	gift, err := s.repo.GetGiftByID(ctx, id)
	if err != nil {
		return nil, err
	}

	collection, err := s.repo.FindCollectionByName(ctx, gift.Title) // Using title as collection name
	if err != nil {
		// If collection not found, create a default one
		collection, err = s.repo.CreateCollection(ctx, &giftDomain.CreateCollectionParams{
			Name:      gift.Title,
			ShortName: giftDomain.ShortName(gift.Title),
		})
		if err != nil {
			return nil, err
		}
	}

	model, err := s.repo.GetGiftModel(ctx, gift.Model.ID)
	if err != nil {
		return nil, err
	}
	backdrop, err := s.repo.GetGiftBackdrop(ctx, gift.Backdrop.ID)
	if err != nil {
		return nil, err
	}
	symbol, err := s.repo.GetGiftSymbol(ctx, gift.Symbol.ID)
	if err != nil {
		return nil, err
	}

	// Add nil checks to prevent panic
	if collection != nil {
		gift.Collection = *collection
	}
	if model != nil {
		gift.Model = *model
	}
	if backdrop != nil {
		gift.Backdrop = *backdrop
	}
	if symbol != nil {
		gift.Symbol = *symbol
	}

	return gift, nil
}

type GetUserGiftsResult struct {
	Gifts      []*giftDomain.Gift
	Total      int32
	TotalValue *tonamount.TonAmount
}

type ExecuteWithdrawResult struct {
	Gifts             []*giftDomain.Gift
	StarsInvoiceURL   string // только для Stars валюты
	IsStarsCommission bool   // указывает, что это Stars комиссия
}

func (s *Service) GetUserGifts(
	ctx context.Context,
	telegramUserID int64,
	pagination *shared.PageRequest,
) (*GetUserGiftsResult, error) {
	res, err := s.repo.GetUserGifts(ctx, pagination.PageSize(), pagination.Offset(), telegramUserID)
	if err != nil {
		return nil, err
	}

	// Populate attributes for all gifts
	if err = s.populateGiftAttributes(ctx, res.Gifts); err != nil {
		return nil, err
	}

	totalValue, err := tonamount.NewTonAmountFromNano(0) // Start with zero TON amount
	if err != nil {
		return nil, err
	}
	for _, g := range res.Gifts {
		if g.Price != nil {
			totalValue = totalValue.Add(g.Price)
		}
	}

	total, err := safecast.ToInt32(res.Total)
	if err != nil {
		return nil, err
	}

	return &GetUserGiftsResult{
		Gifts:      res.Gifts,
		Total:      total,
		TotalValue: totalValue,
	}, nil
}

func (s *Service) GetUserActiveGifts(
	ctx context.Context,
	telegramUserID int64,
	pagination *shared.PageRequest,
) (*GetUserGiftsResult, error) {
	res, err := s.repo.GetUserActiveGifts(
		ctx,
		pagination.PageSize(),
		pagination.Offset(),
		telegramUserID,
	)
	if err != nil {
		return nil, err
	}

	// Populate attributes for all gifts
	if err = s.populateGiftAttributes(ctx, res.Gifts); err != nil {
		return nil, err
	}

	// TODO: calculate total value in one query
	totalValue, err := tonamount.NewTonAmountFromNano(0) // Start with zero TON amount
	if err != nil {
		return nil, err
	}
	for _, g := range res.Gifts {
		if g.Price != nil {
			totalValue = totalValue.Add(g.Price)
		}
	}

	total, err := safecast.ToInt32(res.Total)
	if err != nil {
		return nil, err
	}

	return &GetUserGiftsResult{
		Gifts:      res.Gifts,
		Total:      total,
		TotalValue: totalValue,
	}, nil
}

// populateGiftAttributes populates the Model, Backdrop, and Symbol attributes for a slice of gifts.
func (s *Service) populateGiftAttributes(ctx context.Context, gifts []*giftDomain.Gift) error {
	// Collect all unique IDs
	modelIDs := make(map[int32]bool)
	backdropIDs := make(map[int32]bool)
	symbolIDs := make(map[int32]bool)

	for _, gift := range gifts {
		modelIDs[gift.Model.ID] = true
		backdropIDs[gift.Backdrop.ID] = true
		symbolIDs[gift.Symbol.ID] = true
	}

	// Fetch all models, backdrops, and symbols in parallel
	modelChan := make(chan map[int32]*giftDomain.Model, 1)
	backdropChan := make(chan map[int32]*giftDomain.Backdrop, 1)
	symbolChan := make(chan map[int32]*giftDomain.Symbol, 1)
	//nolint:mnd // 3 is not a magic number
	errorChan := make(chan error, 3)

	// Fetch models
	go func() {
		models := make(map[int32]*giftDomain.Model)
		for modelID := range modelIDs {
			model, err := s.repo.GetGiftModel(ctx, modelID)
			if err != nil {
				errorChan <- err
				return
			}
			models[modelID] = model
		}
		modelChan <- models
	}()

	// Fetch backdrops
	go func() {
		backdrops := make(map[int32]*giftDomain.Backdrop)
		for backdropID := range backdropIDs {
			backdrop, err := s.repo.GetGiftBackdrop(ctx, backdropID)
			if err != nil {
				errorChan <- err
				return
			}
			backdrops[backdropID] = backdrop
		}
		backdropChan <- backdrops
	}()

	// Fetch symbols
	go func() {
		symbols := make(map[int32]*giftDomain.Symbol)
		for symbolID := range symbolIDs {
			symbol, err := s.repo.GetGiftSymbol(ctx, symbolID)
			if err != nil {
				errorChan <- err
				return
			}
			symbols[symbolID] = symbol
		}
		symbolChan <- symbols
	}()

	// Wait for all goroutines to complete
	models := <-modelChan
	backdrops := <-backdropChan
	symbols := <-symbolChan

	// Check for errors
	select {
	case err := <-errorChan:
		return err
	default:
	}

	// Populate the gifts with the fetched data
	for _, gift := range gifts {
		if model, exists := models[gift.Model.ID]; exists && model != nil {
			gift.Model = *model
		} else {
			s.log.Warn("Model not found for gift", zap.String("giftID", gift.ID), zap.Int32("modelID", gift.Model.ID))
		}
		if backdrop, exists := backdrops[gift.Backdrop.ID]; exists && backdrop != nil {
			gift.Backdrop = *backdrop
		} else {
			s.log.Warn("Backdrop not found for gift", zap.String("giftID", gift.ID), zap.Int32("backdropID", gift.Backdrop.ID))
		}
		if symbol, exists := symbols[gift.Symbol.ID]; exists && symbol != nil {
			gift.Symbol = *symbol
		} else {
			s.log.Warn("Symbol not found for gift", zap.String("giftID", gift.ID), zap.Int32("symbolID", gift.Symbol.ID))
		}
	}

	return nil
}

type StakeGiftParams struct {
	GiftID       string
	GameMetadata *giftv1.StakeGiftRequest_GameMetadata
}

type GameMetadata struct {
	GameMode sharedv1.GameMode
	GameID   string
}

func (s *Service) StakeGift(ctx context.Context, params StakeGiftParams) (*giftDomain.Gift, error) {
	gift, err := s.repo.StakeGiftForGame(ctx, params.GiftID)
	if err != nil {
		return nil, err
	}
	return gift, nil
}

func (s *Service) ExecuteWithdraw(
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
		return nil, errors.NewValidationError(
			"commission currency",
			"unspecified commission currency",
		)
	default:
		return nil, errors.NewValidationError("commission currency", "invalid commission currency")
	}
}

func (s *Service) executeWithdrawTON(
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

func (s *Service) preparePublisher() (*forwarder.Publisher, error) {
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

func (s *Service) validateGiftsOwnership(gifts []*giftDomain.Gift, telegramUserID int64) error {
	for _, g := range gifts {
		if !g.CanBeWithdrawnBy(telegramUserID) {
			s.log.Error("gift cannot be withdrawn by user",
				zap.String("giftID", g.ID),
				zap.Int64("userID", telegramUserID),
				zap.Int64("ownerID", g.OwnerTelegramID),
				zap.String("status", string(g.Status)),
			)
			return errors.NewGiftNotOwnedError("one or more gifts cannot be withdrawn")
		}
	}
	return nil
}

func (s *Service) processGiftsForWithdrawal(
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

func (s *Service) publishEvents(
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

func (s *Service) getAndValidateGiftsForStarsWithdrawal(
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
			return nil, errors.NewGiftNotOwnedError("gift cannot be withdrawn: " + g.ID)
		}
	}

	return gifts, nil
}

func (s *Service) calculateStarsCommissions(
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

func (s *Service) createStarsInvoice(
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

func (s *Service) executeWithdrawStars(
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

func (s *Service) CompleteStarsWithdrawal(
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

	var result []*giftDomain.Gift
	var eventsToPublish []*message.Message

	for _, giftID := range giftIDs {
		g, markErr := repo.MarkGiftForWithdrawal(ctx, giftID)
		if markErr != nil {
			commitErr = markErr
			s.log.Error("failed to mark gift for withdrawal", zap.Error(markErr))
			return nil, markErr
		}

		// Используем domain метод для валидации
		if err = g.CompleteWithdrawal(time.Now()); err != nil {
			commitErr = err
			s.log.Error("failed to complete withdrawal in domain",
				zap.String("giftID", g.ID),
				zap.Error(err),
			)
			return nil, err
		}

		result = append(result, g)

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
			commitErr = previewErr
			s.log.Error(
				"failed to preview withdraw for commission calculation",
				zap.Error(previewErr),
			)
			return nil, previewErr
		}

		// готовим событие для публикации
		ev := &giftv1.GiftWithdrawRequestedEvent{
			GiftId:           &sharedv1.GiftId{Value: g.ID},
			OwnerTelegramId:  &sharedv1.TelegramUserId{Value: g.OwnerTelegramID},
			TelegramGiftId:   &sharedv1.GiftTelegramId{Value: g.TelegramGiftID},
			CollectibleId:    g.CollectibleID,
			UpgradeMessageId: g.UpgradeMessageID,
			Price:            &sharedv1.TonAmount{Value: g.Price.String()},
			CommissionAmount: &sharedv1.TonAmount{
				Value: previewResp.GetTotalTonFee().GetValue(),
			},
			Title: g.Title,
			Slug:  g.Slug,
		}
		payload, marshalErr := proto.Marshal(ev)
		if marshalErr != nil {
			commitErr = marshalErr
			s.log.Error("marshal event failed", zap.Error(marshalErr))
			return nil, marshalErr
		}
		msg := message.NewMessage(watermill.NewUUID(), payload)
		eventsToPublish = append(eventsToPublish, msg)
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

func (s *Service) GetGiftsByIDs(ctx context.Context, giftIDs []string) ([]*giftDomain.Gift, error) {
	gifts, err := s.repo.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		return nil, err
	}

	// Populate attributes for all gifts
	if err = s.populateGiftAttributes(ctx, gifts); err != nil {
		return nil, err
	}

	return gifts, nil
}

// ProcessAttributesFromEvent extracts attributes from the event and creates or finds them in the database.
func (s *Service) ProcessAttributesFromEvent(
	ctx context.Context,
	ev *giftv1.TelegramGiftReceivedEvent,
) (*giftDomain.AttributeData, error) {
	// Process collection
	collection, err := s.repo.FindCollectionByName(ctx, ev.GetTitle())
	if err != nil {
		// Create default collection if not found
		collection, err = s.repo.CreateCollection(ctx, &giftDomain.CreateCollectionParams{
			Name:      ev.GetTitle(),
			ShortName: giftDomain.ShortName(ev.GetTitle()),
		})
		if err != nil {
			return nil, fmt.Errorf("create default collection: %w", err)
		}
	}
	collectionID := collection.ID

	// Process model
	model, err := s.repo.FindModelByName(ctx, ev.GetModel().GetName())
	if err != nil {
		// Create default model if not found
		model, err = s.repo.CreateModel(ctx, &giftDomain.CreateModelParams{
			CollectionID:   collectionID,
			Name:           ev.GetModel().GetName(),
			ShortName:      giftDomain.ShortName(ev.GetModel().GetName()),
			RarityPerMille: ev.GetModel().GetRarityPerMille(),
		})
		if err != nil {
			return nil, fmt.Errorf("create default model: %w", err)
		}
	}
	modelID := model.ID

	// Process backdrop
	backdrop, err := s.repo.FindBackdropByName(ctx, ev.GetBackdrop().GetName())
	if err != nil {
		// Create default backdrop if not found
		backdrop, err = s.repo.CreateBackdrop(ctx, &giftDomain.CreateBackdropParams{
			Name:           ev.GetBackdrop().GetName(),
			ShortName:      giftDomain.ShortName(ev.GetBackdrop().GetName()),
			RarityPerMille: ev.GetBackdrop().GetRarityPerMille(),
			CenterColor:    giftDomain.StringPtr(ev.GetBackdrop().GetCenterColor()),
			EdgeColor:      giftDomain.StringPtr(ev.GetBackdrop().GetEdgeColor()),
			PatternColor:   giftDomain.StringPtr(ev.GetBackdrop().GetPatternColor()),
			TextColor:      giftDomain.StringPtr(ev.GetBackdrop().GetTextColor()),
		})
		if err != nil {
			return nil, fmt.Errorf("create default backdrop: %w", err)
		}
	}
	backdropID := backdrop.ID

	// Process symbol
	symbol, err := s.repo.FindSymbolByName(ctx, ev.GetSymbol().GetName())
	if err != nil {
		// Create default symbol if not found
		symbol, err = s.repo.CreateSymbol(ctx, &giftDomain.CreateSymbolParams{
			Name:           ev.GetSymbol().GetName(),
			ShortName:      giftDomain.ShortName(ev.GetSymbol().GetName()),
			RarityPerMille: ev.GetSymbol().GetRarityPerMille(),
		})
		if err != nil {
			return nil, fmt.Errorf("create default symbol: %w", err)
		}
	}
	symbolID := symbol.ID

	return &giftDomain.AttributeData{
		CollectionID: collectionID,
		ModelID:      modelID,
		BackdropID:   backdropID,
		SymbolID:     symbolID,
	}, nil
}

// CreateGiftFromEvent creates a gift from event data.
func (s *Service) CreateGiftFromEvent(
	ctx context.Context,
	giftID string,
	ev *giftv1.TelegramGiftReceivedEvent,
	price *tonamount.TonAmount,
	attrs *giftDomain.AttributeData,
) (*giftDomain.Gift, error) {
	createParams := &giftDomain.CreateGiftParams{
		GiftID:           giftID,
		OwnerTelegramID:  ev.GetOwnerTelegramId().GetValue(),
		Price:            price,
		Title:            ev.GetTitle(),
		Slug:             ev.GetSlug(),
		CollectibleID:    ev.GetCollectibleId(),
		UpgradeMessageID: ev.GetUpgradeMessageId(),
		TelegramGiftID:   ev.GetTelegramGiftId().GetValue(),
	}

	return s.repo.CreateGift(
		ctx,
		createParams,
		attrs.CollectionID,
		attrs.ModelID,
		attrs.BackdropID,
		attrs.SymbolID,
	)
}
