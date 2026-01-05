package app

import (
	"context"
	"log"
	"time"
)

type AlertWorker struct {
	sub    <-chan Event
	cancel func()
	store  Store
}

func NewAlertWorker(pubsub *PubSub, store Store, buffer int) *AlertWorker {
	sub, cancel := pubsub.Subscribe(buffer)
	return &AlertWorker{
		sub:    sub,
		cancel: cancel,
		store:  store,
	}
}

func (w *AlertWorker) Run(ctx context.Context) {
	defer w.cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.sub:
			if !ok {
				return
			}
			w.handleEvent(ctx, event)
		}
	}
}

func (w *AlertWorker) handleEvent(ctx context.Context, event Event) {
	if event.Type != EventTypeVitalReceived {
		return
	}
	if !IsAbnormal(event.Vital) {
		return
	}

	alert := Alert{
		VitalID:    event.Vital.ID,
		PatientID:  event.Vital.PatientID,
		Systolic:   event.Vital.Systolic,
		Diastolic:  event.Vital.Diastolic,
		TakenAt:    event.Vital.TakenAt,
		ReceivedAt: event.Vital.ReceivedAt,
		Reason:     AlertReason(event.Vital),
		Status:     AlertStatusActive,
		Created:    time.Now().UTC(),
	}

	if _, err := w.store.AddAlert(ctx, alert); err != nil {
		log.Printf("alert worker failed to store alert: %v", err)
	}
}
