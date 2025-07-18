package duel

import "time"

type Roll struct {
	TelegramUserID TelegramUserID
	DiceValue      int32
	RolledAt       time.Time
	IsAutoRolled   bool
}

type RollBuilder struct {
	r Roll
}

func NewRollBuilder() *RollBuilder {
	return &RollBuilder{r: Roll{}}
}

func (b *RollBuilder) WithTelegramUserID(id TelegramUserID) *RollBuilder {
	b.r.TelegramUserID = id
	return b
}

func (b *RollBuilder) WithDiceValue(diceValue int32) *RollBuilder {
	b.r.DiceValue = diceValue
	return b
}

func (b *RollBuilder) WithRolledAt(rolledAt time.Time) *RollBuilder {
	b.r.RolledAt = rolledAt
	return b
}

func (b *RollBuilder) WithIsAutoRolled(isAutoRolled bool) *RollBuilder {
	b.r.IsAutoRolled = isAutoRolled
	return b
}

func (b *RollBuilder) validate() error {
	if b.r.DiceValue < 1 || b.r.DiceValue > 6 {
		return ErrInvalidDiceValue
	}
	return nil
}

func (b *RollBuilder) Build() (Roll, error) {
	if err := b.validate(); err != nil {
		return Roll{}, err
	}
	return b.r, nil
}
