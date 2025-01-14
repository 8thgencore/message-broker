package service

import (
	"context"

	"github.com/8thgencore/message-broker/internal/service/broker"
)

// BrokerService is the interface for the broker service.
type BrokerService interface {
	PublishMessage(ctx context.Context, queueName string, data []byte) (string, error)
	Subscribe(ctx context.Context, queueName string) (*broker.SubscribeResult, error)
	Unsubscribe(id string)
}
