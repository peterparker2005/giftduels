package proto

import (
	"errors"

	"github.com/ccoveille/go-safecast"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapDuel(duel *duelDomain.Duel) (*duelv1.Duel, error) {
	params, err := MapDuelParams(duel.Params)
	if err != nil {
		return nil, err
	}

	result := &duelv1.Duel{
		DuelId:        &sharedv1.DuelId{Value: duel.ID.String()},
		Params:        params,
		DisplayNumber: duel.DisplayNumber,
		CreatedAt:     timestamppb.New(duel.CreatedAt),
		UpdatedAt:     timestamppb.New(duel.UpdatedAt),
		TotalStakeValue: &sharedv1.TonAmount{
			Value: duel.TotalStakeValue().String(),
		},
		Status: MapDuelStatus(duel.Status),
	}

	minEntryPrice, maxEntryPrice, err := duel.EntryPriceRange()
	if err != nil {
		return nil, err
	}
	result.EntryPriceRange = &duelv1.EntryPriceRange{
		MinEntryPrice: &sharedv1.TonAmount{Value: minEntryPrice.String()},
		MaxEntryPrice: &sharedv1.TonAmount{Value: maxEntryPrice.String()},
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
		result.Participants[i] = MapDuelParticipant(participant)
	}

	// Map stakes
	result.Stakes = make([]*duelv1.DuelStake, len(duel.Stakes))
	for i, stake := range duel.Stakes {
		result.Stakes[i] = MapDuelStake(stake)
	}

	// Map rounds
	result.RoundsHistory = make([]*duelv1.DuelRound, len(duel.Rounds))
	for i, round := range duel.Rounds {
		round, mapErr := MapDuelRound(round)
		if mapErr != nil {
			return nil, mapErr
		}
		result.RoundsHistory[i] = round
	}

	return result, nil
}

func MapDuelParams(params duelDomain.Params) (*duelv1.DuelParams, error) {
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

func MapDuelStatus(status duelDomain.Status) duelv1.DuelStatus {
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

func MapDuelParticipant(participant duelDomain.Participant) *duelv1.DuelParticipant {
	return &duelv1.DuelParticipant{
		TelegramUserId: &sharedv1.TelegramUserId{
			Value: int64(participant.TelegramUserID),
		},
		PhotoUrl:  participant.PhotoURL,
		IsCreator: participant.IsCreator,
	}
}

func MapDuelStake(stake duelDomain.Stake) *duelv1.DuelStake {
	return &duelv1.DuelStake{
		ParticipantTelegramUserId: &sharedv1.TelegramUserId{
			Value: int64(stake.TelegramUserID),
		},
		Gift: &duelv1.StakedGift{
			GiftId: &sharedv1.GiftId{
				Value: stake.Gift.ID,
			},
			Title: stake.Gift.Title,
			Slug:  stake.Gift.Slug,
			Price: &sharedv1.TonAmount{
				Value: stake.Gift.Price.String(),
			},
		},
		StakeValue: &sharedv1.TonAmount{
			Value: stake.StakeValue().String(),
		},
	}
}

func MapDuelRound(round duelDomain.Round) (*duelv1.DuelRound, error) {
	rolls := make([]*duelv1.DuelRoll, len(round.Rolls))
	for i, roll := range round.Rolls {
		roll, err := MapDuelRoll(roll)
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

func MapDuelRoll(roll duelDomain.Roll) (*duelv1.DuelRoll, error) {
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

func MapDuelFromProto(protoDuel *duelv1.Duel) (*duelDomain.Duel, error) {
	id, err := duelDomain.NewID(protoDuel.GetDuelId().GetValue())
	if err != nil {
		return nil, err
	}

	params, err := MapDuelParamsFromProto(protoDuel.GetParams())
	if err != nil {
		return nil, err
	}

	status, err := MapDuelStatusFromProto(protoDuel.GetStatus())
	if err != nil {
		return nil, err
	}

	duel := &duelDomain.Duel{
		ID:        id,
		Params:    params,
		Status:    status,
		CreatedAt: protoDuel.GetCreatedAt().AsTime(),
		UpdatedAt: protoDuel.GetUpdatedAt().AsTime(),
	}

	// Optional fields
	if protoDuel.GetWinnerTelegramUserId() != nil {
		winnerID, idErr := duelDomain.NewTelegramUserID(
			protoDuel.GetWinnerTelegramUserId().GetValue(),
		)
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
		mapped, mapErr := MapDuelParticipantFromProto(participant)
		if mapErr != nil {
			return nil, mapErr
		}
		duel.Participants[i] = mapped
	}

	// Map stakes
	duel.Stakes = make([]duelDomain.Stake, len(protoDuel.GetStakes()))
	for i, stake := range protoDuel.GetStakes() {
		mapped, mapErr := MapDuelStakeFromProto(stake)
		if mapErr != nil {
			return nil, mapErr
		}
		duel.Stakes[i] = mapped
	}

	// Map rounds
	duel.Rounds = make([]duelDomain.Round, len(protoDuel.GetRoundsHistory()))
	for i, round := range protoDuel.GetRoundsHistory() {
		mapped, mapErr := MapDuelRoundFromProto(round)
		if mapErr != nil {
			return nil, mapErr
		}
		duel.Rounds[i] = mapped
	}

	return duel, nil
}

func MapDuelParamsFromProto(protoParams *duelv1.DuelParams) (duelDomain.Params, error) {
	maxPlayers, err := duelDomain.NewMaxPlayers(protoParams.GetMaxPlayers())
	if err != nil {
		return duelDomain.Params{}, err
	}

	maxGifts, err := duelDomain.NewMaxGifts(protoParams.GetMaxGifts())
	if err != nil {
		return duelDomain.Params{}, err
	}

	return duelDomain.Params{
		IsPrivate:  protoParams.GetIsPrivate(),
		MaxPlayers: maxPlayers,
		MaxGifts:   maxGifts,
	}, nil
}

func MapDuelStatusFromProto(protoStatus duelv1.DuelStatus) (duelDomain.Status, error) {
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

func MapDuelParticipantFromProto(
	protoParticipant *duelv1.DuelParticipant,
) (duelDomain.Participant, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(
		protoParticipant.GetTelegramUserId().GetValue(),
	)
	if err != nil {
		return duelDomain.Participant{}, err
	}

	return duelDomain.Participant{
		TelegramUserID: telegramUserID,
		IsCreator:      protoParticipant.GetIsCreator(),
	}, nil
}

func MapDuelStakeFromProto(protoStake *duelv1.DuelStake) (duelDomain.Stake, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(
		protoStake.GetParticipantTelegramUserId().GetValue(),
	)
	if err != nil {
		return duelDomain.Stake{}, err
	}

	stakeValue, err := tonamount.NewTonAmountFromString(
		protoStake.GetStakeValue().GetValue(),
	)
	if err != nil {
		return duelDomain.Stake{}, err
	}

	gift, err := duelDomain.NewStakedGiftBuilder().
		WithID(protoStake.GetGift().GetGiftId().GetValue()).
		WithTitle(protoStake.GetGift().GetTitle()).
		WithSlug(protoStake.GetGift().GetSlug()).
		WithPrice(stakeValue).
		Build()
	if err != nil {
		return duelDomain.Stake{}, err
	}

	return duelDomain.Stake{
		TelegramUserID: telegramUserID,
		Gift:           gift,
	}, nil
}

func MapDuelRoundFromProto(protoRound *duelv1.DuelRound) (duelDomain.Round, error) {
	rolls := make([]duelDomain.Roll, len(protoRound.GetRolls()))
	for i, roll := range protoRound.GetRolls() {
		mapped, err := MapDuelRollFromProto(roll)
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

func MapDuelRollFromProto(protoRoll *duelv1.DuelRoll) (duelDomain.Roll, error) {
	telegramUserID, err := duelDomain.NewTelegramUserID(
		protoRoll.GetParticipantTelegramUserId().GetValue(),
	)
	if err != nil {
		return duelDomain.Roll{}, err
	}

	return duelDomain.Roll{
		TelegramUserID: telegramUserID,
		DiceValue:      protoRoll.GetDiceValue(),
		RolledAt:       protoRoll.GetRolledAt().AsTime(),
		IsAutoRolled:   protoRoll.GetIsAutoRolled(),
	}, nil
}

func MapDuelCreatedEvent(duel *duelDomain.Duel) (*duelv1.DuelCreatedEvent, error) {
	minEntryPrice, maxEntryPrice, err := duel.EntryPriceRange()
	if err != nil {
		return nil, err
	}
	participants := make([]*duelv1.DuelParticipant, 0, len(duel.Participants))
	for _, participant := range duel.Participants {
		participants = append(participants, &duelv1.DuelParticipant{
			TelegramUserId: &sharedv1.TelegramUserId{Value: participant.TelegramUserID.Int64()},
			PhotoUrl:       participant.PhotoURL,
			IsCreator:      participant.IsCreator,
		})
	}
	stakes := make([]*duelv1.DuelStake, 0, len(duel.Stakes))
	for _, stake := range duel.Stakes {
		stakes = append(stakes, &duelv1.DuelStake{
			ParticipantTelegramUserId: &sharedv1.TelegramUserId{
				Value: stake.TelegramUserID.Int64(),
			},
			Gift: &duelv1.StakedGift{
				GiftId: &sharedv1.GiftId{Value: stake.Gift.ID},
				Title:  stake.Gift.Title,
				Slug:   stake.Gift.Slug,
				Price:  &sharedv1.TonAmount{Value: stake.StakeValue().String()},
			},
			StakeValue: &sharedv1.TonAmount{Value: stake.StakeValue().String()},
		})
	}
	event := duelv1.DuelCreatedEvent{
		DuelId: &sharedv1.DuelId{Value: duel.ID.String()},
		Params: &duelv1.DuelParams{
			IsPrivate:  duel.Params.IsPrivate,
			MaxPlayers: uint32(duel.Params.MaxPlayers),
			MaxGifts:   uint32(duel.Params.MaxGifts),
		},
		CreatedAt:     timestamppb.New(duel.CreatedAt),
		DisplayNumber: duel.DisplayNumber,
		Participants:  participants,
		Stakes:        stakes,
		Status:        duelv1.DuelStatus_DUEL_STATUS_WAITING_FOR_OPPONENT,
		EntryPriceRange: &duelv1.EntryPriceRange{
			MinEntryPrice: &sharedv1.TonAmount{Value: minEntryPrice.String()},
			MaxEntryPrice: &sharedv1.TonAmount{Value: maxEntryPrice.String()},
		},
	}
	return &event, nil
}

func MapDuelJoinedEvent(
	duel *duelDomain.Duel,
	userID duelDomain.TelegramUserID,
) (*duelv1.DuelParticipantEvent, error) {
	// Находим участника
	var participant *duelv1.DuelParticipant
	for _, p := range duel.Participants {
		if p.TelegramUserID == userID {
			participant = &duelv1.DuelParticipant{
				TelegramUserId: &sharedv1.TelegramUserId{Value: p.TelegramUserID.Int64()},
				PhotoUrl:       p.PhotoURL,
				IsCreator:      p.IsCreator,
			}
			break
		}
	}
	if participant == nil {
		return nil, errors.New("participant not found")
	}

	// Находим ставки этого участника
	var userStakes []*duelv1.DuelStake
	var totalStakeValue *tonamount.TonAmount
	for _, stake := range duel.Stakes {
		if stake.TelegramUserID == userID {
			userStakes = append(userStakes, &duelv1.DuelStake{
				ParticipantTelegramUserId: &sharedv1.TelegramUserId{
					Value: stake.TelegramUserID.Int64(),
				},
				Gift: &duelv1.StakedGift{
					GiftId: &sharedv1.GiftId{Value: stake.Gift.ID},
					Title:  stake.Gift.Title,
					Slug:   stake.Gift.Slug,
					Price:  &sharedv1.TonAmount{Value: stake.StakeValue().String()},
				},
				StakeValue: &sharedv1.TonAmount{Value: stake.StakeValue().String()},
			})
			if totalStakeValue == nil {
				totalStakeValue = tonamount.Zero()
			}
			totalStakeValue = totalStakeValue.Add(stake.StakeValue())
		}
	}

	event := duelv1.DuelParticipantEvent{
		DuelId:      &sharedv1.DuelId{Value: duel.ID.String()},
		Participant: participant,
		Stakes:      userStakes,
		TotalStakeValue: &sharedv1.TonAmount{
			Value: totalStakeValue.String(),
		},
	}
	return &event, nil
}
