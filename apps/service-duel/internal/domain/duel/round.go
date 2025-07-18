package duel

type Round struct {
	RoundNumber  int32
	Participants []TelegramUserID
	Rolls        []Roll
}

func (r *Round) AddRoll(roll Roll) {
	r.Rolls = append(r.Rolls, roll)
}

func (r *Round) HasRolled(participant TelegramUserID) bool {
	for _, roll := range r.Rolls {
		if roll.TelegramUserID == participant {
			return true
		}
	}
	return false
}

type RoundBuilder struct {
	r Round
}

func NewRoundBuilder() *RoundBuilder {
	return &RoundBuilder{r: Round{}}
}

func (b *RoundBuilder) WithRoundNumber(roundNumber int32) *RoundBuilder {
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
