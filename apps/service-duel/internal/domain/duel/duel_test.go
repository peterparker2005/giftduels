package duel_test

import (
	"errors"
	"testing"
	"time"

	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
)

func TestJoin(t *testing.T) {
	params, err := dueldomain.NewParamsBuilder().WithIsPrivate(false).
		WithMaxPlayers(dueldomain.MaxPlayers(2)).
		WithMaxGifts(dueldomain.MaxGifts(1)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	d := dueldomain.NewDuel(params)

	// Успешные два Join
	participant1, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(1).
		WithPhoto("").
		AsCreator().
		Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	participant2, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(2).WithPhoto("").Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	if err = d.AddParticipant(participant1); err != nil {
		t.Fatalf("expected Join without error, got %v", err)
	}
	if err = d.AddParticipant(participant2); err != nil {
		t.Fatalf("expected Join without error, got %v", err)
	}

	// Третий превысит лимит
	participant3, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(3).WithPhoto("").Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	if err = d.AddParticipant(participant3); !errors.Is(err, dueldomain.ErrMaxPlayersExceeded) {
		t.Fatalf("expected ErrMaxPlayersExceeded, got %v", err)
	}

	// Дубль
	if err = d.AddParticipant(participant1); !errors.Is(err, dueldomain.ErrAlreadyJoined) {
		t.Fatalf("expected ErrAlreadyJoined, got %v", err)
	}
}

func TestPlaceStake(t *testing.T) {
	params, err := dueldomain.NewParamsBuilder().WithIsPrivate(false).
		WithMaxPlayers(dueldomain.MaxPlayers(2)).
		WithMaxGifts(dueldomain.MaxGifts(1)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	d := dueldomain.NewDuel(params)
	p1, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(1).WithPhoto("").AsCreator().Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	if err = d.AddParticipant(p1); err != nil {
		t.Fatalf("expected AddParticipant without error, got %v", err)
	}

	// Ставка от незарегистрированного игрока
	stakedGift, err := dueldomain.NewStakedGiftBuilder().WithID("gift1").
		WithTitle("Gift 1").
		WithSlug("gift-1").
		WithPrice(mustTon(t, "1.5")).
		Build()
	if err != nil {
		t.Fatalf("expected NewStakedGiftBuilder without error, got %v", err)
	}
	stake, err := dueldomain.NewStakeBuilder(2).WithGift(stakedGift).Build()
	if err != nil {
		t.Fatalf("expected NewStakeBuilder without error, got %v", err)
	}
	if err = d.PlaceStake(stake); !errors.Is(err, dueldomain.ErrParticipantNotFound) {
		t.Fatalf("expected ErrParticipantNotFound, got %v", err)
	}

	// Успешная ставка
	succeedStake, err := dueldomain.NewStakeBuilder(
		1,
	).WithGift(stakedGift).
		Build()
	if err != nil {
		t.Fatalf("expected NewStakeBuilder without error, got %v", err)
	}
	if err = d.PlaceStake(succeedStake); err != nil {
		t.Fatalf("expected PlaceStake without error, got %v", err)
	}
}

func TestRoundFlow_ThreePlayers(t *testing.T) {
	params, err := dueldomain.NewParamsBuilder().WithIsPrivate(false).
		WithMaxPlayers(dueldomain.MaxPlayers(2)).
		WithMaxGifts(dueldomain.MaxGifts(1)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	d := dueldomain.NewDuel(params)
	p1, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(1).WithPhoto("").AsCreator().Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	p2, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(2).WithPhoto("").Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	p3, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(3).WithPhoto("").Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	d.AddParticipant(p1)
	d.AddParticipant(p2)
	d.AddParticipant(p3)

	// Первый раунд
	d.StartRound([]dueldomain.TelegramUserID{p1.TelegramUserID, p2.TelegramUserID, p3.TelegramUserID})

	// Пока не оба броска — не finished
	_, finished := d.EvaluateCurrentRound()
	if finished {
		t.Error("expected finished=false before rolls")
	}

	// Первый бросок
	roll1, err := dueldomain.NewRollBuilder().WithTelegramUserID(p1.TelegramUserID).
		WithDiceValue(4).
		WithRolledAt(time.Now()).
		WithIsAutoRolled(false).
		Build()
	if err != nil {
		t.Fatalf("expected NewRollBuilder without error, got %v", err)
	}
	if err = d.AddRollToCurrentRound(roll1); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}
	_, finished = d.EvaluateCurrentRound()
	if finished {
		t.Error("expected finished=false after one roll")
	}

	// Второй бросок, p2 > p1
	roll2, err := dueldomain.NewRollBuilder().WithTelegramUserID(p2.TelegramUserID).
		WithDiceValue(5).
		WithRolledAt(time.Now()).
		WithIsAutoRolled(false).
		Build()
	if err != nil {
		t.Fatalf("expected NewRollBuilder without error, got %v", err)
	}
	if err = d.AddRollToCurrentRound(roll2); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}

	// Третий бросок, p3 > p2
	roll3, err := dueldomain.NewRollBuilder().WithTelegramUserID(p3.TelegramUserID).
		WithDiceValue(6).
		WithRolledAt(time.Now()).
		WithIsAutoRolled(false).
		Build()
	if err != nil {
		t.Fatalf("expected NewRollBuilder without error, got %v", err)
	}
	if err = d.AddRollToCurrentRound(roll3); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}
	winners, finished := d.EvaluateCurrentRound()
	if !finished {
		t.Fatal("expected finished=true after two rolls")
	}
	if len(winners) != 1 || winners[0] != p3.TelegramUserID {
		t.Errorf("expected winner %d, got %v", p3.TelegramUserID, winners)
	}
}

