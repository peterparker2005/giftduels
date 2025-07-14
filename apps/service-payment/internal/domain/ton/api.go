package ton

import (
	"context"

	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

// MasterchainInfo — минимальная часть данных о мастере.
type MasterchainInfo struct {
	SeqNo uint32
}

// Transaction — доменная модель прихода средств.
type Transaction struct {
	Sender   string
	Amount   *tonamount.TonAmount // строковое представление с нужными десятичными знаками
	Currency string               // "TON" или код джеттона
	Payload  string               // payload/comment from transaction body
	LastLT   uint64               // для сохранения курсора
}

// API — абстракция над tonutils-go.
type API interface {
	CurrentMasterchainInfo(ctx context.Context) (MasterchainInfo, error)
	GetAccountLastLT(ctx context.Context, addr string) (uint64, error)
	SubscribeTransactions(ctx context.Context, addr string, fromLT uint64, out chan<- Transaction) error
}
