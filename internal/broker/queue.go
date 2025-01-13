package broker

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrQueueFull        = errors.New("queue is full")
	ErrTooManySubscribers = errors.New("too many subscribers")
)

type Message struct {
	ID   string
	Data []byte
}

type subscriber struct {
	id     string
	msgCh  chan Message
	done   chan struct{}
}

type Queue struct {
	name           string
	size           int
	maxSubscribers int
	messages       []Message
	subscribers    map[string]*subscriber
	mu            sync.RWMutex
}

func NewQueue(name string, size, maxSubscribers int) *Queue {
	return &Queue{
		name:           name,
		size:           size,
		maxSubscribers: maxSubscribers,
		messages:       make([]Message, 0, size),
		subscribers:    make(map[string]*subscriber),
	}
}

func (q *Queue) Publish(ctx context.Context, data []byte) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) >= q.size {
		return "", ErrQueueFull
	}

	msg := Message{
		ID:   uuid.New().String(),
		Data: data,
	}

	q.messages = append(q.messages, msg)

	// Отправляем сообщение всем подписчикам
	for _, sub := range q.subscribers {
		select {
		case sub.msgCh <- msg:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return msg.ID, nil
}

func (q *Queue) Subscribe(ctx context.Context) (string, <-chan Message, <-chan struct{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.subscribers) >= q.maxSubscribers {
		return "", nil, nil, ErrTooManySubscribers
	}

	id := uuid.New().String()
	msgCh := make(chan Message, q.size)
	done := make(chan struct{})

	sub := &subscriber{
		id:    id,
		msgCh: msgCh,
		done:  done,
	}

	q.subscribers[id] = sub

	// Отправляем все существующие сообщения новому подписчику
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

func (q *Queue) Unsubscribe(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if sub, ok := q.subscribers[id]; ok {
		close(sub.msgCh)
		close(sub.done)
		delete(q.subscribers, id)
	}
} 