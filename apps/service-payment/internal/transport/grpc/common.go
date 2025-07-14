package grpc

import (
	"context"

	paymentdomain "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

func handlePreviewWithdraw(
	ctx context.Context,
	giftsReq []*paymentv1.GiftWithdrawRequest,
	previewFn func(ctx context.Context, gifts []*paymentdomain.GiftWithdrawRequest) (*paymentdomain.WithdrawOptions, error),
) (*paymentv1.PreviewWithdrawResponse, error) {
	gifts := make([]*paymentdomain.GiftWithdrawRequest, 0, len(giftsReq))
	for _, giftReq := range giftsReq {
		tonAmount, err := tonamount.NewTonAmountFromString(giftReq.GetPrice().GetValue())
		if err != nil {
			return nil, err
		}
		gifts = append(gifts, &paymentdomain.GiftWithdrawRequest{
			GiftID: giftReq.GetGiftId().GetValue(),
			Price:  tonAmount,
		})
	}

	resp, err := previewFn(ctx, gifts)
	if err != nil {
		return nil, err
	}

	giftFees := make([]*paymentv1.GiftFee, 0, len(resp.GiftFees))
	for _, fee := range resp.GiftFees {
		giftFees = append(giftFees, &paymentv1.GiftFee{
			GiftId:   &sharedv1.GiftId{Value: fee.GiftID},
			StarsFee: &sharedv1.StarsAmount{Value: fee.StarsFee},
			TonFee:   &sharedv1.TonAmount{Value: fee.TonFee.String()},
		})
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
