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
func makeDuelWithCreatorStakes(t *testing.T, creatorID duel.TelegramUserID, creatorStakes []string) *duel.Duel {
	t.Helper()
	du := duel.NewDuel(duel.NewParams(false, duel.MaxPlayers(2), duel.MaxGifts(10)))
	// добавляем участника‑создателя
	du.Join(duel.NewParticipant(creatorID, "", true))

	// добавляем второго «оппонента», чтобы ValidateEntry не ругался на отсутствие пользователя
	opID := duel.TelegramUserID(999)
	du.Join(duel.NewParticipant(opID, "", false))

	// проставляем ставки создателя
	for i, s := range creatorStakes {
		amt := mustTon(t, s)
		du.PlaceStake(duel.NewStake(creatorID, duel.NewStakedGift(strconv.Itoa(i), "", "", amt), amt))
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
			wantMin: "3.38", wantMax: "3.38",
		},
		{
			name:    "несколько ставок",
			stakes:  []string{"2.0", "2.5", "1.25"},
			wantMin: "1.25", wantMax: "5.75", // 2.0+2.5+1.25 = 5.75
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

func TestValidateEntry(t *testing.T) {
	creator := duel.TelegramUserID(1)
	opponent := duel.TelegramUserID(2)

	// готовим дуэль, создаем creator и opponent
	du := duel.NewDuel(duel.NewParams(false, duel.MaxPlayers(2), duel.MaxGifts(10)))
	du.Join(duel.NewParticipant(creator, "", true))
	du.Join(duel.NewParticipant(opponent, "", false))

	amt2, err := tonamount.NewTonAmountFromString("2.0")
	if err != nil {
		t.Fatalf("не смогли распарсить TonAmount из %q: %v", "2.0", err)
	}
	amt3, err := tonamount.NewTonAmountFromString("3.0")
	if err != nil {
		t.Fatalf("не смогли распарсить TonAmount из %q: %v", "3.0", err)
	}
	// creator ставит два подарка на 2.0 и 3.0 → диапазон [2.0,5.0]
	du.PlaceStake(duel.NewStake(creator, duel.NewStakedGift("g1", "", "", amt2), amt2))
	du.PlaceStake(duel.NewStake(creator, duel.NewStakedGift("g2", "", "", amt3), amt3))

	tests := []struct {
		name       string
		user       duel.TelegramUserID
		userStakes []string
		wantErr    error
	}{
		{
			name:       "stake with one gift in range",
			user:       opponent,
			userStakes: []string{"4.5"},
			wantErr:    nil,
		},
		{
			name:       "multiple gifts in range",
			user:       opponent,
			userStakes: []string{"1.0", "4.0"},
			wantErr:    nil,
		},
		{
			name:       "too low",
			user:       opponent,
			userStakes: []string{"1.5"},
			wantErr:    duel.ErrStakeOutOfRange,
		},
		{
			name:       "too high",
			user:       opponent,
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
				du.PlaceStake(duel.NewStake(tc.user, duel.NewStakedGift("op"+strconv.Itoa(i), "", "", amt), amt))
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
	du := duel.NewDuel(duel.NewParams(false, duel.MaxPlayers(2), duel.MaxGifts(10)))
	// add only «regular» participant
	du.Join(duel.NewParticipant(42, "", false))

	if _, _, err := du.EntryPriceRange(); !errors.Is(err, duel.ErrCreatorNotFound) {
		t.Fatalf("expected ErrCreatorNotFound, got %v", err)
	}
}

func TestValidateEntry_ParticipantNotFound(t *testing.T) {
	// ValidateEntry for unregistered user should return ErrParticipantNotFound
	du := duel.NewDuel(duel.NewParams(false, duel.MaxPlayers(2), duel.MaxGifts(10)))
	du.Join(duel.NewParticipant(1, "", true)) // creator
	// don't register opponent

	if err := du.ValidateEntry(999); !errors.Is(err, duel.ErrParticipantNotFound) {
		t.Fatalf("expected ErrParticipantNotFound, got %v", err)
	}
}
