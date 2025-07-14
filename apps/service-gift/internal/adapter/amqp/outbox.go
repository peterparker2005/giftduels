package amqp

import (
	"github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	giftevents "github.com/peterparker2005/giftduels/packages/events/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

func ProvideOutbox(
	pool *pgxpool.Pool,
	log *logger.Logger,
	pub message.Publisher,
) (*forwarder.Forwarder, error) {
	db := stdlib.OpenDBFromPool(pool)

	sqlSubscriber, err := sql.NewSubscriber(
		db,
		sql.SubscriberConfig{
			SchemaAdapter:    sql.DefaultPostgreSQLSchema{},
			OffsetsAdapter:   sql.DefaultPostgreSQLOffsetsAdapter{},
			InitializeSchema: true,
		},
		nil,
		// logger.NewWatermill(log),
	)
	if err != nil {
		log.Fatal("failed to create sql subscriber", zap.Error(err))
	}

	return forwarder.NewForwarder(
		sqlSubscriber,
		pub,
		logger.NewWatermill(log),
		forwarder.Config{
			ForwarderTopic: giftevents.SQLOutboxTopic,
		},
	)
}
