package app

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type MessageStatus int32

const (
	MessageStatusQueued     MessageStatus = 0
	MessageStatusProcessing MessageStatus = 1
	MessageStatusSent       MessageStatus = 2
)

func (s MessageStatus) String() string {
	switch s {
	case MessageStatusQueued:
		return "QUEUED"
	case MessageStatusProcessing:
		return "PROCESSING"
	case MessageStatusSent:
		return "SENT"
	default:
		return "UNKNOWN"
	}
}

type Message struct {
	ID        int64
	PatientID string
	Content   string
	Status    MessageStatus
	QueuedAt  time.Time
	SentAt    time.Time
}

type MessageListener func(Message)

type MessageQueue struct {
	mu        sync.Mutex
	seq       int64
	messages  []Message
	queue     []Message
	listeners []MessageListener
	minDelay  time.Duration
	maxDelay  time.Duration
}

func NewMessageQueue(minDelay, maxDelay time.Duration) *MessageQueue {
	return &MessageQueue{
		minDelay: minDelay,
		maxDelay: maxDelay,
	}
}

func (q *MessageQueue) Enqueue(patientID, content string) Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.seq++
	msg := Message{
		ID:        q.seq,
		PatientID: patientID,
		Content:   content,
		Status:    MessageStatusQueued,
		QueuedAt:  time.Now().UTC(),
	}
	q.messages = append(q.messages, msg)
	q.queue = append(q.queue, msg)
	q.notifyLocked(msg)
	return msg
}

func (q *MessageQueue) ProcessNext(ctx context.Context) *Message {
	q.mu.Lock()
	if len(q.queue) == 0 {
		q.mu.Unlock()
		return nil
	}
	msg := q.queue[0]
	q.queue = q.queue[1:]
	q.mu.Unlock()

	// Mark as processing
	msg = q.updateMessage(msg, MessageStatusProcessing, false)
	q.notify(msg)

	// Simulate delay
	delay := q.minDelay + time.Duration(rand.Float64()*float64(q.maxDelay-q.minDelay))
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.Now().Add(delay)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			continue
		}
	}

	// Mark as sent
	msg = q.updateMessage(msg, MessageStatusSent, true)
	q.notify(msg)
	return &msg
}

func (q *MessageQueue) updateMessage(msg Message, status MessageStatus, sent bool) Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	if sent {
		msg.SentAt = time.Now().UTC()
	}
	msg.Status = status

	for i, m := range q.messages {
		if m.ID == msg.ID {
			q.messages[i] = msg
			break
		}
	}
	return msg
}

func (q *MessageQueue) ListMessages() []Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	result := make([]Message, len(q.messages))
	copy(result, q.messages)
	return result
}

func (q *MessageQueue) AddListener(listener MessageListener) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.listeners = append(q.listeners, listener)
}

func (q *MessageQueue) notify(msg Message) {
	q.mu.Lock()
	listeners := make([]MessageListener, len(q.listeners))
	copy(listeners, q.listeners)
	q.mu.Unlock()

	for _, listener := range listeners {
		func() {
			defer func() { recover() }()
			listener(msg)
		}()
	}
}

func (q *MessageQueue) notifyLocked(msg Message) {
	listeners := make([]MessageListener, len(q.listeners))
	copy(listeners, q.listeners)

	for _, listener := range listeners {
		func() {
			defer func() { recover() }()
			listener(msg)
		}()
	}
}
