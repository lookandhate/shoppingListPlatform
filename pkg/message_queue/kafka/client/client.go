package kafka

import (
	"context"

	"github.com/cum4ati/shoppingListPlatform/pkg/message_queue/kafka/client/consumer"
)

type Consumer interface {
	Consume(ctx context.Context, topicName string, handler consumer.Handler) (err error)
	Close() error
}