func TestRoundFlow_TieThenResolve(t *testing.T) {
	params, err := dueldomain.NewParamsBuilder().WithIsPrivate(false).
		WithMaxPlayers(dueldomain.MaxPlayers(2)).
		WithMaxGifts(dueldomain.MaxGifts(1)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	d := dueldomain.NewDuel(params)
	p1, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(1).WithPhoto("").AsCreator().Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	p2, err := dueldomain.NewParticipantBuilder().WithTelegramUserID(2).WithPhoto("").Build()
	if err != nil {
		t.Fatalf("expected NewParticipantBuilder without error, got %v", err)
	}
	d.AddParticipant(p1)
	d.AddParticipant(p2)

	// Первый раунд
	d.StartRound([]dueldomain.TelegramUserID{p1.TelegramUserID, p2.TelegramUserID})
	roll1, err := dueldomain.NewRollBuilder().WithTelegramUserID(p1.TelegramUserID).
		WithDiceValue(3).
		WithRolledAt(time.Now()).
		WithIsAutoRolled(false).
		Build()
	if err != nil {
		t.Fatalf("expected NewRollBuilder without error, got %v", err)
	}
	roll2, err := dueldomain.NewRollBuilder().WithTelegramUserID(p2.TelegramUserID).
		WithDiceValue(3).
		WithRolledAt(time.Now()).
		WithIsAutoRolled(false).
		Build()
	if err != nil {
		t.Fatalf("expected NewRollBuilder without error, got %v", err)
	}
	if err = d.AddRollToCurrentRound(roll1); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}
	if err = d.AddRollToCurrentRound(roll2); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}
	winners, finished := d.EvaluateCurrentRound()
	if !finished || len(winners) != 2 {
		t.Fatalf("expected tie between two, got finished=%v winners=%v", finished, winners)
	}

	// Запускаем новый раунд только для tied
	d.StartRound(winners)
	round2 := d.Rounds[len(d.Rounds)-1]
	if len(round2.Participants) != 2 ||
		round2.Participants[0] != p1.TelegramUserID ||
		round2.Participants[1] != p2.TelegramUserID {
		t.Errorf("expected new round with the same two, got %v", round2.Participants)
	}

	// Решающий бросок
	roll3, err := dueldomain.NewRollBuilder().WithTelegramUserID(p1.TelegramUserID).
		WithDiceValue(6).
		WithRolledAt(time.Now()).
		WithIsAutoRolled(false).
		Build()
	if err != nil {
		t.Fatalf("expected NewRollBuilder without error, got %v", err)
	}
	roll4, err := dueldomain.NewRollBuilder().WithTelegramUserID(p2.TelegramUserID).
		WithDiceValue(2).
		WithRolledAt(time.Now()).
		WithIsAutoRolled(false).
		Build()
	if err != nil {
		t.Fatalf("expected NewRollBuilder without error, got %v", err)
	}
	if err = d.AddRollToCurrentRound(roll3); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}
	if err = d.AddRollToCurrentRound(roll4); err != nil {
		t.Fatalf("AddRollToCurrentRound: %v", err)
	}
	finalWinners, finished := d.EvaluateCurrentRound()
	if !finished || len(finalWinners) != 1 || finalWinners[0] != p1.TelegramUserID {
		t.Errorf("expected winner %d, got %v", p1.TelegramUserID, finalWinners)
	}
}

func TestCompleteAndCancel(t *testing.T) {
	params, err := dueldomain.NewParamsBuilder().WithIsPrivate(false).
		WithMaxPlayers(dueldomain.MaxPlayers(2)).
		WithMaxGifts(dueldomain.MaxGifts(1)).
		Build()
	if err != nil {
		t.Fatalf("expected NewParamsBuilder without error, got %v", err)
	}
	d := dueldomain.NewDuel(params)

	// Cancel в waiting
	if err = d.Cancel(); err != nil {
		t.Fatalf("Cancel in waiting should work, got %v", err)
	}
	if d.Status != dueldomain.StatusCancelled {
		t.Errorf("expected StatusCancelled, got %s", d.Status)
	}

	// Complete в wrong status
	if err = d.Complete(1); !errors.Is(err, dueldomain.ErrDuelNotInProgress) {
		t.Errorf("expected ErrDuelNotInProgress, got %v", err)
	}

	// Переводим в in_progress и Complete
	d2 := dueldomain.NewDuel(params)
	d2.Status = dueldomain.StatusInProgress
	if err = d2.Complete(42); err != nil {
		t.Fatalf("Complete in in_progress should work, got %v", err)
	}
	if d2.Status != dueldomain.StatusCompleted || d2.WinnerID == nil || *d2.WinnerID != 42 {
		t.Errorf(
			"expected completed duel with winner 42, got status=%s winner=%v",
			d2.Status,
			d2.WinnerID,
		)
	}
}
