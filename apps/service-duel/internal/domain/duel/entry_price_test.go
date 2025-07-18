package duel_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

func mustTon(t *testing.T, s string) *tonamount.TonAmount {
	t.Helper()
	a, err := tonamount.NewTonAmountFromString(s)
	if err != nil {
		t.Fatalf("cannot parse TonAmount from %q: %v", s, err)
	}
	return a
}

// Вспомогалка: создает дуэль с одним создателем и заданными ставками creatorStakes.
func makeDuelWithCreatorStakes(
	t *testing.T,
	creatorID duel.TelegramUserID,
	creatorStakes []string,
) *duel.Duel {
	t.Helper()
	params, err := duel.NewParamsBuilder().
		WithIsPrivate(false).
		WithMaxPlayers(duel.MaxPlayers(2)).
		WithMaxGifts(duel.MaxGifts(10)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	du := duel.NewDuel(params)
	// добавляем участника‑создателя
	creator, err := duel.NewParticipantBuilder().
		WithTelegramUserID(creatorID).
		WithPhoto("").
		AsCreator().
		Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	du.AddParticipant(creator)

	// добавляем второго «оппонента», чтобы ValidateEntry не ругался на отсутствие пользователя
	opID := duel.TelegramUserID(999)
	op, err := duel.NewParticipantBuilder().WithTelegramUserID(opID).WithPhoto("").Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	du.AddParticipant(op)

	// проставляем ставки создателя
	for i, s := range creatorStakes {
		amt := mustTon(t, s)
		gift, giftErr := duel.NewStakedGiftBuilder().
			WithID(strconv.Itoa(i)).
			WithTitle("Test Gift " + strconv.Itoa(i)).
			WithSlug("test-gift-" + strconv.Itoa(i)).
			WithPrice(amt).
			Build()
		if giftErr != nil {
			t.Fatalf("expected NewStakedGiftBuilder without error, got %v", giftErr)
		}
		stake, stakeErr := duel.NewStakeBuilder(creatorID).WithGift(gift).Build()
		if stakeErr != nil {
			t.Fatalf("expected NewStakeBuilder without error, got %v", stakeErr)
		}
		du.PlaceStake(stake)
	}
	return du
}

func TestEntryPriceRange(t *testing.T) {
	creator := duel.TelegramUserID(123)

	tests := []struct {
		name             string
		stakes           []string
		wantMin, wantMax string
		wantErr          error
	}{
		{
			name:    "одиночная ставка",
			stakes:  []string{"3.38"},
			wantMin: "3.21", wantMax: "3.55",
		},
		{
			name:    "несколько ставок",
			stakes:  []string{"2.0", "2.5", "1.25"},
			wantMin: "5.46", wantMax: "6.04",
		},
		{
			name:    "нет ставок от создателя",
			stakes:  []string{},
			wantErr: duel.ErrNoStakesFromCreator,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			du := makeDuelWithCreatorStakes(t, creator, tc.stakes)

			minStake, maxStake, err := du.EntryPriceRange()
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ожидали ошибку %v, получили %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("не ожидали ошибку, но получили %v", err)
			}
			if minStake.String() != tc.wantMin {
				t.Errorf("min: ожидали %s, получили %s", tc.wantMin, minStake.String())
			}
			if maxStake.String() != tc.wantMax {
				t.Errorf("max: ожидали %s, получили %s", tc.wantMax, maxStake.String())
			}
		})
	}
}

