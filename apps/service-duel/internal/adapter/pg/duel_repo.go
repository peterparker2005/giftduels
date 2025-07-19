package pg

import (
	"context"
	"time"

	"github.com/ccoveille/go-safecast"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg/sqlc"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/peterparker2005/giftduels/packages/shared"
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
	duel *duelDomain.Duel,
) (duelDomain.ID, error) {
	pgDuel, err := r.q.CreateDuel(ctx, sqlc.CreateDuelParams{
		IsPrivate:  duel.Params.IsPrivate,
		MaxPlayers: duel.Params.MaxPlayers.Int32(),
		MaxGifts:   duel.Params.MaxGifts.Int32(),
	})
	if err != nil {
		return "", err
	}

	return duelDomain.NewID(pgDuel.ID.String())
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
	if err = r.loadDuelRelatedData(ctx, duel); err != nil {
		r.logger.Error("failed to load duel related data", zap.Error(err))
		return nil, err
	}

	return duel, nil
}

func (r *duelRepository) GetDuelList(
	ctx context.Context,
	pageRequest *shared.PageRequest,
	filter *duelDomain.Filter,
	telegramUserID *duelDomain.TelegramUserID,
) ([]*duelDomain.Duel, int64, error) {
	var total int64
	var sqlcDuels []sqlc.Duel
	var err error

	// Apply filter based on filter type
	switch filter.FilterType {
	case duelDomain.FilterTypeAll:
		total, err = r.q.GetVisibleDuelsCount(ctx)
		if err != nil {
			r.logger.Error("failed to get visible duels count", zap.Error(err))
			return nil, 0, err
		}
		sqlcDuels, err = r.q.GetDuels(ctx, sqlc.GetDuelsParams{
			Limit:  pageRequest.PageSize(),
			Offset: pageRequest.Offset(),
		})
		if err != nil {
			r.logger.Error("failed to get duels", zap.Error(err))
			return nil, 0, err
		}

	case duelDomain.FilterType1v1:
		total, err = r.q.Get1v1DuelsCount(ctx)
		if err != nil {
			r.logger.Error("failed to get 1v1 duels count", zap.Error(err))
			return nil, 0, err
		}
		sqlcDuels, err = r.q.Get1v1Duels(ctx, sqlc.Get1v1DuelsParams{
			Limit:  pageRequest.PageSize(),
			Offset: pageRequest.Offset(),
		})
		if err != nil {
			r.logger.Error("failed to get 1v1 duels", zap.Error(err))
			return nil, 0, err
		}

	case duelDomain.FilterTypeDailyTop:
		total, err = r.q.GetTopDuelsCount(ctx)
		if err != nil {
			r.logger.Error("failed to get top duels count", zap.Error(err))
			return nil, 0, err
		}
		sqlcDuels, err = r.q.GetTopDuels(ctx, sqlc.GetTopDuelsParams{
			Limit:  pageRequest.PageSize(),
			Offset: pageRequest.Offset(),
		})
		if err != nil {
			r.logger.Error("failed to get top duels", zap.Error(err))
			return nil, 0, err
		}

	case duelDomain.FilterTypeMyDuels:
		total, err = r.q.GetMyDuelsCount(ctx, telegramUserID.Int64())
		if err != nil {
			r.logger.Error("failed to get my duels count", zap.Error(err))
			return nil, 0, err
		}
		sqlcDuels, err = r.q.GetMyDuels(ctx, sqlc.GetMyDuelsParams{
			TelegramUserID: telegramUserID.Int64(),
			Limit:          pageRequest.PageSize(),
			Offset:         pageRequest.Offset(),
		})
		if err != nil {
			r.logger.Error("failed to get my duels", zap.Error(err))
			return nil, 0, err
		}

	default:
		// Default to all visible duels
		total, err = r.q.GetVisibleDuelsCount(ctx)
		if err != nil {
			r.logger.Error("failed to get visible duels count", zap.Error(err))
			return nil, 0, err
		}
		sqlcDuels, err = r.q.GetDuels(ctx, sqlc.GetDuelsParams{
			Limit:  pageRequest.PageSize(),
			Offset: pageRequest.Offset(),
		})
		if err != nil {
			r.logger.Error("failed to get duels", zap.Error(err))
			return nil, 0, err
		}
	}

	duels := make([]*duelDomain.Duel, len(sqlcDuels))
	for i, sqlcDuel := range sqlcDuels {
		duel, mapErr := mapDuel(&sqlcDuel)
		if mapErr != nil {
			r.logger.Error("failed to map duel", zap.Error(mapErr))
			return nil, 0, mapErr
		}

		// Load related data
		if err = r.loadDuelRelatedData(ctx, duel); err != nil {
			r.logger.Error("failed to load duel related data", zap.Error(err))
			return nil, 0, err
		}

		duels[i] = duel
	}

	return duels, total, nil
}

