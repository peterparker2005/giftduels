package payment

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/ton"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/peterparker2005/giftduels/packages/shared"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
)

const (
	// Минимальное базовое кол-во звёзд.
	baseStarsCommission = 25.0

	// Максимальное кол-во звёзд.
	// maxStarsCommission = 250.0
	maxStarsCommission = 1.0

	// Сколько TON стоит одна звезда.
	tonPerStar = 0.2678 / 50.0

	// Какой процент от стоимости подарка мы берём в виде комиссии (в звёздах)
	// Например, 5% от стоимости подарка.
	commissionRate = 0.15
)

type Service struct {
	log     *logger.Logger
	repo    payment.Repository
	tonRepo ton.DepositRepository
	txMgr   pg.TxManager
}

func NewService(
	repo payment.Repository,
	tonRepo ton.DepositRepository,
	log *logger.Logger,
	txMgr pg.TxManager,
) *Service {
	return &Service{
		log:     log,
		repo:    repo,
		tonRepo: tonRepo,
		txMgr:   txMgr,
	}
}

func (s *Service) CreateDeposit(
	ctx context.Context,
	telegramUserID int64,
	amount string,
) (*ton.Deposit, error) {
	rawPayload := uuid.New().String()
	tonAmount, err := tonamount.NewTonAmountFromString(amount)
	if err != nil {
		return nil, err
	}
	nanoAmount, err := tonAmount.ToNano()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(time.Hour)

	params := &ton.CreateDepositParams{
		TelegramUserID: telegramUserID,
		AmountNano:     nanoAmount,
		Payload:        rawPayload, // просто UUID
		ExpiresAt:      expiresAt,
	}
	return s.tonRepo.CreateDeposit(ctx, params)
}

