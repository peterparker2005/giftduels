package tonworker

import (
	"context"
	"time"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/ton"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	"github.com/peterparker2005/giftduels/apps/service-payment/pkg/boc"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

type Processor struct {
	api             ton.API
	depositRepo     ton.DepositRepository
	paymentService  *payment.Service
	treasuryAddress string
	cancel          context.CancelFunc
	logger          *logger.Logger
}

func NewProcessor(
	api ton.API,
	depositRepo ton.DepositRepository,
	paymentService *payment.Service,
	cfg *config.Config,
	logger *logger.Logger,
) *Processor {
	return &Processor{
		api:             api,
		depositRepo:     depositRepo,
		paymentService:  paymentService,
		treasuryAddress: cfg.Ton.WalletAddress,
		logger:          logger,
	}
}

func (p *Processor) Start() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	go func() {
		p.run(ctx)
	}()
}

func (p *Processor) Stop(_ context.Context) error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *Processor) run(ctx context.Context) {
	const retryDelay = 5 * time.Second

	for {
		if ctx.Err() != nil {
			p.logger.Info("🛑 TON Worker stopping")
			return
		}

		lastLT, err := p.depositRepo.GetCursor(ctx, "testnet", p.treasuryAddress)
		if err != nil {
			p.logger.Error("failed to get cursor", zap.Error(err))
			time.Sleep(retryDelay)
			continue
		}

		p.subscribeAndProcess(ctx, lastLT, retryDelay)
	}
}

func (p *Processor) subscribeAndProcess(
	ctx context.Context,
	fromLT uint64,
	retryDelay time.Duration,
) {
	p.logger.Info("🔍 TON Worker", zap.Uint64("fromLT", fromLT))

	txCh := make(chan ton.Transaction)
	if err := p.api.SubscribeTransactions(ctx, p.treasuryAddress, fromLT, txCh); err != nil {
		p.logger.Error("subscribe error", zap.Error(err))
		time.Sleep(retryDelay)
		return
	}
	p.logger.Info("🚀 TON Worker started")

	p.readLoop(ctx, txCh, retryDelay)
}

func (p *Processor) readLoop(
	ctx context.Context,
	txCh chan ton.Transaction,
	retryDelay time.Duration,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case tx, ok := <-txCh:
			if !ok {
				p.logger.Warn("txCh closed, will retry subscription")
				time.Sleep(retryDelay)
				return
			}
			p.handleTx(ctx, tx)
			p.saveCursor(ctx, tx.LastLT)
		}
	}
}

func (p *Processor) handleTx(ctx context.Context, tx ton.Transaction) {
	p.logger.Info("🔔 Received",
		zap.String("amount", tx.Amount.String()),
		zap.String("currency", tx.Currency),
		zap.String("sender", tx.Sender),
		zap.String("payload", tx.Payload),
	)

	// ранний выход вместо вложенных if
	if tx.Payload == "" || tx.Currency != "TON" {
		return
	}

	p.processDeposit(ctx, tx)
}

func (p *Processor) processDeposit(ctx context.Context, tx ton.Transaction) {
	// 1) парсим nano
	nano, err := tx.Amount.ToNano()
	if err != nil {
		p.logger.Warn("invalid amount", zap.String("amount", tx.Amount.String()), zap.Error(err))
		return
	}

	// 2) декодируем BOC
	original, err := boc.DecodeStringFromBOC(tx.Payload)
	if err != nil {
		p.logger.Warn("failed to decode BOC", zap.String("payload", tx.Payload), zap.Error(err))
		return
	}
	p.logger.Info("🔓 Decoded BOC", zap.String("original", original))

	// 3) обрабатываем в сервисе
	if err = p.paymentService.ProcessDepositTransaction(ctx, original, "", tx.LastLT, tx.Amount); err != nil {
		p.logger.Warn("failed to process deposit", zap.String("payload", original), zap.Error(err))
	} else {
		p.logger.Info("✅ Deposit processed", zap.String("payload", original), zap.Uint64("amount", nano))
	}
}

func (p *Processor) saveCursor(ctx context.Context, lastLT uint64) {
	if err := p.depositRepo.UpsertCursor(ctx, "testnet", p.treasuryAddress, lastLT); err != nil {
		p.logger.Warn("failed to save cursor", zap.Error(err))
	}
}
