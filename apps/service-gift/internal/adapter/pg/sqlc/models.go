// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package sqlc

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type GiftEventType string

const (
	GiftEventTypeStake            GiftEventType = "stake"
	GiftEventTypeReturnFromGame   GiftEventType = "return_from_game"
	GiftEventTypeDeposit          GiftEventType = "deposit"
	GiftEventTypeWithdrawRequest  GiftEventType = "withdraw_request"
	GiftEventTypeWithdrawComplete GiftEventType = "withdraw_complete"
	GiftEventTypeWithdrawFail     GiftEventType = "withdraw_fail"
)

func (e *GiftEventType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = GiftEventType(s)
	case string:
		*e = GiftEventType(s)
	default:
		return fmt.Errorf("unsupported scan type for GiftEventType: %T", src)
	}
	return nil
}

type NullGiftEventType struct {
	GiftEventType GiftEventType
	Valid         bool // Valid is true if GiftEventType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullGiftEventType) Scan(value interface{}) error {
	if value == nil {
		ns.GiftEventType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.GiftEventType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullGiftEventType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.GiftEventType), nil
}

type GiftStatus string

const (
	GiftStatusOwned           GiftStatus = "owned"
	GiftStatusInGame          GiftStatus = "in_game"
	GiftStatusWithdrawPending GiftStatus = "withdraw_pending"
	GiftStatusWithdrawn       GiftStatus = "withdrawn"
)

func (e *GiftStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = GiftStatus(s)
	case string:
		*e = GiftStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for GiftStatus: %T", src)
	}
	return nil
}

type NullGiftStatus struct {
	GiftStatus GiftStatus
	Valid      bool // Valid is true if GiftStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullGiftStatus) Scan(value interface{}) error {
	if value == nil {
		ns.GiftStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.GiftStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullGiftStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.GiftStatus), nil
}

type Gift struct {
	ID               pgtype.UUID
	TelegramGiftID   int64
	CollectibleID    int32
	OwnerTelegramID  int64
	UpgradeMessageID int32
	Title            string
	Slug             string
	Price            pgtype.Numeric
	CollectionID     int32
	ModelID          int32
	BackdropID       int32
	SymbolID         int32
	Status           GiftStatus
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	WithdrawnAt      pgtype.Timestamptz
}

type GiftBackdrop struct {
	ID             int32
	Name           string
	ShortName      string
	RarityPerMille int32
	CenterColor    pgtype.Text
	EdgeColor      pgtype.Text
	PatternColor   pgtype.Text
	TextColor      pgtype.Text
}

type GiftCollection struct {
	ID        int32
	Name      string
	ShortName string
}

type GiftEvent struct {
	ID             pgtype.UUID
	GiftID         pgtype.UUID
	EventType      GiftEventType
	TelegramUserID pgtype.Int8
	RelatedGameID  pgtype.UUID
	OccurredAt     pgtype.Timestamptz
}

type GiftModel struct {
	ID             int32
	CollectionID   int32
	Name           string
	ShortName      string
	RarityPerMille int32
}

type GiftSymbol struct {
	ID             int32
	Name           string
	ShortName      string
	RarityPerMille int32
}
