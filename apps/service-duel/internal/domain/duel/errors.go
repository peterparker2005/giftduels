package duel

import "errors"

var (
	ErrInvalidID             = errors.New("id is required")
	ErrInvalidMaxPlayers     = errors.New("max players must be between 2 and 4")
	ErrInvalidMaxGifts       = errors.New("max gifts must be between 1 and 10")
	ErrInvalidTelegramUserID = errors.New("telegram user id must be greater than 0")
	ErrMaxPlayersExceeded    = errors.New("max players exceeded")
	ErrAlreadyJoined         = errors.New("already joined")
	ErrParticipantNotFound   = errors.New("participant not found")
	ErrDuelNotInProgress     = errors.New("duel is not in progress")
	ErrNoRoundStarted        = errors.New("no round started")
	ErrAlreadyRolled         = errors.New("already rolled")

	ErrCreatorNotFound     = errors.New("creator not found")
	ErrNoStakesFromCreator = errors.New("no stakes from creator to determine range")
	ErrStakeOutOfRange     = errors.New("stake is out of allowed entry range")

	ErrInvalidDiceValue   = errors.New("dice value must be between 1 and 6")
	ErrNoParticipants     = errors.New("no participants")
	ErrNilStakeValue      = errors.New("stake value cannot be nil")
	ErrInvalidRoundNumber = errors.New("round number must be greater than 0")
	ErrEmptyGiftID        = errors.New("gift id cannot be empty")
	ErrEmptyGiftTitle     = errors.New("gift title cannot be empty")
)
