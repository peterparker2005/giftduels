package clients

import (
	"fmt"

	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	"google.golang.org/grpc"
)

type DuelClient struct {
	conn    *grpc.ClientConn
	Private duelv1.DuelPrivateServiceClient
	Public  duelv1.DuelPublicServiceClient
}

func NewDuelClient(address string, opts ...grpc.DialOption) (*DuelClient, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial duel service %s: %w", address, err)
	}
	return &DuelClient{
		conn:    conn,
		Private: duelv1.NewDuelPrivateServiceClient(conn),
		Public:  duelv1.NewDuelPublicServiceClient(conn),
	}, nil
}

func (c *DuelClient) Close() error {
	return c.conn.Close()
}
