package command

import "errors"

var (
	// Create duel errors.

	ErrCreateDuel = errors.New("failed to create duel")

	// Join duel errors.

	ErrJoinDuel          = errors.New("failed to join duel")
	ErrDuelNotFound      = errors.New("duel not found")
	ErrDuelFull          = errors.New("duel is full")
	ErrAlreadyJoined     = errors.New("already joined this duel")
	ErrStakeOutOfRange   = errors.New("stake is out of allowed entry range")
	ErrGiftStakingFailed = errors.New("failed to stake gift")
	ErrInvalidGiftPrice  = errors.New("invalid gift price")

	// Auto roll errors.

	ErrAutoRoll                       = errors.New("failed to auto roll")
	ErrRollDiceFailed                 = errors.New("failed to roll dice")
	ErrNoCurrentRound                 = errors.New("no current round")
	ErrRoundEvaluationFailed          = errors.New("failed to evaluate round")
	ErrStartNewRoundFailed            = errors.New("failed to start new round")
	ErrCompleteDuelFailed             = errors.New("failed to complete duel")
	ErrSendDuelCompletedMessageFailed = errors.New("failed to send duel completed message")

	// General command errors.

	ErrTransactionFailed  = errors.New("transaction failed")
	ErrDatabaseOperation  = errors.New("database operation failed")
	ErrInvalidParticipant = errors.New("invalid participant")
	ErrInvalidStake       = errors.New("invalid stake")
	ErrPublishEventFailed = errors.New("failed to publish event")
)

// Create duel error checkers.

func IsCreateDuel(err error) bool {
	return errors.Is(err, ErrCreateDuel)
}

// Join duel error checkers.

func IsJoinDuel(err error) bool {
	return errors.Is(err, ErrJoinDuel)
}

func IsDuelNotFound(err error) bool {
	return errors.Is(err, ErrDuelNotFound)
}

func IsDuelFull(err error) bool {
	return errors.Is(err, ErrDuelFull)
}

func IsAlreadyJoined(err error) bool {
	return errors.Is(err, ErrAlreadyJoined)
}

func IsStakeOutOfRange(err error) bool {
	return errors.Is(err, ErrStakeOutOfRange)
}

func IsGiftStakingFailed(err error) bool {
	return errors.Is(err, ErrGiftStakingFailed)
}

func IsInvalidGiftPrice(err error) bool {
	return errors.Is(err, ErrInvalidGiftPrice)
}

// Auto roll error checkers.

func IsAutoRoll(err error) bool {
	return errors.Is(err, ErrAutoRoll)
}

func IsRollDiceFailed(err error) bool {
	return errors.Is(err, ErrRollDiceFailed)
}

func IsNoCurrentRound(err error) bool {
	return errors.Is(err, ErrNoCurrentRound)
}

func IsRoundEvaluationFailed(err error) bool {
	return errors.Is(err, ErrRoundEvaluationFailed)
}

func IsStartNewRoundFailed(err error) bool {
	return errors.Is(err, ErrStartNewRoundFailed)
}

func IsCompleteDuelFailed(err error) bool {
	return errors.Is(err, ErrCompleteDuelFailed)
}

func IsSendDuelCompletedMessageFailed(err error) bool {
	return errors.Is(err, ErrSendDuelCompletedMessageFailed)
}

// General error checkers.

func IsTransactionFailed(err error) bool {
	return errors.Is(err, ErrTransactionFailed)
}

func IsDatabaseOperation(err error) bool {
	return errors.Is(err, ErrDatabaseOperation)
}

func IsInvalidParticipant(err error) bool {
	return errors.Is(err, ErrInvalidParticipant)
}

func IsInvalidStake(err error) bool {
	return errors.Is(err, ErrInvalidStake)
}

func IsPublishEventFailed(err error) bool {
	return errors.Is(err, ErrPublishEventFailed)
}
