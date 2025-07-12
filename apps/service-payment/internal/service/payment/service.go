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
	"go.uber.org/zap"
)

const (
	// Минимальное базовое кол-во звёзд
	baseStarsCommission = 25.0

	// Максимальное кол-во звёзд
	maxStarsCommission = 250.0

	// Сколько TON стоит одна звезда
	tonPerStar = 0.2678 / 50.0

	// Какой процент от стоимости подарка мы берём в виде комиссии (в звёздах)
	// Например, 5% от стоимости подарка
	commissionRate = 0.15
)

type Service struct {
	log     *logger.Logger
	repo    payment.Repository
	tonRepo ton.DepositRepository
	txMgr   pg.TxManager
}

func NewService(repo payment.Repository, tonRepo ton.DepositRepository, log *logger.Logger, txMgr pg.TxManager) *Service {
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
	tonAmount float64,
) (*ton.Deposit, error) {
	rawPayload := uuid.New().String()
	nanoAmount := int64(tonAmount * 1e9)
	expiresAt := time.Now().Add(time.Hour)

	params := &ton.CreateDepositParams{
		TelegramUserID: telegramUserID,
		AmountNano:     nanoAmount,
		Payload:        rawPayload, // просто UUID
		ExpiresAt:      expiresAt,
	}
	return s.tonRepo.CreateDeposit(ctx, params)
}

func (s *Service) ProcessDepositTransaction(ctx context.Context, payload, txHash string, txLt, amountNano int64) error {
	deposit, err := s.tonRepo.GetDepositByPayload(ctx, payload)
	if err != nil {
		return err
	}

	if deposit.Status != ton.DepositStatusPending {
		// Or log and ignore
		return nil
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

	err = repo.CreateTransaction(ctx, &payment.CreateTransactionParams{
		TelegramUserID: deposit.TelegramUserID,
		Amount:         float64(amountNano) / 1e9,
		Reason:         payment.TransactionReasonDeposit,
	})
	if err != nil {
		return err
	}

	addBalanceParams := &payment.AddUserBalanceParams{
		TelegramUserID: deposit.TelegramUserID,
		Amount:         float64(amountNano) / 1e9,
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

func (s *Service) SpendUserBalance(ctx context.Context, telegramUserID int64, amount float64, reason payment.TransactionReason, metadata *payment.TransactionMetadata) (*payment.Balance, error) {
	log := s.log.With(zap.Int64("telegram_user_id", telegramUserID), zap.Float64("amount", amount), zap.String("reason", string(reason)))

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

	balance, err := repo.SpendUserBalance(ctx, &payment.SpendUserBalanceParams{
		TelegramUserID: telegramUserID,
		Amount:         amount,
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
		Amount:         -amount,
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

func (s *Service) AddUserBalance(ctx context.Context, telegramUserID int64, amount float64, reason payment.TransactionReason, metadata *payment.TransactionMetadata) (*payment.Balance, error) {
	log := s.log.With(zap.Int64("telegram_user_id", telegramUserID), zap.Float64("amount", amount))

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
		Amount:         amount,
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
		Amount:         amount,
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

func (s *Service) PreviewWithdraw(ctx context.Context, tonAmount float64) (*payment.WithdrawOptions, error) {
	starsCost := calculateStarsCommission(tonAmount)
	tonFee := calculateTonCommission(starsCost)

	totalStarsFee := uint32(starsCost)

	return &payment.WithdrawOptions{
		TotalStarsFee: totalStarsFee,
		TotalTonFee:   tonFee,
	}, nil
}

func calculateStarsCommission(giftTonPrice float64) float64 {
	// 1) переводим цену подарка в эквивалент звёзд
	giftStars := giftTonPrice / tonPerStar

	// 2) базовая комиссия — процент от giftStars
	raw := giftStars * commissionRate

	// 3) гарантируем минимум base и максимум max
	if raw < baseStarsCommission {
		raw = baseStarsCommission
	}
	if raw > maxStarsCommission {
		raw = maxStarsCommission
	}

	// 4) округляем вверх до целого
	return math.Ceil(raw)
}

func calculateTonCommission(stars float64) float64 {
	ton := stars * tonPerStar
	return math.Round(ton*100) / 100
}

func (s *Service) RollbackWithdrawalCommission(ctx context.Context, telegramUserID int64, amount float64, metadata payment.TransactionMetadata) error {
	log := s.log.With(zap.Float64("amount", amount), zap.Any("metadata", metadata))

	_, err := s.AddUserBalance(ctx, telegramUserID, amount, payment.TransactionReasonRefund, &metadata)
	if err != nil {
		log.Error("failed to create transaction", zap.Error(err))
		return err
	}

	return nil
}

func (s *Service) GetTransactionHistory(ctx context.Context, telegramUserID int64, pagination *shared.PageRequest) ([]*payment.Transaction, int64, error) {
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
