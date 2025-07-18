package duel

type Round struct {
	RoundNumber  int
	Participants []TelegramUserID
	Rolls        []Roll
}

func (r *Round) AddRoll(roll Roll) {
	r.Rolls = append(r.Rolls, roll)
}

type RoundBuilder struct {
	r Round
}

func NewRoundBuilder() *RoundBuilder {
	return &RoundBuilder{r: Round{}}
}

func (b *RoundBuilder) WithRoundNumber(roundNumber int) *RoundBuilder {
	b.r.RoundNumber = roundNumber
	return b
}

func (b *RoundBuilder) WithParticipants(participants []TelegramUserID) *RoundBuilder {
	b.r.Participants = participants
	return b
}

func (b *RoundBuilder) AddRoll(roll Roll) *RoundBuilder {
	b.r.Rolls = append(b.r.Rolls, roll)
	return b
}

func (b *RoundBuilder) Build() (Round, error) {
	if b.r.RoundNumber <= 0 {
		return Round{}, ErrInvalidRoundNumber
	}
	if len(b.r.Participants) == 0 {
		return Round{}, ErrNoParticipants
	}
	return b.r, nil
}
