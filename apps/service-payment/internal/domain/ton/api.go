package ton

import "context"

// MasterchainInfo — минимальная часть данных о мастере
type MasterchainInfo struct {
	SeqNo uint32
}

// Transaction — доменная модель прихода средств
type Transaction struct {
	Sender   string
	Amount   string // строковое представление с нужными десятичными знаками
	Currency string // "TON" или код джеттона
	LastLT   uint64 // для сохранения курсора
}

// TonAPI — абстракция над tonutils-go
type TonAPI interface {
	CurrentMasterchainInfo(ctx context.Context) (MasterchainInfo, error)
	GetAccountLastLT(ctx context.Context, addr string) (uint64, error)
	SubscribeTransactions(ctx context.Context, addr string, fromLT uint64, out chan<- Transaction) error
}
