package pg

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func uuidToPg(id string) (pgtype.UUID, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: u, Valid: true}, nil
}

func mustPgUUID(id string) pgtype.UUID {
	v, err := uuidToPg(id)
	if err != nil {
		panic(err) // или пробросьте ошибку выше
	}
	return v
}
