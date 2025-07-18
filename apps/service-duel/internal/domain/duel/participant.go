package duel

type Participant struct {
	TelegramUserID TelegramUserID
	PhotoURL       string
	IsCreator      bool
}

type ParticipantBuilder struct {
	p Participant
}

func NewParticipantBuilder() *ParticipantBuilder {
	return &ParticipantBuilder{p: Participant{}}
}

func (b *ParticipantBuilder) WithTelegramUserID(id TelegramUserID) *ParticipantBuilder {
	b.p.TelegramUserID = id
	return b
}

func (b *ParticipantBuilder) WithPhoto(url string) *ParticipantBuilder {
	b.p.PhotoURL = url
	return b
}

func (b *ParticipantBuilder) AsCreator() *ParticipantBuilder {
	b.p.IsCreator = true
	return b
}

func (b *ParticipantBuilder) validate() error {
	if b.p.TelegramUserID <= 0 {
		return ErrInvalidTelegramUserID
	}
	return nil
}

func (b *ParticipantBuilder) Build() (Participant, error) {
	if err := b.validate(); err != nil {
		return Participant{}, err
	}
	return b.p, nil
}
