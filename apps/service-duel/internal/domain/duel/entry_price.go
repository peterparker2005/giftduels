package duel

import (
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"github.com/shopspring/decimal"
)

const tolerance = 0.05 // 5%

func (d *Duel) EntryPriceRange() (*tonamount.TonAmount, *tonamount.TonAmount, error) {
	// 1. Находим ID создателя (как раньше)…
	// 2. Складываем все его ставки в sum

	var creatorID TelegramUserID
	for _, p := range d.Participants {
		if p.IsCreator {
			creatorID = p.TelegramUserID
			break
		}
	}

	sum := tonamount.Zero()
	for _, s := range d.Stakes {
		if s.TelegramUserID == creatorID {
			sum = sum.Add(s.StakeValue)
		}
	}
	if sum.IsZero() {
		return nil, nil, ErrNoStakesFromCreator
	}

	// 3. Умножаем на (1–tolerance) и (1+tolerance)
	total := sum.Decimal()
	low := total.Mul(decimal.NewFromFloat(1 - tolerance))
	high := total.Mul(decimal.NewFromFloat(1 + tolerance))

	lowTA, err := tonamount.NewTonAmountFromString(low.String())
	if err != nil {
		return nil, nil, err
	}
	highTA, err := tonamount.NewTonAmountFromString(high.String())
	if err != nil {
		return nil, nil, err
	}

	return lowTA, highTA, nil
}

func (d *Duel) ValidateEntry(userID TelegramUserID) error {
	var participantFound bool
	for _, p := range d.Participants {
		if p.TelegramUserID == userID {
			participantFound = true
			break
		}
	}
	if !participantFound {
		return ErrParticipantNotFound
	}

	minStake, maxStake, err := d.EntryPriceRange()
	if err != nil {
		return err
	}

	total, _ := tonamount.NewTonAmountFromString("0")
	for _, s := range d.Stakes {
		if s.TelegramUserID == userID {
			total = total.Add(s.StakeValue)
		}
	}

	if total.Decimal().Cmp(minStake.Decimal()) < 0 || total.Decimal().Cmp(maxStake.Decimal()) > 0 {
		return ErrStakeOutOfRange
	}
	return nil
}
