package duel

import "github.com/peterparker2005/giftduels/packages/tonamount-go"

type Stake struct {
	TelegramUserID TelegramUserID
	Gift           StakedGift
}

func (s Stake) StakeValue() *tonamount.TonAmount {
	return s.Gift.Price
}

type StakeBuilder struct {
	s Stake
}

func NewStakeBuilder(telegramUserID TelegramUserID) *StakeBuilder {
	return &StakeBuilder{s: Stake{TelegramUserID: telegramUserID}}
}

func (b *StakeBuilder) WithGift(gift StakedGift) *StakeBuilder {
	b.s.Gift = gift
	return b
}

func (b *StakeBuilder) validate() error {
	return nil
}

func (b *StakeBuilder) Build() (Stake, error) {
	if err := b.validate(); err != nil {
		return Stake{}, err
	}
	return b.s, nil
}

type StakedGift struct {
	ID    string
	Title string
	Slug  string
	Price *tonamount.TonAmount
}

type StakedGiftBuilder struct {
	g StakedGift
}

func NewStakedGiftBuilder() *StakedGiftBuilder {
	return &StakedGiftBuilder{
		g: StakedGift{
			Price: tonamount.Zero(),
		},
	}
}

func (b *StakedGiftBuilder) WithID(id string) *StakedGiftBuilder {
	b.g.ID = id
	return b
}

func (b *StakedGiftBuilder) WithTitle(title string) *StakedGiftBuilder {
	b.g.Title = title
	return b
}

func (b *StakedGiftBuilder) WithSlug(slug string) *StakedGiftBuilder {
	b.g.Slug = slug
	return b
}

func (b *StakedGiftBuilder) WithPrice(price *tonamount.TonAmount) *StakedGiftBuilder {
	if price != nil {
		b.g.Price = price
	}
	return b
}

func (b *StakedGiftBuilder) validate() error {
	if b.g.ID == "" {
		return ErrEmptyGiftID
	}
	return nil
}

func (b *StakedGiftBuilder) Build() (StakedGift, error) {
	if err := b.validate(); err != nil {
		return StakedGift{}, err
	}
	return b.g, nil
}
