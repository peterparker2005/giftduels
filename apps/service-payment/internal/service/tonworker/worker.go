package tonworker

import (
	"context"
	"time"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/ton"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

type Processor struct {
	api             ton.TonAPI
	cursorRepo      ton.CursorRepository
	treasuryAddress string
	cancel          context.CancelFunc
	logger          *logger.Logger
}

func NewProcessor(
	api ton.TonAPI,
	cursorRepo ton.CursorRepository,
	cfg *config.Config,
	logger *logger.Logger,
) *Processor {
	return &Processor{
		api:             api,
		cursorRepo:      cursorRepo,
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

func (p *Processor) Stop(ctx context.Context) error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *Processor) run(ctx context.Context) {
	const retryDelay = 5 * time.Second

	for {
		// 1) Прекращаем работу, если контекст отменён
		select {
		case <-ctx.Done():
			p.logger.Info("🛑 TON Worker stopping")
			return
		default:
		}

		// 2) Читаем курсор из БД
		lastLT, err := p.cursorRepo.Get(ctx, "testnet", p.treasuryAddress)
		if err != nil {
			p.logger.Error("failed to get cursor", zap.Error(err))
			time.Sleep(retryDelay)
			continue
		}
		p.logger.Info("🔍 TON Worker", zap.Uint64("fromLT", lastLT))

		// 3) Подписываемся и обрабатываем канал
		txCh := make(chan ton.Transaction)
		if err := p.api.SubscribeTransactions(ctx, p.treasuryAddress, lastLT, txCh); err != nil {
			p.logger.Error("subscribe error", zap.Error(err))
			time.Sleep(retryDelay)
			continue
		}
		p.logger.Info("🚀 TON Worker started")

		// 4) Читаем из канала, пока он не закроется или не отменится контекст
		for {
			select {
			case <-ctx.Done():
				return
			case tx, ok := <-txCh:
				if !ok {
					p.logger.Warn("⚠️ txCh closed, will retry subscription")
					time.Sleep(retryDelay)
					// выйти из внутреннего цикла, чтобы заново подписаться
					break
				}
				// 5) Обрабатываем транзакцию и сохраняем курсор
				p.logger.Info("🔔 Received",
					zap.String("amount", tx.Amount),
					zap.String("currency", tx.Currency),
					zap.String("sender", tx.Sender),
				)
				if err := p.cursorRepo.Upsert(ctx, "testnet", p.treasuryAddress, tx.LastLT); err != nil {
					p.logger.Warn("failed to save cursor", zap.Error(err))
				}
			}
			// если канал закрылся — выйти наружу и перезапустить подписку
			if ctx.Err() != nil {
				return
			}
			select {
			case <-txCh:
			default:
			}
		}
	}
}
