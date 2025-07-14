package clients

import (
	"context"
	"fmt"
	"io"

	"github.com/peterparker2005/giftduels/packages/configs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients stores all your grpc clients.
type Clients struct {
	Identity    *IdentityClient
	Gift        *GiftClient
	Payment     *PaymentClient
	TelegramBot *TelegramBotClient
	Duel        *DuelClient
}

// CreateDialOptions returns the base set of options for dial.
func CreateDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		// grpc.WithUnaryInterceptor(logger.UnaryClientRequestIDInterceptor()),
		// grpc.WithStreamInterceptor(logger.StreamClientRequestIDInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
}

// NewClients constructs all clients from the config addresses.
// If you need some custom options, you just add them to opts...
func NewClients(
	_ context.Context,
	cfg configs.GRPCConfig,
	opts ...grpc.DialOption,
) (*Clients, error) {
	dialOpts := CreateDialOptions()
	if len(opts) > 0 {
		dialOpts = append(dialOpts, opts...)
	}

	c := &Clients{}
	var err error

	if c.Identity, err = NewIdentityClient(cfg.Identity.Address(), dialOpts...); err != nil {
		return nil, fmt.Errorf("identity client: %w", err)
	}
	if c.Gift, err = NewGiftClient(cfg.Gift.Address(), dialOpts...); err != nil {
		c.Close()
		return nil, fmt.Errorf("gift client: %w", err)
	}
	if c.Payment, err = NewPaymentClient(cfg.Payment.Address(), dialOpts...); err != nil {
		c.Close()
		return nil, fmt.Errorf("payment client: %w", err)
	}
	if c.TelegramBot, err = NewTelegramBotClient(cfg.TelegramBot.Address(), dialOpts...); err != nil {
		c.Close()
		return nil, fmt.Errorf("telegram bot client: %w", err)
	}
	if c.Duel, err = NewDuelClient(cfg.Duel.Address(), dialOpts...); err != nil {
		c.Close()
		return nil, fmt.Errorf("duel client: %w", err)
	}
	return c, nil
}

// Close closes all clients.
func (c *Clients) Close() {
	for _, cl := range []io.Closer{
		c.Identity, c.Gift, c.Payment, c.TelegramBot, c.Duel,
	} {
		if cl != nil {
			_ = cl.Close()
		}
	}
}
