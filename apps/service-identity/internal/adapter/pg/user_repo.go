package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/domain/user"
)

type UserRepository struct {
	db *sqlc.Queries
}

func NewUserRepo(pool *pgxpool.Pool) user.UserRepository {
	return &UserRepository{db: sqlc.New(pool)}
}

func (r *UserRepository) GetByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*user.User, error) {
	dbUser, err := r.db.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return UserToDomain(dbUser), nil
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	params user.CreateUserParams,
) (*user.User, error) {
	sqlcParams := CreateUserParamsToSQLC(params)
	dbUser, err := r.db.CreateUser(ctx, sqlcParams)
	if err != nil {
		return nil, err
	}

	return UserToDomain(dbUser), nil
}

func (r *UserRepository) UpsertUser(
	ctx context.Context,
	params user.CreateUserParams,
) (*user.User, error) {
	sqlcParams := UpsertUserParamsToSQLC(params)
	dbUser, err := r.db.UpsertUser(ctx, sqlcParams)
	if err != nil {
		return nil, err
	}

	return UserToDomain(dbUser), nil
}
