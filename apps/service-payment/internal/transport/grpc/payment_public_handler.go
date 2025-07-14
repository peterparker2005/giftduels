package grpc

import (
	"context"

	"github.com/ccoveille/go-safecast"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	paymentdomain "github.com/peterparker2005/giftduels/apps/service-payment/internal/domain/payment"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	"github.com/peterparker2005/giftduels/apps/service-payment/pkg/boc"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	paymentv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/payment/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentPublicHandler struct {
	paymentv1.UnimplementedPaymentPublicServiceServer

	service *payment.Service
	cfg     *config.Config
}

func NewPaymentPublicHandler(service *payment.Service, cfg *config.Config) paymentv1.PaymentPublicServiceServer {
	return &PaymentPublicHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *PaymentPublicHandler) GetBalance(
	ctx context.Context,
	_ *emptypb.Empty,
) (*paymentv1.GetBalanceResponse, error) {
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
				Value: balance.TonAmount.String(),
			},
		},
	}, nil
}

func (h *PaymentPublicHandler) PreviewWithdraw(
	ctx context.Context,
	req *paymentv1.PreviewWithdrawRequest,
) (*paymentv1.PreviewWithdrawResponse, error) {
	// Конвертируем protobuf запрос в доменные объекты
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

	resp, err := h.service.PreviewWithdraw(ctx, gifts)
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

func (h *PaymentPublicHandler) DepositTon(
	ctx context.Context,
	req *paymentv1.DepositTonRequest,
) (*paymentv1.DepositTonResponse, error) {
	userID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 1) Создаём запись в БД с «сырым» UUID-payload:
	tonAmount := req.GetTonAmount().GetValue()
	deposit, err := h.service.CreateDeposit(ctx, userID, tonAmount)
	if err != nil {
		return nil, err
	}

	// 2) Упаковываем этот UUID в BOC и кодируем
	bocPayload, err := boc.EncodeStringAsBOC(deposit.Payload)
	if err != nil {
		// если вдруг что-то пошло не так
		return nil, errors.NewInternalError("failed to encode payload boc")
	}

	// 3) Отдаём клиенту nanoAmount + BOC-payload
	return &paymentv1.DepositTonResponse{
		DepositId: deposit.ID.String(),
		// FIXME: переделать на TonAmount?
		NanoTonAmount:   deposit.AmountNano,
		Payload:         bocPayload,
		TreasuryAddress: h.cfg.Ton.WalletAddress,
	}, nil
}

func (h *PaymentPublicHandler) GetTransactionHistory(
	ctx context.Context,
	req *paymentv1.GetTransactionHistoryRequest,
) (*paymentv1.GetTransactionHistoryResponse, error) {
	userID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	pagination := shared.NewPageRequest(req.GetPagination().GetPage(), req.GetPagination().GetPageSize())
	transactions, count, err := h.service.GetTransactionHistory(ctx, userID, pagination)
	if err != nil {
		return nil, err
	}

	transactionsProto := make([]*paymentv1.TransactionView, 0, len(transactions))
	for _, transaction := range transactions {
		transactionProto, mapErr := TransactionToProto(transaction)
		if mapErr != nil {
			return nil, mapErr
		}
		transactionsProto = append(transactionsProto, transactionProto)
	}

	countInt32, err := safecast.ToInt32(count)
	if err != nil {
		return nil, err
	}

	return &paymentv1.GetTransactionHistoryResponse{
		Transactions: transactionsProto,
		Pagination: &sharedv1.PageResponse{
			Total:      countInt32,
			Page:       pagination.Page(),
			PageSize:   pagination.PageSize(),
			TotalPages: pagination.TotalPages(countInt32),
		},
	}, nil
}