func (s *Service) ProcessDepositTransaction(
	ctx context.Context,
	payload, txHash string,
	txLt uint64,
	amount *tonamount.TonAmount,
) error {
	deposit, err := s.tonRepo.GetDepositByPayload(ctx, payload)
	if err != nil {
		return err
	}

	if deposit.Status != ton.DepositStatusPending {
		// Or log and ignore
		return nil
	}

	amountNano, err := amount.ToNano()
	if err != nil {
		return err
	}

	if deposit.AmountNano > amountNano {
		// Handle partial payment, for now, we ignore
		return nil
	}

	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				s.log.Error("failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	repo := s.repo.WithTx(tx)
	tonRepo := s.tonRepo.WithTx(tx)

	_, err = tonRepo.SetDepositTransaction(ctx, &ton.SetDepositTransactionParams{
		ID:     deposit.ID.String(),
		TxHash: txHash,
		TxLt:   txLt,
	})
	if err != nil {
		return err
	}

	// Используем domain для конвертации нано в TON
	tonAmount, err := tonamount.NewTonAmountFromNano(amountNano)
	if err != nil {
		return err
	}

	err = repo.CreateTransaction(ctx, &payment.CreateTransactionParams{
		TelegramUserID: deposit.TelegramUserID,
		Amount:         tonAmount,
		Reason:         payment.TransactionReasonDeposit,
	})
	if err != nil {
		return err
	}

	addBalanceParams := &payment.AddUserBalanceParams{
		TelegramUserID: deposit.TelegramUserID,
		Amount:         tonAmount,
	}

	_, err = repo.AddUserBalance(ctx, addBalanceParams)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetBalance(ctx context.Context, telegramUserID int64) (*payment.Balance, error) {
	return s.repo.GetUserBalance(ctx, telegramUserID)
}

func (s *Service) SpendUserBalance(
	ctx context.Context,
	telegramUserID int64,
	amount string,
	reason payment.TransactionReason,
	metadata *payment.TransactionMetadata,
) (*payment.Balance, error) {
	log := s.log.With(
		zap.Int64("telegram_user_id", telegramUserID),
		zap.String("amount", amount),
		zap.String("reason", string(reason)),
	)

	tonAmount, err := tonamount.NewTonAmountFromString(amount)
	if err != nil {
		log.Error("failed to parse amount", zap.Error(err))
		return nil, err
	}

	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		log.Error("failed to begin transaction", zap.Error(err))
		return nil, err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				log.Error("failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	repo := s.repo.WithTx(tx)

	currentBalance, err := repo.GetUserBalance(ctx, telegramUserID)
	if err != nil {
		return nil, err
	}

	if currentBalance.TonAmount.Decimal().Cmp(tonAmount.Decimal()) < 0 {
		return nil, ErrInsufficientBalance
	}

	balance, err := repo.SpendUserBalance(ctx, &payment.SpendUserBalanceParams{
		TelegramUserID: telegramUserID,
		Amount:         tonAmount,
	})
	if err != nil {
		return nil, err
	}

	var metadataBytes []byte
	if metadata != nil {
		metadataBytes, err = json.Marshal(metadata)
		if err != nil {
			log.Error("failed to marshal metadata", zap.Error(err))
			// не возвращаем ошибку, просто логируем
		}
	}

	negativeAmount := tonAmount.Negate()
	err = repo.CreateTransaction(ctx, &payment.CreateTransactionParams{
		TelegramUserID: telegramUserID,
		Amount:         negativeAmount,
		Reason:         reason,
		Metadata:       metadataBytes,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error("failed to commit transaction", zap.Error(err))
		return nil, err
	}

	return balance, nil
}

func (s *Service) AddUserBalance(
	ctx context.Context,
	telegramUserID int64,
	amount string,
	reason payment.TransactionReason,
	metadata *payment.TransactionMetadata,
) (*payment.Balance, error) {
	log := s.log.With(zap.Int64("telegram_user_id", telegramUserID), zap.String("amount", amount))

	tonAmount, err := tonamount.NewTonAmountFromString(amount)
	if err != nil {
		log.Error("failed to parse amount", zap.Error(err))
		return nil, err
	}

	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		log.Error("failed to begin transaction", zap.Error(err))
		return nil, err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				log.Error("failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	repo := s.repo.WithTx(tx)

	balance, err := repo.AddUserBalance(ctx, &payment.AddUserBalanceParams{
		TelegramUserID: telegramUserID,
		Amount:         tonAmount,
	})
	if err != nil {
		return nil, err
	}

	var metadataBytes []byte
	if metadata != nil {
		metadataBytes, err = json.Marshal(metadata)
		if err != nil {
			log.Error("failed to marshal metadata", zap.Error(err))
			// не возвращаем ошибку, просто логируем
		}
	}

	err = repo.CreateTransaction(ctx, &payment.CreateTransactionParams{
		TelegramUserID: telegramUserID,
		Amount:         tonAmount,
		Reason:         reason,
		Metadata:       metadataBytes,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error("failed to commit transaction", zap.Error(err))
		return nil, err
	}

	return balance, nil
}

func (s *Service) PreviewWithdraw(
	_ context.Context,
	gifts []*payment.GiftWithdrawRequest,
) (*payment.WithdrawOptions, error) {
	var totalStarsFee uint32
	var totalTonFee *tonamount.TonAmount
	giftFees := make([]*payment.GiftFee, 0, len(gifts))

	// Инициализируем totalTonFee нулевым значением
	zeroTon, err := tonamount.NewTonAmountFromString("0")
	if err != nil {
		return nil, err
	}
	totalTonFee = zeroTon

	for _, gift := range gifts {
		// Рассчитываем комиссию для каждого подарка индивидуально
		priceFloat, _ := gift.Price.Decimal().Float64()
		starsFee := calculateStarsCommission(priceFloat)
		tonFee, err := calculateTonCommission(starsFee)
		if err != nil {
			return nil, err
		}

		giftFee := &payment.GiftFee{
			GiftID:   gift.GiftID,
			StarsFee: starsFee,
			TonFee:   tonFee,
		}
		giftFees = append(giftFees, giftFee)

		// Суммируем общие комиссии
		totalStarsFee += starsFee
		totalTonFee = totalTonFee.Add(tonFee)
	}

	return &payment.WithdrawOptions{
		GiftFees:      giftFees,
		TotalStarsFee: totalStarsFee,
		TotalTonFee:   totalTonFee,
	}, nil
}

func calculateStarsCommission(giftTonPrice float64) uint32 {
	giftStars := giftTonPrice / tonPerStar
	raw := giftStars * commissionRate

	if raw < baseStarsCommission {
		raw = baseStarsCommission
	}
	if raw > maxStarsCommission {
		raw = maxStarsCommission
	}
	return uint32(math.Ceil(raw))
}

func calculateTonCommission(stars uint32) (*tonamount.TonAmount, error) {
	// 1) считаем «сырое» значение в TON
	raw := float64(stars) * tonPerStar
	// 2) округляем до 2 знаков
	//nolint:mnd // 2 decimal places
	raw = math.Round(raw*100) / 100
	// 3) создаём доменный TonAmount с проверкой и округлением внутри
	return tonamount.NewTonAmountFromFloat64(raw)
}

func (s *Service) RollbackWithdrawalCommission(
	ctx context.Context,
	telegramUserID int64,
	amount string,
	metadata payment.TransactionMetadata,
) error {
	log := s.log.With(zap.String("amount", amount), zap.Any("metadata", metadata))

	tonAmount, err := tonamount.NewTonAmountFromString(amount)
	if err != nil {
		log.Error("failed to parse amount", zap.Error(err))
		return err
	}

	_, err = s.AddUserBalance(
		ctx,
		telegramUserID,
		tonAmount.String(),
		payment.TransactionReasonRefund,
		&metadata,
	)
	if err != nil {
		log.Error("failed to add user balance", zap.Error(err))
		return err
	}

	return nil
}

func (s *Service) GetTransactionHistory(
	ctx context.Context,
	telegramUserID int64,
	pagination *shared.PageRequest,
) ([]*payment.Transaction, int64, error) {
	count, err := s.repo.GetUserTransactionsCount(ctx, telegramUserID)
	if err != nil {
		s.log.Error("failed to get user transactions count", zap.Error(err))
		return nil, 0, err
	}

	transactions, err := s.repo.GetUserTransactions(ctx, telegramUserID, pagination)
	if err != nil {
		s.log.Error("failed to get user transactions", zap.Error(err))
		return nil, 0, err
	}

	return transactions, count, nil
}
