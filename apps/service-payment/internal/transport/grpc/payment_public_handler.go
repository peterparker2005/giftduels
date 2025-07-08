package grpc

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentPublicHandler struct {
	paymentv1.UnimplementedPaymentPublicServiceServer
	service *payment.Service
}

func NewPaymentPublicHandler(service *payment.Service) paymentv1.PaymentPublicServiceServer {
	return &PaymentPublicHandler{
		service: service,
	}
}

func (h *PaymentPublicHandler) GetBalance(ctx context.Context, req *emptypb.Empty) (*paymentv1.GetBalanceResponse, error) {
	telegramUserID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}
	balance, err := h.service.GetBalance(ctx, telegramUserID)
	if err != nil {
		return nil, err
	}
	return &paymentv1.GetBalanceResponse{
		Balance: &paymentv1.UserBalanceView{
			TonAmount: &sharedv1.TonAmount{
				Value: balance.TonAmount,
			},
		},
	}, nil
}

func (h *PaymentPublicHandler) GetWithdrawOptions(ctx context.Context, req *paymentv1.GetWithdrawOptionsRequest) (*paymentv1.GetWithdrawOptionsResponse, error) {
	return &paymentv1.GetWithdrawOptionsResponse{}, nil
}

func (h *PaymentPublicHandler) DepositTon(ctx context.Context, req *paymentv1.DepositTonRequest) (*paymentv1.DepositTonResponse, error) {
	// telegramUserID, err := authctx.TelegramUserID(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// tonAmount := req.GetTonAmount().GetValue()
	return &paymentv1.DepositTonResponse{
		DepositId:     "",
		NanoTonAmount: 0,
		Payload:       "",
	}, nil
}
