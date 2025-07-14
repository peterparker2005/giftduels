package pg

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgUUID(id string) (pgtype.UUID, error) {
	if id == "" {
		return pgtype.UUID{}, errors.New("uuid cannot be empty")
	}

	u, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: u, Valid: true}, nil
}

func mustPgUUID(id string) pgtype.UUID {
	if id == "" {
		panic("uuid cannot be empty")
	}

	v, err := pgUUID(id)
	if err != nil {
		panic(err)
	}
	return v
}

// pgUUIDToString converts pgtype.UUID to string.
func pgUUIDToString(pgUUID pgtype.UUID) string {
	if !pgUUID.Valid {
		return ""
	}
	return uuid.UUID(pgUUID.Bytes).String()
}

// pgTimestampToTime converts pgtype.Timestamptz to *time.Time.
func pgTimestampToTime(pgTimestamp pgtype.Timestamptz) *time.Time {
	if !pgTimestamp.Valid {
		return nil
	}
	return &pgTimestamp.Time
}

// pgTimestampToTimeRequired converts pgtype.Timestamptz to time.Time (panics if invalid).
func pgTimestampToTimeRequired(pgTimestamp pgtype.Timestamptz) time.Time {
	if !pgTimestamp.Valid {
		panic("expected valid timestamp")
	}
	return pgTimestamp.Time
}

// pgTextToString converts pgtype.Text to *string.
func pgTextToString(pgText pgtype.Text) *string {
	if !pgText.Valid {
		return nil
	}
	return &pgText.String
}

// pgInt8ToInt64 converts pgtype.Int8 to *int64.
func pgInt8ToInt64(pgInt8 pgtype.Int8) (int64, error) {
	if !pgInt8.Valid {
		return 0, errors.New("pgInt8ToInt64: invalid int8")
	}
	return pgInt8.Int64, nil
}

// timeToPgTimestamp converts time.Time to pgtype.Timestamptz.
func timeToPgTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// stringPtrToPgText converts *string to pgtype.Text.
func stringPtrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// int64PtrToPgInt8 converts *int64 to pgtype.Int8.
func int64PtrToPgInt8(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

func pgNumeric(amount string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if amount == "" {
		// by default n.Valid == false
		return n, nil
	}

	// Trim whitespace to handle edge cases
	amount = strings.TrimSpace(amount)
	if amount == "" {
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
