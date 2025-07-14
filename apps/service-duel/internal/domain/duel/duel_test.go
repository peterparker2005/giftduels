//nolint:testpackage // test package
package duel

import (
	"errors"
	"testing"
	"time"

	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

func mustTon(s string) *tonamount.TonAmount {
	a, err := tonamount.NewTonAmountFromString(s)
	if err != nil {
		panic(err)
	}
	return a
}

func TestJoin(t *testing.T) {
	params := NewParams(false, MaxPlayers(2), MaxGifts(1))
	d := NewDuel(params)

	// Успешные два Join
	if err := d.Join(NewParticipant(1, "", true)); err != nil {
		t.Fatalf("expected Join without error, got %v", err)
	}
	if err := d.Join(NewParticipant(2, "", false)); err != nil {
		t.Fatalf("expected Join without error, got %v", err)
	}

	// Третий превысит лимит
	if err := d.Join(NewParticipant(3, "", false)); !errors.Is(err, ErrMaxPlayersExceeded) {
		t.Fatalf("expected ErrMaxPlayersExceeded, got %v", err)
	}

	// Дубль
	if err := d.Join(NewParticipant(1, "", false)); !errors.Is(err, ErrAlreadyJoined) {
		t.Fatalf("expected ErrAlreadyJoined, got %v", err)
	}
}

func TestPlaceStake(t *testing.T) {
	params := NewParams(false, MaxPlayers(2), MaxGifts(1))
	d := NewDuel(params)
	p1 := NewParticipant(1, "", true)
	d.Join(p1)

	// Ставка от незарегистрированного игрока
	err := d.PlaceStake(
		NewStake(2, NewStakedGift("gift1", "Gift 1", "gift-1", mustTon("1.5")), mustTon("1.5")),
	)
	if !errors.Is(err, ErrParticipantNotFound) {
		t.Fatalf("expected ErrParticipantNotFound, got %v", err)
	}

	// Успешная ставка
	stake := NewStake(1, NewStakedGift("gift1", "Gift 1", "gift-1", mustTon("2.0")), mustTon("2.0"))
	if err = d.PlaceStake(stake); err != nil {
		t.Fatalf("expected PlaceStake without error, got %v", err)
	}
	// TotalStakeValue выросло
	sum := mustTon("2.0")
	if d.TotalStakeValue.String() != sum.String() {
		t.Errorf("expected TotalStakeValue %s, got %s", sum, d.TotalStakeValue)
	}
}

func TestRoundFlow_ThreePlayers(t *testing.T) {
	params := NewParams(false, MaxPlayers(2), MaxGifts(1))
	d := NewDuel(params)
	p1 := TelegramUserID(1)
	p2 := TelegramUserID(2)
	p3 := TelegramUserID(3)
	d.Join(NewParticipant(p1, "", true))
	d.Join(NewParticipant(p2, "", false))
	d.Join(NewParticipant(p3, "", false))

	// Первый раунд
	d.StartRound([]TelegramUserID{p1, p2, p3})

	// Пока не оба броска — не finished
	_, finished := d.EvaluateCurrentRound()
	if finished {
		t.Error("expected finished=false before rolls")
	}

	// Первый бросок
	roll1 := NewRoll(p1, 4, time.Now(), false)
	if err := d.AddRollToCurrentRound(roll1); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}
	_, finished = d.EvaluateCurrentRound()
	if finished {
		t.Error("expected finished=false after one roll")
	}

	// Второй бросок, p2 > p1
	roll2 := NewRoll(p2, 5, time.Now(), false)
	if err := d.AddRollToCurrentRound(roll2); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}

	// Третий бросок, p3 > p2
	roll3 := NewRoll(p3, 6, time.Now(), false)
	if err := d.AddRollToCurrentRound(roll3); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}
	winners, finished := d.EvaluateCurrentRound()
	if !finished {
		t.Fatal("expected finished=true after two rolls")
	}
	if len(winners) != 1 || winners[0] != p3 {
		t.Errorf("expected winner %d, got %v", p3, winners)
	}
}

func TestRoundFlow_TieThenResolve(t *testing.T) {
	params := NewParams(false, MaxPlayers(2), MaxGifts(1))
	d := NewDuel(params)
	p1 := TelegramUserID(1)
	p2 := TelegramUserID(2)
	d.Join(NewParticipant(p1, "", true))
	d.Join(NewParticipant(p2, "", false))

	// Первый раунд
	d.StartRound([]TelegramUserID{p1, p2})
	_ = d.AddRollToCurrentRound(NewRoll(p1, 3, time.Now(), false))
	_ = d.AddRollToCurrentRound(NewRoll(p2, 3, time.Now(), false))
	winners, finished := d.EvaluateCurrentRound()
	if !finished || len(winners) != 2 {
		t.Fatalf("expected tie between two, got finished=%v winners=%v", finished, winners)
	}

	// Запускаем новый раунд только для tied
	d.StartRound(winners)
	round2 := d.Rounds[len(d.Rounds)-1]
	if len(round2.Participants) != 2 ||
		round2.Participants[0] != p1 ||
		round2.Participants[1] != p2 {
		t.Errorf("expected new round with the same two, got %v", round2.Participants)
	}

	// Решающий бросок
	_ = d.AddRollToCurrentRound(NewRoll(p1, 6, time.Now(), false))
	_ = d.AddRollToCurrentRound(NewRoll(p2, 2, time.Now(), false))
	finalWinners, finished := d.EvaluateCurrentRound()
	if !finished || len(finalWinners) != 1 || finalWinners[0] != p1 {
		t.Errorf("expected winner %d, got %v", p1, finalWinners)
	}
}

func TestCompleteAndCancel(t *testing.T) {
	params := NewParams(false, MaxPlayers(2), MaxGifts(1))
	d := NewDuel(params)

	// Cancel в waiting
	if err := d.Cancel(); err != nil {
		t.Fatalf("Cancel in waiting should work, got %v", err)
	}
	if d.Status != StatusCancelled {
		t.Errorf("expected StatusCancelled, got %s", d.Status)
	}

	// Complete в wrong status
	if err := d.Complete(1); !errors.Is(err, ErrDuelNotInProgress) {
		t.Errorf("expected ErrDuelNotInProgress, got %v", err)
	}

	// Переводим в in_progress и Complete
	d2 := NewDuel(params)
	d2.Status = StatusInProgress
	if err := d2.Complete(42); err != nil {
		t.Fatalf("Complete in in_progress should work, got %v", err)
	}
	if d2.Status != StatusCompleted || d2.WinnerID == nil || *d2.WinnerID != 42 {
		t.Errorf(
			"expected completed duel with winner 42, got status=%s winner=%v",
			d2.Status,
			d2.WinnerID,
		)
	}
}
