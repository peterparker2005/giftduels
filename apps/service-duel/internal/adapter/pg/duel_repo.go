package pg

import (
	"context"
	"time"

	"github.com/ccoveille/go-safecast"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg/sqlc"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/peterparker2005/giftduels/packages/shared"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
)

type duelRepository struct {
	q      *sqlc.Queries
	logger *logger.Logger
}

func NewDuelRepository(pool *pgxpool.Pool, logger *logger.Logger) duelDomain.Repository {
	return &duelRepository{q: sqlc.New(pool), logger: logger}
}

func (r *duelRepository) WithTx(tx pgx.Tx) duelDomain.Repository {
	return &duelRepository{q: r.q.WithTx(tx), logger: r.logger}
}

func (r *duelRepository) CreateDuel(
	ctx context.Context,
	params duelDomain.CreateDuelParams,
) (duelDomain.ID, error) {
	duel, err := r.q.CreateDuel(ctx, sqlc.CreateDuelParams{
		IsPrivate:  params.Params.IsPrivate,
		MaxPlayers: int32(params.Params.MaxPlayers),
		MaxGifts:   int32(params.Params.MaxGifts),
	})
	if err != nil {
		return "", err
	}

	return duelDomain.NewID(duel.ID.String())
}

func (r *duelRepository) GetDuelByID(
	ctx context.Context,
	id duelDomain.ID,
) (*duelDomain.Duel, error) {
	pgDuelID, err := pgUUID(id.String())
	if err != nil {
		r.logger.Error("failed to get duel by id", zap.Error(err))
		return nil, err
	}
	sqlcDuel, err := r.q.GetDuelByID(ctx, pgDuelID)
	if err != nil {
		r.logger.Error("failed to get duel by id", zap.Error(err))
		return nil, err
	}

	duel, err := mapDuel(&sqlcDuel)
	if err != nil {
		r.logger.Error("failed to map duel", zap.Error(err))
		return nil, err
	}

	// Load related data
	if err := r.loadDuelRelatedData(ctx, duel); err != nil {
		r.logger.Error("failed to load duel related data", zap.Error(err))
		return nil, err
	}

	return duel, nil
}

func (r *duelRepository) GetDuelList(
	ctx context.Context,
	pageRequest *shared.PageRequest,
) ([]*duelDomain.Duel, int64, error) {
	total, err := r.q.GetVisibleDuelsCount(ctx)
	if err != nil {
		return nil, 0, err
	}

	sqlcDuels, err := r.q.GetDuels(ctx, sqlc.GetDuelsParams{
		Limit:  pageRequest.PageSize(),
		Offset: pageRequest.Offset(),
	})
	if err != nil {
		return nil, 0, err
	}

	duels := make([]*duelDomain.Duel, len(sqlcDuels))
	for i, sqlcDuel := range sqlcDuels {
		duel, err := mapDuel(&sqlcDuel)
		if err != nil {
			return nil, 0, err
		}

		// Load related data
		if err := r.loadDuelRelatedData(ctx, duel); err != nil {
			return nil, 0, err
		}

		duels[i] = duel
	}

	return duels, total, nil
}

func (r *duelRepository) loadDuelRelatedData(ctx context.Context, duel *duelDomain.Duel) error {
	duelID, err := pgUUID(duel.ID.String())
	if err != nil {
		return err
	}

	// Load participants
	sqlcParticipants, err := r.q.GetDuelParticipants(ctx, duelID)
	if err != nil {
		return err
	}
	duel.Participants = make([]duelDomain.Participant, len(sqlcParticipants))
	for i, p := range sqlcParticipants {
		telegramUserID, err := duelDomain.NewTelegramUserID(p.TelegramUserID)
		if err != nil {
			return err
		}
		duel.Participants[i] = duelDomain.Participant{
			TelegramUserID: telegramUserID,
			IsCreator:      p.IsCreator,
		}
	}

	// Load stakes
	sqlcStakes, err := r.q.GetDuelStakes(ctx, duelID)
	if err != nil {
		return err
	}
	duel.Stakes = make([]duelDomain.Stake, len(sqlcStakes))
	for i, s := range sqlcStakes {
		telegramUserID, err := duelDomain.NewTelegramUserID(s.TelegramUserID)
		if err != nil {
			return err
		}

		stakeValueStr, err := fromPgNumeric(s.StakeValue)
		if err != nil {
			return err
		}
		stakeValue, err := tonamount.NewTonAmountFromString(stakeValueStr)
		if err != nil {
			return err
		}

		duel.Stakes[i] = duelDomain.Stake{
			TelegramUserID: telegramUserID,
			Gift: duelDomain.NewStakedGift(
				s.GiftID.String(),
				"",  // Title will be filled by service layer
				"",  // Slug will be filled by service layer
				nil, // Price will be filled by service layer
			),
			StakeValue: stakeValue,
		}
	}

	// Load rounds and rolls
	sqlcRounds, err := r.q.GetDuelRounds(ctx, duelID)
	if err != nil {
		return err
	}
	duel.Rounds = make([]duelDomain.Round, len(sqlcRounds))
	for i, round := range sqlcRounds {
		duel.Rounds[i] = duelDomain.Round{
			RoundNumber:  int(round.RoundNumber),
			Participants: make([]duelDomain.TelegramUserID, 0), // Will be filled from participants
			Rolls:        make([]duelDomain.Roll, 0),
		}
	}

	// Load rolls for all rounds
	sqlcRolls, err := r.q.GetDuelRolls(ctx, duelID)
	if err != nil {
		return err
	}

	// Group rolls by round number
	rollsByRound := make(map[int32][]sqlc.DuelRoll)
	for _, roll := range sqlcRolls {
		rollsByRound[roll.RoundNumber] = append(rollsByRound[roll.RoundNumber], roll)
	}

	// Fill rolls for each round
	for i, round := range duel.Rounds {
		roundRolls := rollsByRound[int32(round.RoundNumber)]
		duel.Rounds[i].Rolls = make([]duelDomain.Roll, len(roundRolls))
		for j, roll := range roundRolls {
			telegramUserID, err := duelDomain.NewTelegramUserID(roll.TelegramUserID)
			if err != nil {
				return err
			}
			duel.Rounds[i].Rolls[j] = duelDomain.Roll{
				TelegramUserID: telegramUserID,
				DiceValue:      int32(roll.DiceValue),
				RolledAt:       roll.RolledAt.Time,
				IsAutoRolled:   roll.IsAutoRolled,
			}
		}
	}

	return nil
}

