package grpc

import (
	"net"

	"github.com/peterparker2005/giftduels/packages/configs"
)

func NewListener(cfg *configs.ServiceConfig) (net.Listener, error) {
	return net.Listen("tcp", cfg.Address())
}
