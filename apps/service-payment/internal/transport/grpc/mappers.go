package grpc

import (
	"fmt"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TransactionReasonToDomain(reason paymentv1.TransactionReason) payment.TransactionReason {
	switch reason {
	case paymentv1.TransactionReason_TRANSACTION_REASON_WITHDRAW:
		return payment.TransactionReasonWithdraw
	case paymentv1.TransactionReason_TRANSACTION_REASON_REFUND:
		return payment.TransactionReasonRefund
	case paymentv1.TransactionReason_TRANSACTION_REASON_DEPOSIT:
		return payment.TransactionReasonDeposit
	default:
		panic(fmt.Sprintf("unknown transaction reason: %v", reason))
	}
}

func TransactionReasonToProto(reason payment.TransactionReason) paymentv1.TransactionReason {
	switch reason {
	case payment.TransactionReasonWithdraw:
		return paymentv1.TransactionReason_TRANSACTION_REASON_WITHDRAW
	case payment.TransactionReasonRefund:
		return paymentv1.TransactionReason_TRANSACTION_REASON_REFUND
	case payment.TransactionReasonDeposit:
		return paymentv1.TransactionReason_TRANSACTION_REASON_DEPOSIT
	default:
		panic(fmt.Sprintf("unknown transaction reason: %v", reason))
	}
}

func TransactionToProto(t *payment.Transaction) *paymentv1.TransactionView {
	return &paymentv1.TransactionView{
		TransactionId: &sharedv1.TransactionId{
			Value: t.ID,
		},
		TonAmount: &sharedv1.TonAmount{
			Value: t.Amount,
		},
		Reason:    TransactionReasonToProto(t.Reason),
		Metadata:  TransactionMetadataToProto(t.Metadata),
		CreatedAt: timestamppb.New(t.CreatedAt),
	}
}

func TransactionMetadataToProto(m *payment.TransactionMetadata) *paymentv1.TransactionMetadata {
	if m == nil || m.Gift == nil {
		return nil
	}
	return &paymentv1.TransactionMetadata{
		Data: &paymentv1.TransactionMetadata_Gift{
			Gift: &paymentv1.TransactionMetadata_GiftDetails{
				GiftId: m.Gift.GiftID,
				Title:  m.Gift.Title,
				Slug:   m.Gift.Slug,
			},
		},
	}
}

func TransactionMetadataToDomain(m *paymentv1.TransactionMetadata) *payment.TransactionMetadata {
	if m == nil {
		return nil
	}
	switch metadata := m.Data.(type) {
	case *paymentv1.TransactionMetadata_Gift:
		if metadata.Gift == nil {
			return nil
		}
		return &payment.TransactionMetadata{
			Gift: &payment.TransactionMetadata_GiftDetails{
				GiftID: metadata.Gift.GiftId,
				Title:  metadata.Gift.Title,
				Slug:   metadata.Gift.Slug,
			},
		}
	default:
		return nil
	}
}
