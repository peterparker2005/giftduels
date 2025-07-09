package clients

import (
	"context"
	"fmt"

	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"google.golang.org/grpc"
)

type GiftClient struct {
	conn    *grpc.ClientConn
	Public  giftv1.GiftPrivateServiceClient
	Private giftv1.GiftPrivateServiceClient
}

func NewGiftClient(ctx context.Context, address string, opts ...grpc.DialOption) (*GiftClient, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial gift service %s: %w", address, err)
	}
	return &GiftClient{
		conn:    conn,
		Public:  giftv1.NewGiftPrivateServiceClient(conn),
		Private: giftv1.NewGiftPrivateServiceClient(conn),
	}, nil
}

func (c *GiftClient) Close() error {
	return c.conn.Close()
}
