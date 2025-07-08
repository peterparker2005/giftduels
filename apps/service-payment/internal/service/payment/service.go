package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
)

type Service struct {
	repo payment.Repository
}

func NewService(repo payment.Repository) *Service {
	return &Service{
		repo: repo,
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

	return s.repo.AddUserBalance(ctx, addBalanceParams)
}

func (s *Service) GetBalance(ctx context.Context, telegramUserID int64) (*payment.Balance, error) {
	return s.repo.GetUserBalance(ctx, telegramUserID)
}
