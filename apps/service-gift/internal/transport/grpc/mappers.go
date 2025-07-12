package grpc

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DomainGiftToProto преобразует domain Gift в protobuf Gift
func DomainGiftToProto(domainGift *gift.Gift) *giftv1.Gift {
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
		Backdrop:          DomainBackdropToProto(&domainGift.Backdrop),
		Model:             DomainModelToProto(&domainGift.Model),
		Symbol:            DomainSymbolToProto(&domainGift.Symbol),
	}

	if domainGift.WithdrawnAt != nil {
		protoGift.WithdrawnAt = timestamppb.New(*domainGift.WithdrawnAt)
	}

	return protoGift
}

func DomainAttributesToProto(attributes []gift.Attribute) []*giftv1.GiftAttribute {
	protoAttributes := make([]*giftv1.GiftAttribute, len(attributes))
	for i, attr := range attributes {
		var protoAttribute *giftv1.GiftAttribute
		switch attr.Type {
		case gift.AttributeTypeBackdrop:
			protoAttribute = &giftv1.GiftAttribute{
				Type:           giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_BACKDROP,
				Name:           attr.Name,
				RarityPerMille: attr.RarityPerMille,
			}
		case gift.AttributeTypeModel:
			protoAttribute = &giftv1.GiftAttribute{
				Type:           giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_MODEL,
				Name:           attr.Name,
				RarityPerMille: attr.RarityPerMille,
			}
		case gift.AttributeTypeSymbol:
			protoAttribute = &giftv1.GiftAttribute{
				Type:           giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_SYMBOL,
				Name:           attr.Name,
				RarityPerMille: attr.RarityPerMille,
			}
		}
		protoAttributes[i] = protoAttribute
	}
	return protoAttributes
}

// DomainGiftToProtoView преобразует domain Gift в protobuf GiftView
func DomainGiftToProtoView(domainGift *gift.Gift) *giftv1.GiftView {
	protoView := &giftv1.GiftView{
		GiftId:         &sharedv1.GiftId{Value: domainGift.ID},
		Status:         DomainStatusToProto(domainGift.Status),
		Title:          domainGift.Title,
		Slug:           domainGift.Slug,
		Price:          &sharedv1.TonAmount{Value: domainGift.Price},
		TelegramGiftId: &sharedv1.GiftTelegramId{Value: domainGift.TelegramGiftID},
		CollectibleId:  domainGift.CollectibleID,
	}

	protoView.Attributes = append(protoView.Attributes, &giftv1.GiftAttribute{
		Type:           giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_BACKDROP,
		Name:           domainGift.Backdrop.Name,
		RarityPerMille: domainGift.Backdrop.RarityPerMille,
	})

	protoView.Attributes = append(protoView.Attributes, &giftv1.GiftAttribute{
		Type:           giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_MODEL,
		Name:           domainGift.Model.Name,
		RarityPerMille: domainGift.Model.RarityPerMille,
	})

	protoView.Attributes = append(protoView.Attributes, &giftv1.GiftAttribute{
		Type:           giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_SYMBOL,
		Name:           domainGift.Symbol.Name,
		RarityPerMille: domainGift.Symbol.RarityPerMille,
	})

	if domainGift.WithdrawnAt != nil {
		protoView.WithdrawnAt = timestamppb.New(*domainGift.WithdrawnAt)
	}

	return protoView
}

// DomainBackdropToProto преобразует domain Backdrop в protobuf GiftAttributeBackdrop
func DomainBackdropToProto(backdrop *gift.Backdrop) *giftv1.GiftAttributeBackdrop {
	protoBackdrop := &giftv1.GiftAttributeBackdrop{
		Name:           backdrop.Name,
		RarityPerMille: backdrop.RarityPerMille,
	}

	if backdrop.CenterColor != nil {
		protoBackdrop.CenterColor = *backdrop.CenterColor
	}
	if backdrop.EdgeColor != nil {
		protoBackdrop.EdgeColor = *backdrop.EdgeColor
	}
	if backdrop.PatternColor != nil {
		protoBackdrop.PatternColor = *backdrop.PatternColor
	}
	if backdrop.TextColor != nil {
		protoBackdrop.TextColor = *backdrop.TextColor
	}

	return protoBackdrop
}

// DomainModelToProto преобразует domain Model в protobuf GiftAttributeModel
func DomainModelToProto(model *gift.Model) *giftv1.GiftAttributeModel {
	return &giftv1.GiftAttributeModel{
		Name:           model.Name,
		RarityPerMille: model.RarityPerMille,
	}
}

// DomainSymbolToProto преобразует domain Symbol в protobuf GiftAttributeSymbol
func DomainSymbolToProto(symbol *gift.Symbol) *giftv1.GiftAttributeSymbol {
	return &giftv1.GiftAttributeSymbol{
		Name:           symbol.Name,
		RarityPerMille: symbol.RarityPerMille,
	}
}

// DomainStatusToProto преобразует domain статус в protobuf статус
func DomainStatusToProto(domainStatus gift.Status) giftv1.GiftStatus {
	switch domainStatus {
	case gift.StatusOwned:
		return giftv1.GiftStatus_GIFT_STATUS_OWNED
	case gift.StatusWithdrawn:
		return giftv1.GiftStatus_GIFT_STATUS_WITHDRAWN
	case gift.StatusInGame:
		return giftv1.GiftStatus_GIFT_STATUS_IN_GAME
	case gift.StatusWithdrawPending:
		return giftv1.GiftStatus_GIFT_STATUS_WITHDRAW_PENDING
	default:
		return giftv1.GiftStatus_GIFT_STATUS_UNSPECIFIED
	}
}
