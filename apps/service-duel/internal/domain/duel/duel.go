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

func (d *Duel) AddParticipant(p Participant) error {
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

func (d *Duel) Start() error {
	if d.Status != StatusWaitingForOpponent {
		return ErrDuelNotInProgress
	}
	// меняем статус
	d.Status = StatusInProgress
	// собираем список игроков
	var ids []TelegramUserID
	for _, p := range d.Participants {
		ids = append(ids, p.TelegramUserID)
	}
	// создаём первый раунд в памяти
	if err := d.StartRound(ids); err != nil {
		return err
	}
	// рассчитываем дедлайн первого хода
	deadline := time.Now().Add(d.TimeoutForRound())
	d.NextRollDeadline = &deadline
	d.UpdatedAt = time.Now()
	return nil
}

func (d *Duel) StartRound(players []TelegramUserID) error {
	num := int32(len(d.Rounds) + 1)
	r, err := NewRoundBuilder().WithRoundNumber(num).WithParticipants(players).Build()
	if err != nil {
		return err
	}
	d.Rounds = append(d.Rounds, r)
	d.UpdatedAt = time.Now()
	return nil
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
		sum = sum.Add(s.Gift.Price)
	}
	return sum
}

func (d *Duel) IsCompleted() bool {
	return d.Status == StatusCompleted
}

func (d *Duel) IsCancelled() bool {
	return d.Status == StatusCancelled
}

func (d *Duel) IsInProgress() bool {
	return d.Status == StatusInProgress
}

func (d *Duel) IsWaitingForOpponent() bool {
	return d.Status == StatusWaitingForOpponent
}

func (d *Duel) CurrentRound() (*Round, error) {
	if len(d.Rounds) == 0 {
		return nil, ErrNoRoundStarted
	}
	return &d.Rounds[len(d.Rounds)-1], nil
}

//nolint:revive // naming is ok
type DuelBuilder struct {
	d Duel
}

func NewDuelBuilder() *DuelBuilder {
	return &DuelBuilder{d: Duel{}}
}

func (b *DuelBuilder) WithID(id ID) *DuelBuilder {
	b.d.ID = id
	return b
}

func (b *DuelBuilder) WithDisplayNumber(num int64) *DuelBuilder {
	b.d.DisplayNumber = num
	return b
}

func (b *DuelBuilder) WithParams(params Params) *DuelBuilder {
	b.d.Params = params
	return b
}

func (b *DuelBuilder) WithStatus(status Status) *DuelBuilder {
	b.d.Status = status
	return b
}

func (b *DuelBuilder) WithWinner(w *TelegramUserID) *DuelBuilder {
	b.d.WinnerID = w
	return b
}

func (b *DuelBuilder) WithNextRollDeadline(t *time.Time) *DuelBuilder {
	b.d.NextRollDeadline = t
	return b
}

func (b *DuelBuilder) WithCreatedAt(t time.Time) *DuelBuilder {
	b.d.CreatedAt = t
	return b
}

func (b *DuelBuilder) WithUpdatedAt(t time.Time) *DuelBuilder {
	b.d.UpdatedAt = t
	return b
}

func (b *DuelBuilder) WithCompletedAt(t *time.Time) *DuelBuilder {
	b.d.CompletedAt = t
	return b
}

func (b *DuelBuilder) WithParticipants(ps []Participant) *DuelBuilder {
	b.d.Participants = ps
	return b
}

func (b *DuelBuilder) WithStakes(ss []Stake) *DuelBuilder {
	b.d.Stakes = ss
	return b
}

func (b *DuelBuilder) WithRounds(rr []Round) *DuelBuilder {
	b.d.Rounds = rr
	return b
}

func (b *DuelBuilder) validate() error {
	return nil
}

func (b *DuelBuilder) Build() (*Duel, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}
	return &b.d, nil
}
