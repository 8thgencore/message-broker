package broker

import (
	"context"
	"errors"
	"sync"

	"github.com/8thgencore/message-broker/internal/model"
	"github.com/google/uuid"
)

// ErrQueueFull is the error returned when the queue is full.
var ErrQueueFull = errors.New("queue is full")

// ErrTooManySubscribers is the error returned when there are too many subscribers.
var ErrTooManySubscribers = errors.New("too many subscribers")

// subscriber is a subscriber to a queue.
type subscriber struct {
	id    string
	msgCh chan model.Message
	done  chan struct{}
}

// Queue is a queue of messages.
type Queue struct {
	name           string
	size           int
	maxSubscribers int
	messages       []model.Message
	subscribers    map[string]*subscriber
	mu             sync.RWMutex
}

// NewQueue creates a new Queue instance.
func NewQueue(name string, size, maxSubscribers int) *Queue {
	return &Queue{
		name:           name,
		size:           size,
		maxSubscribers: maxSubscribers,
		messages:       make([]model.Message, 0, size),
		subscribers:    make(map[string]*subscriber),
	}
}

// Publish publishes a message to the queue.
func (q *Queue) Publish(ctx context.Context, data []byte) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) >= q.size {
		return "", ErrQueueFull
	}

	msg := model.Message{
		ID:   uuid.New().String(),
		Data: data,
	}

	q.messages = append(q.messages, msg)

	for _, sub := range q.subscribers {
		select {
		case sub.msgCh <- msg:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return msg.ID, nil
}

// Subscribe subscribes to the queue and streams messages to the client.
func (q *Queue) Subscribe(ctx context.Context) (string, <-chan model.Message, <-chan struct{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.subscribers) >= q.maxSubscribers {
		return "", nil, nil, ErrTooManySubscribers
	}

	id := uuid.New().String()
	msgCh := make(chan model.Message, q.size)
	done := make(chan struct{})

	sub := &subscriber{
		id:    id,
		msgCh: msgCh,
		done:  done,
	}

	q.subscribers[id] = sub

	for _, msg := range q.messages {
		select {
		case msgCh <- msg:
		case <-ctx.Done():
			close(msgCh)
			close(done)
			delete(q.subscribers, id)
			return "", nil, nil, ctx.Err()
		}
	}

	return id, msgCh, done, nil
}

// Unsubscribe unsubscribes from the queue.
func (q *Queue) Unsubscribe(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if sub, ok := q.subscribers[id]; ok {
		close(sub.msgCh)
		close(sub.done)
		delete(q.subscribers, id)
	}
}
