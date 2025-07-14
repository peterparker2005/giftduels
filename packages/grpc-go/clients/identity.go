package clients

import (
	"fmt"

	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	"google.golang.org/grpc"
)

type IdentityClient struct {
	conn    *grpc.ClientConn
	Public  identityv1.IdentityPublicServiceClient
	Private identityv1.IdentityPrivateServiceClient
}

func NewIdentityClient(address string, opts ...grpc.DialOption) (*IdentityClient, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial identity service %s: %w", address, err)
	}
	return &IdentityClient{
		conn:    conn,
		Public:  identityv1.NewIdentityPublicServiceClient(conn),
		Private: identityv1.NewIdentityPrivateServiceClient(conn),
	}, nil
}

func (c *IdentityClient) Close() error {
	return c.conn.Close()
}
