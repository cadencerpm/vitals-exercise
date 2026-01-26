package app

import (
	"context"
	"log"
	"time"
)

type MessageWorker struct {
	queue *MessageQueue
}

func NewMessageWorker(queue *MessageQueue) *MessageWorker {
	return &MessageWorker{queue: queue}
}

func (w *MessageWorker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg := w.queue.ProcessNext(ctx)
			if msg == nil && ctx.Err() == nil {
				// No message in queue, wait a bit
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}
			if msg != nil {
				log.Printf("[Message] Sent to %s: %s", msg.PatientID, msg.Content)
			}
		}
	}
}