//nolint:gocognit
func TestValidateEntry(t *testing.T) {
	creatorTelegramUserID := duel.TelegramUserID(1)
	opponentTelegramUserID := duel.TelegramUserID(2)

	// готовим дуэль, создаем creator и opponent
	params, err := duel.NewParamsBuilder().
		WithIsPrivate(false).
		WithMaxPlayers(duel.MaxPlayers(2)).
		WithMaxGifts(duel.MaxGifts(10)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	du := duel.NewDuel(params)

	creator, err := duel.NewParticipantBuilder().
		WithTelegramUserID(creatorTelegramUserID).
		WithPhoto("").
		AsCreator().
		Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	du.AddParticipant(creator)
	opponent, err := duel.NewParticipantBuilder().
		WithTelegramUserID(opponentTelegramUserID).
		WithPhoto("").
		Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	du.AddParticipant(opponent)

	amt2, err := tonamount.NewTonAmountFromString("2.0")
	if err != nil {
		t.Fatalf("не смогли распарсить TonAmount из %q: %v", "2.0", err)
	}
	amt3, err := tonamount.NewTonAmountFromString("3.0")
	if err != nil {
		t.Fatalf("не смогли распарсить TonAmount из %q: %v", "3.0", err)
	}
	// creator ставит два подарка на 2.0 и 3.0 → диапазон [2.0,5.0]
	gift1, err := duel.NewStakedGiftBuilder().
		WithID("g1").
		WithTitle("Gift 1").
		WithSlug("gift-1").
		WithPrice(amt2).
		Build()
	if err != nil {
		t.Fatalf("expected NewStakedGiftBuilder without error, got %v", err)
	}
	gift2, err := duel.NewStakedGiftBuilder().
		WithID("g2").
		WithTitle("Gift 2").
		WithSlug("gift-2").
		WithPrice(amt3).
		Build()
	if err != nil {
		t.Fatalf("expected NewStakedGiftBuilder without error, got %v", err)
	}
	stake1, err := duel.NewStakeBuilder(creatorTelegramUserID).
		WithGift(gift1).
		Build()
	if err != nil {
		t.Fatalf("expected NewStakeBuilder without error, got %v", err)
	}
	du.PlaceStake(stake1)
	stake2, err := duel.NewStakeBuilder(creatorTelegramUserID).
		WithGift(gift2).
		Build()
	if err != nil {
		t.Fatalf("expected NewStakeBuilder without error, got %v", err)
	}
	du.PlaceStake(stake2)

	tests := []struct {
		name       string
		user       duel.TelegramUserID
		userStakes []string
		wantErr    error
	}{
		{
			name:       "stake with one gift in range",
			user:       opponentTelegramUserID,
			userStakes: []string{"5.0"},
			wantErr:    nil,
		},
		{
			name:       "multiple gifts in range",
			user:       opponentTelegramUserID,
			userStakes: []string{"1.0", "4.0"},
			wantErr:    nil,
		},
		{
			name:       "too low",
			user:       opponentTelegramUserID,
			userStakes: []string{"1.5"},
			wantErr:    duel.ErrStakeOutOfRange,
		},
		{
			name:       "too high",
			user:       opponentTelegramUserID,
			userStakes: []string{"6.0"},
			wantErr:    duel.ErrStakeOutOfRange,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// clear old stakes of opponent
			var newStakes []duel.Stake
			for _, s := range du.Stakes {
				if s.TelegramUserID != tc.user {
					newStakes = append(newStakes, s)
				}
			}
			du.Stakes = newStakes

			// add stakes of opponent
			for i, amtStr := range tc.userStakes {
				amt := mustTon(t, amtStr)
				gift, giftErr := duel.NewStakedGiftBuilder().
					WithID("op" + strconv.Itoa(i)).
					WithTitle("Opponent Gift " + strconv.Itoa(i)).
					WithSlug("opponent-gift-" + strconv.Itoa(i)).
					WithPrice(amt).
					Build()
				if giftErr != nil {
					t.Fatalf("expected NewStakedGiftBuilder without error, got %v", giftErr)
				}
				stake, stakeErr := duel.NewStakeBuilder(tc.user).
					WithGift(gift).
					Build()
				if stakeErr != nil {
					t.Fatalf("expected NewStakeBuilder without error, got %v", stakeErr)
				}
				du.PlaceStake(stake)
			}

			err = du.ValidateEntry(tc.user)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateEntry() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestEntryPriceRange_NoCreator(t *testing.T) {
	// duel without Creator flag, should be ErrCreatorNotFound
	params, err := duel.NewParamsBuilder().
		WithIsPrivate(false).
		WithMaxPlayers(duel.MaxPlayers(2)).
		WithMaxGifts(duel.MaxGifts(10)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	du := duel.NewDuel(params)
	// add only «regular» participant
	participant, err := duel.NewParticipantBuilder().WithTelegramUserID(42).WithPhoto("").Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	du.AddParticipant(participant)

	if _, _, err = du.EntryPriceRange(); !errors.Is(err, duel.ErrNoStakesFromCreator) {
		t.Fatalf("expected ErrNoStakesFromCreator, got %v", err)
	}
}

func TestValidateEntry_ParticipantNotFound(t *testing.T) {
	// ValidateEntry for unregistered user should return ErrParticipantNotFound
	params, err := duel.NewParamsBuilder().
		WithIsPrivate(false).
		WithMaxPlayers(duel.MaxPlayers(2)).
		WithMaxGifts(duel.MaxGifts(10)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	du := duel.NewDuel(params)
	creator, err := duel.NewParticipantBuilder().
		WithTelegramUserID(1).
		WithPhoto("").
		AsCreator().
		Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	du.AddParticipant(creator)
	// don't register opponent

	if err = du.ValidateEntry(999); !errors.Is(err, duel.ErrParticipantNotFound) {
		t.Fatalf("expected ErrParticipantNotFound, got %v", err)
	}
}
