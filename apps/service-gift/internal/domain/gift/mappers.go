package gift

import (
	"fmt"

	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func AttributeTypeFromProto(t giftv1.GiftAttributeType) (AttributeType, error) {
	switch t {
	case giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_MODEL:
		return AttributeTypeModel, nil
	case giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_SYMBOL:
		return AttributeTypeSymbol, nil
	case giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_BACKDROP:
		return AttributeTypeBackdrop, nil
	case giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_UNSPECIFIED:
		fallthrough
	default:
		return "", fmt.Errorf("unknown attribute type: %v", t)
	}
}

// ConvertDomainGiftToProto converts domain Gift to protobuf Gift
func ConvertDomainGiftToProto(domainGift *Gift) *giftv1.Gift {
	protoGift := &giftv1.Gift{
		GiftId:            &sharedv1.GiftId{Value: domainGift.ID},
		OwnerTelegramId:   &sharedv1.TelegramUserId{Value: domainGift.OwnerTelegramID},
		Status:            ConvertDomainStatusToProto(domainGift.Status),
		TelegramMessageId: domainGift.UpgradeMessageID,
		Date:              timestamppb.New(domainGift.CreatedAt),
		Price:             &sharedv1.TonAmount{Value: domainGift.Price},
		EmojiId:           domainGift.EmojiID,
	}

	// Handle optional fields
	if domainGift.TelegramGiftID != 0 {
		protoGift.TelegramGiftId = &sharedv1.GiftTelegramId{Value: domainGift.TelegramGiftID}
	}

	if domainGift.CollectibleID != 0 {
		protoGift.CollectibleId = int32(domainGift.CollectibleID)
	}

	if domainGift.Title != "" {
		protoGift.Title = domainGift.Title
	}

	if domainGift.Slug != "" {
		protoGift.Slug = domainGift.Slug
	}

	if domainGift.WithdrawnAt != nil {
		protoGift.WithdrawnAt = timestamppb.New(*domainGift.WithdrawnAt)
	}

	return protoGift
}

// ConvertDomainGiftToProtoView converts domain Gift to protobuf GiftView
func ConvertDomainGiftToProtoView(domainGift *Gift) *giftv1.GiftView {
	protoView := &giftv1.GiftView{
		GiftId:  &sharedv1.GiftId{Value: domainGift.ID},
		Status:  ConvertDomainStatusToProto(domainGift.Status),
		Title:   domainGift.Title,
		Slug:    domainGift.Slug,
		Price:   &sharedv1.TonAmount{Value: domainGift.Price},
		EmojiId: domainGift.EmojiID,
	}

	// Handle optional fields
	if domainGift.TelegramGiftID != 0 {
		protoView.TelegramGiftId = &sharedv1.GiftTelegramId{Value: domainGift.TelegramGiftID}
	}

	if domainGift.CollectibleID != 0 {
		protoView.CollectibleId = int32(domainGift.CollectibleID)
	}

	if domainGift.WithdrawnAt != nil {
		protoView.WithdrawnAt = timestamppb.New(*domainGift.WithdrawnAt)
	}

	return protoView
}

// ConvertDomainStatusToProto converts domain Status to protobuf GiftStatus
func ConvertDomainStatusToProto(domainStatus Status) giftv1.GiftStatus {
	switch domainStatus {
	case StatusPending:
		return giftv1.GiftStatus_GIFT_STATUS_OWNED
	case StatusWithdrawn:
		return giftv1.GiftStatus_GIFT_STATUS_WITHDRAWN
	case StatusInGame:
		return giftv1.GiftStatus_GIFT_STATUS_IN_GAME
	default:
		return giftv1.GiftStatus_GIFT_STATUS_UNSPECIFIED
	}
}
