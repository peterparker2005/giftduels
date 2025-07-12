package pg

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgUUID(id string) (pgtype.UUID, error) {
	var u uuid.UUID
	u, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: u, Valid: true}, nil
}

func mustPgUUID(id string) pgtype.UUID {
	v, err := pgUUID(id)
	if err != nil {
		panic(err)
	}
	return v
}

func mustPgUUIDs(ids []string) []pgtype.UUID {
	out := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		out[i] = mustPgUUID(id)
	}
	return out
}
