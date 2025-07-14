package grpc

import (
	"context"

	paymentdomain "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

type PaymentPrivateHandler struct {
	paymentv1.UnimplementedPaymentPrivateServiceServer

	paymentService *payment.Service
}

func NewPaymentPrivateHandler(paymentService *payment.Service) paymentv1.PaymentPrivateServiceServer {
	return &PaymentPrivateHandler{paymentService: paymentService}
}

func (h *PaymentPrivateHandler) SpendUserBalance(
	ctx context.Context,
	req *paymentv1.SpendUserBalanceRequest,
) (*paymentv1.SpendUserBalanceResponse, error) {
	reason, err := TransactionReasonToDomain(req.GetReason())
	if err != nil {
		return nil, err
	}
	metadata, err := TransactionMetadataToDomain(req.GetMetadata())
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
	reason, err := TransactionReasonToDomain(req.GetReason())
	if err != nil {
		return nil, err
	}
	metadata, err := TransactionMetadataToDomain(req.GetMetadata())
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
	gifts := make([]*paymentdomain.GiftWithdrawRequest, 0, len(req.GetGifts()))
	for _, giftReq := range req.GetGifts() {
		tonAmount, err := tonamount.NewTonAmountFromString(giftReq.GetPrice().GetValue())
		if err != nil {
			return nil, err
		}

		gift := &paymentdomain.GiftWithdrawRequest{
			GiftID: giftReq.GetGiftId().GetValue(),
			Price:  tonAmount,
		}
		gifts = append(gifts, gift)
	}

	resp, err := h.paymentService.PreviewWithdraw(ctx, gifts)
	if err != nil {
		return nil, err
	}

	// Конвертируем доменные объекты в protobuf ответ
	giftFees := make([]*paymentv1.GiftFee, 0, len(resp.GiftFees))
	for _, giftFee := range resp.GiftFees {
		giftFeeProto := &paymentv1.GiftFee{
			GiftId: &sharedv1.GiftId{
				Value: giftFee.GiftID,
			},
			StarsFee: &sharedv1.StarsAmount{
				Value: giftFee.StarsFee,
			},
			TonFee: &sharedv1.TonAmount{
				Value: giftFee.TonFee.String(),
			},
		}
		giftFees = append(giftFees, giftFeeProto)
	}

	return &paymentv1.PreviewWithdrawResponse{
		Fees: giftFees,
		TotalStarsFee: &sharedv1.StarsAmount{
			Value: resp.TotalStarsFee,
		},
		TotalTonFee: &sharedv1.TonAmount{
			Value: resp.TotalTonFee.String(),
		},
	}, nil
}
