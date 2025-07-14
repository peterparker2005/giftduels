package grpc

import (
	"errors"
	"fmt"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TransactionReasonToDomain(
	reason paymentv1.TransactionReason,
) (payment.TransactionReason, error) {
	switch reason {
	case paymentv1.TransactionReason_TRANSACTION_REASON_WITHDRAW:
		return payment.TransactionReasonWithdraw, nil
	case paymentv1.TransactionReason_TRANSACTION_REASON_REFUND:
		return payment.TransactionReasonRefund, nil
	case paymentv1.TransactionReason_TRANSACTION_REASON_DEPOSIT:
		return payment.TransactionReasonDeposit, nil
	case paymentv1.TransactionReason_TRANSACTION_REASON_UNSPECIFIED:
		return "", errors.New("transaction reason is unspecified")
	default:
		return "", fmt.Errorf("unknown transaction reason: %v", reason)
	}
}

func TransactionReasonToProto(
	reason payment.TransactionReason,
) (paymentv1.TransactionReason, error) {
	switch reason {
	case payment.TransactionReasonWithdraw:
		return paymentv1.TransactionReason_TRANSACTION_REASON_WITHDRAW, nil
	case payment.TransactionReasonRefund:
		return paymentv1.TransactionReason_TRANSACTION_REASON_REFUND, nil
	case payment.TransactionReasonDeposit:
		return paymentv1.TransactionReason_TRANSACTION_REASON_DEPOSIT, nil
	default:
		return paymentv1.TransactionReason_TRANSACTION_REASON_UNSPECIFIED, fmt.Errorf(
			"unknown transaction reason: %v",
			reason,
		)
	}
}

func TransactionToProto(t *payment.Transaction) (*paymentv1.TransactionView, error) {
	reason, err := TransactionReasonToProto(t.Reason)
	if err != nil {
		return nil, err
	}
	amount := t.Amount.String()
	metadata, _ := TransactionMetadataToProto(t.Metadata)
	return &paymentv1.TransactionView{
		TransactionId: &sharedv1.TransactionId{
			Value: t.ID,
		},
		TonAmount: &sharedv1.TonAmount{
			Value: amount,
		},
		Reason:    reason,
		Metadata:  metadata,
		CreatedAt: timestamppb.New(t.CreatedAt),
	}, nil
}

func TransactionMetadataToProto(
	m *payment.TransactionMetadata,
) (*paymentv1.TransactionMetadata, error) {
	if m == nil || m.Gift == nil {
		return nil, errors.New("transaction metadata is nil or gift is nil")
	}
	return &paymentv1.TransactionMetadata{
		Data: &paymentv1.TransactionMetadata_Gift{
			Gift: &paymentv1.TransactionMetadata_GiftDetails{
				GiftId: m.Gift.GiftID,
				Title:  m.Gift.Title,
				Slug:   m.Gift.Slug,
			},
		},
	}, nil
}

func TransactionMetadataToDomain(
	m *paymentv1.TransactionMetadata,
) (*payment.TransactionMetadata, error) {
	if m == nil {
		return nil, errors.New("transaction metadata is nil")
	}
	switch metadata := m.GetData().(type) {
	case *paymentv1.TransactionMetadata_Gift:
		if metadata.Gift == nil {
			return nil, errors.New("transaction metadata gift is nil")
		}
		return &payment.TransactionMetadata{
			Gift: &payment.TransactionMetadataGiftDetails{
				GiftID: metadata.Gift.GetGiftId(),
				Title:  metadata.Gift.GetTitle(),
				Slug:   metadata.Gift.GetSlug(),
			},
		}, nil
	default:
		return nil, errors.New("transaction metadata is unknown")
	}
}
