package pg_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
)

func TestMapPGError(t *testing.T) {
	tests := []struct {
		name         string
		input        error
		want         error
		wantNotFound bool
		wantConflict bool
	}{
		{
			name:         "nil → nil",
			input:        nil,
			want:         nil,
			wantNotFound: false,
			wantConflict: false,
		},
		{
			name:         "pgx.ErrNoRows → ErrNotFound",
			input:        pgx.ErrNoRows,
			want:         pg.ErrNotFound,
			wantNotFound: true,
			wantConflict: false,
		},
		{
			name: "unique‐violation 23505 → ErrConflict",
			input: &pgconn.PgError{
				Code: "23505",
			},
			want:         pg.ErrConflict,
			wantNotFound: false,
			wantConflict: true,
		},
		{
			name:         "other PgError → passthrough",
			input:        &pgconn.PgError{Code: "99999", Message: "foo"},
			want:         &pgconn.PgError{Code: "99999", Message: "foo"},
			wantNotFound: false,
			wantConflict: false,
		},
		{
			name:         "arbitrary error → passthrough",
			input:        errors.New("hello"),
			want:         errors.New("hello"),
			wantNotFound: false,
			wantConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pg.MapPGError(tt.input)

			// сравниваем nil
			if (got == nil) != (tt.want == nil) {
				t.Fatalf("MapPGError(%v) = %v; want %v", tt.input, got, tt.want)
			}
			// если не nil, то сравниваем текст
			if got != nil && got.Error() != tt.want.Error() {
				t.Errorf("MapPGError(%v) = %q; want %q", tt.input, got.Error(), tt.want.Error())
			}

			// проверяем IsNotFound
			if pg.IsNotFound(got) != tt.wantNotFound {
				t.Errorf("IsNotFound(%v) = %v; want %v", got, pg.IsNotFound(got), tt.wantNotFound)
			}
			// проверяем IsConflict
			if pg.IsConflict(got) != tt.wantConflict {
				t.Errorf("IsConflict(%v) = %v; want %v", got, pg.IsConflict(got), tt.wantConflict)
			}
		})
	}
}