func (r *duelRepository) loadDuelRelatedData(
	ctx context.Context, duel *duelDomain.Duel,
) error {
	duelID, err := pgUUID(duel.ID.String())
	if err != nil {
		return err
	}

	// Поочерёдно добавляем все связанные сущности
	if err = r.loadParticipants(ctx, duelID, duel); err != nil {
		return err
	}
	if err = r.loadStakes(ctx, duelID, duel); err != nil {
		return err
	}
	if err = r.loadRoundsAndRolls(ctx, duelID, duel); err != nil {
		return err
	}

	return nil
}

// loadParticipants loads participants.
func (r *duelRepository) loadParticipants(
	ctx context.Context,
	duelID pgtype.UUID,
	duel *duelDomain.Duel,
) error {
	rows, err := r.q.GetDuelParticipants(ctx, duelID)
	if err != nil {
		return err
	}
	parts := make([]duelDomain.Participant, len(rows))
	for i, p := range rows {
		uid, uidErr := duelDomain.NewTelegramUserID(p.TelegramUserID)
		if uidErr != nil {
			return uidErr
		}
		pb := duelDomain.NewParticipantBuilder().WithTelegramUserID(uid)
		if p.IsCreator {
			pb = pb.AsCreator()
		}
		parts[i], err = pb.Build()
		if err != nil {
			return err
		}
	}
	duel.Participants = parts
	return nil
}

// loadStakes loads stakes.
func (r *duelRepository) loadStakes(
	ctx context.Context,
	duelID pgtype.UUID,
	duel *duelDomain.Duel,
) error {
	rows, err := r.q.GetDuelStakes(ctx, duelID)
	if err != nil {
		return err
	}
	stakes := make([]duelDomain.Stake, len(rows))
	for i, s := range rows {
		uid, uidErr := duelDomain.NewTelegramUserID(s.TelegramUserID)
		if uidErr != nil {
			return uidErr
		}
		gift, giftErr := duelDomain.NewStakedGiftBuilder().
			WithID(s.GiftID.String()).
			Build()
		if giftErr != nil {
			return giftErr
		}
		stakes[i], err = duelDomain.NewStakeBuilder(uid).
			WithGift(gift).
			Build()
		if err != nil {
			return err
		}
	}
	duel.Stakes = stakes
	return nil
}

// loadRoundsAndRolls loads rounds and rolls.
func (r *duelRepository) loadRoundsAndRolls(
	ctx context.Context,
	duelID pgtype.UUID,
	duel *duelDomain.Duel,
) error {
	roundsRows, err := r.q.GetDuelRounds(ctx, duelID)
	if err != nil {
		return err
	}
	rollsRows, err := r.q.GetDuelRolls(ctx, duelID)
	if err != nil {
		return err
	}
	// Группируем броски по номеру раунда
	rollsBy := make(map[int32][]sqlc.DuelRoll)
	for _, rl := range rollsRows {
		rollsBy[rl.RoundNumber] = append(rollsBy[rl.RoundNumber], rl)
	}
	// Build each Round through the builder
	var rounds []duelDomain.Round
	for _, rr := range roundsRows {
		// take participants from the builder (they are already loaded in DuelBuilder)
		participants := duel.Participants // допустимо, Participants уже в билдере
		rbuilder := duelDomain.NewRoundBuilder().
			WithRoundNumber(rr.RoundNumber).
			WithParticipants(extractIDs(participants))
		// добавляем броски
		for _, rl := range rollsBy[rr.RoundNumber] {
			uid, uidErr := duelDomain.NewTelegramUserID(rl.TelegramUserID)
			if uidErr != nil {
				return uidErr
			}
			rbuilder.AddRoll(duelDomain.Roll{
				TelegramUserID: uid,
				DiceValue:      int32(rl.DiceValue),
				RolledAt:       rl.RolledAt.Time,
				IsAutoRolled:   rl.IsAutoRolled,
			})
		}
		round, roundErr := rbuilder.Build()
		if roundErr != nil {
			return roundErr
		}
		rounds = append(rounds, round)
	}
	duel.Rounds = rounds
	return nil
}

// extractIDs extracts []TelegramUserID from []Participant.
func extractIDs(parts []duelDomain.Participant) []duelDomain.TelegramUserID {
	ids := make([]duelDomain.TelegramUserID, len(parts))
	for i, p := range parts {
		ids[i] = p.TelegramUserID
	}
	return ids
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
	roundNumber int32,
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
		DuelID:            pgDuelID,
		RoundNumber:       roundNumber,
		TelegramUserID:    roll.TelegramUserID.Int64(),
		DiceValue:         diceValue,
		RolledAt:          rolledAt,
		IsAutoRolled:      roll.IsAutoRolled,
		TelegramMessageID: roll.TelegramMessageID,
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

	params := sqlc.UpdateDuelStatusParams{
		ID:          pgDuelID,
		Status:      sqlc.NullDuelStatus{DuelStatus: sqlc.DuelStatus(status), Valid: true},
		CompletedAt: time,
	}

	if winnerID != nil {
		params.WinnerTelegramUserID = pgInt64(winnerID.Int64())
	}

	return r.q.UpdateDuelStatus(ctx, params)
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
