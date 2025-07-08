package ton

import "context"

// CursorRepository хранит и возвращает last_lt для заданного адреса и сети.
type CursorRepository interface {
	// Get возвращает сохранённый lastLT для walletAddress и network.
	// Если записи нет, возвращает 0 и nil-ошибку.
	Get(ctx context.Context, network, walletAddress string) (uint64, error)
	// Upsert сохраняет или обновляет курсор для walletAddress и network.
	Upsert(ctx context.Context, network, walletAddress string, lastLT uint64) error
}
