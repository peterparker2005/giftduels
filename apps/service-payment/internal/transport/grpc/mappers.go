package grpc

import (
	"fmt"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
)

func mapTransactionReason(reason paymentv1.TransactionReason) payment.TransactionReason {
	switch reason {
	case paymentv1.TransactionReason_TRANSACTION_REASON_WITHDRAW:
		return payment.TransactionReasonWithdraw
	case paymentv1.TransactionReason_TRANSACTION_REASON_DEPOSIT:
		return payment.TransactionReasonRefund
	case paymentv1.TransactionReason_TRANSACTION_REASON_REFUND:
		return payment.TransactionReasonDeposit
	default:
		panic(fmt.Sprintf("unknown transaction reason: %v", reason))
	}
}
