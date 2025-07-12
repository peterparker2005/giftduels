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

func DomainAttributesToProto(attributes []Attribute) []*giftv1.GiftAttribute {
	protoAttributes := make([]*giftv1.GiftAttribute, len(attributes))
	for i, attr := range attributes {
		protoAttributes[i] = &giftv1.GiftAttribute{
			Type:           DomainAttributeTypeToProto(attr.Type),
			Name:           attr.Name,
			RarityPerMille: attr.RarityPerMille,
		}
	}
	return protoAttributes
}

func DomainAttributeTypeToProto(attributeType AttributeType) giftv1.GiftAttributeType {
	switch attributeType {
	case AttributeTypeModel:
		return giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_MODEL
	case AttributeTypeSymbol:
		return giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_SYMBOL
	case AttributeTypeBackdrop:
		return giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_BACKDROP
	default:
		return giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_UNSPECIFIED
	}
}

// DomainGiftToProto s domain Gift to protobuf Gift
func DomainGiftToProto(domainGift *Gift) *giftv1.Gift {
	protoGift := &giftv1.Gift{
		GiftId:            &sharedv1.GiftId{Value: domainGift.ID},
		OwnerTelegramId:   &sharedv1.TelegramUserId{Value: domainGift.OwnerTelegramID},
		Status:            DomainStatusToProto(domainGift.Status),
		TelegramMessageId: domainGift.UpgradeMessageID,
		Date:              timestamppb.New(domainGift.CreatedAt),
		Price:             &sharedv1.TonAmount{Value: domainGift.Price},
		TelegramGiftId:    &sharedv1.GiftTelegramId{Value: domainGift.TelegramGiftID},
		CollectibleId:     domainGift.CollectibleID,
		Title:             domainGift.Title,
		Slug:              domainGift.Slug,
		Attributes:        DomainAttributesToProto(domainGift.Attributes),
	}

	if domainGift.WithdrawnAt != nil {
		protoGift.WithdrawnAt = timestamppb.New(*domainGift.WithdrawnAt)
	}

	return protoGift
}

// DomainGiftToProtoView s domain Gift to protobuf GiftView
func DomainGiftToProtoView(domainGift *Gift) *giftv1.GiftView {
	protoView := &giftv1.GiftView{
		GiftId:         &sharedv1.GiftId{Value: domainGift.ID},
		Status:         DomainStatusToProto(domainGift.Status),
		Title:          domainGift.Title,
		Slug:           domainGift.Slug,
		Price:          &sharedv1.TonAmount{Value: domainGift.Price},
		Attributes:     DomainAttributesToProto(domainGift.Attributes),
		TelegramGiftId: &sharedv1.GiftTelegramId{Value: domainGift.TelegramGiftID},
		CollectibleId:  domainGift.CollectibleID,
	}

	if domainGift.WithdrawnAt != nil {
		protoView.WithdrawnAt = timestamppb.New(*domainGift.WithdrawnAt)
	}

	return protoView
}

// DomainStatusToProto s domain Status to protobuf GiftStatus
func DomainStatusToProto(domainStatus Status) giftv1.GiftStatus {
	switch domainStatus {
	case StatusOwned:
		return giftv1.GiftStatus_GIFT_STATUS_OWNED
	case StatusWithdrawn:
		return giftv1.GiftStatus_GIFT_STATUS_WITHDRAWN
	case StatusInGame:
		return giftv1.GiftStatus_GIFT_STATUS_IN_GAME
	default:
		return giftv1.GiftStatus_GIFT_STATUS_UNSPECIFIED
	}
}
