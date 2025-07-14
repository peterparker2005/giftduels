package clients

import (
	"fmt"

	telegrambotv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/telegrambot/v1"
	"google.golang.org/grpc"
)

type TelegramBotClient struct {
	conn    *grpc.ClientConn
	Private telegrambotv1.TelegramBotPrivateServiceClient
}

func NewTelegramBotClient(address string, opts ...grpc.DialOption) (*TelegramBotClient, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial telegram bot service %s: %w", address, err)
	}
	return &TelegramBotClient{
		conn:    conn,
		Private: telegrambotv1.NewTelegramBotPrivateServiceClient(conn),
	}, nil
}

func (c *TelegramBotClient) Close() error {
	return c.conn.Close()
}
