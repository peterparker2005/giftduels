package grpc

import (
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
)

type PaymentPrivateHandler struct {
	paymentv1.UnimplementedPaymentPrivateServiceServer
}

func NewPaymentPrivateHandler() paymentv1.PaymentPrivateServiceServer {
	return &PaymentPrivateHandler{}
}
