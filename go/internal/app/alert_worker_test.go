package app

import (
	"context"
	"testing"
	"time"
)

func TestAlertWorkerCreatesAlertForAbnormalVitals(t *testing.T) {
	store := NewInMemoryStore()
	pubsub := NewPubSub()
	worker := NewAlertWorker(pubsub, store, 8, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)

	normal := Vital{
		ID:         1,
		PatientID:  "patient-normal",
		Systolic:   120,
		Diastolic:  80,
		TakenAt:    time.Now().UTC(),
		ReceivedAt: time.Now().UTC(),
	}
	if err := pubsub.Publish(ctx, Event{Type: EventTypeVitalReceived, Vital: normal}); err != nil {
		t.Fatalf("publish normal vital: %v", err)
	}

	abnormal := Vital{
		ID:         2,
		PatientID:  "patient-abnormal",
		Systolic:   200,
		Diastolic:  130,
		TakenAt:    time.Now().UTC(),
		ReceivedAt: time.Now().UTC(),
	}
	if err := pubsub.Publish(ctx, Event{Type: EventTypeVitalReceived, Vital: abnormal}); err != nil {
		t.Fatalf("publish abnormal vital: %v", err)
	}

	waitFor(t, 500*time.Millisecond, func() bool {
		alerts, err := store.ListAlerts(ctx)
		if err != nil {
			return false
		}
		return len(alerts) == 1
	})

	alerts, err := store.ListAlerts(ctx)
	if err != nil {
		t.Fatalf("list alerts failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].PatientID != "patient-abnormal" {
		t.Fatalf("unexpected patient id: %s", alerts[0].PatientID)
	}
	if alerts[0].VitalID != abnormal.ID {
		t.Fatalf("unexpected vital id: %d", alerts[0].VitalID)
	}
	if alerts[0].Status != AlertStatusActive {
		t.Fatalf("unexpected alert status: %v", alerts[0].Status)
	}
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	deadline := time.After(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for condition")
		case <-ticker.C:
			if cond() {
				return
			}
		}
	}
}
