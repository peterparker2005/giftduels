package grpchandlers

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/adapter/proto"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
)

type PaymentPrivateHandler struct {
	paymentv1.UnimplementedPaymentPrivateServiceServer

	paymentService *payment.Service
}

func NewPaymentPrivateHandler(
	paymentService *payment.Service,
) paymentv1.PaymentPrivateServiceServer {
	return &PaymentPrivateHandler{paymentService: paymentService}
}

func (h *PaymentPrivateHandler) SpendUserBalance(
	ctx context.Context,
	req *paymentv1.SpendUserBalanceRequest,
) (*paymentv1.SpendUserBalanceResponse, error) {
	reason, err := proto.TransactionReasonToDomain(req.GetReason())
	if err != nil {
		return nil, err
	}
	metadata, err := proto.TransactionMetadataToDomain(req.GetMetadata())
	if err != nil {
		return nil, err
	}
	balance, err := h.paymentService.SpendUserBalance(
		ctx,
		req.GetTelegramUserId().GetValue(),
		req.GetTonAmount().GetValue(),
		reason,
		metadata,
	)
	if err != nil {
		return nil, err
	}
	return &paymentv1.SpendUserBalanceResponse{
		NewAmount: &sharedv1.TonAmount{
			Value: balance.TonAmount.String(),
		},
	}, nil
}

func (h *PaymentPrivateHandler) AddUserBalance(
	ctx context.Context,
	req *paymentv1.AddUserBalanceRequest,
) (*paymentv1.AddUserBalanceResponse, error) {
	reason, err := proto.TransactionReasonToDomain(req.GetReason())
	if err != nil {
		return nil, err
	}
	metadata, err := proto.TransactionMetadataToDomain(req.GetMetadata())
	if err != nil {
		return nil, err
	}
	balance, err := h.paymentService.AddUserBalance(
		ctx,
		req.GetTelegramUserId().GetValue(),
		req.GetTonAmount().GetValue(),
		reason,
		metadata,
	)
	if err != nil {
		return nil, err
	}
	return &paymentv1.AddUserBalanceResponse{
		NewAmount: &sharedv1.TonAmount{
			Value: balance.TonAmount.String(),
		},
	}, nil
}

func (h *PaymentPrivateHandler) GetUserBalance(
	ctx context.Context,
	req *paymentv1.GetUserBalanceRequest,
) (*paymentv1.GetUserBalanceResponse, error) {
	balance, err := h.paymentService.GetBalance(ctx, req.GetTelegramUserId().GetValue())
	if err != nil {
		return nil, err
	}
	return &paymentv1.GetUserBalanceResponse{
		Amount: &sharedv1.TonAmount{
			Value: balance.TonAmount.String(),
		},
	}, nil
}

func (h *PaymentPrivateHandler) PreviewWithdraw(
	ctx context.Context,
	req *paymentv1.PreviewWithdrawRequest,
) (*paymentv1.PreviewWithdrawResponse, error) {
	return handlePreviewWithdraw(ctx, req.GetGifts(), h.paymentService.PreviewWithdraw)
}
