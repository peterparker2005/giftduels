package event

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/lucsky/cuid"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/db"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/pricing"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TelegramGiftReceivedHandler struct {
	repo        gift.Repository
	pricingRepo pricing.Repository
	logger      *zap.Logger
}

func NewTelegramGiftReceivedHandler(repo gift.Repository, pricingRepo pricing.Repository, logger *zap.Logger) *TelegramGiftReceivedHandler {
	return &TelegramGiftReceivedHandler{
		repo:        repo,
		pricingRepo: pricingRepo,
		logger:      logger,
	}
}

func (h *TelegramGiftReceivedHandler) Handle(msg *message.Message) error {
	ctx := context.Background()
	h.logger.Info("Processing telegram gift received event", zap.String("message_id", msg.UUID))

	var ev giftv1.TelegramGiftReceivedEvent
	if err := proto.Unmarshal(msg.Payload, &ev); err != nil {
		h.logger.Error("Failed to unmarshal event", zap.Error(err), zap.String("message_id", msg.UUID))
		return fmt.Errorf("unmarshal event: %w", err)
	}

	// Проверяем обязательные поля
	if ev.OwnerTelegramId == nil {
		h.logger.Error("Missing OwnerTelegramId in event", zap.String("message_id", msg.UUID))
		return fmt.Errorf("missing OwnerTelegramId in event")
	}

	h.logger.Info("Event unmarshaled successfully",
		zap.String("message_id", msg.UUID),
		zap.String("title", ev.Title),
		zap.Int32("collectible_id", ev.CollectibleId),
		zap.Int64("owner_id", ev.OwnerTelegramId.Value))

	// собрали ВСЕ атрибуты
	attrs := pricing.Attributes{
		"gift_name": ev.Title,
	}
	for _, a := range ev.Attributes {
		switch a.Type {
		case giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_MODEL:
			attrs["model"] = fmt.Sprintf("%s (%s)", a.Name, rarityToPercent(a.Rarity))
		case giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_SYMBOL:
			attrs["symbol"] = a.Name
		case giftv1.GiftAttributeType_GIFT_ATTRIBUTE_TYPE_BACKDROP:
			attrs["backdrop"] = a.Name
		}
	}

	// порядок «важности» атрибутов: чем раньше, тем важнее
	keyOrder := []string{"gift_name", "model", "symbol", "backdrop"}

	observations, err := h.queryUntilEnough(ctx, attrs, keyOrder)
	if err != nil {
		h.logger.Error("Failed to query pricing", zap.Error(err), zap.String("message_id", msg.UUID))
		return err // логика выше решит, что делать
	}

	h.logger.Info("Pricing data retrieved",
		zap.String("message_id", msg.UUID),
		zap.Int("items_count", len(observations)))

	// Вычисляем floor price в TON
	floorPriceTON := calcFloor(observations)

	id := cuid.New()

	gift := &db.Gift{
		ID:               id,
		OwnerTelegramID:  ev.OwnerTelegramId.Value,
		UpgradeMessageID: ev.UpgradeMessageId,
		Status:           db.GiftStatusOwned,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err = h.repo.CreateGift(ctx, gift)
	if err != nil {
		h.logger.Error("Failed to save gift", zap.Error(err), zap.String("message_id", msg.UUID))
		return err
	}

	h.logger.Info("Gift processed successfully",
		zap.String("message_id", msg.UUID),
		zap.String("gift_id", id),
		zap.Float64("price_ton", floorPriceTON))

	return nil
}

func (h *TelegramGiftReceivedHandler) queryUntilEnough(
	ctx context.Context,
	all pricing.Attributes,
	order []string,
) ([]pricing.Observation, error) {
	filter := copyMap(all)

	for i := 0; ; i++ {
		observations, err := h.pricingRepo.Samples(ctx, pricing.Filter{
			Attributes: filter,
		}, 30)
		if err != nil {
			return nil, fmt.Errorf("pricing samples: %w", err)
		}
		if len(observations) >= 3 {
			return observations, nil
		}

		// если уже убрали все атрибуты — используем то, что есть
		if i >= len(order) {
			// Если есть хотя бы одно наблюдение, используем его
			if len(observations) > 0 {
				return observations, nil
			}
			// Если нет наблюдений вообще, возвращаем дефолтное значение
			// Это позволит обработать сообщение без ошибки
			return []pricing.Observation{{TonPrice: 1.0}}, nil
		}
		delete(filter, order[len(order)-1-i]) // срезаем наименее важный
	}
}

func copyMap(src pricing.Attributes) pricing.Attributes {
	dst := make(pricing.Attributes, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func calcFloor(observations []pricing.Observation) float64 {
	if len(observations) == 0 {
		return 0.1 // или любое другое значение по умолчанию
	}

	min := observations[0].TonPrice
	for _, obs := range observations[1:] {
		if obs.TonPrice < min {
			min = obs.TonPrice
		}
	}
	return min
}

func rarityToPercent(r int32) string {
	percent := float64(r) / 10.0
	s := fmt.Sprintf("%.2f", percent) // максимум два знака
	s = strings.TrimRight(s, "0")     // убираем лишние нули
	s = strings.TrimRight(s, ".")     // если осталось "x." — убираем точку
	return s + "%"
}
