package kafka

import (
	"context"

	"github.com/lookandhate/course_platform_lib/pkg/message_queue/kafka/client/consumer"
)

type Consumer interface {
	Consume(ctx context.Context, topicName string, handler consumer.Handler) (err error)
	Close() error
}
