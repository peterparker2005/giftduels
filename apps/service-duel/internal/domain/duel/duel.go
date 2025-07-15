package duel

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

type ID string

func (id ID) String() string {
	return string(id)
}

type TelegramUserID int64

func (id TelegramUserID) Int64() int64 {
	return int64(id)
}

func (id TelegramUserID) String() string {
	return strconv.FormatInt(id.Int64(), 10)
}

type Status string

const (
	StatusWaitingForOpponent Status = "waiting_for_opponent"
	StatusInProgress         Status = "in_progress"
	StatusCompleted          Status = "completed"
	StatusCancelled          Status = "cancelled"
)

const (
	TimeoutBeforeFirstRound = 60 * time.Second
	TimeoutAfterFirstRound  = 30 * time.Second
)

type Duel struct {
	ID               ID
	DisplayNumber    int64
	Params           Params
	WinnerID         *TelegramUserID
	NextRollDeadline *time.Time
	Status           Status
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      *time.Time

	Participants []Participant
	Stakes       []Stake
	Rounds       []Round
}

func NewDuel(params Params) *Duel {
	now := time.Now()

	return &Duel{
		ID:        ID(uuid.New().String()),
		Params:    params,
		Status:    StatusWaitingForOpponent,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (d *Duel) Join(p Participant) error {
	// check for duplicate
	for _, ex := range d.Participants {
		if ex.TelegramUserID == p.TelegramUserID {
			return ErrAlreadyJoined
		}
	}
	// check max players limit
	if len(d.Participants) >= int(d.Params.MaxPlayers) {
		return ErrMaxPlayersExceeded
	}
	// add participant
	d.Participants = append(d.Participants, p)
	d.UpdatedAt = time.Now()
	return nil
}

func (d *Duel) PlaceStake(s Stake) error {
	found := false
	for _, p := range d.Participants {
		if p.TelegramUserID == s.TelegramUserID {
			found = true
			break
		}
	}
	if !found {
		return ErrParticipantNotFound
	}
	d.Stakes = append(d.Stakes, s)
	d.UpdatedAt = time.Now()
	return nil
}

func (d *Duel) StartRound(players []TelegramUserID) {
	num := len(d.Rounds) + 1
	d.Rounds = append(d.Rounds,
		NewRound(num, players),
	)
	d.UpdatedAt = time.Now()
}

// AddRollToCurrentRound adds a roll to the current round, checking that the player is in the list and has not rolled yet.
func (d *Duel) AddRollToCurrentRound(roll Roll) error {
	if len(d.Rounds) == 0 {
		return ErrNoRoundStarted
	}
	r := &d.Rounds[len(d.Rounds)-1]

	// 1) check that the player is in r.Participants
	var ok bool
	for _, p := range r.Participants {
		if p == roll.TelegramUserID {
			ok = true
			break
		}
	}
	if !ok {
		return ErrParticipantNotFound
	}

	// 2) check that the player has not rolled in this round
	for _, ex := range r.Rolls {
		if ex.TelegramUserID == roll.TelegramUserID {
			return ErrAlreadyRolled
		}
	}

	// 3) add roll
	r.Rolls = append(r.Rolls, roll)
	d.UpdatedAt = time.Now()
	return nil
}

// EvaluateCurrentRound returns all players who rolled the max value, and a flag indicating if the round is finished
// returns winners and a flag indicating if the round is finished.
func (d *Duel) EvaluateCurrentRound() ([]TelegramUserID, bool) {
	winners := []TelegramUserID{}
	r := d.Rounds[len(d.Rounds)-1]
	// round is not finished until there are as many rolls as participants
	if len(r.Rolls) < len(r.Participants) {
		return nil, false
	}
	// 1) find the maximum value
	maxRoll := int32(0)
	for _, rl := range r.Rolls {
		if rl.DiceValue > maxRoll {
			maxRoll = rl.DiceValue
		}
	}
	// 2) iterate over rolls and collect all who rolled max
	for _, rl := range r.Rolls {
		if rl.DiceValue == maxRoll {
			winners = append(winners, rl.TelegramUserID)
		}
	}
	return winners, true
}

func (d *Duel) Complete(winner TelegramUserID) error {
	if d.Status != StatusInProgress {
		return ErrDuelNotInProgress
	}
	now := time.Now()
	d.WinnerID = &winner
	d.Status = StatusCompleted
	d.CompletedAt = &now
	d.UpdatedAt = now
	return nil
}

func (d *Duel) Cancel() error {
	if d.Status != StatusWaitingForOpponent {
		return ErrDuelNotInProgress
	}
	now := time.Now()
	d.Status = StatusCancelled
	d.CompletedAt = &now
	d.UpdatedAt = now
	return nil
}

// TimeoutForRound returns how long to wait in the current round.
func (d *Duel) TimeoutForRound() time.Duration {
	if len(d.Rounds) == 0 {
		return TimeoutBeforeFirstRound
	}
	return TimeoutAfterFirstRound
}

func (d *Duel) TotalStakeValue() *tonamount.TonAmount {
	sum := tonamount.Zero()
	for _, s := range d.Stakes {
		sum = sum.Add(s.StakeValue)
	}
	return sum
}

type Params struct {
	IsPrivate  bool
	MaxPlayers MaxPlayers
	MaxGifts   MaxGifts
}

func NewParams(isPrivate bool, maxPlayers MaxPlayers, maxGifts MaxGifts) Params {
	return Params{
		IsPrivate:  isPrivate,
		MaxPlayers: maxPlayers,
		MaxGifts:   maxGifts,
	}
}

type Participant struct {
	TelegramUserID TelegramUserID
	PhotoURL       string
	IsCreator      bool
}

func NewParticipant(
	telegramUserID TelegramUserID,
	photoURL string,
	isCreator bool,
) Participant {
	return Participant{
		TelegramUserID: telegramUserID,
		PhotoURL:       photoURL,
		IsCreator:      isCreator,
	}
}

type Stake struct {
	TelegramUserID TelegramUserID
	Gift           StakedGift
	StakeValue     *tonamount.TonAmount
}

type StakedGift struct {
	ID    string
	Title string
	Slug  string
	Price *tonamount.TonAmount
}

func NewStakedGift(
	giftID string,
	title string,
	slug string,
	price *tonamount.TonAmount,
) StakedGift {
	if price == nil {
		price, _ = tonamount.NewTonAmountFromString("0")
	}
	return StakedGift{
		ID:    giftID,
		Title: title,
		Slug:  slug,
		Price: price,
	}
}

func NewStake(
	telegramUserID TelegramUserID,
	gift StakedGift,
	stakeValue *tonamount.TonAmount,
) Stake {
	return Stake{
		TelegramUserID: telegramUserID,
		Gift:           gift,
		StakeValue:     stakeValue,
	}
}

type Round struct {
	RoundNumber  int
	Participants []TelegramUserID
	Rolls        []Roll
}

func NewRound(roundNumber int, players []TelegramUserID) Round {
	ps := append([]TelegramUserID(nil), players...)
	return Round{RoundNumber: roundNumber, Participants: ps}
}

func (r *Round) AddRoll(roll Roll) {
	r.Rolls = append(r.Rolls, roll)
}

type Roll struct {
	TelegramUserID TelegramUserID
	DiceValue      int32
	RolledAt       time.Time
	IsAutoRolled   bool
}

func NewRoll(
	telegramUserID TelegramUserID,
	diceValue int32,
	rolledAt time.Time,
	isAutoRolled bool,
) Roll {
	return Roll{
		TelegramUserID: telegramUserID,
		DiceValue:      diceValue,
		RolledAt:       rolledAt,
		IsAutoRolled:   isAutoRolled,
	}
}
