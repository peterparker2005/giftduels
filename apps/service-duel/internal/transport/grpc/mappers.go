package grpc

import (
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapDuel(duel *duelDomain.Duel) *duelv1.Duel {
	result := &duelv1.Duel{
		DuelId:    &sharedv1.DuelId{Value: duel.ID.String()},
		Params:    mapDuelParams(duel.Params),
		CreatedAt: timestamppb.New(duel.CreatedAt),
		UpdatedAt: timestamppb.New(duel.UpdatedAt),
		TotalStakeValue: &sharedv1.TonAmount{
			Value: duel.TotalStakeValue,
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
		result.RoundsHistory[i] = mapDuelRound(round)
	}

	return result
}

func mapDuelParams(params duelDomain.DuelParams) *duelv1.DuelParams {
	return &duelv1.DuelParams{
		IsPrivate:  params.IsPrivate,
		MaxPlayers: uint32(params.MaxPlayers),
		MaxGifts:   uint32(params.MaxGifts),
	}
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
			Value: stake.StakeValue,
		},
	}
}

func mapDuelRound(round duelDomain.Round) *duelv1.DuelRound {
	rolls := make([]*duelv1.DuelRoll, len(round.Rolls))
	for i, roll := range round.Rolls {
		rolls[i] = mapDuelRoll(roll)
	}

	return &duelv1.DuelRound{
		RoundNumber: int32(round.RoundNumber),
		Rolls:       rolls,
	}
}

func mapDuelRoll(roll duelDomain.Roll) *duelv1.DuelRoll {
	return &duelv1.DuelRoll{
		ParticipantTelegramUserId: &sharedv1.TelegramUserId{
			Value: int64(roll.TelegramUserID),
		},
		DiceValue:    int32(roll.DiceValue),
		RolledAt:     timestamppb.New(roll.RolledAt),
		IsAutoRolled: roll.IsAutoRolled,
	}
}

// Reverse mappers: proto -> domain

func mapDuelFromProto(protoDuel *duelv1.Duel) (*duelDomain.Duel, error) {
	id, err := duelDomain.NewID(protoDuel.DuelId.Value)
	if err != nil {
		return nil, err
	}

	params, err := mapDuelParamsFromProto(protoDuel.Params)
	if err != nil {
		return nil, err
	}

	duel := &duelDomain.Duel{
		ID:              id,
		Params:          params,
		Status:          mapDuelStatusFromProto(protoDuel.Status),
		CreatedAt:       protoDuel.CreatedAt.AsTime(),
		UpdatedAt:       protoDuel.UpdatedAt.AsTime(),
		TotalStakeValue: protoDuel.TotalStakeValue.Value,
	}

	// Optional fields
	if protoDuel.WinnerTelegramUserId != nil {
		winnerID, err := duelDomain.NewTelegramUserID(protoDuel.WinnerTelegramUserId.Value)
		if err != nil {
			return nil, err
		}
		duel.WinnerID = &winnerID
	}

	if protoDuel.NextRollDeadline != nil {
		deadline := protoDuel.NextRollDeadline.AsTime()
		duel.NextRollDeadline = &deadline
	}

	if protoDuel.CompletedAt != nil {
		completedAt := protoDuel.CompletedAt.AsTime()
		duel.CompletedAt = &completedAt
	}

	// Map participants
	duel.Participants = make([]duelDomain.Participant, len(protoDuel.Participants))
	for i, participant := range protoDuel.Participants {
		mapped, err := mapDuelParticipantFromProto(participant)
		if err != nil {
			return nil, err
		}
		duel.Participants[i] = mapped
	}

	// Map stakes
	duel.Stakes = make([]duelDomain.Stake, len(protoDuel.Stakes))
	for i, stake := range protoDuel.Stakes {
		mapped, err := mapDuelStakeFromProto(stake)
		if err != nil {
			return nil, err
		}
		duel.Stakes[i] = mapped
	}

	// Map rounds
	duel.Rounds = make([]duelDomain.Round, len(protoDuel.RoundsHistory))
	for i, round := range protoDuel.RoundsHistory {
		mapped, err := mapDuelRoundFromProto(round)
		if err != nil {
			return nil, err
		}
		duel.Rounds[i] = mapped
	}

	return duel, nil
}

func mapDuelParamsFromProto(protoParams *duelv1.DuelParams) (duelDomain.DuelParams, error) {
	maxPlayers, err := duelDomain.NewMaxPlayers(int32(protoParams.MaxPlayers))
	if err != nil {
		return duelDomain.DuelParams{}, err
	}

	maxGifts, err := duelDomain.NewMaxGifts(int32(protoParams.MaxGifts))
	if err != nil {
		return duelDomain.DuelParams{}, err
	}

	return duelDomain.DuelParams{
		IsPrivate:  protoParams.IsPrivate,
		MaxPlayers: maxPlayers,
		MaxGifts:   maxGifts,
	}, nil
}

func mapDuelStatusFromProto(protoStatus duelv1.DuelStatus) duelDomain.Status {
	switch protoStatus {
	case duelv1.DuelStatus_DUEL_STATUS_WAITING_FOR_OPPONENT:
		return duelDomain.StatusWaitingForOpponent
	case duelv1.DuelStatus_DUEL_STATUS_IN_PROGRESS:
		return duelDomain.StatusInProgress
	case duelv1.DuelStatus_DUEL_STATUS_COMPLETED:
		return duelDomain.StatusCompleted
	case duelv1.DuelStatus_DUEL_STATUS_CANCELLED:
		return duelDomain.StatusCancelled
	default:
		return duelDomain.StatusWaitingForOpponent // Default fallback
	}
}

func mapDuelParticipantFromProto(protoParticipant *duelv1.DuelParticipant) (duelDomain.Participant, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(protoParticipant.TelegramUserId.Value)
	if err != nil {
		return duelDomain.Participant{}, err
	}

	return duelDomain.Participant{
		TelegramUserID: telegramUserID,
		IsCreator:      protoParticipant.IsCreator,
	}, nil
}

func mapDuelStakeFromProto(protoStake *duelv1.DuelStake) (duelDomain.Stake, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(protoStake.ParticipantTelegramUserId.Value)
	if err != nil {
		return duelDomain.Stake{}, err
	}

	return duelDomain.Stake{
		TelegramUserID: telegramUserID,
		// TODO: Extract GiftID from protobuf Gift if available
		GiftID:     "", // For now, empty - would need to map from protobuf gift
		StakeValue: protoStake.StakeValue.Value,
	}, nil
}

func mapDuelRoundFromProto(protoRound *duelv1.DuelRound) (duelDomain.Round, error) {
	rolls := make([]duelDomain.Roll, len(protoRound.Rolls))
	for i, roll := range protoRound.Rolls {
		mapped, err := mapDuelRollFromProto(roll)
		if err != nil {
			return duelDomain.Round{}, err
		}
		rolls[i] = mapped
	}

	return duelDomain.Round{
		RoundNumber: int(protoRound.RoundNumber),
		Rolls:       rolls,
	}, nil
}

func mapDuelRollFromProto(protoRoll *duelv1.DuelRoll) (duelDomain.Roll, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(protoRoll.ParticipantTelegramUserId.Value)
	if err != nil {
		return duelDomain.Roll{}, err
	}

	return duelDomain.Roll{
		TelegramUserID: telegramUserID,
		DiceValue:      int(protoRoll.DiceValue),
		RolledAt:       protoRoll.RolledAt.AsTime(),
		IsAutoRolled:   protoRoll.IsAutoRolled,
	}, nil
}
