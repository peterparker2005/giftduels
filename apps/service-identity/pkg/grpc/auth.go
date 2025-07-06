package grpc

import (
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
)

type contextKey string

const (
	TelegramUserIDKey = contextKey("telegram_user_id")
)

// Список методов, которые требуют авторизацию
var protectedMethods = map[string]bool{
	identityv1.IdentityPublicService_ValidateToken_FullMethodName: false,
	identityv1.IdentityPublicService_Authorize_FullMethodName:     false,

	giftv1.GiftPublicService_GetGifts_FullMethodName:           true,
	giftv1.GiftPublicService_GetGift_FullMethodName:            true,
	giftv1.GiftPublicService_GetWithdrawOptions_FullMethodName: true,
	giftv1.GiftPublicService_WithdrawGift_FullMethodName:       true,
	// Пример:
	// "/gice.user.v1.UserPrivateService/UpdateProfile": true,
}
