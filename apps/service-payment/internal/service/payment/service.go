package payment

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
)

const (
	// Минимальное базовое кол-во звёзд
	baseStarsCommission = 25.0

	// Максимальное кол-во звёзд
	maxStarsCommission = 250.0

	// Наша маржа 10%
	profitMargin = 0.10

	// Сколько TON стоит одна звезда
	tonPerStar = 0.2678 / 50.0

	// Какой процент от стоимости подарка мы берём в виде комиссии (в звёздах)
	// Например, 5% от стоимости подарка
	commissionRate = 0.05
)

type Service struct {
	repo              payment.Repository
	giftPrivateClient giftv1.GiftPrivateServiceClient
}

func NewService(repo payment.Repository, clients *clients.Clients) *Service {
	return &Service{
		repo:              repo,
		giftPrivateClient: clients.Gift.Private,
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

func (s *Service) PreviewWithdraw(ctx context.Context, giftIDs []string) (*payment.WithdrawOptions, error) {
	giftIds := make([]*sharedv1.GiftId, len(giftIDs))
	for i, giftID := range giftIDs {
		giftIds[i] = &sharedv1.GiftId{
			Value: giftID,
		}
	}
	gifts, err := s.giftPrivateClient.PrivateGetGifts(ctx, &giftv1.PrivateGetGiftsRequest{
		GiftIds: giftIds,
	})
	if err != nil {
		return nil, err
	}

	fees := make([]payment.GiftFee, len(gifts.Gifts))
	for i, gift := range gifts.Gifts {
		giftTonPrice := gift.GetOriginalPrice().GetValue()
		starsCost := calculateStarsCommission(giftTonPrice)
		tonFee := calculateTonCommission(starsCost)

		fees[i] = payment.GiftFee{
			GiftID:   gift.GetGiftId().GetValue(),
			StarsFee: uint32(starsCost),
			TonFee:   tonFee,
		}
	}

	totalStarsFee := uint32(0)
	totalTonFee := 0.0
	for _, fee := range fees {
		totalStarsFee += fee.StarsFee
		totalTonFee += fee.TonFee
	}

	totalTonFee = math.Round(totalTonFee*100) / 100

	return &payment.WithdrawOptions{
		Fees:          fees,
		TotalStarsFee: totalStarsFee,
		TotalTonFee:   totalTonFee,
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

	// 4) добавляем маржу
	raw = raw * (1 + profitMargin)

	// 5) округляем вверх до целого
	return math.Ceil(raw)
}

func calculateTonCommission(stars float64) float64 {
	ton := stars * tonPerStar
	return math.Round(ton*100) / 100
}
