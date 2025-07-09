package clients

import (
	"context"
	"fmt"
	"io"

	"github.com/peterparker2005/giftduels/packages/configs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients хранит все ваши grpc-клиенты
type Clients struct {
	Identity *IdentityClient
	Gift     *GiftClient
}

// CreateDialOptions возвращает базовый набор опций для dial
func CreateDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		// grpc.WithUnaryInterceptor(logger.UnaryClientRequestIDInterceptor()),
		// grpc.WithStreamInterceptor(logger.StreamClientRequestIDInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
}

// NewClients сконструирует сразу всех клиентов по адресам из конфига.
// Если вам нужны какие-то кастомные опции — вы просто допихиваете их в opts…
func NewClients(ctx context.Context, cfg configs.GRPCConfig, opts ...grpc.DialOption) (*Clients, error) {
	// стартовый набор опций (интерцепторы + insecure)
	dialOpts := CreateDialOptions()
	if len(opts) > 0 {
		dialOpts = append(dialOpts, opts...)
	}

	c := &Clients{}
	var err error

	if c.Identity, err = NewIdentityClient(ctx, cfg.Identity.Address(), dialOpts...); err != nil {
		return nil, fmt.Errorf("identity client: %w", err)
	}
	if c.Gift, err = NewGiftClient(ctx, cfg.Gift.Address(), dialOpts...); err != nil {
		c.Close()
		return nil, fmt.Errorf("user client: %w", err)
	}

	return c, nil
}

// Close аккуратно закроет все клиенты
func (c *Clients) Close() {
	for _, cl := range []io.Closer{
		c.Identity, c.Gift,
	} {
		if cl != nil {
			_ = cl.Close()
		}
	}
}
