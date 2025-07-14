package grpc

import (
	"errors"

	"github.com/ccoveille/go-safecast"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapDuel(duel *duelDomain.Duel) (*duelv1.Duel, error) {
	params, err := mapDuelParams(duel.Params)
	if err != nil {
		return nil, err
	}

	result := &duelv1.Duel{
		DuelId:    &sharedv1.DuelId{Value: duel.ID.String()},
		Params:    params,
		CreatedAt: timestamppb.New(duel.CreatedAt),
		UpdatedAt: timestamppb.New(duel.UpdatedAt),
		TotalStakeValue: &sharedv1.TonAmount{
			Value: duel.TotalStakeValue.String(),
		},
		Status: mapDuelStatus(duel.Status),
	}

	// Optional fields
	if duel.WinnerID != nil {
		result.WinnerTelegramUserId = &sharedv1.TelegramUserId{
			Value: int64(*duel.WinnerID),
		}
	}

	if duel.NextRollDeadline != nil {
		result.NextRollDeadline = timestamppb.New(*duel.NextRollDeadline)
	}

	if duel.CompletedAt != nil {
		result.CompletedAt = timestamppb.New(*duel.CompletedAt)
	}

	// Map participants
	result.Participants = make([]*duelv1.DuelParticipant, len(duel.Participants))
	for i, participant := range duel.Participants {
		result.Participants[i] = mapDuelParticipant(participant)
	}

	// Map stakes
	result.Stakes = make([]*duelv1.DuelStake, len(duel.Stakes))
	for i, stake := range duel.Stakes {
		result.Stakes[i] = mapDuelStake(stake)
	}

	// Map rounds
	result.RoundsHistory = make([]*duelv1.DuelRound, len(duel.Rounds))
	for i, round := range duel.Rounds {
		round, mapErr := mapDuelRound(round)
		if mapErr != nil {
			return nil, mapErr
		}
		result.RoundsHistory[i] = round
	}

	return result, nil
}

func mapDuelParams(params duelDomain.Params) (*duelv1.DuelParams, error) {
	maxPlayers, err := safecast.ToUint32(params.MaxPlayers)
	if err != nil {
		return nil, err
	}

	maxGifts, err := safecast.ToUint32(params.MaxGifts)
	if err != nil {
		return nil, err
	}

	return &duelv1.DuelParams{
		IsPrivate:  params.IsPrivate,
		MaxPlayers: maxPlayers,
		MaxGifts:   maxGifts,
	}, nil
}

func mapDuelStatus(status duelDomain.Status) duelv1.DuelStatus {
	switch status {
	case duelDomain.StatusWaitingForOpponent:
		return duelv1.DuelStatus_DUEL_STATUS_WAITING_FOR_OPPONENT
	case duelDomain.StatusInProgress:
		return duelv1.DuelStatus_DUEL_STATUS_IN_PROGRESS
	case duelDomain.StatusCompleted:
		return duelv1.DuelStatus_DUEL_STATUS_COMPLETED
	case duelDomain.StatusCancelled:
		return duelv1.DuelStatus_DUEL_STATUS_CANCELLED
	default:
		return duelv1.DuelStatus_DUEL_STATUS_UNSPECIFIED
	}
}

func mapDuelParticipant(participant duelDomain.Participant) *duelv1.DuelParticipant {
	return &duelv1.DuelParticipant{
		TelegramUserId: &sharedv1.TelegramUserId{
			Value: int64(participant.TelegramUserID),
		},
		IsCreator: participant.IsCreator,
	}
}

func mapDuelStake(stake duelDomain.Stake) *duelv1.DuelStake {
	return &duelv1.DuelStake{
		ParticipantTelegramUserId: &sharedv1.TelegramUserId{
			Value: int64(stake.TelegramUserID),
		},
		// TODO: Map Gift field using stake.GiftID
		// This requires fetching gift data from gift service
		Gift: nil,
		StakeValue: &sharedv1.TonAmount{
			Value: stake.StakeValue.String(),
		},
	}
}

func mapDuelRound(round duelDomain.Round) (*duelv1.DuelRound, error) {
	rolls := make([]*duelv1.DuelRoll, len(round.Rolls))
	for i, roll := range round.Rolls {
		roll, err := mapDuelRoll(roll)
		if err != nil {
			return nil, err
		}
		rolls[i] = roll
	}

	roundNumber, err := safecast.ToInt32(round.RoundNumber)
	if err != nil {
		return nil, err
	}

	duelRound := &duelv1.DuelRound{
		RoundNumber: roundNumber,
		Rolls:       rolls,
	}

	return duelRound, nil
}

func mapDuelRoll(roll duelDomain.Roll) (*duelv1.DuelRoll, error) {
	diceValue, err := safecast.ToInt32(roll.DiceValue)
	if err != nil {
		return nil, err
	}

	return &duelv1.DuelRoll{
		ParticipantTelegramUserId: &sharedv1.TelegramUserId{
			Value: int64(roll.TelegramUserID),
		},
		DiceValue:    diceValue,
		RolledAt:     timestamppb.New(roll.RolledAt),
		IsAutoRolled: roll.IsAutoRolled,
	}, nil
}

// Reverse mappers: proto -> domain

func mapDuelFromProto(protoDuel *duelv1.Duel) (*duelDomain.Duel, error) {
	id, err := duelDomain.NewID(protoDuel.GetDuelId().GetValue())
	if err != nil {
		return nil, err
	}

	params, err := mapDuelParamsFromProto(protoDuel.GetParams())
	if err != nil {
		return nil, err
	}

	totalStakeValue, err := tonamount.NewTonAmountFromString(protoDuel.GetTotalStakeValue().GetValue())
	if err != nil {
		return nil, err
	}

	status, err := mapDuelStatusFromProto(protoDuel.GetStatus())
	if err != nil {
		return nil, err
	}

	duel := &duelDomain.Duel{
		ID:              id,
		Params:          params,
		Status:          status,
		CreatedAt:       protoDuel.GetCreatedAt().AsTime(),
		UpdatedAt:       protoDuel.GetUpdatedAt().AsTime(),
		TotalStakeValue: totalStakeValue,
	}

	// Optional fields
	if protoDuel.GetWinnerTelegramUserId() != nil {
		winnerID, idErr := duelDomain.NewTelegramUserID(protoDuel.GetWinnerTelegramUserId().GetValue())
		if idErr != nil {
			return nil, idErr
		}
		duel.WinnerID = &winnerID
	}

	if protoDuel.GetNextRollDeadline() != nil {
		deadline := protoDuel.GetNextRollDeadline().AsTime()
		duel.NextRollDeadline = &deadline
	}

	if protoDuel.GetCompletedAt() != nil {
		completedAt := protoDuel.GetCompletedAt().AsTime()
		duel.CompletedAt = &completedAt
	}

	// Map participants
	duel.Participants = make([]duelDomain.Participant, len(protoDuel.GetParticipants()))
	for i, participant := range protoDuel.GetParticipants() {
		mapped, mapErr := mapDuelParticipantFromProto(participant)
		if mapErr != nil {
			return nil, mapErr
		}
		duel.Participants[i] = mapped
	}

	// Map stakes
	duel.Stakes = make([]duelDomain.Stake, len(protoDuel.GetStakes()))
	for i, stake := range protoDuel.GetStakes() {
		mapped, mapErr := mapDuelStakeFromProto(stake)
		if mapErr != nil {
			return nil, mapErr
		}
		duel.Stakes[i] = mapped
	}

	// Map rounds
	duel.Rounds = make([]duelDomain.Round, len(protoDuel.GetRoundsHistory()))
	for i, round := range protoDuel.GetRoundsHistory() {
		mapped, mapErr := mapDuelRoundFromProto(round)
		if mapErr != nil {
			return nil, mapErr
		}
		duel.Rounds[i] = mapped
	}

	return duel, nil
}

func mapDuelParamsFromProto(protoParams *duelv1.DuelParams) (duelDomain.Params, error) {
	maxPlayersInt, err := safecast.ToInt32(protoParams.GetMaxPlayers())
	if err != nil {
		return duelDomain.Params{}, err
	}

	maxGiftsInt, err := safecast.ToInt32(protoParams.GetMaxGifts())
	if err != nil {
		return duelDomain.Params{}, err
	}

	maxPlayers, err := duelDomain.NewMaxPlayers(maxPlayersInt)
	if err != nil {
		return duelDomain.Params{}, err
	}

	maxGifts, err := duelDomain.NewMaxGifts(maxGiftsInt)
	if err != nil {
		return duelDomain.Params{}, err
	}

	return duelDomain.Params{
		IsPrivate:  protoParams.GetIsPrivate(),
		MaxPlayers: maxPlayers,
		MaxGifts:   maxGifts,
	}, nil
}

func mapDuelStatusFromProto(protoStatus duelv1.DuelStatus) (duelDomain.Status, error) {
	switch protoStatus {
	case duelv1.DuelStatus_DUEL_STATUS_WAITING_FOR_OPPONENT:
		return duelDomain.StatusWaitingForOpponent, nil
	case duelv1.DuelStatus_DUEL_STATUS_IN_PROGRESS:
		return duelDomain.StatusInProgress, nil
	case duelv1.DuelStatus_DUEL_STATUS_COMPLETED:
		return duelDomain.StatusCompleted, nil
	case duelv1.DuelStatus_DUEL_STATUS_CANCELLED:
		return duelDomain.StatusCancelled, nil
	case duelv1.DuelStatus_DUEL_STATUS_UNSPECIFIED:
		return "", errors.New("invalid duel status")
	default:
		return "", errors.New("invalid duel status")
	}
}

func mapDuelParticipantFromProto(protoParticipant *duelv1.DuelParticipant) (duelDomain.Participant, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(protoParticipant.GetTelegramUserId().GetValue())
	if err != nil {
		return duelDomain.Participant{}, err
	}

	return duelDomain.Participant{
		TelegramUserID: telegramUserID,
		IsCreator:      protoParticipant.GetIsCreator(),
	}, nil
}

func mapDuelStakeFromProto(protoStake *duelv1.DuelStake) (duelDomain.Stake, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(protoStake.GetParticipantTelegramUserId().GetValue())
	if err != nil {
		return duelDomain.Stake{}, err
	}

	stakeValue, err := tonamount.NewTonAmountFromString(protoStake.GetStakeValue().GetValue())
	if err != nil {
		return duelDomain.Stake{}, err
	}

	return duelDomain.Stake{
		TelegramUserID: telegramUserID,
		// TODO: Extract GiftID from protobuf Gift if available
		GiftID:     "", // For now, empty - would need to map from protobuf gift
		StakeValue: stakeValue,
	}, nil
}

func mapDuelRoundFromProto(protoRound *duelv1.DuelRound) (duelDomain.Round, error) {
	rolls := make([]duelDomain.Roll, len(protoRound.GetRolls()))
	for i, roll := range protoRound.GetRolls() {
		mapped, err := mapDuelRollFromProto(roll)
		if err != nil {
			return duelDomain.Round{}, err
		}
		rolls[i] = mapped
	}

	return duelDomain.Round{
		RoundNumber: int(protoRound.GetRoundNumber()),
		Rolls:       rolls,
	}, nil
}

func mapDuelRollFromProto(protoRoll *duelv1.DuelRoll) (duelDomain.Roll, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(protoRoll.GetParticipantTelegramUserId().GetValue())
	if err != nil {
		return duelDomain.Roll{}, err
	}

	return duelDomain.Roll{
		TelegramUserID: telegramUserID,
		DiceValue:      int(protoRoll.GetDiceValue()),
		RolledAt:       protoRoll.GetRolledAt().AsTime(),
		IsAutoRolled:   protoRoll.GetIsAutoRolled(),
	}, nil
}
