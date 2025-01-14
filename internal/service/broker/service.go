package broker

import (
	"context"
	"fmt"
	"sync"

	"github.com/8thgencore/message-broker/internal/config"
	"github.com/8thgencore/message-broker/internal/model"
)

// SubscribeResult represents the result of a subscription.
type SubscribeResult struct {
	ID       string
	Messages <-chan model.Message
	Done     <-chan struct{}
}

// Service is the broker service.
type Service struct {
	queues map[string]*Queue
	mu     sync.RWMutex
}

// NewService creates a new Service instance.
func NewService(cfg []config.QueueConfig) *Service {
	queues := make(map[string]*Queue, len(cfg))
	for _, qCfg := range cfg {
		queues[qCfg.Name] = NewQueue(qCfg.Name, qCfg.Size, qCfg.MaxSubscribers)
	}

	return &Service{
		queues: queues,
	}
}

// PublishMessage publishes a message to the given queue.
func (s *Service) PublishMessage(ctx context.Context, queueName string, data []byte) (string, error) {
	s.mu.RLock()
	queue, ok := s.queues[queueName]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("queue %s not found", queueName)
	}

	return queue.Publish(ctx, data)
}

// Subscribe subscribes to the given queue and streams messages to the client.
func (s *Service) Subscribe(ctx context.Context, queueName string) (*SubscribeResult, error) {
	s.mu.RLock()
	queue, ok := s.queues[queueName]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("queue %s not found", queueName)
	}

	id, msgCh, done, err := queue.Subscribe(ctx)
	if err != nil {
		return nil, err
	}

	return &SubscribeResult{
		ID:       id,
		Messages: msgCh,
		Done:     done,
	}, nil
}

// Unsubscribe unsubscribes from the given queue.
func (s *Service) Unsubscribe(id string) {
	s.mu.RLock()
	for _, queue := range s.queues {
		queue.Unsubscribe(id)
	}
	s.mu.RUnlock()
}
