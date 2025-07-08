package ton

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	domain "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/ton"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"go.uber.org/zap"
)

type adapter struct {
	api    ton.APIClientWrapped
	cfg    *liteclient.GlobalConfig
	logger *logger.Logger
}

const (
	testnetURL = "https://ton-blockchain.github.io/testnet-global.config.json"
	mainnetURL = "https://ton-blockchain.github.io/global.config.json"
)

// NewTonAPI создаёт адаптер TonAPI
func NewTonAPI(appCfg *config.Config, logger *logger.Logger) (domain.TonAPI, error) {
	url := testnetURL
	if appCfg.Ton.Network == config.TonNetworkMainnet {
		url = mainnetURL
	}
	logger.Info("🔧 TON adapter: network=%s, configURL=%s", zap.String("network", appCfg.Ton.Network.String()), zap.String("configURL", url))
	client := liteclient.NewConnectionPool()
	cfg, err := liteclient.GetConfigFromUrl(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	if err = client.AddConnectionsFromConfig(context.Background(), cfg); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	raw := ton.NewAPIClient(client, ton.ProofCheckPolicyFast)
	raw.SetTrustedBlockFromConfig(cfg)
	api := raw.WithRetry()
	_, err = api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get masterchain info: %w", err)
	}
	return &adapter{api: api, cfg: cfg, logger: logger}, nil
}

func (a *adapter) CurrentMasterchainInfo(ctx context.Context) (domain.MasterchainInfo, error) {
	m, err := a.api.CurrentMasterchainInfo(ctx)
	if err != nil {
		return domain.MasterchainInfo{}, err
	}
	return domain.MasterchainInfo{SeqNo: m.SeqNo}, nil
}

func (a *adapter) GetAccountLastLT(ctx context.Context, addrStr string) (uint64, error) {
	mci, err := a.api.CurrentMasterchainInfo(ctx)
	if err != nil {
		return 0, err
	}
	acc, err := a.api.GetAccount(ctx, mci, address.MustParseAddr(addrStr))
	if err != nil {
		return 0, err
	}
	return acc.LastTxLT, nil
}

func (a *adapter) SubscribeTransactions(ctx context.Context, addrStr string, fromLT uint64, out chan<- domain.Transaction) error {
	addr := address.MustParseAddr(addrStr)
	// запускаем в фоне
	// rawCh – канал для низкоуровневых tlb.Transaction
	rawCh := make(chan *tlb.Transaction)

	// 1) запустить саму подписку (блокирующий вызов) в отдельной горутине
	go func() {
		a.logger.Info("📡 TON adapter: subscribe transactions", zap.String("addr", addrStr), zap.Uint64("fromLT", fromLT))
		a.api.SubscribeOnTransactions(ctx, addr, fromLT, rawCh)
		a.logger.Info("⚠️ SubscribeOnTransactions finished, closing rawCh")
		close(rawCh)
	}()

	// 2) параллельно читать из rawCh, конвертить и форвардить в out
	go func() {
		for raw := range rawCh {
			if raw.IO.In == nil || raw.IO.In.MsgType != tlb.MsgTypeInternal {
				continue
			}
			ti := raw.IO.In.AsInternal()
			sender := ti.SrcAddr.String()
			amountStr := ti.Amount.Nano().String()
			currency := "TON"

			// Extract payload from transaction body
			payload := ""
			if ti.Body != nil {
				// Convert entire body to BOC base64 (this is what tonworker expects)
				bocBytes := ti.Body.ToBOC()
				if len(bocBytes) > 2 { // Skip empty BOC (usually 2 bytes for empty cell)
					payload = a.encodeBOCAsBase64(bocBytes)
					a.logger.Debug("📦 Extracted BOC payload",
						zap.String("payload", payload),
						zap.Int("bocLength", len(bocBytes)))
				}
			}

			out <- domain.Transaction{
				Sender:   sender,
				Amount:   amountStr,
				Currency: currency,
				Payload:  payload,
				LastLT:   raw.LT,
			}
		}
		a.logger.Info("✅ Forwarding loop ended, out channel will not get more messages")
	}()
	return nil
}

// encodeBOCAsBase64 encodes BOC bytes as base64 URL encoding
func (a *adapter) encodeBOCAsBase64(bocBytes []byte) string {
	return base64.URLEncoding.EncodeToString(bocBytes)
}
