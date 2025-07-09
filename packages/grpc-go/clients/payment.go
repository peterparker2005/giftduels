package clients

import (
	"context"
	"fmt"

	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	"google.golang.org/grpc"
)

type PaymentClient struct {
	conn    *grpc.ClientConn
	Public  paymentv1.PaymentPublicServiceClient
	Private paymentv1.PaymentPrivateServiceClient
}

func NewPaymentClient(ctx context.Context, address string, opts ...grpc.DialOption) (*PaymentClient, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial gift service %s: %w", address, err)
	}
	return &PaymentClient{
		conn:    conn,
		Public:  paymentv1.NewPaymentPublicServiceClient(conn),
		Private: paymentv1.NewPaymentPrivateServiceClient(conn),
	}, nil
}

func (c *PaymentClient) Close() error {
	return c.conn.Close()
}
