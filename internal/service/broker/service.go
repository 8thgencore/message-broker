package broker

import (
	"context"
	"fmt"
	"sync"

	"github.com/8thgencore/message-broker/internal/config"
	"github.com/8thgencore/message-broker/internal/model"
)

type Service struct {
	queues map[string]*Queue
	mu     sync.RWMutex
}

func NewService(cfg []config.QueueConfig) *Service {
	queues := make(map[string]*Queue, len(cfg))
	for _, qCfg := range cfg {
		queues[qCfg.Name] = NewQueue(qCfg.Name, qCfg.Size, qCfg.MaxSubscribers)
	}

	return &Service{
		queues: queues,
	}
}

func (s *Service) PublishMessage(ctx context.Context, queueName string, data []byte) (string, error) {
	s.mu.RLock()
	queue, ok := s.queues[queueName]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("queue %s not found", queueName)
	}

	return queue.Publish(ctx, data)
}

func (s *Service) Subscribe(ctx context.Context, queueName string) (string, <-chan model.Message, <-chan struct{}, error) {
	s.mu.RLock()
	queue, ok := s.queues[queueName]
	s.mu.RUnlock()

	if !ok {
		return "", nil, nil, fmt.Errorf("queue %s not found", queueName)
	}

	return queue.Subscribe(ctx)
}

func (s *Service) Unsubscribe(id string) {
	s.mu.RLock()
	for _, queue := range s.queues {
		queue.Unsubscribe(id)
	}
	s.mu.RUnlock()
}
