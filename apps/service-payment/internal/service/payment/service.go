package payment

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/packages/logger-go"
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
	log   *logger.Logger
	repo  payment.Repository
	txMgr pg.TxManager
}

func NewService(repo payment.Repository, log *logger.Logger, txMgr pg.TxManager) *Service {
	return &Service{
		log:   log,
		repo:  repo,
		txMgr: txMgr,
	}
}

func (s *Service) CreateDeposit(
	ctx context.Context,
	telegramUserID int64,
	tonAmount float64,
) (*payment.Deposit, error) {
	rawPayload := uuid.New().String()
	nanoAmount := int64(tonAmount * 1e9)
	expiresAt := time.Now().Add(time.Hour)

	params := &payment.CreateDepositParams{
		TelegramUserID: telegramUserID,
		AmountNano:     nanoAmount,
		Payload:        rawPayload, // просто UUID
		ExpiresAt:      expiresAt,
	}
	return s.repo.CreateDeposit(ctx, params)
}

func (s *Service) ProcessDepositTransaction(ctx context.Context, payload, txHash string, txLt, amountNano int64) error {
	deposit, err := s.repo.GetDepositByPayload(ctx, payload)
	if err != nil {
		return err
	}

	if deposit.Status != payment.DepositStatusPending {
		// Or log and ignore
		return nil
	}

	if deposit.AmountNano > amountNano {
		// Handle partial payment, for now, we ignore
		return nil
	}

	_, err = s.repo.SetDepositTransaction(ctx, &payment.SetDepositTransactionParams{
		ID:     deposit.ID.String(),
		TxHash: txHash,
		TxLt:   txLt,
	})
	if err != nil {
		return err
	}

	addBalanceParams := &payment.AddUserBalanceParams{
		TelegramUserID: deposit.TelegramUserID,
		Amount:         float64(amountNano) / 1e9,
	}

	_, err = s.repo.AddUserBalance(ctx, addBalanceParams)
	return err
}

func (s *Service) GetBalance(ctx context.Context, telegramUserID int64) (*payment.Balance, error) {
	return s.repo.GetUserBalance(ctx, telegramUserID)
}

func (s *Service) SpendUserBalance(ctx context.Context, telegramUserID int64, amount float64) (*payment.Balance, error) {
	balance, err := s.repo.SpendUserBalance(ctx, &payment.SpendUserBalanceParams{
		TelegramUserID: telegramUserID,
		Amount:         amount,
	})
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *Service) AddUserBalance(ctx context.Context, telegramUserID int64, amount float64) (*payment.Balance, error) {
	balance, err := s.repo.AddUserBalance(ctx, &payment.AddUserBalanceParams{
		TelegramUserID: telegramUserID,
		Amount:         amount,
	})
	if err != nil {
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

func (s *Service) SpendWithdrawalCommission(ctx context.Context, telegramUserID int64, tonAmount float64) (*payment.Balance, float64, error) {
	log := s.log.With(zap.Float64("ton_amount", tonAmount))
	preview, err := s.PreviewWithdraw(ctx, tonAmount)
	if err != nil {
		return nil, 0, err
	}

	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		return nil, 0, err
	}

	repo := s.repo.WithTx(tx)

	var pErr error
	defer func() {
		if pErr != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error("failed to rollback commission spend transaction", zap.Error(rbErr))
			}
		}
	}()

	amount := preview.TotalTonFee

	log.Info("spending withdrawal commission", zap.Float64("amount", amount))

	balance, err := repo.SpendUserBalance(ctx, &payment.SpendUserBalanceParams{
		TelegramUserID: telegramUserID,
		Amount:         amount,
	})
	if err != nil {
		pErr = err
		return nil, 0, err
	}

	err = repo.CreateTransaction(ctx, &payment.CreateTransactionParams{
		TelegramUserID: telegramUserID,
		Amount:         amount,
		Reason:         payment.TransactionReasonWithdraw,
	})
	if err != nil {
		pErr = err
		return nil, 0, err
	}

	pErr = tx.Commit(ctx)
	if pErr != nil {
		return nil, 0, pErr
	}

	return balance, amount, nil
}

func (s *Service) RollbackWithdrawalCommission(ctx context.Context, telegramUserID int64, amount float64) error {
	log := s.log.With(zap.Float64("amount", amount))
	tx, err := s.txMgr.BeginTx(ctx)
	if err != nil {
		log.Error("failed to begin transaction", zap.Error(err))
		return err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				log.Error("Failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	repo := s.repo.WithTx(tx)

	_, err = repo.AddUserBalance(ctx, &payment.AddUserBalanceParams{
		TelegramUserID: telegramUserID,
		Amount:         amount,
	})
	if err != nil {
		log.Error("failed to create transaction", zap.Error(err))
		return err
	}

	err = repo.CreateTransaction(ctx, &payment.CreateTransactionParams{
		TelegramUserID: telegramUserID,
		Amount:         amount,
		Reason:         payment.TransactionReasonRefund,
	})
	if err != nil {
		log.Error("failed to create transaction", zap.Error(err))
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	return nil
}
