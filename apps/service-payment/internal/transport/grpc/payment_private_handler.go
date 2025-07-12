package grpc

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
)

type PaymentPrivateHandler struct {
	paymentv1.UnimplementedPaymentPrivateServiceServer
	paymentService *payment.Service
}

func NewPaymentPrivateHandler(paymentService *payment.Service) paymentv1.PaymentPrivateServiceServer {
	return &PaymentPrivateHandler{paymentService: paymentService}
}

func (h *PaymentPrivateHandler) SpendUserBalance(ctx context.Context, req *paymentv1.SpendUserBalanceRequest) (*paymentv1.SpendUserBalanceResponse, error) {
	balance, err := h.paymentService.SpendUserBalance(
		ctx,
		req.TelegramUserId.Value,
		req.TonAmount.Value,
		TransactionReasonToDomain(req.GetReason()),
		TransactionMetadataToDomain(req.Metadata),
	)
	if err != nil {
		return nil, err
	}
	return &paymentv1.SpendUserBalanceResponse{
		NewAmount: &sharedv1.TonAmount{
			Value: balance.TonAmount,
		},
	}, nil
}

func (h *PaymentPrivateHandler) AddUserBalance(ctx context.Context, req *paymentv1.AddUserBalanceRequest) (*paymentv1.AddUserBalanceResponse, error) {
	balance, err := h.paymentService.AddUserBalance(
		ctx,
		req.TelegramUserId.Value,
		req.TonAmount.Value,
		TransactionReasonToDomain(req.GetReason()),
		TransactionMetadataToDomain(req.Metadata),
	)
	if err != nil {
		return nil, err
	}
	return &paymentv1.AddUserBalanceResponse{
		NewAmount: &sharedv1.TonAmount{
			Value: balance.TonAmount,
		},
	}, nil
}

func (h *PaymentPrivateHandler) GetUserBalance(ctx context.Context, req *paymentv1.GetUserBalanceRequest) (*paymentv1.GetUserBalanceResponse, error) {
	balance, err := h.paymentService.GetBalance(ctx, req.TelegramUserId.Value)
	if err != nil {
		return nil, err
	}
	return &paymentv1.GetUserBalanceResponse{
		Amount: &sharedv1.TonAmount{
			Value: balance.TonAmount,
		},
	}, nil
}

func (h *PaymentPrivateHandler) PreviewWithdraw(ctx context.Context, req *paymentv1.PrivatePreviewWithdrawRequest) (*paymentv1.PrivatePreviewWithdrawResponse, error) {
	tonAmount := req.GetTonAmount().GetValue()
	resp, err := h.paymentService.PreviewWithdraw(ctx, tonAmount)
	if err != nil {
		return nil, err
	}

	return &paymentv1.PrivatePreviewWithdrawResponse{
		TotalStarsFee: &sharedv1.StarsAmount{
			Value: resp.TotalStarsFee,
		},
		TotalTonFee: &sharedv1.TonAmount{
			Value: resp.TotalTonFee,
		},
	}, nil
}
