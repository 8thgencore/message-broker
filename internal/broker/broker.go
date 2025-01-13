package broker

import (
	"context"
	"fmt"
	"sync"

	"github.com/8thgencore/message-broker/internal/config"
)

type Broker struct {
	queues map[string]*Queue
	mu     sync.RWMutex
}

func NewBroker(cfg []config.QueueConfig) *Broker {
	queues := make(map[string]*Queue, len(cfg))
	for _, qCfg := range cfg {
		queues[qCfg.Name] = NewQueue(qCfg.Name, qCfg.Size, qCfg.MaxSubscribers)
	}

	return &Broker{
		queues: queues,
	}
}

func (b *Broker) PublishMessage(ctx context.Context, queueName string, data []byte) (string, error) {
	b.mu.RLock()
	queue, ok := b.queues[queueName]
	b.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("queue %s not found", queueName)
	}

	return queue.Publish(ctx, data)
}

func (b *Broker) Subscribe(ctx context.Context, queueName string) (string, <-chan Message, <-chan struct{}, error) {
	b.mu.RLock()
	queue, ok := b.queues[queueName]
	b.mu.RUnlock()

	if !ok {
		return "", nil, nil, fmt.Errorf("queue %s not found", queueName)
	}

	return queue.Subscribe(ctx)
} 