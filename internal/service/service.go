package service

import (
	"context"

	"github.com/8thgencore/message-broker/internal/model"
)

type BrokerService interface {
	PublishMessage(ctx context.Context, queueName string, data []byte) (string, error)
	Subscribe(ctx context.Context, queueName string) (string, <-chan model.Message, <-chan struct{}, error)
	Unsubscribe(id string)
}
