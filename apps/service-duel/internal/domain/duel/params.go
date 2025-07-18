package duel

func NewID(id string) (ID, error) {
	if id == "" {
		return "", ErrInvalidID
	}
	return ID(id), nil
}

func NewTelegramUserID(id int64) (TelegramUserID, error) {
	if id <= 0 {
		return 0, ErrInvalidTelegramUserID
	}
	return TelegramUserID(id), nil
}

type Params struct {
	IsPrivate  bool
	MaxPlayers MaxPlayers
	MaxGifts   MaxGifts
}

type ParamsBuilder struct {
	Params
}

func NewParamsBuilder() *ParamsBuilder {
	return &ParamsBuilder{}
}

func (b *ParamsBuilder) WithIsPrivate(isPrivate bool) *ParamsBuilder {
	b.IsPrivate = isPrivate
	return b
}

func (b *ParamsBuilder) WithMaxPlayers(maxPlayers MaxPlayers) *ParamsBuilder {
	b.MaxPlayers = maxPlayers
	return b
}

func (b *ParamsBuilder) WithMaxGifts(maxGifts MaxGifts) *ParamsBuilder {
	b.MaxGifts = maxGifts
	return b
}

func (b *ParamsBuilder) validate() error {
	if b.MaxPlayers < MinPlayers || b.MaxPlayers > MaxPlayersLimit {
		return ErrInvalidMaxPlayers
	}
	if b.MaxGifts < MinGifts || b.MaxGifts > MaxGiftsLimit {
		return ErrInvalidMaxGifts
	}
	return nil
}

func (b *ParamsBuilder) Build() (Params, error) {
	if err := b.validate(); err != nil {
		return Params{}, err
	}
	return b.Params, nil
}