func (r *duelRepository) CreateParticipant(
	ctx context.Context,
	duelID duelDomain.ID,
	participant duelDomain.Participant,
) error {
	pgDuelID, err := pgUUID(duelID.String())
	if err != nil {
		return err
	}
	_, err = r.q.CreateParticipant(ctx, sqlc.CreateParticipantParams{
		DuelID:         pgDuelID,
		TelegramUserID: participant.TelegramUserID.Int64(),
		IsCreator:      participant.IsCreator,
	})
	return err
}

func (r *duelRepository) CreateStake(
	ctx context.Context,
	duelID duelDomain.ID,
	stake duelDomain.Stake,
) error {
	stakeValue, err := pgNumeric(stake.StakeValue.String())
	if err != nil {
		return err
	}
	pgGiftID, err := pgUUID(stake.Gift.ID)
	if err != nil {
		return err
	}
	pgDuelID, err := pgUUID(duelID.String())
	if err != nil {
		return err
	}
	_, err = r.q.CreateStake(ctx, sqlc.CreateStakeParams{
		DuelID:         pgDuelID,
		TelegramUserID: stake.TelegramUserID.Int64(),
		GiftID:         pgGiftID,
		StakeValue:     stakeValue,
	})
	return err
}

func (r *duelRepository) CreateRound(
	ctx context.Context,
	duelID duelDomain.ID,
	round duelDomain.Round,
) error {
	pgDuelID, err := pgUUID(duelID.String())
	if err != nil {
		return err
	}
	roundNumber, err := safecast.ToInt32(round.RoundNumber)
	if err != nil {
		return err
	}
	_, err = r.q.CreateRound(ctx, sqlc.CreateRoundParams{
		DuelID:      pgDuelID,
		RoundNumber: roundNumber,
	})
	return err
}

func (r *duelRepository) CreateRoll(
	ctx context.Context,
	duelID duelDomain.ID,
	roll duelDomain.Roll,
) error {
	rolledAt := pgTimestamptz(roll.RolledAt)
	pgDuelID, err := pgUUID(duelID.String())
	if err != nil {
		return err
	}

	diceValue, err := safecast.ToInt16(roll.DiceValue)
	if err != nil {
		return err
	}

	_, err = r.q.CreateRoll(ctx, sqlc.CreateRollParams{
		DuelID:         pgDuelID,
		TelegramUserID: roll.TelegramUserID.Int64(),
		DiceValue:      diceValue,
		RolledAt:       rolledAt,
		IsAutoRolled:   roll.IsAutoRolled,
	})
	return err
}

func (r *duelRepository) UpdateDuelStatus(
	ctx context.Context,
	duelID duelDomain.ID,
	status duelDomain.Status,
	winnerID *duelDomain.TelegramUserID,
	completedAt *time.Time,
) error {
	pgDuelID, err := pgUUID(duelID.String())
	if err != nil {
		return err
	}
	time := pgTimestamptz(time.Now())
	if completedAt != nil {
		time = pgTimestamptz(*completedAt)
	}

	return r.q.UpdateDuelStatus(ctx, sqlc.UpdateDuelStatusParams{
		ID:                   pgDuelID,
		Status:               sqlc.NullDuelStatus{DuelStatus: sqlc.DuelStatus(status), Valid: true},
		WinnerTelegramUserID: pgInt64(winnerID.Int64()),
		CompletedAt:          time,
	})
}

func (r *duelRepository) UpdateNextRollDeadline(
	ctx context.Context,
	duelID duelDomain.ID,
	nextRollDeadline time.Time,
) error {
	pgDuelID, err := pgUUID(duelID.String())
	if err != nil {
		return err
	}
	return r.q.UpdateNextRollDeadline(ctx, sqlc.UpdateNextRollDeadlineParams{
		ID:               pgDuelID,
		NextRollDeadline: pgTimestamptz(nextRollDeadline),
	})
}

func (r *duelRepository) FindDuelByGiftID(
	ctx context.Context,
	giftID string,
) (duelDomain.ID, error) {
	pgGiftID, err := pgUUID(giftID)
	if err != nil {
		return "", err
	}
	duelID, err := r.q.FindDuelByGiftID(ctx, pgGiftID)
	if err != nil {
		return "", err
	}
	return duelDomain.NewID(duelID.String())
}
