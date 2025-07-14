package pg

import (
	"errors"
	"fmt"
	"time"

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

func pgNumeric(amount string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if amount == "" {
		// by default n.Valid == false
		return n, nil
	}
	// Numeric.Scan can parse string to pgtype.Numeric
	if err := n.Scan(amount); err != nil {
		return n, fmt.Errorf("pgNumeric: invalid numeric %q: %w", amount, err)
	}
	return n, nil
}

func fromPgNumeric(n pgtype.Numeric) (string, error) {
	if !n.Valid {
		return "", errors.New("pgNumeric: invalid numeric")
	}
	// Value returns driver.Value (usually string)
	v, err := n.Value()
	if err != nil {
		// should not happen, but just in case
		return "", err
	}
	switch s := v.(type) {
	case string:
		return s, nil
	case []byte:
		return string(s), nil
	default:
		// just in case
		return fmt.Sprint(v), nil
	}
}

func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgInt64(i int64) pgtype.Int8 {
	return pgtype.Int8{Int64: i, Valid: true}
}
