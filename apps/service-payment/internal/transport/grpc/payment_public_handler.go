package grpc

import (
	"context"

	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentPublicHandler struct {
	paymentv1.UnimplementedPaymentPublicServiceServer
}

func NewPaymentPublicHandler() *PaymentPublicHandler {
	return &PaymentPublicHandler{}
}

func (h *PaymentPublicHandler) GetBalance(ctx context.Context, req *emptypb.Empty) (*paymentv1.GetBalanceResponse, error) {
	return &paymentv1.GetBalanceResponse{}, nil
}

func (h *PaymentPublicHandler) GetWithdrawOptions(ctx context.Context, req *paymentv1.GetWithdrawOptionsRequest) (*paymentv1.GetWithdrawOptionsResponse, error) {
	return &paymentv1.GetWithdrawOptionsResponse{}, nil
}

func (h *PaymentPublicHandler) DepositTon(ctx context.Context, req *paymentv1.DepositTonRequest) (*paymentv1.DepositTonResponse, error) {
	return &paymentv1.DepositTonResponse{}, nil
}
