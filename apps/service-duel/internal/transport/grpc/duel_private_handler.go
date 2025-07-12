package grpc

import duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"

type DuelPrivateHandler struct {
	duelv1.UnimplementedDuelPrivateServiceServer
}

func NewDuelPrivateHandler() duelv1.DuelPrivateServiceServer {
	return &DuelPrivateHandler{}
}
